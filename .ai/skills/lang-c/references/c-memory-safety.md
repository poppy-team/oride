# ISO C Memory Safety, Strict Aliasing & Provenance Reference

## 1. Explicit Bounds & Buffer Protocol
- In C, arrays decay to raw pointers when passed to functions, losing size information.
- Every function accepting or populating a buffer must enforce explicit bounds:
  ```c
  int process_stream(const uint8_t *src, size_t src_len, uint8_t *dest, size_t dest_cap, size_t *out_written);
  ```
- Before writing, verify `*out_written + chunk_size <= dest_cap`.

---

## 2. Allocation & Integer Overflow Defense
1. **Multiplication Overflow Hazard**:
   ```c
   // CATASTROPHIC BUG: If count * sizeof(Item) overflows, small buffer is allocated,
   // leading to catastrophic heap buffer overflow!
   Item *items = malloc(count * sizeof(Item));
   ```
2. **Defensive Arithmetic Check**:
   ```c
   if (count > SIZE_MAX / sizeof(Item)) {
       return ERR_INTEGER_OVERFLOW;
   }
   Item *items = malloc(count * sizeof(Item));
   if (items == NULL) {
       return ERR_OUT_OF_MEMORY;
   }
   ```
3. Alternatively, use `calloc(count, sizeof(Item))`, which guarantees overflow checking on allocation parameters.

---

## 3. Strict Aliasing Rules (ISO C §6.5/7)
The compiler assumes two pointers of different types do not refer to the same memory location, allowing aggressive reordering of loads and stores.

### Permitted Aliasing Accesses:
An object's stored value may only be accessed by an lvalue expression having:
1. A type compatible with the effective type of the object.
2. A qualified version of a type compatible with the effective type.
3. A type that is the signed or unsigned type corresponding to the effective type.
4. An aggregate or union type that includes one of the aforementioned types among its members.
5. A character type (`char`, `signed char`, `unsigned char`).

**Illegal Type Punning**:
```c
// STRICT ALIASING VIOLATION (Undefined Behavior in ISO C):
float f = 5.0f;
int *i = (int *)&f; // UB! Compiler may reorder reads/writes assuming f and i are distinct.
```

**Legal Reinterpretation via memcpy**:
```c
float f = 5.0f;
int i;
memcpy(&i, &f, sizeof(i)); // 100% standard-compliant; optimized by compiler to register move.
```

---

## 4. Pointer Provenance & The Abstract Machine
In modern optimizing C compilers (Clang, GCC):
- A pointer is not merely an integer address; it carries **provenance** (association with the original allocation).
- Computing an address from an integer (`(char *)0x12345678`) or performing pointer arithmetic outside the bounds of the original allocated object produces an invalid pointer. Dereferencing or comparing such pointers with unrelated allocations is undefined behavior.

---

## 5. Scope-Based Cleanup with `__attribute__((cleanup))`
On GCC and Clang, cleanups can be executed automatically when variables exit scope:
```c
static inline void auto_free(void *p) {
    void **ptr = (void **)p;
    if (*ptr != NULL) {
        free(*ptr);
        *ptr = NULL;
    }
}

#define RAII_AUTOFREE __attribute__((cleanup(auto_free)))

void execute_task(void) {
    RAII_AUTOFREE char *buf = malloc(1024);
    if (!buf) return;
    // buf is automatically freed when execute_task exits or returns early
}
```
