'use strict';

// ── API client ────────────────────────────────────────────────────────────

async function api(method, path, body) {
  const opts = { method, headers: {} };
  if (body !== undefined) {
    opts.headers['Content-Type'] = 'application/json';
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(path, opts);
  const ct = res.headers.get('content-type') || '';
  if (ct.includes('application/json')) {
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || res.statusText);
    return data;
  }
  const text = await res.text();
  if (!res.ok) throw new Error(text || res.statusText);
  return text;
}

// ── Toast notifications ───────────────────────────────────────────────────

function toast(msg, type = 'success') {
  const el = document.createElement('div');
  el.className = `toast toast-${type}`;
  el.textContent = msg;
  document.getElementById('toast-container').appendChild(el);
  setTimeout(() => el.remove(), 3200);
}

// ── DOM helpers ───────────────────────────────────────────────────────────

function el(tag, attrs = {}, ...children) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === 'class') node.className = v;
    else if (k.startsWith('on')) node.addEventListener(k.slice(2), v);
    else node.setAttribute(k, v);
  }
  for (const child of children) {
    if (child == null) continue;
    if (typeof child === 'string') node.appendChild(document.createTextNode(child));
    else node.appendChild(child);
  }
  return node;
}

function text(str) { return document.createTextNode(String(str ?? '')); }

function badge(method) {
  const b = document.createElement('span');
  b.className = `badge badge-${(method || '').toLowerCase()}`;
  b.textContent = (method || '').toUpperCase();
  return b;
}

function methodBadge(m) {
  const map = { GET: 'get', POST: 'post', PUT: 'put', DELETE: 'delete', PATCH: 'patch' };
  const b = document.createElement('span');
  b.className = `badge badge-${map[(m || '').toUpperCase()] || 'get'}`;
  b.textContent = (m || 'GET').toUpperCase();
  return b;
}

// ── Router ────────────────────────────────────────────────────────────────

const sections = [
  'dashboard', 'config', 'requests', 'environments',
  'memory', 'variables', 'history', 'api-graph', 'exports',
];

const renderers = {};

function showSection(name) {
  document.querySelectorAll('.nav-item').forEach(a => {
    a.classList.toggle('active', a.dataset.section === name);
  });
  const content = document.getElementById('content');
  content.innerHTML = '';
  const loader = el('div', { class: 'loading' }, 'Loading…');
  content.appendChild(loader);
  (renderers[name] || renderNotFound)(content);
}

function renderNotFound(container) {
  container.innerHTML = '';
  container.appendChild(el('div', { class: 'empty-state' }, 'Section not found.'));
}

// ── Dashboard ─────────────────────────────────────────────────────────────

renderers['dashboard'] = async function (container) {
  try {
    const data = await api('GET', '/api/dashboard');
    const { manifest, config_summary } = data;
    const counts = manifest?.counts || {};

    container.innerHTML = '';

    const header = el('div', { class: 'section-header' });
    const title = el('h1', { class: 'section-title' }, 'Dashboard');
    header.appendChild(title);
    container.appendChild(header);

    // Stats
    const statGrid = el('div', { class: 'stat-grid' });
    const statItems = [
      ['Requests', counts.requests ?? 0],
      ['Environments', counts.environments ?? 0],
      ['Baselines', counts.baselines ?? 0],
      ['Variables', counts.variables ?? 0],
    ];
    for (const [label, value] of statItems) {
      const card = el('div', { class: 'stat-card' });
      const val = el('div', { class: 'stat-value' });
      val.textContent = value;
      const lbl = el('div', { class: 'stat-label' });
      lbl.textContent = label;
      card.append(val, lbl);
      statGrid.appendChild(card);
    }
    container.appendChild(statGrid);

    // Config summary
    const cfgCard = el('div', { class: 'card' });
    const cfgTitle = el('div', { class: 'card-title' }, 'Active Configuration');
    cfgCard.appendChild(cfgTitle);
    const rows = [
      ['Provider', config_summary?.provider ?? '—'],
      ['Model', config_summary?.model ?? '—'],
      ['Framework', config_summary?.framework ?? '—'],
      ['Last Updated', manifest?.last_updated ? new Date(manifest.last_updated).toLocaleString() : '—'],
    ];
    const tbl = el('div', { class: 'table-wrapper' });
    const t = el('table');
    for (const [k, v] of rows) {
      const tr = el('tr');
      tr.appendChild(el('td', { class: 'muted' }, k));
      const td = el('td', { class: 'mono' });
      td.textContent = v;
      tr.appendChild(td);
      t.appendChild(tr);
    }
    tbl.appendChild(t);
    cfgCard.appendChild(tbl);
    container.appendChild(cfgCard);
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed to load dashboard: ' + e.message));
  }
};

// ── Config ────────────────────────────────────────────────────────────────

renderers['config'] = async function (container) {
  try {
    const cfg = await api('GET', '/api/config');
    container.innerHTML = '';

    const header = el('div', { class: 'section-header' });
    header.appendChild(el('h1', { class: 'section-title' }, 'Configuration'));
    container.appendChild(header);

    const form = el('form');
    form.addEventListener('submit', async e => {
      e.preventDefault();
      const updated = buildConfigFromForm(form, cfg);
      try {
        await api('PUT', '/api/config', updated);
        toast('Configuration saved');
      } catch (err) {
        toast('Save failed: ' + err.message, 'error');
      }
    });

    // Provider
    const provCard = el('div', { class: 'card' });
    provCard.appendChild(el('div', { class: 'card-title' }, 'LLM Provider'));
    const provRow = el('div', { class: 'form-row' });

    const provGroup = el('div', { class: 'form-group' });
    provGroup.appendChild(el('label', {}, 'Provider'));
    const provSel = el('select', { name: 'provider' },
      el('option', { value: 'ollama' }, 'Ollama'),
      el('option', { value: 'gemini' }, 'Gemini'),
    );
    provSel.value = cfg.provider || 'ollama';
    provGroup.appendChild(provSel);
    provRow.appendChild(provGroup);

    const modelGroup = el('div', { class: 'form-group' });
    modelGroup.appendChild(el('label', {}, 'Default Model'));
    const modelInput = el('input', { type: 'text', name: 'default_model', value: cfg.default_model || '' });
    modelGroup.appendChild(modelInput);
    provRow.appendChild(modelGroup);
    provCard.appendChild(provRow);

    // Ollama sub-config
    const ollamaCard = el('div', { class: 'card' });
    ollamaCard.appendChild(el('div', { class: 'card-title' }, 'Ollama Settings'));
    const ollamaRow = el('div', { class: 'form-row' });

    const ollMode = el('div', { class: 'form-group' });
    ollMode.appendChild(el('label', {}, 'Mode'));
    const ollModeSel = el('select', { name: 'ollama_mode' },
      el('option', { value: 'local' }, 'Local'),
      el('option', { value: 'cloud' }, 'Cloud'),
    );
    ollModeSel.value = cfg.ollama?.mode || 'local';
    ollMode.appendChild(ollModeSel);
    ollamaRow.appendChild(ollMode);

    const ollURL = el('div', { class: 'form-group' });
    ollURL.appendChild(el('label', {}, 'URL'));
    ollURL.appendChild(el('input', { type: 'text', name: 'ollama_url', value: cfg.ollama?.url || '' }));
    ollamaRow.appendChild(ollURL);

    const ollKey = el('div', { class: 'form-group' });
    ollKey.appendChild(el('label', {}, 'API Key'));
    ollKey.appendChild(el('input', { type: 'password', name: 'ollama_api_key', value: cfg.ollama?.api_key || '' }));
    ollamaCard.appendChild(ollamaRow);
    ollamaCard.appendChild(ollKey);

    // Gemini sub-config
    const geminiCard = el('div', { class: 'card' });
    geminiCard.appendChild(el('div', { class: 'card-title' }, 'Gemini Settings'));
    const gemKeyGroup = el('div', { class: 'form-group' });
    gemKeyGroup.appendChild(el('label', {}, 'API Key'));
    gemKeyGroup.appendChild(el('input', { type: 'password', name: 'gemini_api_key', value: cfg.gemini?.api_key || '' }));
    geminiCard.appendChild(gemKeyGroup);

    // Framework
    const fwCard = el('div', { class: 'card' });
    fwCard.appendChild(el('div', { class: 'card-title' }, 'Framework'));
    const fwGroup = el('div', { class: 'form-group' });
    fwGroup.appendChild(el('label', {}, 'API Framework'));
    const fwInput = el('input', { type: 'text', name: 'framework', value: cfg.framework || '' });
    fwGroup.appendChild(fwInput);
    fwCard.appendChild(fwGroup);

    // Tool Limits
    const limCard = el('div', { class: 'card' });
    limCard.appendChild(el('div', { class: 'card-title' }, 'Tool Limits'));

    const globalRow = el('div', { class: 'form-row' });
    const defLim = el('div', { class: 'form-group' });
    defLim.appendChild(el('label', {}, 'Default Limit'));
    defLim.appendChild(el('input', { type: 'number', name: 'limit_default', value: cfg.tool_limits?.default_limit ?? 50 }));
    globalRow.appendChild(defLim);

    const totLim = el('div', { class: 'form-group' });
    totLim.appendChild(el('label', {}, 'Total Limit'));
    totLim.appendChild(el('input', { type: 'number', name: 'limit_total', value: cfg.tool_limits?.total_limit ?? 200 }));
    globalRow.appendChild(totLim);
    limCard.appendChild(globalRow);

    const perTool = cfg.tool_limits?.per_tool || {};
    if (Object.keys(perTool).length > 0) {
      const perToolLabel = el('div', { class: 'card-title' }, 'Per-Tool Limits');
      limCard.appendChild(perToolLabel);
      const limGrid = el('div', { class: 'limits-grid' });
      for (const [toolName, limit] of Object.entries(perTool)) {
        const item = el('div', { class: 'limit-item' });
        const lbl = el('label', {});
        lbl.textContent = toolName;
        const inp = el('input', { type: 'number', 'data-tool': toolName, value: limit });
        item.append(lbl, inp);
        limGrid.appendChild(item);
      }
      limCard.appendChild(limGrid);
    }

    // Web UI config
    const webCard = el('div', { class: 'card' });
    webCard.appendChild(el('div', { class: 'card-title' }, 'Web UI'));
    const webRow = el('div', { class: 'form-row' });
    const webPort = el('div', { class: 'form-group' });
    webPort.appendChild(el('label', {}, 'Port (0 = random)'));
    webPort.appendChild(el('input', { type: 'number', name: 'webui_port', value: cfg.web_ui?.port ?? 0 }));
    webRow.appendChild(webPort);
    webCard.appendChild(webRow);

    const actions = el('div', { class: 'form-actions' });
    actions.appendChild(el('button', { type: 'submit', class: 'btn btn-primary' }, 'Save Configuration'));

    form.append(provCard, ollamaCard, geminiCard, fwCard, limCard, webCard, actions);
    container.appendChild(form);
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed to load config: ' + e.message));
  }
};

function buildConfigFromForm(form, original) {
  const v = name => form.querySelector(`[name="${name}"]`)?.value ?? '';
  const n = name => parseInt(form.querySelector(`[name="${name}"]`)?.value ?? '0', 10);

  const perTool = {};
  form.querySelectorAll('[data-tool]').forEach(inp => {
    const toolName = inp.getAttribute('data-tool');
    const val = parseInt(inp.value, 10);
    if (!isNaN(val)) perTool[toolName] = val;
  });

  return {
    ...original,
    provider: v('provider'),
    default_model: v('default_model'),
    framework: v('framework'),
    ollama: {
      mode: v('ollama_mode'),
      url: v('ollama_url'),
      api_key: v('ollama_api_key'),
    },
    gemini: {
      api_key: v('gemini_api_key'),
    },
    tool_limits: {
      default_limit: n('limit_default'),
      total_limit: n('limit_total'),
      per_tool: perTool,
    },
    web_ui: {
      enabled: original.web_ui?.enabled ?? true,
      port: n('webui_port'),
    },
  };
}

// ── Requests ──────────────────────────────────────────────────────────────

renderers['requests'] = async function (container) {
  let names, selected = null;

  async function load() {
    names = await api('GET', '/api/requests');
  }

  async function render() {
    container.innerHTML = '';
    const header = el('div', { class: 'section-header' });
    header.appendChild(el('h1', { class: 'section-title' }, 'Saved Requests'));
    const newBtn = el('button', { class: 'btn btn-primary btn-sm', onclick: () => openForm(null) }, '+ New');
    header.appendChild(newBtn);
    container.appendChild(header);

    const split = el('div', { class: 'split-layout' });

    const listEl = el('div', { class: 'split-list' });
    if (names.length === 0) {
      listEl.appendChild(el('div', { class: 'split-empty' }, 'No requests saved.'));
    } else {
      for (const name of names) {
        const item = el('div', { class: 'split-list-item' + (name === selected ? ' selected' : '') });
        item.setAttribute('data-name', name);
        item.textContent = name;
        item.addEventListener('click', () => { selected = name; openForm(name); });
        listEl.appendChild(item);
      }
    }

    const panel = el('div', { class: 'split-panel' });
    panel.appendChild(el('div', { class: 'split-empty' }, 'Select a request or create a new one.'));

    split.append(listEl, panel);
    container.appendChild(split);
  }

  function openForm(name) {
    const panel = container.querySelector('.split-panel');
    if (!panel) return;
    panel.innerHTML = '';

    // Update list selection highlight
    container.querySelectorAll('.split-list-item').forEach(i => {
      i.classList.toggle('selected', i.getAttribute('data-name') === name);
    });

    if (name) {
      api('GET', `/api/requests/${encodeURIComponent(name)}`)
        .then(req => renderRequestForm(panel, req, name))
        .catch(e => { panel.appendChild(el('div', { class: 'empty-state' }, e.message)); });
    } else {
      renderRequestForm(panel, { name: '', method: 'GET', url: '', headers: {}, body: null }, null);
    }
  }

  function renderRequestForm(panel, req, existingName) {
    const isNew = !existingName;

    const fg = (labelText, inputEl) => {
      const g = el('div', { class: 'form-group' });
      g.appendChild(el('label', {}, labelText));
      g.appendChild(inputEl);
      return g;
    };

    const nameInput = el('input', { type: 'text', name: 'req-name', value: req.name || existingName || '' });
    const methodSel = el('select', { name: 'req-method' },
      ...['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map(m => el('option', { value: m }, m))
    );
    methodSel.value = req.method || 'GET';

    const urlInput = el('input', { type: 'text', name: 'req-url', value: req.url || '' });

    // Headers KV
    const headersLabel = el('div', { class: 'card-title' }, 'Headers');
    const headersKV = kvEditor(req.headers || {}, 'header');

    // Body
    let bodyStr = '';
    if (req.body) {
      bodyStr = typeof req.body === 'string' ? req.body : JSON.stringify(req.body, null, 2);
    }
    const bodyTA = el('textarea', { name: 'req-body', rows: '6' });
    bodyTA.value = bodyStr;

    const topRow = el('div', { class: 'form-row' });
    topRow.append(fg('Method', methodSel), fg('URL', urlInput));

    const actions = el('div', { class: 'form-actions' });
    const saveBtn = el('button', { type: 'button', class: 'btn btn-primary' }, isNew ? 'Create' : 'Save');
    saveBtn.addEventListener('click', async () => {
      const name = nameInput.value.trim();
      if (!name) { toast('Name is required', 'error'); return; }
      const headers = collectKV(headersKV);
      let body = bodyTA.value.trim();
      if (body) {
        try { body = JSON.parse(body); } catch { /* keep as string */ }
      } else {
        body = null;
      }
      const payload = { name, method: methodSel.value, url: urlInput.value, headers, body };
      try {
        if (isNew) {
          await api('POST', '/api/requests', payload);
          toast('Request created');
        } else {
          await api('PUT', `/api/requests/${encodeURIComponent(existingName)}`, payload);
          toast('Request saved');
        }
        await load();
        selected = name;
        await render();
        openForm(name);
      } catch (e) { toast('Error: ' + e.message, 'error'); }
    });

    if (!isNew) {
      const delBtn = el('button', { type: 'button', class: 'btn btn-danger' }, 'Delete');
      delBtn.addEventListener('click', async () => {
        if (!confirm(`Delete request "${existingName}"?`)) return;
        try {
          await api('DELETE', `/api/requests/${encodeURIComponent(existingName)}`);
          toast('Request deleted');
          selected = null;
          await load();
          await render();
        } catch (e) { toast('Error: ' + e.message, 'error'); }
      });
      actions.append(saveBtn, delBtn);
    } else {
      actions.appendChild(saveBtn);
    }

    if (isNew) panel.appendChild(fg('Name', nameInput));
    panel.append(topRow, headersLabel, headersKV, fg('Body (JSON)', bodyTA), actions);
  }

  try {
    await load();
    await render();
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed to load requests: ' + e.message));
  }
};

// ── Environments ──────────────────────────────────────────────────────────

renderers['environments'] = async function (container) {
  let names, selectedEnv = null;

  async function load() { names = await api('GET', '/api/environments'); }

  async function render() {
    container.innerHTML = '';
    const header = el('div', { class: 'section-header' });
    header.appendChild(el('h1', { class: 'section-title' }, 'Environments'));
    const newBtn = el('button', { class: 'btn btn-primary btn-sm', onclick: () => openForm(null) }, '+ New');
    header.appendChild(newBtn);
    container.appendChild(header);

    const split = el('div', { class: 'split-layout' });
    const listEl = el('div', { class: 'split-list' });
    if (names.length === 0) {
      listEl.appendChild(el('div', { class: 'split-empty' }, 'No environments.'));
    } else {
      for (const name of names) {
        const item = el('div', { class: 'split-list-item' + (name === selectedEnv ? ' selected' : '') });
        item.setAttribute('data-name', name);
        item.textContent = name;
        item.addEventListener('click', () => { selectedEnv = name; openForm(name); });
        listEl.appendChild(item);
      }
    }

    const panel = el('div', { class: 'split-panel' });
    panel.appendChild(el('div', { class: 'split-empty' }, 'Select an environment or create a new one.'));
    split.append(listEl, panel);
    container.appendChild(split);
  }

  function openForm(name) {
    const panel = container.querySelector('.split-panel');
    if (!panel) return;
    panel.innerHTML = '';
    container.querySelectorAll('.split-list-item').forEach(i => {
      i.classList.toggle('selected', i.getAttribute('data-name') === name);
    });

    if (name) {
      api('GET', `/api/environments/${encodeURIComponent(name)}`)
        .then(vars => renderEnvForm(panel, vars, name))
        .catch(e => { panel.appendChild(el('div', { class: 'empty-state' }, e.message)); });
    } else {
      renderEnvForm(panel, {}, null);
    }
  }

  function renderEnvForm(panel, vars, existingName) {
    const isNew = !existingName;
    const nameInput = isNew ? el('input', { type: 'text', placeholder: 'Environment name' }) : null;
    if (nameInput) {
      const ng = el('div', { class: 'form-group' });
      ng.appendChild(el('label', {}, 'Name'));
      ng.appendChild(nameInput);
      panel.appendChild(ng);
    } else {
      const title = el('h2', {});
      title.textContent = existingName;
      title.style.cssText = 'font-size:16px;font-weight:600;margin-bottom:16px;';
      panel.appendChild(title);
    }

    const kvEl = kvEditor(vars, 'env');
    panel.appendChild(kvEl);

    const addBtn = el('button', { type: 'button', class: 'btn btn-secondary btn-sm' }, '+ Add Variable');
    addBtn.addEventListener('click', () => addKVRow(kvEl));
    panel.appendChild(addBtn);

    const actions = el('div', { class: 'form-actions', style: 'margin-top:12px;' });
    const saveBtn = el('button', { type: 'button', class: 'btn btn-primary' }, isNew ? 'Create' : 'Save');
    saveBtn.addEventListener('click', async () => {
      const envName = isNew ? nameInput.value.trim() : existingName;
      if (!envName) { toast('Name is required', 'error'); return; }
      const collected = collectKV(kvEl);
      try {
        if (isNew) {
          await api('POST', '/api/environments', { name: envName, vars: collected });
          toast('Environment created');
        } else {
          await api('PUT', `/api/environments/${encodeURIComponent(envName)}`, collected);
          toast('Environment saved');
        }
        await load();
        selectedEnv = envName;
        await render();
        openForm(envName);
      } catch (e) { toast('Error: ' + e.message, 'error'); }
    });

    if (!isNew) {
      const delBtn = el('button', { type: 'button', class: 'btn btn-danger' }, 'Delete');
      delBtn.addEventListener('click', async () => {
        if (!confirm(`Delete environment "${existingName}"?`)) return;
        try {
          await api('DELETE', `/api/environments/${encodeURIComponent(existingName)}`);
          toast('Environment deleted');
          selectedEnv = null;
          await load();
          await render();
        } catch (e) { toast('Error: ' + e.message, 'error'); }
      });
      actions.append(saveBtn, delBtn);
    } else {
      actions.appendChild(saveBtn);
    }
    panel.appendChild(actions);
  }

  try {
    await load();
    await render();
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed to load environments: ' + e.message));
  }
};

// ── Memory ────────────────────────────────────────────────────────────────

renderers['memory'] = async function (container) {
  async function render() {
    let entries;
    try { entries = await api('GET', '/api/memory'); } catch (e) {
      container.innerHTML = '';
      container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
      return;
    }

    container.innerHTML = '';
    const header = el('div', { class: 'section-header' });
    header.appendChild(el('h1', { class: 'section-title' }, 'Agent Memory'));
    const addBtn = el('button', { class: 'btn btn-primary btn-sm', onclick: () => openAddForm() }, '+ Add');
    header.appendChild(addBtn);
    container.appendChild(header);

    if (!entries || entries.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, 'No memory entries yet.'));
      return;
    }

    const wrapper = el('div', { class: 'table-wrapper' });
    const t = el('table');
    const thead = el('tr');
    for (const h of ['Key', 'Value', 'Category', 'Timestamp', 'Source', '']) {
      thead.appendChild(el('th', {}, h));
    }
    t.appendChild(el('thead').appendChild(thead) && el('thead', {}, thead));

    const tbody = el('tbody');
    for (const entry of entries) {
      const tr = renderMemoryRow(entry, render);
      tbody.appendChild(tr);
    }
    t.appendChild(tbody);
    wrapper.appendChild(t);
    container.appendChild(wrapper);
  }

  function renderMemoryRow(entry, onUpdate) {
    const tr = el('tr');
    let editing = false;

    const keyTd   = el('td', { class: 'mono' }); keyTd.textContent = entry.key;
    const valTd   = el('td');                     valTd.textContent = entry.value;
    const catTd   = el('td', { class: 'muted' }); catTd.textContent = entry.category;
    const tsTd    = el('td', { class: 'muted' }); tsTd.textContent = entry.timestamp ? new Date(entry.timestamp).toLocaleString() : '';
    const srcTd   = el('td', { class: 'muted' }); srcTd.textContent = entry.source;
    const actTd   = el('td');

    function setView() {
      valTd.textContent = entry.value;
      catTd.textContent = entry.category;
      actTd.innerHTML = '';
      const editBtn = el('button', { class: 'btn btn-secondary btn-sm' }, 'Edit');
      const delBtn  = el('button', { class: 'btn btn-danger btn-sm' }, 'Del');
      editBtn.addEventListener('click', setEdit);
      delBtn.addEventListener('click', async () => {
        if (!confirm(`Delete memory key "${entry.key}"?`)) return;
        try {
          await api('DELETE', `/api/memory/${encodeURIComponent(entry.key)}`);
          toast('Entry deleted');
          onUpdate();
        } catch (e) { toast('Error: ' + e.message, 'error'); }
      });
      actTd.append(editBtn, ' ', delBtn);
    }

    function setEdit() {
      editing = true;
      const valInput = el('input', { type: 'text', value: entry.value });
      const catSel = el('select',  {},
        ...['preference','endpoint','error','project','general'].map(c => el('option', { value: c }, c))
      );
      catSel.value = entry.category;
      valTd.innerHTML = ''; valTd.appendChild(valInput);
      catTd.innerHTML = ''; catTd.appendChild(catSel);
      actTd.innerHTML = '';
      const saveBtn = el('button', { class: 'btn btn-primary btn-sm' }, 'Save');
      const cancelBtn = el('button', { class: 'btn btn-secondary btn-sm' }, 'Cancel');
      saveBtn.addEventListener('click', async () => {
        try {
          await api('PUT', `/api/memory/${encodeURIComponent(entry.key)}`, {
            value: valInput.value, category: catSel.value,
          });
          entry.value = valInput.value;
          entry.category = catSel.value;
          toast('Entry saved');
          setView();
        } catch (e) { toast('Error: ' + e.message, 'error'); }
      });
      cancelBtn.addEventListener('click', setView);
      actTd.append(saveBtn, ' ', cancelBtn);
    }

    tr.append(keyTd, valTd, catTd, tsTd, srcTd, actTd);
    setView();
    return tr;
  }

  function openAddForm() {
    const existing = container.querySelector('.add-memory-form');
    if (existing) { existing.remove(); return; }
    const form = el('div', { class: 'card add-memory-form' });
    form.appendChild(el('div', { class: 'card-title' }, 'Add Memory Entry'));
    const row = el('div', { class: 'form-row' });
    const keyInput = el('input', { type: 'text', placeholder: 'Key' });
    const valInput = el('input', { type: 'text', placeholder: 'Value' });
    const catSel = el('select', {},
      ...['preference','endpoint','error','project','general'].map(c => el('option', { value: c }, c))
    );
    const kg = el('div', { class: 'form-group' }); kg.appendChild(el('label', {}, 'Key')); kg.appendChild(keyInput);
    const vg = el('div', { class: 'form-group' }); vg.appendChild(el('label', {}, 'Value')); vg.appendChild(valInput);
    const cg = el('div', { class: 'form-group' }); cg.appendChild(el('label', {}, 'Category')); cg.appendChild(catSel);
    row.append(kg, vg, cg);
    form.appendChild(row);
    const saveBtn = el('button', { class: 'btn btn-primary', style: 'margin-top:8px;' }, 'Add');
    saveBtn.addEventListener('click', async () => {
      const key = keyInput.value.trim();
      if (!key) { toast('Key is required', 'error'); return; }
      try {
        await api('PUT', `/api/memory/${encodeURIComponent(key)}`, { value: valInput.value, category: catSel.value });
        toast('Memory entry added');
        form.remove();
        render();
      } catch (e) { toast('Error: ' + e.message, 'error'); }
    });
    form.appendChild(saveBtn);
    container.insertBefore(form, container.querySelector('.table-wrapper') || null);
  }

  render();
};

// ── Variables ─────────────────────────────────────────────────────────────

renderers['variables'] = async function (container) {
  async function render() {
    let vars;
    try { vars = await api('GET', '/api/variables'); } catch (e) {
      container.innerHTML = '';
      container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
      return;
    }

    container.innerHTML = '';
    const header = el('div', { class: 'section-header' });
    header.appendChild(el('h1', { class: 'section-title' }, 'Global Variables'));
    const addBtn = el('button', { class: 'btn btn-primary btn-sm' }, '+ Add');
    addBtn.addEventListener('click', () => openAddForm(vars, render));
    header.appendChild(addBtn);
    container.appendChild(header);

    const entries = Object.entries(vars);
    if (entries.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, 'No global variables.'));
      return;
    }

    const wrapper = el('div', { class: 'table-wrapper' });
    const t = el('table');
    const thead = el('thead'); const htr = el('tr');
    for (const h of ['Name', 'Value', '']) htr.appendChild(el('th', {}, h));
    thead.appendChild(htr); t.appendChild(thead);

    const tbody = el('tbody');
    for (const [name, value] of entries) {
      const tr = el('tr');
      const nameTd = el('td', { class: 'mono' }); nameTd.textContent = name;
      const valTd  = el('td');
      const actTd  = el('td');

      let curVal = value;

      function setView() {
        valTd.innerHTML = ''; const span = el('span', { class: 'mono' }); span.textContent = curVal; valTd.appendChild(span);
        actTd.innerHTML = '';
        const editBtn = el('button', { class: 'btn btn-secondary btn-sm' }, 'Edit');
        const delBtn  = el('button', { class: 'btn btn-danger btn-sm' }, 'Del');
        editBtn.addEventListener('click', setEdit);
        delBtn.addEventListener('click', async () => {
          if (!confirm(`Delete variable "${name}"?`)) return;
          try { await api('DELETE', `/api/variables/${encodeURIComponent(name)}`); toast('Variable deleted'); render(); }
          catch (e) { toast('Error: ' + e.message, 'error'); }
        });
        actTd.append(editBtn, ' ', delBtn);
      }

      function setEdit() {
        const inp = el('input', { type: 'text', value: curVal });
        valTd.innerHTML = ''; valTd.appendChild(inp);
        actTd.innerHTML = '';
        const saveBtn   = el('button', { class: 'btn btn-primary btn-sm' }, 'Save');
        const cancelBtn = el('button', { class: 'btn btn-secondary btn-sm' }, 'Cancel');
        saveBtn.addEventListener('click', async () => {
          try { await api('PUT', `/api/variables/${encodeURIComponent(name)}`, { value: inp.value }); curVal = inp.value; toast('Saved'); setView(); }
          catch (e) { toast('Error: ' + e.message, 'error'); }
        });
        cancelBtn.addEventListener('click', setView);
        actTd.append(saveBtn, ' ', cancelBtn);
      }

      tr.append(nameTd, valTd, actTd);
      setView();
      tbody.appendChild(tr);
    }
    t.appendChild(tbody);
    wrapper.appendChild(t);
    container.appendChild(wrapper);
  }

  function openAddForm(vars, onDone) {
    const existing = container.querySelector('.add-var-form');
    if (existing) { existing.remove(); return; }
    const form = el('div', { class: 'card add-var-form' });
    form.appendChild(el('div', { class: 'card-title' }, 'Add Variable'));
    const row = el('div', { class: 'form-row' });
    const nameInput = el('input', { type: 'text', placeholder: 'NAME' });
    const valInput  = el('input', { type: 'text', placeholder: 'value' });
    const ng = el('div', { class: 'form-group' }); ng.appendChild(el('label', {}, 'Name')); ng.appendChild(nameInput);
    const vg = el('div', { class: 'form-group' }); vg.appendChild(el('label', {}, 'Value')); vg.appendChild(valInput);
    row.append(ng, vg);
    form.appendChild(row);
    const saveBtn = el('button', { class: 'btn btn-primary', style: 'margin-top:8px;' }, 'Add');
    saveBtn.addEventListener('click', async () => {
      const name = nameInput.value.trim();
      if (!name) { toast('Name is required', 'error'); return; }
      try { await api('PUT', `/api/variables/${encodeURIComponent(name)}`, { value: valInput.value }); toast('Variable added'); form.remove(); onDone(); }
      catch (e) { toast('Error: ' + e.message, 'error'); }
    });
    form.appendChild(saveBtn);
    const wrapper = container.querySelector('.table-wrapper');
    container.insertBefore(form, wrapper || null);
  }

  render();
};

// ── History ───────────────────────────────────────────────────────────────

renderers['history'] = async function (container) {
  try {
    const sessions = await api('GET', '/api/history');
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Session History')
    ));

    if (!sessions || sessions.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, 'No sessions recorded yet.'));
      return;
    }

    const wrapper = el('div', { class: 'table-wrapper' });
    const t = el('table');
    const thead = el('thead'); const htr = el('tr');
    for (const h of ['Session', 'Start', 'End', 'Turns', 'Tools Used', 'Summary']) htr.appendChild(el('th', {}, h));
    thead.appendChild(htr); t.appendChild(thead);

    const tbody = el('tbody');
    for (const s of [...sessions].reverse()) {
      const tr = el('tr');
      const sid = el('td', { class: 'mono muted' }); sid.textContent = s.session_id || '—';
      const start = el('td', { class: 'muted' }); start.textContent = s.start_time ? new Date(s.start_time).toLocaleString() : '—';
      const end   = el('td', { class: 'muted' }); end.textContent = s.end_time ? new Date(s.end_time).toLocaleString() : '—';
      const turns = el('td'); turns.textContent = s.turn_count ?? 0;
      const tools = el('td', { class: 'muted' }); tools.textContent = (s.tools_used || []).join(', ') || '—';
      const sum   = el('td'); sum.textContent = s.summary || '—';
      tr.append(sid, start, end, turns, tools, sum);
      tbody.appendChild(tr);
    }
    t.appendChild(tbody);
    wrapper.appendChild(t);
    container.appendChild(wrapper);
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
  }
};

// ── API Graph ─────────────────────────────────────────────────────────────

renderers['api-graph'] = async function (container) {
  try {
    const graph = await api('GET', '/api/api-graph');
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'API Graph')
    ));

    if (!graph || graph === 'null') {
      container.appendChild(el('div', { class: 'empty-state' }, 'No API graph found. Run the ingest_spec tool first.'));
      return;
    }

    const endpoints = graph.endpoints || {};
    if (Object.keys(endpoints).length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, 'No endpoints in graph.'));
      return;
    }

    // Context card
    if (graph.context) {
      const ctx = graph.context;
      const ctxCard = el('div', { class: 'card' });
      ctxCard.appendChild(el('div', { class: 'card-title' }, 'Project Context'));
      const rows = [['Framework', ctx.framework], ['Language', ctx.language], ['Spec Path', ctx.spec_path]];
      for (const [k, v] of rows) {
        if (!v) continue;
        const row = el('div', { style: 'display:flex;gap:12px;margin-bottom:4px;' });
        const lbl = el('span', { class: 'muted', style: 'width:100px;' }); lbl.textContent = k;
        const val = el('span', { class: 'mono' }); val.textContent = v;
        row.append(lbl, val);
        ctxCard.appendChild(row);
      }
      container.appendChild(ctxCard);
    }

    // Endpoints
    const epCard = el('div', { class: 'card' });
    epCard.appendChild(el('div', { class: 'card-title' }, `Endpoints (${Object.keys(endpoints).length})`));

    for (const [path, ep] of Object.entries(endpoints)) {
      const wrap = el('div', { style: 'border:1px solid var(--border);border-radius:var(--radius);margin-bottom:8px;overflow:hidden;' });
      const hdr = el('div', { class: 'collapsible-header' });
      const arrow = el('span', { class: 'collapsible-arrow' }, '▶');
      const pathSpan = el('span', { class: 'mono', style: 'flex:1;' }); pathSpan.textContent = path;
      const authBadge = ep.auth_type ? (() => { const b = el('span', { class: 'badge badge-post' }); b.textContent = ep.auth_type; return b; })() : null;
      hdr.append(arrow, pathSpan);
      if (authBadge) hdr.appendChild(authBadge);
      hdr.addEventListener('click', () => {
        hdr.classList.toggle('open');
        body.classList.toggle('open');
      });

      const body = el('div', { class: 'collapsible-body' });

      if (ep.summary) {
        const s = el('p', { style: 'margin-bottom:8px;color:var(--text-muted);font-size:13px;' }); s.textContent = ep.summary;
        body.appendChild(s);
      }

      if (ep.parameters && ep.parameters.length > 0) {
        body.appendChild(el('div', { class: 'card-title', style: 'margin-top:4px;' }, 'Parameters'));
        for (const p of ep.parameters) {
          const row = el('div', { class: 'mono', style: 'font-size:12px;margin-bottom:2px;' });
          row.textContent = `${p.name} (${p.type || 'any'})${p.required ? ' *' : ''} — ${p.description || ''}`;
          body.appendChild(row);
        }
      }

      if (ep.security && ep.security.length > 0) {
        body.appendChild(el('div', { class: 'card-title', style: 'margin-top:8px;' }, 'Security Risks'));
        for (const risk of ep.security) {
          const sev = (risk.severity || 'low').toLowerCase();
          const row = el('div', { style: 'display:flex;gap:8px;align-items:center;margin-bottom:4px;' });
          const badge2 = el('span', { class: `badge badge-${sev}` }); badge2.textContent = risk.severity;
          const desc = el('span', { style: 'font-size:12px;' }); desc.textContent = risk.risk || risk.description || '';
          row.append(badge2, desc);
          body.appendChild(row);
        }
      }

      wrap.append(hdr, body);
      epCard.appendChild(wrap);
    }
    container.appendChild(epCard);
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
  }
};

// ── Exports ───────────────────────────────────────────────────────────────

renderers['exports'] = async function (container) {
  let names, selectedExport = null;

  async function load() { names = await api('GET', '/api/exports'); }

  async function render() {
    container.innerHTML = '';
    const header = el('div', { class: 'section-header' });
    header.appendChild(el('h1', { class: 'section-title' }, 'Exports'));
    container.appendChild(header);

    const split = el('div', { class: 'split-layout' });
    const listEl = el('div', { class: 'split-list' });
    if (!names || names.length === 0) {
      listEl.appendChild(el('div', { class: 'split-empty' }, 'No exports yet.'));
    } else {
      for (const name of names) {
        const item = el('div', { class: 'split-list-item' + (name === selectedExport ? ' selected' : '') });
        item.setAttribute('data-name', name);
        item.textContent = name;
        item.addEventListener('click', () => { selectedExport = name; openFile(name); });
        listEl.appendChild(item);
      }
    }

    const panel = el('div', { class: 'split-panel' });
    panel.appendChild(el('div', { class: 'split-empty' }, 'Select a file to view.'));
    split.append(listEl, panel);
    container.appendChild(split);
  }

  function openFile(name) {
    const panel = container.querySelector('.split-panel');
    if (!panel) return;
    panel.innerHTML = '';
    container.querySelectorAll('.split-list-item').forEach(i => {
      i.classList.toggle('selected', i.getAttribute('data-name') === name);
    });

    fetch(`/api/exports/${encodeURIComponent(name)}`).then(async res => {
      const text = await res.text();
      panel.innerHTML = '';
      const title = el('h2', {}); title.textContent = name;
      title.style.cssText = 'font-size:14px;font-weight:600;margin-bottom:12px;font-family:monospace;';
      panel.appendChild(title);
      const pre = el('pre'); pre.textContent = text;
      panel.appendChild(pre);
    }).catch(e => {
      panel.appendChild(el('div', { class: 'empty-state' }, e.message));
    });
  }

  try {
    await load();
    await render();
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
  }
};

// ── KV editor helpers ─────────────────────────────────────────────────────

function kvEditor(initialVars, prefix) {
  const table = el('table', { class: 'kv-table' });
  const tbody = el('tbody');
  table.appendChild(tbody);

  for (const [k, v] of Object.entries(initialVars)) {
    tbody.appendChild(kvRow(k, v, prefix));
  }
  return table;
}

function kvRow(k, v, prefix) {
  const tr = el('tr');
  const keyTd = el('td', { class: 'kv-key-col' });
  const valTd = el('td', { class: 'kv-val-col' });
  const actTd = el('td', { class: 'kv-act-col' });

  const keyInput = el('input', { type: 'text', 'data-kv-key': '1' });
  keyInput.value = k;
  const valInput = el('input', { type: 'text', 'data-kv-val': '1' });
  valInput.value = v;

  const delBtn = el('button', { type: 'button', class: 'btn btn-danger btn-sm' }, '×');
  delBtn.addEventListener('click', () => tr.remove());

  keyTd.appendChild(keyInput);
  valTd.appendChild(valInput);
  actTd.appendChild(delBtn);
  tr.append(keyTd, valTd, actTd);
  return tr;
}

function addKVRow(table) {
  const tbody = table.querySelector('tbody') || table;
  tbody.appendChild(kvRow('', '', 'kv'));
}

function collectKV(table) {
  const result = {};
  const keys = table.querySelectorAll('[data-kv-key]');
  const vals = table.querySelectorAll('[data-kv-val]');
  keys.forEach((k, i) => {
    const key = k.value.trim();
    if (key) result[key] = vals[i]?.value ?? '';
  });
  return result;
}

// ── Init ──────────────────────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', () => {
  document.querySelectorAll('.nav-item').forEach(a => {
    a.addEventListener('click', e => {
      e.preventDefault();
      showSection(a.dataset.section);
    });
  });
  showSection('dashboard');
});
