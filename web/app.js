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
  chevron: svg('<path d="m9 6 6 6-6 6"/>', 'stroke-width="2.4"'),
  play: svg('<path fill="currentColor" d="M7 4v16l13-8z"/>'),
  image: svg('<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-5-5L5 21"/>'),
  gear: svg('<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z"/>'),
  pinterest: '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M12 0a12 12 0 0 0-4.37 23.17c-.1-.95-.2-2.4.04-3.44l1.4-5.96s-.35-.72-.35-1.78c0-1.67.96-2.91 2.17-2.91 1.02 0 1.52.77 1.52 1.69 0 1.03-.66 2.57-1 4-.28 1.19.6 2.16 1.78 2.16 2.13 0 3.77-2.25 3.77-5.49 0-2.86-2.06-4.87-5-4.87-3.41 0-5.41 2.56-5.41 5.2 0 1.03.4 2.14.89 2.74.1.12.11.22.08.34l-.33 1.37c-.05.22-.18.27-.4.16-1.5-.7-2.44-2.88-2.44-4.65 0-3.77 2.75-7.25 7.92-7.25 4.16 0 7.39 2.97 7.39 6.93 0 4.13-2.6 7.46-6.23 7.46-1.21 0-2.36-.63-2.75-1.38l-.75 2.85c-.27 1.04-1 2.35-1.5 3.15A12 12 0 1 0 12 0z"/></svg>',
  link: svg('<path d="M10 14a4.5 4.5 0 0 0 6.4 0l3-3a4.5 4.5 0 0 0-6.4-6.4l-1 1"/><path d="M14 10a4.5 4.5 0 0 0-6.4 0l-3 3a4.5 4.5 0 0 0 6.4 6.4l1-1"/>'),
  search: svg('<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>'),
  check: svg('<path d="M20 6 9 17l-5-5"/>', 'stroke-width="3"'),
  x: '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M18.24 2.25h3.31l-7.23 8.26 8.5 11.24h-6.65l-5.21-6.82-5.97 6.82H1.68l7.73-8.84L1.25 2.25h6.83l4.71 6.23zm-1.16 17.52h1.83L7.08 4.13H5.12z"/></svg>',
  youtube: '<svg viewBox="0 0 24 24"><rect x="1.5" y="4.5" width="21" height="15" rx="4.5" fill="currentColor"/><path d="M10 8.8v6.4l5.5-3.2z" fill="#fff"/></svg>',
  instagram: svg('<rect x="3" y="3" width="18" height="18" rx="5"/><circle cx="12" cy="12" r="4"/><circle cx="17.5" cy="6.5" r=".6" fill="currentColor"/>', 'stroke-width="2.2"'),
};

// Sites posts come from. Older bookmarks have no source: they are from X.
const sources = {
  x: { name: "X", icon: icons.x },
  youtube: { name: "YouTube", icon: icons.youtube },
  instagram: { name: "Instagram", icon: icons.instagram },
  pinterest: { name: "Pinterest", icon: icons.pinterest },
};
const sourceOf = (b) => (sources[b.source] ? b.source : "x");
const sourceIcon = (src) => `<i class="src src-${src}" title="${sources[src].name}">${sources[src].icon}</i>`;
// A YouTube channel without a handle keeps its ID, not worth showing.
const handle = (b) => (/^UC[\w-]{22}$/.test(b.author) ? "" : "@" + b.author);

// --- State ------------------------------------------------------------------

const prefs = load("xdll.prefs", { sort: "added-desc", zoom: 220, quality: "economy", exportOriginal: true });
const state = {
  folders: [],
  bookmarks: [],
  view: { type: "all" }, // all | fav | unfiled | folder (with id)
  search: "",
  kind: "",
  source: "",
  sort: prefs.sort,
  quality: prefs.quality, // of downloaded videos: "economy" or "max"
  exportOriginal: prefs.exportOriginal, // export economy videos in maximum quality
  selected: new Set(),
  pending: [], // { key, url, id, error }
  viewer: null, // { ids, pos, media }
  settings: false, // the settings page is shown instead of the library
  collapsed: new Set(loadList("xdll.collapsed")), // folders whose subfolders are hidden
};

function loadList(key) {
  try {
    return JSON.parse(localStorage.getItem(key) || "[]");
  } catch {
    return [];
  }
}
function saveCollapsed() {
  try {
    localStorage.setItem("xdll.collapsed", JSON.stringify([...state.collapsed]));
  } catch {}
}

function load(key, fallback) {
  try {
    return { ...fallback, ...JSON.parse(localStorage.getItem(key) || "{}") };
  } catch {
    return fallback;
  }
}
function savePrefs() {
  try {
    localStorage.setItem("xdll.prefs", JSON.stringify({ sort: state.sort, zoom: +$("#zoom").value, quality: state.quality, exportOriginal: state.exportOriginal }));
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
const folderOf = (id) => (id ? state.folders.find((f) => f.id === id) : undefined);
const folderName = (id) => folderOf(id)?.name;
const postOf = (id) => state.bookmarks.find((b) => b.id === id);
// The parents of a folder, from the top, and the folder. The Go side
// removes cycles; the bound guards against a bad library anyway.
function folderAncestors(id) {
  const out = [];
  for (let f = folderOf(id); f && out.length < 64; f = folderOf(f.parent)) out.unshift(f);
  return out;
}
const folderPath = (id) => folderAncestors(id).map((f) => f.name).join(" / ");
// isWithin reports whether folder id is ancestor or inside it.
const isWithin = (id, ancestor) => folderAncestors(id).some((f) => f.id === ancestor);
const subfolders = (id) => state.folders.filter((f) => (f.parent || "") === id);
// folderTree lists the folders depth first, with their depth, without the
// subfolders of those skip returns true for.
function folderTree(skip = () => false) {
  const out = [];
  const walk = (parent, depth) => {
    for (const f of subfolders(parent)) {
      out.push({ f, depth });
      if (!skip(f)) walk(f.id, depth + 1);
    }
  };
  walk("", 0);
  return out;
}
const indent = (depth) => "\u00a0\u00a0\u00a0".repeat(depth);
const folderOptions = () => folderTree().map(({ f, depth }) => `<option value="${esc(f.id)}">${indent(depth)}${esc(f.name)}</option>`).join("");
const dateFmt = new Intl.DateTimeFormat("fr-FR", { dateStyle: "medium", timeStyle: "short" });
const fmtDate = (s) => (s && !s.startsWith("0001") ? dateFmt.format(new Date(s)) : "-");
const totalSize = (b) => b.media.reduce((n, m) => n + (m.size || 0), 0);
function fmtSize(n) {
  if (n < 1024 * 1024) return Math.max(1, Math.round(n / 1024)) + " Ko";
  if (n < 1024 * 1024 * 1024) return (n / 1024 / 1024).toFixed(1) + " Mo";
  return (n / 1024 / 1024 / 1024).toFixed(2) + " Go";
}

function toast(text, error = false, onclick = null) {
  const el = document.createElement("div");
  el.className = "toast" + (error ? " error" : "") + (onclick ? " action" : "");
  el.textContent = text;
  if (onclick) el.onclick = () => (el.remove(), onclick());
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

// openMenu shows a menu at x, y. An item is "-" (a separator) or
// { label, run, danger, disabled, hint (a shortcut), checked (true or false
// for an item that can be checked), items (a submenu) }.
function openMenu(x, y, items) {
  closeMenus();
  fillMenu($("#menu"), items);
  placeMenu($("#menu"), x, y);
}

function fillMenu(menu, items) {
  menu.innerHTML = "";
  menu.classList.toggle("checks", items.some((i) => i.checked !== undefined));
  for (const item of items) {
    if (item === "-") {
      menu.append(document.createElement("hr"));
      continue;
    }
    const b = document.createElement("button");
    b.className = item.danger ? "danger" : "";
    b.disabled = !!item.disabled;
    b.innerHTML =
      `<span class="mark">${item.checked ? icons.check : ""}</span><span class="text">${esc(item.label)}</span>` +
      (item.hint ? `<kbd>${esc(item.hint)}</kbd>` : "") +
      (item.items ? icons.right : "");
    if (item.items) {
      b.classList.add("sub");
      b.onmouseenter = b.onclick = () => openSubmenu(b, item.items);
    } else {
      if (menu.id === "menu") b.onmouseenter = () => closeSubmenu();
      b.onclick = () => {
        closeMenus();
        item.run();
      };
    }
    menu.append(b);
  }
}

// placeMenu shows menu at x, y, or at altX when it does not fit on the right.
function placeMenu(menu, x, y, altX = innerWidth) {
  menu.hidden = false;
  const r = menu.getBoundingClientRect();
  const left = x + r.width <= innerWidth - 8 ? x : Math.min(altX - r.width, innerWidth - r.width - 8);
  menu.style.left = Math.max(8, left) + "px";
  menu.style.top = Math.max(8, Math.min(y, innerHeight - r.height - 8)) + "px";
}

function openSubmenu(button, items) {
  const sub = $("#submenu");
  if (sub.owner === button && !sub.hidden) return;
  closeSubmenu();
  sub.owner = button;
  button.classList.add("open");
  fillMenu(sub, items);
  const r = button.getBoundingClientRect();
  placeMenu(sub, r.right + 4, r.top - 7, r.left - 4);
}

function closeSubmenu() {
  const sub = $("#submenu");
  sub.owner?.classList.remove("open");
  sub.owner = null;
  sub.hidden = true;
}

function closeMenus() {
  closeSubmenu();
  $("#menu").hidden = true;
}
const menuOpen = () => !$("#menu").hidden;
addEventListener("mousedown", (e) => !e.target.closest(".menu") && closeMenus(), true);
addEventListener("blur", closeMenus);

// The browser's menu (Back, Reload, Print, Inspect) has no place in the app,
// and Reload stops downloads. Text fields and selected text keep it, for cut,
// copy and paste.
addEventListener("contextmenu", (e) => {
  if (e.defaultPrevented) return;
  if (e.target.closest("textarea, input:not([type=range], [type=radio], [type=checkbox])")) return;
  if (String(getSelection()).trim()) return;
  e.preventDefault();
});

// --- Filtering and sorting ---------------------------------------------------

function inView(b, view = state.view) {
  switch (view.type) {
    case "fav":
      return b.favorite;
    case "unfiled":
      return !b.folderId;
    case "folder": // with its subfolders
      return isWithin(b.folderId, view.id);
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
    if (state.source && sourceOf(b) !== state.source) return false;
    if (!terms.length) return true;
    const hay = `${b.text} @${b.author} ${b.authorName} ${b.note} ${folderPath(b.folderId)} ${sources[sourceOf(b)].name}`.toLowerCase();
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
  if (state.settings) renderStorage();
  renderGrid();
  renderSelection();
  if (state.viewer) renderViewer();
}

function navLink({ icon, label, count, on, onClick, folderId, onContext, depth, open }) {
  const a = document.createElement("a");
  a.className = on ? "on" : "";
  a.innerHTML = `${icon}<span class="label">${esc(label)}</span><span class="count">${count}</span>`;
  a.onclick = onClick;
  if (onContext) a.oncontextmenu = (e) => (e.preventDefault(), onContext(e));
  if (depth !== undefined) {
    // A folder: indented, with a chevron when it has subfolders, and
    // draggable into another one.
    a.style.setProperty("--depth", depth);
    a.classList.add("tree");
    const chev = document.createElement("button");
    chev.className = "chev" + (open ? " open" : "");
    chev.tabIndex = -1;
    if (open !== undefined) {
      chev.innerHTML = icons.chevron;
      chev.title = open ? "Replier" : "Déplier";
      chev.onclick = (e) => {
        e.stopPropagation();
        open ? state.collapsed.add(folderId) : state.collapsed.delete(folderId);
        saveCollapsed();
        renderSidebar();
      };
    }
    a.prepend(chev);
    a.draggable = true;
    a.ondragstart = (e) => {
      e.dataTransfer.setData("text/x-xdll-folder", folderId);
      e.dataTransfer.effectAllowed = "move";
    };
  }
  if (folderId !== undefined) dropTarget(a, folderId);
  return a;
}

// dropTarget lets posts be dropped on el to file them in folderId, and
// folders to put them inside it (at the top level for "").
function dropTarget(el, folderId, postsToo = true) {
  const accepts = (e) => {
    const t = e.dataTransfer.types;
    return (postsToo && t.includes("text/x-xdll")) || t.includes("text/x-xdll-folder");
  };
  el.ondragover = (e) => {
    if (!accepts(e)) return;
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    el.classList.add("drop");
  };
  el.ondragleave = () => el.classList.remove("drop");
  el.ondrop = (e) => {
    if (!accepts(e)) return;
    e.preventDefault();
    el.classList.remove("drop");
    const where = folderPath(folderId);
    const folder = e.dataTransfer.getData("text/x-xdll-folder");
    if (folder) {
      if (folder === folderId || (folderOf(folder)?.parent || "") === folderId) return;
      return run(call("MoveFolder", folder, folderId), (s) => {
        if (state.collapsed.delete(folderId)) saveCollapsed();
        apply(s);
        toast(where ? `« ${folderName(folder)} » déplacé dans ${where}` : `« ${folderName(folder)} » remis au premier niveau`);
      });
    }
    if (!postsToo) return;
    const ids = JSON.parse(e.dataTransfer.getData("text/x-xdll") || "[]");
    if (ids.length)
      run(call("Move", ids, folderId), (s) => {
        apply(s);
        toast(`Déplacé dans ${where || "Non classés"}`);
      });
  };
}

function setView(view) {
  if (state.settings) closeSettings();
  state.view = view;
  state.selected.clear();
  render();
  $("#grid").scrollTop = 0;
}

function renderSidebar() {
  const v = state.settings ? {} : state.view;
  const n = (view) => state.bookmarks.filter((b) => inView(b, view)).length;
  const views = $("#views");
  views.replaceChildren(
    navLink({ icon: icons.all, label: "Tout", count: state.bookmarks.length, on: v.type === "all", onClick: () => setView({ type: "all" }) }),
    navLink({ icon: icons.star, label: "Favoris", count: n({ type: "fav" }), on: v.type === "fav", onClick: () => setView({ type: "fav" }) }),
    navLink({ icon: icons.inbox, label: "Non classés", count: n({ type: "unfiled" }), on: v.type === "unfiled", onClick: () => setView({ type: "unfiled" }), folderId: "" }),
  );
  const folders = $("#folders");
  folders.replaceChildren(
    ...folderTree((f) => state.collapsed.has(f.id)).map(({ f, depth }) =>
      navLink({
        icon: icons.folder,
        label: f.name,
        count: n({ type: "folder", id: f.id }),
        on: v.type === "folder" && v.id === f.id,
        onClick: () => setView({ type: "folder", id: f.id }),
        folderId: f.id,
        onContext: (e) => folderMenu(f, e.clientX, e.clientY),
        depth,
        open: subfolders(f.id).length ? !state.collapsed.has(f.id) : undefined,
      }),
    ),
  );
  if (!state.folders.length) {
    folders.innerHTML = `<div class="hint">Aucun dossier</div>`;
  }
  $("#open-settings").className = state.settings ? "on" : "";
  $("#title").textContent = viewTitle();
  $("#title").title = v.type === "folder" ? folderPath(v.id) : "";
  const size = state.bookmarks.reduce((s, b) => s + totalSize(b), 0);
  $("#brand-stats").textContent = `${state.bookmarks.length} post${state.bookmarks.length > 1 ? "s" : ""}` + (size ? ` · ${fmtSize(size)}` : "");
}

function statsText() {
  const size = state.bookmarks.reduce((s, b) => s + totalSize(b), 0);
  const media = state.bookmarks.reduce((s, b) => s + b.media.length, 0);
  return `${state.bookmarks.length} posts · ${media} médias · ${fmtSize(size)}`;
}

function folderMenu(f, x, y) {
  openMenu(x, y, [
    { label: "Nouveau sous-dossier", run: () => newFolder(f.id) },
    { label: "Renommer", run: () => renameFolder(f) },
    { label: "Exporter", run: () => exportIds(state.bookmarks.filter((b) => isWithin(b.folderId, f.id)).map((b) => b.id)) },
    ...(f.parent ? [{ label: "Remettre au premier niveau", run: () => run(call("MoveFolder", f.id, ""), apply) }] : []),
    "-",
    { label: "Supprimer le dossier", danger: true, run: () => deleteFolder(f) },
  ]);
}

// newFolder creates a folder in parent, files ids in it and opens it.
async function newFolder(parent = "", ids = [...state.selected]) {
  const name = await ask({ title: parent ? "Nouveau sous-dossier" : "Nouveau dossier", text: parent ? `Dans ${folderPath(parent)}` : "", value: "", ok: "Créer" });
  if (name === null) return;
  const f = await run(call("CreateFolder", name, parent));
  if (!f) return;
  if (state.collapsed.delete(parent)) saveCollapsed();
  if (ids.length) await run(call("Move", ids, f.id), apply);
  else apply(await call("List"));
  setView({ type: "folder", id: f.id });
}

async function renameFolder(f) {
  const name = await ask({ title: "Renommer le dossier", value: f.name, ok: "Renommer" });
  if (name !== null) run(call("RenameFolder", f.id, name), apply);
}

async function deleteFolder(f) {
  const count = state.bookmarks.filter((b) => b.folderId === f.id).length;
  const subs = subfolders(f.id).length;
  const what = [count && `${count} post${count > 1 ? "s" : ""}`, subs && `${subs} sous-dossier${subs > 1 ? "s" : ""}`].filter(Boolean).join(" et ");
  const to = f.parent ? `dans ${folderPath(f.parent)}` : !subs ? "dans Non classés" : count ? "au premier niveau (les posts dans Non classés)" : "au premier niveau";
  const ok = await ask({
    title: `Supprimer « ${f.name} » ?`,
    text: what ? `${what} ${count + subs > 1 ? "iront" : "ira"} ${to}.` : "",
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
      <div class="info"><div class="author">${p.source ? sourceIcon(p.source) : ""}<b>${p.error ? "Échec" : "Téléchargement"}</b></div><div class="snippet">${esc(p.url)}</div></div>`;
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
    const folder = b.folderId && !(state.view.type === "folder" && b.folderId === state.view.id) ? folderName(b.folderId) : "";
    el.innerHTML = `<div class="thumb">${thumbHTML(b.media[0])}<div class="badges">${badgesHTML(b)}</div>
        <span class="check">${icons.check}</span>
        <button class="star${b.favorite ? " on" : ""}" title="Favori">${b.favorite ? icons.starOn : icons.star}</button></div>
      <div class="info">
        <div class="author">${sourceIcon(sourceOf(b))}<b>${esc(b.authorName || b.author)}</b><span>${esc(handle(b))}</span></div>
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
      ? `<span class="empty-icon">${icons.search}</span><b>Aucun résultat</b><p>Aucun post ne correspond à ces filtres.</p>`
      : `<span class="empty-icon">${icons.link}</span><b>Aucun post</b><p>Colle le lien d'un post X, Instagram, Pinterest ou d'une vidéo YouTube avec <kbd>Ctrl</kbd> <kbd>V</kbd>, ou glisse-le ici.</p>`;
  }
}

function cardMenu(b, x, y) {
  const ids = state.selected.has(b.id) ? [...state.selected] : [b.id];
  const n = ids.length;
  const many = n > 1;
  const posts = ids.map(postOf).filter(Boolean);
  const files = posts.reduce((s, p) => s + p.media.length, 0);
  const fav = posts.every((p) => p.favorite);
  openMenu(x, y, [
    ...(many
      ? []
      : [
          { label: "Ouvrir", run: () => openViewer([b.id], 0) },
          { label: `Ouvrir sur ${sources[sourceOf(b)].name}`, run: () => run(call("OpenPost", b.id)) },
          { label: "Afficher dans l'Explorateur", run: () => call("Reveal", b.id) },
          "-",
          { label: "Copier pour Claude", run: () => copyForClaude(b.id, 0, "all") },
          { label: "Copier le texte", run: () => copyForClaude(b.id, 0, "text") },
          { label: "Copier l'image", run: () => copyForClaude(b.id, 0, "image") },
        ]),
    { label: files > 1 ? `Copier les fichiers (${files})` : "Copier le fichier", run: () => copyFiles(ids, -1) },
    { label: many ? `Copier les liens (${n})` : "Copier le lien", run: () => copyLinks(ids) },
    "-",
    { label: fav ? "Retirer des favoris" : "Ajouter aux favoris", run: () => setFavorites(ids, !fav) },
    { label: "Déplacer vers", items: moveItems(ids) },
    many
      ? { label: "Tout désélectionner", hint: "Échap", run: () => (state.selected.clear(), render()) }
      : { label: state.selected.has(b.id) ? "Désélectionner" : "Sélectionner", run: () => toggleSelect(b.id) },
    { label: many ? `Exporter (${n})` : "Exporter", run: () => exportIds(ids) },
    "-",
    { label: many ? `Supprimer (${n})` : "Supprimer", danger: true, run: () => deleteIds(ids) },
  ]);
}

// moveItems lists the folders to file ids in, checking the one they are in.
function moveItems(ids) {
  const at = new Set(ids.map((id) => postOf(id)?.folderId || ""));
  const only = at.size === 1 ? [...at][0] : null;
  return [
    { label: "Non classés", checked: only === "", run: () => moveTo(ids, "") },
    ...folderTree().map(({ f, depth }) => ({ label: indent(depth) + f.name, checked: only === f.id, run: () => moveTo(ids, f.id) })),
    "-",
    { label: "Nouveau dossier…", run: () => newFolder(state.view.type === "folder" ? state.view.id : "", ids) },
  ];
}

function moveTo(ids, folderId) {
  run(call("Move", ids, folderId), (s) => {
    apply(s);
    toast(`Déplacé dans ${folderPath(folderId) || "Non classés"}`);
  });
}

async function setFavorites(ids, favorite) {
  let snap;
  for (const id of ids) snap = await run(call("SetFavorite", id, favorite));
  if (!snap) return;
  apply(snap);
  if (ids.length > 1) toast(favorite ? `${ids.length} posts ajoutés aux favoris` : `${ids.length} posts retirés des favoris`);
}

async function copyFiles(ids, index) {
  const n = await run(call("CopyFiles", ids, index));
  if (n) toast(n > 1 ? `${n} fichiers copiés : à coller dans l'Explorateur ou une discussion` : "Fichier copié : à coller dans l'Explorateur ou une discussion");
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
    folderOptions() +
    `<option value="__new">Nouveau dossier</option>`;
}

function copyLinks(ids) {
  const urls = ids.map((id) => state.bookmarks.find((b) => b.id === id)?.url).filter(Boolean);
  navigator.clipboard.writeText(urls.join("\n")).then(() => toast(urls.length > 1 ? "Liens copiés" : "Lien copié"));
}

async function copyForClaude(id, index, what) {
  const copied = await run(call("CopyForClaude", id, index, what));
  if (copied) toast({ all: "Copié : texte + image", text: "Copié : texte", image: "Copié : image" }[copied]);
}

async function exportIds(ids) {
  if (!ids.length) return toast("Rien à exporter");
  const light = ids.some((id) => state.bookmarks.find((b) => b.id === id)?.media.some((m) => m.light));
  const r = await run(call("Export", ids, state.exportOriginal && light));
  if (!r?.files) return;
  const s = (n) => (n > 1 ? "s" : "");
  let text = `${r.files} fichier${s(r.files)} exporté${s(r.files)}`;
  if (r.downloaded) text += `, dont ${r.downloaded} vidéo${s(r.downloaded)} en qualité maximale`;
  if (r.light) text += ` (${r.light} plus en ligne, en version économie)`;
  toast(text);
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

// Same as parsePostID in Go: the tweet ID, "yt-", "ig-" or "pin-" and the site's ID.
const youtubeID = /(?:youtube\.com\/(?:shorts\/|live\/|embed\/|watch\?(?:[^#]*&)?v=)|youtu\.be\/)([\w-]{11})/i;
const instagramCode = /instagram\.com\/(?:[\w.]+\/)?(?:p|reels?|tv)\/([\w-]{5,})/i;
const pinterestID = /pinterest\.[a-z.]+\/pin\/(?:[^/?#]*--)?(\d{6,20})/i;
const pinItLink = /^\s*(?:https?:\/\/)?pin\.it\/\w+/i; // resolved by Go
const looksLikePost = (s) => /\/status(es)?\/\d+/i.test(s) || /^\d{6,}$/.test(s.trim()) || youtubeID.test(s) || instagramCode.test(s) || pinterestID.test(s) || pinItLink.test(s);
function postID(s) {
  const yt = s.match(youtubeID)?.[1];
  if (yt) return "yt-" + yt;
  const ig = s.match(instagramCode)?.[1];
  if (ig) return "ig-" + ig;
  const pin = s.match(pinterestID)?.[1];
  if (pin) return "pin-" + pin;
  return s.match(/\/status(?:es)?\/(\d+)/i)?.[1] ?? (/^\d+$/.test(s) ? s : null);
}
const sourceOfID = (id) => (id?.startsWith("yt-") ? "youtube" : id?.startsWith("ig-") ? "instagram" : id?.startsWith("pin-") ? "pinterest" : id ? "x" : "");

async function add(url) {
  url = url.trim();
  if (!url) return;
  const id = postID(url);
  const dup = state.bookmarks.find((b) => b.id === id);
  if (dup) {
    const where = folderPath(dup.folderId) || "Non classés";
    return toast(`Déjà enregistré : ${dup.authorName || dup.author} sur ${sources[sourceOf(dup)].name} (${where}). Cliquer pour l'ouvrir`, true, () => openViewer([dup.id], 0));
  }
  if (id && state.pending.some((p) => p.id === id && !p.error)) return toast("Déjà en cours de téléchargement", true);
  if (state.settings) closeSettings();
  const folderId = state.view.type === "folder" ? state.view.id : "";
  const p = { key: Math.random(), url, id, source: pinItLink.test(url) ? "pinterest" : sourceOfID(id) };
  state.pending.unshift(p);
  renderGrid();
  try {
    const b = await call("Add", url, folderId, state.quality);
    state.pending = state.pending.filter((x) => x !== p);
    if (state.view.type === "fav") await call("SetFavorite", b.id, true);
    apply(await call("List"));
    toast(`${sources[sourceOf(b)].name}, ${b.authorName || b.author} : ${b.media.length} média${b.media.length > 1 ? "s" : ""}`);
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

  const src = sourceOf(b);
  $("#v-name").textContent = b.authorName || b.author;
  $("#v-handle").innerHTML = `${sourceIcon(src)}<span>${esc([handle(b), sources[src].name].filter(Boolean).join(" · "))}</span>`;
  $("#v-open").textContent = `Ouvrir sur ${sources[src].name}`;
  $("#v-text").textContent = b.text;
  $("#v-posted").textContent = fmtDate(b.postedAt);
  $("#v-added").textContent = fmtDate(b.addedAt);
  $("#v-count").textContent = `${v.media + 1} / ${b.media.length} · ${fmtSize(totalSize(b))}` + (m.width ? ` · ${m.width}×${m.height}` : "");
  $("#v-folder").innerHTML = `<option value="">Non classés</option>` + folderOptions();
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
$("#v-claude").onclick = () => copyForClaude(current().id, state.viewer.media, "all");
$("#v-claude-text").onclick = () => copyForClaude(current().id, state.viewer.media, "text");
$("#v-claude-image").onclick = () => copyForClaude(current().id, state.viewer.media, "image");
$("#v-reveal").onclick = () => call("Reveal", current().id);
$("#v-file").onclick = () => run(call("OpenMedia", current().media[state.viewer.media].file));
$("#v-export").onclick = () => exportIds([current().id]);
$("#v-delete").onclick = () => deleteIds([current().id]);

// --- Settings -------------------------------------------------------------------

async function openSettings() {
  if (state.viewer) closeViewer();
  state.settings = true;
  $(".main").hidden = true;
  $("#settings").hidden = false;
  render();
  showTheme(await run(call("Theme")));
  showSettings(await run(call("Settings")));
  showUpdates(await run(call("Updates")));
  if ((await run(call("OptimizeStatus")))?.running) watchOptimize();
}

// --- Appearance ---

const themes = { system: "Système", light: "Clair", dark: "Sombre" };
// A small window of the app in that theme; the system's is half of each.
const themeMock = (t) => `<span class="mock mock-${t}"><i class="m-side"></i><i class="m-card"><i></i><i></i><i class="m-btn"></i></i></span>`;
$("#themes").innerHTML = Object.entries(themes)
  .map(([k, label]) => `<label class="theme-choice"><input type="radio" name="theme" value="${k}" />
    <span class="theme-preview">${k === "system" ? themeMock("light") + themeMock("dark") : themeMock(k)}</span><b>${label}</b></label>`)
  .join("");
function showTheme(t) {
  if (!t) return;
  for (const r of document.querySelectorAll('input[name="theme"]')) r.checked = r.value === t;
}
for (const r of document.querySelectorAll('input[name="theme"]')) {
  r.onchange = async () => showTheme((await run(call("SetTheme", r.value))) ?? (await call("Theme")));
}

function showQuality() {
  for (const r of document.querySelectorAll('input[name="quality"]')) r.checked = r.value === state.quality;
  $("#export-original").checked = state.exportOriginal;
}
$("#export-original").onchange = (e) => {
  state.exportOriginal = e.target.checked;
  savePrefs();
};
for (const r of document.querySelectorAll('input[name="quality"]')) {
  r.onchange = () => {
    state.quality = r.value;
    savePrefs();
  };
}
showQuality();

// --- Lighter library ---

// Videos economy could make lighter: Instagram and Pinterest have a single version.
const heavyVideos = () =>
  state.bookmarks.filter((b) => !["instagram", "pinterest"].includes(sourceOf(b))).flatMap((b) => b.media.filter((m) => m.kind === "video" && !m.light));

let optimizing = false;

function renderOptimize() {
  if (optimizing) return;
  const videos = heavyVideos();
  const size = videos.reduce((n, m) => n + (m.size || 0), 0);
  $("#opt-text").textContent = videos.length
    ? `${videos.length} vidéo${videos.length > 1 ? "s" : ""} X ou YouTube en qualité maximale (${fmtSize(size)}). ` +
      "Chacune est re-téléchargée en version économie, gardée si elle est plus légère. Celles qui ne sont plus en ligne restent telles quelles."
    : "Toutes les vidéos sont en version économie.";
  $("#optimize").disabled = !videos.length;
}

async function optimize() {
  const n = heavyVideos().length;
  const ok = await ask({
    title: "Alléger la bibliothèque ?",
    text: `${n} vidéo${n > 1 ? "s" : ""} en qualité maximale ser${n > 1 ? "ont remplacées" : "a remplacée"} par la version économie.` + (state.exportOriginal ? " À l'export, l'original est re-téléchargé tant qu'il est en ligne." : ""),
    ok: "Alléger",
  });
  if (!ok) return;
  run(call("Optimize"));
  watchOptimize();
}

// watchOptimize shows the progress of Optimize until it ends.
async function watchOptimize() {
  if (optimizing) return;
  optimizing = true;
  $("#optimize").hidden = true;
  $("#optimize-cancel").hidden = false;
  $("#opt-progress").hidden = false;
  let st;
  let done = -1;
  do {
    await new Promise((r) => setTimeout(r, 400));
    st = await call("OptimizeStatus");
    $("#opt-progress i").style.width = (st.total ? (100 * st.done) / st.total : 100) + "%";
    $("#opt-status").textContent = `${st.done} / ${st.total} posts · ${fmtSize(Math.max(st.saved, 1024))} libérés`;
    if (st.done !== done) {
      done = st.done;
      apply(await call("List"));
    }
  } while (st.running);
  optimizing = false;
  $("#optimize").hidden = false;
  $("#optimize-cancel").hidden = true;
  $("#opt-progress").hidden = true;
  $("#opt-status").textContent = "";
  renderOptimize();
  const left = st.failed ? `, ${st.failed} post${st.failed > 1 ? "s" : ""} laissé${st.failed > 1 ? "s" : ""} tel${st.failed > 1 ? "s" : ""} quel${st.failed > 1 ? "s" : ""}` : "";
  toast((st.done < st.total ? "Arrêté : " : "") + `${fmtSize(Math.max(st.saved, 1024))} libérés` + left);
}

function closeSettings() {
  state.settings = false;
  $("#settings").hidden = true;
  $(".main").hidden = false;
  renderSidebar();
}

// renderStorage shows what the library holds, by site and by kind.
function renderStorage() {
  const bySource = {};
  const byKind = {};
  let media = 0;
  for (const b of state.bookmarks) {
    const src = (bySource[sourceOf(b)] ??= { posts: 0, size: 0 });
    src.posts++;
    media += b.media.length;
    for (const m of b.media) {
      const k = (byKind[m.kind] ??= { n: 0, size: 0 });
      k.n++;
      k.size += m.size || 0;
      src.size += m.size || 0;
    }
  }
  const size = Object.values(bySource).reduce((n, s) => n + s.size, 0);
  const sizeText = (n) => (n ? fmtSize(n) : "0 Ko");
  $("#st-posts").textContent = state.bookmarks.length;
  $("#st-media").textContent = media;
  $("#st-size").textContent = sizeText(size);
  const present = Object.keys(sources).filter((k) => bySource[k]);
  $("#st-bar").innerHTML = present
    .filter((k) => bySource[k].size)
    .map((k) => `<i class="src-${k}" style="flex: ${bySource[k].size}" title="${sources[k].name} : ${fmtSize(bySource[k].size)}"></i>`)
    .join("");
  $("#st-sources").innerHTML = present
    .map((k) => `<span>${sourceIcon(k)}${sources[k].name} <small>${bySource[k].posts} post${bySource[k].posts > 1 ? "s" : ""} · ${sizeText(bySource[k].size)}</small></span>`)
    .join("");
  renderOptimize();
  $("#st-kinds").innerHTML = [["image", "Images"], ["video", "Vidéos"], ["gif", "GIF"]]
    .filter(([k]) => byKind[k])
    .map(([k, label]) => `<span>${label} : ${byKind[k].n} · ${sizeText(byKind[k].size)}</span>`)
    .join("");
}

function showSettings(info) {
  if (!info) return;
  $("#data-dir").textContent = info.dir;
  $("#data-version").textContent = info.version || "dev";
  $("#reset-data-dir").hidden = info.default;
  $("#data-missing").hidden = !info.missing;
  $("#data-missing").textContent = `Dossier introuvable : ${info.missing}. Le dossier par défaut est utilisé en attendant (disque débranché ?).`;
}

// moveData moves the library to c.dir, or switches to the library there.
async function moveData(c) {
  if (!c?.dir) return;
  const ok = await ask(
    c.hasLibrary
      ? { title: "Utiliser la bibliothèque de ce dossier ?", text: `${c.dir} contient déjà une bibliothèque xdll. L'actuelle reste où elle est.`, ok: "Utiliser" }
      : { title: "Déplacer les données ?", text: `${statsText()} vont être déplacés vers ${c.dir}.`, ok: "Déplacer" },
  );
  if (!ok) return;
  await busy($("#change-data-dir"), c.hasLibrary ? "Changement…" : "Déplacement…", async () => {
    const s = await run(call(c.hasLibrary ? "UseData" : "MoveData", c.dir));
    if (!s) return;
    showSettings(s);
    apply(await call("List"));
    toast("Données dans " + s.dir);
  });
}

function showUpdates(u) {
  if (!u) return;
  $("#update-off").hidden = u.enabled;
  for (const el of [$("#auto-check"), $("#auto-install"), $("#check-updates")]) el.disabled = !u.enabled;
  $("#auto-check").checked = u.automaticChecks;
  $("#auto-install").checked = u.automaticDownloads;
  $("#last-check").textContent = u.lastCheck && !u.lastCheck.startsWith("0001") ? "Dernière vérification : " + fmtDate(u.lastCheck) : "";
}

const setUpdates = () => run(call("SetUpdates", $("#auto-check").checked, $("#auto-install").checked), showUpdates);

// busy disables the settings buttons while fn runs.
async function busy(btn, label, fn) {
  const buttons = [...$("#settings").querySelectorAll(".settings-row button:not(:disabled)")];
  const text = btn.textContent;
  buttons.forEach((b) => (b.disabled = true));
  btn.textContent = label;
  try {
    await fn();
  } finally {
    buttons.forEach((b) => (b.disabled = false));
    btn.textContent = text;
  }
}

function backup() {
  let prefs = "";
  try {
    prefs = localStorage.getItem("xdll.prefs") || "";
  } catch {}
  return busy($("#backup"), "Sauvegarde…", async () => {
    const r = await run(call("Backup", prefs));
    if (r?.file) toast(`Sauvegardé : ${r.posts} posts, ${r.files} fichiers`);
  });
}

async function restore(replace) {
  if (replace) {
    const ok = await ask({
      title: "Remplacer la bibliothèque ?",
      text: `Les ${state.bookmarks.length} posts actuels et leurs fichiers seront effacés et remplacés par la sauvegarde.`,
      ok: "Choisir la sauvegarde",
      danger: true,
    });
    if (!ok) return;
  }
  const btn = replace ? $("#restore-replace") : $("#restore-merge");
  return busy(btn, "Import…", async () => {
    const r = await run(call("Restore", replace));
    if (!r) return;
    if (r.prefs) {
      try {
        localStorage.setItem("xdll.prefs", r.prefs);
        const p = JSON.parse(r.prefs);
        if (p.sort) $("#sort").value = state.sort = p.sort;
        if (p.zoom) zoom(($("#zoom").value = p.zoom));
        if (p.quality) state.quality = p.quality;
        if (p.exportOriginal !== undefined) state.exportOriginal = p.exportOriginal;
        showQuality();
      } catch {}
    }
    state.selected.clear();
    if (replace) state.view = { type: "all" };
    apply(r.library);
    toast(`Importé : ${r.added} post${r.added > 1 ? "s" : ""}` + (r.skipped ? `, ${r.skipped} déjà présent${r.skipped > 1 ? "s" : ""}` : ""));
  });
}

$("#open-settings").innerHTML = `${icons.gear}<span class="label">Paramètres</span>`;
$("#open-settings").onclick = openSettings;
$("#backup").onclick = backup;
$("#optimize").onclick = optimize;
$("#optimize-cancel").onclick = () => call("CancelOptimize");
$("#restore-merge").onclick = () => restore(false);
$("#restore-replace").onclick = () => restore(true);
$("#auto-check").onchange = setUpdates;
$("#auto-install").onchange = setUpdates;
$("#check-updates").onclick = () => call("CheckForUpdates");
$("#change-data-dir").onclick = async () => moveData(await run(call("ChooseDataDir")));
$("#reset-data-dir").onclick = async () => moveData(await run(call("DefaultDataDir")));
$("#open-data-dir").onclick = () => run(call("OpenDataDir"));

// --- Toolbar and keyboard ----------------------------------------------------------

$("#new-folder").innerHTML = icons.plus;
$("#add-form").insertAdjacentHTML("afterbegin", icons.link);
$(".search").insertAdjacentHTML("afterbegin", icons.search);
$("#new-folder").onclick = () => newFolder();
dropTarget($("#folders-head"), "", false);
$("#search").oninput = (e) => {
  state.search = e.target.value;
  renderGrid();
};
$("#source-filter").innerHTML =
  `<button data-source="" class="on">Tout</button>` +
  Object.entries(sources).map(([k, s]) => `<button data-source="${k}" title="${s.name}" aria-label="${s.name}">${sourceIcon(k)}</button>`).join("");
$("#source-filter").onclick = (e) => {
  const btn = e.target.closest("button");
  if (!btn) return;
  state.source = btn.dataset.source;
  for (const b of $("#source-filter").children) b.classList.toggle("on", b === btn);
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
  if (v === "__new") return newFolder(state.view.type === "folder" ? state.view.id : "");
  await run(call("Move", [...state.selected], v === "__none" ? "" : v), (s) => {
    state.selected.clear();
    apply(s);
  });
};

addEventListener("keydown", (e) => {
  if (menuOpen()) {
    if (e.key === "Escape") closeMenus();
    return;
  }
  const typing = e.target.closest("input, textarea, select");
  if (!$("#dialog").hidden) return;
  if (state.settings) {
    if (e.key === "Escape" && !typing) closeSettings();
    return;
  }
  if (state.viewer) {
    if (e.key === "Escape") closeViewer();
    else if (typing) return;
    else if (e.key === "ArrowLeft") step(-1);
    else if (e.key === "ArrowRight") step(1);
    else if (e.key.toLowerCase() === "f") $("#v-fav").click();
    else if (e.ctrlKey || e.metaKey || e.altKey) return;
    else if (e.key.toLowerCase() === "c") $("#v-claude").click();
    else if (e.key.toLowerCase() === "t") $("#v-claude-text").click();
    else if (e.key.toLowerCase() === "i") $("#v-claude-image").click();
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
call("Settings").then((s) => s.missing && toast(`Dossier des données introuvable : ${s.missing}. Voir Paramètres`, true, openSettings));
$("#add-url").focus();
