# Exercise 4: Type Hints — `type_hints.py`

**Goal:** Understand Python's type system vs Go's compile-time types and TS's structural types.

**Environment:** Start in ipython for the shock value (Step 1), then work in `type_hints.py`.

---

## The Fundamental Shift

**Go:** Types are checked at compile time. Wrong type → won't compile. Period.

**TypeScript:** Types are checked at compile time by tsc. Wrong type → red squiggle, won't compile. Very similar to Go in this regard.

**Python:** Type hints are **completely ignored at runtime.** The interpreter doesn't read them, doesn't check them, doesn't care about them. They exist for:
1. Humans reading the code
2. Static analysis tools (`mypy`, `pyright`, IDE support)
3. Documentation generators

This is the hardest adjustment coming from Go or TS. You'll write `def add(a: int, b: int) -> int:` and Python will happily let you pass strings. The hints are a suggestion, not a contract.

---

## Part A: Hints Don't Enforce Anything

### Action Steps

**Step 1: Runtime proof (ipython)**

```python
def add(a: int, b: int) -> int:
    return a + b

# This "should" be an error. In Go and TS, it would be.
result = add("hello", " world")
print(result)     # "hello world" — no error!
print(type(result))  # <class 'str'>
```

Sit with this for a moment. In Go, `func add(a int, b int) int` means the compiler will reject `add("hello", " world")`. In TS, `function add(a: number, b: number): number` means tsc will reject it. In Python, it just runs.

Now try:
```python
def divide(a: int, b: int) -> float:
    return a / b

divide("hello", "world")   # TypeError — but from the OPERATION, not the hint
```

The error here isn't because of the type hint. It's because `/` doesn't work on strings. Python checked nothing about the hint — the operation itself failed.

**Write a comment for your .py file:** Explain that type hints are metadata, not contracts. Python's runtime ignores them entirely. Errors come from operations failing, not from type mismatches. Compare to Go/TS where the compiler rejects type mismatches before the code ever runs.

**Step 2: Modern annotation syntax (ipython)**

```python
# Old style (Python 3.8 and earlier — you'll see this in older code)
from typing import List, Dict, Optional, Tuple

def old_style(items: List[str], config: Dict[str, int]) -> Optional[Tuple[str, int]]:
    pass

# Modern style (Python 3.10+ — use this)
def modern_style(items: list[str], config: dict[str, int]) -> tuple[str, int] | None:
    pass
```

Key differences from Go/TS:
- `list[str]` instead of `[]string` (Go) or `string[]` (TS)
- `dict[str, int]` instead of `map[string]int` (Go) or `Record<string, number>` (TS)
- `str | None` instead of `*string` / nullable (Go) or `string | null` (TS)
- `tuple[str, int]` — fixed-size typed sequence. Go: no equivalent. TS: `[string, number]`

**Write a comment:** Show the modern syntax for common types. Note the `|` union operator (Python 3.10+) replaces `Optional` from the `typing` module.

---

## Part B: Protocol — Python's Go Interfaces

### Concept

This is the closest Python gets to Go's structural interfaces.

**Go interfaces** are structural: any type with the right methods satisfies the interface. You never write `implements`. The compiler checks it.

**Python ABCs** (Exercise 2) are nominal: you must explicitly inherit. Runtime error if you forget.

**Python Protocol** is structural: any class with the right methods satisfies it. No inheritance needed. But — and this is key — it's only checked by static analysis tools (`mypy`), not at runtime.

### Action Steps

**Step 3: Define a Protocol (ipython first, then .py file)**

```python
from typing import Protocol


class Embeddable(Protocol):
    """Any object that can produce an embedding vector.

    This is like a Go interface:
        type Embeddable interface {
            ToEmbedding() []float64
        }
    """
    def to_embedding(self) -> list[float]: ...
```

The `...` (Ellipsis) means "no implementation." It's like an interface method declaration.

**Step 4: Classes that satisfy the Protocol — no inheritance needed**

```python
class TextChunk:
    """Satisfies Embeddable without knowing Embeddable exists."""
    def __init__(self, text: str):
        self.text = text

    def to_embedding(self) -> list[float]:
        return [float(ord(c)) for c in self.text[:5]]


class ImageFeature:
    """Also satisfies Embeddable — completely unrelated class."""
    def __init__(self, label: str, features: list[float]):
        self.label = label
        self.features = features

    def to_embedding(self) -> list[float]:
        return self.features
```

Neither class inherits from `Embeddable`. Neither class even imports `Embeddable`. They just happen to have a `to_embedding()` method. This is exactly how Go interfaces work.

**Step 5: Use the Protocol as a type hint**

```python
def cosine_similarity(a: Embeddable, b: Embeddable) -> float:
    """Works with any Embeddable — TextChunk, ImageFeature, or anything
    else that has to_embedding(). Just like accepting an interface in Go."""
    vec_a = a.to_embedding()
    vec_b = b.to_embedding()
    dot = sum(x * y for x, y in zip(vec_a, vec_b))
    mag_a = sum(x**2 for x in vec_a) ** 0.5
    mag_b = sum(x**2 for x in vec_b) ** 0.5
    if mag_a == 0 or mag_b == 0:
        return 0.0
    return dot / (mag_a * mag_b)

# Both work — structural subtyping
chunk = TextChunk("hello")
image = ImageFeature("cat", [104.0, 101.0, 108.0, 108.0, 111.0])
cosine_similarity(chunk, image)
```

But remember: Python doesn't check this at runtime. You could pass an `int` and it would only fail when it tries to call `.to_embedding()`. To get compile-time-like checking, run `mypy`:

```bash
pip install mypy
mypy type_hints.py
```

**Write a comment:** Explain Protocol as Python's version of Go interfaces. Structural, not nominal. Checked by mypy, not the runtime. Compare to ABCs which require explicit inheritance.

---

## Part C: TypeVar and Generics

### Concept

**Go (1.18+):** `func First[T any](items []T) T` — generic function, checked at compile time.

**TS:** `function first<T>(items: T[]): T` — generic function, checked at compile time.

**Python:** More verbose declaration, and the generics are only checked by mypy, not at runtime.

### Action Steps

**Step 6: Generic functions (ipython)**

```python
from typing import TypeVar

T = TypeVar("T")

def first_or_default(items: list[T], default: T) -> T:
    """Return first item or default.

    Go:  func First[T any](items []T, def T) T
    TS:  function first<T>(items: T[], def: T): T
    """
    return items[0] if items else default


# Type is preserved through the generic
first_or_default([1, 2, 3], 0)          # int
first_or_default(["a", "b"], "none")    # str
first_or_default([], "fallback")         # str
```

Notice `T = TypeVar("T")` is a separate line. In Go/TS, the type parameter is declared inline. Python requires this separate declaration because the type system was added later — it's bolted on, not built in.

**Step 7: Generic function — chunk_list (ipython)**

```python
def chunk_list(items: list[T], size: int) -> list[list[T]]:
    """Split list into chunks.

    Go: func Chunk[T any](items []T, size int) [][]T
    """
    return [items[i:i + size] for i in range(0, len(items), size)]

chunk_list([1, 2, 3, 4, 5, 6, 7], 3)
# [[1, 2, 3], [4, 5, 6], [7]]
```

**Step 8: Generic class (ipython)**

```python
from typing import Generic
from dataclasses import dataclass

@dataclass
class Result(Generic[T]):
    """A typed result wrapper.

    Go:  type Result[T any] struct { Value T; Confidence float64 }
    TS:  interface Result<T> { value: T; confidence: number }
    """
    value: T
    confidence: float

r1: Result[str] = Result(value="positive", confidence=0.95)
r2: Result[list[str]] = Result(value=["python", "nlp"], confidence=0.72)

# Python doesn't stop this — no runtime checking
r3: Result[int] = Result(value="not an int", confidence=0.5)  # No error!
```

That last line is the point. In Go, `Result[int]{Value: "not an int"}` wouldn't compile. In Python, it runs fine. The hint `Result[int]` is a suggestion for mypy, not a constraint.

**Write a comment:** Explain TypeVar and Generic. Note the extra verbosity vs Go/TS. Emphasize that generics are enforced by mypy, not Python itself.

---

## Part D: Running mypy

### Action Steps

**Step 9: Static type checking**

After writing your `type_hints.py`, run:
```bash
pip install mypy
mypy type_hints.py
```

Intentionally add a type error:
```python
x: int = "hello"
```

Run `mypy` again — it should catch it. Then run `python type_hints.py` — it runs without error. This demonstrates the gap: mypy catches it, Python doesn't.

**Write a comment:** Explain that mypy is Python's equivalent of `go vet` + compiler type checking. It's optional, external, and not part of the language runtime. In a production codebase, you'd run it in CI.

---

## Writing Your .py File

```python
"""Type Hints: Python vs Go/TypeScript

[Your summary: Python type hints are metadata, not contracts.
The runtime ignores them. mypy is the optional enforcer.]
"""

from typing import Protocol, TypeVar, Generic
from dataclasses import dataclass

# [Your comment: runtime proof — hints don't prevent wrong types]


# --- Protocol (structural interfaces) ---

# [Your comment: Protocol ≈ Go interfaces. Structural, no inheritance needed.
#  Only checked by mypy, not at runtime.]


# --- TypeVar and Generic functions ---

# [Your comment: More verbose than Go/TS generics.
#  TypeVar requires separate declaration.]

T = TypeVar("T")


# --- Generic class ---

# [Your comment: Generic[T] in class definition.
#  Go: type Result[T any] struct {}
#  TS: class Result<T> {}]


def demo():
    # Demonstrate runtime proof
    # Demonstrate Protocol with two unrelated classes
    # Demonstrate generic functions
    # Demonstrate generic class
    pass


if __name__ == "__main__":
    demo()
```

---

## Action Checklist

- [ ] Step 1 in ipython: prove to yourself that hints don't enforce anything
- [ ] Build your Protocol + two classes that satisfy it
- [ ] Write generic functions with TypeVar
- [ ] Write a Generic class
- [ ] Run `mypy type_hints.py` — see what it catches
- [ ] Intentionally add type errors, run mypy vs python — see the difference
- [ ] Write all comments in your own words
- [ ] Run `python type_hints.py` end to end

When this feels solid, move on to Exercise 5.
