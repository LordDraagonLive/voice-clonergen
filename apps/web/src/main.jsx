import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

const API = import.meta.env.VITE_API_URL || "http://localhost:8080";
const MODELS = ["qwen3-tts", "xtts-v2"];

function App() {
  const [profiles, setProfiles] = useState([]);
  const [samples, setSamples] = useState([]);
  const [jobs, setJobs] = useState([]);
  const [selectedProfile, setSelectedProfile] = useState("");
  const [message, setMessage] = useState("");

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
    refresh().catch(() => setMessage("API is not reachable yet."));
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
    if (!selectedProfile || !samples.some((sample) => sample.status === "pending" || sample.status === "processing")) return;
    const timer = setInterval(() => {
      refreshSamples().catch(() => {});
    }, 3000);
    return () => clearInterval(timer);
  }, [selectedProfile, samples]);

  useEffect(() => {
    if (!jobs.some((job) => job.status === "pending" || job.status === "running")) return;
    const timer = setInterval(() => {
      refresh().catch(() => {});
    }, 3000);
    return () => clearInterval(timer);
  }, [jobs]);

  const hasReadySample = samples.some(isUsableReadySample);

  return (
    <main>
      <header>
        <h1>Personal Voice Cloner</h1>
        <p>Consent-based, self-hosted voice profile and generation workspace.</p>
      </header>

      <section className="grid">
        <CreateProfile onDone={refresh} setMessage={setMessage} />
        <UploadSample profiles={profiles} samples={samples} selectedProfile={selectedProfile} setSelectedProfile={setSelectedProfile} onDone={() => { refresh(); refreshSamples(); }} setMessage={setMessage} />
        <GenerateSpeech profiles={profiles} selectedProfile={selectedProfile} setSelectedProfile={setSelectedProfile} hasReadySample={hasReadySample} onDone={refresh} setMessage={setMessage} />
        <Benchmark profiles={profiles} selectedProfile={selectedProfile} setSelectedProfile={setSelectedProfile} setMessage={setMessage} />
      </section>

      {message && <p className="notice">{message}</p>}
      <Profiles profiles={profiles} />
      <History jobs={jobs} />
    </main>
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

function CreateProfile({ onDone, setMessage }) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [modelDefault, setModelDefault] = useState("qwen3-tts");
  const [consentConfirmed, setConsentConfirmed] = useState(false);

  async function submit(event) {
    event.preventDefault();
    const res = await fetch(`${API}/api/voice-profiles`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name,
        description,
        modelDefault,
        consentConfirmed,
        consentText: "I confirm I own this voice or have explicit permission to clone it for personal use.",
      }),
    });
    setMessage(res.ok ? "Voice profile created." : (await res.json()).error);
    if (res.ok) {
      setName("");
      setDescription("");
      setConsentConfirmed(false);
      onDone();
    }
  }

  return (
    <form onSubmit={submit} className="panel">
      <h2>Create Voice Profile</h2>
      <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Profile name" />
      <textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Description" />
      <select value={modelDefault} onChange={(e) => setModelDefault(e.target.value)}>
        {MODELS.map((model) => <option key={model}>{model}</option>)}
      </select>
      <label className="checkbox">
        <input type="checkbox" checked={consentConfirmed} onChange={(e) => setConsentConfirmed(e.target.checked)} />
        I own this voice or have explicit permission to clone it.
      </label>
      <button>Create</button>
    </form>
  );
}

function UploadSample({ profiles, samples, selectedProfile, setSelectedProfile, onDone, setMessage }) {
  const [file, setFile] = useState(null);
  const [transcript, setTranscript] = useState("");

  async function submit(event) {
    event.preventDefault();
    if (!selectedProfile || !file) return setMessage("Choose a profile and audio file.");
    const form = new FormData();
    form.append("audio", file);
    form.append("transcript", transcript);
    const res = await fetch(`${API}/api/voice-profiles/${selectedProfile}/samples`, { method: "POST", body: form });
    setMessage(res.ok ? "Sample uploaded. Processing will finish automatically." : (await res.json()).error);
    if (res.ok) {
      setFile(null);
      onDone();
    }
  }

  return (
    <form onSubmit={submit} className="panel">
      <h2>Upload Samples</h2>
      <ProfileSelect profiles={profiles} value={selectedProfile} onChange={setSelectedProfile} />
      <input type="file" accept=".wav,.mp3,.m4a,.flac,audio/*" onChange={(e) => setFile(e.target.files?.[0])} />
      <textarea value={transcript} onChange={(e) => setTranscript(e.target.value)} placeholder="Optional transcript" />
      <button>Upload</button>
      <SampleList samples={samples} />
    </form>
  );
}

function SampleList({ samples }) {
  samples = samples || [];
  if (!samples.length) return <p className="muted">No samples uploaded for this profile yet.</p>;
  return (
    <div className="samples">
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

function GenerateSpeech({ profiles, selectedProfile, setSelectedProfile, hasReadySample, onDone, setMessage }) {
  const [text, setText] = useState("");
  const [modelName, setModelName] = useState("qwen3-tts");

  async function submit(event) {
    event.preventDefault();
    if (!hasReadySample) return setMessage("Wait for at least one uploaded sample to be ready before generating.");
    const res = await fetch(`${API}/api/generations`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ voiceProfileId: selectedProfile, modelName, text, format: "wav" }),
    });
    setMessage(res.ok ? "Generation queued. Status will update automatically." : (await res.json()).error);
    if (res.ok) onDone();
  }

  return (
    <form onSubmit={submit} className="panel">
      <h2>Generate Speech</h2>
      <ProfileSelect profiles={profiles} value={selectedProfile} onChange={setSelectedProfile} />
      <select value={modelName} onChange={(e) => setModelName(e.target.value)}>
        {MODELS.map((model) => <option key={model}>{model}</option>)}
      </select>
      <textarea value={text} onChange={(e) => setText(e.target.value)} placeholder="Text to speak" />
      <button disabled={!hasReadySample}>Generate</button>
      {!hasReadySample && <p className="muted">Upload a 3-15 second ready sample before generating.</p>}
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
  async function submit(event) {
    event.preventDefault();
    const res = await fetch(`${API}/api/benchmarks`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ voiceProfileId: selectedProfile, text, models: MODELS }),
    });
    setMessage(res.ok ? "Benchmark queued." : (await res.json()).error);
  }
  return (
    <form onSubmit={submit} className="panel">
      <h2>Benchmark Models</h2>
      <ProfileSelect profiles={profiles} value={selectedProfile} onChange={setSelectedProfile} />
      <textarea value={text} onChange={(e) => setText(e.target.value)} placeholder="Benchmark text" />
      <button>Compare</button>
    </form>
  );
}

function Profiles({ profiles }) {
  profiles = profiles || [];
  return (
    <section>
      <h2>Voice Profiles</h2>
      <div className="list">
        {profiles.map((profile) => (
          <article key={profile.id}>
            <strong>{profile.name}</strong>
            <span>{profile.modelDefault}</span>
            <p>{profile.description || "No description"}</p>
          </article>
        ))}
      </div>
    </section>
  );
}

function History({ jobs }) {
  jobs = jobs || [];
  return (
    <section>
      <h2>Generation History</h2>
      <div className="list">
        {jobs.map((job) => (
          <article key={job.id} className={`job ${job.status}`}>
            <strong>{job.modelName}</strong>
            <span>{job.status}</span>
            {job.progressMessage && <p className="progress">{job.progressMessage}</p>}
            <p>{job.inputText}</p>
            {job.errorMessage && <p className="error">{job.errorMessage}</p>}
            {job.outputFilePath && (
              <>
                <audio controls src={`${API}/api/generations/${job.id}/download`} />
                <a href={`${API}/api/generations/${job.id}/download`}>Download</a>
              </>
            )}
          </article>
        ))}
      </div>
    </section>
  );
}

createRoot(document.getElementById("root")).render(<App />);
