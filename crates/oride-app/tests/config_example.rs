//! O exemplo de config é documentação executável, e documentação que ninguém
//! verifica diverge.
//!
//! Um `[editor]` duplicado, um id de ação que não existe ou um chord inválido
//! passam despercebidos em revisão — quem copia o arquivo descobre com o editor
//! quebrado. Estes testes fazem o exemplo falhar o build, que é o único lugar
//! onde esse erro custa pouco.

use oride_config::Config;
use oride_keymap::{parse_action, parse_chord};

const EXAMPLE: &str = include_str!("../../../assets/config.example.toml");

fn example_value() -> toml::Value {
    toml::from_str(EXAMPLE).expect("assets/config.example.toml precisa ser TOML válido")
}

#[test]
fn example_has_no_duplicate_tables() {
    // Uma tabela `[editor]` declarada duas vezes é erro de TOML, não estilo: o
    // arquivo anterior tinha duas e teria quebrado qualquer parser estrito.
    let value = example_value();
    assert!(
        value.get("editor").is_some_and(toml::Value::is_table),
        "seção [editor] ausente ou não é tabela"
    );
}

#[test]
fn every_documented_chord_and_action_exists() {
    let value = example_value();
    let keys = value
        .get("keys")
        .and_then(toml::Value::as_table)
        .expect("seção [keys] ausente no exemplo");

    assert!(!keys.is_empty(), "seção [keys] vazia não documenta nada");

    for (chord, action) in keys {
        parse_chord(chord)
            .unwrap_or_else(|error| panic!("chord `{chord}` inválido no exemplo: {error}"));

        let action_id = action
            .as_str()
            .unwrap_or_else(|| panic!("ação do chord `{chord}` precisa ser string"));
        parse_action(action_id).unwrap_or_else(|_| {
            panic!("ação `{action_id}` (chord `{chord}`) não existe em Action")
        });
    }
}

#[test]
fn documented_mouse_default_matches_the_code() {
    // O README, o guia e o CHANGELOG afirmaram por muito tempo que o mouse vem
    // desligado enquanto o código dizia ligado. O exemplo é o lado que o usuário
    // lê, então ele falha o build quando os dois divergem.
    let value = example_value();
    let documented = value
        .get("mouse")
        .and_then(toml::Value::as_bool)
        .expect("chave `mouse` ausente no exemplo");
    assert_eq!(
        documented,
        Config::default().mouse,
        "o default documentado de `mouse` divergiu do default do código"
    );
}

#[test]
fn documented_editor_defaults_match_the_code() {
    let value = example_value();
    let editor = value
        .get("editor")
        .and_then(toml::Value::as_table)
        .expect("seção [editor] ausente");
    let defaults = Config::default();

    for (key, expected) in [
        ("tab_size", u64::from(defaults.editor.tab_size)),
        (
            "completion_min_chars",
            u64::from(defaults.editor.completion_min_chars),
        ),
    ] {
        if let Some(documented) = editor.get(key).and_then(toml::Value::as_integer) {
            assert_eq!(
                documented, expected as i64,
                "`editor.{key}` documentado como {documented} mas o default é {expected}"
            );
        }
    }

    for (key, expected) in [
        ("insert_spaces", defaults.editor.insert_spaces),
        ("format_on_save", defaults.editor.format_on_save),
        ("use_editorconfig", defaults.editor.use_editorconfig),
        ("completion_auto", defaults.editor.completion_auto),
    ] {
        if let Some(documented) = editor.get(key).and_then(toml::Value::as_bool) {
            assert_eq!(
                documented, expected,
                "`editor.{key}` documentado como {documented} mas o default é {expected}"
            );
        }
    }
}

#[test]
fn documented_keys_section_only_overrides_real_defaults() {
    // A seção [keys] do exemplo lista bindings que já são default. Se um deles
    // sair dos defaults, o exemplo passa a mentir sobre o que vem embutido.
    let value = example_value();
    let keys = value
        .get("keys")
        .and_then(toml::Value::as_table)
        .expect("seção [keys] ausente");
    let defaults = Config::default().keys;

    for (chord, action) in keys {
        let action_id = action.as_str().expect("ação precisa ser string");
        match defaults.get(chord) {
            Some(default_action) => assert_eq!(
                default_action, action_id,
                "chord `{chord}`: o exemplo diz `{action_id}` mas o default é `{default_action}`"
            ),
            None => panic!("chord `{chord}` está no exemplo mas não é um default do produto"),
        }
    }
}
