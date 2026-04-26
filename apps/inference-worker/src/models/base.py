from abc import ABC, abstractmethod
from dataclasses import dataclass


@dataclass
class GenerationResult:
    output_path: str
    latency_ms: int
    realtime_factor: float
    model_name: str
    sample_rate: int
    duration_seconds: float


class TTSModelAdapter(ABC):
    name: str

    @abstractmethod
    def load(self) -> None:
        raise NotImplementedError

    @abstractmethod
    def generate(
        self,
        text: str,
        reference_audio_path: str,
        reference_text: str | None,
        output_path: str,
        language: str | None = None,
        style_prompt: str | None = None,
    ) -> dict:
        raise NotImplementedError
