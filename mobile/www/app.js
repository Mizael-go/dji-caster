/* Relais de test : publie la caméra du téléphone en RTMP sur le serveur MediaMTX
   de DJI Caster, exactement comme le ferait une caméra DJI. */

const plugin = window.Capacitor?.Plugins?.RtmpRelay;

const el = {
  host: document.getElementById("host"),
  port: document.getElementById("port"),
  path: document.getElementById("path"),
  resolution: document.getElementById("resolution"),
  bitrate: document.getElementById("bitrate"),
  facing: document.getElementById("facing"),
  url: document.getElementById("url"),
  start: document.getElementById("start"),
  stop: document.getElementById("stop"),
  status: document.getElementById("status"),
  dot: document.getElementById("dot"),
  previewEmpty: document.getElementById("preview-empty"),
};

const RESOLUTIONS = {
  480: { width: 854, height: 480 },
  720: { width: 1280, height: 720 },
  1080: { width: 1920, height: 1080 },
};

const STORE = "dji-relay";

function buildUrl() {
  const host = el.host.value.trim() || "192.168.1.20";
  const port = el.port.value.trim() || "1935";
  const path = (el.path.value.trim() || "live/dji").replace(/^\/+/, "");
  return `rtmp://${host}:${port}/${path}`;
}

function refreshUrl() {
  el.url.textContent = buildUrl();
}

function setStatus(text, kind) {
  el.status.textContent = text;
  el.dot.className = "dot" + (kind ? " " + kind : "");
}

function setRunning(running) {
  el.start.disabled = running;
  el.stop.disabled = !running;
  el.previewEmpty.style.display = running ? "none" : "grid";
  for (const field of [el.host, el.port, el.path, el.resolution, el.bitrate, el.facing]) {
    field.disabled = running;
  }
}

function save() {
  localStorage.setItem(
    STORE,
    JSON.stringify({
      host: el.host.value,
      port: el.port.value,
      path: el.path.value,
      resolution: el.resolution.value,
      bitrate: el.bitrate.value,
      facing: el.facing.value,
    }),
  );
}

function restore() {
  let saved = {};
  try {
    saved = JSON.parse(localStorage.getItem(STORE) || "{}");
  } catch {
    saved = {};
  }
  for (const key of ["host", "port", "path", "resolution", "bitrate", "facing"]) {
    if (saved[key]) el[key].value = saved[key];
  }
}

async function start() {
  const host = el.host.value.trim();
  if (!host) {
    setStatus("Renseignez l'IP du PC qui fait tourner DJI Caster.", "error");
    el.host.focus();
    return;
  }
  if (!plugin) {
    setStatus("Le relais natif n'est disponible que dans l'application Android.", "error");
    return;
  }
  const size = RESOLUTIONS[el.resolution.value] || RESOLUTIONS[720];
  save();
  setStatus("Connexion à " + buildUrl() + "…");
  try {
    await plugin.start({
      url: buildUrl(),
      width: size.width,
      height: size.height,
      fps: 30,
      bitrate: Number(el.bitrate.value) * 1000,
      facing: el.facing.value,
    });
    setRunning(true);
  } catch (error) {
    setStatus(String(error?.message || error), "error");
    setRunning(false);
  }
}

async function stop() {
  if (!plugin) return;
  try {
    await plugin.stop();
  } finally {
    setRunning(false);
    setStatus("Diffusion arrêtée.");
  }
}

el.start.addEventListener("click", start);
el.stop.addEventListener("click", stop);
for (const field of [el.host, el.port, el.path]) {
  field.addEventListener("input", refreshUrl);
}
for (const field of [el.host, el.port, el.path, el.resolution, el.bitrate, el.facing]) {
  field.addEventListener("change", save);
}

plugin?.addListener("state", (event) => {
  const state = event.state;
  if (state === "connecting") setStatus("Négociation RTMP…");
  else if (state === "live") {
    setStatus("En direct — " + (event.message || buildUrl()), "live");
    setRunning(true);
  } else if (state === "idle") {
    setStatus("Diffusion arrêtée.");
    setRunning(false);
  } else if (state === "error") {
    setStatus(event.message || "Échec de la diffusion.", "error");
    setRunning(false);
  }
});

restore();
refreshUrl();
setRunning(false);
