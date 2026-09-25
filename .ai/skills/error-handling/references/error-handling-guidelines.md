# Error Handling Guidelines

## 1. Exceptions for Exceptional Cases
- Use exceptions for exceptional situations (e.g., database connection failed, disk full).
- Do not use exceptions for control flow (e.g., using an exception to break out of a loop).

## 2. Catch Specific Exceptions
- Avoid bare `except:` or catching `BaseException` unless you are writing a top-level error handler that logs and crashes.
- Catch the most specific exception possible so you don't inadvertently mask bugs like `NameError` or `TypeError`.

## 3. Use the Result Pattern for Expected Failures
- For domain logic where failure is an expected outcome (e.g., user not found, invalid credentials), consider returning a `Result` object (or a union type) rather than throwing an exception.
- This forces the caller to explicitly handle the failure case.

## 4. Wrap Exceptions
- When crossing architectural boundaries, wrap low-level exceptions in domain-specific exceptions.
- For example, catch `psycopg2.OperationalError` in your repository adapter and throw `DatabaseUnavailableError` so your domain layer is unaware of the specific DB technology.

## 5. Provide Context
- When raising exceptions, provide meaningful error messages with context.
- Use `raise CustomError("message") from e` to chain exceptions in Python, preserving the original traceback.
