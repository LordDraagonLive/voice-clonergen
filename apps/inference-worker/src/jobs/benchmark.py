from pathlib import Path


def run_benchmark(adapters, job: dict, output_root: str) -> dict:
    data = job.get("data", {})
    models = data.get("models") or ["qwen3-tts", "xtts-v2"]
    results = []
    for model_name in models:
        output_path = str(Path(output_root) / "benchmarks" / job["id"] / f"{model_name}.wav")
        results.append(
            adapters[model_name].generate(
                text=data.get("text", ""),
                reference_audio_path=data.get("referenceAudioPath", ""),
                reference_text=data.get("referenceText"),
                output_path=output_path,
                language=data.get("language"),
                style_prompt=data.get("stylePrompt"),
            )
        )
    return {"results": results}
