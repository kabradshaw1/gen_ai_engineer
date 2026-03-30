# Exercises: Environments

After completing the reference notebook, test yourself with these.

## ipython Exercises

Type each into ipython. **Predict the output BEFORE you hit enter.**

### 1. Module search path

```python
import sys
print(len(sys.path) > 0)
print(sys.path[0] == '')
```

### 2. Type of a module

```python
import os
print(type(os))
print(type(os.path))
```

### 3. Reloading a module

```python
import importlib
import math
math.pi = 3
print(math.pi)
importlib.reload(math)
print(math.pi)
```

### 4. The __name__ trick

```python
def main():
    return "running"

print(__name__)
print(main())
```

### 5. sys.modules cache

```python
import sys
import json
print('json' in sys.modules)
del sys.modules['json']
print('json' in sys.modules)
import json
print('json' in sys.modules)
```

### 6. Chained magic commands

```python
%timeit -n 1 -r 1 sum(range(100))
%timeit -n 1 -r 1 sum(list(range(100)))
```

## .py Challenge

Create `env_check.py` that produces this exact output:

```
=== Environment Report ===
Python: 3.11
Platform: [your platform, e.g. macOS-15.3.2-arm64-arm-64bit]
Executable: [your python path]

Installed packages:
  numpy ✓
  pandas ✓
  jupyter ✓

sys.path entries: [number]
```

Notes: The bracketed values will vary by machine — match the format, not the exact values. Use only the standard library plus the packages listed.

No hints. No function signatures. Figure it out from the output.
