# CI Failure Diagnosis Record

## Summary
- Failing run: `https://github.com/acme/api/actions/runs/1842`
- Job: `test (ubuntu-22.04, node 20)`
- First failing step: `build`
- Exit code: `1`

## Classification
- Last green run: `1841`
- Suspect delta: Ubuntu runner image updated OpenSSL defaults.
- Classification: runner-image drift, not an application regression.

## First Error
```text
Error: error:0308010C:digital envelope routines::unsupported
```

## Fix and Evidence
- Added `NODE_OPTIONS=--openssl-legacy-provider` only to the affected matrix cell.
- Reproduced with the matching `ubuntu:22.04` container before editing YAML.
- Re-run `1843` is green; no required check was disabled.
- Follow-up issue #1856 tracks the permanent webpack upgrade.
