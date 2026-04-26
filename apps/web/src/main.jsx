import React, { useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

const API = import.meta.env.VITE_API_URL || "http://localhost:8081";
const MODELS = ["qwen3-tts", "xtts-v2"];
const MODEL_LABELS = {
  "qwen3-tts": "Qwen3 TTS",
  "xtts-v2": "XTTS v2",
};
const ACTIVE_SAMPLE_STATUSES = new Set(["pending", "processing"]);
const ACTIVE_JOB_STATUSES = new Set(["pending", "running"]);

function App() {
  const [profiles, setProfiles] = useState([]);
  const [samples, setSamples] = useState([]);
  const [jobs, setJobs] = useState([]);
  const [selectedProfile, setSelectedProfile] = useState("");
  const [message, setMessage] = useState("");
  const [messageTone, setMessageTone] = useState("info");
  const [loading, setLoading] = useState(true);

  const selectedProfileData = useMemo(
    () => profiles.find((profile) => profile.id === selectedProfile),
    [profiles, selectedProfile],
  );
  const readySamples = samples.filter(isUsableReadySample);
  const hasReadySample = readySamples.length > 0;
  const activeJobs = jobs.filter((job) => ACTIVE_JOB_STATUSES.has(job.status));
  const completedJobs = jobs.filter((job) => job.status === "completed");

  function notify(text, tone = "info") {
    setMessage(text || "");
    setMessageTone(tone);
  }

  async function refresh() {
    const [profilesRes, jobsRes] = await Promise.all([
      fetch(`${API}/api/voice-profiles`),
      fetch(`${API}/api/generations`),
    ]);
    if (profilesRes.ok) {
      const nextProfiles = (await profilesRes.json()) || [];
      setProfiles(nextProfiles);
      setSelectedProfile((current) => current || nextProfiles[0]?.id || "");
    }
    if (jobsRes.ok) setJobs((await jobsRes.json()) || []);
  }

  useEffect(() => {
    refresh()
      .catch(() => notify("API is not reachable yet. Start the stack, then refresh this page.", "warning"))
      .finally(() => setLoading(false));
  }, []);

  async function refreshSamples(profileId = selectedProfile) {
    if (!profileId) {
      setSamples([]);
      return;
    }
    const res = await fetch(`${API}/api/voice-profiles/${profileId}/samples`);
    if (res.ok) setSamples((await res.json()) || []);
  }

  useEffect(() => {
    refreshSamples().catch(() => setSamples([]));
  }, [selectedProfile]);

  useEffect(() => {
    if (!selectedProfile || !samples.some((sample) => ACTIVE_SAMPLE_STATUSES.has(sample.status))) return;
    const timer = setInterval(() => {
      refreshSamples().catch(() => {});
    }, 3000);
    return () => clearInterval(timer);
  }, [selectedProfile, samples]);

  useEffect(() => {
    if (!jobs.some((job) => ACTIVE_JOB_STATUSES.has(job.status))) return;
    const timer = setInterval(() => {
      refresh().catch(() => {});
    }, 3000);
    return () => clearInterval(timer);
  }, [jobs]);

  return (
    <main>
      <header className="hero">
        <div>
          <p className="eyebrow">Consent-first voice workspace</p>
          <h1>Personal Voice Cloner</h1>
          <p className="hero-copy">
            Create a permitted voice profile, prepare clean reference audio, generate speech, and compare models from one self-hosted control room.
          </p>
        </div>
        <div className="hero-status" aria-label="Workspace status">
          <StatusMetric label="Profiles" value={profiles.length} />
          <StatusMetric label="Ready samples" value={readySamples.length} />
          <StatusMetric label="Active jobs" value={activeJobs.length} />
        </div>
      </header>

      {message && <p className={`notice ${messageTone}`}>{message}</p>}

      <section className="workspace" aria-label="Voice cloning workflow">
        <aside className="rail">
          <h2>Workflow</h2>
          <Step number="1" title="Create" text="Name the voice and confirm permission." active={!profiles.length} complete={profiles.length > 0} />
          <Step number="2" title="Upload" text="Add a clear 3-20 second sample." active={Boolean(selectedProfile && !hasReadySample)} complete={hasReadySample} />
          <Step number="3" title="Generate" text="Queue speech once a sample is ready." active={hasReadySample} complete={completedJobs.length > 0} />
          <Step number="4" title="Compare" text="Benchmark both adapters with one prompt." active={completedJobs.length > 0} />
        </aside>

        <div className="cards">
          <CreateProfile onDone={refresh} setMessage={notify} />
          <UploadSample
            profiles={profiles}
            samples={samples}
            selectedProfile={selectedProfile}
            selectedProfileData={selectedProfileData}
            setSelectedProfile={setSelectedProfile}
            onDone={() => {
              refresh();
              refreshSamples();
            }}
            setMessage={notify}
          />
          <GenerateSpeech
            profiles={profiles}
            selectedProfile={selectedProfile}
            selectedProfileData={selectedProfileData}
            setSelectedProfile={setSelectedProfile}
            hasReadySample={hasReadySample}
            onDone={refresh}
            setMessage={notify}
          />
          <Benchmark profiles={profiles} selectedProfile={selectedProfile} setSelectedProfile={setSelectedProfile} setMessage={notify} />
        </div>
      </section>

      <section className="results-grid">
        <Profiles profiles={profiles} loading={loading} />
        <History jobs={jobs} loading={loading} />
      </section>
    </main>
  );
}

function StatusMetric({ label, value }) {
  return (
    <div>
      <strong>{value}</strong>
      <span>{label}</span>
    </div>
  );
}

function Step({ number, title, text, active, complete }) {
  return (
    <div className={`step ${active ? "active" : ""} ${complete ? "complete" : ""}`}>
      <span>{complete ? "OK" : number}</span>
      <div>
        <strong>{title}</strong>
        <p>{text}</p>
      </div>
    </div>
  );
}

function Field({ label, hint, children }) {
  return (
    <label className="field">
      <span>{label}</span>
      {children}
      {hint && <small>{hint}</small>}
    </label>
  );
}

function ProfileSelect({ profiles, value, onChange }) {
  profiles = profiles || [];
  return (
    <select value={value} onChange={(event) => onChange(event.target.value)}>
      <option value="">Select voice profile</option>
      {profiles.map((profile) => (
        <option key={profile.id} value={profile.id}>{profile.name}</option>
      ))}
    </select>
  );
}

function ModelSelect({ value, onChange }) {
  return (
    <select value={value} onChange={(event) => onChange(event.target.value)}>
      {MODELS.map((model) => <option key={model} value={model}>{MODEL_LABELS[model]}</option>)}
    </select>
  );
}

function CreateProfile({ onDone, setMessage }) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [modelDefault, setModelDefault] = useState("qwen3-tts");
  const [consentConfirmed, setConsentConfirmed] = useState(false);
  const canSubmit = name.trim().length > 0 && consentConfirmed;

  async function submit(event) {
    event.preventDefault();
    if (!canSubmit) return setMessage("Add a profile name and confirm consent before creating.", "warning");
    const res = await fetch(`${API}/api/voice-profiles`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: name.trim(),
        description: description.trim(),
        modelDefault,
        consentConfirmed,
        consentText: "I confirm I own this voice or have explicit permission to clone it for personal use.",
      }),
    });
    setMessage(res.ok ? "Voice profile created." : await errorMessage(res), res.ok ? "success" : "error");
    if (res.ok) {
      setName("");
      setDescription("");
      setConsentConfirmed(false);
      onDone();
    }
  }

  return (
    <form onSubmit={submit} className="panel featured">
      <PanelTitle title="Create Voice Profile" detail="Start with ownership, permission, and a default model." />
      <Field label="Profile name">
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="My narration voice" />
      </Field>
      <Field label="Description" hint="A short note helps when you manage multiple profiles.">
        <textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Warm, calm delivery for personal projects" />
      </Field>
      <Field label="Default model">
        <ModelSelect value={modelDefault} onChange={setModelDefault} />
      </Field>
      <label className="checkbox">
        <input type="checkbox" checked={consentConfirmed} onChange={(e) => setConsentConfirmed(e.target.checked)} />
        <span>I own this voice or have explicit permission to clone it.</span>
      </label>
      <button disabled={!canSubmit}>Create profile</button>
    </form>
  );
}

function PanelTitle({ title, detail, action }) {
  return (
    <div className="panel-title">
      <div>
        <h2>{title}</h2>
        {detail && <p>{detail}</p>}
      </div>
      {action}
    </div>
  );
}

function UploadSample({ profiles, samples, selectedProfile, selectedProfileData, setSelectedProfile, onDone, setMessage }) {
  const [file, setFile] = useState(null);
  const [fileInputKey, setFileInputKey] = useState(0);
  const [transcript, setTranscript] = useState("");
  const canSubmit = Boolean(selectedProfile && file);

  async function submit(event) {
    event.preventDefault();
    if (!canSubmit) return setMessage("Choose a profile and audio file.", "warning");
    const form = new FormData();
    form.append("audio", file);
    form.append("transcript", transcript.trim());
    const res = await fetch(`${API}/api/voice-profiles/${selectedProfile}/samples`, { method: "POST", body: form });
    setMessage(res.ok ? "Sample uploaded. Processing will finish automatically." : await errorMessage(res), res.ok ? "success" : "error");
    if (res.ok) {
      setFile(null);
      setFileInputKey((key) => key + 1);
      setTranscript("");
      onDone();
    }
  }

  return (
    <form onSubmit={submit} className="panel">
      <PanelTitle
        title="Upload Samples"
        detail={selectedProfileData ? `Adding reference audio for ${selectedProfileData.name}.` : "Choose a profile before uploading audio."}
      />
      <Field label="Voice profile">
        <ProfileSelect profiles={profiles} value={selectedProfile} onChange={setSelectedProfile} />
      </Field>
      <Field label="Reference audio" hint={file ? `${file.name} (${formatBytes(file.size)})` : "WAV, MP3, M4A, or FLAC. Aim for clean speech."}>
        <input key={fileInputKey} type="file" accept=".wav,.mp3,.m4a,.flac,audio/*" onChange={(e) => setFile(e.target.files?.[0] || null)} />
      </Field>
      <Field label="Transcript" hint="Optional, but useful for later quality checks.">
        <textarea value={transcript} onChange={(e) => setTranscript(e.target.value)} placeholder="Words spoken in the sample" />
      </Field>
      <button disabled={!canSubmit}>Upload sample</button>
      <SampleList samples={samples} />
    </form>
  );
}

function SampleList({ samples }) {
  samples = samples || [];
  if (!samples.length) return <EmptyState text="No samples uploaded for this profile yet." />;
  return (
    <div className="samples" aria-label="Uploaded samples">
      {samples.map((sample) => (
        <div key={sample.id} className={`sample ${sample.status} ${sample.status === "ready" && !isUsableReadySample(sample) ? "not-usable" : ""}`}>
          <strong>{sampleLabel(sample)}</strong>
          <span>{sample.durationSeconds ? `${sample.durationSeconds.toFixed(1)}s` : "waiting"}</span>
          {sample.errorMessage && <p>{sample.errorMessage}</p>}
        </div>
      ))}
    </div>
  );
}

function GenerateSpeech({ profiles, selectedProfile, selectedProfileData, setSelectedProfile, hasReadySample, onDone, setMessage }) {
  const [text, setText] = useState("");
  const [modelName, setModelName] = useState(selectedProfileData?.modelDefault || "qwen3-tts");
  const charCount = text.trim().length;
  const canSubmit = Boolean(selectedProfile && hasReadySample && charCount);

  useEffect(() => {
    if (selectedProfileData?.modelDefault) setModelName(selectedProfileData.modelDefault);
  }, [selectedProfileData]);

  async function submit(event) {
    event.preventDefault();
    if (!hasReadySample) return setMessage("Wait for at least one uploaded sample to be ready before generating.", "warning");
    if (!charCount) return setMessage("Add text before generating speech.", "warning");
    const res = await fetch(`${API}/api/generations`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ voiceProfileId: selectedProfile, modelName, text: text.trim(), format: "wav" }),
    });
    setMessage(res.ok ? "Generation queued. Status will update automatically." : await errorMessage(res), res.ok ? "success" : "error");
    if (res.ok) {
      setText("");
      onDone();
    }
  }

  return (
    <form onSubmit={submit} className="panel primary-action">
      <PanelTitle title="Generate Speech" detail="Use a ready sample to create downloadable WAV output." action={<span className={`pill ${hasReadySample ? "ready" : "waiting"}`}>{hasReadySample ? "Ready" : "Needs sample"}</span>} />
      <Field label="Voice profile">
        <ProfileSelect profiles={profiles} value={selectedProfile} onChange={setSelectedProfile} />
      </Field>
      <Field label="Model">
        <ModelSelect value={modelName} onChange={setModelName} />
      </Field>
      <Field label="Text to speak" hint={`${charCount} characters`}>
        <textarea className="large-text" value={text} onChange={(e) => setText(e.target.value)} placeholder="Paste or type the line you want this voice to say." />
      </Field>
      <button disabled={!canSubmit}>Generate speech</button>
      {!hasReadySample && <p className="muted">Upload a 3-20 second ready sample before generating.</p>}
    </form>
  );
}

function isUsableReadySample(sample) {
  return sample.status === "ready" && sample.durationSeconds >= 3 && sample.durationSeconds <= 20;
}

function sampleLabel(sample) {
  if (sample.status === "ready" && !isUsableReadySample(sample)) return "reupload needed";
  return sample.status;
}

function Benchmark({ profiles, selectedProfile, setSelectedProfile, setMessage }) {
  const [text, setText] = useState("");
  const canSubmit = Boolean(selectedProfile && text.trim());

  async function submit(event) {
    event.preventDefault();
    if (!canSubmit) return setMessage("Choose a profile and add benchmark text.", "warning");
    const res = await fetch(`${API}/api/benchmarks`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ voiceProfileId: selectedProfile, text: text.trim(), models: MODELS }),
    });
    setMessage(res.ok ? "Benchmark queued." : await errorMessage(res), res.ok ? "success" : "error");
    if (res.ok) setText("");
  }

  return (
    <form onSubmit={submit} className="panel">
      <PanelTitle title="Benchmark Models" detail="Compare model output with the same profile and prompt." />
      <Field label="Voice profile">
        <ProfileSelect profiles={profiles} value={selectedProfile} onChange={setSelectedProfile} />
      </Field>
      <Field label="Benchmark text" hint={`${text.trim().length} characters`}>
        <textarea value={text} onChange={(e) => setText(e.target.value)} placeholder="Use a short, expressive phrase for comparison." />
      </Field>
      <button disabled={!canSubmit}>Compare models</button>
    </form>
  );
}

function Profiles({ profiles, loading }) {
  profiles = profiles || [];
  return (
    <section className="section-block">
      <div className="section-title">
        <h2>Voice Profiles</h2>
        <span>{profiles.length} total</span>
      </div>
      <div className="list">
        {profiles.map((profile) => (
          <article key={profile.id}>
            <div className="article-head">
              <strong>{profile.name}</strong>
              <span>{MODEL_LABELS[profile.modelDefault] || profile.modelDefault}</span>
            </div>
            <p>{profile.description || "No description added yet."}</p>
          </article>
        ))}
        {!profiles.length && <EmptyState text={loading ? "Loading profiles..." : "Create your first profile to begin."} />}
      </div>
    </section>
  );
}

function History({ jobs, loading }) {
  jobs = jobs || [];
  return (
    <section className="section-block">
      <div className="section-title">
        <h2>Generation History</h2>
        <span>{jobs.length} jobs</span>
      </div>
      <div className="list">
        {jobs.map((job) => (
          <article key={job.id} className={`job ${job.status}`}>
            <div className="article-head">
              <strong>{MODEL_LABELS[job.modelName] || job.modelName}</strong>
              <span className="status-chip">{job.status}</span>
            </div>
            {job.progressMessage && <p className="progress">{job.progressMessage}</p>}
            <p>{job.inputText}</p>
            {job.errorMessage && <p className="error">{job.errorMessage}</p>}
            {job.outputFilePath && (
              <div className="audio-result">
                <audio controls src={`${API}/api/generations/${job.id}/download`} />
                <a href={`${API}/api/generations/${job.id}/download`}>Download WAV</a>
              </div>
            )}
          </article>
        ))}
        {!jobs.length && <EmptyState text={loading ? "Loading generations..." : "Generated audio will appear here."} />}
      </div>
    </section>
  );
}

function EmptyState({ text }) {
  return <p className="empty">{text}</p>;
}

function formatBytes(bytes) {
  if (!Number.isFinite(bytes)) return "";
  if (bytes < 1024 * 1024) return `${Math.max(1, Math.round(bytes / 1024))} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

async function errorMessage(res) {
  try {
    const body = await res.json();
    return body?.error || "Something went wrong.";
  } catch {
    return "Something went wrong.";
  }
}

createRoot(document.getElementById("root")).render(<App />);
