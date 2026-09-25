//! Production-grade Rust demonstrating custom error handling, safe concurrency,
//! memory boundary encapsulation, and explicit safety proofs.

use std::fmt;
use std::ptr::NonNull;
use std::sync::{Arc, Mutex};
use std::thread;

/// Custom domain error enum without external crate dependencies for universal standalone compilation.
#[derive(Debug, PartialEq, Eq)]
pub enum DataError {
    EmptyInput,
    InvalidPayload(String),
    BufferOverflow { index: usize, capacity: usize },
    LockPoisoned,
}

impl fmt::Display for DataError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::EmptyInput => write!(f, "input data slice cannot be empty"),
            Self::InvalidPayload(msg) => write!(f, "invalid payload format: {}", msg),
            Self::BufferOverflow { index, capacity } => {
                write!(f, "index {} out of bounds for buffer capacity {}", index, capacity)
            }
            Self::LockPoisoned => write!(f, "synchronization mutex was poisoned by a panicked thread"),
        }
    }
}

impl std::error::Error for DataError {}

/// A memory-safe, bounds-checked wrapper encapsulating a raw buffer.
pub struct AuditedBuffer<T> {
    ptr: NonNull<T>,
    len: usize,
    capacity: usize,
}

impl<T> AuditedBuffer<T> {
    /// Allocates an audited buffer with a specified initial capacity.
    pub fn with_capacity(capacity: usize) -> Self {
        assert!(capacity > 0, "Capacity must be strictly greater than zero");
        let mut vec = Vec::with_capacity(capacity);
        let ptr = NonNull::new(vec.as_mut_ptr()).expect("Vec allocation cannot produce null pointer");
        std::mem::forget(vec); // Manual ownership transfer

        Self {
            ptr,
            len: 0,
            capacity,
        }
    }

    /// Appends an element into the buffer, checking capacity invariants.
    pub fn push(&mut self, item: T) -> Result<(), DataError> {
        if self.len >= self.capacity {
            return Err(DataError::BufferOverflow {
                index: self.len,
                capacity: self.capacity,
            });
        }

        // SAFETY:
        // 1. self.len < self.capacity, guaranteeing self.ptr.add(self.len) is within the allocated block.
        // 2. self.ptr is a valid, aligned NonNull pointer derived from Vec::with_capacity.
        // 3. We use ptr::write to prevent dropping uninitialized memory.
        unsafe {
            let target = self.ptr.as_ptr().add(self.len);
            std::ptr::write(target, item);
        }
        self.len += 1;
        Ok(())
    }

    /// Returns a shared slice to the initialized elements.
    pub fn as_slice(&self) -> &[T] {
        // SAFETY:
        // 1. self.ptr points to self.len consecutive initialized elements of type T.
        // 2. The lifetime is bound to &'self, preventing concurrent mutation or reallocation.
        unsafe {
            std::slice::from_raw_parts(self.ptr.as_ptr(), self.len)
        }
    }

    pub fn len(&self) -> usize {
        self.len
    }

    pub fn is_empty(&self) -> bool {
        self.len == 0
    }
}

impl<T> Drop for AuditedBuffer<T> {
    fn drop(&mut self) {
        // SAFETY:
        // 1. Reconstruct Vec to invoke standard deallocation and drop all active elements.
        // 2. self.ptr, self.len, and self.capacity were created together from a Vec allocation.
        unsafe {
            let _ = Vec::from_raw_parts(self.ptr.as_ptr(), self.len, self.capacity);
        }
    }
}

// Send and Sync markers proven safe because access across threads is serialized by Rust type system
unsafe impl<T: Send> Send for AuditedBuffer<T> {}
unsafe impl<T: Sync> Sync for AuditedBuffer<T> {}

/// Thread-safe concurrent worker coordinator
pub struct ConcurrentWorkerPool {
    storage: Arc<Mutex<AuditedBuffer<u64>>>,
}

impl ConcurrentWorkerPool {
    pub fn new(capacity: usize) -> Self {
        Self {
            storage: Arc::new(Mutex::new(AuditedBuffer::with_capacity(capacity))),
        }
    }

    pub fn execute_parallel_work(&self, worker_count: usize, items_per_worker: u64) -> Result<usize, DataError> {
        let mut handles = Vec::with_capacity(worker_count);

        for w in 0..worker_count {
            let storage_clone = Arc::clone(&self.storage);
            handles.push(thread::spawn(move || -> Result<(), DataError> {
                for i in 0..items_per_worker {
                    let val = (w as u64) * 1000 + i;
                    let mut lock = storage_clone.lock().map_err(|_| DataError::LockPoisoned)?;
                    lock.push(val)?;
                }
                Ok(())
            }));
        }

        for handle in handles {
            let thread_res = handle.join().map_err(|_| DataError::LockPoisoned)?;
            thread_res?;
        }

        let lock = self.storage.lock().map_err(|_| DataError::LockPoisoned)?;
        Ok(lock.len())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_audited_buffer_basic() {
        let mut buf = AuditedBuffer::with_capacity(4);
        assert!(buf.is_empty());
        assert_eq!(buf.push(42), Ok(()));
        assert_eq!(buf.push(99), Ok(()));
        assert_eq!(buf.as_slice(), &[42, 99]);
        assert_eq!(buf.len(), 2);
    }

    #[test]
    fn test_concurrent_worker_pool() {
        let pool = ConcurrentWorkerPool::new(100);
        let total = pool.execute_parallel_work(4, 10).expect("Workers must finish without poison");
        assert_eq!(total, 40);
    }
}
