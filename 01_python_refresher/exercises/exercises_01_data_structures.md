# Exercises: Data Structures

After completing the reference notebook, test yourself with these.

## ipython Exercises

Type each into ipython. **Predict the output BEFORE you hit enter.**

### 1. Aliasing through a function

```python
def append_to(item, target=[]):
    target.append(item)
    return target

print(append_to(1))
print(append_to(2))
print(append_to(3))
```

### 2. Comprehension with condition and transform

```python
words = ["hello", "WORLD", "Python", "GO", "ts"]
result = {w.lower(): len(w) for w in words if len(w) > 2}
print(result)
```

### 3. Nested unpacking

```python
data = [(1, (2, 3)), (4, (5, 6))]
result = [a + b + c for a, (b, c) in data]
print(result)
```

### 4. Set operations chain

```python
a = {1, 2, 3, 4, 5}
b = {4, 5, 6, 7}
c = {5, 6, 8, 9}
print(a & b | c)
print((a & b) | c)
print(a & (b | c))
```

### 5. Generator exhaustion

```python
gen = (x for x in range(3))
print(list(gen))
print(list(gen))
print(sum(x for x in range(3)))
```

### 6. Dict merge operators

```python
defaults = {"color": "red", "size": 10, "visible": True}
overrides = {"size": 20, "name": "box"}
merged = defaults | overrides
print(merged)
print(len(merged))
```

### 7. Slice assignment with different length

```python
a = [0, 1, 2, 3, 4]
a[1:4] = [10, 20, 30, 40, 50]
print(a)
print(len(a))
```

### 8. Tuple as dict key with computation

```python
grid = {}
for x in range(3):
    for y in range(3):
        grid[(x, y)] = x * 3 + y

diag = [grid[(i, i)] for i in range(3)]
print(diag)
```

## .py Challenge

Create `data_cruncher.py` that produces this exact output:

```
Original: [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]
Unique sorted: [1, 2, 3, 4, 5, 6, 9]
Frequency: {1: 2, 2: 1, 3: 2, 4: 1, 5: 3, 6: 1, 9: 1}
Most common: 5 (3 times)
Pairs that sum to 10: [(1, 9), (4, 6), (5, 5)]
Squared evens: [4, 36]
Running total: [3, 4, 8, 9, 14, 23, 25, 31, 36, 39, 44]
```

No hints. No function signatures. Figure it out from the output.
