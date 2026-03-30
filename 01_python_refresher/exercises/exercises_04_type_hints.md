# Exercises: Type Hints

After completing the reference notebook, test yourself with these.

## ipython Exercises

Type each into ipython. **Predict the output BEFORE you hit enter.**

### 1. Runtime doesn't care

```python
def strict_add(a: int, b: int) -> int:
    return a + b

print(strict_add([1, 2], [3, 4]))
print(type(strict_add([1, 2], [3, 4])))
```

### 2. isinstance vs type hints

```python
from typing import Protocol

class Printable(Protocol):
    def __str__(self) -> str: ...

x = 42
print(isinstance(x, Printable))
```

### 3. TypedDict at runtime

```python
from typing import TypedDict

class Config(TypedDict):
    host: str
    port: int

c: Config = {"host": "localhost", "port": 8080}
print(type(c))
print(c.__class__.__name__)
c["extra"] = True
print(len(c))
```

### 4. TypeVar identity

```python
from typing import TypeVar

T = TypeVar('T')

def identity(x: T) -> T:
    return x

result = identity("hello")
print(result)
print(type(result).__name__)
result2 = identity(42)
print(result2)
```

### 5. get_type_hints introspection

```python
from typing import get_type_hints, Optional

def process(name: str, age: int, email: Optional[str] = None) -> bool:
    return True

hints = get_type_hints(process)
print(list(hints.keys()))
print(hints['return'].__name__)
```

### 6. Annotations are just data

```python
x: int = "not an int"
y: str = 42

print(x, y)
print(__annotations__)
print(type(__annotations__['x']).__name__)
```

## .py Challenge

Create `type_safe.py` that produces this exact output:

```
=== Type Registry ===

Registered: IntValidator (validates: int)
Registered: StringValidator (validates: str)
Registered: ListValidator (validates: list)

Validating 42 with IntValidator: ✓ valid
Validating "hello" with IntValidator: ✗ invalid (expected int, got str)
Validating "hello" with StringValidator: ✓ valid
Validating [1, 2] with ListValidator: ✓ valid
Validating (1, 2) with ListValidator: ✗ invalid (expected list, got tuple)

Registry contents:
  int -> IntValidator
  str -> StringValidator
  list -> ListValidator

Lookup validator for int: IntValidator
Lookup validator for float: None (not registered)
```

Notes: Must pass `mypy --strict` with no errors. Use Protocol, TypeVar, or generics — not just isinstance checks with string formatting.

No hints. No function signatures. Figure it out from the output.
