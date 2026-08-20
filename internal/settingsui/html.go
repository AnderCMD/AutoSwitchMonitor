package settingsui

// pageHTML es la ventana de configuración de hotkeys: una sola página HTML
// autocontenida (sin red, sin CDN) que webview carga vía SetHtml. Habla con
// Go a través de las funciones expuestas con w.Bind(...): getState,
// saveHotkeys y closeWindow.
const pageHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<style>
  :root { color-scheme: light dark; }
  body {
    font-family: -apple-system, "Segoe UI", system-ui, sans-serif;
    margin: 0; padding: 16px 20px 20px;
    background: Canvas; color: CanvasText;
  }
  h1 { font-size: 15px; margin: 0 0 4px; }
  p.hint { font-size: 12px; opacity: 0.7; margin: 0 0 14px; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th { text-align: left; font-size: 11px; text-transform: uppercase; opacity: 0.6; padding: 4px 6px; }
  td { padding: 6px; border-top: 1px solid rgba(128,128,128,0.25); vertical-align: middle; }
  select, .combo {
    font-size: 13px; padding: 4px 6px; border-radius: 6px;
    border: 1px solid rgba(128,128,128,0.4); background: Field; color: FieldText;
  }
  .combo { display: inline-block; min-width: 120px; }
  .combo.empty { opacity: 0.5; font-style: italic; }
  .combo.recording { border-color: #f59e0b; color: #f59e0b; }
  button {
    font-size: 12px; padding: 5px 10px; border-radius: 6px; border: 1px solid rgba(128,128,128,0.4);
    background: ButtonFace; color: ButtonText; cursor: pointer;
  }
  button.primary { background: #3b82f6; border-color: #3b82f6; color: white; }
  button.danger { border-color: #ef4444; color: #ef4444; background: transparent; }
  .row-actions { display: flex; gap: 6px; }
  .footer { display: flex; justify-content: space-between; align-items: center; margin-top: 16px; }
  .footer-buttons { display: flex; gap: 8px; }
  #error { color: #ef4444; font-size: 12px; min-height: 16px; }
  #addRow { margin-top: 10px; }
</style>
</head>
<body>
  <h1>Atajos de teclado</h1>
  <p class="hint">Click "Grabar" y presiona la combinación (ej. Ctrl+Alt+1). Se guarda al presionar "Guardar".</p>

  <table>
    <thead>
      <tr><th>Cambiar a</th><th>Atajo</th><th></th></tr>
    </thead>
    <tbody id="rows"></tbody>
  </table>

  <button id="addRow">+ Agregar atajo</button>

  <div class="footer">
    <div id="error"></div>
    <div class="footer-buttons">
      <button id="cancel">Cancelar</button>
      <button id="save" class="primary">Guardar</button>
    </div>
  </div>

<script>
let state = { hotkeys: [], inputs: [] };
let recordingIndex = -1;

function comboLabel(hk) {
  if (!hk.modifiers || !hk.modifiers.length || !hk.key) return null;
  const mods = hk.modifiers.map(m => m.charAt(0).toUpperCase() + m.slice(1));
  return mods.concat([hk.key.toUpperCase()]).join("+");
}

function render() {
  const tbody = document.getElementById("rows");
  tbody.innerHTML = "";
  state.hotkeys.forEach((hk, i) => {
    const tr = document.createElement("tr");

    const tdInput = document.createElement("td");
    const select = document.createElement("select");
    state.inputs.forEach(name => {
      const opt = document.createElement("option");
      opt.value = name;
      opt.textContent = name.toUpperCase();
      if (name === hk.target) opt.selected = true;
      select.appendChild(opt);
    });
    select.onchange = () => { hk.target = select.value; };
    tdInput.appendChild(select);

    const tdCombo = document.createElement("td");
    const comboSpan = document.createElement("span");
    const label = comboLabel(hk);
    comboSpan.className = "combo" + (label ? "" : " empty") + (recordingIndex === i ? " recording" : "");
    comboSpan.textContent = recordingIndex === i ? "Presiona una combinación…" : (label || "sin asignar");
    tdCombo.appendChild(comboSpan);

    const tdActions = document.createElement("td");
    const actions = document.createElement("div");
    actions.className = "row-actions";

    const recBtn = document.createElement("button");
    recBtn.textContent = recordingIndex === i ? "Cancelar" : "Grabar";
    recBtn.onclick = () => {
      recordingIndex = (recordingIndex === i) ? -1 : i;
      render();
    };
    actions.appendChild(recBtn);

    const delBtn = document.createElement("button");
    delBtn.className = "danger";
    delBtn.textContent = "Eliminar";
    delBtn.onclick = () => {
      state.hotkeys.splice(i, 1);
      if (recordingIndex === i) recordingIndex = -1;
      render();
    };
    actions.appendChild(delBtn);

    tdActions.appendChild(actions);
    tr.appendChild(tdInput);
    tr.appendChild(tdCombo);
    tr.appendChild(tdActions);
    tbody.appendChild(tr);
  });
}

document.getElementById("addRow").onclick = () => {
  state.hotkeys.push({ modifiers: [], key: "", target: state.inputs[0] || "" });
  render();
};

document.getElementById("cancel").onclick = () => { closeWindow(); };

document.getElementById("save").onclick = () => {
  const err = document.getElementById("error");
  err.textContent = "";
  for (const hk of state.hotkeys) {
    if (!hk.modifiers.length || !hk.key) {
      err.textContent = "Hay un atajo sin grabar. Grábalo o elimínalo.";
      return;
    }
  }
  saveHotkeys(JSON.stringify(state.hotkeys)).then(() => {
    closeWindow();
  }).catch(e => {
    err.textContent = "Error: " + e;
  });
};

const MODIFIER_KEYS = { "Control": "ctrl", "Alt": "alt", "Shift": "shift", "Meta": "cmd" };

document.addEventListener("keydown", (e) => {
  if (recordingIndex < 0) return;
  e.preventDefault();
  e.stopPropagation();

  if (MODIFIER_KEYS[e.key]) return; // esperando la tecla no-modificadora

  const mods = [];
  if (e.ctrlKey) mods.push("ctrl");
  if (e.altKey) mods.push("alt");
  if (e.shiftKey) mods.push("shift");
  if (e.metaKey) mods.push("cmd");

  let key = e.key.toLowerCase();
  if (key === "escape") { recordingIndex = -1; render(); return; }
  if (!/^[a-z0-9]$/.test(key)) return; // solo 0-9 y a-z por ahora

  if (mods.length === 0) return; // exige al menos un modificador

  state.hotkeys[recordingIndex].modifiers = mods;
  state.hotkeys[recordingIndex].key = key;
  recordingIndex = -1;
  render();
}, true);

getState().then(s => {
  state = s;
  if (!state.hotkeys) state.hotkeys = [];
  render();
});
</script>
</body>
</html>`
