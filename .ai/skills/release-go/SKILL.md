---
name: release-go
description: Go release pipeline for a single static binary — version from one source, CGO-free builds proven on every change, cross-compilation without a matrix of toolchains, and packaging that does not hardcode the version.
---

# Release (Go)

## Purpose

Ship one static binary per platform, built from one version number, without a
step that only works on the release engineer's machine.

## Version from a single source

The version is injected at build time and read from one place:

```bash
go build -trimpath \
  -ldflags "-s -w -X github.com/ori-team/oride/internal/buildinfo.Version=${VERSION}" \
  ./cmd/oride
```

The ldflags target is a dedicated `internal/buildinfo` package so the path stays
stable when `main` moves. The previous build hardcoded the version in twelve
files — installers, three distro recipes, Nix, and a CI default — and every
release edited all of them.

## CGO-free is asserted, not hoped

```bash
CGO_ENABLED=0 go build ./...
```

Run on every change, not at release time. A dependency that quietly requires C
turns the static single binary into a fragile one, and the failure appears
exactly when it is most expensive.

Consequence worth stating: with cross-compilation and no CGO, the previous split
between glibc and musl builds disappears. One binary per platform instead of two.

## Targets

| OS | Architectures |
|---|---|
| linux | amd64, arm64 |
| darwin | amd64, arm64 |
| windows | amd64, arm64, 386 |

Windows arm64 and 386 are cheap to include and are the difference between a
package manager accepting the project and rejecting it.

## Packaging

- Checksums are published for every artifact, generated from the artifacts
  themselves.
- Distro recipes and installers read the version from the release tag. A
  filename with a version baked in, and a default version in an installer, are
  two places to forget.
- The Nix build uses the Go builder, not the Rust one it inherited.
- The license and the canonical documentation travel in the archive.

## The packaging trap this project already has

`packaging/debian/build-deb.sh` generates `oride_0.2.0_amd64.deb` from an
internal constant while CI expects that exact filename. The version and the
filename are the same fact stored twice.

## Verification

- The pipeline builds every target from a clean checkout.
- A test asserts the built binary reports the version the tag claims.
- `SHA256SUMS` is checked after publish, not before.
