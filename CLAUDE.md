# CLAUDE.md

## Project Intent

This is a portfolio project for a Gen AI Engineer job application. Kyle is the sole author of all Python code in this repo.

## Code Authorship Rules

- **NEVER generate Python implementation files (.py, .ipynb).** Kyle writes all Python himself.
- Code snippets in exercise guides and chat are fine — Kyle will type them out and make sure he understands them before moving on.

## What Claude SHOULD Do

- **Generate exercise/action plan markdown files** (like EXERCISES.md) when Kyle says he needs to learn something. These outline what to learn, tasks to complete, concepts to understand, and what to look for.
- **Generate infrastructure/config files** — .gitignore, Dockerfile, docker-compose.yml, requirements.txt, CLAUDE.md, README files. These aren't learning exercises.
- **Review Kyle's code** when asked — point out issues, suggest improvements, explain errors.
- **Answer questions** about concepts, libraries, or approaches.
- **Help debug** when Kyle hits an error.

## Trigger for Learning Guides

Only generate EXERCISES.md / action plan files when Kyle **explicitly states** he needs to learn or refresh something. Don't proactively create learning materials for sections he hasn't asked about yet.

## Project Structure

See README.md for the full layout. Three main sections:
- `01_python_refresher/` — Kyle writes all scripts
- `02_nlp_fundamentals/` — Kyle writes all scripts and the notebook
- `03_rag_app/` — Kyle writes the Python app code; Claude can help with Docker/config
