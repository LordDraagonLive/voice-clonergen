from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    database_url: str = "postgres://voice:voice@localhost:5432/voice_cloner?sslmode=disable"
    redis_url: str = "redis://localhost:6379/0"
    job_queue: str = "voice-cloner-jobs"
    storage_local_path: str = "./data/audio"
    default_model: str = "qwen3-tts"
    sample_rate: int = 24000
    enable_placeholder_tts: bool = False
    max_tts_chunk_chars: int = 350
    reference_min_seconds: float = 3.0
    reference_max_seconds: float = 15.0
    qwen_model_name: str = "Qwen/Qwen3-TTS-12Hz-0.6B-Base"
    qwen_device_map: str = "auto"
    qwen_dtype: str = "auto"
    qwen_attn_implementation: str = "sdpa"
    qwen_language: str = "English"
    enable_real_xtts: bool = False
    xtts_model_name: str = "tts_models/multilingual/multi-dataset/xtts_v2"
    xtts_device: str = "auto"
    xtts_language: str = "en"


settings = Settings()
