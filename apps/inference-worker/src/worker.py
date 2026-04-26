import json
import logging
import time

from redis import Redis

from .config import settings
from .db import (
    load_generation_context,
    mark_generation_completed,
    mark_generation_failed,
    mark_generation_running,
    mark_sample_failed,
    mark_sample_processing,
    mark_sample_ready,
    update_generation_progress,
)
from .jobs.benchmark import run_benchmark
from .jobs.generate import generate_speech
from .jobs.preprocess import preprocess_sample
from .models import build_adapters

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger("personal-voice-cloner-worker")


def main() -> None:
    redis = Redis.from_url(settings.redis_url, decode_responses=True)
    adapters = build_adapters(sample_rate=settings.sample_rate)
    log.info("worker listening queue=%s", settings.job_queue)
    while True:
        item = redis.brpop(settings.job_queue, timeout=5)
        if item is None:
            continue
        _, payload = item
        job = json.loads(payload)
        try:
            handle_job(adapters, job)
        except Exception as exc:
            if job.get("type") == "preprocess_sample":
                mark_sample_failed(job["id"], str(exc))
            elif job.get("type") == "generate_speech":
                mark_generation_failed(job["id"], str(exc))
            log.exception("job failed payload=%s", payload)


def handle_job(adapters, job: dict) -> None:
    job_type = job.get("type")
    started = time.perf_counter()
    if job_type == "generate_speech":
        mark_generation_running(job["id"], "Loading generation context")
        job["data"] = {**load_generation_context(job["id"]), **job.get("data", {})}
        result = generate_speech(
            adapters,
            job,
            settings.storage_local_path,
            max_chunk_chars=settings.max_tts_chunk_chars,
            progress=lambda message: update_generation_progress(job["id"], message),
        )
        mark_generation_completed(job["id"], result)
    elif job_type == "benchmark_models":
        result = run_benchmark(adapters, job, settings.storage_local_path)
    elif job_type == "preprocess_sample":
        mark_sample_processing(job["id"])
        result = preprocess_sample(
            job,
            settings.storage_local_path,
            settings.sample_rate,
            settings.reference_min_seconds,
            settings.reference_max_seconds,
        )
        mark_sample_ready(job["id"], result)
    else:
        raise ValueError(f"unknown job type: {job_type}")
    elapsed_ms = int((time.perf_counter() - started) * 1000)
    log.info("job complete id=%s type=%s elapsed_ms=%s result=%s", job.get("id"), job_type, elapsed_ms, result)


if __name__ == "__main__":
    main()
