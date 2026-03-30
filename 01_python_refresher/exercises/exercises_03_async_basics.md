# Exercises: Async Basics

After completing the reference notebook, test yourself with these.

## ipython Exercises

Type each into ipython. **Predict the output BEFORE you hit enter.**

### 1. Coroutine without await

```python
import asyncio

async def double(x):
    return x * 2

coro = double(5)
print(type(coro))
result = await coro
print(result)
```

### 2. Gather ordering

```python
import asyncio

async def delayed_value(val, delay):
    await asyncio.sleep(delay)
    return val

results = await asyncio.gather(
    delayed_value("slow", 0.3),
    delayed_value("fast", 0.1),
    delayed_value("medium", 0.2),
)
print(results)
```

### 3. Task cancellation

```python
import asyncio

async def long_task():
    try:
        await asyncio.sleep(10)
        return "done"
    except asyncio.CancelledError:
        return "cancelled"

task = asyncio.create_task(long_task())
await asyncio.sleep(0.1)
task.cancel()
try:
    result = await task
except asyncio.CancelledError:
    result = "caught outside"
print(result)
```

### 4. Semaphore fairness

```python
import asyncio

sem = asyncio.Semaphore(1)
order = []

async def worker(name, delay):
    async with sem:
        order.append(f"{name}-start")
        await asyncio.sleep(delay)
        order.append(f"{name}-end")

await asyncio.gather(
    worker("A", 0.2),
    worker("B", 0.1),
    worker("C", 0.1),
)
print(order)
```

### 5. Return exceptions vs raise

```python
import asyncio

async def fail():
    raise ValueError("boom")

async def succeed():
    return "ok"

results = await asyncio.gather(
    succeed(),
    fail(),
    succeed(),
    return_exceptions=True,
)
print([type(r).__name__ for r in results])
print([str(r) for r in results])
```

### 6. Async generator collection

```python
import asyncio

async def trickle(n):
    for i in range(n):
        await asyncio.sleep(0.05)
        yield i * i

squares = [x async for x in trickle(5)]
print(squares)
```

## .py Challenge

Create `async_pipeline.py` that produces this exact output when run with `python async_pipeline.py`:

```
Starting pipeline...

[fetch] Fetching user_001... done (0.3s)
[fetch] Fetching user_002... done (0.1s)
[fetch] Fetching user_003... done (0.2s)
All fetches complete in 0.3s

[process] Processing user_001... done
[process] Processing user_002... done
[process] Processing user_003... done

Pipeline results:
  user_001: score=82 (grade: B)
  user_002: score=95 (grade: A)
  user_003: score=67 (grade: D)

Pipeline finished in ~0.3s (concurrent fetch + sequential process)
```

Notes: Timing values are approximate. The fetches must run concurrently (total fetch time ≈ max delay, not sum). Processing is sequential. Grading: A=90+, B=80+, C=70+, D=60+, F=below 60.

No hints. No function signatures. Figure it out from the output.
