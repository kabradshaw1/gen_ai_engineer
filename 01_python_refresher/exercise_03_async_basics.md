# Exercise 3: Async Basics — `async_basics.py`

**Goal:** Understand Python's async model vs Go goroutines and TS async/await.

**Environment:** This exercise works best in a .py file from the start — async code doesn't run naturally in a plain ipython session. (Jupyter notebooks support top-level `await`, so you could also explore in a notebook.)

---

## The Mental Model Shift

This is the exercise with the biggest mental model gap from Go. Read this section carefully before writing code.

**Go concurrency:**
- `go func()` launches a goroutine on the M:N scheduler
- Goroutines run on real OS threads — they can execute in parallel on multiple cores
- The scheduler preempts goroutines — they don't have to cooperate
- Channels are the communication primitive
- You think in terms of: "launch independent tasks that run simultaneously"

**TS async/await:**
- Single-threaded event loop (Node.js)
- `await` yields control back to the event loop
- I/O operations run in a thread pool behind the scenes (libuv)
- Promises are the primitive
- You think in terms of: "start I/O, do other things while waiting, resume when done"

**Python async/await:**
- Single-threaded event loop (asyncio)
- `await` yields control to the event loop — NOTHING else runs until you `await`
- No background thread pool by default (unlike Node)
- Coroutines are the primitive
- You think in terms of: same as TS, but the mechanics are more visible

**The critical difference from Go:** Python async is **cooperative**, not preemptive. If a coroutine does a long computation without `await`-ing, everything else freezes. In Go, the scheduler would preempt it.

---

## Part A: Coroutines Are Not Functions

### Action Steps

**Step 1: A coroutine is just an object (create `async_basics.py`)**

Start your file:
```python
"""Async Basics: Python vs Go/TypeScript

[Your summary: how Python's cooperative single-threaded async differs
from Go's preemptive multi-threaded goroutines]
"""

import asyncio
import time
```

Now write this and run it:
```python
async def fetch(source: str, delay: float) -> dict:
    print(f"  Fetching from {source}...")
    await asyncio.sleep(delay)
    return {"source": source, "records": int(delay * 100)}
```

In Go, calling `fetch("db", 1.0)` would just... call the function. In Python:
```python
# This does NOT run the function. It creates a coroutine object.
coro = fetch("db", 1.0)
print(type(coro))    # <class 'coroutine'>
print(coro)          # <coroutine object fetch at 0x...>
```

You must `await` a coroutine or schedule it on the event loop. Calling it is just step one. This is like creating a goroutine but not launching it — except in Go, `go` handles both in one step.

**Write a comment:** Explain why `async def` creates a coroutine factory, not a regular function. Compare to Go where calling a function runs it immediately.

**Step 2: Running a coroutine**

You can't just `await` at the top level of a .py file (you'd get a SyntaxError). You need an event loop:
```python
async def main():
    result = await fetch("database", 0.5)
    print(f"Got: {result}")

if __name__ == "__main__":
    asyncio.run(main())
```

`asyncio.run()` creates an event loop, runs `main()` to completion, then closes the loop. In Go, goroutines just run from `main()` with no ceremony. In Python, you need this bootstrap.

**Write a comment:** Explain `asyncio.run()` — it's the entry point that starts the event loop. In Go you just `go doThing()` from `main()`. In Python you need this wrapper.

---

## Part B: Sequential vs Concurrent

### Action Steps

**Step 3: Sequential — the slow way**

```python
async def sequential():
    print("=== Sequential ===")
    start = time.perf_counter()

    r1 = await fetch("database", 1.0)
    r2 = await fetch("cache", 0.5)
    r3 = await fetch("api", 1.5)

    elapsed = time.perf_counter() - start
    print(f"Results: {[r1, r2, r3]}")
    print(f"Time: {elapsed:.2f}s")    # ~3.0s — each waited for the previous
    return elapsed
```

This is like calling three functions in order in any language — nothing special. Total time = sum of all delays.

**Step 4: Concurrent with `asyncio.gather()` — the fast way**

```python
async def concurrent():
    print("\n=== Concurrent (gather) ===")
    start = time.perf_counter()

    results = await asyncio.gather(
        fetch("database", 1.0),
        fetch("cache", 0.5),
        fetch("api", 1.5),
    )

    elapsed = time.perf_counter() - start
    print(f"Results: {results}")
    print(f"Time: {elapsed:.2f}s")    # ~1.5s — all ran concurrently
    return elapsed
```

`asyncio.gather()` is like `sync.WaitGroup` + `go func()` combined:
- Go: launch 3 goroutines, wait with `wg.Wait()`
- Python: schedule 3 coroutines, await all with `gather()`

But the mechanism is completely different. Go's goroutines can run on different CPU cores simultaneously. Python's coroutines take turns on ONE thread — when one awaits sleep, the event loop switches to another.

**Step 5: Measure the difference**

```python
async def main():
    seq_time = await sequential()
    conc_time = await concurrent()
    print(f"\nSpeedup: {seq_time / conc_time:.1f}x")
```

Run it. You should see ~2x speedup. The concurrent version finishes in the time of the slowest call, not the sum.

**Write a comment:** Explain gather vs WaitGroup. Emphasize that the speedup comes from overlapping I/O waits, not parallel execution. If these were CPU-bound tasks instead of sleeps, there would be NO speedup.

---

## Part C: create_task vs go func()

### Concept

In Go, `go doSomething()` fires and forgets — the goroutine starts immediately on a separate thread.

In Python, `asyncio.create_task()` schedules a coroutine, but it doesn't start running until the current coroutine hits an `await`. This is the cooperative scheduling in action.

### Action Steps

**Step 6: Task scheduling**

```python
async def task_demo():
    print("\n=== create_task vs go func() ===")

    task = asyncio.create_task(fetch("background", 2.0))
    print("Task created, but hasn't started yet")

    # Do other work — the task starts when we await something
    await asyncio.sleep(0.1)
    print("Now the task is running in the background")

    # Wait for the task to finish
    result = await task
    print(f"Task result: {result}")
```

In Go, this would be:
```go
ch := make(chan Result)
go func() { ch <- fetch("background", 2*time.Second) }()
// goroutine starts immediately — no need to yield
result := <-ch
```

**Write a comment:** Explain that `create_task()` is like `go func()` but the task doesn't start until you `await` something. This is the core of cooperative scheduling — nothing preempts you.

---

## Part D: Timeouts

### Concept

In Go, you'd use `context.WithTimeout` — create a context, pass it through, the operation checks the context for cancellation.

In Python, `asyncio.wait_for()` wraps a coroutine with a timeout. If the coroutine doesn't finish in time, it raises `asyncio.TimeoutError`.

### Action Steps

**Step 7: Timeout handling**

```python
async def timeout_demo():
    print("\n=== Timeout Handling ===")

    # This will timeout — fetch takes 5s, we only wait 1s
    try:
        result = await asyncio.wait_for(
            fetch("slow_service", 5.0),
            timeout=1.0
        )
        print(f"Got: {result}")
    except asyncio.TimeoutError:
        print("  Timed out after 1.0s (expected)")

    # This will succeed — fetch takes 0.5s, we wait 2s
    try:
        result = await asyncio.wait_for(
            fetch("fast_service", 0.5),
            timeout=2.0
        )
        print(f"  Got: {result}")
    except asyncio.TimeoutError:
        print("  Timed out (unexpected)")
```

In Go, the pattern is more explicit — you thread a `ctx` through every function:
```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
result, err := fetch(ctx, "slow_service")
```

Python's `wait_for` is less invasive — you don't need to change the function signature. But Go's context pattern is more powerful because it composes through deep call chains.

**Write a comment:** Compare `asyncio.wait_for()` to Go's `context.WithTimeout()`. Note the tradeoff: Python's is simpler to use, Go's is more composable.

---

## Part E: Queues vs Channels

### Concept

**Go channels** are the core concurrency primitive: `ch := make(chan string, 5)`. Send with `ch <- item`. Receive with `item := <-ch`. Range over with `for item := range ch`.

**Python's `asyncio.Queue`** serves the same purpose but is single-threaded and cooperative.

### Action Steps

**Step 8: Producer-consumer**

```python
async def producer_consumer():
    print("\n=== Producer-Consumer (Queue vs Channel) ===")
    queue: asyncio.Queue[str] = asyncio.Queue(maxsize=3)

    async def producer():
        for item in ["doc1.pdf", "doc2.txt", "doc3.md", "doc4.json"]:
            await queue.put(item)           # like ch <- item
            print(f"  Produced: {item}")

    async def consumer():
        processed = []
        for _ in range(4):
            item = await queue.get()        # like item := <-ch
            print(f"  Consumed: {item}")
            processed.append(item)
            queue.task_done()               # no Go equivalent — marks item done
        return processed

    # Run both concurrently — like launching two goroutines
    producer_task = asyncio.create_task(producer())
    consumer_task = asyncio.create_task(consumer())

    await producer_task
    results = await consumer_task
    print(f"  All done: {results}")
```

Key differences from Go channels:
- `queue.put()` and `queue.get()` are `await`-able — they block the coroutine, not the thread
- No `close()` on the queue — you have to know when to stop consuming (usually by count or sentinel value)
- `queue.task_done()` has no Go equivalent — it's for `queue.join()` which waits until all items are processed

**Write a comment:** Compare `asyncio.Queue` to Go channels. Note the key differences: no `close()`, need `task_done()`, everything is cooperative. The producer must `await` to let the consumer run — in Go, both goroutines run independently.

---

## Writing Your .py File

Your `async_basics.py` should have this flow:

```python
"""Async Basics: Python vs Go/TypeScript

[Your summary of cooperative vs preemptive concurrency.
The ONE thing you want to remember about Python async.]
"""

import asyncio
import time


# [Your comment: coroutines are objects, not function calls]
async def fetch(source: str, delay: float) -> dict:
    ...


# [Your comment: sequential — same as any language]
async def sequential():
    ...


# [Your comment: gather = WaitGroup + go func(). Same thread, overlapping waits.]
async def concurrent():
    ...


# [Your comment: create_task starts at next await, not immediately]
async def task_demo():
    ...


# [Your comment: wait_for vs context.WithTimeout]
async def timeout_demo():
    ...


# [Your comment: Queue vs channels — no close(), cooperative]
async def producer_consumer():
    ...


async def main():
    seq = await sequential()
    conc = await concurrent()
    print(f"\nSpeedup: {seq / conc:.1f}x")
    await task_demo()
    await timeout_demo()
    await producer_consumer()


if __name__ == "__main__":
    asyncio.run(main())
```

---

## Action Checklist

- [ ] Read the mental model section at the top — make sure you can explain the Go vs Python difference
- [ ] Build `async_basics.py` incrementally — write fetch + sequential + main first, run it
- [ ] Add concurrent, run it, see the speedup
- [ ] Add task_demo, timeout_demo, producer_consumer one at a time
- [ ] Write comments in your own words at each stage
- [ ] Intentionally break things: call a coroutine without `await`, try `await` at top level
- [ ] Run `python async_basics.py` end to end

When this feels solid, move on to Exercise 4.
