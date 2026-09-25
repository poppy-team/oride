// ISO C17 Safe Bounded Dynamic Buffer Reference
// Demonstrates explicit capacity checks, immediate NULL validation,
// zero after free, and safe string formatting with snprintf.

#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef enum {
    BUFFER_OK = 0,
    BUFFER_ERR_INVALID_PARAM = -1,
    BUFFER_ERR_OUT_OF_MEMORY = -2,
    BUFFER_ERR_OVERFLOW = -3,
} BufferStatus;

typedef struct {
    char *data;
    size_t length;
    size_t capacity;
} SafeByteBuffer;

// 1. Initialize buffer with explicit capacity
BufferStatus buffer_init(SafeByteBuffer *buf, size_t initial_cap) {
    if (buf == NULL || initial_cap == 0) {
        return BUFFER_ERR_INVALID_PARAM;
    }

    buf->data = (char *)calloc(initial_cap, sizeof(char));
    if (buf->data == NULL) {
        buf->length = 0;
        buf->capacity = 0;
        return BUFFER_ERR_OUT_OF_MEMORY;
    }

    buf->length = 0;
    buf->capacity = initial_cap;
    return BUFFER_OK;
}

// 2. Safe append with dynamic reallocation check
BufferStatus buffer_append_string(SafeByteBuffer *buf, const char *src) {
    if (buf == NULL || buf->data == NULL || src == NULL) {
        return BUFFER_ERR_INVALID_PARAM;
    }

    size_t src_len = strlen(src);
    size_t needed = buf->length + src_len + 1; // +1 for null terminator

    if (needed > buf->capacity) {
        // Grow exponentially with overflow check
        size_t new_cap = buf->capacity * 2;
        if (new_cap < needed) {
            new_cap = needed;
        }

        char *new_data = (char *)realloc(buf->data, new_cap);
        if (new_data == NULL) {
            return BUFFER_ERR_OUT_OF_MEMORY;
        }

        buf->data = new_data;
        buf->capacity = new_cap;
    }

    // Explicit bounded write
    int written = snprintf(buf->data + buf->length, buf->capacity - buf->length, "%s", src);
    if (written < 0 || (size_t)written >= (buf->capacity - buf->length)) {
        return BUFFER_ERR_OVERFLOW;
    }

    buf->length += (size_t)written;
    return BUFFER_OK;
}

// 3. Deterministic cleanup with pointer nullification
void buffer_free(SafeByteBuffer *buf) {
    if (buf != NULL && buf->data != NULL) {
        free(buf->data);
        buf->data = NULL;
        buf->length = 0;
        buf->capacity = 0;
    }
}

int main(void) {
    printf("=== ISO C17 Safe Bounded Buffer Demonstration ===\n");

    SafeByteBuffer buffer;
    BufferStatus status = buffer_init(&buffer, 16);
    if (status != BUFFER_OK) {
        fprintf(stderr, "Failed to initialize buffer: %d\n", status);
        return 1;
    }

    // Append safe chunks
    status = buffer_append_string(&buffer, "Header: Prumo Framework | ");
    if (status != BUFFER_OK) {
        buffer_free(&buffer);
        return 1;
    }

    status = buffer_append_string(&buffer, "Payload: Bounded Memory Invariant Verified.");
    if (status != BUFFER_OK) {
        buffer_free(&buffer);
        return 1;
    }

    printf("Buffer Content: %s\n", buffer.data);
    printf("Buffer Length: %zu, Capacity: %zu\n", buffer.length, buffer.capacity);

    // Free and verify nullification
    buffer_free(&buffer);
    printf("Buffer safely freed, pointer is NULL: %s\n", buffer.data == NULL ? "true" : "false");

    return 0;
}
