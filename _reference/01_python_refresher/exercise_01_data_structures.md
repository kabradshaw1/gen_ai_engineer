# Exercise 1: Data Structures — `data_structures.py`

**Goal:** Understand how Python's collections differ from Go slices/maps and TS arrays/objects.

**Environment:** Start each section in ipython to experiment, then write the final version in `data_structures.py` with your own comments.

---

## Part A: Lists vs Slices/Arrays

### Concept

**In Go** a slice is a struct: pointer to a backing array, length, and capacity. It holds values of one type contiguously in memory. `[]int{1, 2, 3}` — every element is an int, stored directly.

**In TS** an array `number[]` is typed and V8 optimizes dense arrays into contiguous memory under the hood.

**In Python** a list is an array of *pointers*. Each pointer points to an object somewhere on the heap. The objects can be any type. There is no capacity concept exposed to you.

This means:
- Lists are flexible (mixed types) but slower (pointer chasing)
- Assignment creates an alias, not a copy (both variables point to the same list object)
- Slicing always creates a new list (unlike Go where a sub-slice shares the backing array)

### Action Steps

**Step 1: Mixed types (ipython)**

Open ipython and type:
```python
items = [1, "hello", 3.14, None, [1, 2]]
type(items)
type(items[0])
type(items[1])
```

In Go, this would require `[]interface{}` (or `[]any` in 1.18+). In TS, `(number | string | null | number[])[]`. In Python, it just works — no special syntax. Think about the tradeoff: you gain flexibility but lose compile-time type safety.

**Write a comment for your .py file:** In your own words, explain why Python lists can hold mixed types. What's the underlying mechanism? (Hint: it's pointers.)

**Step 2: Reference semantics — the big gotcha (ipython)**

Type each line one at a time and predict the output before hitting Enter:
```python
a = [1, 2, 3]
b = a
b.append(4)
print(a)
```

Now try:
```python
a = [1, 2, 3]
b = a.copy()
b.append(4)
print(a)
print(b)
```

And the deep copy case:
```python
import copy
nested = [[1, 2], [3, 4]]
shallow = nested.copy()
deep = copy.deepcopy(nested)

shallow[0].append(99)
print(f"original after shallow mutate: {nested}")
print(f"shallow: {shallow}")
print(f"deep: {deep}")
```

**Write a comment for your .py file:** Explain in your own words:
- Why `b = a` doesn't copy (both are pointers to the same list object)
- What `.copy()` does and doesn't copy (new outer list, same inner objects)
- When you need `deepcopy` (when the list contains mutable objects you might change)
- How this compares to Go: assigning a slice copies the header (pointer, len, cap) but the backing array is shared. Python is even simpler — it's a straight alias.

**Step 3: Slicing (ipython)**

```python
nums = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]

# Basic slicing — same as Go s[low:high]
nums[2:5]

# Python extras — Go can't do these without helper functions
nums[::-1]       # Reverse. Go: slices.Reverse() or manual loop
nums[::2]        # Every other element. Go: manual loop
nums[-3:]        # Last 3. Go: s[len(s)-3:]
nums[1:7:2]      # From index 1 to 7, step 2. Go: no equivalent
```

Now verify that slicing creates a new list:
```python
a = [1, 2, 3, 4, 5]
b = a[1:4]
b[0] = 99
print(a)    # unchanged — b is a new list
```

In Go, `b := a[1:4]` shares the backing array. `b[0] = 99` would change `a[1]`. This is a critical difference.

**Write a comment for your .py file:** Explain the slice syntax `[start:stop:step]`. Note that `stop` is exclusive (same as Go). Explain that Python slices are always new lists — contrast with Go's shared backing array.

**Step 4: Comprehensions (ipython)**

This has no Go equivalent. TS has `.map()` and `.filter()` chains.

```python
numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

# TS: numbers.filter(x => x % 2 === 0)
# Go: manual loop with append
# Python:
evens = [x for x in numbers if x % 2 == 0]
evens

# TS: numbers.filter(x => x % 2 === 0).map(x => x * x)
# Python:
even_squares = [x * x for x in numbers if x % 2 == 0]
even_squares

# Dict comprehension — no TS/Go equivalent
scores = {"alice": 85, "bob": 92, "charlie": 78}
passing = {name: score for name, score in scores.items() if score >= 80}
passing

# Set comprehension
words = ["hello", "world", "hello", "python", "world"]
unique_lengths = {len(w) for w in words}
unique_lengths

# Nested comprehension — flatten a list of lists
# Go: two nested for loops with append
matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
flat = [val for row in matrix for val in row]
flat
```

**Write a comment for your .py file:** Explain the comprehension syntax pattern: `[expression for item in iterable if condition]`. Note the reading order: the `for` clause comes first, then the `if` filter, then the expression to produce. For nested comprehensions, the outer loop comes first (left to right), which reads like nested for loops.

---

## Part B: Dicts vs Maps/Objects

### Concept

**In Go** `map[string]int` is typed and unordered — iterating a map gives random order each time. You check existence with `if val, ok := m[key]; ok`.

**In TS** objects and `Record<string, number>` are ordered by insertion (for string keys). You check existence with `if (key in obj)` or optional chaining.

**In Python** dicts are insertion-ordered (guaranteed since 3.7) and untyped at runtime.

### Action Steps

**Step 5: Merging dicts (ipython)**

```python
defaults = {"color": "blue", "size": "medium", "verbose": False}
overrides = {"size": "large", "debug": True}

# TS: {...defaults, ...overrides}
# Go: manual loop
# Python (spread equivalent):
merged = {**defaults, **overrides}
merged

# Python 3.9+ (union operator):
merged2 = defaults | overrides
merged2
```

**Write a comment for your .py file:** Explain that `**` unpacks a dict (like TS spread `...`). Note that the right dict wins on conflicts — same as TS spread.

**Step 6: Grouping with `.setdefault()` (ipython)**

In Go, you'd write:
```go
groups := map[string][]string{}
if _, ok := groups[key]; !ok {
    groups[key] = []string{}
}
groups[key] = append(groups[key], value)
```

In Python:
```python
students = [
    ("math", "alice"),
    ("math", "bob"),
    ("science", "alice"),
    ("science", "charlie"),
    ("math", "diana"),
]

groups = {}
for subject, student in students:
    groups.setdefault(subject, []).append(student)

groups
```

Also try `defaultdict` — the more Pythonic way for this pattern:
```python
from collections import defaultdict
groups2 = defaultdict(list)
for subject, student in students:
    groups2[subject].append(student)

groups2
```

**Write a comment for your .py file:** Explain what `.setdefault(key, default)` does — returns the existing value if the key exists, otherwise inserts the default and returns it. Compare to Go's `if _, ok` pattern. Note `defaultdict` as the preferred approach when you're always grouping.

**Step 7: Sorting by value (ipython)**

```python
scores = {"alice": 85, "bob": 92, "charlie": 78, "diana": 95}

# Go: you'd implement sort.Interface or use slices.SortFunc
# TS: Object.entries(scores).sort(([,a], [,b]) => b - a)
# Python:
ranked = sorted(scores.items(), key=lambda item: item[1], reverse=True)
ranked

# Turn it back into a dict (preserves order since 3.7)
ranked_dict = dict(ranked)
ranked_dict
```

**Write a comment for your .py file:** Explain `sorted()` with `key=`. In Go, sorting requires implementing an interface or passing a comparison function. In Python, `key` extracts the value to sort by — it's more declarative. A `lambda` is an inline anonymous function (like Go's `func(a, b int) bool { ... }` or TS's arrow function `(a, b) => ...`).

---

## Part C: Sets

### Concept

**Go** has no built-in set. The idiomatic pattern is `map[T]struct{}` — you use an empty struct as the value because it takes zero bytes.

**TS** has `Set<T>` but it only supports `.has()`, `.add()`, `.delete()`. No union/intersection operators.

**Python** sets have first-class operators and are a core data structure.

### Action Steps

**Step 8: Set operations (ipython)**

```python
python_devs = {"alice", "bob", "charlie", "diana"}
ml_devs = {"charlie", "diana", "eve", "frank"}

# Intersection — who does both?
# Go: manual loop checking map existence
# TS: [...a].filter(x => b.has(x))
python_devs & ml_devs

# Union — everyone
python_devs | ml_devs

# Difference — Python only
python_devs - ml_devs

# Symmetric difference — one but not both
python_devs ^ ml_devs

# Deduplication — the most common use case
words = ["the", "cat", "sat", "on", "the", "mat", "the"]
unique = set(words)
unique
len(words), len(unique)
```

**Write a comment for your .py file:** Explain each operator (`&`, `|`, `-`, `^`). Note that sets are unordered and only hold hashable (immutable) elements — you can't put a list in a set (`{[1, 2]}` fails) but you can put a tuple (`{(1, 2)}` works). Compare to Go where you'd use `map[T]struct{}` for all of this.

---

## Part D: Generators

### Concept

**In Go** the closest analogy is a goroutine sending values on a channel. The goroutine produces values; the `range` loop consumes them. The producer and consumer run concurrently.

**In Python** a generator is the same idea but single-threaded. `yield` produces a value and pauses the function. The consumer pulls the next value, which resumes the function. No concurrency — it's cooperative.

### Action Steps

**Step 9: Generator function (ipython)**

```python
def fibonacci(n):
    a, b = 0, 1
    for _ in range(n):
        yield a          # pauses here, returns a
        a, b = b, a + b  # resumes here on next call

# Create the generator — nothing runs yet
gen = fibonacci(10)
gen                       # <generator object fibonacci at 0x...>
type(gen)

# Pull values one at a time
next(gen)   # 0
next(gen)   # 1
next(gen)   # 1
next(gen)   # 2

# Or consume all remaining values
list(gen)   # the rest — [3, 5, 8, 13, 21, 34]

# Try to get more — it's exhausted
# next(gen)  # StopIteration error — like a closed channel
```

In Go this would be:
```go
func fibonacci(n int, ch chan int) {
    a, b := 0, 1
    for i := 0; i < n; i++ {
        ch <- a
        a, b = b, a+b
    }
    close(ch)
}
```

The key difference: Go's version runs on a separate goroutine. Python's generator runs on the same thread, resuming only when you ask for the next value.

**Step 10: Generator expressions (ipython)**

```python
# List comprehension — builds the whole list in memory
squares_list = [x**2 for x in range(1000000)]

# Generator expression — lazy, builds nothing until asked
squares_gen = (x**2 for x in range(1000000))

# Check memory difference
import sys
sys.getsizeof(squares_list)   # ~8MB
sys.getsizeof(squares_gen)    # ~200 bytes — just the generator object
```

Generator expressions use `()` instead of `[]`. Same syntax, completely different behavior. Use generators when you don't need all values at once — just like you'd use a channel in Go when you don't need all values in a slice.

**Step 11: Chaining generators (ipython)**

```python
def evens(iterable):
    for x in iterable:
        if x % 2 == 0:
            yield x

def doubled(iterable):
    for x in iterable:
        yield x * 2

# Chain them — like piping channels
numbers = range(20)
result = doubled(evens(numbers))
list(result)
```

Each generator is a processing stage. Values flow through one at a time, never all held in memory at once. In Go, you'd wire this up with channels. In TS, you'd chain `.filter().map()` on an array (but that builds intermediate arrays).

**Write comments for your .py file:** Explain:
- `yield` pauses the function and produces a value (like sending on a channel)
- Generators are lazy — they compute values on demand, not all upfront
- You can only iterate a generator once (like reading from a channel — once consumed, it's done)
- Generator expressions `()` vs list comprehensions `[]` — same syntax, lazy vs eager
- Chaining generators is like piping data through stages without intermediate storage

---

## Writing Your .py File

Now that you've explored everything in ipython, create `data_structures.py`. Here's the structure to follow:

```python
"""Data Structures: Python vs Go/TypeScript

[Write 2-3 sentences in your own words about the fundamental difference
between Python lists and Go slices / TS arrays.]
"""


def list_demos():
    """Lists — reference types, flexible, pointer-based."""
    # [Your comment: explain what a Python list actually is under the hood]

    # Reference semantics
    # [Your comment: explain why b = a doesn't copy, and how this
    #  differs from Go slice assignment]

    # Slicing
    # [Your comment: explain [start:stop:step] and why Python slices
    #  create new lists unlike Go sub-slices]

    # Comprehensions
    # [Your comment: explain the syntax pattern and how this replaces
    #  map/filter in TS and for-loops in Go]


def dict_demos():
    """Dicts — ordered, untyped, with powerful merging."""
    # [Your comment: how dicts differ from Go maps (ordered vs random)
    #  and TS objects]

    # Merging
    # [Your comment: ** unpacking, like TS spread]

    # Grouping
    # [Your comment: setdefault vs Go's if _, ok pattern]

    # Sorting by value
    # [Your comment: sorted() with key= lambda]


def set_demos():
    """Sets — first-class, unlike Go's map[T]struct{} workaround."""
    # [Your comment: operators &, |, -, ^ and when to use sets]


def generator_demos():
    """Generators — like Go channels but single-threaded and lazy."""
    # [Your comment: yield pauses the function, values produced on demand]

    # Generator function

    # Generator expression vs list comprehension

    # Chaining generators


if __name__ == "__main__":
    list_demos()
    dict_demos()
    set_demos()
    generator_demos()
```

**Rules for your comments:**
- Don't copy the explanations from this guide word-for-word. Restate them in how *you* understand it.
- If you can't explain it in your own words yet, go back to ipython and experiment more.
- Include the Go/TS comparison where it helps your understanding.
- Each comment block should be something a hiring manager could read and think "this person actually understands the difference."

---

## Action Checklist

- [ ] Open ipython and work through Steps 1-11, typing each example
- [ ] For each step, predict the output before running it
- [ ] When something surprises you, experiment until it doesn't
- [ ] Create `data_structures.py` following the structure above
- [ ] Write all comments in your own words
- [ ] Run `python data_structures.py` and verify it works top to bottom
- [ ] Read through your comments — could you explain each concept to someone else?

When this feels solid, move on to Exercise 2.
