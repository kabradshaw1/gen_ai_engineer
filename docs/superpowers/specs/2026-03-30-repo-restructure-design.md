# Repo Restructure: Jupyter Notebook-Based Learning

**Date:** 2026-03-30
**Status:** Draft

## Problem

The current workflow (read markdown guide → explore in ipython → write `.py` file with comments → run) has too much mental overhead. Four artifacts per topic creates friction that slows learning.

## Solution

Replace the markdown-guide + `.py` file approach with self-contained Jupyter notebooks. One notebook per topic. Kyle retypes them cell-by-cell to learn the material.

## Repo Structure

### Archive existing work

Move all current content into `_reference/`, preserving directory structure:

```
_reference/
  01_python_refresher/    (all existing .md guides, .py files, etc.)
  02_demo_of_gv/
  03_nlp_fundamentals/
  04_rag_app/
  Inital.md
  README.md
  CLAUDE.md               (old CLAUDE.md for reference)
```

### Keep at repo root (updated)

- `CLAUDE.md` — updated for the new approach
- `README.md` — rewritten for the new structure
- `.gitignore`

### Fresh Python refresher

```
01_python_refresher/
  README.md
  00_environments.ipynb
  01_data_structures.ipynb
  02_oop_patterns.ipynb
  03_async_basics.ipynb
  04_type_hints.ipynb
  05_data_processing.ipynb
  requirements.txt
```

### Other sections

`02_demo_of_gv/`, `03_nlp_fundamentals/`, `04_rag_app/` — not part of this restructure. Will be addressed later once the notebook format is validated.

## Notebook Format

Each notebook follows a consistent repeating pattern:

### Top cell (markdown)
Title, one-sentence goal, prereqs (e.g., "finish notebook 00 first").

### Per concept (repeating)
1. **Markdown cell — Go/TS comparison:** "In Go you'd do X. In Python, here's what's different and why."
2. **Code cell — example:** Complete, runnable code demonstrating the concept. Kyle retypes this.
3. **Code cell — experiment:** A small variation or "try this" prompt to poke at the concept.
4. **Markdown cell — "In your own words":** Prompt for Kyle to write a comment explaining what he learned.

### Bottom cell (markdown)
Quick recap checklist of concepts covered.

### What notebooks do NOT include
- No `.py` file skeletons — the notebook is the artifact
- No graded exercises or test assertions — this is a refresher, not a course

## Kyle's Workflow (new)

1. Open the reference notebook (Claude-generated)
2. Create his own notebook side-by-side (or retype into a copy)
3. Work through it cell-by-cell — retype code, run it, write own comments
4. Commit his version

## Topic Coverage

Same 6 topics, same order — well-scoped and builds on each other:

| # | Notebook | Coverage |
|---|----------|----------|
| 00 | environments | .py files, ipython, notebooks, conda |
| 01 | data_structures | lists, dicts, sets, tuples, generators, comprehensions |
| 02 | oop_patterns | classes, self, inheritance, ABCs, dunder methods |
| 03 | async_basics | async/await, asyncio, event loop vs goroutines |
| 04 | type_hints | type hints, Protocol, generics, mypy |
| 05 | data_processing | numpy, pandas basics |

## CLAUDE.md Changes

**Old rule:** "NEVER generate Python implementation files (.py, .ipynb)"

**New rule:** Claude CAN generate reference `.ipynb` notebooks for the Python refresher section. These are learning materials (like the existing markdown guides), not Kyle's work. Kyle's deliverable is his own retyped copy with his own comments.

The old `CLAUDE.md` is preserved in `_reference/CLAUDE.md`.

All other CLAUDE.md rules remain unchanged — Claude still doesn't write Kyle's implementation code for other sections.
