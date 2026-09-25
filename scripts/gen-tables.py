#!/usr/bin/env python3
"""Generates Go tables from the Rust canonical sources.

Two tables are generated, both because transcription is how a port drifts:

  * ``internal/action/action.go`` from ``crates/oride-keymap/src/action.rs`` —
    the 96 action ids plus the palette order. One id missing means a user's
    config silently stops matching a binding.
  * ``internal/config/keybindings.go`` from
    ``crates/oride-config/src/model.rs`` — the 98 default chord→action pairs.
    A chord missing means a documented shortcut stops working.

Each binding's action id is checked against the action table extracted in the
same run, so a default that names a non-existent action fails generation instead
of failing at runtime.

Output is piped through ``gofmt`` so ``--check`` compares like with like.

Usage: scripts/gen-tables.py [--check]
"""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from collections import Counter
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
ACTION_SOURCE = REPO_ROOT / "crates" / "oride-keymap" / "src" / "action.rs"
CONFIG_SOURCE = REPO_ROOT / "crates" / "oride-config" / "src" / "model.rs"
THEME_SOURCE = REPO_ROOT / "crates" / "oride-config" / "src" / "theme.rs"
ACTION_TARGET = REPO_ROOT / "internal" / "action" / "action.go"
KEYBINDINGS_TARGET = REPO_ROOT / "internal" / "config" / "keybindings.go"
THEMES_TARGET = REPO_ROOT / "internal" / "config" / "themes.go"

ID_ARM = re.compile(
    r'^\s*Self::(?P<variant>[A-Za-z]+)(?:\s*\{\s*extend:\s*(?P<extend>true|false)\s*\})?'
    r'\s*=>\s*"(?P<id>[a-z0-9_]+)"\s*,\s*$'
)

PALETTE_ARM = re.compile(r"^\s*Action::(?P<variant>[A-Za-z]+)(?:\s*\{[^}]*\})?\s*,\s*$")

# `fn dracula() -> ThemeDefinition {` starts a theme definition.
THEME_HEAD = re.compile(r"^fn (?P<fn>[a-z_]+)\(\) -> ThemeDefinition \{$")

# `name: "Dracula".into(),` / `is_dark: true,` / `background: "#282a36".into(),`
THEME_FIELD = re.compile(
    r'^\s+(?P<key>[a-z_]+):\s+(?:"(?P<text>[^"]*)"\.into\(\)|(?P<bool>true|false)|(?P<number>\d+))'
)

# `ui: ThemeUiConfig::default(),` — the theme inherits the product defaults.
THEME_DEFAULT_SECTION = re.compile(r"^\s+(?P<section>ui|syntax):\s+\w+::default\(\),")

THEME_ORDER = re.compile(r"^\s+(?P<fn>[a-z_]+)\(\),")

BINDING_PAIR = re.compile(
    r'^\s*\(\s*"(?P<chord>(?:[^"\\]|\\.)*)"\s*,\s*"(?P<action>[a-z0-9_]+)"\s*\)\s*,\s*$'
)

# Any line inside the pairs array that starts an entry. Used to refuse silent
# skips: if a line looks like an entry and the strict pattern above did not
# match it, the generator stops instead of dropping a binding.
BINDING_LOOKS_LIKE_ENTRY = re.compile(r"^\s*\(\s*\"")

# Rust string escapes that appear in chord literals.
RUST_ESCAPES = {"\\\"": '"', "\\\\": "\\"}


def unescape_rust_literal(raw: str) -> str:
    out: list[str] = []
    i = 0
    while i < len(raw):
        pair = raw[i : i + 2]
        if pair in RUST_ESCAPES:
            out.append(RUST_ESCAPES[pair])
            i += 2
            continue
        out.append(raw[i])
        i += 1
    return "".join(out)


def require(path: Path) -> str:
    if not path.is_file():
        sys.exit(f"error: {path} não encontrado")
    return path.read_text(encoding="utf-8")


def gofmt(source: str) -> str:
    try:
        result = subprocess.run(
            ["gofmt"], input=source, capture_output=True, text=True, check=True
        )
    except FileNotFoundError:
        sys.exit("error: gofmt não encontrado no PATH")
    except subprocess.CalledProcessError as error:
        sys.exit(f"error: gofmt rejeitou a saída do gerador:\n{error.stderr}")
    return result.stdout


# --------------------------------------------------------------------------
# Action table
# --------------------------------------------------------------------------


def extract_id_arms(source: str) -> list[tuple[str, bool | None, str]]:
    start = source.index("pub fn id(self)")
    end = source.index("pub fn palette_label", start)
    arms: list[tuple[str, bool | None, str]] = []
    for line in source[start:end].splitlines():
        match = ID_ARM.match(line)
        if not match:
            continue
        extend_raw = match.group("extend")
        arms.append(
            (
                match.group("variant"),
                None if extend_raw is None else extend_raw == "true",
                match.group("id"),
            )
        )
    if not arms:
        sys.exit("error: nenhum braço de id() extraído — o formato do Rust mudou")
    return arms


def extract_palette(source: str) -> list[str]:
    start = source.index("pub fn palette_actions")
    end = source.index("]", source.index("&[", start))
    names: list[str] = []
    for line in source[start:end].splitlines():
        match = PALETTE_ARM.match(line)
        if match:
            names.append(match.group("variant"))
    return names


def go_constant(variant: str, extend: bool | None) -> str:
    if extend is None:
        return variant
    return f"{variant}{'Extend' if extend else 'Plain'}"


def render_actions(arms: list[tuple[str, bool | None, str]], palette: list[str]) -> str:
    by_variant: dict[str, list[str]] = {}
    for variant, extend, _ in arms:
        by_variant.setdefault(variant, []).append(go_constant(variant, extend))

    lines = [
        "// Code generated by scripts/gen-tables.py from crates/oride-keymap/src/action.rs. DO NOT EDIT.",
        "",
        "// Package action is the canonical table of editor actions.",
        "//",
        "// An action is a stable string id, not an enum: in the Rust implementation a",
        "// variant and a TOML id were two identities joined by a parser that could",
        "// disagree with itself. Here the string is the identity, so a config keymap, a",
        "// palette entry and a keybinding all name the same thing by construction.",
        "package action",
        "",
        "// Action is a stable action id. The zero value is not a valid action.",
        "type Action string",
        "",
        "// Action ids, in the declaration order of the Rust canonical table.",
        "const (",
    ]
    seen: set[str] = set()
    for variant, extend, action_id in arms:
        constant = go_constant(variant, extend)
        if constant in seen:
            continue
        seen.add(constant)
        lines.append(f'\t{constant} Action = "{action_id}"')
    lines.append(")")
    lines += ["", "// all is the canonical table, in declaration order.", "var all = []Action{"]
    lines += [f"\t{go_constant(v, e)}," for v, e, _ in arms]
    lines += ["}", "", "// paletteOrder is the curated order of the command palette.", "var paletteOrder = []Action{"]
    for variant in palette:
        candidates = by_variant.get(variant)
        if not candidates:
            sys.exit(f"error: palette cita '{variant}', que não tem id canônico")
        if len(candidates) > 1:
            sys.exit(f"error: palette cita '{variant}', que tem variantes; seja explícito")
        lines.append(f"\t{candidates[0]},")
    lines += ["}", ""]
    return "\n".join(lines)


# --------------------------------------------------------------------------
# Default key bindings
# --------------------------------------------------------------------------


def extract_bindings(source: str) -> list[tuple[str, str]]:
    start = source.index("pub fn default_key_bindings")
    end = source.index("pairs\n", start)
    bindings: list[tuple[str, str]] = []

    for number, line in enumerate(source[start:end].splitlines(), start=1):
        match = BINDING_PAIR.match(line)
        if match:
            bindings.append(
                (unescape_rust_literal(match.group("chord")), match.group("action"))
            )
            continue
        # A line that starts an entry but did not parse is a binding about to be
        # dropped in silence. That is exactly the bug this generator shipped once
        # — an escaped quote in `("ctrl+\"", …)` — so it is now a hard failure.
        if BINDING_LOOKS_LIKE_ENTRY.match(line):
            sys.exit(
                f"error: linha {number} parece um binding mas não foi interpretada:\n  {line.strip()}"
            )

    if not bindings:
        sys.exit("error: nenhum binding extraído — o formato do Rust mudou")
    return bindings


def render_keybindings(bindings: list[tuple[str, str]], known_ids: set[str]) -> str:
    unknown = sorted({a for _, a in bindings if a not in known_ids})
    if unknown:
        sys.exit(f"error: bindings default citam ações inexistentes: {', '.join(unknown)}")

    duplicates = sorted(
        chord for chord, count in Counter(chord for chord, _ in bindings).items() if count > 1
    )
    if duplicates:
        sys.exit(f"error: chord repetido nos defaults: {', '.join(duplicates)}")

    lines = [
        "// Code generated by scripts/gen-tables.py from crates/oride-config/src/model.rs. DO NOT EDIT.",
        "",
        "package config",
        "",
        "// defaultKeyBindings is the shipped chord → action id map.",
        "//",
        "// Emitted sorted, not in declaration order: a Go map has no order, and a",
        "// stable file is what makes --check meaningful.",
        "//",
        "// Values are ids, not resolved actions: parsing belongs to internal/keymap,",
        "// exactly as `Keymap::from_string_map` owns it in the Rust implementation.",
        "// A test asserts every pair parses, so an id that stops existing fails the",
        "// build rather than one user's keybinding.",
        "var defaultKeyBindings = map[string]string{",
    ]
    for chord, action_id in sorted(bindings):
        lines.append(f"\t{go_string(chord)}: {go_string(action_id)},")
    lines += ["}", ""]
    return "\n".join(lines)


# --------------------------------------------------------------------------


def go_string(value: str) -> str:
    """Renders a Go string literal.

    Backslash first, then quote: the other order would escape the backslash that
    the quote escape just introduced. A chord like `ctrl+"` is a legal binding
    and would otherwise produce invalid Go.
    """
    return '"' + value.replace("\\", "\\\\").replace('"', '\\"') + '"'


def extract_themes(source: str) -> list[dict]:
    """Reads every ThemeDefinition function, in the order built_in_themes lists them."""
    start = source.index("pub fn built_in_themes")
    order_block = source[start:source.index("]", source.index("vec![", start))]
    order = [m.group("fn") for m in (THEME_ORDER.match(line) for line in order_block.splitlines()) if m]

    definitions: dict[str, dict] = {}
    for fn in order:
        head = source.index(f"fn {fn}() -> ThemeDefinition {{")
        body = source[head:source.index("\n}", head)]

        name_match = re.search(r'name: "([^"]*)"\.into\(\)', body)
        if not name_match:
            sys.exit(f"error: tema {fn} sem nome")
        dark_match = re.search(r"is_dark: (true|false)", body)
        if not dark_match:
            sys.exit(f"error: tema {fn} sem is_dark")

        theme = {
            "fn": fn,
            "name": name_match.group(1),
            "is_dark": dark_match.group(1) == "true",
            "ui": {},
            "syntax": {},
            "ui_default": "ui: ThemeUiConfig::default()" in body,
            "syntax_default": "syntax: SyntaxColorsConfig::default()" in body,
        }

        section = None
        for line in body.splitlines():
            if THEME_DEFAULT_SECTION.match(line):
                section = None
                continue
            if re.match(r"^\s+ui: ThemeUiConfig \{", line):
                section = "ui"
                continue
            if re.match(r"^\s+syntax: SyntaxColorsConfig \{", line):
                section = "syntax"
                continue
            if section and line.strip() == "},":
                section = None
                continue
            field = THEME_FIELD.match(line)
            if section and field:
                key = field.group("key")
                if key in ("name", "is_dark"):
                    continue
                # Integers are kept as numbers, not as strings: gutter_width is a
                # count, and emitting it quoted would not compile.
                if field.group("number") is not None:
                    theme[section][key] = int(field.group("number"))
                elif field.group("bool") is not None:
                    theme[section][key] = field.group("bool") == "true"
                else:
                    theme[section][key] = field.group("text")
        definitions[fn] = theme

    if not definitions:
        sys.exit("error: nenhum tema extraído — o formato do Rust mudou")
    return [definitions[fn] for fn in order]


def render_themes(themes: list[dict]) -> str:
    lines = [
        "// Code generated by scripts/gen-tables.py from crates/oride-config/src/theme.rs. DO NOT EDIT.",
        "",
        "package config",
        "",
        "// builtInThemes is the shipped theme set, in the order the reference lists it.",
        "//",
        "// A theme that inherits the product defaults carries no colours here; the",
        "// registry fills them from Default() so the two cannot drift apart.",
        "var builtInThemes = []ThemeDefinition{",
    ]
    for theme in themes:
        lines.append("\t{")
        lines.append(f'\t\tName:   {go_string(theme["name"])},')
        lines.append(f'\t\tIsDark: {"true" if theme["is_dark"] else "false"},')
        if not theme["ui_default"]:
            lines.append("\t\tUI: UIConfig{")
            for key, value in sorted(theme["ui"].items()):
                rendered = go_string(value) if isinstance(value, str) else str(value).lower()
                lines.append(f"\t\t\t{goField(key)}: {rendered},")
            lines.append("\t\t},")
        if not theme["syntax_default"]:
            lines.append("\t\tSyntax: SyntaxColors{")
            for key, value in sorted(theme["syntax"].items()):
                lines.append(f"\t\t\t{goField(key)}: {go_string(value)},")
            lines.append("\t\t},")
        lines.append("\t},")
    lines += ["}", ""]
    return "\n".join(lines)


def goField(snake: str) -> str:
    """`type_name` → `TypeName`, `status_bg` → `StatusBG`, `cursor_fg` → `CursorFG`."""
    special = {"bg": "BG", "fg": "FG", "lsp": "LSP", "ui": "UI"}
    return "".join(special.get(part, part.capitalize()) for part in snake.split("_"))


def write_or_check(target: Path, content: str, check: bool, label: str) -> bool:
    if check:
        if not target.is_file():
            print(f"error: {target.relative_to(REPO_ROOT)} não existe; rode scripts/gen-tables.py", file=sys.stderr)
            return False
        if target.read_text(encoding="utf-8") != content:
            print(
                f"error: {target.relative_to(REPO_ROOT)} está desatualizado em relação ao Rust;"
                " rode scripts/gen-tables.py",
                file=sys.stderr,
            )
            return False
        print(f"ok: {target.name} em sincronia ({label})")
        return True

    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(content, encoding="utf-8")
    print(f"escrito: {target.relative_to(REPO_ROOT)} ({label})")
    return True


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true", help="falha se algum arquivo estiver desatualizado")
    args = parser.parse_args()

    action_source = require(ACTION_SOURCE)
    arms = extract_id_arms(action_source)
    palette = extract_palette(action_source)
    known_ids = {i for _, _, i in arms}

    binding_source = require(CONFIG_SOURCE)
    bindings = extract_bindings(binding_source)

    ok = write_or_check(
        ACTION_TARGET,
        gofmt(render_actions(arms, palette)),
        args.check,
        f"{len(arms)} ids, {len(palette)} na palette",
    )
    ok &= write_or_check(
        KEYBINDINGS_TARGET,
        gofmt(render_keybindings(bindings, known_ids)),
        args.check,
        f"{len(bindings)} bindings",
    )

    theme_source = require(THEME_SOURCE)
    themes = extract_themes(theme_source)
    ok &= write_or_check(
        THEMES_TARGET,
        gofmt(render_themes(themes)),
        args.check,
        f"{len(themes)} temas",
    )
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
