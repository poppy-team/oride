//! Preview Markdown → linhas semânticas (sem ratatui).
//!
//! O UI mapeia `PreviewStyle` para cores. Não é HTML; é “ANSI-like” em TUI.
//! Imagens viram **placeholders** (alt + path + existe?); sem bitmap.

use std::path::{Path, PathBuf};

/// Estilo de um segmento de preview.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum PreviewStyle {
    Normal,
    Heading(u8),
    Bold,
    Italic,
    Code,
    Link,
    Quote,
    ListMarker,
    Hr,
    FenceLang,
    /// Moldura do card de imagem.
    Image,
    /// Texto alt da imagem.
    ImageAlt,
    /// Path / URL da imagem.
    ImagePath,
    /// Arquivo local encontrado.
    ImageOk,
    /// Arquivo local ausente.
    ImageMissing,
    /// Célula / linha de tabela.
    Table,
    /// ~~riscado~~
    Strike,
    /// Texto secundário / dica.
    Dim,
    /// Sintaxe destacada via gramática.
    Syntax(crate::HighlightKind),
}

/// Link detectado no preview com intervalo de colunas lógicas na linha.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PreviewLink {
    pub start_col: usize,
    pub end_col: usize,
    pub url: String,
}

/// Uma linha do painel de preview.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PreviewLine {
    pub segments: Vec<(String, PreviewStyle)>,
    pub links: Vec<PreviewLink>,
}

impl PreviewLine {
    pub fn plain(s: impl Into<String>) -> Self {
        Self {
            segments: vec![(s.into(), PreviewStyle::Normal)],
            links: Vec::new(),
        }
    }

    pub fn styled(s: impl Into<String>, style: PreviewStyle) -> Self {
        Self {
            segments: vec![(s.into(), style)],
            links: Vec::new(),
        }
    }

    pub fn empty() -> Self {
        Self {
            segments: vec![(" ".into(), PreviewStyle::Normal)],
            links: Vec::new(),
        }
    }

    pub fn multi(segments: Vec<(String, PreviewStyle)>) -> Self {
        if segments.is_empty() {
            Self::empty()
        } else {
            Self {
                segments,
                links: Vec::new(),
            }
        }
    }

    #[must_use]
    pub fn with_links(mut self, links: Vec<PreviewLink>) -> Self {
        self.links = links;
        self
    }

    #[must_use]
    pub fn link_at_col(&self, col: usize) -> Option<&PreviewLink> {
        self.links
            .iter()
            .find(|l| col >= l.start_col && col <= l.end_col)
    }

    #[must_use]
    pub fn first_link(&self) -> Option<&PreviewLink> {
        self.links.first()
    }
}

/// Metadados extraídos de cabeçalho de imagem (PNG, JPEG, GIF, SVG, WebP).
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ImageMetadata {
    pub format: &'static str,
    pub dimensions: Option<(u32, u32)>,
    pub file_size_bytes: u64,
}

/// Capacidade de renderização gráfica do terminal corrente.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum TerminalGraphicsCapability {
    None,
    Kitty,
    Sixel,
    Iterm2,
}

/// As variáveis de ambiente que a detecção consulta.
///
/// Um valor em vez de `std::env::var` direto: a detecção passa a ser uma função
/// pura sobre este tipo, testável sem tocar no ambiente do processo — que é
/// estado global compartilhado por todos os testes que rodam em paralelo, e
/// escrever nele é o que torna um teste instável sob carga.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct TerminalEnvironment {
    vars: Vec<(String, String)>,
}

impl TerminalEnvironment {
    /// Lê o ambiente do processo corrente.
    #[must_use]
    pub fn from_process() -> Self {
        Self {
            vars: std::env::vars().collect(),
        }
    }

    /// Constrói um ambiente a partir de pares explícitos.
    #[must_use]
    pub fn from_pairs(pairs: &[(&str, &str)]) -> Self {
        Self {
            vars: pairs
                .iter()
                .map(|(k, v)| ((*k).to_string(), (*v).to_string()))
                .collect(),
        }
    }

    /// Consulta uma variável.
    #[must_use]
    pub fn get(&self, key: &str) -> Option<&str> {
        self.vars
            .iter()
            .find(|(name, _)| name == key)
            .map(|(_, value)| value.as_str())
    }
}

/// Detecta a compatibilidade do emulador de terminal com protocolos de imagem.
///
/// Lê o ambiente do processo. Um chamador que já conheça a capacidade — o app,
/// que a resolve uma vez — deve usar [`TerminalGraphicsCapability::detect`]
/// diretamente, para não reler o ambiente a cada quadro nem impedir um teste de
/// injetá-la.
#[must_use]
pub fn detect_terminal_graphics() -> TerminalGraphicsCapability {
    TerminalGraphicsCapability::detect(&TerminalEnvironment::from_process())
}

impl TerminalGraphicsCapability {
    /// Detecta a capacidade a partir de um ambiente.
    #[must_use]
    pub fn detect(env: &TerminalEnvironment) -> Self {
        if env.get("KITTY_WINDOW_ID").is_some()
            || env.get("GHOSTTY_RESOURCES_DIR").is_some()
            || env.get("WEZTERM_PANE").is_some()
        {
            return Self::Kitty;
        }
        if let Some(term) = env.get("TERM") {
            let term_lower = term.to_ascii_lowercase();
            if term_lower.contains("kitty") || term_lower.contains("ghostty") {
                return Self::Kitty;
            }
            if term_lower.contains("sixel") || term_lower.contains("foot") {
                return Self::Sixel;
            }
            if term_lower.contains("iterm") {
                return Self::Iterm2;
            }
        }
        if let Some(term_prog) = env.get("TERM_PROGRAM") {
            let prog_lower = term_prog.to_ascii_lowercase();
            if prog_lower.contains("wezterm")
                || prog_lower.contains("ghostty")
                || prog_lower.contains("kitty")
            {
                return Self::Kitty;
            }
            if prog_lower.contains("iterm") {
                return Self::Iterm2;
            }
        }
        Self::None
    }
}

/// Formata contagem de bytes para string legível (B, KB, MB).
#[must_use]
pub fn format_file_size(bytes: u64) -> String {
    if bytes < 1024 {
        format!("{bytes} B")
    } else if bytes < 1024 * 1024 {
        format!("{:.1} KB", bytes as f64 / 1024.0)
    } else {
        format!("{:.1} MB", bytes as f64 / (1024.0 * 1024.0))
    }
}

/// Inspeciona o cabeçalho de um arquivo local de imagem sem dependência de bibliotecas externas pesadas.
#[must_use]
pub fn inspect_image_file(path: &Path) -> Option<ImageMetadata> {
    let file_size_bytes = std::fs::metadata(path).ok()?.len();
    let file = std::fs::File::open(path).ok()?;
    use std::io::Read;
    let mut header = [0u8; 1024];
    let mut reader = std::io::BufReader::new(file);
    let bytes_read = reader.read(&mut header).ok()?;
    let bytes = &header[..bytes_read];

    // PNG: \x89PNG\r\n\x1a\n seguido de IHDR
    if bytes.starts_with(b"\x89PNG\r\n\x1a\n") && bytes.len() >= 24 {
        let width = u32::from_be_bytes(bytes[16..20].try_into().ok()?);
        let height = u32::from_be_bytes(bytes[20..24].try_into().ok()?);
        return Some(ImageMetadata {
            format: "PNG",
            dimensions: Some((width, height)),
            file_size_bytes,
        });
    }

    // GIF: GIF87a ou GIF89a
    if (bytes.starts_with(b"GIF87a") || bytes.starts_with(b"GIF89a")) && bytes.len() >= 10 {
        let width = u16::from_le_bytes(bytes[6..8].try_into().ok()?) as u32;
        let height = u16::from_le_bytes(bytes[8..10].try_into().ok()?) as u32;
        return Some(ImageMetadata {
            format: "GIF",
            dimensions: Some((width, height)),
            file_size_bytes,
        });
    }

    // JPEG: \xFF\xD8
    if bytes.starts_with(b"\xFF\xD8") {
        let mut dimensions = None;
        let mut i = 2;
        while i + 4 <= bytes.len() {
            if bytes[i] != 0xFF {
                i += 1;
                continue;
            }
            let marker = bytes[i + 1];
            if matches!(marker, 0xC0..=0xC3 | 0xC5..=0xC7 | 0xC9..=0xCB | 0xCD..=0xCF)
                && i + 9 <= bytes.len()
            {
                let height = u16::from_be_bytes(bytes[i + 5..i + 7].try_into().ok()?) as u32;
                let width = u16::from_be_bytes(bytes[i + 7..i + 9].try_into().ok()?) as u32;
                dimensions = Some((width, height));
                break;
            }
            let segment_len = u16::from_be_bytes(bytes[i + 2..i + 4].try_into().ok()?) as usize;
            i += 2 + segment_len;
        }
        return Some(ImageMetadata {
            format: "JPEG",
            dimensions,
            file_size_bytes,
        });
    }

    // WebP: RIFF .... WEBP
    if bytes.starts_with(b"RIFF") && bytes.len() >= 30 && &bytes[8..12] == b"WEBP" {
        let mut dimensions = None;
        if &bytes[12..16] == b"VP8 " && bytes.len() >= 30 {
            if bytes[23..26] == [0x9d, 0x01, 0x2a] {
                let width = (u16::from_le_bytes(bytes[26..28].try_into().ok()?) & 0x3fff) as u32;
                let height = (u16::from_le_bytes(bytes[28..30].try_into().ok()?) & 0x3fff) as u32;
                dimensions = Some((width, height));
            }
        } else if &bytes[12..16] == b"VP8L" && bytes.len() >= 25 {
            if bytes[20] == 0x2f {
                let b1 = bytes[21] as u32;
                let b2 = bytes[22] as u32;
                let b3 = bytes[23] as u32;
                let b4 = bytes[24] as u32;
                let width = 1 + (((b2 & 0x3f) << 8) | b1);
                let height = 1 + (((b4 & 0x0f) << 10) | (b3 << 2) | ((b2 & 0xc0) >> 6));
                dimensions = Some((width, height));
            }
        } else if &bytes[12..16] == b"VP8X" && bytes.len() >= 30 {
            let width =
                1 + (bytes[24] as u32 | ((bytes[25] as u32) << 8) | ((bytes[26] as u32) << 16));
            let height =
                1 + (bytes[27] as u32 | ((bytes[28] as u32) << 8) | ((bytes[29] as u32) << 16));
            dimensions = Some((width, height));
        }
        return Some(ImageMetadata {
            format: "WebP",
            dimensions,
            file_size_bytes,
        });
    }

    // SVG
    if let Ok(text) = std::str::from_utf8(bytes) {
        if text.contains("<svg") {
            return Some(ImageMetadata {
                format: "SVG",
                dimensions: None,
                file_size_bytes,
            });
        }
    }

    None
}

/// Renderiza Markdown simples (sem base path — imagens sem check de disco).
#[must_use]
pub fn render_preview_lines(source: &str) -> Vec<PreviewLine> {
    render_preview_lines_in(source, None)
}

/// Como [`render_preview_lines`], resolvendo paths de imagem relativos a `base_dir`
/// (normalmente o diretório do arquivo `.md` aberto).
#[must_use]
pub fn render_preview_lines_in(source: &str, base_dir: Option<&Path>) -> Vec<PreviewLine> {
    render_preview_lines_with_config(source, base_dir, false)
}

/// Renderiza Markdown com configuração de suporte a imagens de terminal.
///
/// A capacidade gráfica vem do ambiente do processo. Um chamador que já a
/// conheça usa [`render_preview_lines_with_graphics`].
#[must_use]
pub fn render_preview_lines_with_config(
    source: &str,
    base_dir: Option<&Path>,
    terminal_images: bool,
) -> Vec<PreviewLine> {
    render_preview_lines_with_graphics(
        source,
        base_dir,
        terminal_images,
        detect_terminal_graphics(),
    )
}

/// Renderiza Markdown com a capacidade gráfica fornecida pelo chamador.
///
/// É esta a porta que um teste deve usar: injetar a capacidade evita escrever no
/// ambiente do processo, que é global e compartilhado.
#[must_use]
pub fn render_preview_lines_with_graphics(
    source: &str,
    base_dir: Option<&Path>,
    terminal_images: bool,
    graphics: TerminalGraphicsCapability,
) -> Vec<PreviewLine> {
    let mut out = Vec::new();
    let lines: Vec<&str> = source.lines().collect();
    let mut idx = 0usize;

    // frontmatter YAML simples no topo
    if lines.first().map(|l| l.trim() == "---").unwrap_or(false) {
        out.push(PreviewLine::styled("── frontmatter ──", PreviewStyle::Dim));
        idx = 1;
        while idx < lines.len() {
            if lines[idx].trim() == "---" {
                out.push(PreviewLine::styled("────────", PreviewStyle::Hr));
                idx += 1;
                break;
            }
            out.push(PreviewLine::styled(
                format!("  {}", lines[idx]),
                PreviewStyle::Dim,
            ));
            idx += 1;
        }
    }

    while idx < lines.len() {
        let line = lines[idx];

        // fences
        if let Some(rest) = line.strip_prefix("```") {
            let fence_lang = rest.trim();
            let label = if fence_lang.is_empty() {
                "code".to_string()
            } else {
                fence_lang.to_string()
            };
            out.push(PreviewLine::styled(
                format!("┌ {label}"),
                PreviewStyle::FenceLang,
            ));
            idx += 1;

            let lang_id = crate::markdown::fence_lang_alias(fence_lang);
            let mut code_lines = Vec::new();
            while idx < lines.len() {
                if lines[idx].trim_start().starts_with("```") {
                    idx += 1;
                    break;
                }
                code_lines.push(lines[idx]);
                idx += 1;
            }

            if let Some(lang) = lang_id {
                let code_block = code_lines.join("\n");
                let highlights = crate::highlight::highlight_language_slice(lang, &code_block, 0);
                let mut line_offset = 0;
                for code_line in code_lines {
                    let mut segs = vec![("│ ".to_string(), PreviewStyle::FenceLang)];
                    let spans = crate::highlight::line_spans(code_line, line_offset, &highlights);
                    if spans.is_empty() {
                        segs.push((code_line.to_string(), PreviewStyle::Code));
                    } else {
                        for (token, kind) in spans {
                            let style = if kind == crate::HighlightKind::Normal {
                                PreviewStyle::Code
                            } else {
                                PreviewStyle::Syntax(kind)
                            };
                            segs.push((token.to_string(), style));
                        }
                    }
                    out.push(PreviewLine::multi(segs));
                    line_offset += code_line.len() + 1;
                }
            } else {
                for code_line in code_lines {
                    out.push(PreviewLine::styled(
                        format!("│ {code_line}"),
                        PreviewStyle::Code,
                    ));
                }
            }
            out.push(PreviewLine::styled("└───", PreviewStyle::Hr));
            continue;
        }

        // setext heading: Title\n=== or ---
        if idx + 1 < lines.len() {
            let next = lines[idx + 1].trim();
            if !line.trim().is_empty()
                && next.len() >= 2
                && (next.chars().all(|c| c == '=') || next.chars().all(|c| c == '-'))
            {
                let level = if next.starts_with('=') { 1u8 } else { 2u8 };
                out.push(PreviewLine::styled(
                    line.trim().to_string(),
                    PreviewStyle::Heading(level),
                ));
                if level == 1 {
                    out.push(PreviewLine::styled("════════", PreviewStyle::Heading(1)));
                } else {
                    out.push(PreviewLine::styled("────────", PreviewStyle::Heading(2)));
                }
                idx += 2;
                continue;
            }
        }

        // HR (só se não for setext — já tratado)
        let t = line.trim();
        if is_hr(t) {
            out.push(PreviewLine::styled("────────────", PreviewStyle::Hr));
            idx += 1;
            continue;
        }

        // headings ATX
        if let Some((level, text)) = parse_atx_heading(line) {
            let mark = match level {
                1 => "█ ",
                2 => "▓ ",
                3 => "▒ ",
                _ => "· ",
            };
            let (inlines, links) = inline_segments_and_links(text, base_dir, false);
            let mark_len = mark.chars().count();
            let shifted_links = links
                .into_iter()
                .map(|mut l| {
                    l.start_col += mark_len;
                    l.end_col += mark_len;
                    l
                })
                .collect();
            let mut segs = vec![(mark.into(), PreviewStyle::Heading(level))];
            segs.extend(inlines);
            out.push(PreviewLine::multi(segs).with_links(shifted_links));
            if level <= 2 {
                out.push(PreviewLine::empty());
            }
            idx += 1;
            continue;
        }

        // blockquote (com inline)
        if let Some(rest) = line.trim_start().strip_prefix("> ") {
            let (inlines, links) = inline_segments_and_links(rest, base_dir, false);
            let shifted_links = links
                .into_iter()
                .map(|mut l| {
                    l.start_col += 2;
                    l.end_col += 2;
                    l
                })
                .collect();
            let mut segs = vec![("│ ".into(), PreviewStyle::Quote)];
            for mut s in inlines {
                if matches!(s.1, PreviewStyle::Normal) {
                    s.1 = PreviewStyle::Quote;
                }
                segs.push(s);
            }
            out.push(PreviewLine::multi(segs).with_links(shifted_links));
            idx += 1;
            continue;
        }
        if line.trim_start() == ">" {
            out.push(PreviewLine::styled("│", PreviewStyle::Quote));
            idx += 1;
            continue;
        }

        // table block
        if is_table_row(line) {
            let mut table_lines = Vec::new();
            while idx < lines.len() && is_table_row(lines[idx]) {
                table_lines.push(lines[idx]);
                idx += 1;
            }
            out.extend(format_table_block(&table_lines, base_dir));
            continue;
        }

        // task list
        if let Some((done, rest)) = strip_task(line) {
            let mark = if done { " ☑ " } else { " ☐ " };
            let (inlines, links) = inline_segments_and_links(rest, base_dir, false);
            let shifted_links = links
                .into_iter()
                .map(|mut l| {
                    l.start_col += 3;
                    l.end_col += 3;
                    l
                })
                .collect();
            let mut segs = vec![(mark.into(), PreviewStyle::ListMarker)];
            segs.extend(inlines);
            out.push(PreviewLine::multi(segs).with_links(shifted_links));
            idx += 1;
            continue;
        }

        // list unordered
        if let Some(rest) = strip_ul(line) {
            let (inlines, links) = inline_segments_and_links(rest, base_dir, false);
            let shifted_links = links
                .into_iter()
                .map(|mut l| {
                    l.start_col += 3;
                    l.end_col += 3;
                    l
                })
                .collect();
            let mut segs = vec![(" • ".into(), PreviewStyle::ListMarker)];
            segs.extend(inlines);
            out.push(PreviewLine::multi(segs).with_links(shifted_links));
            idx += 1;
            continue;
        }

        // list ordered
        if let Some((num, rest)) = strip_ol(line) {
            let prefix = format!(" {num}. ");
            let prefix_len = prefix.chars().count();
            let (inlines, links) = inline_segments_and_links(rest, base_dir, false);
            let shifted_links = links
                .into_iter()
                .map(|mut l| {
                    l.start_col += prefix_len;
                    l.end_col += prefix_len;
                    l
                })
                .collect();
            let mut segs = vec![(prefix, PreviewStyle::ListMarker)];
            segs.extend(inlines);
            out.push(PreviewLine::multi(segs).with_links(shifted_links));
            idx += 1;
            continue;
        }

        // empty
        if line.trim().is_empty() {
            out.push(PreviewLine::empty());
            idx += 1;
            continue;
        }

        // linha só com imagem → card multi-linha
        if let Some((alt, url)) = parse_standalone_image(line) {
            out.extend(image_card(&alt, &url, base_dir, terminal_images, graphics));
            idx += 1;
            continue;
        }

        // parágrafo com inline (imagens inline → card compacto no fluxo)
        let (segs, links) = inline_segments_and_links(line, base_dir, true);
        out.push(PreviewLine::multi(segs).with_links(links));
        idx += 1;
    }

    if out.is_empty() {
        out.push(PreviewLine::plain("(vazio)"));
    }
    out
}

fn is_hr(t: &str) -> bool {
    matches!(t, "---" | "***" | "___" | "* * *" | "- - -")
        || (t.len() >= 3
            && t.chars()
                .all(|c| c == '-' || c == '*' || c == '_' || c == ' ')
            && t.chars().filter(|c| *c != ' ').count() >= 3
            && !t.contains('|'))
}

fn is_table_row(line: &str) -> bool {
    let t = line.trim();
    t.starts_with('|') && t.matches('|').count() >= 2
}

fn is_table_sep(line: &str) -> bool {
    let t = line.trim().trim_matches('|');
    !t.is_empty()
        && t.chars()
            .all(|c| c == '-' || c == ':' || c == '|' || c == ' ')
}

#[derive(Clone, Copy, PartialEq, Eq)]
enum TableAlignment {
    Left,
    Center,
    Right,
}

fn parse_table_alignment(cell: &str) -> TableAlignment {
    let trimmed = cell.trim();
    let starts = trimmed.starts_with(':');
    let ends = trimmed.ends_with(':');
    if starts && ends {
        TableAlignment::Center
    } else if ends {
        TableAlignment::Right
    } else {
        TableAlignment::Left
    }
}

fn parse_table_cells(line: &str) -> Vec<String> {
    let t = line.trim();
    let content = t.strip_prefix('|').unwrap_or(t);
    let content = content.strip_suffix('|').unwrap_or(content);
    content.split('|').map(|c| c.trim().to_string()).collect()
}

fn format_table_block(table_lines: &[&str], base_dir: Option<&Path>) -> Vec<PreviewLine> {
    if table_lines.is_empty() {
        return Vec::new();
    }
    let mut raw_rows: Vec<Vec<String>> = Vec::new();
    let mut sep_index = None;

    for (i, line) in table_lines.iter().enumerate() {
        if is_table_sep(line) {
            if sep_index.is_none() {
                sep_index = Some(i);
            }
        } else {
            raw_rows.push(parse_table_cells(line));
        }
    }

    if raw_rows.is_empty() {
        return Vec::new();
    }

    let num_cols = raw_rows.iter().map(|r| r.len()).max().unwrap_or(0);
    if num_cols == 0 {
        return Vec::new();
    }

    let mut alignments = vec![TableAlignment::Left; num_cols];
    if let Some(sep_idx) = sep_index {
        let sep_cells = parse_table_cells(table_lines[sep_idx]);
        for (col, cell) in sep_cells.iter().enumerate().take(num_cols) {
            alignments[col] = parse_table_alignment(cell);
        }
    }

    for row in &mut raw_rows {
        while row.len() < num_cols {
            row.push(String::new());
        }
    }

    let mut col_widths = vec![3usize; num_cols];
    for row in &raw_rows {
        for (col, cell) in row.iter().enumerate() {
            let char_count = cell.chars().count();
            if char_count > col_widths[col] {
                col_widths[col] = char_count;
            }
        }
    }

    let mut out = Vec::new();

    // Top border: ┌───┬───┐
    let mut top_border = String::from("  ┌");
    for (col, &w) in col_widths.iter().enumerate() {
        top_border.push_str(&"─".repeat(w + 2));
        if col + 1 < num_cols {
            top_border.push('┬');
        } else {
            top_border.push('┐');
        }
    }
    out.push(PreviewLine::styled(top_border, PreviewStyle::Dim));

    let has_header = sep_index.is_some() && !raw_rows.is_empty();

    for (row_idx, row) in raw_rows.iter().enumerate() {
        let mut row_segs = Vec::new();
        let mut row_links = Vec::new();
        row_segs.push(("  │".to_string(), PreviewStyle::Dim));
        let mut current_col = 3;

        for (col, cell) in row.iter().enumerate() {
            let w = col_widths[col];
            let cell_count = cell.chars().count();
            let pad = w.saturating_sub(cell_count);
            let (left_pad, right_pad) = match alignments[col] {
                TableAlignment::Left => (1, pad + 1),
                TableAlignment::Right => (pad + 1, 1),
                TableAlignment::Center => {
                    let half = pad / 2;
                    (half + 1, pad - half + 1)
                }
            };
            if left_pad > 0 {
                row_segs.push((" ".repeat(left_pad), PreviewStyle::Table));
                current_col += left_pad;
            }
            let (inlines, links) = inline_segments_and_links(cell, base_dir, false);
            for mut link in links {
                link.start_col += current_col;
                link.end_col += current_col;
                row_links.push(link);
            }
            for (text, style) in inlines {
                let text_len = text.chars().count();
                current_col += text_len;
                let cell_style = if style == PreviewStyle::Normal {
                    PreviewStyle::Table
                } else {
                    style
                };
                row_segs.push((text, cell_style));
            }
            if right_pad > 0 {
                row_segs.push((" ".repeat(right_pad), PreviewStyle::Table));
                current_col += right_pad;
            }
            row_segs.push(("│".to_string(), PreviewStyle::Dim));
            current_col += 1;
        }
        out.push(PreviewLine::multi(row_segs).with_links(row_links));

        // Header separator: ├───┼───┤
        if has_header && row_idx == 0 && raw_rows.len() > 1 {
            let mut mid_border = String::from("  ├");
            for (col, &w) in col_widths.iter().enumerate() {
                mid_border.push_str(&"─".repeat(w + 2));
                if col + 1 < num_cols {
                    mid_border.push('┼');
                } else {
                    mid_border.push('┤');
                }
            }
            out.push(PreviewLine::styled(mid_border, PreviewStyle::Dim));
        }
    }

    // Bottom border: └───┴───┘
    let mut bot_border = String::from("  └");
    for (col, &w) in col_widths.iter().enumerate() {
        bot_border.push_str(&"─".repeat(w + 2));
        if col + 1 < num_cols {
            bot_border.push('┴');
        } else {
            bot_border.push('┘');
        }
    }
    out.push(PreviewLine::styled(bot_border, PreviewStyle::Dim));

    out
}

fn parse_atx_heading(line: &str) -> Option<(u8, &str)> {
    let trimmed = line.trim_start();
    let mut level = 0u8;
    let bytes = trimmed.as_bytes();
    while (level as usize) < bytes.len() && bytes[level as usize] == b'#' && level < 6 {
        level += 1;
    }
    if level == 0 {
        return None;
    }
    if (level as usize) < bytes.len() && bytes[level as usize] != b' ' {
        if (level as usize) == bytes.len() {
            return Some((level, ""));
        }
        return None;
    }
    let text = trimmed[level as usize..].trim_start();
    Some((level, text))
}

fn strip_task(line: &str) -> Option<(bool, &str)> {
    let t = line.trim_start();
    for (p, done) in [
        ("- [ ] ", false),
        ("- [x] ", true),
        ("- [X] ", true),
        ("* [ ] ", false),
        ("* [x] ", true),
        ("* [X] ", true),
        ("+ [ ] ", false),
        ("+ [x] ", true),
    ] {
        if let Some(r) = t.strip_prefix(p) {
            return Some((done, r));
        }
    }
    None
}

fn strip_ul(line: &str) -> Option<&str> {
    let t = line.trim_start();
    for p in ["- ", "* ", "+ "] {
        if let Some(r) = t.strip_prefix(p) {
            // task lists handled elsewhere
            if r.starts_with("[ ]") || r.starts_with("[x]") || r.starts_with("[X]") {
                return None;
            }
            return Some(r);
        }
    }
    None
}

fn strip_ol(line: &str) -> Option<(&str, &str)> {
    let t = line.trim_start();
    let mut i = 0;
    let b = t.as_bytes();
    while i < b.len() && b[i].is_ascii_digit() {
        i += 1;
    }
    if i == 0 || i >= b.len() {
        return None;
    }
    if b[i] != b'.' {
        return None;
    }
    if i + 1 < b.len() && b[i + 1] == b' ' {
        return Some((&t[..i], &t[i + 2..]));
    }
    None
}

/// Linha que é só `![alt](url)` (espaços ok).
fn parse_standalone_image(line: &str) -> Option<(String, String)> {
    let t = line.trim();
    if !t.starts_with("![") {
        return None;
    }
    let chars: Vec<char> = t.chars().collect();
    let (alt, url, next) = parse_image(&chars, 0)?;
    if next == chars.len() {
        Some((alt, url))
    } else {
        None
    }
}

fn image_card(
    alt: &str,
    url: &str,
    base_dir: Option<&Path>,
    terminal_images: bool,
    graphics: TerminalGraphicsCapability,
) -> Vec<PreviewLine> {
    let alt_show = if alt.is_empty() {
        "(sem texto alt)"
    } else {
        alt
    };
    let (path_line, status_style, status_note, meta_note) =
        describe_image_target(url, base_dir, terminal_images, graphics);

    let mut lines = vec![
        PreviewLine::styled("┌ 🖼  imagem", PreviewStyle::Image),
        PreviewLine::multi(vec![
            ("│  ".into(), PreviewStyle::Image),
            (alt_show.into(), PreviewStyle::ImageAlt),
        ]),
        PreviewLine::multi(vec![
            ("│  ".into(), PreviewStyle::Image),
            (path_line, PreviewStyle::ImagePath),
        ])
        .with_links(vec![PreviewLink {
            start_col: 0,
            end_col: usize::MAX,
            url: url.to_string(),
        }]),
    ];

    if let Some(meta) = meta_note {
        lines.push(PreviewLine::multi(vec![
            ("│  ".into(), PreviewStyle::Image),
            (meta, PreviewStyle::Dim),
        ]));
    }

    lines.push(PreviewLine::multi(vec![
        ("│  ".into(), PreviewStyle::Image),
        (status_note, status_style),
    ]));
    lines.push(PreviewLine::styled("└", PreviewStyle::Image));
    lines
}

fn describe_image_target(
    url: &str,
    base_dir: Option<&Path>,
    terminal_images: bool,
    graphics: TerminalGraphicsCapability,
) -> (String, PreviewStyle, String, Option<String>) {
    let url = url.trim();
    if url.is_empty() {
        return (
            "(sem path)".into(),
            PreviewStyle::ImageMissing,
            "⚠ path vazio".into(),
            None,
        );
    }
    if url.starts_with("http://") || url.starts_with("https://") || url.starts_with("data:") {
        return (
            truncate_mid(url, 48),
            PreviewStyle::Dim,
            "🔗 URL remota · não embutida no TUI".into(),
            None,
        );
    }
    // local path
    let path = PathBuf::from(url);
    let resolved = if path.is_absolute() {
        path
    } else if let Some(base) = base_dir {
        base.join(&path)
    } else {
        path
    };
    let display = truncate_mid(&resolved.display().to_string(), 48);
    if resolved.is_file() {
        let meta = inspect_image_file(&resolved);
        let meta_str = meta.as_ref().map(|m| {
            let size = format_file_size(m.file_size_bytes);
            if let Some((w, h)) = m.dimensions {
                format!("{} · {}x{} px · {}", m.format, w, h, size)
            } else {
                format!("{} · {}", m.format, size)
            }
        });

        let cap = graphics;
        let status = if terminal_images {
            match cap {
                TerminalGraphicsCapability::Kitty => {
                    "✓ arquivo local · Kitty Graphics ativo".into()
                }
                TerminalGraphicsCapability::Sixel => {
                    "✓ arquivo local · Sixel Graphics ativo".into()
                }
                TerminalGraphicsCapability::Iterm2 => {
                    "✓ arquivo local · iTerm2 Protocol ativo".into()
                }
                TerminalGraphicsCapability::None => {
                    "✓ arquivo local encontrado (terminal sem suporte gráfico)".into()
                }
            }
        } else {
            "✓ arquivo local encontrado".into()
        };

        (display, PreviewStyle::ImageOk, status, meta_str)
    } else if resolved.exists() {
        (
            display,
            PreviewStyle::ImageMissing,
            "⚠ path existe mas não é arquivo".into(),
            None,
        )
    } else {
        (
            display,
            PreviewStyle::ImageMissing,
            "✗ arquivo não encontrado".into(),
            None,
        )
    }
}

fn truncate_mid(s: &str, max: usize) -> String {
    let n = s.chars().count();
    if n <= max {
        return s.to_string();
    }
    let keep = max.saturating_sub(1) / 2;
    let left: String = s.chars().take(keep).collect();
    let right: String = s.chars().skip(n - keep).collect();
    format!("{left}…{right}")
}

/// Inline: code, bold, italic, strike, image, link.
/// Retorna segmentos visuais e metadados de links com intervalo de colunas lógicas.
fn inline_segments_and_links(
    text: &str,
    base_dir: Option<&Path>,
    expand_images: bool,
) -> (Vec<(String, PreviewStyle)>, Vec<PreviewLink>) {
    let _ = base_dir; // reservado para tooltips futuros
    let mut out = Vec::new();
    let mut links = Vec::new();
    let chars: Vec<char> = text.chars().collect();
    let mut i = 0;
    let mut buf = String::new();
    let mut current_col = 0usize;

    let flush = |buf: &mut String,
                 out: &mut Vec<(String, PreviewStyle)>,
                 style: PreviewStyle,
                 col: &mut usize| {
        if !buf.is_empty() {
            let s = std::mem::take(buf);
            *col += s.chars().count();
            out.push((s, style));
        }
    };

    while i < chars.len() {
        // image ![alt](url) — antes de link
        if chars[i] == '!' && chars.get(i + 1) == Some(&'[') {
            if let Some((alt, url, next)) = parse_image(&chars, i) {
                flush(&mut buf, &mut out, PreviewStyle::Normal, &mut current_col);
                let link_start = current_col;
                if expand_images {
                    let label = if alt.is_empty() {
                        "🖼".into()
                    } else {
                        format!("🖼 {alt}")
                    };
                    current_col += label.chars().count();
                    out.push((label, PreviewStyle::Image));
                    if !url.is_empty() {
                        let note = format!(" ({})", truncate_mid(&url, 24));
                        current_col += note.chars().count();
                        out.push((note, PreviewStyle::Dim));
                    }
                } else {
                    let label = format!("🖼 {}", if alt.is_empty() { "img" } else { &alt });
                    current_col += label.chars().count();
                    out.push((label, PreviewStyle::Image));
                }
                if !url.is_empty() {
                    links.push(PreviewLink {
                        start_col: link_start,
                        end_col: current_col,
                        url,
                    });
                }
                i = next;
                continue;
            }
        }

        // code `...`
        if chars[i] == '`' {
            flush(&mut buf, &mut out, PreviewStyle::Normal, &mut current_col);
            i += 1;
            let start = i;
            while i < chars.len() && chars[i] != '`' {
                i += 1;
            }
            let code: String = chars[start..i].iter().collect();
            let seg = format!(" {code} ");
            current_col += seg.chars().count();
            out.push((seg, PreviewStyle::Code));
            if i < chars.len() {
                i += 1;
            }
            continue;
        }

        // ~~strike~~
        if i + 1 < chars.len() && chars[i] == '~' && chars[i + 1] == '~' {
            flush(&mut buf, &mut out, PreviewStyle::Normal, &mut current_col);
            i += 2;
            let start = i;
            while i + 1 < chars.len() && !(chars[i] == '~' && chars[i + 1] == '~') {
                i += 1;
            }
            let s: String = chars[start..i].iter().collect();
            current_col += s.chars().count();
            out.push((s, PreviewStyle::Strike));
            if i + 1 < chars.len() {
                i += 2;
            }
            continue;
        }

        // **bold** or __bold__
        if i + 1 < chars.len()
            && ((chars[i] == '*' && chars[i + 1] == '*')
                || (chars[i] == '_' && chars[i + 1] == '_'))
        {
            let mark = chars[i];
            flush(&mut buf, &mut out, PreviewStyle::Normal, &mut current_col);
            i += 2;
            let start = i;
            while i + 1 < chars.len() && !(chars[i] == mark && chars[i + 1] == mark) {
                i += 1;
            }
            let bold: String = chars[start..i].iter().collect();
            current_col += bold.chars().count();
            out.push((bold, PreviewStyle::Bold));
            if i + 1 < chars.len() {
                i += 2;
            }
            continue;
        }

        // *italic* or _italic_ (não confundir com __)
        if (chars[i] == '*' || chars[i] == '_') && chars.get(i + 1) != Some(&chars[i]) {
            let mark = chars[i];
            // _word_ mid-word: skip if alnum before
            if mark == '_' && i > 0 && chars[i - 1].is_alphanumeric() {
                buf.push(chars[i]);
                i += 1;
                continue;
            }
            flush(&mut buf, &mut out, PreviewStyle::Normal, &mut current_col);
            i += 1;
            let start = i;
            while i < chars.len() && chars[i] != mark {
                i += 1;
            }
            let it: String = chars[start..i].iter().collect();
            current_col += it.chars().count();
            out.push((it, PreviewStyle::Italic));
            if i < chars.len() {
                i += 1;
            }
            continue;
        }

        // [text](url) link
        if chars[i] == '[' {
            if let Some((label, url, next)) = parse_link(&chars, i) {
                flush(&mut buf, &mut out, PreviewStyle::Normal, &mut current_col);
                let link_start = current_col;
                current_col += label.chars().count();
                out.push((label, PreviewStyle::Link));
                if !url.is_empty() {
                    let arrow = format!(" → {}", truncate_mid(&url, 28));
                    current_col += arrow.chars().count();
                    out.push((arrow, PreviewStyle::Dim));
                }
                let link_end = current_col;
                links.push(PreviewLink {
                    start_col: link_start,
                    end_col: link_end,
                    url,
                });
                i = next;
                continue;
            }
        }

        // autolink <http...>
        if chars[i] == '<' {
            if let Some((url, next)) = parse_autolink(&chars, i) {
                flush(&mut buf, &mut out, PreviewStyle::Normal, &mut current_col);
                let link_start = current_col;
                current_col += url.chars().count();
                out.push((url.clone(), PreviewStyle::Link));
                links.push(PreviewLink {
                    start_col: link_start,
                    end_col: current_col,
                    url,
                });
                i = next;
                continue;
            }
        }

        buf.push(chars[i]);
        i += 1;
    }
    flush(&mut buf, &mut out, PreviewStyle::Normal, &mut current_col);
    if out.is_empty() {
        out.push((String::new(), PreviewStyle::Normal));
    }
    (out, links)
}

/// Inline: code, bold, italic, strike, image, link.
/// `expand_images`: se true, imagem vira `🖼 alt` compacto (não card).
pub fn inline_segments(
    text: &str,
    base_dir: Option<&Path>,
    expand_images: bool,
) -> Vec<(String, PreviewStyle)> {
    inline_segments_and_links(text, base_dir, expand_images).0
}

fn parse_image(chars: &[char], start: usize) -> Option<(String, String, usize)> {
    // ![alt](url)
    if chars.get(start) != Some(&'!') || chars.get(start + 1) != Some(&'[') {
        return None;
    }
    let mut i = start + 2;
    let alt_start = i;
    while i < chars.len() && chars[i] != ']' {
        i += 1;
    }
    if i >= chars.len() {
        return None;
    }
    let alt: String = chars[alt_start..i].iter().collect();
    i += 1;
    if chars.get(i) != Some(&'(') {
        return None;
    }
    i += 1;
    let url_start = i;
    while i < chars.len() && chars[i] != ')' {
        i += 1;
    }
    if i >= chars.len() {
        return None;
    }
    let url: String = chars[url_start..i].iter().collect();
    i += 1;
    Some((alt, url.trim().to_string(), i))
}

fn parse_link(chars: &[char], start: usize) -> Option<(String, String, usize)> {
    if chars.get(start) != Some(&'[') {
        return None;
    }
    let mut i = start + 1;
    let label_start = i;
    let mut depth = 1i32;
    while i < chars.len() {
        if chars[i] == '[' {
            depth += 1;
        } else if chars[i] == ']' {
            depth -= 1;
            if depth == 0 {
                break;
            }
        }
        i += 1;
    }
    if i >= chars.len() || depth != 0 {
        return None;
    }
    let label: String = chars[label_start..i].iter().collect();
    i += 1;
    if chars.get(i) != Some(&'(') {
        return None;
    }
    i += 1;
    let url_start = i;
    while i < chars.len() && chars[i] != ')' {
        i += 1;
    }
    if i >= chars.len() {
        return None;
    }
    let url: String = chars[url_start..i].iter().collect();
    i += 1;
    Some((label, url.trim().to_string(), i))
}

fn parse_autolink(chars: &[char], start: usize) -> Option<(String, usize)> {
    if chars.get(start) != Some(&'<') {
        return None;
    }
    let mut i = start + 1;
    let s = i;
    while i < chars.len() && chars[i] != '>' {
        i += 1;
    }
    if i >= chars.len() {
        return None;
    }
    let url: String = chars[s..i].iter().collect();
    if !(url.starts_with("http://") || url.starts_with("https://") || url.starts_with("mailto:")) {
        return None;
    }
    Some((url, i + 1))
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    #[test]
    fn heading_and_list() {
        let lines = render_preview_lines("# Title\n\n- item **bold**\n");
        assert!(lines[0]
            .segments
            .iter()
            .any(|(_, s)| matches!(s, PreviewStyle::Heading(1))));
        assert!(lines.iter().any(|l| {
            l.segments
                .iter()
                .any(|(t, s)| t.contains("item") || matches!(s, PreviewStyle::ListMarker))
        }));
    }

    #[test]
    fn fence_block() {
        let lines = render_preview_lines("```oris\nfn main() {}\n```\n");
        assert!(lines
            .iter()
            .any(|l| l.segments.iter().any(|(_, s)| *s == PreviewStyle::Code)));
        assert!(lines.iter().any(|l| {
            l.segments
                .iter()
                .any(|(t, s)| *s == PreviewStyle::FenceLang && t.contains("oris"))
        }));
    }

    #[test]
    fn inline_code() {
        let lines = render_preview_lines("use `code` here\n");
        assert!(lines[0]
            .segments
            .iter()
            .any(|(t, s)| t.contains("code") && *s == PreviewStyle::Code));
    }

    #[test]
    fn image_placeholder_card() {
        let lines = render_preview_lines("![Diagrama](./arch.png)\n");
        assert!(
            lines.iter().any(|l| {
                l.segments
                    .iter()
                    .any(|(t, s)| *s == PreviewStyle::Image && t.contains("imagem"))
            }),
            "{lines:?}"
        );
        assert!(lines.iter().any(|l| {
            l.segments
                .iter()
                .any(|(t, s)| *s == PreviewStyle::ImageAlt && t.contains("Diagrama"))
        }));
        assert!(lines.iter().any(|l| {
            l.segments.iter().any(|(_, s)| {
                matches!(
                    s,
                    PreviewStyle::ImageMissing | PreviewStyle::ImageOk | PreviewStyle::Dim
                )
            })
        }));
    }

    #[test]
    fn image_local_exists() {
        let dir = tempfile::tempdir().unwrap();
        let img = dir.path().join("pic.png");
        std::fs::File::create(&img)
            .unwrap()
            .write_all(b"fake")
            .unwrap();
        let md = dir.path().join("doc.md");
        std::fs::write(&md, "![x](pic.png)\n").unwrap();
        let lines = render_preview_lines_in("![x](pic.png)\n", Some(dir.path()));
        assert!(
            lines
                .iter()
                .any(|l| { l.segments.iter().any(|(_, s)| *s == PreviewStyle::ImageOk) }),
            "{lines:?}"
        );
    }

    #[test]
    fn remote_image_note() {
        let lines = render_preview_lines("![logo](https://example.com/a.png)\n");
        assert!(lines.iter().any(|l| {
            l.segments
                .iter()
                .any(|(t, _)| t.contains("URL remota") || t.contains("https://"))
        }));
    }

    #[test]
    fn task_list_and_table() {
        let src = "- [x] done\n- [ ] todo\n\n| a | b |\n| --- | --- |\n| 1 | 2 |\n";
        let lines = render_preview_lines(src);
        assert!(lines.iter().any(|l| {
            l.segments
                .iter()
                .any(|(t, _)| t.contains('☑') || t.contains('☐'))
        }));
        assert!(lines
            .iter()
            .any(|l| l.segments.iter().any(|(_, s)| *s == PreviewStyle::Table)));
    }

    #[test]
    fn link_and_strike() {
        let lines = render_preview_lines("see [docs](./a.md) and ~~old~~\n");
        assert!(lines[0]
            .segments
            .iter()
            .any(|(t, s)| t == "docs" && *s == PreviewStyle::Link));
        assert!(lines[0]
            .segments
            .iter()
            .any(|(t, s)| t == "old" && *s == PreviewStyle::Strike));
    }

    #[test]
    fn inline_image_compact() {
        let lines = render_preview_lines("antes ![a](b.png) depois\n");
        assert!(lines[0]
            .segments
            .iter()
            .any(|(t, s)| *s == PreviewStyle::Image && t.contains('🖼')));
    }

    #[test]
    fn box_drawing_table_with_alignments() {
        let src = "| Col1 | Col2 |\n| :--- | ---: |\n| left | right |\n";
        let lines = render_preview_lines(src);
        assert!(lines.iter().any(|l| l
            .segments
            .iter()
            .any(|(t, _)| t.contains('┌') && t.contains('┬') && t.contains('┐'))));
        assert!(lines.iter().any(|l| l
            .segments
            .iter()
            .any(|(t, _)| t.contains('├') && t.contains('┼') && t.contains('┤'))));
        assert!(lines.iter().any(|l| l
            .segments
            .iter()
            .any(|(t, _)| t.contains('└') && t.contains('┴') && t.contains('┘'))));
        assert!(lines.iter().any(|l| l
            .segments
            .iter()
            .any(|(t, s)| t.contains("left") && *s == PreviewStyle::Table)));
    }

    #[test]
    fn links_register_logical_columns() {
        let lines = render_preview_lines("Click [here](https://example.com) for details\n");
        assert_eq!(lines[0].links.len(), 1);
        let link = &lines[0].links[0];
        assert_eq!(link.url, "https://example.com");
        assert!(link.start_col > 0);
        assert!(link.end_col > link.start_col);
        assert_eq!(lines[0].link_at_col(link.start_col), Some(link));
    }

    #[test]
    fn fenced_code_highlights_syntax() {
        let lines = render_preview_lines("```rust\nfn main() {\n    let x = 42;\n}\n```\n");
        assert!(lines.iter().any(|l| {
            l.segments
                .iter()
                .any(|(_, s)| matches!(s, PreviewStyle::Syntax(_)))
        }));
    }

    #[test]
    fn inspect_png_and_gif_headers() {
        let temp_dir = tempfile::tempdir().unwrap();

        // 1. PNG Header (width 640, height 480)
        let png_path = temp_dir.path().join("test.png");
        let mut png_data = vec![0x89, b'P', b'N', b'G', b'\r', b'\n', 0x1a, b'\n'];
        png_data.extend_from_slice(&[0, 0, 0, 13]); // chunk len
        png_data.extend_from_slice(b"IHDR");
        png_data.extend_from_slice(&640u32.to_be_bytes());
        png_data.extend_from_slice(&480u32.to_be_bytes());
        png_data.extend_from_slice(&[8, 6, 0, 0, 0]);
        std::fs::write(&png_path, &png_data).unwrap();

        let meta_png = inspect_image_file(&png_path).expect("PNG metadata should be parsed");
        assert_eq!(meta_png.format, "PNG");
        assert_eq!(meta_png.dimensions, Some((640, 480)));

        // 2. GIF Header (width 128, height 64)
        let gif_path = temp_dir.path().join("test.gif");
        let mut gif_data = b"GIF89a".to_vec();
        gif_data.extend_from_slice(&128u16.to_le_bytes());
        gif_data.extend_from_slice(&64u16.to_le_bytes());
        std::fs::write(&gif_path, &gif_data).unwrap();

        let meta_gif = inspect_image_file(&gif_path).expect("GIF metadata should be parsed");
        assert_eq!(meta_gif.format, "GIF");
        assert_eq!(meta_gif.dimensions, Some((128, 64)));

        // 3. Render Preview with image card metadata
        let preview = render_preview_lines_in("![Diagrama](test.png)\n", Some(temp_dir.path()));
        assert!(preview
            .iter()
            .any(|l| l.segments.iter().any(|(t, s)| t.contains("PNG")
                && t.contains("640x480 px")
                && *s == PreviewStyle::Dim)));
    }
}

#[cfg(test)]
mod graphics_tests {
    use super::{TerminalEnvironment, TerminalGraphicsCapability};

    fn detect(pairs: &[(&str, &str)]) -> TerminalGraphicsCapability {
        TerminalGraphicsCapability::detect(&TerminalEnvironment::from_pairs(pairs))
    }

    #[test]
    fn a_bare_environment_has_no_graphics() {
        assert_eq!(detect(&[]), TerminalGraphicsCapability::None);
    }

    #[test]
    fn marker_variables_imply_kitty() {
        // Cada uma destas é a marca de um emulador que fala o protocolo.
        for key in ["KITTY_WINDOW_ID", "GHOSTTY_RESOURCES_DIR", "WEZTERM_PANE"] {
            assert_eq!(
                detect(&[(key, "1")]),
                TerminalGraphicsCapability::Kitty,
                "{key} deveria implicar Kitty"
            );
        }
    }

    #[test]
    fn term_names_each_protocol() {
        assert_eq!(
            detect(&[("TERM", "xterm-kitty")]),
            TerminalGraphicsCapability::Kitty
        );
        assert_eq!(
            detect(&[("TERM", "xterm-ghostty")]),
            TerminalGraphicsCapability::Kitty
        );
        assert_eq!(
            detect(&[("TERM", "foot")]),
            TerminalGraphicsCapability::Sixel
        );
        assert_eq!(
            detect(&[("TERM", "xterm-sixel")]),
            TerminalGraphicsCapability::Sixel
        );
        assert_eq!(
            detect(&[("TERM", "xterm-256color")]),
            TerminalGraphicsCapability::None
        );
    }

    #[test]
    fn term_matching_ignores_case() {
        assert_eq!(
            detect(&[("TERM", "XTERM-KITTY")]),
            TerminalGraphicsCapability::Kitty
        );
    }

    #[test]
    fn term_program_is_consulted_after_term() {
        assert_eq!(
            detect(&[("TERM", "xterm-256color"), ("TERM_PROGRAM", "WezTerm")]),
            TerminalGraphicsCapability::Kitty
        );
        assert_eq!(
            detect(&[("TERM", "xterm-256color"), ("TERM_PROGRAM", "iTerm.app")]),
            TerminalGraphicsCapability::Iterm2
        );
    }

    #[test]
    fn the_marker_wins_over_a_plain_term() {
        assert_eq!(
            detect(&[("TERM", "xterm-256color"), ("KITTY_WINDOW_ID", "1")]),
            TerminalGraphicsCapability::Kitty
        );
    }

    #[test]
    fn a_plain_terminal_reports_nothing() {
        assert_eq!(
            detect(&[
                ("TERM", "xterm-256color"),
                ("TERM_PROGRAM", "Apple_Terminal")
            ]),
            TerminalGraphicsCapability::None
        );
    }
}
