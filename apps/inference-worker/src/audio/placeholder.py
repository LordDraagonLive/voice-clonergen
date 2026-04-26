from pathlib import Path

import numpy as np
import soundfile as sf


def write_placeholder_wav(
    output_path: str,
    text: str,
    sample_rate: int = 24000,
    frequency: float = 440.0,
) -> float:
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)
    duration_seconds = min(max(len(text) / 14.0, 1.0), 12.0)
    t = np.linspace(0, duration_seconds, int(sample_rate * duration_seconds), endpoint=False)
    envelope = np.linspace(0.2, 0.8, t.size)
    audio = 0.08 * np.sin(2 * np.pi * frequency * t) * envelope
    sf.write(output_path, audio, sample_rate)
    return duration_seconds
