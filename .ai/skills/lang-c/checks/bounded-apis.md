# Banned and Safe C APIs

| Banned Unsafe API | Safe Bounded Replacement |
|---|---|
| `gets(buf)` | `fgets(buf, sizeof(buf), stdin)` |
| `strcpy(dest, src)` | `snprintf(dest, sizeof(dest), "%s", src)` or `strncpy` with explicit null termination |
| `strcat(dest, src)` | `snprintf(dest, sizeof(dest), "%s%s", dest, src)` |
| `sprintf(buf, fmt, ...)` | `snprintf(buf, sizeof(buf), fmt, ...)` |
| `vsprintf(...)` | `vsnprintf(...)` |
| `atoi(str)` | `strtol(str, &endptr, 10)` with boundary and overflow verification |
