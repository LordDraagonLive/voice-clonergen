import time
from pathlib import Path

import soundfile as sf

from .base import TTSModelAdapter
from ..config import settings
from ..audio.placeholder import write_placeholder_wav


class XTTSV2Adapter(TTSModelAdapter):
    name = "xtts-v2"

    def __init__(self, sample_rate: int = 24000) -> None:
        self.sample_rate = sample_rate
        self.loaded = False
        self.tts = None
        self.device = "cpu"

    def load(self) -> None:
        if not settings.enable_real_xtts and settings.enable_placeholder_tts:
            self.loaded = True
            return
        if not settings.enable_real_xtts:
            raise RuntimeError(
                "xtts-v2 real inference is disabled. Set ENABLE_REAL_XTTS=true, or set "
                "ENABLE_PLACEHOLDER_TTS=true for test beep output."
            )

        try:
            import torch
            from TTS.api import TTS
        except ImportError as exc:
            raise RuntimeError(
                "ENABLE_REAL_XTTS=true but Coqui TTS is not installed. "
                "Install the worker with `pip install .[xtts]` in a compatible Python environment."
            ) from exc

        if settings.xtts_device == "auto":
            self.device = "cuda" if torch.cuda.is_available() else "cpu"
        else:
            self.device = settings.xtts_device

        if self.device == "cuda" and not torch.cuda.is_available():
            raise RuntimeError("XTTS_DEVICE=cuda was requested but CUDA is not available")

        self.tts = TTS(settings.xtts_model_name).to(self.device)
        self.loaded = True

    def generate(
        self,
        text: str,
        reference_audio_path: str,
        reference_text: str | None,
        output_path: str,
        language: str | None = None,
        style_prompt: str | None = None,
    ) -> dict:
        if not self.loaded:
            self.load()
        started = time.perf_counter()
        if settings.enable_real_xtts:
            if not reference_audio_path:
                raise ValueError("XTTS-v2 requires a ready reference audio sample")
            Path(output_path).parent.mkdir(parents=True, exist_ok=True)
            self.tts.tts_to_file(
                text=text,
                file_path=output_path,
                speaker_wav=reference_audio_path,
                language=language or settings.xtts_language,
            )
            info = sf.info(output_path)
            latency_ms = int((time.perf_counter() - started) * 1000)
            duration_seconds = info.frames / info.samplerate
            return {
                "output_path": output_path,
                "latency_ms": latency_ms,
                "realtime_factor": latency_ms / max(duration_seconds * 1000, 1),
                "model_name": self.name,
                "sample_rate": info.samplerate,
                "duration_seconds": duration_seconds,
            }

        if not settings.enable_placeholder_tts:
            raise RuntimeError("placeholder TTS is disabled")
        duration_seconds = write_placeholder_wav(
            output_path,
            text=text,
            sample_rate=self.sample_rate,
            frequency=554.37,
        )
        latency_ms = int((time.perf_counter() - started) * 1000)
        return {
            "output_path": output_path,
            "latency_ms": latency_ms,
            "realtime_factor": latency_ms / max(duration_seconds * 1000, 1),
            "model_name": self.name,
            "sample_rate": self.sample_rate,
            "duration_seconds": duration_seconds,
        }
