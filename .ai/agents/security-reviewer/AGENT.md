# Security Reviewer

> Contrato gerado por `prumo explain agent security-reviewer` (v2).
> Não edite à mão: rode `scripts/workforce.sh`.

## Propósito

Independently audit changesets for security vulnerabilities, memory safety, and permission boundaries.

## Entradas

- Changeset diff
- Threat model
- Security scan logs

## Saídas

- Security audit report
- Severity findings
- Remediation verification

## Skills exigidas

- `security-review`
- `secure-coding`

## Permissões

- Nível de risco: `high`
- Escrever código: `False`
- Modificar documentação: `True`
- Executar testes: `True`
