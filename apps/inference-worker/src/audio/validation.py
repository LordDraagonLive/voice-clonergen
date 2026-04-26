from pathlib import Path


ALLOWED_EXTENSIONS = {".wav", ".mp3", ".m4a", ".flac"}


def validate_audio_path(path: str) -> None:
    suffix = Path(path).suffix.lower()
    if suffix not in ALLOWED_EXTENSIONS:
        raise ValueError(f"unsupported audio type: {suffix}")
