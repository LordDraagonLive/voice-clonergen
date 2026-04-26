from pathlib import Path

from ..audio.preprocess import preprocess_audio


def preprocess_sample(
    job: dict,
    output_root: str,
    sample_rate: int,
    min_duration_seconds: float,
    max_duration_seconds: float,
) -> dict:
    data = job.get("data", {})
    original_path = data["originalFilePath"]
    cleaned_output_path = data.get("cleanedOutputPath")
    if cleaned_output_path:
        output_path = str(Path(output_root) / cleaned_output_path)
    else:
        output_path = str(Path(output_root) / "cleaned" / f"{job['id']}.wav")
    return preprocess_audio(
        original_path,
        output_path,
        sample_rate=sample_rate,
        min_duration_seconds=min_duration_seconds,
        max_duration_seconds=max_duration_seconds,
    )
