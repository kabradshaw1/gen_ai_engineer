# NLP Fundamentals Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create 3 reference Jupyter notebooks covering NLP fundamentals — tokenization/embeddings, similarity/search, and NER/classification.

**Architecture:** New `02_nlp_fundamentals/` directory with README, requirements.txt, and 3 reference notebooks following the same format as the Python refresher. Update root CLAUDE.md and README.md to reflect the new section.

**Tech Stack:** Jupyter, HuggingFace transformers, sentence-transformers, spaCy, scikit-learn, PyTorch

---

## File Structure

```
02_nlp_fundamentals/
  README.md                                  # Section intro and notebook index
  requirements.txt                           # NLP dependencies
  00_tokenization_and_embeddings.ipynb       # Text → tokens → vectors
  01_similarity_and_search.ipynb             # Cosine similarity, mini search engine
  02_ner_and_classification.ipynb            # spaCy NER, HuggingFace classification
CLAUDE.md                                    # Add 02_nlp_fundamentals to structure
README.md                                    # Change NLP status to "In progress"
```

---

### Task 1: Scaffold 02_nlp_fundamentals/

**Files:**
- Create: `02_nlp_fundamentals/README.md`
- Create: `02_nlp_fundamentals/requirements.txt`

- [ ] **Step 1: Create directory and README**

```bash
mkdir -p /Users/kylebradshaw/repos/gen_ai_engineer/02_nlp_fundamentals
```

Write `02_nlp_fundamentals/README.md`:

```markdown
# NLP Fundamentals — Jupyter Notebooks

Core NLP concepts for a developer who's used HuggingFace before but needs a refresher. Each notebook is self-contained — work through it cell-by-cell, retyping code and adding your own explanations.

## Notebooks

| # | File | Topic |
|---|------|-------|
| 0 | `00_tokenization_and_embeddings.ipynb` | How text becomes numbers — tokenizers, subword splitting, sentence embeddings |
| 1 | `01_similarity_and_search.ipynb` | Cosine similarity, semantic vs lexical matching, building a mini search engine |
| 2 | `02_ner_and_classification.ipynb` | Named Entity Recognition (spaCy), sentiment analysis, zero-shot classification |

## Setup

```bash
conda activate gen_ai
pip install -r requirements.txt
python -m spacy download en_core_web_sm
jupyter notebook
```

## Workflow

1. Open the reference notebook
2. Create your own copy — retype code cell-by-cell, run it, add your own comments
3. Commit your version
```

- [ ] **Step 2: Write requirements.txt**

```
jupyter
transformers
sentence-transformers
torch
spacy
scikit-learn
```

- [ ] **Step 3: Commit**

```bash
git add 02_nlp_fundamentals/
git commit -m "scaffold: create 02_nlp_fundamentals with README and requirements"
```

---

### Task 2: Update root CLAUDE.md and README.md

**Files:**
- Modify: `CLAUDE.md`
- Modify: `README.md`

- [ ] **Step 1: Update CLAUDE.md project structure section**

In `CLAUDE.md`, replace the Project Structure section:

```markdown
## Project Structure

- `01_python_refresher/` — Reference notebooks + Kyle's retyped versions
- `02_nlp_fundamentals/` — NLP reference notebooks + Kyle's retyped versions
- `_reference/` — Archived original approach (markdown guides, `.py` files)
- Other sections (RAG app) — TBD, will be addressed after NLP
```

- [ ] **Step 2: Update README.md planned sections table**

In `README.md`, change the NLP row from "Planned" to "In progress":

```markdown
| Section | Description | Status |
|---------|-------------|--------|
| `01_python_refresher/` | Core Python via Jupyter notebooks | In progress |
| `02_nlp_fundamentals/` | Tokenization, embeddings, NER, text classification | In progress |
| RAG App | FastAPI + LangChain + ChromaDB | Planned |
```

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md README.md
git commit -m "docs: add 02_nlp_fundamentals to project structure"
```

---

### Task 3: Generate 00_tokenization_and_embeddings.ipynb

**Files:**
- Create: `02_nlp_fundamentals/00_tokenization_and_embeddings.ipynb`

- [ ] **Step 1: Create the notebook with the following cells**

1. **Markdown:** "# Tokenization and Embeddings\n\n**Goal:** Understand how text becomes numbers — the two-step process that powers all of NLP.\n\n**Prereqs:** Python basics. Familiarity with lists and numpy arrays.\n\n**Libraries:** `transformers`, `sentence-transformers`"

2. **Markdown:** "## Why Text Needs to Become Numbers\n\n**Go/TS comparison:** In Go, you work with `[]byte` or `string` — text is just data you parse with string functions. In NLP, models can't see text at all. They only see numbers. Tokenization converts text → integer IDs, and embeddings convert those IDs → dense float vectors that capture meaning. Every NLP pipeline starts here."

3. **Markdown:** "## Part 1: Tokenization\n\n### Word Tokenization — The Naive Approach"

4. **Code cell:**
```python
# The simplest tokenizer: split on spaces
sentence = "The cat sat on the mat."
tokens = sentence.split()
print(f"Tokens: {tokens}")
print(f"Count: {len(tokens)}")

# Problem: punctuation sticks to words
# Problem: "don't" → should it be "do" + "n't"?
sentence2 = "I don't think it's working!"
print(f"\nNaive split: {sentence2.split()}")
```

5. **Markdown:** "**Go/TS comparison:** `strings.Split(s, \" \")` in Go or `s.split(' ')` in TS gives you the same naive result. Real tokenization is much harder — punctuation, contractions, unicode, and languages without spaces (Chinese, Japanese) all break simple splitting."

6. **Code cell:**
```python
# Experiment: try to handle punctuation yourself — see why this is hard
import re

def better_tokenize(text):
    """Split on spaces and punctuation, but keep contractions together"""
    return re.findall(r"\w+(?:'\w+)?|[^\w\s]", text)

print(better_tokenize("I don't think it's working!"))
print(better_tokenize("The U.K. startup raised $1.5B"))
# Still breaks on edge cases — this is why we use pretrained tokenizers
```

7. **Markdown:** "### ✍️ In Your Own Words\n\nWhy is splitting on spaces insufficient for tokenization? What edge cases break it? Write your answer here."

8. **Markdown:** "### Subword Tokenization — What Modern Models Use"

9. **Markdown:** "**Go/TS comparison:** No equivalent concept. Subword tokenization (BPE, WordPiece) is unique to NLP. It keeps common words whole but splits rare words into known pieces. \"unfrigginbelievable\" → \"un\", \"##fri\", \"##gg\", \"##in\", \"##bel\", \"##ie\", \"##va\", \"##ble\". This is the sweet spot between word-level (can't handle unknown words) and character-level (sequences get too long)."

10. **Code cell:**
```python
from transformers import AutoTokenizer

# Load BERT's tokenizer — uses WordPiece subword tokenization
tokenizer = AutoTokenizer.from_pretrained("bert-base-uncased")

sentence = "The cat sat on the mat"

# Three levels of tokenization:
tokens = tokenizer.tokenize(sentence)        # Human-readable tokens
ids = tokenizer.encode(sentence)             # Integer IDs (what the model sees)
decoded = tokenizer.decode(ids)              # Back to text

print(f"Tokens:  {tokens}")
print(f"IDs:     {ids}")
print(f"Decoded: {decoded}")
```

11. **Code cell:**
```python
# See how subword tokenization handles unknown words
weird_words = [
    "unfrigginbelievable",
    "supercalifragilistic",
    "ChatGPT",
    "kubernetes",
]

for word in weird_words:
    tokens = tokenizer.tokenize(word)
    print(f"{word:30s} → {tokens}")
```

12. **Code cell:**
```python
# Experiment: compare BERT vs GPT-2 tokenizers — different strategies
gpt2_tokenizer = AutoTokenizer.from_pretrained("gpt2")

test = "The quick brown fox jumps over the lazy dog"
bert_tokens = tokenizer.tokenize(test)
gpt2_tokens = gpt2_tokenizer.tokenize(test)

print(f"BERT tokens ({len(bert_tokens)}): {bert_tokens}")
print(f"GPT2 tokens ({len(gpt2_tokens)}): {gpt2_tokens}")
print(f"\nBERT vocab size: {tokenizer.vocab_size}")
print(f"GPT2 vocab size: {gpt2_tokenizer.vocab_size}")
```

13. **Markdown:** "### ✍️ In Your Own Words\n\nWhy do BERT and GPT-2 tokenize the same text differently? What determines a model's tokenization strategy? Write your answer here."

14. **Markdown:** "### Special Tokens"

15. **Code cell:**
```python
# Special tokens — metadata the model needs
print(f"Special tokens: {tokenizer.special_tokens_map}")

# See them in action
sentence = "Hello world"
ids = tokenizer.encode(sentence)
tokens = tokenizer.convert_ids_to_tokens(ids)
print(f"\nTokens with special: {tokens}")
print(f"IDs:                 {ids}")

# [CLS] = "start of input" (used for classification tasks)
# [SEP] = "separator" (marks end of input or between sentence pairs)
# [PAD] = "padding" (fills sequences to same length in a batch)
```

16. **Code cell:**
```python
# Experiment: token count matters for LLM context windows
long_text = "This is a test. " * 100
ids = tokenizer.encode(long_text)
print(f"Characters: {len(long_text)}")
print(f"Words:      {len(long_text.split())}")
print(f"Tokens:     {len(ids)}")
print(f"\nTokens ≠ words. This is why LLM context limits are in tokens, not words.")
```

17. **Markdown:** "### ✍️ In Your Own Words\n\nWhat are special tokens and why do models need them? Write your answer here."

18. **Markdown:** "## Part 2: Embeddings\n\n### From Tokens to Meaning"

19. **Markdown:** "**Go/TS comparison:** In Go, if you wanted to compare two strings for similarity, you'd use Levenshtein distance or exact matching. That only catches lexical similarity (\"cat\" vs \"bat\"). Embeddings capture *semantic* similarity — \"cat\" and \"feline\" end up close together in vector space even though they share zero characters."

20. **Code cell:**
```python
from sentence_transformers import SentenceTransformer

# Load a sentence embedding model — small and fast
model = SentenceTransformer("all-MiniLM-L6-v2")

# Embed a single sentence
sentence = "The cat sat on the mat"
embedding = model.encode(sentence)

print(f"Input:      '{sentence}'")
print(f"Output:     float array of shape {embedding.shape}")
print(f"Dimensions: {len(embedding)}")
print(f"First 10:   {embedding[:10].round(4)}")
print(f"dtype:      {embedding.dtype}")
```

21. **Code cell:**
```python
# Embeddings capture meaning, not just words
sentences = [
    "The cat sat on the mat",
    "A feline rested on the rug",           # Same meaning, different words
    "The dog chased the ball",               # Different meaning, some shared words
    "bank of the river",                     # Ambiguous word: bank
    "bank account balance",                  # Different meaning of "bank"
]

embeddings = model.encode(sentences)
print(f"Batch shape: {embeddings.shape}")
print(f"\nEach sentence → a {embeddings.shape[1]}-dimensional vector")
print(f"Regardless of sentence length!")
```

22. **Code cell:**
```python
# Experiment: embedding properties
import numpy as np

# Same input always gives same output (deterministic)
e1 = model.encode("hello world")
e2 = model.encode("hello world")
print(f"Deterministic? {np.allclose(e1, e2)}")

# Empty and very long inputs both work
e_empty = model.encode("")
e_long = model.encode("word " * 500)
print(f"Empty string shape: {e_empty.shape}")
print(f"500-word shape:     {e_long.shape}")
print(f"Same dimensions regardless of input length!")
```

23. **Markdown:** "### ✍️ In Your Own Words\n\nHow is semantic similarity (embeddings) different from lexical similarity (string matching)? Why does this matter for RAG? Write your answer here."

24. **Markdown:** "### Comparing Embedding Models"

25. **Code cell:**
```python
# Different models produce different-sized embeddings
model_small = SentenceTransformer("all-MiniLM-L6-v2")
model_large = SentenceTransformer("all-mpnet-base-v2")

sentence = "The quick brown fox"
e_small = model_small.encode(sentence)
e_large = model_large.encode(sentence)

print(f"MiniLM (small): {e_small.shape[0]} dimensions")
print(f"MPNet (large):  {e_large.shape[0]} dimensions")
print(f"\nMore dimensions = more nuance, but slower and more memory")
print(f"MiniLM is usually good enough for most use cases")
```

26. **Markdown:** "### ✍️ In Your Own Words\n\nWhy might you choose a smaller embedding model over a larger one? Write your answer here."

27. **Markdown:** "## Recap\n\n- ✅ Tokenization converts text → integer IDs that models can process\n- ✅ Subword tokenization (BPE/WordPiece) handles unknown words by splitting into known pieces\n- ✅ Different models use different tokenizers and vocabularies\n- ✅ Special tokens (`[CLS]`, `[SEP]`, `[PAD]`) provide metadata to models\n- ✅ Token count ≠ word count — matters for context windows\n- ✅ Embeddings convert text → dense float vectors capturing semantic meaning\n- ✅ Similar meanings → nearby vectors, even with different words\n- ✅ Embedding dimensions vary by model — more dimensions = more nuance\n\n**Next:** [01_similarity_and_search.ipynb](./01_similarity_and_search.ipynb) — use these embeddings to find similar text"

- [ ] **Step 2: Verify notebook opens**

```bash
cd /Users/kylebradshaw/repos/gen_ai_engineer/02_nlp_fundamentals
jupyter nbconvert --to notebook --execute 00_tokenization_and_embeddings.ipynb --output /dev/null 2>&1 || echo "May need model downloads - OK"
```

- [ ] **Step 3: Commit**

```bash
git add 02_nlp_fundamentals/00_tokenization_and_embeddings.ipynb
git commit -m "lesson: add 00_tokenization_and_embeddings reference notebook"
```

---

### Task 4: Generate 01_similarity_and_search.ipynb

**Files:**
- Create: `02_nlp_fundamentals/01_similarity_and_search.ipynb`

- [ ] **Step 1: Create the notebook with the following cells**

1. **Markdown:** "# Similarity and Search\n\n**Goal:** Learn how to compare text using cosine similarity and build a mini semantic search engine.\n\n**Prereqs:** Complete [00_tokenization_and_embeddings.ipynb](./00_tokenization_and_embeddings.ipynb) first.\n\n**Libraries:** `sentence-transformers`, `scikit-learn`, `numpy`"

2. **Markdown:** "## The Foundation of RAG\n\n**Go/TS comparison:** In your Go RAG project, you sent text to an LLM and got answers back. But how did the system know *which* documents to send? That's the retrieval step — and it works by computing similarity between the query embedding and all document embeddings. This notebook builds that from scratch."

3. **Markdown:** "## Cosine Similarity from Scratch"

4. **Markdown:** "**Go/TS comparison:** In Go, you'd compare strings with `==` or maybe Levenshtein distance. Cosine similarity compares *vectors* — it measures the angle between them, ignoring magnitude. Two vectors pointing in the same direction score 1.0, perpendicular score 0.0, opposite score -1.0."

5. **Code cell:**
```python
import numpy as np

def cosine_similarity_manual(vec_a, vec_b):
    """Compute cosine similarity from scratch — no libraries"""
    dot_product = sum(a * b for a, b in zip(vec_a, vec_b))
    magnitude_a = sum(x ** 2 for x in vec_a) ** 0.5
    magnitude_b = sum(x ** 2 for x in vec_b) ** 0.5
    if magnitude_a == 0 or magnitude_b == 0:
        return 0.0
    return dot_product / (magnitude_a * magnitude_b)

# Test with simple vectors
a = [1, 0, 0]
b = [1, 0, 0]  # Same direction
c = [0, 1, 0]  # Perpendicular
d = [-1, 0, 0] # Opposite

print(f"Same direction:  {cosine_similarity_manual(a, b):.4f}")  # 1.0
print(f"Perpendicular:   {cosine_similarity_manual(a, c):.4f}")  # 0.0
print(f"Opposite:        {cosine_similarity_manual(a, d):.4f}")  # -1.0
```

6. **Code cell:**
```python
# Verify against sklearn
from sklearn.metrics.pairwise import cosine_similarity as sklearn_cosine

# sklearn expects 2D arrays
a_2d = np.array([[1, 2, 3]])
b_2d = np.array([[4, 5, 6]])

manual = cosine_similarity_manual(a_2d[0], b_2d[0])
sklearn = sklearn_cosine(a_2d, b_2d)[0][0]

print(f"Manual:  {manual:.6f}")
print(f"sklearn: {sklearn:.6f}")
print(f"Match:   {abs(manual - sklearn) < 1e-10}")
```

7. **Markdown:** "### ✍️ In Your Own Words\n\nWhy does cosine similarity ignore vector magnitude? When would this be an advantage? Write your answer here."

8. **Markdown:** "## Semantic vs Lexical Similarity"

9. **Code cell:**
```python
from sentence_transformers import SentenceTransformer
from sklearn.metrics.pairwise import cosine_similarity

model = SentenceTransformer("all-MiniLM-L6-v2")

# Pairs to compare
pairs = [
    ("The cat sat on the mat", "A feline rested on the rug"),       # Same meaning, different words
    ("The cat sat on the mat", "The cat sat on the mat"),           # Identical
    ("bank of the river", "bank account balance"),                   # Same word, different meaning
    ("I love programming", "Coding is my passion"),                  # Semantic match
    ("The weather is nice", "Python is a programming language"),     # Unrelated
]

for s1, s2 in pairs:
    emb = model.encode([s1, s2])
    sim = cosine_similarity([emb[0]], [emb[1]])[0][0]
    print(f"{sim:.3f}  '{s1}' ↔ '{s2}'")
```

10. **Code cell:**
```python
# Experiment: lexical overlap doesn't mean semantic similarity
# These share lots of words but mean different things
tricky_pairs = [
    ("The dog bit the man", "The man bit the dog"),                 # Same words, different meaning
    ("He is not happy", "He is happy"),                              # Negation
    ("The movie was not bad", "The movie was good"),                 # Double negative ≈ positive
]

for s1, s2 in tricky_pairs:
    emb = model.encode([s1, s2])
    sim = cosine_similarity([emb[0]], [emb[1]])[0][0]
    print(f"{sim:.3f}  '{s1}' ↔ '{s2}'")

print("\nNote: models often struggle with negation — a known limitation")
```

11. **Markdown:** "### ✍️ In Your Own Words\n\nGive an example where cosine similarity on embeddings would give a better result than simple string matching. Write your answer here."

12. **Markdown:** "## Building a Mini Search Engine"

13. **Markdown:** "**Go/TS comparison:** This is literally what a vector database (ChromaDB, Pinecone, Weaviate) does under the hood. You're building the core retrieval algorithm that powers RAG — embed everything, embed the query, find the closest matches."

14. **Code cell:**
```python
# Our "document database"
documents = [
    "Python is a popular programming language for data science",
    "JavaScript is widely used for web development",
    "Machine learning models need large amounts of training data",
    "Docker containers package applications with their dependencies",
    "Neural networks are inspired by biological brain structures",
    "REST APIs use HTTP methods like GET, POST, PUT, DELETE",
    "Vector databases store embeddings for similarity search",
    "Go is known for its concurrency model using goroutines",
    "Natural language processing converts text to numerical representations",
    "Kubernetes orchestrates containerized applications at scale",
]

# Embed all documents (do this once, store the results)
doc_embeddings = model.encode(documents)
print(f"Embedded {len(documents)} documents")
print(f"Embedding matrix shape: {doc_embeddings.shape}")
```

15. **Code cell:**
```python
def search(query, documents, doc_embeddings, model, top_k=3):
    """Semantic search — embed query, find closest documents"""
    query_embedding = model.encode([query])
    similarities = cosine_similarity(query_embedding, doc_embeddings)[0]

    # Rank by similarity
    ranked = sorted(enumerate(similarities), key=lambda x: x[1], reverse=True)

    print(f"Query: '{query}'\n")
    for rank, (idx, score) in enumerate(ranked[:top_k], 1):
        print(f"  {rank}. [{score:.3f}] {documents[idx]}")
    print()

# Try different queries
search("How do I build a web app?", documents, doc_embeddings, model)
search("What is AI?", documents, doc_embeddings, model)
search("container orchestration", documents, doc_embeddings, model)
```

16. **Code cell:**
```python
# Experiment: the query doesn't need to match any document words
# Semantic search finds meaning, not keywords
search("fast concurrent server language", documents, doc_embeddings, model)
search("turning human language into numbers", documents, doc_embeddings, model)
search("packaging software for deployment", documents, doc_embeddings, model)
```

17. **Markdown:** "### ✍️ In Your Own Words\n\nHow does this mini search engine differ from a keyword search (like grep)? What are the advantages and limitations? Write your answer here."

18. **Markdown:** "## Similarity Matrix"

19. **Code cell:**
```python
# Compare all documents against each other
sim_matrix = cosine_similarity(doc_embeddings)

# Print as a readable table (first 5 docs for readability)
import pandas as pd

labels = [f"doc_{i}" for i in range(5)]
df = pd.DataFrame(
    sim_matrix[:5, :5].round(2),
    index=labels,
    columns=labels,
)
print("Similarity matrix (first 5 documents):")
print(df)
print("\nDiagonal is always 1.0 (document vs itself)")
print("High off-diagonal values indicate similar documents")
```

20. **Code cell:**
```python
# Experiment: find the most similar pair of documents
max_sim = 0
max_pair = (0, 0)
for i in range(len(documents)):
    for j in range(i + 1, len(documents)):
        if sim_matrix[i][j] > max_sim:
            max_sim = sim_matrix[i][j]
            max_pair = (i, j)

i, j = max_pair
print(f"Most similar pair (score: {max_sim:.3f}):")
print(f"  '{documents[i]}'")
print(f"  '{documents[j]}'")
```

21. **Markdown:** "### ✍️ In Your Own Words\n\nHow could a similarity matrix be useful in a real application? Think about clustering, deduplication, or recommendation systems. Write your answer here."

22. **Markdown:** "## Recap\n\n- ✅ Cosine similarity measures angle between vectors (−1 to 1)\n- ✅ Your manual implementation matches sklearn's — it's just dot product / magnitudes\n- ✅ Semantic similarity ≠ lexical similarity — embeddings capture meaning\n- ✅ Models struggle with negation — a known limitation\n- ✅ Semantic search: embed query → compare to document embeddings → rank by similarity\n- ✅ This is literally what vector databases do — you just built the core algorithm\n- ✅ Similarity matrices reveal document clusters and relationships\n\n**Next:** [02_ner_and_classification.ipynb](./02_ner_and_classification.ipynb) — extract structure from text"

- [ ] **Step 2: Commit**

```bash
git add 02_nlp_fundamentals/01_similarity_and_search.ipynb
git commit -m "lesson: add 01_similarity_and_search reference notebook"
```

---

### Task 5: Generate 02_ner_and_classification.ipynb

**Files:**
- Create: `02_nlp_fundamentals/02_ner_and_classification.ipynb`

- [ ] **Step 1: Create the notebook with the following cells**

1. **Markdown:** "# NER and Text Classification\n\n**Goal:** Extract structured entities from text and classify text into categories.\n\n**Prereqs:** Complete [00_tokenization_and_embeddings.ipynb](./00_tokenization_and_embeddings.ipynb) first.\n\n**Libraries:** `spacy`, `transformers`\n\n**Setup:** Run `python -m spacy download en_core_web_sm` if you haven't already."

2. **Markdown:** "## Part 1: Named Entity Recognition (NER)\n\n**Go/TS comparison:** In a web service, you parse JSON into structs with known fields. NER does something similar for natural language — it finds structured entities (people, places, organizations, dates, money) hiding in unstructured text. Think of it as `json.Unmarshal()` for English."

3. **Code cell:**
```python
import spacy

# Load the small English model
nlp = spacy.load("en_core_web_sm")

# Process a sentence — NER happens automatically
doc = nlp("Apple is looking at buying U.K. startup for $1 billion")

print("Entities found:")
for ent in doc.ents:
    print(f"  {ent.text:20s}  {ent.label_:10s}  ({spacy.explain(ent.label_)})")
```

4. **Code cell:**
```python
# Try different types of text
texts = [
    "Elon Musk announced that Tesla will open a new factory in Austin, Texas by March 2025.",
    "The European Central Bank raised interest rates by 0.25% on Thursday.",
    "Dr. Sarah Chen published her research on quantum computing at MIT.",
]

for text in texts:
    doc = nlp(text)
    print(f"Text: {text[:60]}...")
    for ent in doc.ents:
        print(f"  {ent.text:25s} → {ent.label_}")
    print()
```

5. **Code cell:**
```python
# Experiment: NER isn't perfect — see where it fails
tricky_texts = [
    "I love Paris Hilton",                    # Paris = person or place?
    "Apple stock rose after the iPhone launch", # Apple = company, not fruit
    "Jordan played basketball in Jordan",      # Same word, different entities
    "Reading is a city in England",            # Reading = place, not the verb
]

for text in tricky_texts:
    doc = nlp(text)
    print(f"'{text}'")
    if doc.ents:
        for ent in doc.ents:
            print(f"  {ent.text} → {ent.label_}")
    else:
        print("  (no entities found)")
    print()
```

6. **Markdown:** "### ✍️ In Your Own Words\n\nWhy does NER sometimes get things wrong? What makes entity recognition ambiguous? Write your answer here."

7. **Markdown:** "### Entity Frequency Analysis"

8. **Code cell:**
```python
# Process a longer text and count entity types
text = """
Google announced a partnership with NASA to advance quantum computing research.
CEO Sundar Pichai met with NASA Administrator Bill Nelson in Washington, D.C.
on January 15, 2025. The deal is worth approximately $500 million over five years.
Microsoft and Amazon are also investing heavily in quantum technology.
The European Union has allocated €1 billion for quantum research through 2030.
"""

doc = nlp(text)

# Count entity types
from collections import Counter
entity_counts = Counter(ent.label_ for ent in doc.ents)

print("Entity frequency:")
for label, count in entity_counts.most_common():
    print(f"  {label:10s} ({spacy.explain(label):30s}): {count}")

print(f"\nAll entities:")
for ent in doc.ents:
    print(f"  {ent.text:25s} → {ent.label_}")
```

9. **Markdown:** "### ✍️ In Your Own Words\n\nHow could entity frequency analysis be useful in a real data pipeline? Write your answer here."

10. **Markdown:** "## Part 2: Text Classification\n\n**Go/TS comparison:** In a web service, you route requests to handlers based on the URL path or method. Text classification does the same for natural language — it assigns a label to text so you can route it, filter it, or act on it. Sentiment analysis is the most common example."

11. **Markdown:** "### Sentiment Analysis"

12. **Code cell:**
```python
from transformers import pipeline

# Load a pretrained sentiment analysis model
classifier = pipeline("sentiment-analysis")

sentences = [
    "I absolutely love this product, it's amazing!",
    "This is the worst experience I've ever had.",
    "The movie was okay, nothing special.",
    "I'm not unhappy with the results.",           # Double negative
    "Great product, terrible customer service.",     # Mixed
]

for sentence in sentences:
    result = classifier(sentence)[0]
    print(f"[{result['label']:8s} {result['score']:.3f}]  {sentence}")
```

13. **Code cell:**
```python
# Experiment: sarcasm and nuance — where models struggle
sarcastic = [
    "Oh great, another Monday morning.",
    "Wow, what a surprise, the train is late again.",
    "Sure, because that worked so well last time.",
]

print("Sarcasm test (models usually miss this):")
for s in sarcastic:
    result = classifier(s)[0]
    print(f"  [{result['label']:8s} {result['score']:.3f}]  {s}")
```

14. **Markdown:** "### ✍️ In Your Own Words\n\nWhy do models struggle with sarcasm? What would it take to detect sarcasm reliably? Write your answer here."

15. **Markdown:** "### Zero-Shot Classification\n\n**Go/TS comparison:** This is the powerful one. Imagine a Go HTTP router where you could add new routes at runtime without writing new handler code. Zero-shot classification lets you define categories *at runtime* — no training needed. The model uses natural language inference to score how well the text matches each label."

16. **Code cell:**
```python
# Zero-shot: you provide the labels at runtime
zero_shot = pipeline("zero-shot-classification")

text = "The new iPhone has an amazing camera but the battery life is disappointing"

result = zero_shot(
    text,
    candidate_labels=["technology", "sports", "food", "politics"],
)

print(f"Text: {text}\n")
for label, score in zip(result['labels'], result['scores']):
    print(f"  {label:15s}  {score:.3f}")
```

17. **Code cell:**
```python
# Experiment: change the labels and watch scores shift
texts_and_labels = [
    (
        "The stock market crashed after the Fed raised interest rates",
        ["finance", "politics", "technology", "sports"],
    ),
    (
        "The patient was diagnosed with type 2 diabetes",
        ["medical", "fitness", "nutrition", "insurance"],
    ),
    (
        "The team scored a last-minute goal to win the championship",
        ["sports", "business", "entertainment", "politics"],
    ),
]

for text, labels in texts_and_labels:
    result = zero_shot(text, candidate_labels=labels)
    print(f"Text: {text[:60]}...")
    for label, score in zip(result['labels'][:3], result['scores'][:3]):
        print(f"  {label:15s}  {score:.3f}")
    print()
```

18. **Code cell:**
```python
# Experiment: labels that are close in meaning
text = "I just bought a new gaming laptop"
label_sets = [
    ["tech", "technology", "gadgets", "electronics"],    # Synonyms
    ["shopping", "gaming", "computers", "hardware"],     # Different aspects
]

for labels in label_sets:
    result = zero_shot(text, candidate_labels=labels)
    print(f"Labels: {labels}")
    for label, score in zip(result['labels'], result['scores']):
        print(f"  {label:15s}  {score:.3f}")
    print()
```

19. **Markdown:** "### ✍️ In Your Own Words\n\nHow does zero-shot classification work under the hood? Why can it classify text into categories it was never trained on? Write your answer here."

20. **Markdown:** "### Confidence Thresholds"

21. **Code cell:**
```python
# Not all predictions are confident — filter by threshold
texts = [
    "Apple released a new MacBook Pro with M3 chip",
    "I went for a walk in the park this morning",
    "The government announced new tax policies for 2025",
    "My cat knocked over the Christmas tree again",
    "SpaceX successfully launched 50 Starlink satellites",
]

labels = ["technology", "politics", "lifestyle", "science"]
threshold = 0.5

print(f"Threshold: {threshold}\n")
for text in texts:
    result = zero_shot(text, candidate_labels=labels)
    top_label = result['labels'][0]
    top_score = result['scores'][0]
    status = "✓" if top_score >= threshold else "✗ (low confidence)"
    print(f"  [{top_score:.3f}] {top_label:12s} {status}  — {text[:50]}")
```

22. **Markdown:** "### ✍️ In Your Own Words\n\nWhen would you use a confidence threshold in production? What's the trade-off between a high and low threshold? Write your answer here."

23. **Markdown:** "## Recap\n\n- ✅ **NER** extracts structured entities (PERSON, ORG, GPE, DATE, MONEY) from text\n- ✅ spaCy provides production-grade NER out of the box\n- ✅ NER is statistical, not perfect — ambiguity is inherent in language\n- ✅ **Sentiment analysis** classifies text as positive/negative with confidence scores\n- ✅ Models struggle with sarcasm, negation, and mixed sentiment\n- ✅ **Zero-shot classification** defines categories at runtime — no training needed\n- ✅ It works via natural language inference (NLI) — treating each label as a hypothesis\n- ✅ Confidence thresholds filter uncertain predictions — trade precision for recall\n\n**You've completed the NLP Fundamentals section!** These concepts (tokenization, embeddings, similarity, NER, classification) are the building blocks used in the RAG app."

- [ ] **Step 2: Commit**

```bash
git add 02_nlp_fundamentals/02_ner_and_classification.ipynb
git commit -m "lesson: add 02_ner_and_classification reference notebook"
```

---

## Summary

| Task | What | Commit |
|------|------|--------|
| 1 | Scaffold `02_nlp_fundamentals/` | `scaffold: create 02_nlp_fundamentals with README and requirements` |
| 2 | Update root CLAUDE.md and README.md | `docs: add 02_nlp_fundamentals to project structure` |
| 3 | Generate `00_tokenization_and_embeddings.ipynb` | `lesson: add 00_tokenization_and_embeddings reference notebook` |
| 4 | Generate `01_similarity_and_search.ipynb` | `lesson: add 01_similarity_and_search reference notebook` |
| 5 | Generate `02_ner_and_classification.ipynb` | `lesson: add 02_ner_and_classification reference notebook` |
