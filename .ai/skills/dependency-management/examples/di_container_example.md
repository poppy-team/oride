# Example: Constructor Injection

This is the standard and preferred way to manage dependencies.

## The Interfaces and Implementations

```python
from abc import ABC, abstractmethod

class EmailSender(ABC):
    @abstractmethod
    def send(self, to: str, message: str) -> None:
        pass

class SmtpEmailSender(EmailSender):
    def send(self, to: str, message: str) -> None:
        print(f"Sending via SMTP to {to}: {message}")

class MockEmailSender(EmailSender):
    def send(self, to: str, message: str) -> None:
        print(f"Mock send to {to}")
```

## The Service

```python
class NotificationService:
    # Dependency is injected via constructor
    def __init__(self, email_sender: EmailSender):
        self.email_sender = email_sender

    def notify_user(self, user_email: str) -> None:
        self.email_sender.send(user_email, "Hello!")
```

## The Composition Root (main.py)

```python
def main():
    # Environment variable check or config
    is_prod = False
    
    # Wire dependencies
    if is_prod:
        sender = SmtpEmailSender()
    else:
        sender = MockEmailSender()
        
    service = NotificationService(sender)
    
    # Run application
    service.notify_user("test@example.com")

if __name__ == "__main__":
    main()
```
