#!/usr/bin/env bash
#
# Regenerates the Prumo workforce for this project.
#
# Why this exists instead of a bare `prumo workforce sync`: the sync rewrites
# `.ai/skills/manifest.json` to exactly the skills it was asked for. Skills that
# the framework catalog does not cover — the ones authored in this repository —
# are discovered by directory, but they must also appear in that manifest or
# they silently drop out of the workforce. Running sync alone would therefore
# erase the local skills from the selection every time.
#
# So: sync the catalog set resolved from the profile, then re-add the local set,
# then compile the adapters. One command, and the local list below is the single
# place it can drift from.
#
# Usage: scripts/workforce.sh [--target <harness>]

set -euo pipefail

TARGET="${2:-antigravity}"
if [ "${1:-}" = "--target" ] && [ -z "${2:-}" ]; then
  echo "error: --target requires a value" >&2
  exit 2
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# Skills authored here, because the framework catalog has no equivalent.
# Changing this list changes the agent surface of the project.
LOCAL_SKILLS=(
  bubbletea-v2
  conformance-parity
  cross-platform-terminal
  doc-drift
  lsp-client
  pty-vt
  release-go
  tui-editor-engine
)

for skill in "${LOCAL_SKILLS[@]}"; do
  if [ ! -f ".ai/skills/${skill}/SKILL.md" ]; then
    echo "error: skill local '${skill}' está na lista mas não tem SKILL.md" >&2
    exit 1
  fi
done

# `prumo` blocks on stdin; every invocation is closed explicitly.
echo "== resolvendo o workforce a partir de project-profile.json"
CATALOG_SKILLS="$(prumo --json resolve project-profile.json </dev/null \
  | python3 -c 'import json,sys; print(",".join(json.load(sys.stdin)["data"]["skills"]))')"

if [ -z "$CATALOG_SKILLS" ]; then
  echo "error: a resolução não devolveu nenhuma skill" >&2
  exit 1
fi

echo "== sincronizando $(tr ',' '\n' <<<"$CATALOG_SKILLS" | wc -l) skills do catálogo"
prumo workforce sync --offline --skills "$CATALOG_SKILLS" --path "$REPO_ROOT" </dev/null

CATALOG_SKILLS="$CATALOG_SKILLS" python3 - "$REPO_ROOT" "${LOCAL_SKILLS[@]}" <<'PY'
import json
import os
import sys

root, local = sys.argv[1], sys.argv[2:]
path = os.path.join(root, ".ai", "skills", "manifest.json")

with open(path) as handle:
    manifest = json.load(handle)

catalog = [s for s in manifest.get("skills", []) if s not in local]
merged = catalog + [s for s in local if s not in catalog]
manifest["skills"] = merged

with open(path, "w") as handle:
    json.dump(manifest, handle, indent=2, ensure_ascii=False)
    handle.write("\n")

print(f"== manifest: {len(catalog)} do catálogo + {len(local)} locais = {len(merged)}")
PY

echo "== gerando as roles de agente a partir do contrato canônico"
# `prumo explain agent <id>` é a única forma suportada de ler o contrato de um
# agente: `workforce sync` não instala agentes, e o corpo em prosa do AGENT.md
# só existe no FS embutido do binário. O que sai daqui é o contrato
# estruturado — propósito, entradas, saídas, skills exigidas, permissões e
# nível de risco — que é o que um harness precisa para escolher e limitar um
# papel. Gerado, não transcrito: reexecutar o script atualiza.
for agent in $(python3 -c "
import json
print(' '.join(json.load(open('.ai/agents/manifest.json'))['agents']))
"); do
  mkdir -p ".ai/agents/$agent"
  prumo --json explain agent "$agent" </dev/null > "/tmp/prumo-agent-$agent.json"
  python3 - "$agent" <<'PY'
import json
import sys

agent = sys.argv[1]
with open(f"/tmp/prumo-agent-{agent}.json") as handle:
    payload = json.load(handle)
data = payload.get("data") or {}
if not data:
    raise SystemExit(f"error: explain agent {agent} não devolveu contrato")

def bullets(items, prefix=""):
    return "\n".join(f"- {prefix}{item}" for item in items) or "- (nenhum)"

permissions = data.get("permissions") or {}
required = data.get("required_skills") or []
skills_block = "\n".join(f"- `{s}`" for s in required) or "- (nenhuma)"

lines = [
    f"# {data.get('name', agent)}",
    "",
    f"> Contrato gerado por `prumo explain agent {agent}` (v{data.get('version', '?')}).",
    "> Não edite à mão: rode `scripts/workforce.sh`.",
    "",
    "## Propósito",
    "",
    data.get("purpose", "").strip() or "(não declarado)",
    "",
    "## Entradas",
    "",
    bullets(data.get("inputs") or []),
    "",
    "## Saídas",
    "",
    bullets(data.get("outputs") or []),
    "",
    "## Skills exigidas",
    "",
    skills_block,
    "",
    "## Permissões",
    "",
    f"- Nível de risco: `{data.get('risk_level', '?')}`",
    f"- Escrever código: `{permissions.get('write_code')}`",
    f"- Modificar documentação: `{permissions.get('modify_docs')}`",
    f"- Executar testes: `{permissions.get('execute_tests')}`",
    "",
]
with open(f".ai/agents/{agent}/AGENT.md", "w") as handle:
    handle.write("\n".join(lines))
PY
done
echo "   $(ls -d .ai/agents/*/ 2>/dev/null | wc -l) roles geradas"

echo "== compilando o adaptador '$TARGET'"
prumo compile --target "$TARGET" --path "$REPO_ROOT" </dev/null

echo "== validando"
prumo validate "$REPO_ROOT" </dev/null
prumo doctor "$REPO_ROOT" </dev/null
