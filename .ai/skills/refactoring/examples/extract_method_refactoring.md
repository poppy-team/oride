# Example: Extract Method Refactoring

## Before

```python
def print_owing(name: str, amounts: list[float]):
    outstanding = 0.0

    # print banner
    print("*************************")
    print("***** Customer Owes *****")
    print("*************************")

    # calculate outstanding
    for amount in amounts:
        outstanding += amount

    # print details
    print(f"name: {name}")
    print(f"amount: {outstanding}")
```

## After

```python
def print_owing(name: str, amounts: list[float]):
    print_banner()
    outstanding = calculate_outstanding(amounts)
    print_details(name, outstanding)

def print_banner():
    print("*************************")
    print("***** Customer Owes *****")
    print("*************************")

def calculate_outstanding(amounts: list[float]) -> float:
    return sum(amounts)

def print_details(name: str, outstanding: float):
    print(f"name: {name}")
    print(f"amount: {outstanding}")
```
