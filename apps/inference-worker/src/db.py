import psycopg

from .config import settings


def connect():
    return psycopg.connect(settings.database_url)


def mark_sample_processing(sample_id: str) -> None:
    with connect() as conn:
        conn.execute(
            """
            UPDATE voice_samples
            SET status = 'processing', error_message = ''
            WHERE id = %s
            """,
            (sample_id,),
        )


def mark_sample_ready(sample_id: str, result: dict) -> None:
    with connect() as conn:
        conn.execute(
            """
            UPDATE voice_samples
            SET cleaned_file_path = %s,
                duration_seconds = %s,
                sample_rate = %s,
                quality_score = %s,
                status = 'ready',
                error_message = ''
            WHERE id = %s
            """,
            (
                result["cleaned_file_path"],
                result["duration_seconds"],
                result["sample_rate"],
                result["quality_score"],
                sample_id,
            ),
        )


def mark_sample_failed(sample_id: str, error_message: str) -> None:
    with connect() as conn:
        conn.execute(
            """
            UPDATE voice_samples
            SET status = 'failed', error_message = %s
            WHERE id = %s
            """,
            (error_message[:1000], sample_id),
        )


def mark_generation_running(job_id: str, progress_message: str = "Starting generation") -> None:
    with connect() as conn:
        conn.execute(
            """
            UPDATE generation_jobs
            SET status = 'running', error_message = '', progress_message = %s
            WHERE id = %s
            """,
            (progress_message, job_id),
        )


def update_generation_progress(job_id: str, progress_message: str) -> None:
    with connect() as conn:
        conn.execute(
            """
            UPDATE generation_jobs
            SET progress_message = %s
            WHERE id = %s
              AND status IN ('pending', 'running')
            """,
            (progress_message, job_id),
        )


def load_generation_context(job_id: str) -> dict:
    with connect() as conn:
        row = conn.execute(
            """
            SELECT
                gj.model_name,
                gj.input_text,
                vs.cleaned_file_path,
                vs.transcript_optional
            FROM generation_jobs gj
            LEFT JOIN LATERAL (
                SELECT cleaned_file_path, transcript_optional
                FROM voice_samples
                WHERE voice_profile_id = gj.voice_profile_id
                  AND status = 'ready'
                  AND cleaned_file_path <> ''
                  AND duration_seconds BETWEEN 3 AND 20
                ORDER BY quality_score DESC, created_at DESC
                LIMIT 1
            ) vs ON true
            WHERE gj.id = %s
            """,
            (job_id,),
        ).fetchone()
    if row is None:
        raise ValueError(f"generation job not found: {job_id}")
    model_name, text, reference_audio_path, reference_text = row
    if not reference_audio_path:
        raise ValueError("no ready voice sample found for this generation job")
    return {
        "modelName": model_name,
        "text": text,
        "referenceAudioPath": reference_audio_path,
        "referenceText": reference_text,
    }


def mark_generation_completed(job_id: str, result: dict) -> None:
    with connect() as conn:
        conn.execute(
            """
            UPDATE generation_jobs
            SET status = 'completed',
                error_message = '',
                progress_message = 'Generation completed',
                output_file_path = %s,
                latency_ms = %s,
                realtime_factor = %s,
                completed_at = now()
            WHERE id = %s
            """,
            (
                result["output_path"],
                result["latency_ms"],
                result["realtime_factor"],
                job_id,
            ),
        )


def mark_generation_failed(job_id: str, error_message: str) -> None:
    with connect() as conn:
        conn.execute(
            """
            UPDATE generation_jobs
            SET status = 'failed',
                error_message = %s,
                progress_message = 'Generation failed',
                completed_at = now()
            WHERE id = %s
            """,
            (error_message[:1000], job_id),
        )
