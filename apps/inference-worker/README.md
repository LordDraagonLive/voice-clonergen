# Inference Worker

Python worker for audio preprocessing, generation, and model benchmarks.

The current implementation is intentionally adapter-first: Qwen3-TTS and XTTS-v2 expose the same interface and emit placeholder WAV files until the production model packages are installed.

Run locally:

```bash
python -m src.worker
```
