"use strict";

// --- Go ---------------------------------------------------------------------

const call = (method, ...args) => window.mygo.call("Library." + method, ...args);
const errText = (e) => String(e?.message ?? e).replace(/^Error: /, "");

// --- Icons ------------------------------------------------------------------

const svg = (d, extra = "") =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" ${extra}>${d}</svg>`;
const icons = {
  all: svg('<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>'),
  star: svg('<path d="m12 3 2.8 5.7 6.2.9-4.5 4.4 1 6.2L12 17.3 6.5 20.2l1-6.2L3 9.6l6.2-.9z"/>'),
  starOn: svg('<path fill="currentColor" d="m12 3 2.8 5.7 6.2.9-4.5 4.4 1 6.2L12 17.3 6.5 20.2l1-6.2L3 9.6l6.2-.9z"/>'),
  inbox: svg('<path d="M22 12h-6l-2 3h-4l-2-3H2"/><path d="M5.5 5h13L22 12v6a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2v-6z"/>'),
  folder: svg('<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>'),
  plus: svg('<path d="M12 5v14M5 12h14"/>'),
  close: svg('<path d="M18 6 6 18M6 6l12 12"/>'),
  left: svg('<path d="m15 18-6-6 6-6"/>'),
  right: svg('<path d="m9 18 6-6-6-6"/>'),
  play: svg('<path fill="currentColor" d="M7 4v16l13-8z"/>'),
  image: svg('<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-5-5L5 21"/>'),
  check: svg('<path d="M20 6 9 17l-5-5"/>', 'stroke-width="3"'),
};

// --- State ------------------------------------------------------------------

const prefs = load("xdll.prefs", { sort: "added-desc", zoom: 220 });
const state = {
  folders: [],
  bookmarks: [],
  view: { type: "all" }, // all | fav | unfiled | folder (with id)
  search: "",
  kind: "",
  sort: prefs.sort,
  selected: new Set(),
  pending: [], // { key, url, error }
  viewer: null, // { ids, pos, media }
};

function load(key, fallback) {
  try {
    return { ...fallback, ...JSON.parse(localStorage.getItem(key) || "{}") };
  } catch {
    return fallback;
  }
}
function savePrefs() {
  try {
    localStorage.setItem("xdll.prefs", JSON.stringify({ sort: state.sort, zoom: +$("#zoom").value }));
  } catch {}
}

function apply(snapshot) {
  state.folders = snapshot.folders || [];
  state.bookmarks = snapshot.bookmarks || [];
  const ids = new Set(state.bookmarks.map((b) => b.id));
  for (const id of state.selected) if (!ids.has(id)) state.selected.delete(id);
  if (state.view.type === "folder" && !state.folders.some((f) => f.id === state.view.id)) state.view = { type: "all" };
  render();
}

// --- Helpers ----------------------------------------------------------------

const $ = (s, root = document) => root.querySelector(s);
const esc = (s) => String(s ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]);
const mediaURL = (file) => "/media/" + file.split("/").map(encodeURIComponent).join("/");
const folderName = (id) => state.folders.find((f) => f.id === id)?.name;
const dateFmt = new Intl.DateTimeFormat("fr-FR", { dateStyle: "medium", timeStyle: "short" });
const fmtDate = (s) => (s && !s.startsWith("0001") ? dateFmt.format(new Date(s)) : "-");
const totalSize = (b) => b.media.reduce((n, m) => n + (m.size || 0), 0);
function fmtSize(n) {
  if (n < 1024 * 1024) return Math.max(1, Math.round(n / 1024)) + " Ko";
  if (n < 1024 * 1024 * 1024) return (n / 1024 / 1024).toFixed(1) + " Mo";
  return (n / 1024 / 1024 / 1024).toFixed(2) + " Go";
}

function toast(text, error = false) {
  const el = document.createElement("div");
  el.className = "toast" + (error ? " error" : "");
  el.textContent = text;
  $("#toasts").append(el);
  setTimeout(() => el.remove(), error ? 6000 : 3000);
}

async function run(promise, ok) {
  try {
    const r = await promise;
    if (ok) ok(r);
    return r;
  } catch (e) {
    toast(errText(e), true);
  }
}

// ask resolves to the input (or true), or null when canceled.
function ask({ title, text = "", value = null, ok = "OK", danger = false }) {
  const dlg = $("#dialog");
  $("#dialog-title").textContent = title;
  $("#dialog-text").textContent = text;
  $("#dialog-text").hidden = !text;
  const input = $("#dialog-input");
  input.hidden = value === null;
  input.value = value ?? "";
  const okBtn = $("#dialog-ok");
  okBtn.textContent = ok;
  okBtn.className = danger ? "danger" : "primary";
  dlg.hidden = false;
  (value === null ? okBtn : input).focus();
  input.select();
  return new Promise((resolve) => {
    const done = (r) => {
      dlg.hidden = true;
      $("#dialog-form").onsubmit = $("#dialog-cancel").onclick = dlg.onkeydown = dlg.onmousedown = null;
      resolve(r);
    };
    $("#dialog-form").onsubmit = (e) => {
      e.preventDefault();
      done(value === null ? true : input.value);
    };
    $("#dialog-cancel").onclick = () => done(null);
    // Block bodies: an on* handler returning false cancels the event.
    dlg.onkeydown = (e) => {
      if (e.key === "Escape") {
        e.stopPropagation();
        done(null);
      }
    };
    dlg.onmousedown = (e) => {
      if (e.target === dlg) done(null);
    };
  });
}

function openMenu(x, y, items) {
  const menu = $("#menu");
  menu.innerHTML = "";
  for (const item of items) {
    if (item === "-") {
      menu.append(document.createElement("hr"));
      continue;
    }
    const b = document.createElement("button");
    b.textContent = item.label;
    if (item.danger) b.className = "danger";
    b.onclick = () => {
      menu.hidden = true;
      item.run();
    };
    menu.append(b);
  }
  menu.hidden = false;
  const r = menu.getBoundingClientRect();
  menu.style.left = Math.min(x, innerWidth - r.width - 8) + "px";
  menu.style.top = Math.min(y, innerHeight - r.height - 8) + "px";
}
addEventListener("mousedown", (e) => !e.target.closest("#menu") && ($("#menu").hidden = true), true);
addEventListener("blur", () => ($("#menu").hidden = true));

// --- Filtering and sorting ---------------------------------------------------

function inView(b, view = state.view) {
  switch (view.type) {
    case "fav":
      return b.favorite;
    case "unfiled":
      return !b.folderId;
    case "folder":
      return b.folderId === view.id;
    default:
      return true;
  }
}

function visible() {
  const q = state.search.trim().toLowerCase();
  const terms = q ? q.split(/\s+/) : [];
  const list = state.bookmarks.filter((b) => {
    if (!inView(b)) return false;
    if (state.kind && !b.media.some((m) => m.kind === state.kind)) return false;
    if (!terms.length) return true;
    const hay = `${b.text} @${b.author} ${b.authorName} ${b.note} ${folderName(b.folderId) ?? ""}`.toLowerCase();
    return terms.every((t) => hay.includes(t));
  });
  const by = {
    "added-desc": (a, b) => b.addedAt.localeCompare(a.addedAt),
    "added-asc": (a, b) => a.addedAt.localeCompare(b.addedAt),
    "posted-desc": (a, b) => b.postedAt.localeCompare(a.postedAt),
    "posted-asc": (a, b) => a.postedAt.localeCompare(b.postedAt),
    author: (a, b) => a.author.localeCompare(b.author, "fr", { sensitivity: "base" }) || b.addedAt.localeCompare(a.addedAt),
    size: (a, b) => totalSize(b) - totalSize(a),
  }[state.sort];
  return list.sort(by);
}

function viewTitle() {
  return { all: "Tout", fav: "Favoris", unfiled: "Non classés" }[state.view.type] ?? folderName(state.view.id) ?? "Dossier";
}

// --- Rendering ----------------------------------------------------------------

function render() {
  renderSidebar();
  renderGrid();
  renderSelection();
  if (state.viewer) renderViewer();
}

function navLink({ icon, label, count, on, onClick, folderId, onContext }) {
  const a = document.createElement("a");
  a.className = on ? "on" : "";
  a.innerHTML = `${icon}<span class="label">${esc(label)}</span><span class="count">${count}</span>`;
  a.onclick = onClick;
  if (onContext) a.oncontextmenu = (e) => (e.preventDefault(), onContext(e));
  if (folderId !== undefined) {
    a.ondragover = (e) => {
      if (!e.dataTransfer.types.includes("text/x-xdll")) return;
      e.preventDefault();
      e.dataTransfer.dropEffect = "move";
      a.classList.add("drop");
    };
    a.ondragleave = () => a.classList.remove("drop");
    a.ondrop = (e) => {
      e.preventDefault();
      a.classList.remove("drop");
      const ids = JSON.parse(e.dataTransfer.getData("text/x-xdll") || "[]");
      if (ids.length)
        run(call("Move", ids, folderId), (s) => {
          apply(s);
          toast(`Déplacé dans ${folderName(folderId) ?? "Non classés"}`);
        });
    };
  }
  return a;
}

function setView(view) {
  state.view = view;
  state.selected.clear();
  render();
  $("#grid").scrollTop = 0;
}

function renderSidebar() {
  const v = state.view;
  const n = (view) => state.bookmarks.filter((b) => inView(b, view)).length;
  const views = $("#views");
  views.replaceChildren(
    navLink({ icon: icons.all, label: "Tout", count: state.bookmarks.length, on: v.type === "all", onClick: () => setView({ type: "all" }) }),
    navLink({ icon: icons.star, label: "Favoris", count: n({ type: "fav" }), on: v.type === "fav", onClick: () => setView({ type: "fav" }) }),
    navLink({ icon: icons.inbox, label: "Non classés", count: n({ type: "unfiled" }), on: v.type === "unfiled", onClick: () => setView({ type: "unfiled" }), folderId: "" }),
  );
  const folders = $("#folders");
  folders.replaceChildren(
    ...state.folders.map((f) =>
      navLink({
        icon: icons.folder,
        label: f.name,
        count: n({ type: "folder", id: f.id }),
        on: v.type === "folder" && v.id === f.id,
        onClick: () => setView({ type: "folder", id: f.id }),
        folderId: f.id,
        onContext: (e) => folderMenu(f, e.clientX, e.clientY),
      }),
    ),
  );
  if (!state.folders.length) {
    folders.innerHTML = `<div class="hint">Aucun dossier</div>`;
  }
  const size = state.bookmarks.reduce((s, b) => s + totalSize(b), 0);
  const media = state.bookmarks.reduce((s, b) => s + b.media.length, 0);
  $("#stats").textContent = `${state.bookmarks.length} posts · ${media} médias · ${fmtSize(size)}`;
  $("#title").textContent = viewTitle();
}

function folderMenu(f, x, y) {
  openMenu(x, y, [
    { label: "Renommer", run: () => renameFolder(f) },
    { label: "Exporter", run: () => exportIds(state.bookmarks.filter((b) => b.folderId === f.id).map((b) => b.id)) },
    "-",
    { label: "Supprimer le dossier", danger: true, run: () => deleteFolder(f) },
  ]);
}

async function newFolder() {
  const name = await ask({ title: "Nouveau dossier", value: "", ok: "Créer" });
  if (name === null) return;
  const f = await run(call("CreateFolder", name));
  if (!f) return;
  if (state.selected.size) await run(call("Move", [...state.selected], f.id), apply);
  else apply(await call("List"));
  setView({ type: "folder", id: f.id });
}

async function renameFolder(f) {
  const name = await ask({ title: "Renommer le dossier", value: f.name, ok: "Renommer" });
  if (name !== null) run(call("RenameFolder", f.id, name), apply);
}

async function deleteFolder(f) {
  const count = state.bookmarks.filter((b) => b.folderId === f.id).length;
  const ok = await ask({
    title: `Supprimer « ${f.name} » ?`,
    text: count ? `${count} post(s) iront dans Non classés.` : "",
    ok: "Supprimer",
    danger: true,
  });
  if (ok) run(call("DeleteFolder", f.id), apply);
}

function thumbHTML(m) {
  if (!m) return "";
  if (m.kind === "image") return `<img loading="lazy" src="${mediaURL(m.file)}" alt="" />`;
  if (m.kind === "gif") return `<video src="${mediaURL(m.file)}" autoplay loop muted playsinline></video>`;
  return m.thumb ? `<img loading="lazy" src="${mediaURL(m.thumb)}" alt="" />` : `<video src="${mediaURL(m.file)}#t=0.1" muted preload="metadata"></video>`;
}

function badgesHTML(b) {
  const out = [];
  const videos = b.media.filter((m) => m.kind === "video").length;
  const gifs = b.media.filter((m) => m.kind === "gif").length;
  const images = b.media.length - videos - gifs;
  if (videos) out.push(`<span class="badge">${icons.play}${videos > 1 ? videos : "Vidéo"}</span>`);
  if (gifs) out.push(`<span class="badge">GIF${gifs > 1 ? " ×" + gifs : ""}</span>`);
  if (images > 1 || (images && out.length)) out.push(`<span class="badge">${icons.image}${images}</span>`);
  return out.join("");
}

let lastClicked = null;

function renderGrid() {
  const grid = $("#grid");
  const list = visible();
  const cards = [];
  for (const p of state.pending) {
    const el = document.createElement("div");
    el.className = "card pending" + (p.error ? " failed" : "");
    el.innerHTML = `<div class="thumb">${p.error ? esc(p.error) : ""}</div>
      <div class="info"><div class="author"><b>${p.error ? "Échec" : "Téléchargement"}</b></div><div class="snippet">${esc(p.url)}</div></div>`;
    if (p.error) {
      el.style.cursor = "pointer";
      el.onclick = () => {
        state.pending = state.pending.filter((x) => x !== p);
        renderGrid();
      };
    }
    cards.push(el);
  }
  list.forEach((b, i) => {
    const el = document.createElement("div");
    el.className = "card" + (state.selected.has(b.id) ? " selected" : "");
    el.draggable = true;
    const folder = state.view.type !== "folder" && b.folderId ? folderName(b.folderId) : "";
    el.innerHTML = `<div class="thumb">${thumbHTML(b.media[0])}<div class="badges">${badgesHTML(b)}</div>
        <span class="check">${icons.check}</span>
        <button class="star${b.favorite ? " on" : ""}" title="Favori">${b.favorite ? icons.starOn : icons.star}</button></div>
      <div class="info">
        <div class="author"><b>${esc(b.authorName || b.author)}</b><span>@${esc(b.author)}</span></div>
        <div class="snippet">${esc(b.note || b.text) || "&nbsp;"}</div>
        ${folder ? `<div class="folder-tag">${esc(folder)}</div>` : ""}
      </div>`;
    el.querySelector(".star").onclick = (e) => {
      e.stopPropagation();
      run(call("SetFavorite", b.id, !b.favorite), apply);
    };
    el.querySelector(".check").onclick = (e) => {
      e.stopPropagation();
      toggleSelect(b.id);
      lastClicked = i;
    };
    el.onclick = (e) => {
      if (e.shiftKey && lastClicked !== null) {
        const [a, z] = [Math.min(lastClicked, i), Math.max(lastClicked, i)];
        for (let k = a; k <= z; k++) state.selected.add(list[k].id);
        render();
      } else if (e.ctrlKey || e.metaKey || state.selected.size) {
        toggleSelect(b.id);
        lastClicked = i;
      } else {
        lastClicked = i;
        openViewer(list.map((x) => x.id), i);
      }
    };
    el.oncontextmenu = (e) => {
      e.preventDefault();
      cardMenu(b, e.clientX, e.clientY);
    };
    el.ondragstart = (e) => {
      const ids = state.selected.has(b.id) ? [...state.selected] : [b.id];
      e.dataTransfer.setData("text/x-xdll", JSON.stringify(ids));
      e.dataTransfer.effectAllowed = "move";
      el.classList.add("dragging");
    };
    el.ondragend = () => el.classList.remove("dragging");
    cards.push(el);
  });
  grid.replaceChildren(...cards);

  const empty = $("#empty");
  empty.hidden = cards.length > 0;
  grid.hidden = cards.length === 0;
  if (!cards.length) {
    empty.innerHTML = state.bookmarks.length
      ? `<b>Aucun résultat</b>`
      : `<b>Aucun post</b>`;
  }
}

function cardMenu(b, x, y) {
  const ids = state.selected.has(b.id) ? [...state.selected] : [b.id];
  const many = ids.length > 1;
  openMenu(x, y, [
    ...(many ? [] : [{ label: "Ouvrir", run: () => openViewer([b.id], 0) }, { label: "Ouvrir sur X", run: () => run(call("OpenPost", b.id)) }, { label: "Explorateur", run: () => call("Reveal", b.id) }]),
    { label: many ? `Exporter (${ids.length})` : "Exporter", run: () => exportIds(ids) },
    { label: "Copier le lien", run: () => copyLinks(ids) },
    "-",
    { label: "Non classés", run: () => run(call("Move", ids, ""), apply) },
    ...state.folders.map((f) => ({ label: f.name, run: () => run(call("Move", ids, f.id), apply) })),
    "-",
    { label: many ? `Supprimer (${ids.length})` : "Supprimer", danger: true, run: () => deleteIds(ids) },
  ]);
}

function toggleSelect(id) {
  state.selected.has(id) ? state.selected.delete(id) : state.selected.add(id);
  render();
}

function renderSelection() {
  const n = state.selected.size;
  $("#selection-bar").hidden = n === 0;
  if (!n) return;
  $("#selection-count").textContent = `${n} sélectionné${n > 1 ? "s" : ""}`;
  const sel = $("#selection-move");
  sel.innerHTML = `<option value="" disabled selected>Déplacer</option><option value="__none">Non classés</option>` +
    state.folders.map((f) => `<option value="${esc(f.id)}">${esc(f.name)}</option>`).join("") +
    `<option value="__new">Nouveau dossier</option>`;
}

function copyLinks(ids) {
  const urls = ids.map((id) => state.bookmarks.find((b) => b.id === id)?.url).filter(Boolean);
  navigator.clipboard.writeText(urls.join("\n")).then(() => toast(urls.length > 1 ? "Liens copiés" : "Lien copié"));
}

async function exportIds(ids) {
  if (!ids.length) return toast("Rien à exporter");
  const n = await run(call("Export", ids));
  if (n) toast(`${n} fichier${n > 1 ? "s" : ""} exporté${n > 1 ? "s" : ""}`);
}

async function deleteIds(ids) {
  const ok = await ask({
    title: ids.length > 1 ? `Supprimer ${ids.length} posts ?` : "Supprimer ce post ?",
    ok: "Supprimer",
    danger: true,
  });
  if (!ok) return;
  let snap;
  for (const id of ids) snap = await run(call("Delete", id));
  if (state.viewer) closeViewer();
  if (snap) apply(snap);
}

// --- Adding -------------------------------------------------------------------

const looksLikePost = (s) => /\/status(es)?\/\d+/i.test(s) || /^\d{6,}$/.test(s.trim());

async function add(url) {
  url = url.trim();
  if (!url) return;
  const folderId = state.view.type === "folder" ? state.view.id : "";
  const p = { key: Math.random(), url };
  state.pending.unshift(p);
  renderGrid();
  try {
    const b = await call("Add", url, folderId);
    state.pending = state.pending.filter((x) => x !== p);
    if (state.view.type === "fav") await call("SetFavorite", b.id, true);
    apply(await call("List"));
    toast(`@${b.author} : ${b.media.length} média${b.media.length > 1 ? "s" : ""}`);
  } catch (e) {
    p.error = errText(e);
    renderGrid();
  }
}

$("#add-form").onsubmit = (e) => {
  e.preventDefault();
  const input = $("#add-url");
  const urls = input.value.split(/\s+/).filter(Boolean);
  input.value = "";
  urls.forEach(add);
};

// Ctrl+V outside a text field.
addEventListener("paste", (e) => {
  if (e.target.closest("input, textarea")) return;
  const text = e.clipboardData.getData("text");
  if (looksLikePost(text)) {
    e.preventDefault();
    text.split(/\s+/).filter(looksLikePost).forEach(add);
  }
});

// Links dropped from the browser.
addEventListener("dragover", (e) => {
  if (e.dataTransfer.types.includes("text/uri-list") || (e.dataTransfer.types.includes("text/plain") && !e.dataTransfer.types.includes("text/x-xdll"))) e.preventDefault();
});
addEventListener("drop", (e) => {
  if (e.dataTransfer.types.includes("text/x-xdll")) return;
  const text = e.dataTransfer.getData("text/uri-list") || e.dataTransfer.getData("text/plain");
  if (text && looksLikePost(text)) {
    e.preventDefault();
    text.split(/\s+/).filter(looksLikePost).forEach(add);
  }
});

// --- Viewer -------------------------------------------------------------------

function openViewer(ids, pos) {
  state.viewer = { ids, pos, media: 0 };
  $("#viewer").hidden = false;
  renderViewer();
}

function closeViewer() {
  state.viewer = null;
  $("#viewer").hidden = true;
  $("#v-media").innerHTML = "";
}

function current() {
  const v = state.viewer;
  return v && state.bookmarks.find((b) => b.id === v.ids[v.pos]);
}

let shown = ""; // keeps a playing video across renders
function renderViewer() {
  const b = current();
  if (!b) return closeViewer();
  const v = state.viewer;
  v.media = Math.max(0, Math.min(v.media, b.media.length - 1));
  const m = b.media[v.media];
  if (shown !== m.file || !$("#v-media").firstChild) {
    shown = m.file;
    $("#v-media").innerHTML =
      m.kind === "image"
        ? `<img src="${mediaURL(m.file)}" alt="" />`
        : `<video src="${mediaURL(m.file)}" controls autoplay ${m.kind === "gif" ? "loop muted" : ""} playsinline></video>`;
  }
  $("#v-dots").innerHTML = b.media.length > 1 ? b.media.map((_, i) => `<i class="${i === v.media ? "on" : ""}"></i>`).join("") : "";
  const hasPrev = v.media > 0 || v.pos > 0;
  const hasNext = v.media < b.media.length - 1 || v.pos < v.ids.length - 1;
  $("#v-prev").style.visibility = hasPrev ? "" : "hidden";
  $("#v-next").style.visibility = hasNext ? "" : "hidden";

  $("#v-name").textContent = b.authorName || b.author;
  $("#v-handle").textContent = "@" + b.author;
  $("#v-text").textContent = b.text;
  $("#v-posted").textContent = fmtDate(b.postedAt);
  $("#v-added").textContent = fmtDate(b.addedAt);
  $("#v-count").textContent = `${v.media + 1} / ${b.media.length} · ${fmtSize(totalSize(b))}` + (m.width ? ` · ${m.width}×${m.height}` : "");
  $("#v-folder").innerHTML = `<option value="">Non classés</option>` + state.folders.map((f) => `<option value="${esc(f.id)}">${esc(f.name)}</option>`).join("");
  $("#v-folder").value = b.folderId || "";
  if (document.activeElement !== $("#v-note")) $("#v-note").value = b.note || "";
  const fav = $("#v-fav");
  fav.className = b.favorite ? "on" : "";
  fav.innerHTML = (b.favorite ? icons.starOn : icons.star) + (b.favorite ? "Favori" : "Ajouter aux favoris");
}

function step(d) {
  const v = state.viewer;
  const b = current();
  if (!v || !b) return;
  const i = v.media + d;
  if (i >= 0 && i < b.media.length) v.media = i;
  else if (v.pos + d >= 0 && v.pos + d < v.ids.length) {
    v.pos += d;
    const nb = current();
    v.media = d > 0 ? 0 : nb ? nb.media.length - 1 : 0;
  }
  renderViewer();
}

$("#v-prev").innerHTML = icons.left;
$("#v-next").innerHTML = icons.right;
$("#v-close").innerHTML = icons.close;
$("#v-prev").onclick = () => step(-1);
$("#v-next").onclick = () => step(1);
$("#v-close").onclick = closeViewer;
$(".viewer-stage").onclick = (e) => {
  if (e.target === e.currentTarget) closeViewer();
};
$("#v-folder").onchange = (e) => run(call("Move", [current().id], e.target.value), apply);
$("#v-note").onchange = (e) => run(call("SetNote", current().id, e.target.value), apply);
$("#v-fav").onclick = () => {
  const b = current();
  run(call("SetFavorite", b.id, !b.favorite), apply);
};
$("#v-open").onclick = () => run(call("OpenPost", current().id));
$("#v-reveal").onclick = () => call("Reveal", current().id);
$("#v-file").onclick = () => run(call("OpenMedia", current().media[state.viewer.media].file));
$("#v-export").onclick = () => exportIds([current().id]);
$("#v-delete").onclick = () => deleteIds([current().id]);

// --- Toolbar and keyboard ----------------------------------------------------------

$("#new-folder").innerHTML = icons.plus;
$("#new-folder").onclick = newFolder;
$("#search").oninput = (e) => {
  state.search = e.target.value;
  renderGrid();
};
$("#kind-filter").onclick = (e) => {
  const btn = e.target.closest("button");
  if (!btn) return;
  state.kind = btn.dataset.kind;
  for (const b of $("#kind-filter").children) b.classList.toggle("on", b === btn);
  renderGrid();
};
$("#sort").value = state.sort;
$("#sort").onchange = (e) => {
  state.sort = e.target.value;
  savePrefs();
  renderGrid();
};
const zoom = (px) => document.documentElement.style.setProperty("--thumb", px + "px");
$("#zoom").value = prefs.zoom;
zoom(prefs.zoom);
$("#zoom").oninput = (e) => {
  zoom(e.target.value);
  savePrefs();
};
$("#selection-clear").onclick = () => {
  state.selected.clear();
  render();
};
$("#selection-export").onclick = () => exportIds([...state.selected]);
$("#selection-delete").onclick = () => deleteIds([...state.selected]);
$("#selection-move").onchange = async (e) => {
  const v = e.target.value;
  if (v === "__new") return newFolder();
  await run(call("Move", [...state.selected], v === "__none" ? "" : v), (s) => {
    state.selected.clear();
    apply(s);
  });
};

addEventListener("keydown", (e) => {
  const typing = e.target.closest("input, textarea, select");
  if (!$("#dialog").hidden) return;
  if (state.viewer) {
    if (e.key === "Escape") closeViewer();
    else if (typing) return;
    else if (e.key === "ArrowLeft") step(-1);
    else if (e.key === "ArrowRight") step(1);
    else if (e.key.toLowerCase() === "f") $("#v-fav").click();
    else if (e.key === "Delete") $("#v-delete").click();
    return;
  }
  if (typing) {
    if (e.key === "Escape") e.target.blur();
    return;
  }
  if (e.key === "Escape") {
    state.selected.clear();
    render();
  } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "a") {
    e.preventDefault();
    visible().forEach((b) => state.selected.add(b.id));
    render();
  } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "f") {
    e.preventDefault();
    $("#search").focus();
  } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "l") {
    e.preventDefault();
    $("#add-url").focus();
  } else if (e.key === "Delete" && state.selected.size) {
    deleteIds([...state.selected]);
  }
});

// --- Start -------------------------------------------------------------------------

call("List").then(apply, (e) => toast(errText(e), true));
$("#add-url").focus();
