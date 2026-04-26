import numpy as np


def trim_silence(audio: np.ndarray, threshold: float = 0.01) -> np.ndarray:
    if audio.size == 0:
        return audio
    mask = np.abs(audio) > threshold
    if not mask.any():
        return audio
    indexes = np.flatnonzero(mask)
    return audio[indexes[0] : indexes[-1] + 1]
