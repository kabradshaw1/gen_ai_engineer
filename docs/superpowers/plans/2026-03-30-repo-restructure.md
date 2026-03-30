# Repo Restructure: Jupyter Notebook-Based Learning — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restructure the gen_ai_engineer repo from markdown-guide + `.py` file workflow to self-contained Jupyter notebooks, archiving all existing content.

**Architecture:** Move all existing content into `_reference/`. Create fresh `01_python_refresher/` with 6 reference notebooks (00-05). Update CLAUDE.md and README.md for the new approach.

**Tech Stack:** Jupyter notebooks, Python 3.11, Miniconda (conda env: `gen_ai`)

---

## File Structure

```
_reference/                         # Archive of all existing content
  01_python_refresher/              # Existing .md guides, .py files
  02_demo_of_gv/
  03_nlp_fundamentals/
  04_rag_app/
  Inital.md
  README.md
  CLAUDE.md
CLAUDE.md                           # Updated for new approach
README.md                           # Rewritten for new structure
.gitignore                          # Unchanged
01_python_refresher/
  README.md                         # New, short intro
  requirements.txt                  # Updated
  00_environments.ipynb             # Reference notebook
  01_data_structures.ipynb          # Reference notebook
  02_oop_patterns.ipynb             # Reference notebook
  03_async_basics.ipynb             # Reference notebook
  04_type_hints.ipynb               # Reference notebook
  05_data_processing.ipynb          # Reference notebook
docs/superpowers/specs/             # Design doc (already exists)
docs/superpowers/plans/             # This plan (already exists)
```

---

### Task 1: Archive existing content to `_reference/`

**Files:**
- Create: `_reference/` (directory)
- Move: all existing content directories and files

- [ ] **Step 1: Create `_reference/` and move content**

```bash
cd /Users/kylebradshaw/repos/gen_ai_engineer
mkdir -p _reference
# Move directories
mv 01_python_refresher _reference/
mv 02_demo_of_gv _reference/
mv 03_nlp_fundamentals _reference/
mv 04_rag_app _reference/
# Move files (keep CLAUDE.md, .gitignore, docs/ at root)
cp CLAUDE.md _reference/CLAUDE.md
mv Inital.md _reference/
mv README.md _reference/
```

- [ ] **Step 2: Verify the archive**

```bash
ls -la _reference/
```

Expected: `01_python_refresher/`, `02_demo_of_gv/`, `03_nlp_fundamentals/`, `04_rag_app/`, `CLAUDE.md`, `Inital.md`, `README.md`

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "archive: move existing content to _reference/"
```

---

### Task 2: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Rewrite CLAUDE.md for the new approach**

Replace the full contents of `CLAUDE.md` with:

```markdown
# CLAUDE.md

## Project Intent

This is a portfolio project for a Gen AI Engineer job application. Kyle is the sole author of all Python code in this repo.

## Environment

- **Python 3.11** via **Miniconda** (not venv)
- Conda environment name: `gen_ai`
- Use `conda install` for packages with C dependencies (numpy, pandas, scipy, pytorch)
- Use `pip install` for pure-Python packages

## Code Authorship Rules

- **Kyle writes all portfolio code himself.** His deliverable is his own retyped notebooks with his own comments.
- **Claude CAN generate reference `.ipynb` notebooks** for the Python refresher section. These are learning materials (like exercise guides), not Kyle's work.
- **Claude CANNOT generate Kyle's implementation code** for other sections (NLP, RAG app, etc.).
- Code snippets in chat are fine — Kyle will type them out and make sure he understands them before moving on.

## What Claude SHOULD Do

- **Generate reference Jupyter notebooks** when Kyle says he needs to learn something (see notebook format below).
- **Generate infrastructure/config files** — .gitignore, Dockerfile, docker-compose.yml, requirements.txt, CLAUDE.md, README files.
- **Review Kyle's code** when asked — point out issues, suggest improvements, explain errors.
- **Answer questions** about concepts, libraries, or approaches.
- **Help debug** when Kyle hits an error.

## Trigger for Learning Guides

Only generate reference notebooks when Kyle **explicitly states** he needs to learn or refresh something. Don't proactively create learning materials for sections he hasn't asked about yet.

## Reference Notebook Format

Each notebook follows a consistent repeating pattern:

### Top cell (markdown)
Title, one-sentence goal, prereqs.

### Per concept (repeating)
1. **Markdown cell — Go/TS comparison:** "In Go you'd do X. In Python, here's what's different and why."
2. **Code cell — example:** Complete, runnable code demonstrating the concept.
3. **Code cell — experiment:** A variation or "try this" prompt.
4. **Markdown cell — "In your own words":** Prompt for Kyle to write his own explanation.

### Bottom cell (markdown)
Quick recap checklist of concepts covered.

## Kyle's Workflow

1. Open the reference notebook (Claude-generated)
2. Create his own notebook — retype code, run it, write own comments
3. Commit his version

## Project Structure

- `01_python_refresher/` — Reference notebooks + Kyle's retyped versions
- `_reference/` — Archived original approach (markdown guides, `.py` files)
- Other sections (NLP, RAG app) — TBD, will be addressed after validating the notebook format
```

- [ ] **Step 2: Verify the file reads correctly**

Read `CLAUDE.md` and confirm it matches the above.

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for notebook-based learning approach"
```

---

### Task 3: Create new README.md

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write the new README**

```markdown
# Gen AI Engineer Portfolio

A portfolio project demonstrating proficiency in Python, AI/ML workflows, and Generative AI — built to showcase skills relevant to a Gen AI Engineer role.

## Current Focus: Python Refresher

Working through Python fundamentals via Jupyter notebooks, with Go/TypeScript comparison framing.

See [`01_python_refresher/README.md`](./01_python_refresher/README.md) for details.

## Planned Sections

| Section | Description | Status |
|---------|-------------|--------|
| `01_python_refresher/` | Core Python via Jupyter notebooks | In progress |
| NLP Fundamentals | Tokenization, embeddings, NER, text classification | Planned |
| RAG App | FastAPI + LangChain + ChromaDB | Planned |

## Prior Work

I previously built an LLM-powered workflow using **Ollama** with **RAG** in a larger Go-based application. That project demonstrates agent-like system design and prompt engineering strategies — retrieving information from a database based on user prompts and structuring prompts for the LLM.

> Repository: *[Link to Go project — coming soon]*

## Getting Started

```bash
cd 01_python_refresher
conda activate gen_ai
pip install -r requirements.txt
jupyter notebook
```
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: rewrite README for restructured repo"
```

---

### Task 4: Create fresh `01_python_refresher/` scaffolding

**Files:**
- Create: `01_python_refresher/README.md`
- Create: `01_python_refresher/requirements.txt`

- [ ] **Step 1: Create the directory and README**

```bash
mkdir -p /Users/kylebradshaw/repos/gen_ai_engineer/01_python_refresher
```

Write `01_python_refresher/README.md`:

```markdown
# Python Refresher — Jupyter Notebooks

Core Python exercises for a developer coming from Go and TypeScript. Each notebook is self-contained — work through it cell-by-cell, retyping code and adding your own explanations.

## Notebooks

| # | File | Topic |
|---|------|-------|
| 0 | `00_environments.ipynb` | .py files, ipython, notebooks, conda |
| 1 | `01_data_structures.ipynb` | Lists, dicts, sets, tuples, generators, comprehensions |
| 2 | `02_oop_patterns.ipynb` | Classes, self, inheritance, ABCs, dunder methods |
| 3 | `03_async_basics.ipynb` | async/await, asyncio, event loop vs goroutines |
| 4 | `04_type_hints.ipynb` | Type hints, Protocol, generics, mypy |
| 5 | `05_data_processing.ipynb` | numpy, pandas basics |

## Workflow

1. Open the reference notebook
2. Create your own copy — retype code cell-by-cell, run it, add your own comments
3. Commit your version
```

- [ ] **Step 2: Write `requirements.txt`**

```
jupyter
numpy
pandas
```

- [ ] **Step 3: Commit**

```bash
git add 01_python_refresher/
git commit -m "scaffold: create 01_python_refresher with README and requirements"
```

---

### Task 5: Generate `00_environments.ipynb`

**Files:**
- Create: `01_python_refresher/00_environments.ipynb`

**Content outline — all cells must be present in the notebook:**

- [ ] **Step 1: Create the notebook with the following cells**

1. *Markdown:* Title — "Python Environments: .py, ipython, and Notebooks". Goal: understand the 3 ways to run Python and when to use each. No prereqs.
2. *Markdown:* Go/TS comparison — In Go you have `go run`, `go build`, and no REPL. In TS you have `ts-node`, compiled JS, and browser console. Python has 3 modes: `.py` scripts, ipython REPL, and Jupyter notebooks. Each serves a different purpose.
3. *Markdown:* Section — "Running .py files"
4. *Code cell:* `print("hello from a .py file")` — explain this is equivalent to `go run main.go` or `ts-node script.ts`.
5. *Markdown:* Go/TS comparison — `.py` files are your production code. Like Go, you run them from the terminal. Unlike Go, no compilation step.
6. *Code cell:* Experiment — `import sys; print(sys.version); print(sys.executable)` — check which Python you're running (matters with conda).
7. *Markdown:* "In your own words" prompt — When would you use a `.py` file vs the other modes?
8. *Markdown:* Section — "ipython REPL"
9. *Code cell:* Show ipython-style exploration — `x = [1, 2, 3]`, `type(x)`, `dir(x)`, `help(x.append)` — explain that `dir()` and `help()` are like GoDoc but live and interactive.
10. *Markdown:* Go/TS comparison — Go has no REPL. TS has `ts-node` but it's clunky. ipython is where you experiment — tab completion, `?` for help, `%timeit` for benchmarks.
11. *Code cell:* Experiment — show `%timeit` magic: `%timeit sum(range(1000))`.
12. *Markdown:* "In your own words" prompt — What's the difference between `dir()` and `help()`?
13. *Markdown:* Section — "Jupyter Notebooks"
14. *Markdown:* Go/TS comparison — No equivalent in Go/TS. Notebooks mix code, output, and documentation in one file. They run on a kernel (a live Python process). Cells share state — a variable defined in cell 1 is available in cell 10.
15. *Code cell:* `name = "Kyle"` — demonstrate shared state.
16. *Code cell:* `print(f"Hello, {name}")` — this works because the kernel remembers `name` from the previous cell.
17. *Code cell:* Experiment — define a list in one cell, modify it in the next, show the gotcha of running cells out of order.
18. *Markdown:* Go/TS comparison — The out-of-order execution gotcha. In Go, code runs top to bottom. In a notebook, you can run cell 5 before cell 3. This is both powerful and dangerous. When confused, "Restart & Run All" is your friend.
19. *Markdown:* "In your own words" prompt — What's the biggest risk of notebook cell ordering? How do you mitigate it?
20. *Markdown:* Section — "Conda Environments"
21. *Markdown:* Go/TS comparison — Like `go mod` or `node_modules` — isolates dependencies per project. Conda also manages the Python version itself (npm only manages packages, not Node).
22. *Code cell:* `import sys; print(sys.prefix)` — shows which conda env is active.
23. *Code cell:* Experiment — `!conda list | head -20` — show installed packages. The `!` prefix runs shell commands from a notebook.
24. *Markdown:* "In your own words" prompt — How is conda different from pip? When would you use each?
25. *Markdown:* Recap checklist — `.py` files for production code, ipython for exploration, notebooks for learning/prototyping, conda for environment isolation, `!` prefix for shell commands in notebooks.

- [ ] **Step 2: Verify notebook opens and runs**

```bash
cd /Users/kylebradshaw/repos/gen_ai_engineer/01_python_refresher
jupyter nbconvert --to notebook --execute 00_environments.ipynb --output /dev/null 2>&1 || echo "Some cells may need interactive features - that's OK"
```

- [ ] **Step 3: Commit**

```bash
git add 01_python_refresher/00_environments.ipynb
git commit -m "lesson: add 00_environments reference notebook"
```

---

### Task 6: Generate `01_data_structures.ipynb`

**Files:**
- Create: `01_python_refresher/01_data_structures.ipynb`

**Content outline — use existing `_reference/01_python_refresher/exercise_01_data_structures.md` as the source material. All cells must be present:**

- [ ] **Step 1: Create the notebook with the following cells**

1. *Markdown:* Title — "Python Data Structures". Goal: understand lists, dicts, sets, tuples, generators, and comprehensions. Prereq: notebook 00.
2. *Markdown:* Section — "Lists"
3. *Markdown:* Go/TS comparison — In Go, slices are contiguous typed values with `len()` and `cap()`. Python lists are arrays of pointers — `[1, "hello", 3.14, None]` is valid. No `cap()`, they just grow on `.append()`. This is why NumPy exists.
4. *Code cell:* `a = [1, 2, 3]; b = a; b.append(4); print(a); print(b)` — demonstrate aliasing. Both print `[1, 2, 3, 4]`.
5. *Code cell:* `a = [1, 2, 3]; b = a.copy(); b.append(4); print(a); print(b)` — demonstrate shallow copy. `a` stays `[1, 2, 3]`.
6. *Markdown:* Go/TS comparison — The aliasing gotcha. In Go, assigning a slice to a new variable creates a new header but shares the backing array (until append triggers a new allocation). In Python, `b = a` means both names point to the SAME list object. Always. Use `.copy()` or `[:]` for a shallow copy, `copy.deepcopy()` for nested structures.
7. *Code cell:* Experiment — `import copy; nested = [[1, 2], [3, 4]]; shallow = nested.copy(); deep = copy.deepcopy(nested); nested[0].append(99); print(shallow); print(deep)` — show shallow vs deep copy.
8. *Markdown:* "In your own words" prompt — When do you need deep copy vs shallow copy?
9. *Markdown:* Section — "Slicing"
10. *Code cell:* `numbers = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]` then show `numbers[2:5]`, `numbers[:3]`, `numbers[7:]`, `numbers[::2]`, `numbers[::-1]`.
11. *Markdown:* Go/TS comparison — Go slicing `a[2:5]` is similar syntax but creates a view of the same backing array. Python slicing creates a NEW list every time. `[::-1]` for reverse has no Go equivalent — you'd write a loop.
12. *Code cell:* Experiment — slice assignment: `numbers[2:5] = [20, 30, 40]; print(numbers)`.
13. *Markdown:* "In your own words" prompt — How does Python slicing differ from Go slicing in terms of memory?
14. *Markdown:* Section — "List Comprehensions"
15. *Markdown:* Go/TS comparison — No Go equivalent (you write loops). TS has `.map()` and `.filter()`. Python comprehensions are the idiomatic way to transform/filter lists.
16. *Code cell:* `squares = [x**2 for x in range(10)]; print(squares)`.
17. *Code cell:* `evens = [x for x in range(20) if x % 2 == 0]; print(evens)`.
18. *Code cell:* Experiment — nested comprehension: `matrix = [[i*j for j in range(1, 4)] for i in range(1, 4)]; print(matrix)`.
19. *Markdown:* "In your own words" prompt — When would you use a comprehension vs a for loop?
20. *Markdown:* Section — "Dictionaries"
21. *Markdown:* Go/TS comparison — Like Go's `map[string]T` or TS objects/Maps. Python dicts are ordered (since 3.7). Keys must be hashable (immutable). No need to check `ok` like Go — use `.get()` for safe access.
22. *Code cell:* Create a dict, access keys, use `.get()` with default, iterate with `.items()`.
23. *Code cell:* Dict comprehension: `{k: v**2 for k, v in {"a": 1, "b": 2, "c": 3}.items()}`.
24. *Markdown:* "In your own words" prompt — How does `.get(key, default)` compare to Go's comma-ok pattern?
25. *Markdown:* Section — "Sets"
26. *Markdown:* Go/TS comparison — Go has no built-in set (you use `map[T]struct{}`). Python sets are first-class with `|`, `&`, `-` operators.
27. *Code cell:* `a = {1, 2, 3}; b = {2, 3, 4}; print(a | b); print(a & b); print(a - b)`.
28. *Code cell:* Experiment — using sets for dedup: `names = ["alice", "bob", "alice", "charlie", "bob"]; unique = list(set(names)); print(unique)`.
29. *Markdown:* "In your own words" prompt — When would you use a set vs a list?
30. *Markdown:* Section — "Tuples"
31. *Markdown:* Go/TS comparison — Like TS tuples `[string, number]`. Immutable. Used for function return values (like Go's multiple returns), dict keys, and fixed structures.
32. *Code cell:* `point = (3, 4); x, y = point; print(x, y)` — tuple unpacking.
33. *Code cell:* Experiment — tuple as dict key: `grid = {}; grid[(0, 0)] = "origin"; grid[(1, 2)] = "point"; print(grid)`.
34. *Markdown:* "In your own words" prompt — Why can tuples be dict keys but lists can't?
35. *Markdown:* Section — "Generators"
36. *Markdown:* Go/TS comparison — Similar to Go channels in concept — produce values one at a time, lazily. Like TS generators (`function*`). Use `yield` instead of `return`. Memory efficient for large sequences.
37. *Code cell:* `def count_up(n): i = 0; ...` (while loop with yield). Use it with `for x in count_up(5)`.
38. *Code cell:* Generator expression: `gen = (x**2 for x in range(1000000)); print(next(gen)); print(next(gen))` — show lazy evaluation.
39. *Code cell:* Experiment — compare memory: `import sys; list_size = sys.getsizeof([x**2 for x in range(10000)]); gen_size = sys.getsizeof(x**2 for x in range(10000)); print(f"List: {list_size}, Generator: {gen_size}")`.
40. *Markdown:* "In your own words" prompt — When would you use a generator vs a list comprehension?
41. *Markdown:* Recap checklist — lists are pointer arrays (aliasing gotcha), `.copy()` vs `deepcopy()`, slicing creates new lists, comprehensions replace loops, dicts are ordered and use `.get()`, sets have `|`/`&`/`-` operators, tuples are immutable (usable as dict keys), generators are lazy and memory efficient.

- [ ] **Step 2: Verify notebook opens**

```bash
jupyter nbconvert --to notebook --execute 01_data_structures.ipynb --output /dev/null 2>&1 || echo "Check for errors"
```

- [ ] **Step 3: Commit**

```bash
git add 01_python_refresher/01_data_structures.ipynb
git commit -m "lesson: add 01_data_structures reference notebook"
```

---

### Task 7: Generate `02_oop_patterns.ipynb`

**Files:**
- Create: `01_python_refresher/02_oop_patterns.ipynb`

**Content outline — use `_reference/01_python_refresher/exercise_02_oop_patterns.md` as source material:**

- [ ] **Step 1: Create the notebook with the following cells**

1. *Markdown:* Title — "Python OOP Patterns". Goal: classes, `self`, inheritance, ABCs, dunder methods. Prereq: notebook 01.
2. *Markdown:* Section — "Classes and `self`"
3. *Markdown:* Go/TS comparison — Go uses structs + methods with receivers (`func (d Dog) Speak()`). TS uses `class` with implicit `this`. Python uses `class` with EXPLICIT `self` — you must pass it as the first param to every method.
4. *Code cell:* Define a `Dog` class with `__init__(self, name, breed)` and a `speak(self)` method. Create an instance, call `speak()`.
5. *Code cell:* Experiment — call `Dog.speak(my_dog)` to show that `self` is just the instance passed explicitly.
6. *Markdown:* "In your own words" prompt — Why does Python require explicit `self`?
7. *Markdown:* Section — "Dunder Methods"
8. *Markdown:* Go/TS comparison — Like implementing `String()` or `Stringer` interface in Go, or `toString()` in TS. Python uses double-underscore methods (`__str__`, `__repr__`, `__len__`, `__eq__`) to hook into built-in behaviors.
9. *Code cell:* Add `__str__`, `__repr__`, `__eq__` to a `Point` class. Show `print(p)`, `repr(p)`, `p1 == p2`.
10. *Code cell:* Add `__len__` and `__getitem__` to a `Deck` class — make it work with `len()` and `[]` indexing.
11. *Code cell:* Experiment — `__add__` for custom `+` behavior on `Point`.
12. *Markdown:* "In your own words" prompt — What's the difference between `__str__` and `__repr__`?
13. *Markdown:* Section — "Inheritance"
14. *Markdown:* Go/TS comparison — Go has no inheritance (composition via embedding). TS has `extends`. Python has full inheritance with `super()`.
15. *Code cell:* `Animal` base class, `Cat` and `Dog` subclasses overriding `speak()`. Show `super().__init__()`.
16. *Code cell:* Experiment — `isinstance()` and `issubclass()` checks.
17. *Markdown:* "In your own words" prompt — When would you use inheritance vs composition in Python?
18. *Markdown:* Section — "Abstract Base Classes (ABCs)"
19. *Markdown:* Go/TS comparison — Like Go interfaces but enforced at class definition time. Go interfaces are satisfied implicitly. Python ABCs must be explicitly inherited and `@abstractmethod` decorated.
20. *Code cell:* `from abc import ABC, abstractmethod`. Define `Shape` ABC with `area()` and `perimeter()`. Implement `Circle` and `Rectangle`.
21. *Code cell:* Experiment — try to instantiate `Shape()` directly — show the error.
22. *Markdown:* "In your own words" prompt — How do Python ABCs compare to Go interfaces?
23. *Markdown:* Section — "Properties and Access Control"
24. *Markdown:* Go/TS comparison — Go has exported/unexported (capital letter). TS has `private`/`public`. Python has no real private — `_name` is a convention, `__name` triggers name mangling but isn't truly private.
25. *Code cell:* Show `@property` decorator for getters/setters. `BankAccount` with `_balance` and a property that validates deposits.
26. *Code cell:* Experiment — access `_balance` directly to show it's not actually enforced.
27. *Markdown:* "In your own words" prompt — How does Python's approach to privacy differ from Go/TS?
28. *Markdown:* Recap checklist — explicit `self`, dunder methods hook into builtins, `super()` for inheritance, ABCs for enforced interfaces, `@property` for getters/setters, `_convention` not enforcement for privacy.

- [ ] **Step 2: Verify and commit**

```bash
jupyter nbconvert --to notebook --execute 02_oop_patterns.ipynb --output /dev/null 2>&1 || echo "Check for errors"
git add 01_python_refresher/02_oop_patterns.ipynb
git commit -m "lesson: add 02_oop_patterns reference notebook"
```

---

### Task 8: Generate `03_async_basics.ipynb`

**Files:**
- Create: `01_python_refresher/03_async_basics.ipynb`

**Content outline — use `_reference/01_python_refresher/exercise_03_async_basics.md` as source material:**

- [ ] **Step 1: Create the notebook with the following cells**

1. *Markdown:* Title — "Python Async Basics". Goal: async/await, asyncio, event loop. Prereq: notebook 02.
2. *Markdown:* Section — "The Mental Model Shift"
3. *Markdown:* Go/TS comparison — Go goroutines are preemptive (the runtime switches between them). Python async is cooperative — your code must explicitly `await` to let other tasks run. Like TS `async/await` but with `asyncio` as the event loop (similar to Node's event loop).
4. *Code cell:* `import asyncio` — basic coroutine: `async def greet(name): await asyncio.sleep(1); return f"Hello, {name}"`. Run with `await greet("Kyle")` (works in notebooks directly).
5. *Markdown:* Go/TS comparison — In Go, `go func()` fires and forgets. In Python, calling `greet("Kyle")` returns a coroutine object — it doesn't run until you `await` it.
6. *Code cell:* Experiment — `coro = greet("Kyle"); print(type(coro)); result = await coro; print(result)`.
7. *Markdown:* "In your own words" prompt — What happens when you call an async function without `await`?
8. *Markdown:* Section — "Concurrent Tasks"
9. *Code cell:* `asyncio.gather()` — run multiple coroutines concurrently. Time the difference between sequential awaits vs gathered.
10. *Code cell:* `asyncio.create_task()` — fire off a task and await it later.
11. *Code cell:* Experiment — `asyncio.as_completed()` — process results as they finish.
12. *Markdown:* Go/TS comparison — `gather()` is like `sync.WaitGroup` in Go or `Promise.all()` in TS. `create_task()` is like `go func()`. `as_completed()` is like `select` on channels.
13. *Markdown:* "In your own words" prompt — When would you use `gather()` vs `create_task()`?
14. *Markdown:* Section — "Error Handling in Async"
15. *Code cell:* try/except in async functions. Show what happens when one task in `gather()` raises.
16. *Code cell:* `gather(return_exceptions=True)` — collect errors without crashing.
17. *Markdown:* "In your own words" prompt — How does error handling in `gather()` compare to error groups in Go?
18. *Markdown:* Section — "Async Patterns"
19. *Code cell:* Async context manager — `async with` for resource cleanup.
20. *Code cell:* Async iteration — `async for` with an async generator.
21. *Code cell:* Experiment — `asyncio.Semaphore` to limit concurrency (like a buffered Go channel).
22. *Markdown:* "In your own words" prompt — What's the Python equivalent of a buffered Go channel for concurrency limiting?
23. *Markdown:* Section — "The Event Loop"
24. *Markdown:* Go/TS comparison — Node has one event loop you never see. Go has a scheduler you never see. Python makes you aware of the event loop — `asyncio.run()` creates one, notebooks already have one running.
25. *Code cell:* `asyncio.get_event_loop()` — show the running loop. Explain why `asyncio.run()` is used in `.py` files but not in notebooks.
26. *Markdown:* "In your own words" prompt — Why can't you call `asyncio.run()` inside a Jupyter notebook?
27. *Markdown:* Recap checklist — async is cooperative (not preemptive), `await` yields control, `gather()` for concurrent execution, `create_task()` for fire-and-await-later, `return_exceptions=True` for resilient gathering, notebooks have a running event loop.

- [ ] **Step 2: Verify and commit**

```bash
jupyter nbconvert --to notebook --execute 03_async_basics.ipynb --output /dev/null 2>&1 || echo "Check for errors"
git add 01_python_refresher/03_async_basics.ipynb
git commit -m "lesson: add 03_async_basics reference notebook"
```

---

### Task 9: Generate `04_type_hints.ipynb`

**Files:**
- Create: `01_python_refresher/04_type_hints.ipynb`

**Content outline — use `_reference/01_python_refresher/exercise_04_type_hints.md` as source material:**

- [ ] **Step 1: Create the notebook with the following cells**

1. *Markdown:* Title — "Python Type Hints". Goal: type hints, Protocol, generics, mypy. Prereq: notebook 02.
2. *Markdown:* Section — "The Big Shock: Hints Are Ignored at Runtime"
3. *Markdown:* Go/TS comparison — In Go, types are enforced at compile time. In TS, types are enforced at compile time but erased at runtime. In Python, type hints are COMPLETELY ignored at runtime — they're just documentation that tools like mypy can check. `x: int = "hello"` runs fine.
4. *Code cell:* `x: int = "hello"; print(x, type(x))` — demonstrate runtime doesn't care.
5. *Code cell:* Experiment — `def add(a: int, b: int) -> int: return a + b; print(add("foo", "bar"))` — no error.
6. *Markdown:* "In your own words" prompt — How does Python's approach to types compare to Go and TypeScript?
7. *Markdown:* Section — "Basic Type Hints"
8. *Code cell:* Function signatures: `def greet(name: str) -> str`, variables: `count: int = 0`, collections: `items: list[str] = []`.
9. *Code cell:* `Optional` and `Union`: `from typing import Optional; def find(items: list[str], target: str) -> Optional[int]`.
10. *Code cell:* Experiment — `x: int | str = "hello"` (Python 3.10+ union syntax).
11. *Markdown:* "In your own words" prompt — What's the difference between `Optional[int]` and `int | None`?
12. *Markdown:* Section — "Complex Types"
13. *Code cell:* `dict[str, list[int]]`, `tuple[str, int, float]` (fixed), `tuple[int, ...]` (variable).
14. *Code cell:* `Callable[[int, int], int]` for function types.
15. *Code cell:* `TypedDict` — like TS interfaces for dict shapes.
16. *Markdown:* "In your own words" prompt — How does `TypedDict` compare to Go structs or TS interfaces?
17. *Markdown:* Section — "Protocol (Structural Typing)"
18. *Markdown:* Go/TS comparison — THIS is the one that feels like Go interfaces. `Protocol` defines structural types — if an object has the right methods, it satisfies the protocol. No explicit inheritance needed.
19. *Code cell:* Define `Drawable` protocol with `draw(self) -> str`. Create `Circle` and `Square` that satisfy it WITHOUT inheriting. Type-check a function that accepts `Drawable`.
20. *Code cell:* Experiment — create a class that DOESN'T satisfy the protocol, show mypy would catch it.
21. *Markdown:* "In your own words" prompt — How is `Protocol` different from ABCs?
22. *Markdown:* Section — "Generics"
23. *Markdown:* Go/TS comparison — Like Go generics (`func Foo[T any](x T)`) or TS generics (`function foo<T>(x: T)`).
24. *Code cell:* `from typing import TypeVar; T = TypeVar('T'); def first(items: list[T]) -> T: return items[0]`.
25. *Code cell:* Python 3.12 syntax: `def first[T](items: list[T]) -> T: return items[0]`.
26. *Code cell:* Bounded generics: `T = TypeVar('T', bound=Comparable)`.
27. *Markdown:* "In your own words" prompt — How do Python generics compare to Go's type constraints?
28. *Markdown:* Section — "Using mypy"
29. *Code cell:* `!pip install mypy` (if not installed). Write a `.py` file with type errors, run `!mypy filename.py`, show the output.
30. *Markdown:* "In your own words" prompt — When would you use mypy in a real project?
31. *Markdown:* Recap checklist — hints ignored at runtime, `Optional`/`Union`/`|` for nullable, `Protocol` for structural typing (like Go interfaces), `TypeVar` for generics, mypy for static checking.

- [ ] **Step 2: Verify and commit**

```bash
jupyter nbconvert --to notebook --execute 04_type_hints.ipynb --output /dev/null 2>&1 || echo "Check for errors"
git add 01_python_refresher/04_type_hints.ipynb
git commit -m "lesson: add 04_type_hints reference notebook"
```

---

### Task 10: Generate `05_data_processing.ipynb`

**Files:**
- Create: `01_python_refresher/05_data_processing.ipynb`

**Content outline — use `_reference/01_python_refresher/exercise_05_data_processing.md` as source material:**

- [ ] **Step 1: Create the notebook with the following cells**

1. *Markdown:* Title — "Data Processing with NumPy and Pandas". Goal: numpy arrays, pandas DataFrames, basic data manipulation. Prereq: notebook 01.
2. *Markdown:* Section — "Why NumPy Exists"
3. *Markdown:* Go/TS comparison — No direct equivalent in Go or TS. Python lists are slow for numeric work (pointers, not values). NumPy stores actual numbers contiguously in memory — like Go's `[]float64` but with built-in math operations.
4. *Code cell:* `import numpy as np; a = np.array([1, 2, 3, 4, 5]); print(a, a.dtype, a.shape)`.
5. *Code cell:* Vectorized operations: `a * 2`, `a + a`, `np.sqrt(a)` — no loops needed.
6. *Code cell:* Experiment — time comparison: `%timeit sum(range(1000000))` vs `%timeit np.sum(np.arange(1000000))`.
7. *Markdown:* "In your own words" prompt — Why is NumPy faster than Python lists for numeric operations?
8. *Markdown:* Section — "Array Creation and Shaping"
9. *Code cell:* `np.zeros()`, `np.ones()`, `np.arange()`, `np.linspace()`.
10. *Code cell:* `reshape()`, `.T` for transpose, `np.concatenate()`.
11. *Code cell:* Experiment — `np.random.default_rng(42).normal(0, 1, (3, 4))` — create a random matrix.
12. *Markdown:* "In your own words" prompt — What does `reshape()` do and when would you use it?
13. *Markdown:* Section — "Indexing and Boolean Masks"
14. *Code cell:* Array slicing (same syntax as lists but returns VIEWS, not copies).
15. *Code cell:* Boolean masking: `a = np.array([1, 2, 3, 4, 5]); mask = a > 3; print(a[mask])`.
16. *Code cell:* Experiment — fancy indexing: `a[[0, 2, 4]]`.
17. *Markdown:* Go/TS comparison — NumPy slicing returns VIEWS (like Go slices sharing a backing array). Modifying the slice modifies the original. Use `.copy()` to avoid this.
18. *Markdown:* "In your own words" prompt — How does NumPy slicing differ from Python list slicing?
19. *Markdown:* Section — "Pandas Basics"
20. *Markdown:* Go/TS comparison — No Go/TS equivalent. Think of a DataFrame as a spreadsheet or SQL table in memory. Series is a single column. This is the primary tool for data manipulation in Python.
21. *Code cell:* `import pandas as pd; df = pd.DataFrame({"name": ["Alice", "Bob", "Charlie"], "age": [25, 30, 35], "score": [88.5, 92.0, 78.5]}); print(df)`.
22. *Code cell:* Column access: `df["name"]`, `df.age`, `df[["name", "score"]]`.
23. *Code cell:* Filtering: `df[df["age"] > 25]`, `df.query("age > 25 and score > 80")`.
24. *Markdown:* "In your own words" prompt — How is filtering a DataFrame similar to SQL WHERE clauses?
25. *Markdown:* Section — "Data Manipulation"
26. *Code cell:* `df.sort_values("score", ascending=False)`.
27. *Code cell:* `df.groupby("category").mean()` — aggregate by group.
28. *Code cell:* Adding columns: `df["grade"] = df["score"].apply(lambda x: "A" if x >= 90 else "B")`.
29. *Code cell:* Experiment — `df.describe()` for summary statistics.
30. *Markdown:* "In your own words" prompt — How does `groupby()` compare to SQL GROUP BY?
31. *Markdown:* Section — "Reading and Writing Data"
32. *Code cell:* `df.to_csv("example.csv", index=False); df2 = pd.read_csv("example.csv"); print(df2)`.
33. *Code cell:* Experiment — `df.to_json()`, `pd.read_json()`.
34. *Markdown:* "In your own words" prompt — What does `index=False` do when saving a CSV?
35. *Markdown:* Recap checklist — NumPy for fast numeric arrays (contiguous memory), vectorized operations replace loops, slicing returns views (not copies), pandas DataFrame = spreadsheet/SQL table, filtering with boolean masks or `.query()`, `groupby()` for aggregation, `read_csv()`/`to_csv()` for I/O.

- [ ] **Step 2: Verify and commit**

```bash
jupyter nbconvert --to notebook --execute 05_data_processing.ipynb --output /dev/null 2>&1 || echo "Check for errors"
git add 01_python_refresher/05_data_processing.ipynb
git commit -m "lesson: add 05_data_processing reference notebook"
```

---

## Summary

| Task | What | Commit |
|------|------|--------|
| 1 | Archive existing content to `_reference/` | `archive: move existing content to _reference/` |
| 2 | Update CLAUDE.md | `docs: update CLAUDE.md for notebook-based learning approach` |
| 3 | Create new README.md | `docs: rewrite README for restructured repo` |
| 4 | Scaffold `01_python_refresher/` | `scaffold: create 01_python_refresher with README and requirements` |
| 5 | Generate `00_environments.ipynb` | `lesson: add 00_environments reference notebook` |
| 6 | Generate `01_data_structures.ipynb` | `lesson: add 01_data_structures reference notebook` |
| 7 | Generate `02_oop_patterns.ipynb` | `lesson: add 02_oop_patterns reference notebook` |
| 8 | Generate `03_async_basics.ipynb` | `lesson: add 03_async_basics reference notebook` |
| 9 | Generate `04_type_hints.ipynb` | `lesson: add 04_type_hints reference notebook` |
| 10 | Generate `05_data_processing.ipynb` | `lesson: add 05_data_processing reference notebook` |
