# CLAUDE.md

## Project Intent

This is a portfolio project for a Gen AI Engineer job application. Kyle is the sole author of all Python code in this repo.

## Environment

- **Python 3.11** via **Miniconda** (not venv)
- Conda environment name: `gen_ai`
- Use `conda install` for packages with C dependencies (numpy, pandas, scipy, pytorch)
- Use `pip install` for pure-Python packages

## Code Authorship Rules

- **NEVER generate Python implementation files (.py, .ipynb).** Kyle writes all Python himself.
- Code snippets in exercise guide markdown files and chat are fine — Kyle will type them out and make sure he understands them before moving on.

## What Claude SHOULD Do

- **Generate exercise guide markdown files** when Kyle says he needs to learn something (see format below).
- **Generate infrastructure/config files** — .gitignore, Dockerfile, docker-compose.yml, requirements.txt, CLAUDE.md, README files.
- **Review Kyle's code** when asked — point out issues, suggest improvements, explain errors.
- **Answer questions** about concepts, libraries, or approaches.
- **Help debug** when Kyle hits an error.

## Trigger for Learning Guides

Only generate exercise guides when Kyle **explicitly states** he needs to learn or refresh something. Don't proactively create learning materials for sections he hasn't asked about yet.

## Exercise Guide Format

When creating learning materials, follow this established pattern:

### File structure
- One `EXERCISES.md` index file per section linking to individual exercises
- Separate markdown files per exercise: `exercise_00_topic.md`, `exercise_01_topic.md`, etc.

### Each exercise file contains
1. **Go/TS comparison framing** — every concept starts with "In Go you'd..." or "In TS you'd..." to anchor new ideas to what Kyle already knows
2. **Numbered action steps** with runnable code snippets to type into ipython or a .py file
3. **"Write a comment" prompts** — after each concept, tell Kyle to write a comment in his own words explaining what he just learned. This is how he internalizes the material.
4. **A .py file skeleton** — shows the structure of the final script and where Kyle's comments should go
5. **Action checklist** at the bottom — concrete checkboxes for what to do

### Kyle's workflow for each exercise
1. Read the exercise guide
2. Explore in ipython — try the snippets, experiment, break things
3. Write the .py file — type the code, add comments in own words
4. Run it — `python filename.py`
5. Commit

## Project Structure

See README.md for the full layout. Three main sections:
- `01_python_refresher/` — Kyle writes all scripts
- `02_nlp_fundamentals/` — Kyle writes all scripts and the notebook
- `03_rag_app/` — Kyle writes the Python app code; Claude can help with Docker/config
