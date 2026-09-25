# Example: Result Pattern for Error Handling

Instead of throwing exceptions for expected domain errors, use a Result pattern to make error handling explicit.

## Implementation

```python
from typing import Generic, TypeVar, Optional

T = TypeVar('T')
E = TypeVar('E')

class Result(Generic[T, E]):
    def __init__(self, is_success: bool, value: Optional[T], error: Optional[E]):
        self.is_success = is_success
        self.value = value
        self.error = error

    @classmethod
    def ok(cls, value: T) -> 'Result[T, E]':
        return cls(True, value, None)

    @classmethod
    def fail(cls, error: E) -> 'Result[T, E]':
        return cls(False, None, error)
        
    def is_failure(self) -> bool:
        return not self.is_success
```

## Usage

```python
class InsufficientFundsError(Exception):
    pass

class BankAccount:
    def __init__(self, balance: float):
        self.balance = balance

    def withdraw(self, amount: float) -> Result[float, str]:
        if amount > self.balance:
            return Result.fail("Insufficient funds")
        
        self.balance -= amount
        return Result.ok(self.balance)

# Calling code is forced to handle the result
account = BankAccount(100)
result = account.withdraw(150)

if result.is_failure():
    print(f"Transaction failed: {result.error}")
else:
    print(f"Transaction successful. New balance: {result.value}")
```
