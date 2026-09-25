//! Highlight de sintaxe baseado em tree-sitter.

pub mod dynamic_grammar;
mod highlight;
mod kind;
mod language;
mod lexical;
mod markdown;
mod md_preview;

pub use highlight::{line_spans, HighlightEngine, HighlightSpan};
pub use kind::HighlightKind;
pub use language::{detect_language, LanguageId};
pub use markdown::{continue_list_on_enter, list_prefix};
pub use md_preview::{
    detect_terminal_graphics, format_file_size, inline_segments, inspect_image_file,
    render_preview_lines, render_preview_lines_in, render_preview_lines_with_config,
    render_preview_lines_with_graphics, ImageMetadata, PreviewLine, PreviewLink, PreviewStyle,
    TerminalEnvironment, TerminalGraphicsCapability,
};
