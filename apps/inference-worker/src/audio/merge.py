from pathlib import Path

import numpy as np
import soundfile as sf


def concatenate_wavs(input_paths: list[str], output_path: str, silence_seconds: float = 0.18) -> dict:
    if not input_paths:
        raise ValueError("no generated chunks to merge")

    audio_parts = []
    sample_rate = None
    for path in input_paths:
        audio, sr = sf.read(path, always_2d=False)
        if sample_rate is None:
            sample_rate = sr
        elif sr != sample_rate:
            raise ValueError("cannot merge WAV chunks with different sample rates")
        audio_parts.append(np.asarray(audio))
        silence = np.zeros(int(sr * silence_seconds), dtype=np.float32)
        audio_parts.append(silence)

    merged = np.concatenate(audio_parts[:-1])
    Path(output_path).parent.mkdir(parents=True, exist_ok=True)
    sf.write(output_path, merged, sample_rate)
    return {
        "output_path": output_path,
        "sample_rate": sample_rate,
        "duration_seconds": len(merged) / sample_rate,
    }
