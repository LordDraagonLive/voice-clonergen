# RunPod Notes

Use RunPod for the `inference-worker` when GPU inference is required.

1. Build a GPU image from `infra/docker/inference-worker.Dockerfile` and extend it with CUDA, PyTorch, Qwen3-TTS, and XTTS-v2 dependencies.
2. Mount persistent storage or configure S3/MinIO-compatible storage for reference and output audio.
3. Point `REDIS_URL` at a queue reachable from the pod, or replace Redis with a secure RPC worker endpoint.
4. Set `INFERENCE_RPC_URL` on the Go API to the worker endpoint once gRPC/ConnectRPC is enabled.
5. Restrict network access with firewall rules or private networking.
