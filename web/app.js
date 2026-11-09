// Gemeinsame Helfer für alle Seiten: API-Aufrufe, Tab-Leiste, Formatierung.

const TABS = [
  ["index.html", "Dashboard"],
  ["bikes.html", "Fahrräder"],
  ["customers.html", "Kunden"],
  ["rent.html", "Verleih"],
  ["return.html", "Rückgabe"],
];

function renderTabs() {
  const nav = document.getElementById("tabs");
  if (!nav) return;
  const current = location.pathname.split("/").pop() || "index.html";
  nav.innerHTML =
    TABS.map(([href, label]) =>
      `<a href="${href}" class="tab${href === current ? " active" : ""}">${label}</a>`
    ).join("") +
    `<a href="reserve.html" class="tab external" target="_blank">Kundenseite ↗</a>`;
}

// api ruft einen Go-Endpoint auf und wirft bei Fehlern die Meldung des Servers.
async function api(method, path, body) {
  const opts = { method, headers: {} };
  if (body !== undefined) {
    opts.headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(path, opts);
  if (res.status === 204) return null;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || res.statusText);
  return data;
}

const pad = (n) => String(n).padStart(2, "0");

// fmtDateTime: 05.11.2026-14:00
function fmtDateTime(iso) {
  const d = new Date(iso);
  return `${pad(d.getDate())}.${pad(d.getMonth() + 1)}.${d.getFullYear()}-${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// fmtUntil zeigt ein exklusives Ende als letzte Minute an: Ende 08:00 -> 07:59.
function fmtUntil(iso) {
  return fmtDateTime(new Date(new Date(iso).getTime() - 60000));
}

// hourOptions: <option>s für die Stundenblöcke 00:00–00:59 … 23:00–23:59.
function hourOptions() {
  return [...Array(24).keys()].map((h) => `<option value="${h}">${pad(h)}:00 – ${pad(h)}:59</option>`).join("");
}

// isoDate: Date -> 2026-11-05 (lokale Zeit)
function isoDate(d) {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, (c) =>
    ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

function showMsg(el, text, ok = false) {
  el.textContent = text;
  el.className = "msg " + (ok ? "ok" : "err");
}

function bikeLabel(b) {
  return `${b.inventoryNo} – ${esc(b.brand)} ${esc(b.model)} (${esc(b.type)})`;
}

renderTabs();
