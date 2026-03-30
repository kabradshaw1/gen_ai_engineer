# Lists

"""
python calls this a list.  When compared to arrays (js) or slices (go)
list operate very differently.  

In python, a list is an array of pointers
not values scattered across the heap. [1, "hello", 3.14, None] is valid in 
a list.  You can add type hints like list[str], but they are not inforced 
at runtime.

Python lists over-allocate capacity.  You never see the capacity, there is 
no cap().  It just grows automatically on .append()

Because python list stores pointers, not values, iterating a list of numbers
is much slower than other langues.  This is exactly why NumPy exists: 
numpy.ndarray stores actual values contiguously.

Practical rules of thumb from Go/TS:                                              
    - Where you'd use a []T slice in Go → use a list in Python     
    - Where you'd care about cap() in Go → don't worry about it in Python             
    - Where you'd use a typed array for performance in Go → use numpy in Python
    - Where you'd use a tuple in TS ([string, number]) → use a tuple in Python        
(immutable, fixed structure)     

The big gotcha: because lists hold references, a = b doesn't copy - both 
variables point to the same list.  use a = b.copy() or a = b[:] for a shallow copy.
A shallow copy creates a new list, but the elements inside still point to 
"""
numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9]

print(numbers)

numbers[::-1]