# personal-voice-cloner

Self-hosted personal voice cloning app for consent-based voice profiles, sample preprocessing, text-to-speech generation, and benchmark-first model comparison.

This project is designed as a clean MVP scaffold. The Go API and React UI expose the real product workflow, while the Python worker contains adapter placeholders for Qwen3-TTS and XTTS-v2 so production inference can be inserted without changing the public API shape.

## Safety and Consent

This application is only for cloning voices you own or have explicit permission to use. Creating a voice profile requires an affirmative consent checkbox, and the consent text is stored with the profile. The app intentionally avoids impersonation, deception, identity-bypass, or unauthorized cloning features.

## Architecture

```text
frontend
  -> Go API gateway
  -> PostgreSQL metadata
  -> Redis queue
  -> local filesystem or MinIO/S3 object storage
  -> Python GPU inference worker
  -> Qwen3-TTS / XTTS-v2 model adapters
```

## Monorepo Layout

```text
apps/web                  Vite + React UI
apps/api                  Go Chi API gateway
apps/inference-worker     Python preprocessing and model adapters
proto                     Future gRPC/ConnectRPC contract
infra/docker              Local container definitions
infra/runpod              GPU worker deployment notes
infra/scripts             Utility scripts
```

## Local Setup

1. Copy `.env.example` to `.env` and adjust values if needed.
2. Start local services:

```bash
docker compose up --build
```

3. Open the app at `http://localhost:5173`.
4. API health check is available at `http://localhost:8081/health` by default.

PostgreSQL initializes the schema from `apps/api/internal/db/migrations/001_init.sql` on first container startup.

## Environment Variables

`DATABASE_URL`: PostgreSQL connection string.

`REDIS_URL`: Redis queue connection string.

`STORAGE_PROVIDER`: `local` for development, later `s3`.

`STORAGE_LOCAL_PATH`: shared path for uploaded and generated audio.

`S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET`: S3-compatible storage settings for MinIO or production object storage.

`INFERENCE_RPC_URL`: future RPC endpoint for remote GPU worker.

`DEFAULT_MODEL`: defaults to `qwen3-tts`.

`MAX_UPLOAD_MB`: maximum accepted upload size.

`ALLOWED_AUDIO_TYPES`: comma-separated allowed MIME types.

`MAX_TTS_CHUNK_CHARS`: maximum text size sent to the model per generated chunk.

`ENABLE_PLACEHOLDER_TTS`: set to `true` only when you intentionally want test beep output without a model.

`REFERENCE_MIN_SECONDS`, `REFERENCE_MAX_SECONDS`: reference window bounds saved during preprocessing.

`QWEN_MODEL_NAME`, `QWEN_DEVICE_MAP`, `QWEN_DTYPE`, `QWEN_ATTN_IMPLEMENTATION`, `QWEN_LANGUAGE`: Qwen3-TTS runtime configuration.

`ENABLE_REAL_XTTS`: set to `true` to use the real Coqui XTTS-v2 adapter instead of placeholder WAV output.

`XTTS_MODEL_NAME`, `XTTS_DEVICE`, `XTTS_LANGUAGE`: XTTS-v2 runtime configuration.

## Run API

```bash
cd apps/api
go mod download
go run ./cmd/server
```

The API exposes:

- `GET /health`
- `POST /api/voice-profiles`
- `GET /api/voice-profiles`
- `GET /api/voice-profiles/{id}`
- `POST /api/voice-profiles/{id}/samples`
- `POST /api/generations`
- `GET /api/generations`
- `GET /api/generations/{id}`
- `GET /api/generations/{id}/download`
- `POST /api/benchmarks`

## Run Inference Worker

```bash
cd apps/inference-worker
pip install .
python -m src.worker
```

The current Docker worker installs `qwen-tts` and uses Qwen3-TTS for `qwen3-tts` generations. First generation may be slow because model weights are downloaded and cached. If you need a fast pipeline smoke test, set `ENABLE_PLACEHOLDER_TTS=true`.

To try real XTTS-v2 in a compatible GPU Python environment:

```bash
pip install ".[xtts]"
ENABLE_REAL_XTTS=true XTTS_DEVICE=cuda python -m src.worker
```

The default Docker worker keeps XTTS disabled so the local stack can run without downloading model weights.

## Run Frontend

```bash
cd apps/web
npm install
npm run dev
```

Set `VITE_API_URL` if the API is not running at `http://localhost:8081`.

## Adding a Model Adapter

1. Create `apps/inference-worker/src/models/new_model.py`.
2. Implement `TTSModelAdapter` from `models/base.py`.
3. Return generation metadata with `output_path`, `latency_ms`, `realtime_factor`, `model_name`, `sample_rate`, and `duration_seconds`.
4. Register the adapter in `models/__init__.py`.
5. Add the model name to the frontend model list and API validation once hard validation is added.

The adapter layer is ready for Fish Speech, CosyVoice, and F5-TTS without changing the benchmark API shape.

## Audio Pipeline

The preprocessing module accepts WAV, MP3, M4A, and FLAC. It converts audio to mono, resamples to the configured sample rate, trims silence, peak normalizes, rejects too-short segments, writes cleaned WAV, and returns basic quality metadata.

Production follow-ups should add stronger VAD using webrtcvad or silero-vad, clipping/noise analysis, segment splitting around 5-15 seconds, and optional denoising.

## RunPod Deployment Notes

Run the API, database, queue, and object storage on a normal VPS first. Run the Python worker on RunPod when real GPU inference is enabled.

Use a custom worker image with CUDA, PyTorch, Qwen3-TTS or vLLM-Omni dependencies, and XTTS-v2 dependencies. Point the worker at the shared queue or expose a secured RPC endpoint. Set `INFERENCE_RPC_URL` on the API to that secured endpoint once the generated RPC client/server is wired in.

## License

This project is released under the HelagenHQ Source-Available License. The code is publicly viewable for evaluation and personal experimentation, but redistribution, commercial use, hosted service use, and sublicensing require prior written permission from HelagenHQ.com. Attribution to HelagenHQ.com is required when using, referencing, demonstrating, publishing about, or building upon this project.

## Current Implementation Status

Implemented:

- Monorepo structure
- Go API skeleton with real endpoints
- PostgreSQL schema migration
- Local storage abstraction
- Redis queue abstraction
- Python worker skeleton
- Qwen3-TTS and XTTS-v2 adapter placeholders
- Optional real XTTS-v2 adapter path
- Generation progress messages
- Long-text chunking and WAV merge
- Simple React UI
- Docker Compose for local development

Next:

- Wire generated gRPC/ConnectRPC client and server
- Update generation job rows when worker completes
- Add production S3 storage implementation
- Replace placeholder WAV output with real Qwen3-TTS and XTTS-v2 inference
- Add authentication before exposing beyond localhost
