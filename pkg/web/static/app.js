'use strict';

// ── Theme ──────────────────────────────────────────────────────────────────

(function initTheme() {
  const saved = localStorage.getItem('falcon-theme') || 'dark';
  document.documentElement.setAttribute('data-theme', saved);
})();

document.getElementById('theme-btn').addEventListener('click', () => {
  const cur = document.documentElement.getAttribute('data-theme');
  const next = cur === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', next);
  localStorage.setItem('falcon-theme', next);
});

// ── API client ─────────────────────────────────────────────────────────────

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

// ── Toast ──────────────────────────────────────────────────────────────────

function toast(msg, type = 'success') {
  const t = document.createElement('div');
  t.className = `toast toast-${type}`;
  t.textContent = msg;
  document.getElementById('toast-container').appendChild(t);
  setTimeout(() => t.remove(), 3200);
}

// ── DOM helpers ────────────────────────────────────────────────────────────

function el(tag, attrs = {}, ...children) {
  const node = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === 'class') node.className = v;
    else if (k === 'style') node.style.cssText = v;
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

function methodBadge(m) {
  const map = { GET: 'get', POST: 'post', PUT: 'put', DELETE: 'delete', PATCH: 'patch' };
  const b = document.createElement('span');
  b.className = `badge badge-${map[(m || '').toUpperCase()] || 'get'}`;
  b.textContent = (m || 'GET').toUpperCase();
  return b;
}

function severityBadge(sev) {
  const s = (sev || 'low').toLowerCase();
  const b = document.createElement('span');
  b.className = `badge badge-${s}`;
  b.textContent = sev || 'low';
  return b;
}

// ── Router ─────────────────────────────────────────────────────────────────

const renderers = {};

function showSection(name) {
  document.querySelectorAll('.nav-item').forEach(a => {
    a.classList.toggle('active', a.dataset.section === name);
  });
  const content = document.getElementById('content');
  // Animate out, swap, animate in
  content.style.opacity = '0';
  content.style.transform = 'translateY(4px)';
  setTimeout(() => {
    content.innerHTML = '';
    content.appendChild(el('div', { class: 'loading' }, 'Loading'));
    content.style.transition = 'opacity 0.18s ease, transform 0.18s ease';
    content.style.opacity = '1';
    content.style.transform = 'translateY(0)';
    (renderers[name] || renderNotFound)(content);
  }, 80);
}

function renderNotFound(container) {
  container.innerHTML = '';
  container.appendChild(el('div', { class: 'empty-state' }, '404 — Section not found'));
}

// ── Dashboard ──────────────────────────────────────────────────────────────

renderers['dashboard'] = async function(container) {
  try {
    const data = await api('GET', '/api/dashboard');
    const { manifest, config_summary } = data;
    const counts = manifest?.counts || {};

    container.innerHTML = '';

    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Dashboard')
    ));

    // Stats
    const statGrid = el('div', { class: 'stat-grid' });
    for (const [label, value] of [
      ['Requests',     counts.requests ?? 0],
      ['Environments', counts.environments ?? 0],
      ['Baselines',    counts.baselines ?? 0],
      ['Variables',    counts.variables ?? 0],
    ]) {
      statGrid.appendChild(
        el('div', { class: 'stat-card' },
          el('div', { class: 'stat-value' }, String(value)),
          el('div', { class: 'stat-label' }, label)
        )
      );
    }
    container.appendChild(statGrid);

    // Config summary
    const cfgCard = el('div', { class: 'card' });
    cfgCard.appendChild(el('div', { class: 'card-title' }, 'Active Configuration'));
    const t = el('table');
    for (const [k, v] of [
      ['Provider',     config_summary?.provider ?? '—'],
      ['Model',        config_summary?.model ?? '—'],
      ['Framework',    config_summary?.framework ?? '—'],
      ['Last Updated', manifest?.last_updated ? new Date(manifest.last_updated).toLocaleString() : '—'],
    ]) {
      t.appendChild(el('tr',  {},
        el('td', { class: 'muted', style: 'width:140px' }, k),
        el('td', { class: 'mono' }, v)
      ));
    }
    cfgCard.appendChild(el('div', { class: 'table-wrapper' }, t));
    container.appendChild(cfgCard);
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed to load dashboard: ' + e.message));
  }
};

// ── Config ─────────────────────────────────────────────────────────────────

renderers['config'] = async function(container) {
  try {
    const cfg = await api('GET', '/api/config');
    container.innerHTML = '';

    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Configuration')
    ));

    const form = el('form');
    form.addEventListener('submit', async e => {
      e.preventDefault();
      try {
        await api('PUT', '/api/config', buildConfigFromForm(form, cfg));
        toast('Configuration saved');
      } catch (err) {
        toast('Save failed: ' + err.message, 'error');
      }
    });

    function fg(labelText, inputEl) {
      return el('div', { class: 'form-group' }, el('label', {}, labelText), inputEl);
    }

    // Provider card
    const provSel = el('select', { name: 'provider' },
      el('option', { value: 'ollama' }, 'Ollama'),
      el('option', { value: 'gemini' }, 'Gemini'),
    );
    provSel.value = cfg.provider || 'ollama';
    const modelInput = el('input', { type: 'text', name: 'default_model', value: cfg.default_model || '' });
    const provCard = el('div', { class: 'card' },
      el('div', { class: 'card-title' }, 'LLM Provider'),
      el('div', { class: 'form-row' }, fg('Provider', provSel), fg('Default Model', modelInput))
    );

    // Ollama card
    const ollModeSel = el('select', { name: 'ollama_mode' },
      el('option', { value: 'local' }, 'Local'),
      el('option', { value: 'cloud' }, 'Cloud'),
    );
    ollModeSel.value = cfg.ollama?.mode || 'local';
    const ollamaCard = el('div', { class: 'card' },
      el('div', { class: 'card-title' }, 'Ollama Settings'),
      el('div', { class: 'form-row' },
        fg('Mode', ollModeSel),
        fg('URL', el('input', { type: 'text', name: 'ollama_url', value: cfg.ollama?.url || '' }))
      ),
      fg('API Key', el('input', { type: 'password', name: 'ollama_api_key', value: cfg.ollama?.api_key || '' }))
    );

    // Gemini card
    const geminiCard = el('div', { class: 'card' },
      el('div', { class: 'card-title' }, 'Gemini Settings'),
      fg('API Key', el('input', { type: 'password', name: 'gemini_api_key', value: cfg.gemini?.api_key || '' }))
    );

    // Framework
    const fwCard = el('div', { class: 'card' },
      el('div', { class: 'card-title' }, 'Project'),
      fg('API Framework', el('input', { type: 'text', name: 'framework', value: cfg.framework || '' }))
    );

    // Tool limits
    const limCard = el('div', { class: 'card' });
    limCard.appendChild(el('div', { class: 'card-title' }, 'Tool Limits'));
    limCard.appendChild(
      el('div', { class: 'form-row' },
        fg('Default Limit', el('input', { type: 'number', name: 'limit_default', value: cfg.tool_limits?.default_limit ?? 50 })),
        fg('Total Limit',   el('input', { type: 'number', name: 'limit_total',   value: cfg.tool_limits?.total_limit ?? 200 }))
      )
    );
    const perTool = cfg.tool_limits?.per_tool || {};
    if (Object.keys(perTool).length > 0) {
      limCard.appendChild(el('div', { class: 'divider' }));
      limCard.appendChild(el('div', { class: 'card-title' }, 'Per-Tool Limits'));
      const limGrid = el('div', { class: 'limits-grid' });
      for (const [toolName, limit] of Object.entries(perTool)) {
        const lbl = el('label', {}); lbl.textContent = toolName;
        limGrid.appendChild(
          el('div', { class: 'limit-item' },
            lbl,
            el('input', { type: 'number', 'data-tool': toolName, value: limit })
          )
        );
      }
      limCard.appendChild(limGrid);
    }

    // Web UI
    const webCard = el('div', { class: 'card' },
      el('div', { class: 'card-title' }, 'Web UI'),
      fg('Port (0 = random)', el('input', { type: 'number', name: 'webui_port', value: cfg.web_ui?.port ?? 0 }))
    );

    const actions = el('div', { class: 'form-actions' },
      el('button', { type: 'submit', class: 'btn btn-primary' }, 'Save Configuration')
    );

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
    ollama: { mode: v('ollama_mode'), url: v('ollama_url'), api_key: v('ollama_api_key') },
    gemini: { api_key: v('gemini_api_key') },
    tool_limits: { default_limit: n('limit_default'), total_limit: n('limit_total'), per_tool: perTool },
    web_ui: { enabled: original.web_ui?.enabled ?? true, port: n('webui_port') },
  };
}

// ── Requests ───────────────────────────────────────────────────────────────

renderers['requests'] = async function(container) {
  let names, selected = null;

  async function load() { names = await api('GET', '/api/requests'); }

  async function render() {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Saved Requests'),
      el('button', { class: 'btn btn-primary btn-sm', onclick: () => openForm(null) }, '+ New')
    ));

    const split = el('div', { class: 'split-layout' });
    const listEl = el('div', { class: 'split-list' });

    if (names.length === 0) {
      listEl.appendChild(el('div', { class: 'split-empty' }, 'No requests saved'));
    } else {
      for (const name of names) {
        const item = el('div', {
          class: 'split-list-item' + (name === selected ? ' selected' : ''),
          'data-name': name,
        });
        item.textContent = name;
        item.addEventListener('click', () => { selected = name; openForm(name); });
        listEl.appendChild(item);
      }
    }

    const panel = el('div', { class: 'split-panel' },
      el('div', { class: 'split-empty' }, 'Select a request or create a new one')
    );
    split.append(listEl, panel);
    container.appendChild(split);
  }

  function openForm(name) {
    const panel = container.querySelector('.split-panel');
    if (!panel) return;
    panel.innerHTML = '';
    container.querySelectorAll('.split-list-item').forEach(i =>
      i.classList.toggle('selected', i.getAttribute('data-name') === name)
    );
    if (name) {
      api('GET', `/api/requests/${encodeURIComponent(name)}`)
        .then(req => renderRequestForm(panel, req, name))
        .catch(e => panel.appendChild(el('div', { class: 'empty-state' }, e.message)));
    } else {
      renderRequestForm(panel, { name: '', method: 'GET', url: '', headers: {}, body: null }, null);
    }
  }

  function renderRequestForm(panel, req, existingName) {
    const isNew = !existingName;

    function fg(lbl, inp) { return el('div', { class: 'form-group' }, el('label', {}, lbl), inp); }

    const nameInput  = el('input', { type: 'text', name: 'req-name', value: req.name || existingName || '' });
    const methodSel  = el('select', { name: 'req-method' },
      ...['GET','POST','PUT','PATCH','DELETE'].map(m => el('option', { value: m }, m))
    );
    methodSel.value = req.method || 'GET';
    const urlInput   = el('input', { type: 'text', name: 'req-url', value: req.url || '' });
    const headersKV  = kvEditor(req.headers || {}, 'header');
    let bodyStr = req.body ? (typeof req.body === 'string' ? req.body : JSON.stringify(req.body, null, 2)) : '';
    const bodyTA = el('textarea', { name: 'req-body', rows: '6' });
    bodyTA.value = bodyStr;

    const saveBtn = el('button', { type: 'button', class: 'btn btn-primary' }, isNew ? 'Create' : 'Save');
    saveBtn.addEventListener('click', async () => {
      const name = nameInput.value.trim();
      if (!name) { toast('Name is required', 'error'); return; }
      let body = bodyTA.value.trim();
      if (body) { try { body = JSON.parse(body); } catch { /* keep string */ } } else { body = null; }
      const payload = { name, method: methodSel.value, url: urlInput.value, headers: collectKV(headersKV), body };
      try {
        if (isNew) { await api('POST', '/api/requests', payload); toast('Request created'); }
        else { await api('PUT', `/api/requests/${encodeURIComponent(existingName)}`, payload); toast('Request saved'); }
        await load(); selected = name; await render(); openForm(name);
      } catch (e) { toast('Error: ' + e.message, 'error'); }
    });

    const actions = el('div', { class: 'form-actions' }, saveBtn);
    if (!isNew) {
      const delBtn = el('button', { type: 'button', class: 'btn btn-danger' }, 'Delete');
      delBtn.addEventListener('click', async () => {
        if (!confirm(`Delete "${existingName}"?`)) return;
        try {
          await api('DELETE', `/api/requests/${encodeURIComponent(existingName)}`);
          toast('Request deleted'); selected = null; await load(); await render();
        } catch (e) { toast('Error: ' + e.message, 'error'); }
      });
      actions.appendChild(delBtn);
    }

    const addHdrBtn = el('button', { type: 'button', class: 'btn btn-secondary btn-sm', style: 'margin-bottom:12px' }, '+ Header');
    addHdrBtn.addEventListener('click', () => addKVRow(headersKV));

    if (isNew) panel.appendChild(fg('Name', nameInput));
    panel.append(
      el('div', { class: 'form-row' }, fg('Method', methodSel), fg('URL', urlInput)),
      el('div', { class: 'card-title', style: 'margin-top:4px' }, 'Headers'),
      headersKV, addHdrBtn,
      fg('Body (JSON)', bodyTA),
      actions
    );
  }

  try { await load(); await render(); }
  catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
  }
};

// ── Environments ───────────────────────────────────────────────────────────

renderers['environments'] = async function(container) {
  let names, selectedEnv = null;

  async function load() { names = await api('GET', '/api/environments'); }

  async function render() {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Environments'),
      el('button', { class: 'btn btn-primary btn-sm', onclick: () => openForm(null) }, '+ New')
    ));

    const split = el('div', { class: 'split-layout' });
    const listEl = el('div', { class: 'split-list' });

    if (names.length === 0) {
      listEl.appendChild(el('div', { class: 'split-empty' }, 'No environments'));
    } else {
      for (const name of names) {
        const item = el('div', {
          class: 'split-list-item' + (name === selectedEnv ? ' selected' : ''),
          'data-name': name,
        });
        item.textContent = name;
        item.addEventListener('click', () => { selectedEnv = name; openForm(name); });
        listEl.appendChild(item);
      }
    }

    const panel = el('div', { class: 'split-panel' },
      el('div', { class: 'split-empty' }, 'Select an environment or create a new one')
    );
    split.append(listEl, panel);
    container.appendChild(split);
  }

  function openForm(name) {
    const panel = container.querySelector('.split-panel');
    if (!panel) return;
    panel.innerHTML = '';
    container.querySelectorAll('.split-list-item').forEach(i =>
      i.classList.toggle('selected', i.getAttribute('data-name') === name)
    );
    if (name) {
      api('GET', `/api/environments/${encodeURIComponent(name)}`)
        .then(vars => renderEnvForm(panel, vars, name))
        .catch(e => panel.appendChild(el('div', { class: 'empty-state' }, e.message)));
    } else {
      renderEnvForm(panel, {}, null);
    }
  }

  function renderEnvForm(panel, vars, existingName) {
    const isNew = !existingName;
    const nameInput = isNew ? el('input', { type: 'text', placeholder: 'environment-name' }) : null;

    if (isNew) {
      const ng = el('div', { class: 'form-group' });
      ng.appendChild(el('label', {}, 'Name'));
      ng.appendChild(nameInput);
      panel.appendChild(ng);
    } else {
      panel.appendChild(el('h2', { style: 'font-size:15px;font-weight:600;margin-bottom:16px;' }, existingName));
    }

    const kvEl = kvEditor(vars, 'env');
    panel.appendChild(kvEl);

    const addBtn = el('button', { type: 'button', class: 'btn btn-secondary btn-sm', style: 'margin-top:8px' }, '+ Add Variable');
    addBtn.addEventListener('click', () => addKVRow(kvEl));
    panel.appendChild(addBtn);

    const saveBtn = el('button', { type: 'button', class: 'btn btn-primary' }, isNew ? 'Create' : 'Save');
    saveBtn.addEventListener('click', async () => {
      const envName = isNew ? nameInput.value.trim() : existingName;
      if (!envName) { toast('Name is required', 'error'); return; }
      try {
        if (isNew) { await api('POST', '/api/environments', { name: envName, vars: collectKV(kvEl) }); toast('Environment created'); }
        else { await api('PUT', `/api/environments/${encodeURIComponent(envName)}`, collectKV(kvEl)); toast('Environment saved'); }
        await load(); selectedEnv = envName; await render(); openForm(envName);
      } catch (e) { toast('Error: ' + e.message, 'error'); }
    });

    const actions = el('div', { class: 'form-actions', style: 'margin-top:14px' }, saveBtn);
    if (!isNew) {
      const delBtn = el('button', { type: 'button', class: 'btn btn-danger' }, 'Delete');
      delBtn.addEventListener('click', async () => {
        if (!confirm(`Delete "${existingName}"?`)) return;
        try {
          await api('DELETE', `/api/environments/${encodeURIComponent(existingName)}`);
          toast('Environment deleted'); selectedEnv = null; await load(); await render();
        } catch (e) { toast('Error: ' + e.message, 'error'); }
      });
      actions.appendChild(delBtn);
    }
    panel.appendChild(actions);
  }

  try { await load(); await render(); }
  catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
  }
};

// ── Memory ─────────────────────────────────────────────────────────────────

renderers['memory'] = async function(container) {
  async function render() {
    let entries;
    try { entries = await api('GET', '/api/memory'); }
    catch (e) {
      container.innerHTML = '';
      container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
      return;
    }

    container.innerHTML = '';
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Agent Memory'),
      el('button', { class: 'btn btn-primary btn-sm', onclick: openAddForm }, '+ Add')
    ));

    if (!entries || entries.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, 'No memory entries yet'));
      return;
    }

    const thead = el('thead', {},
      el('tr', {},
        ...['Key', 'Value', 'Category', 'Timestamp', 'Source', ''].map(h => el('th', {}, h))
      )
    );
    const tbody = el('tbody');
    for (const entry of entries) tbody.appendChild(renderMemoryRow(entry, render));
    container.appendChild(el('div', { class: 'table-wrapper' }, el('table', {}, thead, tbody)));
  }

  function renderMemoryRow(entry, onUpdate) {
    const tr = el('tr');
    const keyTd = el('td', { class: 'mono' }); keyTd.textContent = entry.key;
    const valTd = el('td');
    const catTd = el('td', { class: 'muted' });
    const tsTd  = el('td', { class: 'muted' });
    const srcTd = el('td', { class: 'muted' });
    const actTd = el('td');

    tsTd.textContent  = entry.timestamp ? new Date(entry.timestamp).toLocaleString() : '';
    srcTd.textContent = entry.source || '';

    function setView() {
      valTd.textContent = entry.value;
      catTd.textContent = entry.category;
      actTd.innerHTML = '';
      const editBtn = el('button', { class: 'btn btn-secondary btn-xs' }, 'Edit');
      const delBtn  = el('button', { class: 'btn btn-danger btn-xs' }, 'Del');
      editBtn.addEventListener('click', setEdit);
      delBtn.addEventListener('click', async () => {
        if (!confirm(`Delete "${entry.key}"?`)) return;
        try { await api('DELETE', `/api/memory/${encodeURIComponent(entry.key)}`); toast('Entry deleted'); onUpdate(); }
        catch (e) { toast('Error: ' + e.message, 'error'); }
      });
      actTd.appendChild(el('div', { class: 'act-cell' }, editBtn, delBtn));
    }

    function setEdit() {
      const valInput = el('input', { type: 'text', value: entry.value });
      const catSel = el('select', {},
        ...['preference','endpoint','error','project','general'].map(c => el('option', { value: c }, c))
      );
      catSel.value = entry.category;
      valTd.innerHTML = ''; valTd.appendChild(valInput);
      catTd.innerHTML = ''; catTd.appendChild(catSel);
      actTd.innerHTML = '';
      const saveBtn   = el('button', { class: 'btn btn-primary btn-xs' }, 'Save');
      const cancelBtn = el('button', { class: 'btn btn-secondary btn-xs' }, 'Cancel');
      saveBtn.addEventListener('click', async () => {
        try {
          await api('PUT', `/api/memory/${encodeURIComponent(entry.key)}`, { value: valInput.value, category: catSel.value });
          entry.value = valInput.value; entry.category = catSel.value;
          toast('Saved'); setView();
        } catch (e) { toast('Error: ' + e.message, 'error'); }
      });
      cancelBtn.addEventListener('click', setView);
      actTd.appendChild(el('div', { class: 'act-cell' }, saveBtn, cancelBtn));
    }

    tr.append(keyTd, valTd, catTd, tsTd, srcTd, actTd);
    setView();
    return tr;
  }

  function openAddForm() {
    const existing = container.querySelector('.add-memory-form');
    if (existing) { existing.remove(); return; }
    const keyInput = el('input', { type: 'text', placeholder: 'key' });
    const valInput = el('input', { type: 'text', placeholder: 'value' });
    const catSel = el('select', {},
      ...['preference','endpoint','error','project','general'].map(c => el('option', { value: c }, c))
    );
    const saveBtn = el('button', { class: 'btn btn-primary', style: 'margin-top:10px' }, 'Add Entry');
    saveBtn.addEventListener('click', async () => {
      const key = keyInput.value.trim();
      if (!key) { toast('Key is required', 'error'); return; }
      try {
        await api('PUT', `/api/memory/${encodeURIComponent(key)}`, { value: valInput.value, category: catSel.value });
        toast('Memory entry added'); form.remove(); render();
      } catch (e) { toast('Error: ' + e.message, 'error'); }
    });

    const form = el('div', { class: 'card add-memory-form' },
      el('div', { class: 'card-title' }, 'New Memory Entry'),
      el('div', { class: 'form-row' },
        el('div', { class: 'form-group' }, el('label', {}, 'Key'), keyInput),
        el('div', { class: 'form-group' }, el('label', {}, 'Value'), valInput),
        el('div', { class: 'form-group' }, el('label', {}, 'Category'), catSel)
      ),
      saveBtn
    );
    container.insertBefore(form, container.querySelector('.table-wrapper') || null);
  }

  render();
};

// ── Variables ──────────────────────────────────────────────────────────────

renderers['variables'] = async function(container) {
  async function render() {
    let vars;
    try { vars = await api('GET', '/api/variables'); }
    catch (e) {
      container.innerHTML = '';
      container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
      return;
    }

    container.innerHTML = '';
    const addBtn = el('button', { class: 'btn btn-primary btn-sm' }, '+ Add');
    addBtn.addEventListener('click', () => openAddForm(vars, render));
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Global Variables'),
      addBtn
    ));

    const entries = Object.entries(vars);
    if (entries.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, 'No global variables'));
      return;
    }

    const thead = el('thead', {}, el('tr', {}, ...['Name', 'Value', ''].map(h => el('th', {}, h))));
    const tbody = el('tbody');

    for (const [name, value] of entries) {
      const nameTd = el('td', { class: 'mono' }); nameTd.textContent = name;
      const valTd  = el('td');
      const actTd  = el('td');
      let curVal = value;

      function setView() {
        valTd.innerHTML = ''; valTd.appendChild(el('span', { class: 'mono' }));
        valTd.firstChild.textContent = curVal;
        actTd.innerHTML = '';
        const editBtn = el('button', { class: 'btn btn-secondary btn-xs' }, 'Edit');
        const delBtn  = el('button', { class: 'btn btn-danger btn-xs' }, 'Del');
        editBtn.addEventListener('click', setEdit);
        delBtn.addEventListener('click', async () => {
          if (!confirm(`Delete "${name}"?`)) return;
          try { await api('DELETE', `/api/variables/${encodeURIComponent(name)}`); toast('Variable deleted'); render(); }
          catch (e) { toast('Error: ' + e.message, 'error'); }
        });
        actTd.appendChild(el('div', { class: 'act-cell' }, editBtn, delBtn));
      }

      function setEdit() {
        const inp = el('input', { type: 'text', value: curVal });
        valTd.innerHTML = ''; valTd.appendChild(inp);
        actTd.innerHTML = '';
        const saveBtn   = el('button', { class: 'btn btn-primary btn-xs' }, 'Save');
        const cancelBtn = el('button', { class: 'btn btn-secondary btn-xs' }, 'Cancel');
        saveBtn.addEventListener('click', async () => {
          try { await api('PUT', `/api/variables/${encodeURIComponent(name)}`, { value: inp.value }); curVal = inp.value; toast('Saved'); setView(); }
          catch (e) { toast('Error: ' + e.message, 'error'); }
        });
        cancelBtn.addEventListener('click', setView);
        actTd.appendChild(el('div', { class: 'act-cell' }, saveBtn, cancelBtn));
      }

      const tr = el('tr', {}, nameTd, valTd, actTd);
      setView();
      tbody.appendChild(tr);
    }
    container.appendChild(el('div', { class: 'table-wrapper' }, el('table', {}, thead, tbody)));
  }

  function openAddForm(vars, onDone) {
    const existing = container.querySelector('.add-var-form');
    if (existing) { existing.remove(); return; }
    const nameInput = el('input', { type: 'text', placeholder: 'VAR_NAME' });
    const valInput  = el('input', { type: 'text', placeholder: 'value' });
    const saveBtn = el('button', { class: 'btn btn-primary', style: 'margin-top:10px' }, 'Add Variable');
    saveBtn.addEventListener('click', async () => {
      const name = nameInput.value.trim();
      if (!name) { toast('Name is required', 'error'); return; }
      try { await api('PUT', `/api/variables/${encodeURIComponent(name)}`, { value: valInput.value }); toast('Variable added'); form.remove(); onDone(); }
      catch (e) { toast('Error: ' + e.message, 'error'); }
    });
    const form = el('div', { class: 'card add-var-form' },
      el('div', { class: 'card-title' }, 'New Variable'),
      el('div', { class: 'form-row' },
        el('div', { class: 'form-group' }, el('label', {}, 'Name'), nameInput),
        el('div', { class: 'form-group' }, el('label', {}, 'Value'), valInput)
      ),
      saveBtn
    );
    container.insertBefore(form, container.querySelector('.table-wrapper') || null);
  }

  render();
};

// ── History ────────────────────────────────────────────────────────────────

renderers['history'] = async function(container) {
  try {
    const sessions = await api('GET', '/api/history');
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Session History')
    ));

    if (!sessions || sessions.length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, 'No sessions recorded yet'));
      return;
    }

    const thead = el('thead', {},
      el('tr', {}, ...['Session', 'Start', 'End', 'Turns', 'Tools Used', 'Summary'].map(h => el('th', {}, h)))
    );
    const tbody = el('tbody');
    for (const s of [...sessions].reverse()) {
      const tr = el('tr');
      const sid = el('td', { class: 'mono muted' }); sid.textContent = (s.session_id || '—').slice(0, 8);
      const start = el('td', { class: 'muted' }); start.textContent = s.start_time ? new Date(s.start_time).toLocaleString() : '—';
      const end   = el('td', { class: 'muted' }); end.textContent = s.end_time ? new Date(s.end_time).toLocaleString() : '—';
      const turns = el('td'); turns.textContent = s.turn_count ?? 0;
      const tools = el('td', { class: 'muted' }); tools.textContent = (s.tools_used || []).join(', ') || '—';
      const sum   = el('td'); sum.textContent = s.summary || '—';
      tr.append(sid, start, end, turns, tools, sum);
      tbody.appendChild(tr);
    }
    container.appendChild(el('div', { class: 'table-wrapper' }, el('table', {}, thead, tbody)));
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
  }
};

// ── API Graph ──────────────────────────────────────────────────────────────

renderers['api-graph'] = async function(container) {
  try {
    const graph = await api('GET', '/api/api-graph');
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'API Graph')
    ));

    if (!graph || graph === 'null') {
      container.appendChild(el('div', { class: 'empty-state' }, 'No API graph found — run ingest_spec first'));
      return;
    }

    const endpoints = graph.endpoints || {};
    if (Object.keys(endpoints).length === 0) {
      container.appendChild(el('div', { class: 'empty-state' }, 'No endpoints in graph'));
      return;
    }

    if (graph.context) {
      const ctx = graph.context;
      const rows = [['Framework', ctx.framework], ['Language', ctx.language], ['Spec', ctx.spec_path]].filter(r => r[1]);
      const ctxCard = el('div', { class: 'card' },
        el('div', { class: 'card-title' }, 'Project Context'),
        ...rows.map(([k, v]) =>
          el('div', { style: 'display:flex;gap:12px;margin-bottom:5px;align-items:baseline' },
            el('span', { class: 'muted', style: 'width:90px;flex-shrink:0' }, k),
            el('span', { class: 'mono' }, v)
          )
        )
      );
      container.appendChild(ctxCard);
    }

    const epCard = el('div', { class: 'card' });
    epCard.appendChild(el('div', { class: 'card-title' }, `Endpoints (${Object.keys(endpoints).length})`));

    for (const [path, ep] of Object.entries(endpoints)) {
      const wrap = el('div', { class: 'collapsible-wrap' });
      const hdr = el('div', { class: 'collapsible-header' },
        el('span', { class: 'collapsible-arrow' }, '▶'),
        methodBadge(ep.method || 'GET'),
        el('span', { class: 'mono', style: 'flex:1;font-size:13px' }, path)
      );
      if (ep.auth_type) hdr.appendChild(el('span', { class: 'badge badge-post', style: 'margin-left:auto' }, ep.auth_type));

      const body = el('div', { class: 'collapsible-body' });
      if (ep.summary) body.appendChild(el('p', { style: 'margin-bottom:8px;color:var(--text-muted);font-size:13px' }, ep.summary));

      if (ep.parameters && ep.parameters.length > 0) {
        body.appendChild(el('div', { class: 'card-title' }, 'Parameters'));
        for (const p of ep.parameters) {
          const row = el('div', { class: 'mono', style: 'font-size:12px;margin-bottom:3px;' });
          row.textContent = `${p.name} (${p.type || 'any'})${p.required ? ' *' : ''} — ${p.description || ''}`;
          body.appendChild(row);
        }
      }

      if (ep.security && ep.security.length > 0) {
        body.appendChild(el('div', { class: 'card-title', style: 'margin-top:10px' }, 'Security Risks'));
        for (const risk of ep.security) {
          body.appendChild(
            el('div', { style: 'display:flex;gap:8px;align-items:center;margin-bottom:5px' },
              severityBadge(risk.severity),
              el('span', { style: 'font-size:12px' }, risk.risk || risk.description || '')
            )
          );
        }
      }

      hdr.addEventListener('click', () => {
        hdr.classList.toggle('open');
        body.classList.toggle('open');
      });

      wrap.append(hdr, body);
      epCard.appendChild(wrap);
    }
    container.appendChild(epCard);
  } catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
  }
};

// ── Exports ────────────────────────────────────────────────────────────────

renderers['exports'] = async function(container) {
  let names, selectedExport = null;

  async function load() { names = await api('GET', '/api/exports'); }

  async function render() {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'section-header' },
      el('h1', { class: 'section-title' }, 'Exports')
    ));

    const split = el('div', { class: 'split-layout' });
    const listEl = el('div', { class: 'split-list' });

    if (!names || names.length === 0) {
      listEl.appendChild(el('div', { class: 'split-empty' }, 'No exports yet'));
    } else {
      for (const name of names) {
        const item = el('div', {
          class: 'split-list-item' + (name === selectedExport ? ' selected' : ''),
          'data-name': name,
        });
        item.textContent = name;
        item.addEventListener('click', () => { selectedExport = name; openFile(name); });
        listEl.appendChild(item);
      }
    }

    const panel = el('div', { class: 'split-panel' },
      el('div', { class: 'split-empty' }, 'Select a file to view')
    );
    split.append(listEl, panel);
    container.appendChild(split);
  }

  function openFile(name) {
    const panel = container.querySelector('.split-panel');
    if (!panel) return;
    panel.innerHTML = '';
    container.querySelectorAll('.split-list-item').forEach(i =>
      i.classList.toggle('selected', i.getAttribute('data-name') === name)
    );
    fetch(`/api/exports/${encodeURIComponent(name)}`).then(async res => {
      const text = await res.text();
      panel.innerHTML = '';
      panel.appendChild(el('div', { style: 'font-size:12px;font-family:monospace;color:var(--text-muted);margin-bottom:12px' }, name));
      panel.appendChild(el('pre', {}, text));
    }).catch(e => panel.appendChild(el('div', { class: 'empty-state' }, e.message)));
  }

  try { await load(); await render(); }
  catch (e) {
    container.innerHTML = '';
    container.appendChild(el('div', { class: 'empty-state' }, 'Failed: ' + e.message));
  }
};

// ── KV editor ─────────────────────────────────────────────────────────────

function kvEditor(initialVars, prefix) {
  const table = el('table', { class: 'kv-table' });
  const tbody = el('tbody');
  table.appendChild(tbody);
  for (const [k, v] of Object.entries(initialVars)) tbody.appendChild(kvRow(k, v, prefix));
  return table;
}

function kvRow(k, v, prefix) {
  const keyInput = el('input', { type: 'text', 'data-kv-key': '1' }); keyInput.value = k;
  const valInput = el('input', { type: 'text', 'data-kv-val': '1' }); valInput.value = v;
  const delBtn = el('button', { type: 'button', class: 'btn btn-danger btn-xs' }, '×');
  const tr = el('tr',
    el('td', { class: 'kv-key-col' }, keyInput),
    el('td', { class: 'kv-val-col' }, valInput),
    el('td', { class: 'kv-act-col' }, delBtn)
  );
  delBtn.addEventListener('click', () => tr.remove());
  return tr;
}

function addKVRow(table) {
  (table.querySelector('tbody') || table).appendChild(kvRow('', '', 'kv'));
}

function collectKV(table) {
  const result = {};
  const keys = table.querySelectorAll('[data-kv-key]');
  const vals = table.querySelectorAll('[data-kv-val]');
  keys.forEach((k, i) => { const key = k.value.trim(); if (key) result[key] = vals[i]?.value ?? ''; });
  return result;
}

// ── Init ───────────────────────────────────────────────────────────────────

document.addEventListener('DOMContentLoaded', () => {
  document.querySelectorAll('.nav-item').forEach(a => {
    a.addEventListener('click', e => {
      e.preventDefault();
      showSection(a.dataset.section);
    });
  });
  showSection('dashboard');
});
