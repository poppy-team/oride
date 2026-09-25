//! Documento (buffer + path + seleção + undo) e store multi-tab.

use std::path::{Path, PathBuf};

use thiserror::Error;

use crate::buffer::{Buffer, BufferError};
use crate::position::{ByteOffset, Caret};
use crate::selection::Selection;
use crate::undo::{Edit, UndoStack};

/// Identificador opaco de documento aberto (tab).
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct DocumentId(u64);

impl DocumentId {
    #[must_use]
    pub const fn as_u64(self) -> u64 {
        self.0
    }

    /// Construtor para testes / split state.
    #[must_use]
    pub const fn from_raw(id: u64) -> Self {
        Self(id)
    }
}

/// Resumo de uma tab para a UI.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct TabSummary {
    pub id: DocumentId,
    pub title: String,
    pub dirty: bool,
    pub active: bool,
}

#[derive(Debug, Default)]
pub struct SaveAllReport {
    pub saved: usize,
    pub saved_paths: Vec<PathBuf>,
    pub skipped_without_path: usize,
    pub failures: Vec<SaveFailure>,
}

#[derive(Debug)]
pub struct SaveFailure {
    pub path: PathBuf,
    pub error: DocumentError,
}

#[derive(Debug, Error)]
pub enum DocumentError {
    #[error(transparent)]
    Buffer(#[from] BufferError),
    #[error("io: {0}")]
    Io(#[from] std::io::Error),
    #[error("document {0:?} not found")]
    NotFound(DocumentId),
    #[error("no active document")]
    NoActiveDocument,
}

/// Um arquivo (ou buffer sem path) aberto no editor.
#[derive(Debug)]
pub struct Document {
    id: DocumentId,
    path: Option<PathBuf>,
    buffer: Buffer,
    selection: Selection,
    /// Cursores extras (byte offsets); o primário é `selection.head`.
    extra_carets: Vec<ByteOffset>,
    undo: UndoStack,
    dirty: bool,
    /// Coluna preferida ao mover ↑/↓ (estilo editores clássicos).
    preferred_column: Option<usize>,
    version: u64,
}

impl Document {
    #[must_use]
    pub fn id(&self) -> DocumentId {
        self.id
    }

    #[must_use]
    pub fn version(&self) -> u64 {
        self.version
    }

    #[must_use]
    pub fn path(&self) -> Option<&Path> {
        self.path.as_deref()
    }

    #[must_use]
    pub fn is_dirty(&self) -> bool {
        self.dirty
    }

    #[must_use]
    pub fn buffer(&self) -> &Buffer {
        &self.buffer
    }

    #[must_use]
    pub fn selection(&self) -> Selection {
        self.selection
    }

    pub fn set_selection(&mut self, selection: Selection) {
        self.selection = selection;
        self.extra_carets.clear();
        self.preferred_column = None;
        self.undo.commit_group();
    }

    /// Caret atual (linha/coluna).
    pub fn caret(&self) -> Result<Caret, DocumentError> {
        Ok(self.buffer.byte_to_caret(self.selection.head)?)
    }

    /// Cursores extras (além do primário), em ordem de inserção.
    #[must_use]
    pub fn extra_carets(&self) -> &[ByteOffset] {
        &self.extra_carets
    }

    /// Todos os heads (primário + extras), ordenados por offset crescente.
    #[must_use]
    pub fn all_caret_offsets(&self) -> Vec<ByteOffset> {
        let mut v = vec![self.selection.head];
        v.extend_from_slice(&self.extra_carets);
        v.sort_by_key(|b| b.as_usize());
        v.dedup();
        v
    }

    /// Limpa todos os cursores secundários.
    pub fn clear_extra_carets(&mut self) {
        self.extra_carets.clear();
    }

    /// Colapsa a seleção ativa em um cursor no head.
    pub fn collapse_selection(&mut self) {
        self.selection = Selection::caret(self.selection.head);
    }

    /// Adiciona um cursor em posição absoluta de byte.
    pub fn add_cursor_at(&mut self, offset: ByteOffset) {
        let clamped = self.buffer.clamp_to_char_boundary(offset);
        self.push_extra_caret(clamped);
    }

    /// Adiciona caret uma linha acima do cursor mais acima.
    pub fn add_cursor_above(&mut self) -> Result<(), DocumentError> {
        let all_carets = self.all_caret_offsets();
        let topmost_offset = all_carets
            .iter()
            .min_by_key(|b| b.as_usize())
            .copied()
            .unwrap_or(self.selection.head);
        let caret = self.buffer.byte_to_caret(topmost_offset)?;
        if caret.line == 0 {
            return Ok(());
        }
        let col = self.preferred_column.unwrap_or(caret.column);
        self.preferred_column = Some(col);
        let target = Caret::new(caret.line - 1, col);
        let off = self.buffer.caret_to_byte(target)?;
        self.push_extra_caret(off);
        Ok(())
    }

    /// Adiciona caret uma linha abaixo do cursor mais abaixo.
    pub fn add_cursor_below(&mut self) -> Result<(), DocumentError> {
        let all_carets = self.all_caret_offsets();
        let bottommost_offset = all_carets
            .iter()
            .max_by_key(|b| b.as_usize())
            .copied()
            .unwrap_or(self.selection.head);
        let caret = self.buffer.byte_to_caret(bottommost_offset)?;
        let last_line = self.buffer.line_count().saturating_sub(1);
        if caret.line >= last_line {
            return Ok(());
        }
        let col = self.preferred_column.unwrap_or(caret.column);
        self.preferred_column = Some(col);
        let target = Caret::new(caret.line + 1, col);
        let off = self.buffer.caret_to_byte(target)?;
        self.push_extra_caret(off);
        Ok(())
    }

    fn push_extra_caret(&mut self, off: ByteOffset) {
        if off == self.selection.head {
            return;
        }
        if !self.extra_carets.contains(&off) {
            self.extra_carets.push(off);
        }
    }

    fn set_carets_from_offsets(&mut self, mut offsets: Vec<ByteOffset>) {
        offsets.sort_by_key(|b| b.as_usize());
        offsets.dedup();
        if let Some((first, rest)) = offsets.split_first() {
            self.selection = Selection::caret(*first);
            self.extra_carets = rest.to_vec();
        }
        self.undo.commit_group();
    }

    fn move_head_to(&mut self, head: ByteOffset, extend: bool) {
        self.selection = self.selection.move_head(head, extend);
        if !extend {
            self.extra_carets.clear();
        }
        self.undo.commit_group();
    }

    /// Move o caret um caractere à esquerda (todos os cursores se multi-cursor).
    pub fn move_left(&mut self, extend: bool) -> Result<(), DocumentError> {
        if !self.extra_carets.is_empty() && !extend {
            let all = self.all_caret_offsets();
            let mut new_offsets = Vec::with_capacity(all.len());
            for off in all {
                let prev = self.buffer.prev_char_offset(off).unwrap_or(off);
                new_offsets.push(prev);
            }
            self.set_carets_from_offsets(new_offsets);
            self.preferred_column = None;
            return Ok(());
        }
        let head = self.buffer.prev_char_offset(self.selection.head)?;
        self.preferred_column = None;
        self.move_head_to(head, extend);
        Ok(())
    }

    /// Move o caret um caractere à direita (todos os cursores se multi-cursor).
    pub fn move_right(&mut self, extend: bool) -> Result<(), DocumentError> {
        if !self.extra_carets.is_empty() && !extend {
            let all = self.all_caret_offsets();
            let mut new_offsets = Vec::with_capacity(all.len());
            for off in all {
                let next = self.buffer.next_char_offset(off).unwrap_or(off);
                new_offsets.push(next);
            }
            self.set_carets_from_offsets(new_offsets);
            self.preferred_column = None;
            return Ok(());
        }
        let head = self.buffer.next_char_offset(self.selection.head)?;
        self.preferred_column = None;
        self.move_head_to(head, extend);
        Ok(())
    }

    /// Move o caret uma linha acima, preservando coluna preferida.
    pub fn move_up(&mut self, extend: bool) -> Result<(), DocumentError> {
        if !self.extra_carets.is_empty() && !extend {
            let all = self.all_caret_offsets();
            let mut new_offsets = Vec::with_capacity(all.len());
            for off in all {
                if let Ok(caret) = self.buffer.byte_to_caret(off) {
                    if caret.line > 0 {
                        let col = self.preferred_column.unwrap_or(caret.column);
                        let target = Caret::new(caret.line - 1, col);
                        if let Ok(new_off) = self.buffer.caret_to_byte(target) {
                            new_offsets.push(new_off);
                            continue;
                        }
                    }
                }
                new_offsets.push(off);
            }
            self.set_carets_from_offsets(new_offsets);
            return Ok(());
        }
        let caret = self.buffer.byte_to_caret(self.selection.head)?;
        if caret.line == 0 {
            return Ok(());
        }
        let col = self.preferred_column.unwrap_or(caret.column);
        self.preferred_column = Some(col);
        let target = Caret::new(caret.line - 1, col);
        let head = self.buffer.caret_to_byte(target)?;
        self.move_head_to(head, extend);
        Ok(())
    }

    /// Move o caret uma linha abaixo, preservando coluna preferida.
    pub fn move_down(&mut self, extend: bool) -> Result<(), DocumentError> {
        if !self.extra_carets.is_empty() && !extend {
            let all = self.all_caret_offsets();
            let last_line = self.buffer.line_count().saturating_sub(1);
            let mut new_offsets = Vec::with_capacity(all.len());
            for off in all {
                if let Ok(caret) = self.buffer.byte_to_caret(off) {
                    if caret.line < last_line {
                        let col = self.preferred_column.unwrap_or(caret.column);
                        let target = Caret::new(caret.line + 1, col);
                        if let Ok(new_off) = self.buffer.caret_to_byte(target) {
                            new_offsets.push(new_off);
                            continue;
                        }
                    }
                }
                new_offsets.push(off);
            }
            self.set_carets_from_offsets(new_offsets);
            return Ok(());
        }
        let caret = self.buffer.byte_to_caret(self.selection.head)?;
        let last_line = self.buffer.line_count().saturating_sub(1);
        if caret.line >= last_line {
            return Ok(());
        }
        let col = self.preferred_column.unwrap_or(caret.column);
        self.preferred_column = Some(col);
        let target = Caret::new(caret.line + 1, col);
        let head = self.buffer.caret_to_byte(target)?;
        self.move_head_to(head, extend);
        Ok(())
    }

    /// Home da linha atual.
    pub fn move_line_start(&mut self, extend: bool) -> Result<(), DocumentError> {
        if !self.extra_carets.is_empty() && !extend {
            let all = self.all_caret_offsets();
            let mut new_offsets = Vec::with_capacity(all.len());
            for off in all {
                if let Ok(caret) = self.buffer.byte_to_caret(off) {
                    if let Ok(new_off) = self.buffer.caret_to_byte(Caret::new(caret.line, 0)) {
                        new_offsets.push(new_off);
                        continue;
                    }
                }
                new_offsets.push(off);
            }
            self.set_carets_from_offsets(new_offsets);
            self.preferred_column = Some(0);
            return Ok(());
        }
        let caret = self.buffer.byte_to_caret(self.selection.head)?;
        let head = self.buffer.caret_to_byte(Caret::new(caret.line, 0))?;
        self.preferred_column = Some(0);
        self.move_head_to(head, extend);
        Ok(())
    }

    /// End da linha atual.
    pub fn move_line_end(&mut self, extend: bool) -> Result<(), DocumentError> {
        if !self.extra_carets.is_empty() && !extend {
            let all = self.all_caret_offsets();
            let mut new_offsets = Vec::with_capacity(all.len());
            for off in all {
                if let Ok(caret) = self.buffer.byte_to_caret(off) {
                    if let Ok(line_text) = self.buffer.line_text(caret.line) {
                        let trimmed = line_text.trim_end_matches(['\n', '\r']);
                        let col = trimmed.chars().count();
                        if let Ok(new_off) = self.buffer.caret_to_byte(Caret::new(caret.line, col))
                        {
                            new_offsets.push(new_off);
                            continue;
                        }
                    }
                }
                new_offsets.push(off);
            }
            self.set_carets_from_offsets(new_offsets);
            return Ok(());
        }
        let caret = self.buffer.byte_to_caret(self.selection.head)?;
        let line = self.buffer.line_text(caret.line)?;
        let col = line.chars().count();
        let head = self.buffer.caret_to_byte(Caret::new(caret.line, col))?;
        self.preferred_column = Some(col);
        self.move_head_to(head, extend);
        Ok(())
    }

    /// Início do buffer (Ctrl+Home).
    pub fn move_buffer_start(&mut self, extend: bool) -> Result<(), DocumentError> {
        self.preferred_column = Some(0);
        self.move_head_to(ByteOffset::new(0), extend);
        Ok(())
    }

    /// Fim do buffer (Ctrl+End).
    pub fn move_buffer_end(&mut self, extend: bool) -> Result<(), DocumentError> {
        let head = ByteOffset::new(self.buffer.len_bytes());
        self.preferred_column = None;
        self.move_head_to(head, extend);
        Ok(())
    }

    /// Seleciona todo o documento (Ctrl+A).
    pub fn select_all(&mut self) {
        let end = ByteOffset::new(self.buffer.len_bytes());
        self.selection = Selection::new(ByteOffset::new(0), end);
        self.extra_carets.clear();
        self.preferred_column = None;
        self.undo.commit_group();
    }

    /// Título para tab: nome do arquivo ou "untitled".
    #[must_use]
    pub fn tab_title(&self) -> String {
        match &self.path {
            Some(p) => p
                .file_name()
                .map(|n| n.to_string_lossy().into_owned())
                .unwrap_or_else(|| p.display().to_string()),
            None => "untitled".into(),
        }
    }

    /// Insere texto na posição do caret (substitui seleção se houver).
    /// Com multi-cursor: insere em todos os carets (do fim para o início).
    pub fn insert_text(&mut self, text: &str) -> Result<(), DocumentError> {
        if !self.extra_carets.is_empty() && self.selection.is_empty() {
            return self.insert_text_multi(text);
        }
        if !self.selection.is_empty() {
            self.delete_selection()?;
        }
        let at = self.selection.head;
        self.buffer.insert(at, text)?;
        self.undo.push_applied(Edit::Insert {
            at,
            text: text.to_string(),
        });
        let new_head = ByteOffset::new(at.as_usize() + text.len());
        self.selection = Selection::caret(new_head);
        self.mark_modified();
        self.commit_newline_boundary(text);
        Ok(())
    }

    /// Enter fecha o grupo de undo aberto.
    ///
    /// Sem isso, digitar, pressionar Enter e digitar de novo coalesce num único
    /// grupo, e um `Ctrl+Z` apaga as duas digitações e a quebra de linha. Todo
    /// editor sério cria um ponto de undo no Enter, e o usuário conta com isso.
    fn commit_newline_boundary(&mut self, inserted: &str) {
        if inserted.contains('\n') {
            self.undo.commit_group();
        }
    }

    fn mark_modified(&mut self) {
        self.dirty = true;
        self.version = self.version.wrapping_add(1);
    }

    fn insert_text_multi(&mut self, text: &str) -> Result<(), DocumentError> {
        let carets = self.all_caret_offsets();
        let step = text.len();
        let new_offsets: Vec<ByteOffset> = carets
            .iter()
            .enumerate()
            .map(|(i, at)| ByteOffset::new(at.as_usize() + (i + 1) * step))
            .collect();

        // Insere do maior offset para o menor para manter índices válidos
        for at in carets.into_iter().rev() {
            self.buffer.insert(at, text)?;
            self.undo.push_applied(Edit::Insert {
                at,
                text: text.to_string(),
            });
        }
        if let Some((first, rest)) = new_offsets.split_first() {
            self.selection = Selection::caret(*first);
            self.extra_carets = rest.to_vec();
        }
        self.mark_modified();
        self.commit_newline_boundary(text);
        Ok(())
    }

    /// Apaga a seleção atual, ou o caractere anterior se vazia (backspace).
    pub fn backspace(&mut self) -> Result<(), DocumentError> {
        if !self.selection.is_empty() {
            return self.delete_selection();
        }
        if !self.extra_carets.is_empty() {
            return self.backspace_multi();
        }
        let end = self.selection.head;
        if end.as_usize() == 0 {
            return Ok(());
        }
        let start = self.buffer.prev_char_offset(end)?;
        let removed = self.buffer.delete_range(start, end)?;
        self.undo.push_applied(Edit::Delete {
            at: start,
            text: removed,
        });
        self.selection = Selection::caret(start);
        self.preferred_column = None;
        self.mark_modified();
        Ok(())
    }

    fn backspace_multi(&mut self) -> Result<(), DocumentError> {
        let carets = self.all_caret_offsets();
        let mut ops = Vec::with_capacity(carets.len());
        for end in carets {
            if end.as_usize() == 0 {
                ops.push((end, end, 0usize));
            } else {
                let start = self.buffer.prev_char_offset(end)?;
                let len = end.as_usize().saturating_sub(start.as_usize());
                ops.push((start, end, len));
            }
        }

        let mut new_offsets = Vec::with_capacity(ops.len());
        let mut cumulative_deleted = 0usize;
        for (start, _end, len) in &ops {
            let final_offset = start.as_usize().saturating_sub(cumulative_deleted);
            new_offsets.push(ByteOffset::new(final_offset));
            cumulative_deleted += len;
        }

        for (start, end, len) in ops.into_iter().rev() {
            if len > 0 {
                let removed = self.buffer.delete_range(start, end)?;
                self.undo.push_applied(Edit::Delete {
                    at: start,
                    text: removed,
                });
            }
        }

        new_offsets.sort_by_key(|b| b.as_usize());
        new_offsets.dedup();
        if let Some((first, rest)) = new_offsets.split_first() {
            self.selection = Selection::caret(*first);
            self.extra_carets = rest.to_vec();
        }
        self.preferred_column = None;
        self.mark_modified();
        Ok(())
    }

    /// Delete à frente do caret (ou a seleção).
    pub fn delete_forward(&mut self) -> Result<(), DocumentError> {
        if !self.selection.is_empty() {
            return self.delete_selection();
        }
        if !self.extra_carets.is_empty() {
            return self.delete_forward_multi();
        }
        let start = self.selection.head;
        if start.as_usize() >= self.buffer.len_bytes() {
            return Ok(());
        }
        let end = self.buffer.next_char_offset(start)?;
        let removed = self.buffer.delete_range(start, end)?;
        self.undo.push_applied(Edit::Delete {
            at: start,
            text: removed,
        });
        self.preferred_column = None;
        self.mark_modified();
        Ok(())
    }

    fn delete_forward_multi(&mut self) -> Result<(), DocumentError> {
        let carets = self.all_caret_offsets();
        let buf_len = self.buffer.len_bytes();
        let mut ops = Vec::with_capacity(carets.len());
        for start in carets {
            if start.as_usize() >= buf_len {
                ops.push((start, start, 0usize));
            } else {
                let end = self.buffer.next_char_offset(start)?;
                let len = end.as_usize().saturating_sub(start.as_usize());
                ops.push((start, end, len));
            }
        }

        let mut new_offsets = Vec::with_capacity(ops.len());
        let mut cumulative_deleted = 0usize;
        for (start, _end, len) in &ops {
            let final_offset = start.as_usize().saturating_sub(cumulative_deleted);
            new_offsets.push(ByteOffset::new(final_offset));
            cumulative_deleted += len;
        }

        for (start, end, len) in ops.into_iter().rev() {
            if len > 0 {
                let removed = self.buffer.delete_range(start, end)?;
                self.undo.push_applied(Edit::Delete {
                    at: start,
                    text: removed,
                });
            }
        }

        new_offsets.sort_by_key(|b| b.as_usize());
        new_offsets.dedup();
        if let Some((first, rest)) = new_offsets.split_first() {
            self.selection = Selection::caret(*first);
            self.extra_carets = rest.to_vec();
        }
        self.preferred_column = None;
        self.mark_modified();
        Ok(())
    }

    pub fn delete_selection(&mut self) -> Result<(), DocumentError> {
        if self.selection.is_empty() {
            return Ok(());
        }
        let start = self.selection.start();
        let end = self.selection.end();
        let removed = self.buffer.delete_range(start, end)?;
        self.undo.push_applied(Edit::Delete {
            at: start,
            text: removed,
        });
        self.selection = Selection::caret(start);
        self.extra_carets.clear();
        self.mark_modified();
        Ok(())
    }

    pub fn undo(&mut self) -> bool {
        let changed = self.undo.undo(&mut self.buffer);
        if changed {
            self.mark_modified();
            let len = self.buffer.len_bytes();
            let head = self.selection.head.as_usize().min(len);
            self.selection = Selection::caret(ByteOffset::new(head));
        }
        changed
    }

    /// Resumos do histórico de undo (UI tier B).
    #[must_use]
    pub fn undo_history_labels(&self) -> Vec<String> {
        self.undo.undo_summaries()
    }

    #[must_use]
    pub fn redo_history_labels(&self) -> Vec<String> {
        self.undo.redo_summaries()
    }

    pub fn redo(&mut self) -> bool {
        let changed = self.redo_inner();
        if changed {
            self.mark_modified();
            let len = self.buffer.len_bytes();
            let head = self.selection.head.as_usize().min(len);
            self.selection = Selection::caret(ByteOffset::new(head));
        }
        changed
    }

    fn redo_inner(&mut self) -> bool {
        self.undo.redo(&mut self.buffer)
    }

    pub fn commit_edit_group(&mut self) {
        self.undo.commit_group();
    }

    /// Marca limpo após save bem-sucedido.
    pub fn mark_saved(&mut self) {
        self.undo.commit_group();
        self.dirty = false;
    }

    pub fn set_path(&mut self, path: PathBuf) {
        self.path = Some(path);
    }

    /// Recarrega o conteúdo do path no disco (descarta mudanças locais).
    pub fn reload_from_disk(&mut self) -> Result<(), DocumentError> {
        let path = self.path.clone().ok_or_else(|| {
            DocumentError::Io(std::io::Error::new(
                std::io::ErrorKind::InvalidInput,
                "document has no path",
            ))
        })?;
        let text = std::fs::read_to_string(&path)?;
        self.buffer = Buffer::from_text(&text);
        let len = self.buffer.len_bytes();
        let head = self.selection.head.as_usize().min(len);
        self.selection = Selection::caret(ByteOffset::new(head));
        self.extra_carets.clear();
        self.dirty = false;
        self.preferred_column = None;
        self.undo = UndoStack::new();
        self.version = self.version.wrapping_add(1);
        Ok(())
    }

    /// Substitui o texto inteiro do buffer (ex.: format LSP), preservando path.
    pub fn replace_full_text(&mut self, text: &str) -> Result<(), DocumentError> {
        let old = self.buffer.as_string();
        if old == text {
            return Ok(());
        }
        // delete all + insert
        let end = ByteOffset::new(self.buffer.len_bytes());
        if end.as_usize() > 0 {
            let removed = self.buffer.delete_range(ByteOffset::new(0), end)?;
            self.undo.push_applied(Edit::Delete {
                at: ByteOffset::new(0),
                text: removed,
            });
        }
        self.buffer.insert(ByteOffset::new(0), text)?;
        self.undo.push_applied(Edit::Insert {
            at: ByteOffset::new(0),
            text: text.to_string(),
        });
        self.selection = Selection::caret(ByteOffset::new(0));
        self.preferred_column = None;
        self.mark_modified();
        Ok(())
    }

    /// Serializa o buffer para disco no path atual ou `path` fornecido.
    pub fn save_to(&mut self, path: Option<&Path>) -> Result<(), DocumentError> {
        let target = match path {
            Some(p) => p.to_path_buf(),
            None => self.path.clone().ok_or_else(|| {
                DocumentError::Io(std::io::Error::new(
                    std::io::ErrorKind::InvalidInput,
                    "document has no path; provide one to save",
                ))
            })?,
        };
        if let Some(parent) = target.parent() {
            if !parent.as_os_str().is_empty() && !parent.exists() {
                std::fs::create_dir_all(parent)?;
            }
        }
        std::fs::write(&target, self.buffer.as_string())?;
        self.path = Some(target);
        self.mark_saved();
        Ok(())
    }

    /// Texto da seleção (vazio se caret).
    #[must_use]
    pub fn selected_text(&self) -> String {
        if self.selection.is_empty() {
            return String::new();
        }
        self.buffer
            .text_range(self.selection.start(), self.selection.end())
            .unwrap_or_default()
    }

    /// Move o caret para `byte` (seleção colapsada), ajustando para boundary UTF-8.
    pub fn jump_to_byte(&mut self, byte: ByteOffset) {
        let snapped = self.buffer.clamp_to_char_boundary(byte);
        self.selection = Selection::caret(snapped);
        self.extra_carets.clear();
        self.preferred_column = None;
        self.undo.commit_group();
    }

    /// Seleciona o intervalo `[start, end)` em bytes e coloca o head no fim, ajustando para boundaries UTF-8.
    pub fn select_byte_range(&mut self, start: ByteOffset, end: ByteOffset) {
        let snapped_start = self.buffer.clamp_to_char_boundary(start);
        let snapped_end = self.buffer.clamp_to_char_boundary(end);
        self.selection = Selection::new(snapped_start, snapped_end);
        self.preferred_column = None;
        self.undo.commit_group();
    }

    /// Atualiza seleção sem fechar grupo de undo (drag do mouse), ajustando para boundaries UTF-8.
    pub fn set_selection_live(&mut self, anchor: ByteOffset, head: ByteOffset) {
        let snapped_anchor = self.buffer.clamp_to_char_boundary(anchor);
        let snapped_head = self.buffer.clamp_to_char_boundary(head);
        self.selection = Selection::new(snapped_anchor, snapped_head);
        self.extra_carets.clear();
        self.preferred_column = None;
        // não commit_group — drag gera dezenas de updates por segundo
    }
}

/// Coleção de documentos abertos + tab ativa.
#[derive(Debug, Default)]
pub struct DocumentStore {
    next_id: u64,
    docs: Vec<Document>,
    active: Option<DocumentId>,
}

impl DocumentStore {
    #[must_use]
    pub fn new() -> Self {
        Self::default()
    }

    fn alloc_id(&mut self) -> DocumentId {
        let id = DocumentId(self.next_id);
        self.next_id += 1;
        id
    }

    /// Buffer vazio sem path.
    pub fn open_empty(&mut self) -> DocumentId {
        let id = self.alloc_id();
        self.docs.push(Document {
            id,
            path: None,
            buffer: Buffer::new(),
            selection: Selection::default(),
            extra_carets: Vec::new(),
            undo: UndoStack::new(),
            dirty: false,
            preferred_column: None,
            version: 0,
        });
        self.active = Some(id);
        id
    }

    /// Abre arquivo do disco (UTF-8). Se já estiver aberto, só ativa a tab.
    pub fn open_path(&mut self, path: impl AsRef<Path>) -> Result<DocumentId, DocumentError> {
        let path = path.as_ref();
        let canonical = std::fs::canonicalize(path).unwrap_or_else(|_| path.to_path_buf());
        if let Some(existing) = self.docs.iter().find(|d| {
            d.path
                .as_ref()
                .is_some_and(|p| p == &canonical || p == path)
        }) {
            let id = existing.id;
            self.active = Some(id);
            return Ok(id);
        }
        let text = std::fs::read_to_string(path)?;
        let id = self.alloc_id();
        self.docs.push(Document {
            id,
            path: Some(canonical),
            buffer: Buffer::from_text(&text),
            selection: Selection::default(),
            extra_carets: Vec::new(),
            undo: UndoStack::new(),
            dirty: false,
            preferred_column: None,
            version: 0,
        });
        self.active = Some(id);
        Ok(id)
    }

    /// Tab seguinte (cíclico).
    pub fn activate_next_tab(&mut self) -> Option<DocumentId> {
        let ids = self.tab_ids();
        if ids.is_empty() {
            return None;
        }
        let cur = self.active?;
        let idx = ids.iter().position(|id| *id == cur).unwrap_or(0);
        let next = ids[(idx + 1) % ids.len()];
        self.active = Some(next);
        Some(next)
    }

    /// Tab anterior (cíclico).
    pub fn activate_prev_tab(&mut self) -> Option<DocumentId> {
        let ids = self.tab_ids();
        if ids.is_empty() {
            return None;
        }
        let cur = self.active?;
        let idx = ids.iter().position(|id| *id == cur).unwrap_or(0);
        let prev = ids[(idx + ids.len() - 1) % ids.len()];
        self.active = Some(prev);
        Some(prev)
    }

    /// Metadados leves para a barra de tabs.
    #[must_use]
    pub fn tab_summaries(&self) -> Vec<TabSummary> {
        self.docs
            .iter()
            .map(|d| TabSummary {
                id: d.id,
                title: d.tab_title(),
                dirty: d.dirty,
                active: Some(d.id) == self.active,
            })
            .collect()
    }

    #[must_use]
    pub fn active_id(&self) -> Option<DocumentId> {
        self.active
    }

    pub fn set_active(&mut self, id: DocumentId) -> Result<(), DocumentError> {
        if !self.docs.iter().any(|d| d.id == id) {
            return Err(DocumentError::NotFound(id));
        }
        self.active = Some(id);
        Ok(())
    }

    #[must_use]
    pub fn get(&self, id: DocumentId) -> Option<&Document> {
        self.docs.iter().find(|d| d.id == id)
    }

    pub fn get_mut(&mut self, id: DocumentId) -> Option<&mut Document> {
        self.docs.iter_mut().find(|d| d.id == id)
    }

    pub fn active(&self) -> Result<&Document, DocumentError> {
        let id = self.active.ok_or(DocumentError::NoActiveDocument)?;
        self.get(id).ok_or(DocumentError::NotFound(id))
    }

    pub fn active_mut(&mut self) -> Result<&mut Document, DocumentError> {
        let id = self.active.ok_or(DocumentError::NoActiveDocument)?;
        // borrow checker: find index
        let idx = self
            .docs
            .iter()
            .position(|d| d.id == id)
            .ok_or(DocumentError::NotFound(id))?;
        Ok(&mut self.docs[idx])
    }

    /// Ordem das tabs.
    #[must_use]
    pub fn tab_ids(&self) -> Vec<DocumentId> {
        self.docs.iter().map(|d| d.id).collect()
    }

    #[must_use]
    pub fn dirty_count(&self) -> usize {
        self.docs.iter().filter(|document| document.dirty).count()
    }

    /// Localiza uma tab pelo path, normalizando o path do evento quando possível.
    #[must_use]
    pub fn document_id_for_path(&self, path: &Path) -> Option<DocumentId> {
        let canonical = std::fs::canonicalize(path).ok();
        self.docs
            .iter()
            .find(|document| {
                document.path.as_deref().is_some_and(|document_path| {
                    document_path == path
                        || canonical
                            .as_deref()
                            .is_some_and(|canonical_path| document_path == canonical_path)
                })
            })
            .map(|document| document.id)
    }

    /// Fecha tab. Retorna se o doc estava dirty (chamador deve confirmar).
    pub fn close(&mut self, id: DocumentId) -> Result<bool, DocumentError> {
        let idx = self
            .docs
            .iter()
            .position(|d| d.id == id)
            .ok_or(DocumentError::NotFound(id))?;
        let dirty = self.docs[idx].dirty;
        self.docs.remove(idx);
        if self.active == Some(id) {
            self.active = self.docs.last().map(|d| d.id);
        }
        Ok(dirty)
    }

    /// Salva todos os documentos com path sem ocultar falhas parciais.
    pub fn save_all(&mut self) -> SaveAllReport {
        let mut report = SaveAllReport::default();
        let ids: Vec<_> = self.docs.iter().map(|d| d.id).collect();
        for id in ids {
            if let Some(doc) = self.get_mut(id) {
                if doc.path().is_none() {
                    report.skipped_without_path += 1;
                    continue;
                }
                let path = doc.path().map(Path::to_path_buf).unwrap_or_default();
                match doc.save_to(None) {
                    Ok(()) => {
                        report.saved += 1;
                        report.saved_paths.push(path);
                    }
                    Err(error) => report.failures.push(SaveFailure { path, error }),
                }
            }
        }
        report
    }

    /// Paths de abas abertas (ordem das tabs).
    #[must_use]
    pub fn open_paths(&self) -> Vec<PathBuf> {
        self.docs
            .iter()
            .filter_map(|d| d.path().map(Path::to_path_buf))
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn insert_and_undo_through_document() {
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        {
            let doc = store.get_mut(id).unwrap();
            doc.insert_text("abc").unwrap();
            assert_eq!(doc.buffer().as_string(), "abc");
            assert!(doc.is_dirty());
            doc.commit_edit_group();
            assert!(doc.undo());
            assert_eq!(doc.buffer().as_string(), "");
        }
    }

    #[test]
    fn backspace_utf8() {
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        let doc = store.get_mut(id).unwrap();
        doc.insert_text("a✨b").unwrap();
        doc.backspace().unwrap();
        assert_eq!(doc.buffer().as_string(), "a✨");
        doc.backspace().unwrap();
        assert_eq!(doc.buffer().as_string(), "a");
    }

    #[test]
    fn multi_tab_active() {
        let mut store = DocumentStore::new();
        let a = store.open_empty();
        let b = store.open_empty();
        assert_eq!(store.active_id(), Some(b));
        store.set_active(a).unwrap();
        assert_eq!(store.active_id(), Some(a));
        assert_eq!(store.tab_ids().len(), 2);
    }

    #[test]
    fn dirty_count_includes_inactive_tabs() {
        let mut store = DocumentStore::new();
        let dirty = store.open_empty();
        store
            .get_mut(dirty)
            .unwrap()
            .insert_text("changed")
            .unwrap();
        store.open_empty();

        assert_eq!(store.dirty_count(), 1);
        assert!(!store.active().unwrap().is_dirty());
    }

    #[test]
    fn save_all_reports_write_failures() {
        let directory = tempfile::tempdir().unwrap();
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        let document = store.get_mut(id).unwrap();
        document.set_path(directory.path().to_path_buf());
        document
            .insert_text("cannot overwrite a directory")
            .unwrap();

        let report = store.save_all();

        assert_eq!(report.saved, 0);
        assert_eq!(report.failures.len(), 1);
        assert!(store.get(id).unwrap().is_dirty());
    }

    #[test]
    fn move_up_down_preserves_preferred_column() {
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        let doc = store.get_mut(id).unwrap();
        doc.insert_text("hello\nxy\nworld").unwrap();
        doc.move_line_end(false).unwrap();
        assert_eq!(doc.caret().unwrap().column, 5);
        doc.move_up(false).unwrap();
        assert_eq!(doc.caret().unwrap(), Caret::new(1, 2)); // "xy" só tem 2 cols
        doc.move_up(false).unwrap();
        assert_eq!(doc.caret().unwrap(), Caret::new(0, 5));
    }
}

#[cfg(test)]
mod multi_cursor_tests {
    use super::*;

    #[test]
    fn insert_at_two_carets() {
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        let doc = store.get_mut(id).unwrap();
        doc.insert_text("ab\ncd").unwrap();
        doc.jump_to_byte(ByteOffset::new(0));
        doc.add_cursor_below().unwrap();
        assert_eq!(doc.extra_carets().len(), 1);
        doc.insert_text("XY").unwrap();
        assert_eq!(doc.buffer().as_string(), "XYab\nXYcd");
        assert_eq!(doc.selection().head, ByteOffset::new(2));
        assert_eq!(doc.extra_carets(), &[ByteOffset::new(7)]);
    }

    #[test]
    fn backspace_at_two_carets() {
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        let doc = store.get_mut(id).unwrap();
        doc.insert_text("ab\ncd").unwrap();
        // C1 at end of line 0 ("ab" -> offset 2), C2 at end of line 1 ("cd" -> offset 5)
        doc.jump_to_byte(ByteOffset::new(2));
        doc.add_cursor_below().unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(2), ByteOffset::new(5)]
        );

        doc.backspace().unwrap();
        // Deletes 'b' (offset 1..2) and 'd' (offset 4..5)
        assert_eq!(doc.buffer().as_string(), "a\nc");
        assert_eq!(doc.selection().head, ByteOffset::new(1));
        assert_eq!(doc.extra_carets(), &[ByteOffset::new(3)]);
    }

    #[test]
    fn delete_forward_at_two_carets() {
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        let doc = store.get_mut(id).unwrap();
        doc.insert_text("ab\ncd").unwrap();
        // C1 at 0 ('a'), C2 at 3 ('c')
        doc.jump_to_byte(ByteOffset::new(0));
        doc.add_cursor_below().unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(0), ByteOffset::new(3)]
        );

        doc.delete_forward().unwrap();
        // Deletes 'a' (offset 0..1) and 'c' (offset 3..4)
        assert_eq!(doc.buffer().as_string(), "b\nd");
        assert_eq!(doc.selection().head, ByteOffset::new(0));
        assert_eq!(doc.extra_carets(), &[ByteOffset::new(2)]);
    }

    #[test]
    fn move_all_carets_in_sync() {
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        let doc = store.get_mut(id).unwrap();
        doc.insert_text("hello\nworld").unwrap();
        doc.jump_to_byte(ByteOffset::new(0));
        doc.add_cursor_below().unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(0), ByteOffset::new(6)]
        );

        // Move right: both carets advance 1 char
        doc.move_right(false).unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(1), ByteOffset::new(7)]
        );

        // Move left: both carets retreat 1 char
        doc.move_left(false).unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(0), ByteOffset::new(6)]
        );

        // Move line end: both carets move to end of their respective line
        doc.move_line_end(false).unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(5), ByteOffset::new(11)]
        );

        // Move line start: both carets move to start of line
        doc.move_line_start(false).unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(0), ByteOffset::new(6)]
        );

        // Clear extra carets
        doc.clear_extra_carets();
        assert_eq!(doc.extra_carets().len(), 0);
        assert_eq!(doc.selection().head, ByteOffset::new(0));
    }

    #[test]
    fn multi_cursor_merges_on_collision() {
        let mut store = DocumentStore::new();
        let id = store.open_empty();
        let doc = store.get_mut(id).unwrap();
        doc.insert_text("ab").unwrap();
        doc.jump_to_byte(ByteOffset::new(0));
        doc.add_cursor_at(ByteOffset::new(1));
        assert_eq!(doc.all_caret_offsets().len(), 2);

        // Move first cursor right by 1, while second is at 1 -> collision at offset 1 or 2
        doc.move_right(false).unwrap();
        // C1 was 0 -> 1, C2 was 1 -> 2. Still 2 carets:
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(1), ByteOffset::new(2)]
        );

        // Moving right again: C1 -> 2, C2 was at 2 -> 2. They collide at 2!
        doc.move_right(false).unwrap();
        assert_eq!(doc.all_caret_offsets(), vec![ByteOffset::new(2)]);
        assert_eq!(doc.extra_carets().len(), 0);
    }
}
