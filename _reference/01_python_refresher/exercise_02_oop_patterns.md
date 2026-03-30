# Exercise 2: OOP Patterns — `oop_patterns.py`

**Goal:** Understand Python's class system vs Go structs+interfaces and TS classes.

**Environment:** Explore in ipython, then write `oop_patterns.py` with your own comments.

---

## Part A: Classes vs Structs

### Concept

**In Go** there are no classes. You define a struct for data, then attach methods with receivers:
```go
type Document struct {
    Title   string
    Content string
}

func (d Document) WordCount() int {
    return len(strings.Fields(d.Content))
}
```

**In TS** you have `class` with explicit visibility modifiers:
```typescript
class Document {
    constructor(public title: string, private content: string) {}
    wordCount(): number { return this.content.split(' ').length }
}
```

**In Python** everything about classes is more explicit and more open:
- `self` is always the first parameter (Go's receiver, but written inside the method signature)
- No visibility modifiers — everything is public
- `__init__` is the constructor (like Go's `NewDocument()` constructor function)
- No compilation step — errors are caught at runtime

### Action Steps

**Step 1: Basic class (ipython)**

Type this line by line:
```python
class Document:
    def __init__(self, title: str, content: str):
        self.title = title
        self.content = content

    def word_count(self) -> int:
        return len(self.content.split())

    def __repr__(self) -> str:
        return f"Document({self.title!r}, {len(self.content)} chars)"

doc = Document("My Doc", "hello world foo bar")
doc
doc.title
doc.word_count()
```

Now try forgetting `self`:
```python
class Broken:
    def greet(name):   # forgot self!
        return f"Hello {name}"

b = Broken()
b.greet("Kyle")   # TypeError — Python passed the instance as 'name'
```

Read the error message carefully. Python automatically passes the instance as the first argument. If you don't name that parameter `self`, it gets used as whatever the first parameter is.

In Go, the receiver is *outside* the function signature: `func (d Document) WordCount()`. In Python, it's *inside*: `def word_count(self)`. Both are explicit, just positioned differently. TS's `this` is implicit — you never declare it.

**Write a comment for your .py file:** Explain `self` — what it is, why it's explicit, and how it compares to Go receivers and TS's `this`.

**Step 2: No private fields (ipython)**

```python
class User:
    def __init__(self, name: str, password: str):
        self.name = name
        self._password = password        # convention: "private"
        self.__secret = "hidden"          # name mangling

u = User("Kyle", "hunter2")
u.name           # works — public
u._password      # works — "private" by convention only
u.__secret       # AttributeError
u._User__secret  # works — Python just mangled the name
```

In Go, lowercase = unexported (enforced by the compiler). In TS, `private` is enforced by tsc. In Python, `_prefix` is a *suggestion*. Nothing stops you from accessing it. `__double_prefix` does name mangling (prepends `_ClassName`), but even that is accessible if you know the mangled name.

**Write a comment for your .py file:** Explain the `_` and `__` conventions. Note that Python trusts developers rather than enforcing access control — "we're all consenting adults here." Compare to Go's export rules.

**Step 3: Dataclasses (ipython)**

```python
from dataclasses import dataclass, field

@dataclass
class Config:
    name: str
    debug: bool = False
    tags: list[str] = field(default_factory=list)

c1 = Config("dev", debug=True, tags=["fast"])
c2 = Config("dev", debug=True, tags=["fast"])

print(c1)          # free __repr__: Config(name='dev', debug=True, tags=['fast'])
print(c1 == c2)    # free __eq__: True (compares all fields)
```

Without `@dataclass`, you'd have to write `__init__`, `__repr__`, and `__eq__` yourself. A dataclass auto-generates them from the field annotations.

Compare to Go: a struct automatically supports `==` (if fields are comparable) and `fmt.Printf("%+v", s)` for printing. Python's regular classes don't — you need `@dataclass` or manual dunder methods.

Compare to TS: similar to declaring a class with all public readonly fields, but with auto-generated equality.

**Gotcha — mutable default arguments:**
```python
# WRONG — this shares one list across all instances
@dataclass
class Bad:
    items: list[str] = []      # This will error — Python catches this

# RIGHT — field(default_factory=list) creates a new list per instance
@dataclass
class Good:
    items: list[str] = field(default_factory=list)
```

In Go, each struct instance gets its own zero-value slice. Python needs `default_factory` because default values are evaluated once at class definition time, not per-instance.

**Write a comment for your .py file:** Explain what `@dataclass` generates for you. Explain `field(default_factory=list)` and why you can't use `= []` as a default. Compare to Go struct defaults and TS constructor shorthand.

---

## Part B: Inheritance vs Composition/Embedding

### Concept

**Go** has no inheritance. You embed structs and get method forwarding:
```go
type Animal struct { Name string }
func (a Animal) Speak() string { return "..." }

type Dog struct { Animal }  // embedding, not inheritance
```

**TS** has `extends` for single inheritance.

**Python** has full inheritance, including multiple inheritance (which Go and TS don't have).

### Action Steps

**Step 4: Basic inheritance (ipython)**

```python
class Animal:
    def __init__(self, name: str):
        self.name = name

    def speak(self) -> str:
        return "..."

class Dog(Animal):
    def speak(self) -> str:
        return f"{self.name} says Woof!"

class Cat(Animal):
    def speak(self) -> str:
        return f"{self.name} says Meow!"

dog = Dog("Rex")
cat = Cat("Whiskers")
dog.speak()
cat.speak()
isinstance(dog, Animal)   # True — like Go's interface satisfaction, but nominal
```

**Step 5: super() (ipython)**

```python
class Animal:
    def __init__(self, name: str):
        self.name = name

class Dog(Animal):
    def __init__(self, name: str, breed: str):
        super().__init__(name)    # calls Animal.__init__
        self.breed = breed

d = Dog("Rex", "Lab")
d.name    # from Animal
d.breed   # from Dog
```

In Go with embedding, you'd access the embedded struct directly: `d.Animal.Name`. In Python, `super()` walks the inheritance chain. They solve the same problem — reusing parent behavior — but the mechanism is different.

**Step 6: Multiple inheritance and MRO (ipython)**

```python
class Flyable:
    def move(self):
        return "flying"

class Swimmable:
    def move(self):
        return "swimming"

class Duck(Flyable, Swimmable):
    pass

d = Duck()
d.move()          # "flying" — Flyable comes first in the class definition

Duck.__mro__      # Method Resolution Order — shows the lookup chain
```

Go can't do this — embedding two structs with the same method name is a compile error. TS can't do this — single inheritance only (though mixins exist). Python resolves it with MRO (C3 linearization). The class listed first wins.

**Write a comment for your .py file:** Explain inheritance vs Go embedding. Note that `super()` walks the MRO, not just "the parent." Mention multiple inheritance exists but composition is usually preferred in practice.

---

## Part C: ABCs vs Interfaces

### Concept

**Go interfaces** are structural: if a type has the right methods, it satisfies the interface automatically. You never write `implements`.

**Python ABCs** are nominal: you must explicitly inherit from the ABC. If you forget to implement an abstract method, you get a runtime error when you try to instantiate.

(Python's `Protocol` — covered in Exercise 4 — is the structural equivalent of Go interfaces.)

### Action Steps

**Step 7: Define and use an ABC (ipython)**

```python
from abc import ABC, abstractmethod

class Processor(ABC):
    @abstractmethod
    def process(self, text: str) -> str:
        ...

    @abstractmethod
    def name(self) -> str:
        ...
```

Now try to instantiate it:
```python
p = Processor()   # TypeError: Can't instantiate abstract class
```

In Go, you can't instantiate an interface either — but you'd never try because interfaces aren't types you construct. In Python, someone could try `Processor()` and would get a runtime error.

Implement a subclass that's missing a method:
```python
class Incomplete(Processor):
    def process(self, text: str) -> str:
        return text.upper()
    # forgot name()!

i = Incomplete()   # TypeError — name() not implemented
```

In Go, this would be a compile error: "Incomplete does not implement Processor (missing method name)." In Python, it's a runtime error — you only discover it when you try to create an instance.

Now implement it correctly:
```python
class Uppercaser(Processor):
    def process(self, text: str) -> str:
        return text.upper()

    def name(self) -> str:
        return "Uppercaser"

u = Uppercaser()
u.process("hello")
u.name()
isinstance(u, Processor)   # True — nominal subtyping
```

**Write a comment for your .py file:** Explain ABCs — they're like Go interfaces but nominal (must inherit explicitly) and errors are caught at runtime, not compile time. Note that Exercise 4 covers `Protocol` which is the Python equivalent of Go's structural interfaces.

---

## Writing Your .py File

Create `oop_patterns.py` with this structure:

```python
"""OOP Patterns: Python vs Go/TypeScript

[Your own summary: how Python classes differ from Go structs
and TS classes. What surprised you most?]
"""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field


# --- Part A: Classes ---

# [Your comment: explain self, compare to Go receivers and TS this]

# [Your comment: explain _ and __ privacy conventions, compare to Go exports]

# [Your comment: explain @dataclass — what it generates, why default_factory]


# --- Part B: Inheritance ---

# [Your comment: inheritance vs Go embedding, super() and MRO]


# --- Part C: ABCs ---

# [Your comment: ABCs are nominal interfaces, errors at runtime not compile time]


def demo():
    """Run all demonstrations."""
    # Build your classes above, then demonstrate them here
    pass


if __name__ == "__main__":
    demo()
```

**Rules for your comments:**
- Rewrite every explanation in your own words
- Include the Go/TS comparison where it clarifies the concept
- If you can't explain something clearly, go back to ipython and experiment

---

## Action Checklist

- [ ] Work through Steps 1-7 in ipython
- [ ] Intentionally trigger each error (missing self, instantiating ABC, incomplete subclass)
- [ ] Read each error message — know what it looks like so you recognize it
- [ ] Create `oop_patterns.py` with your own classes and comments
- [ ] Run `python oop_patterns.py` and verify it works
- [ ] Review your comments — would they help another Go/TS developer understand Python OOP?

When this feels solid, move on to Exercise 3.
