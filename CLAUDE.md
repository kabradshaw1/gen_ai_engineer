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

- **Generate reference Jupyter notebooks** when Kyle says he needs to learn or refresh something (see notebook format below).
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
- `02_nlp_fundamentals/` — NLP reference notebooks + Kyle's retyped versions
- `_reference/` — Archived original approach (markdown guides, `.py` files)
- Other sections (RAG app) — TBD, will be addressed after NLP
