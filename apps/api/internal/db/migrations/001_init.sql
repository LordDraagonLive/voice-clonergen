CREATE TABLE IF NOT EXISTS voice_profiles (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  model_default TEXT NOT NULL DEFAULT 'qwen3-tts',
  consent_confirmed BOOLEAN NOT NULL DEFAULT false,
  consent_text TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS voice_samples (
  id UUID PRIMARY KEY,
  voice_profile_id UUID NOT NULL REFERENCES voice_profiles(id) ON DELETE CASCADE,
  original_file_path TEXT NOT NULL,
  cleaned_file_path TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'ready', 'failed')),
  error_message TEXT NOT NULL DEFAULT '',
  duration_seconds DOUBLE PRECISION NOT NULL DEFAULT 0,
  sample_rate INTEGER NOT NULL DEFAULT 0,
  quality_score DOUBLE PRECISION NOT NULL DEFAULT 0,
  transcript_optional TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE voice_samples ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE voice_samples ADD COLUMN IF NOT EXISTS error_message TEXT NOT NULL DEFAULT '';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'voice_samples_status_check'
  ) THEN
    ALTER TABLE voice_samples
      ADD CONSTRAINT voice_samples_status_check
      CHECK (status IN ('pending', 'processing', 'ready', 'failed'));
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS generation_jobs (
  id UUID PRIMARY KEY,
  voice_profile_id UUID NOT NULL REFERENCES voice_profiles(id) ON DELETE CASCADE,
  model_name TEXT NOT NULL,
  input_text TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
  error_message TEXT NOT NULL DEFAULT '',
  progress_message TEXT NOT NULL DEFAULT '',
  output_file_path TEXT NOT NULL DEFAULT '',
  latency_ms BIGINT NOT NULL DEFAULT 0,
  realtime_factor DOUBLE PRECISION NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

ALTER TABLE generation_jobs ADD COLUMN IF NOT EXISTS progress_message TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS benchmark_runs (
  id UUID PRIMARY KEY,
  voice_profile_id UUID NOT NULL REFERENCES voice_profiles(id) ON DELETE CASCADE,
  input_text TEXT NOT NULL,
  models_tested TEXT[] NOT NULL,
  results_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
