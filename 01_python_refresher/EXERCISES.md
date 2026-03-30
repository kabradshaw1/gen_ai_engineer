# Python Refresher — For Go & TypeScript Developers

You know how to program. These exercises focus on where Python *differs* from Go and TypeScript — the gotchas, the mental model shifts, and the idioms that don't translate directly.

## Workflow

1. **Read the exercise guide** — understand the concept and how it differs from Go/TS
2. **Explore in ipython** — try the code examples, experiment, break things
3. **Write your .py file** — type the code yourself, add comments in your own words
4. **Run it** — `python your_file.py`

Your comments are the deliverable. They prove you understand the material, not just that you can copy code.

## Exercises

| # | File | Topic | Key Difference from Go/TS |
|---|------|-------|--------------------------|
| 0 | [exercise_00_environments.md](exercise_00_environments.md) | .py files, ipython, notebooks | Python has 3 ways to run code — learn when to use each |
| 1 | [exercise_01_data_structures.md](exercise_01_data_structures.md) | Lists, dicts, sets, generators | Lists are pointer arrays, assignment aliases, comprehensions replace loops |
| 2 | [exercise_02_oop_patterns.md](exercise_02_oop_patterns.md) | Classes, inheritance, ABCs | Explicit `self`, no real private, nominal interfaces |
| 3 | [exercise_03_async_basics.md](exercise_03_async_basics.md) | async/await, asyncio | Cooperative single-thread vs Go's preemptive goroutines |
| 4 | [exercise_04_type_hints.md](exercise_04_type_hints.md) | Type hints, Protocol, generics | Hints are ignored at runtime — the biggest shock |
| 5 | [exercise_05_data_processing.md](exercise_05_data_processing.md) | numpy, pandas | No Go/TS equivalent — genuinely new territory |

Start with Exercise 0 to set up your environments, then work through 1-5 in order.
