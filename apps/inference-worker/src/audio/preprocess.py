from pathlib import Path

import librosa
import soundfile as sf

from .normalize import peak_normalize
from .validation import validate_audio_path
from .vad import trim_silence


def preprocess_audio(
    input_path: str,
    output_path: str,
    sample_rate: int = 24000,
    min_duration_seconds: float = 3.0,
    max_duration_seconds: float = 15.0,
) -> dict:
    validate_audio_path(input_path)
    audio, _ = librosa.load(input_path, sr=sample_rate, mono=True)
    audio = trim_silence(audio)
    audio = peak_normalize(audio)
    if len(audio) > int(sample_rate * max_duration_seconds):
        audio = _select_reference_window(audio, sample_rate, max_duration_seconds)
    duration = len(audio) / sample_rate if sample_rate else 0
    if duration < min_duration_seconds:
        raise ValueError("audio segment is too short after trimming")
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)
    sf.write(output_path, audio, sample_rate)
    return {
        "cleaned_file_path": output_path,
        "duration_seconds": duration,
        "sample_rate": sample_rate,
        "quality_score": _quality_score(audio),
    }


def _quality_score(audio) -> float:
    if audio.size == 0:
        return 0.0
    peak = float(abs(audio).max())
    if peak > 0.999:
        return 0.5
    return min(max(peak, 0.0), 1.0)


def _select_reference_window(audio, sample_rate: int, max_duration_seconds: float):
    window_size = int(sample_rate * max_duration_seconds)
    hop_size = max(int(sample_rate * 1.0), 1)
    best_start = 0
    best_score = -1.0
    for start in range(0, max(len(audio) - window_size + 1, 1), hop_size):
        window = audio[start : start + window_size]
        if len(window) < window_size:
            break
        score = float((window**2).mean())
        if score > best_score:
            best_score = score
            best_start = start
    return audio[best_start : best_start + window_size]
