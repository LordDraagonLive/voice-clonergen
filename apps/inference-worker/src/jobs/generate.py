from pathlib import Path
import time

from ..audio.merge import concatenate_wavs
from ..audio.text import split_text


def generate_speech(adapters, job: dict, output_root: str, max_chunk_chars: int = 350, progress=None) -> dict:
    data = job.get("data", {})
    model_name = data.get("modelName") or "qwen3-tts"
    adapter = adapters[model_name]
    text = data.get("text", "")
    chunks = split_text(text, max_chars=max_chunk_chars)
    if not chunks:
        raise ValueError("generation text is empty")

    final_output_path = data.get("outputPath") or str(Path(output_root) / "generated" / f"{job['id']}.wav")
    started = time.perf_counter()
    if len(chunks) == 1:
        if progress:
            progress("Generating audio")
        return adapter.generate(
            text=chunks[0],
            reference_audio_path=data.get("referenceAudioPath", ""),
            reference_text=data.get("referenceText"),
            output_path=final_output_path,
            language=data.get("language"),
            style_prompt=data.get("stylePrompt"),
        )

    chunk_paths: list[str] = []
    chunk_results: list[dict] = []
    chunk_root = Path(output_root) / "generated" / job["id"]
    for index, chunk in enumerate(chunks, start=1):
        if progress:
            progress(f"Generating chunk {index}/{len(chunks)}")
        chunk_path = str(chunk_root / f"chunk_{index:03d}.wav")
        chunk_results.append(
            adapter.generate(
                text=chunk,
                reference_audio_path=data.get("referenceAudioPath", ""),
                reference_text=data.get("referenceText"),
                output_path=chunk_path,
                language=data.get("language"),
                style_prompt=data.get("stylePrompt"),
            )
        )
        chunk_paths.append(chunk_path)

    if progress:
        progress("Merging audio chunks")
    merged = concatenate_wavs(chunk_paths, final_output_path)
    latency_ms = int((time.perf_counter() - started) * 1000)
    duration_seconds = merged["duration_seconds"]
    return {
        "output_path": final_output_path,
        "latency_ms": latency_ms,
        "realtime_factor": latency_ms / max(duration_seconds * 1000, 1),
        "model_name": model_name,
        "sample_rate": merged["sample_rate"],
        "duration_seconds": duration_seconds,
        "chunks": len(chunk_results),
    }
