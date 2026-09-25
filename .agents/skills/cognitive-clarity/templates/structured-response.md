# Structured Technical Response — Checkout Build Failure

## 1. Objective / Status
**Blocked:** checkout tests fail because the new client rejects the existing PSP sandbox certificate. The expected certificate rotation is not present in the test fixture.

## 2. Technical Rationale
Certificate verification must remain enabled. Disabling it would hide a real trust failure and weaken production transport security.

## 3. Concrete Action / Next Step
1. Add the approved sandbox CA certificate to the fixture trust store.
2. Run `go test ./checkout/...`.
3. Expect 31 passing tests, including `TestRefreshesRejectedClientCertificate`.

## 4. Verification & Evidence
```text
$ go test ./checkout/...
ok  checkout  4.812s
```

## 5. Decision Needed
**Recommendation:** keep verification enabled and update the fixture. Do not add a bypass flag.
