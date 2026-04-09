FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
# Pinned versions; no need for --no-cache-dir cleanup beyond the slim base.
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py .

EXPOSE 11434

# Bind to 0.0.0.0 so other compose services can reach us via the service name.
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "11434"]
