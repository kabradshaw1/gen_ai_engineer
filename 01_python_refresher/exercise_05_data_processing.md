# Exercise 5: Data Processing — `data_processing.py`

**Goal:** Learn pandas and numpy — this has no Go or TypeScript equivalent.

**Environment:** This exercise works great in ipython for exploration. Pandas DataFrames render nicely in ipython with auto-formatting. You could also use a Jupyter notebook for this one since seeing tabular output inline is helpful.

---

## Why This Is Different

The other exercises are about translating concepts you already know. This one is genuinely new. Go and TS don't have anything like pandas or numpy. The closest equivalents would be:

- **numpy** → Writing C code to do math on arrays, or using a math library. Go's standard library has nothing comparable. TS has typed arrays for WebGL but no general-purpose numeric computing.
- **pandas** → SQL. Seriously. If you think of a DataFrame as a SQL table and pandas operations as queries, the API clicks fast. In Go, you'd query a database or loop over slices of structs. In TS, you'd use lodash or SQL.

The job listing says "Data Processing" — in Python, that means pandas.

---

## Part A: NumPy — Typed Arrays with Vectorized Operations

### Concept

A Python list is an array of pointers. A numpy array is a contiguous block of typed values — much closer to a Go slice in memory layout, but with built-in math operations.

The core idea: **operate on the whole array at once, no loops.** Numpy runs C code under the hood, so vectorized operations on 1M elements are 10-100x faster than a Python loop.

### Action Steps

**Step 1: Arrays vs lists (ipython)**

```python
import numpy as np

# Python list — array of pointers, mixed types allowed
py_list = [1, 2, 3, 4, 5]
type(py_list[0])     # int — each element is a Python object

# NumPy array — contiguous block of typed values (like a Go slice)
np_array = np.array([1, 2, 3, 4, 5])
np_array.dtype       # int64 — all elements are the same C type
np_array.shape       # (5,) — shape is a concept lists don't have
```

In Go, `[]int{1, 2, 3, 4, 5}` stores 5 contiguous ints. A numpy array does the same thing. A Python list stores 5 pointers to int objects scattered on the heap.

**Step 2: Vectorized operations — no loops (ipython)**

```python
a = np.array([1, 2, 3, 4, 5])
b = np.array([10, 20, 30, 40, 50])

# In Go, each of these would be a for loop:
a + b          # [11, 22, 33, 44, 55]
a * 2          # [2, 4, 6, 8, 10]
a ** 2         # [1, 4, 9, 16, 25]
a > 3          # [False, False, False, True, True] — boolean array!
```

**Step 3: Boolean indexing — filtering without loops (ipython)**

```python
a = np.array([1, 2, 3, 4, 5, 6, 7, 8, 9, 10])

# Go: manual loop with if and append
# TS: a.filter(x => x > 5)
# NumPy:
mask = a > 5              # [False, False, ..., True, True, True, True, True]
a[mask]                   # [6, 7, 8, 9, 10]

# Or in one line:
a[a > 5]                  # [6, 7, 8, 9, 10]

# Combine conditions (& for AND, | for OR — note the parentheses)
a[(a > 3) & (a < 8)]     # [4, 5, 6, 7]
```

This is like a `WHERE` clause in SQL: `SELECT * FROM a WHERE a > 5`.

**Step 4: Reshaping and aggregation (ipython)**

```python
# Create a 3x4 matrix
m = np.arange(12).reshape(3, 4)
m
# array([[ 0,  1,  2,  3],
#        [ 4,  5,  6,  7],
#        [ 8,  9, 10, 11]])

# Slicing — 2D indexing (Go has no equivalent for multidimensional slicing)
m[0, :]       # first row: [0, 1, 2, 3]
m[:, 1]       # second column: [1, 5, 9]
m[1:, 2:]     # rows 1+, columns 2+: [[6, 7], [10, 11]]

# Aggregation
np.mean(m)           # mean of all elements
np.mean(m, axis=0)   # mean of each column
np.mean(m, axis=1)   # mean of each row
np.std(m)            # standard deviation
```

**Step 5: Performance comparison (ipython)**

```python
import time

size = 1_000_000
py_list = list(range(size))
np_array = np.arange(size)

# Python loop
start = time.perf_counter()
result_py = [x * 2 for x in py_list]
py_time = time.perf_counter() - start

# NumPy vectorized
start = time.perf_counter()
result_np = np_array * 2
np_time = time.perf_counter() - start

print(f"Python loop: {py_time:.4f}s")
print(f"NumPy:       {np_time:.4f}s")
print(f"Speedup:     {py_time / np_time:.0f}x")
```

You should see 10-100x speedup. This is why data science uses numpy — Python loops are slow because of all the pointer chasing and object overhead. Numpy bypasses this by running C loops on contiguous memory.

**Write a comment for your .py file:** Explain why numpy exists — Python lists are slow for math because they store pointers. Numpy stores typed values contiguously (like Go slices) and runs C code. Vectorized = no Python loop.

---

## Part B: Pandas — SQL Tables in Python

### Concept

Think of a pandas DataFrame as a SQL table:
- Columns = fields with types
- Rows = records
- Operations = queries

If you've written SQL, pandas will feel familiar. If you've only used Go structs, think of a DataFrame as `[]struct{}` with built-in query methods.

### Action Steps

**Step 6: Create a DataFrame (ipython)**

```python
import pandas as pd

df = pd.DataFrame({
    "name": ["Alice", "Bob", "Charlie", "Diana", "Eve"],
    "department": ["ML", "Backend", "ML", "Frontend", "Backend"],
    "salary": [95000, 87000, 92000, 78000, 91000],
    "years": [5, 3, 4, 2, 6],
})

df            # ipython renders this as a nice table
df.dtypes     # column types — like checking struct field types
df.shape      # (5, 4) — 5 rows, 4 columns
df.columns    # column names
```

In Go, this data would be a `[]Employee` where `Employee` is a struct. Every operation below would be a `for` loop. In TS, it'd be an array of objects.

**Step 7: Filtering = WHERE (ipython)**

```python
# SQL: SELECT * FROM df WHERE salary > 90000
df[df["salary"] > 90000]

# SQL: SELECT * FROM df WHERE department = 'ML'
df[df["department"] == "ML"]

# SQL: SELECT * FROM df WHERE department = 'ML' AND salary > 90000
df[(df["department"] == "ML") & (df["salary"] > 90000)]
```

The syntax `df[condition]` is boolean indexing — same concept as numpy. The condition creates a True/False series, and pandas keeps only the True rows.

**Step 8: Groupby = GROUP BY (ipython)**

This is the most important pandas operation to learn.

```python
# SQL: SELECT department, AVG(salary) FROM df GROUP BY department
df.groupby("department")["salary"].mean()

# SQL: SELECT department, COUNT(*), AVG(salary), MAX(salary)
#      FROM df GROUP BY department
df.groupby("department")["salary"].agg(["count", "mean", "max"])

# SQL: SELECT department, AVG(salary), AVG(years)
#      FROM df GROUP BY department
df.groupby("department")[["salary", "years"]].mean()
```

In Go, you'd build a `map[string][]Employee`, loop through, accumulate, then compute averages manually. Pandas does it in one line.

**Step 9: Sorting = ORDER BY (ipython)**

```python
# SQL: SELECT * FROM df ORDER BY salary DESC
df.sort_values("salary", ascending=False)

# SQL: SELECT * FROM df ORDER BY department, salary DESC
df.sort_values(["department", "salary"], ascending=[True, False])
```

**Step 10: Adding columns = computed fields (ipython)**

```python
# Add a bonus column (10% of salary)
df["bonus"] = df["salary"] * 0.1

# Add a seniority label
df["senior"] = df["years"] >= 4

df
```

This is vectorized — no loop. Every row gets the computation at once.

**Step 11: Data cleaning (ipython)**

Create a messy DataFrame:
```python
messy = pd.DataFrame({
    "name": ["Alice", "Bob", None, "Diana", "Bob"],
    "score": [85, None, 72, 90, None],
    "grade": ["A", "B", "C", "A", "B"],
})

messy
```

Clean it:
```python
# Find nulls
messy.isnull()                    # True/False for every cell
messy.isnull().sum()              # count nulls per column

# Drop rows with any null
messy.dropna()                    # 2 rows left

# Fill nulls with a value
messy.fillna({"name": "Unknown", "score": 0})

# Drop duplicate rows (Bob appears twice with different scores — or same?)
messy.drop_duplicates(subset=["name"])   # keeps first occurrence

# Check and cast types
messy.dtypes
messy["score"].astype(float)      # explicit type casting
```

**Write a comment for your .py file:** Frame pandas operations as SQL equivalents: filter = WHERE, groupby = GROUP BY, sort_values = ORDER BY, merge = JOIN. Note that everything is vectorized — no Python loops.

**Step 12: Method chaining (ipython)**

Pandas supports chaining operations — similar to piping in Go or method chaining in TS:

```python
# SQL: SELECT department, AVG(salary) as avg_salary
#      FROM df WHERE years >= 3
#      GROUP BY department
#      ORDER BY avg_salary DESC

result = (
    df[df["years"] >= 3]
    .groupby("department")["salary"]
    .mean()
    .sort_values(ascending=False)
)
result
```

Each line transforms the data. Read it top to bottom like a pipeline. This is the idiomatic way to write pandas — not intermediate variables for each step.

**Write a comment:** Explain method chaining as building a data pipeline. Each step transforms the output of the previous step. Parentheses `()` allow multi-line chaining.

---

## Writing Your .py File

```python
"""Data Processing: pandas and numpy

[Your summary: why these libraries exist (Python lists are slow for math,
Go/TS have no equivalent for tabular data). Think of numpy as typed arrays
with vectorized ops, pandas as SQL tables in Python.]
"""

import numpy as np
import pandas as pd


def numpy_demo():
    """NumPy — typed arrays with vectorized operations."""
    # [Your comment: numpy arrays vs Python lists — memory layout,
    #  typed values, no pointer chasing]

    # Vectorized operations
    # [Your comment: every operation here would be a for loop in Go]

    # Boolean indexing
    # [Your comment: like a WHERE clause — filter without loops]

    # Reshaping and aggregation
    # [Your comment: 2D arrays and axis-based operations]


def pandas_demo():
    """Pandas — SQL tables in Python."""
    # [Your comment: DataFrame = SQL table. Columns = fields, rows = records.]

    # Create a DataFrame

    # Filtering (= WHERE)
    # [Your comment: boolean indexing on DataFrames]

    # Groupby (= GROUP BY)
    # [Your comment: most important pandas operation. Compare to Go's
    #  manual map[string][]T + loop approach]

    # Sorting (= ORDER BY)

    # Adding columns (= computed fields)

    # Method chaining
    # [Your comment: read top to bottom like a pipeline]


def cleaning_demo():
    """Data cleaning — handling nulls and duplicates."""
    # [Your comment: isnull, dropna, fillna, drop_duplicates]


if __name__ == "__main__":
    numpy_demo()
    pandas_demo()
    cleaning_demo()
```

---

## Action Checklist

- [ ] Work through Steps 1-12 in ipython (or a notebook — DataFrames render nicely)
- [ ] Run the performance comparison (Step 5) — see the numpy speedup yourself
- [ ] For every pandas operation, think "what SQL query is this?"
- [ ] Try method chaining — write a multi-step pipeline in one expression
- [ ] Create `data_processing.py` with your own comments
- [ ] Run `python data_processing.py` end to end
- [ ] Review your comments — would they help a Go developer understand pandas?

---

## Congratulations

When you finish this exercise, you've completed the Python refresher. You should now be able to:

- Recognize how Python's data structures differ from Go/TS (reference semantics, comprehensions, generators)
- Write classes and understand Python's OOP model vs Go structs and TS classes
- Write async code and know why it's fundamentally different from goroutines
- Use type hints and know their limitations (compared to compiled languages)
- Use numpy and pandas for data processing

Go back and ask me to review any of your .py files — I'll check your comments and code for accuracy.

When you're ready, let me know and we'll move on to the NLP fundamentals section.
