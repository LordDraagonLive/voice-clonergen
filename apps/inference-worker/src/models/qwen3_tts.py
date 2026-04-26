import time
from pathlib import Path

import soundfile as sf

from .base import TTSModelAdapter
from ..config import settings
from ..audio.placeholder import write_placeholder_wav


class Qwen3TTSAdapter(TTSModelAdapter):
    name = "qwen3-tts"

    def __init__(self, sample_rate: int = 24000) -> None:
        self.sample_rate = sample_rate
        self.loaded = False
        self.model = None
        self.torch = None

    def load(self) -> None:
        if settings.enable_placeholder_tts:
            self.loaded = True
            return

        try:
            import torch
            from qwen_tts import Qwen3TTSModel
        except ImportError as exc:
            raise RuntimeError(
                "qwen3-tts requires the qwen-tts package. Rebuild the worker image or install "
                "the worker with `pip install .[qwen]`. For a test beep only, set "
                "ENABLE_PLACEHOLDER_TTS=true."
            ) from exc

        dtype = _resolve_dtype(torch, settings.qwen_dtype)
        kwargs = {
            "device_map": _resolve_device_map(torch, settings.qwen_device_map),
            "dtype": dtype,
        }
        if settings.qwen_attn_implementation:
            kwargs["attn_implementation"] = settings.qwen_attn_implementation

        self.torch = torch
        self.model = Qwen3TTSModel.from_pretrained(settings.qwen_model_name, **kwargs)
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
        if not settings.enable_placeholder_tts:
            if not reference_audio_path:
                raise ValueError("Qwen3-TTS requires a ready reference audio sample")
            Path(output_path).parent.mkdir(parents=True, exist_ok=True)
            kwargs = {
                "text": text,
                "language": language or settings.qwen_language,
                "ref_audio": reference_audio_path,
            }
            if reference_text:
                kwargs["ref_text"] = reference_text
            else:
                kwargs["x_vector_only_mode"] = True
            if style_prompt:
                kwargs["instruct"] = style_prompt

            try:
                wavs, sr = self.model.generate_voice_clone(**kwargs)
            except TypeError:
                kwargs.pop("instruct", None)
                wavs, sr = self.model.generate_voice_clone(**kwargs)

            sf.write(output_path, wavs[0], sr)
            latency_ms = int((time.perf_counter() - started) * 1000)
            duration_seconds = len(wavs[0]) / sr
            return {
                "output_path": output_path,
                "latency_ms": latency_ms,
                "realtime_factor": latency_ms / max(duration_seconds * 1000, 1),
                "model_name": self.name,
                "sample_rate": sr,
                "duration_seconds": duration_seconds,
            }

        duration_seconds = write_placeholder_wav(
            output_path,
            text=text,
            sample_rate=self.sample_rate,
            frequency=440.0,
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


def _resolve_device_map(torch, value: str):
    if value == "auto":
        return "cuda:0" if torch.cuda.is_available() else "cpu"
    return value


def _resolve_dtype(torch, value: str):
    if value == "auto":
        return torch.bfloat16 if torch.cuda.is_available() else torch.float32
    if value == "bfloat16":
        return torch.bfloat16
    if value == "float16":
        return torch.float16
    if value == "float32":
        return torch.float32
    raise ValueError(f"unsupported QWEN_DTYPE: {value}")
