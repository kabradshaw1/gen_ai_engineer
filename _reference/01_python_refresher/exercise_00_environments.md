# Exercise 0: Python Environments — .py Files, ipython, and Notebooks

Before you write any exercises, get comfortable with the three ways you'll run Python. Each has a different purpose. You'll try all three on the same example so you can feel the difference.

---

## Setup

### Environment with Miniconda

Create a conda environment for this project:

```bash
conda create -n gen_ai python=3.11 -y
conda activate gen_ai
```

**Conda vs venv — what's the difference?**

You'll see `python -m venv` in most tutorials. Both solve the same problem (isolated dependencies), but conda is more powerful:

- **venv** — Python-only. Creates a folder with a Python binary and `site-packages`. Only manages Python packages via `pip`.
- **conda** — Manages Python itself AND system-level C/Fortran libraries. Useful for data science where numpy, scipy, and pytorch depend on compiled code. Conda installs pre-built binaries so you don't need a C compiler.

In Go terms: `venv` is like `GOPATH` per project. Conda is more like having separate Go toolchain installs with different system libraries.

**`conda install` vs `pip install`:**
- Use `conda install` for packages with C dependencies (numpy, pandas, scipy, pytorch) — conda handles the compiled bits better
- Use `pip install` for everything else. Pip works fine inside a conda environment.
- Don't mix both for the same package — pick one.

```bash
conda activate gen_ai
conda install ipython jupyter pandas numpy -y
```

**Verify your setup:**
```bash
which python        # Should point to your conda env, not system Python
python --version    # Should be 3.11.x
conda list          # Shows installed packages
```

**Useful conda commands:**
```bash
conda env list              # List all environments
conda activate gen_ai       # Activate
conda deactivate            # Back to base
conda install <package> -y  # Install a package
```

---

## The Three Environments

### 1. ipython — The Exploration Shell

**What it is:** An enhanced Python REPL. Like running `go run` on a snippet or using the Node REPL, but much more powerful.

**When to use it:** When you're experimenting. Trying a new function, checking how something works, testing a quick idea. You wouldn't write a whole Go program to see what `strings.Split` does — you'd look it up. In Python, you just try it in ipython.

**Start it:**
```bash
ipython
```

**Try this — type each line and hit Enter:**
```python
a = [1, 2, 3, 4, 5]
a[::2]
```

Notice: ipython automatically prints the result of the last expression. No `fmt.Println` or `console.log` needed. This is why it's great for exploration.

**Useful ipython features you don't get in Go/TS:**
```python
a = [1, 2, 3]
a.          # Hit Tab — see all methods on a list
a.append?   # ? shows the docstring
a.append??  # ?? shows the source code
%timeit sum(range(1000))   # Built-in benchmarking — like Go's testing.B
```

**When NOT to use it:** When you're writing something you want to keep. ipython is scratch paper. Nothing is saved when you close it.

**Action step:** Open ipython now. Create a list, a dict, and a set. Use Tab completion to explore what methods are available on each. Type `exit` when done.

### 2. .py Files — The Permanent Record

**What it is:** A Python script. Like a `.go` file or `.ts` file. This is where your portfolio code lives.

**When to use it:** When you're writing something you want to commit. Your exercises, your demos, your portfolio pieces. This is the finished product.

**Create `hello.py`** (you type this — don't copy/paste):
```python
"""
My first Python script.

In Go, I'd have package main + func main().
In TS, I'd just write code at the top level.
In Python, I use the if __name__ block to make it runnable AND importable.
"""


def greet(name: str) -> str:
    # In Go: func greet(name string) string
    # In TS: function greet(name: string): string
    return f"Hello, {name}!"


if __name__ == "__main__":
    # This block runs when you execute: python hello.py
    # It does NOT run when another file does: from hello import greet
    # Go equivalent: func main()
    # TS equivalent: no direct equivalent — TS files are either scripts or modules
    print(greet("Kyle"))
```

**Run it:**
```bash
python hello.py
```

**Why `if __name__ == "__main__":`?**
In Go, `main()` is special — the compiler knows it's the entry point. In Python, every `.py` file can be both a script AND a module. This guard lets you write code that works both ways. When you run `python hello.py`, Python sets `__name__` to `"__main__"`. When another file imports it, `__name__` is set to `"hello"`.

**Action step:** Create `hello.py`, type it out yourself. Run it with `python hello.py`. Then open ipython and type `from hello import greet` — notice the `if __name__` block doesn't run. Then type `greet("World")` — the function is available.

### 3. Jupyter Notebooks — The Explainer

**What it is:** A document that mixes code, output, and prose. Like a Google Doc where some paragraphs are runnable code. There's no Go or TS equivalent — this is unique to the Python/data science ecosystem.

**When to use it:** When you're telling a story with code. Your NLP unified notebook. Anything where you want to show your thought process alongside the results. Also great for data exploration because outputs (tables, charts) render inline.

**Start it:**
```bash
jupyter notebook
```

This opens a browser. Click "New → Python 3" to create a notebook.

**Try this:**

Cell 1 (Markdown cell — change cell type in the dropdown):
```markdown
# List Reference Semantics
In Go, assigning a slice copies the header. In Python, it's a straight alias.
```

Cell 2 (Code cell):
```python
a = [1, 2, 3]
b = a
b.append(4)
print(f"a = {a}")  # Shows [1, 2, 3, 4]
print(f"b = {b}")  # Shows [1, 2, 3, 4]
print(f"Same object? {a is b}")  # True
```

Run each cell with Shift+Enter. Notice: the output appears right below the code. This is what makes notebooks powerful for explanation.

**Key notebook concepts:**
- Cells run independently, but share state. If Cell 1 defines `x = 5`, Cell 2 can use `x`.
- You can run cells out of order. This is a feature AND a footgun — it means the notebook state might not match top-to-bottom reading.
- Restart the kernel (Kernel → Restart & Run All) to verify your notebook works from scratch.

**When NOT to use it:** For code that will be imported by other code. Notebooks are documents, not modules. Your `.py` files are the building blocks; notebooks are the presentations.

**Action step:** Create a notebook. Make a Markdown cell with a title. Make a code cell that creates a list and demonstrates the aliasing behavior. Run both cells. Save it, then delete it — this was just practice.

---

## When to Use What — Summary

| Situation | Use | Why |
|-----------|-----|-----|
| "What does this function do?" | ipython | Instant feedback, tab completion, ? for docs |
| "I'm writing exercise code for my portfolio" | .py file | Permanent, runnable, importable, committable |
| "I want to explain a concept with code + prose" | Notebook | Mix markdown and code, inline output, tells a story |
| "I'm exploring a new library" | ipython first, then .py | Explore in ipython, solidify in a script |
| "I'm debugging something weird" | ipython | Isolate the problem in a clean environment |

---

## Your Workflow for Each Exercise

For exercises 1-5, here's the workflow:

1. **Read the exercise guide** — understand what you're building and why it differs from Go/TS
2. **Explore in ipython** — try the code snippets from the guide. Experiment. Break things.
3. **Write your .py file** — type out the code (don't copy-paste). Add comments in your own words explaining each concept. These comments are your proof that you understand it.
4. **Run your .py file** — `python exercise_name.py`. Make sure it works top to bottom.
5. **Commit** — this is portfolio code.

For the unified notebooks (NLP section), the notebook IS the deliverable — it's where explanation and code live together.

---

## Action Checklist

- [ ] Create and activate conda environment (`conda create -n gen_ai python=3.11 -y`)
- [ ] Install ipython and jupyter (`conda install ipython jupyter pandas numpy -y`)
- [ ] Open ipython, explore a list with Tab completion, then exit
- [ ] Create `hello.py`, run it, then import from it in ipython
- [ ] Create a throwaway Jupyter notebook, make a markdown cell and a code cell, run both
- [ ] Delete the throwaway notebook and `hello.py` — they were just practice

Once this feels comfortable, move on to Exercise 1.
