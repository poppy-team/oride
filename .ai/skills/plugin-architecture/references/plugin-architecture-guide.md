# Plugin Architecture Reference Guide

## 1. Core Concepts

### 1.1 Capabilities vs Ambient Authority
Ambient authority (Node's `require('fs')`, Python's `open()`) gives code everything the user has. Capability design inverts this: the host mints narrow handles (`vfs.read(workspacePath)`, `fetch(allowlistedUrl)`) and plugins receive only what their manifest declares. A path-traversal bug in a capability VFS fails closed at the root check; the same bug with ambient `fs` reads `/etc/passwd`.

### 1.2 Permission Manifests and Install-Time Consent
`plugin.json` lists scopes (`fs.read:workspace`, `net.fetch:api.example.com`, `ui.commands:register`) with human-readable justifications shown at install. Runtime enforcement mirrors the manifest: the host proxy checks each call against granted scopes and throws `PermissionDenied { plugin, scope, call }`. Consent recorded at install is auditable; silent scope creep on update re-prompts.

### 1.3 Sandbox Mechanisms Compared
| Mechanism | Isolation strength | Startup cost | Best for |
|---|---|---|---|
| WebAssembly (wasmtime/wasmer) | Strong: linear memory, no ambient syscalls | ~1 ms | Untrusted third-party logic |
| Isolated JS realm (isolated-vm, ShadowRealm) | Medium: no DOM/Node globals, host bridged | <1 ms | Editor extensions, themes |
| OS process + IPC | Strongest: UID/seccomp/cgroups | ~50 ms | Native dylibs, compilers |
| In-process modules | None: full host access | Zero | First-party code only, never third-party |

Default to WASM for untrusted code; reserve in-process loading for signed first-party plugins.

### 1.4 Quotas and Watchdogs
Memory caps (128 MB linear memory), CPU ceilings (500 ms synchronous call, watchdog thread aborts runaway WASM), and handle budgets (max 32 file watches, 8 sockets) convert denial-of-service attempts into per-plugin suspensions. Log quota events with plugin id; three suspensions in 24 h auto-disables pending review.

### 1.5 Versioning With apiVersion
Hosts evolve; plugins pin `apiVersion: "2.4"`. Minor releases add functions (old plugins unaffected), majors may remove after the 2-release deprecation window. The compat suite installs the top 50 marketplace plugins against each LTS host and fails on undeclared breakage — the only reliable way to keep the "no silent breaks" promise.

## 2. Patterns and Anti-Patterns

| Pattern (do) | Anti-pattern (do not) |
|---|---|
| Capability VFS rooted at workspace; traversal rejected | Pass real cwd paths and "validate" with a regex |
| Permission proxy throwing typed errors per call | Document-only permissions enforced by code review |
| Deprecation warnings for 2 minors, then removal | Rename a host function and fix in-tree callers only |
| Dispose-all-handles reload with 1 s assertion | Reload that leaks listeners until restart |
| Sigstore-signed bundles on every channel | HTTPS download with a checksum in the same JSON |

## 3. Minimal Example: Manifest and Permission Proxy (TypeScript)

```jsonc
// plugin.json — wordcount-plus v1.3.0
{
  "id": "wordcount-plus",
  "version": "1.3.0",
  "apiVersion": "2.4",
  "permissions": [
    { "scope": "fs.read:workspace", "why": "Count words in open Markdown files" },
    { "scope": "ui.commands:register", "why": "Expose 'Count words' palette command" }
  ]
}
```

```ts
// host/permissions.ts — runtime enforcement proxy.
export function grantProxy(pluginId: string, granted: Set<string>, host: HostApi) {
  return new Proxy(host, {
    get(target, prop: string) {
      const need = SCOPE_FOR_CALL[prop]; // e.g. readFile -> "fs.read:workspace"
      if (need && !granted.has(need)) {
        throw new PermissionDenied(pluginId, need, prop);
      }
      return Reflect.get(target, prop);
    },
  });
}
```

Undeclared `readFile` throws before touching disk; the error names plugin, scope, and call for audit. Pair with the hostile battery: attempt traversal, oversized allocation, infinite loop, and metadata socket, asserting containment each release.
