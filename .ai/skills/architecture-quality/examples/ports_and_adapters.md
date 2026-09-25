# Example: Ports and Adapters (Hexagonal Architecture)

## The Port (Domain Interface)

```python
from abc import ABC, abstractmethod
from typing import List
from domain.entities import User

class UserRepository(ABC):
    @abstractmethod
    def save(self, user: User) -> None:
        pass

    @abstractmethod
    def find_by_id(self, user_id: str) -> User | None:
        pass
```

## The Use Case (Application Layer)

```python
from domain.entities import User
from application.ports import UserRepository

class RegisterUserUseCase:
    def __init__(self, user_repository: UserRepository):
        self.user_repository = user_repository

    def execute(self, user_id: str, email: str, password: str) -> User:
        if self.user_repository.find_by_id(user_id):
            raise ValueError("User already exists")
            
        user = User(id=user_id, email=email, password=password)
        self.user_repository.save(user)
        return user
```

## The Adapter (Infrastructure Layer)

```python
from application.ports import UserRepository
from domain.entities import User
from infrastructure.database import Session

class PostgresUserRepository(UserRepository):
    def __init__(self, session: Session):
        self.session = session

    def save(self, user: User) -> None:
        # SQL/ORM logic to save user
        pass

    def find_by_id(self, user_id: str) -> User | None:
        # SQL/ORM logic to find user
        pass
```
