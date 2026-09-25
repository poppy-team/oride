# ISO C Systems Programming & Memory Safety Contract

## 1. Purpose
Author, refactor, and audit systems software written in modern **ISO C** (C11, C17, or C23). This skill mandates strict bounded buffer operations, elimination of banned and dangerous standard library functions, pointer provenance discipline, zero unchecked allocations, automatic scope cleanup via `__attribute__((cleanup))` where supported, zero-warning compilation with `-Werror`, and runtime verification with AddressSanitizer (ASan) and UndefinedBehaviorSanitizer (UBSan).

---

## 2. Use When
- Developing low-level systems libraries, operating system interfaces, embedded modules, or runtime runloops in C.
- Authoring C-ABI foreign function interface (FFI) bindings for Go, Rust, Python, or WebAssembly.
- Writing high-throughput data processing algorithms requiring raw memory access with zero compiler-induced overhead.
- Auditing legacy C codebases for memory corruption vulnerabilities, buffer overflows, or integer wrap-around defects.
- Operating in mode(s): `implementation`, `review`, `testing`, `audit`.

---

## 3. Do Not Use When
- Developing in C++ with RAII, templates, and classes (use `lang-cpp`).
- Writing memory-safe systems code where Rust's compiler borrow checker is available and mandated (use `lang-rust`).
- Working purely on web frontend UI or scripting environments.

---

## 4. Required Context
Before implementing or modifying C code, determine:
- **Standard Dialect**: ISO C11 (`-std=c11`), C17 (`-std=c17`), or C23 (`-std=c2x` / `-std=c23`).
- **Platform ABI & Alignment Constraints**: Data model (LP64 on 64-bit Linux/macOS vs LLP64 on Windows), pointer size, and structure alignment requirements.
- **Buffer Safety Protocol**: Use of explicit `(void *buf, size_t cap, size_t *out_len)` parameter tuples or structured slice types (`struct BufferSlice { uint8_t *data; size_t len; size_t cap; }`).
- **Fortification & Compiler Hardening**: Presence of `_FORTIFY_SOURCE=2` or `_FORTIFY_SOURCE=3`, stack canaries (`-fstack-protector-strong`).

---

## 5. Procedure

```
[API Design & Type Definition]
         |
         v
[1. Buffer Tuple Specification] ----> (ptr, cap, out_len) mandatory on all writes
         |
         v
[2. Dangerous API Ban & Scan] -------> Eradicate gets, strcpy, strcat, sprintf
         |
         v
[3. Allocation & Pointer Discipline]-> Immediate NULL checks, zero after free
         |
         v
[4. Static Analysis & Fortify] ------> -Wall -Wextra -Wpedantic -Werror -D_FORTIFY_SOURCE=3
         |
         v
[5. Dynamic ASan/UBSan Gauntlet] ----> Zero leaks, zero out-of-bounds, zero UB
```

### Step 1: Bounded Buffer Operations
- Every function accepting or populating a buffer must accept explicit capacity constraints:
  ```c
  int safe_copy_string(char *dest, size_t dest_cap, const char *src);
  ```
- Use structured buffer descriptors for binary I/O:
  ```c
  typedef struct {
      uint8_t *data;
      size_t len;
      size_t cap;
  } ByteBuffer;
  ```
- Variable Length Arrays (VLAs, e.g. `char buf[n]`) are **strictly prohibited**. Use fixed-size stack buffers with overflow checks, or heap/arena allocation.

### Step 2: Elimination of Banned APIs
The following legacy C standard library functions are strictly forbidden:
- `gets()` $\rightarrow$ use `fgets()` with bounded length.
- `strcpy()`, `strcat()` $\rightarrow$ use `snprintf()` or explicit bounded loops with null termination.
- `sprintf()`, `vsprintf()` $\rightarrow$ use `snprintf()`, `vsnprintf()`.
- `atoi()`, `atol()` $\rightarrow$ use `strtol()`, `strtoll()` with `endptr` and `errno` validation.

### Step 3: Allocation Safety & Lifecycle Hygiene
1. Check every allocation immediately:
   ```c
   void *ptr = malloc(size);
   if (ptr == NULL) {
       return ERR_OUT_OF_MEMORY;
   }
   ```
2. Free and nullify:
   ```c
   free(ptr);
   ptr = NULL; // Prevent Use-After-Free
   ```
3. When targeting GCC or Clang on Linux/POSIX, utilize `__attribute__((cleanup))` to implement deterministic RAII-like scope cleanups for dynamic resources.

### Step 4: Strict Aliasing & Provenance Invariants
- Never cast between incompatible pointer types (e.g., `float *` to `int *`).
- To inspect or reinterpret binary representation, use `memcpy()` or copy into `unsigned char *`:
  ```c
  uint32_t raw_bits;
  memcpy(&raw_bits, &float_val, sizeof(raw_bits));
  ```
- Obey pointer provenance: arithmetic on a pointer must remain within the bounds of the original allocated object.

### Step 5: Static Hardening & Dynamic Sanitizers
1. Compile with strict flags:
   `-std=c17 -Wall -Wextra -Wpedantic -Wconversion -Wsign-conversion -Wshadow -Wformat=2 -Wstrict-prototypes -Wmissing-prototypes -Werror -D_FORTIFY_SOURCE=3 -fstack-protector-strong`.
2. Compile and run the test suite under AddressSanitizer and UndefinedBehaviorSanitizer:
   `-fsanitize=address,undefined -fno-omit-frame-pointer`.

---

## 6. Decision Rules & Invariants
- **RULE 1 (Explicit Buffer Bounds)**: Any function writing to a pointer without receiving an explicit capacity parameter is rejected immediately.
- **RULE 2 (No Unchecked Heap Operations)**: Any dereference of a pointer returned by `malloc`, `calloc`, or `realloc` without an antecedent `NULL` check is a fatal quality defect.
- **RULE 3 (No VLAs on Stack)**: Variable-length arrays (`type arr[n]`) are forbidden; they risk unbounded stack growth and remote code execution vulnerabilities.
- **RULE 4 (Explicit Return Codes)**: Every fallible function must return an explicit status code (enum or negative error integer) or error struct. Functions must never rely on side effects to communicate failure.

---

## 7. Evidence Required
- **Compiler Evidence**: Clean compilation under `-std=c17` or `-std=c2x` with `-Werror` and `-D_FORTIFY_SOURCE=3`.
- **Sanitizer Evidence**: Zero memory leaks, zero buffer overflows, and zero undefined behavior warnings under ASan/UBSan execution.
- **Boundary Verification**: Unit test evidence verifying that buffer truncated conditions return expected error codes without heap or stack corruption.

---

## 8. Output Contract
- Hardened C source (`.c`) and header (`.h`) files with full prototype signatures and header guards.
- Makefile or CMake configuration specifying strict flags and sanitizer profiles.
- C Verification Report (`templates/c-verification-report.md`).

---

## 9. Stop Conditions
- All tests pass with zero findings under ASan and UBSan.
- Static analysis and compiler run with `-Werror` and report zero diagnostics.
- Token budget exhausted.
- Blocked on external dependency.

---

## 10. Escalation Rules
- Escalate immediately if an external legacy C library with known buffer overflow risks must be linked without a memory-safe wrapper boundary.
- Escalate if hardware-specific memory-mapped I/O requires disabling strict aliasing flags (`-fno-strict-aliasing`).
