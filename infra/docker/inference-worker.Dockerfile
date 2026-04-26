FROM python:3.12-slim
WORKDIR /app/apps/inference-worker
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ffmpeg \
    libsox-dev \
    sox \
    && rm -rf /var/lib/apt/lists/*
COPY apps/inference-worker/pyproject.toml ./
RUN pip install --no-cache-dir --timeout 1000 --retries 10 \
    --index-url https://download.pytorch.org/whl/cpu \
    torch torchaudio
RUN pip install --no-cache-dir --timeout 1000 --retries 10 ".[qwen]"
COPY apps/inference-worker ./
WORKDIR /app
WORKDIR /app/apps/inference-worker
CMD ["python", "-m", "src.worker"]
