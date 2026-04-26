from .qwen3_tts import Qwen3TTSAdapter
from .xtts_v2 import XTTSV2Adapter


def build_adapters(sample_rate: int = 24000):
    adapters = [
        Qwen3TTSAdapter(sample_rate=sample_rate),
        XTTSV2Adapter(sample_rate=sample_rate),
    ]
    return {adapter.name: adapter for adapter in adapters}
