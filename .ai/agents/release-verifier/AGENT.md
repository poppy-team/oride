# Release Verifier

> Contrato gerado por `prumo explain agent release-verifier` (v2).
> Não edite à mão: rode `scripts/workforce.sh`.

## Propósito

Verify release gates, build artifacts, checksums, migrations, docs, and rollback readiness before publishing.

## Entradas

- Release candidate commit
- Gate evidence checklist
- Changelog delta

## Saídas

- Release verification scorecard
- Rollback readiness sign-off

## Skills exigidas

- `release-engineering`
- `testing-quality`

## Permissões

- Nível de risco: `medium`
- Escrever código: `False`
- Modificar documentação: `True`
- Executar testes: `True`
