---
name: lang-yaml
description: YAML schema-first validation, safe deserialization (safe_load), duplicate key rejection, anchor limits, and secret scrubbing.
---

# YAML Engineering & Safety Contract

## 1. Schema-First Validation
- All structured YAML manifests (Kubernetes, GitHub Actions, application configs) must be validated against a formal JSON Schema or tool (e.g. \`kubeconform\`).
- Adhere strictly to Yamllint rules (two-space indentation, document start \`---\`, newline at EOF).

## 2. Safe Deserialization & Parser Security
- Application parsers must use safe loaders (\`yaml.safe_load\`, PyYAML \`SafeLoader\`, Go \`gopkg.in/yaml.v3\`). Prohibit custom object instantiation.
- Strictly reject duplicate mapping keys.
- Limit YAML anchor (\`&anchor\`) and alias (\`*alias\`) nesting to prevent exponential entity expansion (billion-laughs DOS).

## 3. Explicit Typing & Secret Protection
- Quote ambiguous string literals that can be coerced to booleans or numbers (e.g. \`"yes"\`, \`"no"\`, \`"true"\`, \`"1.0"\`).
- Zero plaintext secrets in YAML files. Use Secret references, KMS, or Vault injections.\n