# YAML Safety Reference
1. **Safe Loading**: Always use `yaml.safe_load(content)` in Python and equivalent safe unmarshalers.
2. **Scalar Quoting**: Values like `yes`, `no`, `on`, `off` must be quoted if intended as strings.\n