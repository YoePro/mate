// app.js — main entry point

// ---- Modal state ----
let modalEntityType = null;
let modalEntityId = null;
let modalX = 0, modalY = 0;
let attributeOwnerType = null;
let attributeOwnerId = null;

let temporaryIdCounter = 1;

function isTemporaryGraphMode() {
  return Boolean(window.MATE_CONFIG && window.MATE_CONFIG.temporaryGraph);
}

function isOrganizationEntityType(entityType) {
  return [
    'company',
    'association',
    'school',
    'government',
    'political_party',
    'religious_organization',
    'sports_club',
    'military_unit',
    'ngo',
    'community',
  ].includes(entityType);
}

function isTemporaryApiError(err) {
  return err && /^(404|501)\b/.test(err.message || '');
}

function isArchiveAction(entityType) {
  return entityType === 'project' || isOrganizationEntityType(entityType) || (entityType === 'person' && window.currentNetworkId);
}

function destructiveActionLabel(entityType) {
  return isArchiveAction(entityType) ? 'Archive' : 'Delete';
}

function createTemporaryEntity(entityType, data) {
  const record = Object.assign({}, data, {
    id: `tmp-${entityType}-${Date.now()}-${temporaryIdCounter++}`,
    temporary: true,
  });

  if (isOrganizationEntityType(entityType)) {
    record.type = entityType;
  }

  return record;
}

function createTemporaryRelationship(source, target, type, data) {
  return Object.assign({
    id: `tmp-rel-${Date.now()}-${temporaryIdCounter++}`,
    source_id: source.id,
    source_type: source.entityType,
    target_id: target.id,
    target_type: target.entityType,
    type,
  }, data || {}, {
    temporary: true,
  });
}

function localOnlyWarning(action, err) {
  console.warn(`${action} is using temporary frontend state:`, err.message);
}

// ---- Boot ----

let currentAccount = null;
let currentNetwork = null;
let ownedNetworks = [];
let currentCustomRelationshipTypes = [];
window.currentNetworkId = null;

async function boot() {
  currentAccount = await requireAccount();
  if (!currentAccount) return;

  await initNetworks();
  initToolbox();
  initInspector();
  initModals();
  initLinkModal();
  initAttributeModal();
  initSearch();
  initKeyboard();
  initHeaderButtons();

  graph.on('nodes-changed', renderAllNodes);
  graph.on('node-updated', (node) => {
    renderNode(node);
    if (graph.selectedNodeId === node.id) renderInspector(node.id);
  });
  graph.on('node-moved', (node) => {
    updateLinksForNode(node.id);
  });
  graph.on('links-changed', renderAllLinks);
  graph.on('selection-changed', () => {
    updateNodeSelection();
    if (graph.selectedNodeId) {
      renderInspector(graph.selectedNodeId);
      el('inspector').classList.remove('inspector-hidden');
    } else {
      el('inspector').classList.add('inspector-hidden');
    }
  });
  graph.on('link-source-changed', updateLinkSource);

  await loadGraph();
}

async function loadGraph() {
  setStatus('loading');
  try {
    if (!window.currentNetworkId) {
      currentCustomRelationshipTypes = [];
      graph.clear();
      renderAllNodes();
      renderAllLinks();
      setStatus('ok');
      return;
    }
    const data = await apiLoadAll();
    currentCustomRelationshipTypes = data.custom_relationship_types || [];
    graph.load(data);
    renderAllNodes();
    renderAllLinks();
    setStatus('ok');
  } catch (err) {
    console.error('Failed to load graph:', err);
    setStatus('error');
  }
}

async function initNetworks() {
  const select = el('network-select');
  const createBtn = el('btn-new-network');
  const renameBtn = el('btn-rename-network');
  createBtn.addEventListener('click', handleCreateNetwork);
  renameBtn.addEventListener('click', handleRenameNetwork);
  select.addEventListener('change', async () => {
    const selected = ownedNetworks.find(n => n.id === select.value);
    setCurrentNetwork(selected || null);
    await loadGraph();
  });
  initNetworkSearch();
  await refreshNetworks();
}

async function refreshNetworks() {
  ownedNetworks = await apiListNetworks();
  if (!ownedNetworks.length && currentAccount && currentAccount.role !== 'viewer') {
    const created = await apiCreateNetwork({ name: 'Default network' });
    ownedNetworks.push(created);
  }
  const savedID = window.localStorage.getItem('mate.currentNetworkId');
  const selected = ownedNetworks.find(n => n.id === savedID) || ownedNetworks[0] || null;
  renderNetworkOptions(ownedNetworks, selected ? selected.id : '');
  setCurrentNetwork(selected);
}

function updateNetworkActions() {
  const createBtn = el('btn-new-network');
  const renameBtn = el('btn-rename-network');
  const canWrite = currentAccount && currentAccount.role !== 'viewer';
  if (createBtn) createBtn.disabled = !canWrite;
  if (renameBtn) renameBtn.disabled = !canWrite || !currentNetwork;
}

function renderNetworkOptions(networks, selectedID) {
  const select = el('network-select');
  clearChildren(select);
  if (!networks.length) {
    const option = createEl('option', '');
    option.value = '';
    option.textContent = 'No networks';
    select.appendChild(option);
    select.disabled = true;
    return;
  }
  select.disabled = false;
  networks.forEach(network => {
    const option = createEl('option', '');
    option.value = network.id;
    option.textContent = network.name;
    option.selected = network.id === selectedID;
    select.appendChild(option);
  });
}

function setCurrentNetwork(network) {
  currentNetwork = network;
  window.currentNetworkId = network ? network.id : null;
  if (window.currentNetworkId) {
    window.localStorage.setItem('mate.currentNetworkId', window.currentNetworkId);
  } else {
    window.localStorage.removeItem('mate.currentNetworkId');
  }
  updateNetworkActions();
}

async function handleCreateNetwork() {
  if (currentAccount && currentAccount.role === 'viewer') return;
  const name = prompt('Network name');
  if (!name || !name.trim()) return;
  try {
    const network = await apiCreateNetwork({ name: name.trim() });
    ownedNetworks = await apiListNetworks();
    renderNetworkOptions(ownedNetworks, network.id);
    setCurrentNetwork(network);
    graph.selectNode(null);
    await loadGraph();
  } catch (err) {
    console.error('Network create failed:', err);
    alert('Could not create network.');
  }
}

async function handleRenameNetwork() {
  if (!currentNetwork || (currentAccount && currentAccount.role === 'viewer')) return;
  const name = prompt('Network name', currentNetwork.name || '');
  if (!name || !name.trim() || name.trim() === currentNetwork.name) return;
  try {
    const updated = await apiUpdateNetwork(currentNetwork.id, {
      name: name.trim(),
      description: currentNetwork.description || '',
    });
    ownedNetworks = ownedNetworks.map(network => network.id === updated.id ? updated : network);
    renderNetworkOptions(ownedNetworks, updated.id);
    setCurrentNetwork(updated);
  } catch (err) {
    console.error('Network rename failed:', err);
    alert('Could not rename network.');
  }
}

function initNetworkSearch() {
  const input = el('network-search-input');
  let timer = null;
  input.addEventListener('input', () => {
    clearTimeout(timer);
    const query = input.value.trim();
    if (query.length < 2) {
      removeNetworkResults();
      return;
    }
    timer = setTimeout(async () => {
      try {
        const results = await apiSearchNetworks(query);
        showNetworkResults(results);
      } catch (err) {
        console.warn('Network search failed:', err.message);
        removeNetworkResults();
      }
    }, 180);
  });
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      input.value = '';
      removeNetworkResults();
    }
  });
  document.addEventListener('click', (e) => {
    const wrap = input.closest('.network-search-wrap');
    if (wrap && !wrap.contains(e.target)) removeNetworkResults();
  });
}

function showNetworkResults(results) {
  removeNetworkResults();
  const input = el('network-search-input');
  const wrap = input.closest('.network-search-wrap');
  const box = createEl('div', 'network-results');
  if (!results.length) {
    const empty = createEl('button', 'network-result-item');
    empty.type = 'button';
    empty.disabled = true;
    empty.textContent = 'No matching networks';
    box.appendChild(empty);
  }
  results.forEach(result => {
    const item = createEl('button', 'network-result-item');
    item.type = 'button';
    item.disabled = !result.owned;
    const name = createEl('span', 'network-result-name');
    name.textContent = result.name || 'Untitled network';
    const meta = createEl('span', 'network-result-meta');
    meta.textContent = result.owned ? 'Owned by you' : 'Discoverable - no access';
    item.appendChild(name);
    item.appendChild(meta);
    if (result.description) {
      const desc = createEl('span', 'network-result-meta');
      desc.textContent = result.description;
      item.appendChild(desc);
    }
    if (result.owned) {
      item.addEventListener('click', async () => {
        const selected = ownedNetworks.find(n => n.id === result.id);
        if (!selected) return;
        renderNetworkOptions(ownedNetworks, selected.id);
        setCurrentNetwork(selected);
        input.value = '';
        removeNetworkResults();
        graph.selectNode(null);
        await loadGraph();
      });
    }
    box.appendChild(item);
  });
  wrap.appendChild(box);
}

function removeNetworkResults() {
  qsa('.network-results').forEach(node => node.remove());
}

function setStatus(state) {
  const dot = el('status-indicator');
  dot.classList.remove('status-ok', 'status-loading', 'status-error');
  dot.classList.add('status-dot', `status-${state}`);
}


async function requireAccount() {
  try {
    return await apiCurrentAccount();
  } catch (err) {
    if (/^401\b/.test(err.message || '')) {
      window.location.href = '/login';
      return null;
    }
    console.error('Session check failed:', err);
    window.location.href = '/login';
    return null;
  }
}

function renderCurrentAccount() {
  const label = el('account-label');
  if (!label || !currentAccount) return;
  label.textContent = currentAccount.display_name || currentAccount.email || '';
  label.title = currentAccount.role ? `${currentAccount.email} (${currentAccount.role})` : currentAccount.email;
}

async function handleLogout() {
  el('btn-logout').disabled = true;
  try {
    await apiLogout();
  } catch (err) {
    console.warn('Logout failed:', err.message);
  } finally {
    window.location.href = '/login';
  }
}

// ---- Header buttons ----

function initHeaderButtons() {
  renderCurrentAccount();
  el('btn-logout').addEventListener('click', handleLogout);

  el('btn-fit').addEventListener('click', () => {
    canvas.fitToNodes(graph.nodes);
  });

  el('btn-reload').addEventListener('click', async () => {
    graph.selectNode(null);
    await loadGraph();
  });
}

// ---- Inspector ----

function initInspector() {
  el('inspector-close').addEventListener('click', () => {
    graph.selectNode(null);
  });
}

function renderInspector(nodeId) {
  const node = graph.getNode(nodeId);
  if (!node) return;

  el('inspector-title').textContent = ENTITY_LABELS[node.entityType] || 'Details';

  const body = el('inspector-body');
  clearChildren(body);

  const badge = createEl('div', 'inspector-entity-type');
  badge.style.background = getNodeColor(node.entityType) + '22';
  badge.style.color = getNodeColor(node.entityType);
  badge.textContent = ENTITY_LABELS[node.entityType] || node.entityType;
  body.appendChild(badge);

  if (node.data && node.data.deceased) {
    const dec = createEl('div', 'inspector-deceased');
    dec.textContent = '\u2020 Deceased / Legacy';
    body.appendChild(dec);
  }

  const name = createEl('div', 'inspector-name');
  name.textContent = node.label || '?';
  body.appendChild(name);

  if (node.data) {
    if (node.data.nickname && node.data.nickname !== node.label) {
      const nick = createEl('div', 'inspector-subtitle');
      nick.textContent = `"${node.data.nickname}"`;
      body.appendChild(nick);
    }
    if (node.data.title) {
      const title = createEl('div', 'inspector-subtitle');
      title.textContent = node.data.title;
      body.appendChild(title);
    }
    if (node.data.gender) {
      const gender = createEl('div', 'inspector-subtitle');
      gender.textContent = `Gender: ${GENDER_LABELS[node.data.gender] || node.data.gender}`;
      body.appendChild(gender);
    }
    if (node.data.web) {
      const web = createEl('div', 'inspector-subtitle');
      web.textContent = node.data.web;
      body.appendChild(web);
    }
    if (node.data.description) {
      const sec = createEl('div', 'inspector-section');
      const title = createEl('div', 'inspector-section-title');
      title.textContent = 'Description';
      const description = createEl('div', 'inspector-notes');
      description.textContent = node.data.description;
      sec.appendChild(title);
      sec.appendChild(description);
      body.appendChild(sec);
    }
    if (node.data.notes) {
      const sec = createEl('div', 'inspector-section');
      const title = createEl('div', 'inspector-section-title');
      title.textContent = 'Notes';
      const notes = createEl('div', 'inspector-notes');
      notes.textContent = node.data.notes;
      sec.appendChild(title);
      sec.appendChild(notes);
      body.appendChild(sec);
    }
  }

  if (node.entityType === 'person' || isOrganizationEntityType(node.entityType)) {
    renderAttributes(node.entityType === 'person' ? 'person' : 'organization', node.id, body);
  }

  const nodeLinks = graph.getNodeLinks(nodeId);
  if (nodeLinks.length > 0) {
    const sec = createEl('div', 'inspector-section');
    const title = createEl('div', 'inspector-section-title');
    title.textContent = `Connections (${nodeLinks.length})`;
    sec.appendChild(title);

    nodeLinks.forEach(link => {
      const otherId = link.sourceId === nodeId ? link.targetId : link.sourceId;
      const other = graph.getNode(otherId);
      if (!other) return;

      const item = createEl('div', 'inspector-rel-item');
      const dot = createEl('span', 'inspector-rel-dot');
      dot.style.background = getNodeColor(other.entityType);
      const relName = createEl('span', 'inspector-rel-name');
      relName.textContent = other.label || '?';
      const relType = createEl('span', 'inspector-rel-type');
      const context = [link.role, link.startDate, link.endDate].filter(Boolean).join(', ');
      relType.textContent = context ? `${relationshipLabel(link)} (${context})` : relationshipLabel(link);
      item.appendChild(dot);
      item.appendChild(relName);
      item.appendChild(relType);
      item.addEventListener('click', () => graph.selectNode(otherId));
      sec.appendChild(item);
    });
    body.appendChild(sec);
  }

  const actions = createEl('div', 'inspector-actions');
  const editBtn = createEl('button', 'btn btn-ghost');
  editBtn.textContent = 'Edit';
  editBtn.addEventListener('click', () => openEditModal(nodeId));
  actions.appendChild(editBtn);

  if (node.entityType === 'person' || isOrganizationEntityType(node.entityType)) {
    const attrBtn = createEl('button', 'btn btn-ghost');
    attrBtn.textContent = 'Add attribute';
    attrBtn.addEventListener('click', () => openAttributeModal(nodeId));
    actions.appendChild(attrBtn);
  }

  body.appendChild(actions);
}

async function renderAttributes(ownerType, ownerId, body) {
  const sec = createEl('div', 'inspector-section');
  const title = createEl('div', 'inspector-section-title');
  title.textContent = 'Attributes';
  const list = createEl('div', 'inspector-attribute-list');
  list.textContent = 'Loading...';
  sec.appendChild(title);
  sec.appendChild(list);
  body.appendChild(sec);

  try {
    const attributes = await apiListAttributes(ownerType, ownerId);
    clearChildren(list);
    if (!attributes.length) {
      list.textContent = 'No attributes yet.';
      return;
    }
    attributes.forEach(attribute => {
      list.appendChild(buildAttributeItem(attribute));
    });
  } catch (err) {
    list.textContent = 'Could not load attributes.';
    console.error('Attribute load failed:', err);
  }
}

function buildAttributeItem(attribute) {
  const item = createEl('div', 'inspector-attribute-item');
  const type = createEl('div', 'inspector-attribute-type');
  type.textContent = ATTRIBUTE_LABELS[attribute.type] || attribute.type;
  const value = createEl('div', 'inspector-attribute-value');
  value.textContent = attribute.value;
  item.appendChild(type);
  item.appendChild(value);

  const metaParts = [];
  if (attribute.start_date) metaParts.push(attribute.start_date);
  if (attribute.end_date) metaParts.push(attribute.end_date);
  if (attribute.current) metaParts.push('current');
  if (metaParts.length) {
    const meta = createEl('div', 'inspector-attribute-meta');
    meta.textContent = metaParts.join(' - ');
    item.appendChild(meta);
  }

  if (attribute.notes) {
    const notes = createEl('div', 'inspector-notes');
    notes.textContent = attribute.notes;
    item.appendChild(notes);
  }

  return item;
}

// ---- Node creation modal ----

const FORM_FIELDS = {
  person: [
    { key: 'name',     label: 'Full name',       type: 'text',     required: true },
    { key: 'nickname', label: 'Nickname',         type: 'text'  },
    { key: 'gender',   label: 'Gender',           type: 'select',   options: [
      { value: '',  label: 'Unspecified' },
      { value: 'm', label: 'Male' },
      { value: 'f', label: 'Female' },
      { value: 'o', label: 'Other' },
    ] },
    { key: 'title',    label: 'Title / Role',     type: 'text'  },
    { key: 'notes',    label: 'Notes',            type: 'textarea' },
    { key: 'deceased', label: 'Deceased / Legacy', type: 'checkbox' },
  ],
  company: [
    { key: 'name', label: 'Company name', type: 'text', required: true },
    { key: 'description', label: 'Description', type: 'textarea' },
    { key: 'web', label: 'Web / Reference', type: 'text' },
    { key: 'notes', label: 'Notes', type: 'textarea' },
  ],
  association: [
    { key: 'name', label: 'Association name', type: 'text', required: true },
    { key: 'description', label: 'Description', type: 'textarea' },
    { key: 'web', label: 'Web / Reference', type: 'text' },
    { key: 'notes', label: 'Notes', type: 'textarea' },
  ],
  school: [
    { key: 'name', label: 'School name', type: 'text', required: true },
    { key: 'description', label: 'Description', type: 'textarea' },
    { key: 'web', label: 'Web / Reference', type: 'text' },
    { key: 'notes', label: 'Notes', type: 'textarea' },
  ],
  organization: [
    { key: 'name', label: 'Organization name', type: 'text', required: true },
    { key: 'description', label: 'Description', type: 'textarea' },
    { key: 'web', label: 'Web / Reference', type: 'text' },
    { key: 'notes', label: 'Notes', type: 'textarea' },
  ],
  project: [
    { key: 'name', label: 'Project name', type: 'text', required: true },
    { key: 'status', label: 'Status', type: 'text' },
    { key: 'description', label: 'Description', type: 'textarea' },
    { key: 'web', label: 'Web / Reference', type: 'text' },
    { key: 'notes', label: 'Notes', type: 'textarea' },
  ],
  location:    [{ key: 'name', label: 'Location name',    type: 'text', required: true }, { key: 'notes', label: 'Notes', type: 'textarea' }],
  tag:         [{ key: 'name', label: 'Tag name',         type: 'text', required: true }],
};

const GENDER_LABELS = {
  m: 'Male',
  f: 'Female',
  o: 'Other',
};

const ATTRIBUTE_LABELS = {
  title: 'Title',
  role: 'Role',
  employment: 'Employment',
  education: 'Education',
  certification: 'Certification',
  award: 'Award',
  board_membership: 'Board membership',
  board_role: 'Board role',
  membership: 'Membership',
  milestone: 'Milestone',
  competition: 'Competition',
  achievement: 'Achievement',
};

const ATTRIBUTE_TYPES = {
  person: [
    ['title', 'Title'],
    ['role', 'Role'],
    ['employment', 'Employment'],
    ['education', 'Education'],
    ['certification', 'Certification'],
    ['award', 'Award'],
    ['board_membership', 'Board membership'],
    ['competition', 'Competition'],
    ['achievement', 'Achievement'],
  ],
  organization: [
    ['role', 'Role'],
    ['membership', 'Membership'],
    ['board_role', 'Board role'],
    ['certification', 'Certification'],
    ['award', 'Award'],
    ['milestone', 'Milestone'],
  ],
};

const RELATIONSHIP_OPTIONS = {
  'person:person': [
    ['knows', 'Knows'],
    ['spouse_of', 'Spouse of'],
    ['parent_of', 'Parent of'],
    ['sibling_of', 'Sibling of'],
  ],
  'person:organization': [
    ['works_at', 'Works at'],
    ['member_of', 'Member of'],
    ['studied_at', 'Studied at'],
  ],
  'organization:person': [
    ['member_of', 'Has member'],
    ['works_at', 'Employs'],
  ],
  'organization:organization': [
    ['partner_of', 'Partner of'],
    ['owns', 'Owns'],
    ['sponsors', 'Sponsors'],
  ],
  'person:location': [
    ['lives_in', 'Lives in'],
  ],
  'person:tag': [
    ['has_tag', 'Has tag'],
  ],
  'organization:tag': [
    ['has_tag', 'Has tag'],
  ],
  'person:project': [
    ['works_on', 'Works on'],
  ],
  'organization:project': [
    ['sponsors', 'Sponsors'],
  ],
};

const RELATIONSHIP_CONTEXT_TYPES = new Set([
  'works_at',
  'member_of',
  'studied_at',
  'works_on',
  'sponsors',
]);

function initModals() {
  el('modal-close').addEventListener('click', closeModal);
  el('modal-cancel').addEventListener('click', closeModal);
  el('modal-overlay').addEventListener('click', (e) => {
    if (e.target === el('modal-overlay')) closeModal();
  });
  el('modal-save').addEventListener('click', handleModalSave);
  el('modal-delete').addEventListener('click', handleModalDelete);
}

function openAddModal(entityType, x, y, defaults) {
  modalEntityType = entityType;
  modalEntityId = null;
  modalX = x;
  modalY = y;
  el('modal-title').textContent = `Add ${ENTITY_LABELS[entityType] || entityType}`;
  el('modal-delete').style.display = 'none';
  buildModalForm(entityType, defaults || null);
  el('modal-overlay').classList.remove('modal-hidden');
  const first = qs('.form-input', el('modal-form'));
  if (first) first.focus();
}

function openEditModal(nodeId) {
  const node = graph.getNode(nodeId);
  if (!node) return;
  modalEntityType = node.entityType;
  modalEntityId = nodeId;
  el('modal-title').textContent = `Edit ${ENTITY_LABELS[node.entityType] || node.entityType}`;
  el('modal-delete').style.display = 'inline-flex';
  el('modal-delete').textContent = destructiveActionLabel(node.entityType);
  buildModalForm(node.entityType, node.data);
  el('modal-overlay').classList.remove('modal-hidden');
}

function buildModalForm(entityType, data) {
  const form = el('modal-form');
  clearChildren(form);
  const fields = FORM_FIELDS[entityType] || (isOrganizationEntityType(entityType) ? FORM_FIELDS.organization : []);

  fields.forEach(field => {
    const group = createEl('div', 'form-group');
    const label = createEl('label', 'form-label');
    label.setAttribute('for', `field-${field.key}`);
    label.textContent = field.label + (field.required ? ' *' : '');

    let input;
    if (field.type === 'textarea') {
      input = createEl('textarea', 'form-input');
      input.rows = 3;
    } else if (field.type === 'select') {
      input = createEl('select', 'form-select');
      (field.options || []).forEach(option => {
        const opt = createEl('option', '');
        opt.value = option.value;
        opt.textContent = option.label;
        input.appendChild(opt);
      });
    } else if (field.type === 'checkbox') {
      const wrap = createEl('label', 'form-check');
      input = createEl('input', '');
      input.type = 'checkbox';
      if (data && data[field.key]) input.checked = true;
      const checkLabel = createEl('span', 'form-check-label');
      checkLabel.textContent = field.label;
      wrap.appendChild(input);
      wrap.appendChild(checkLabel);
      input.id = `field-${field.key}`;
      input.dataset.key = field.key;
      group.appendChild(wrap);
      form.appendChild(group);
      return;
    } else {
      input = createEl('input', 'form-input');
      input.type = 'text';
    }

    input.id = `field-${field.key}`;
    input.dataset.key = field.key;
    if (field.required) input.required = true;
    if (data && data[field.key] != null) input.value = data[field.key];

    group.appendChild(label);
    group.appendChild(input);
    form.appendChild(group);
  });
}

async function handleModalSave() {
  const form = el('modal-form');
  const data = {};

  qsa('[data-key]', form).forEach(input => {
    const key = input.dataset.key;
    data[key] = input.type === 'checkbox' ? input.checked : input.value.trim();
  });

  const nameField = qsa('[data-key]', form).find(i => i.dataset.key === 'name');
  if (nameField && !nameField.value.trim()) {
    nameField.focus();
    nameField.style.borderColor = 'var(--error)';
    setTimeout(() => { nameField.style.borderColor = ''; }, 1200);
    return;
  }

  el('modal-save').disabled = true;
  el('modal-save').textContent = 'Saving...';

  try {
    if (modalEntityId) {
      let result;
      if (isTemporaryGraphMode()) {
        result = Object.assign({}, graph.getNode(modalEntityId).data, data);
      } else {
        try {
          result = await apiUpdate(modalEntityType, modalEntityId, data);
        } catch (err) {
          if (!isTemporaryApiError(err)) throw err;
          localOnlyWarning('Node update', err);
          result = Object.assign({}, graph.getNode(modalEntityId).data, data);
        }
      }
      graph.updateNodeData(modalEntityId, result);
    } else {
      let result;
      if (isTemporaryGraphMode()) {
        result = createTemporaryEntity(modalEntityType, data);
      } else {
        try {
          if (modalEntityType === 'person' && window.currentNetworkId) {
            const existing = await maybeUseExistingPerson(data);
            if (existing) {
              data.id = existing.id;
            }
          }
          result = await apiCreate(modalEntityType, data);
          await apiSavePosition(result.id, modalEntityType, modalX, modalY);
        } catch (err) {
          if (!isTemporaryApiError(err)) throw err;
          localOnlyWarning('Node creation', err);
          result = createTemporaryEntity(modalEntityType, data);
        }
      }
      graph.addNode({
        id: result.id,
        entityType: modalEntityType,
        label: result.name || result.nickname || '?',
        x: modalX,
        y: modalY,
        data: result,
      });
    }
    closeModal();
  } catch (err) {
    console.error('Save failed:', err);
    alert('Could not save. Please try again.');
  } finally {
    el('modal-save').disabled = false;
    el('modal-save').textContent = 'Save';
  }
}

async function maybeUseExistingPerson(data) {
  if (!data.name && !data.nickname) return null;
  const matches = await apiPersonMatches({ name: data.name || '', nickname: data.nickname || '' });
  if (!matches.length) return null;
  const top = matches[0];
  const reasons = (top.reasons || []).join(', ') || 'possible duplicate';
  const confidence = Math.round((top.confidence || 0) * 100);
  const message = `Possible match: ${top.person.name} (${confidence}%, ${reasons}).\n\nLink this person to the selected network instead of creating a new global person?`;
  return confirm(message) ? top.person : null;
}

async function handleModalDelete() {
  if (!modalEntityId) return;
  const node = graph.getNode(modalEntityId);
  const action = destructiveActionLabel(modalEntityType);
  if (!confirm(`${action} "${node ? node.label : '?'}"?`)) return;

  try {
    if (!isTemporaryGraphMode()) {
      try {
        await apiDelete(modalEntityType, modalEntityId);
      } catch (err) {
        if (!isTemporaryApiError(err)) throw err;
        localOnlyWarning(`Node ${action.toLowerCase()}`, err);
      }
    }
    removeNodeEl(modalEntityId);
    graph.removeNode(modalEntityId);
    closeModal();
  } catch (err) {
    console.error(`${action} failed:`, err);
    alert(`Could not ${action.toLowerCase()}. Please try again.`);
  }
}

function closeModal() {
  el('modal-overlay').classList.add('modal-hidden');
  modalEntityType = null;
  modalEntityId = null;
}

// ---- Attribute modal ----

function initAttributeModal() {
  el('attribute-modal-close').addEventListener('click', closeAttributeModal);
  el('attribute-modal-cancel').addEventListener('click', closeAttributeModal);
  el('attribute-modal-overlay').addEventListener('click', (e) => {
    if (e.target === el('attribute-modal-overlay')) closeAttributeModal();
  });
  el('attribute-modal-save').addEventListener('click', handleAttributeSave);
}

function openAttributeModal(nodeId) {
  const node = graph.getNode(nodeId);
  if (!node) return;
  attributeOwnerType = node.entityType === 'person' ? 'person' : 'organization';
  attributeOwnerId = nodeId;
  el('attribute-form').reset();
  configureAttributeModal(attributeOwnerType);
  el('attribute-modal-overlay').classList.remove('modal-hidden');
  el('attribute-value').focus();
}

function configureAttributeModal(ownerType) {
  const select = el('attribute-type');
  clearChildren(select);
  (ATTRIBUTE_TYPES[ownerType] || []).forEach(([value, label]) => {
    const option = createEl('option', '');
    option.value = value;
    option.textContent = label;
    select.appendChild(option);
  });

  const relatedLabel = el('attribute-related-label');
  relatedLabel.textContent = ownerType === 'person' ? 'Organization ID' : 'Person ID';
}

function closeAttributeModal() {
  el('attribute-modal-overlay').classList.add('modal-hidden');
  attributeOwnerType = null;
  attributeOwnerId = null;
}

async function handleAttributeSave() {
  if (!attributeOwnerType || !attributeOwnerId) return;

  const value = el('attribute-value').value.trim();
  if (!value) {
    el('attribute-value').focus();
    el('attribute-value').style.borderColor = 'var(--error)';
    setTimeout(() => { el('attribute-value').style.borderColor = ''; }, 1200);
    return;
  }

  const data = {
    type: el('attribute-type').value,
    value,
    start_date: el('attribute-start').value.trim(),
    end_date: el('attribute-end').value.trim(),
    current: el('attribute-current').checked,
    notes: el('attribute-notes').value.trim(),
  };
  const relatedId = el('attribute-related').value.trim();
  if (attributeOwnerType === 'person') {
    data.organization_id = relatedId;
  } else {
    data.person_id = relatedId;
  }

  el('attribute-modal-save').disabled = true;
  el('attribute-modal-save').textContent = 'Saving...';

  try {
    await apiCreateAttribute(attributeOwnerType, attributeOwnerId, data);
    const nodeId = attributeOwnerId;
    closeAttributeModal();
    renderInspector(nodeId);
  } catch (err) {
    console.error('Attribute save failed:', err);
    alert('Could not save attribute.');
  } finally {
    el('attribute-modal-save').disabled = false;
    el('attribute-modal-save').textContent = 'Save';
  }
}

// ---- Link modal ----

let linkSourceNodeId = null;
let linkTargetNodeId = null;

function broadEntityType(entityType) {
  if (entityType === 'person') return 'person';
  if (isOrganizationEntityType(entityType)) return 'organization';
  if (entityType === 'project') return 'project';
  if (entityType === 'location') return 'location';
  if (entityType === 'tag') return 'tag';
  return entityType;
}

function relationshipOptionsFor(source, target) {
  const directKey = `${broadEntityType(source.entityType)}:${broadEntityType(target.entityType)}`;
  const direct = RELATIONSHIP_OPTIONS[directKey];
  let options = direct || null;

  const reverseKey = `${broadEntityType(target.entityType)}:${broadEntityType(source.entityType)}`;
  const reverse = RELATIONSHIP_OPTIONS[reverseKey];
  if (!options && reverse) options = reverse;

  const base = options || [['knows', 'Knows']];
  const sourceType = broadEntityType(source.entityType);
  const targetType = broadEntityType(target.entityType);
  const custom = currentCustomRelationshipTypes
    .filter(item => item.source_type === sourceType && item.target_type === targetType)
    .map(item => [item.key, item.label]);
  return base.concat(custom);
}

function relationshipNeedsContext(type) {
  return RELATIONSHIP_CONTEXT_TYPES.has(type) || type.startsWith('custom_');
}

function customRelationshipType(label) {
  const slug = label
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '')
    .replace(/^([^a-z])/, 'x_$1')
    .slice(0, 56);
  return slug ? `custom_${slug}` : '';
}

function customRelationshipLabel(type) {
  const match = currentCustomRelationshipTypes.find(item => item.key === type);
  return match ? match.label : '';
}

function configureLinkTypeOptions(source, target) {
  const select = el('link-type-select');
  clearChildren(select);
  relationshipOptionsFor(source, target).forEach(([value, label]) => {
    const option = createEl('option', '');
    option.value = value;
    option.textContent = label;
    select.appendChild(option);
  });
  const custom = createEl('option', '');
  custom.value = '__custom__';
  custom.textContent = 'Custom...';
  select.appendChild(custom);
  updateLinkContextFields();
}

function updateLinkContextFields() {
  const selected = el('link-type-select').value;
  const custom = selected === '__custom__';
  const contextual = custom || relationshipNeedsContext(selected);
  el('link-custom-group').style.display = custom ? '' : 'none';
  el('link-role-group').style.display = contextual ? '' : 'none';
  el('link-dates-group').style.display = contextual ? '' : 'none';
  el('link-current-group').style.display = contextual ? '' : 'none';
}

function initLinkModal() {
  el('link-modal-close').addEventListener('click', closeLinkModal);
  el('link-modal-cancel').addEventListener('click', closeLinkModal);
  el('link-modal-overlay').addEventListener('click', (e) => {
    if (e.target === el('link-modal-overlay')) closeLinkModal();
  });
  el('link-type-select').addEventListener('change', updateLinkContextFields);
  el('link-modal-save').addEventListener('click', handleLinkSave);
}

function openLinkModal(sourceId, targetId) {
  linkSourceNodeId = sourceId;
  linkTargetNodeId = targetId;

  const source = graph.getNode(sourceId);
  const target = graph.getNode(targetId);

  const preview = el('link-nodes-preview');
  clearChildren(preview);

  function nodeChip(node) {
    const wrap = createEl('div', 'link-preview-node');
    const dot = createEl('div', 'link-preview-dot');
    dot.style.background = getNodeColor(node.entityType);
    const name = createEl('div', 'link-preview-name');
    name.textContent = node ? node.label : '?';
    wrap.appendChild(dot);
    wrap.appendChild(name);
    return wrap;
  }

  preview.appendChild(nodeChip(source));
  const arrow = createEl('div', 'link-preview-arrow');
  arrow.textContent = '\u2192';
  preview.appendChild(arrow);
  preview.appendChild(nodeChip(target));

  configureLinkTypeOptions(source, target);
  el('link-custom-label').value = '';
  el('link-role').value = '';
  el('link-start').value = '';
  el('link-end').value = '';
  el('link-current').checked = false;
  el('link-notes').value = '';
  el('link-modal-overlay').classList.remove('modal-hidden');
  el('link-type-select').focus();

  graph.setLinkSource(null);
}

async function handleLinkSave() {
  let type = el('link-type-select').value;
  const customLabel = el('link-custom-label').value.trim();
  let resolvedCustomLabel = customRelationshipLabel(type);
  if (type === '__custom__') {
    type = customRelationshipType(customLabel);
    if (!type) {
      el('link-custom-label').focus();
      el('link-custom-label').style.borderColor = 'var(--error)';
      setTimeout(() => { el('link-custom-label').style.borderColor = ''; }, 1200);
      return;
    }
    resolvedCustomLabel = customLabel;
  }
  const data = {
    custom_label: resolvedCustomLabel || '',
    role: el('link-role').value.trim(),
    start_date: el('link-start').value.trim(),
    end_date: el('link-end').value.trim(),
    current: el('link-current').checked,
    notes: el('link-notes').value.trim(),
  };
  const source = graph.getNode(linkSourceNodeId);
  const target = graph.getNode(linkTargetNodeId);
  if (!source || !target) return;

  el('link-modal-save').disabled = true;
  try {
    let rel;
    if (isTemporaryGraphMode()) {
      rel = createTemporaryRelationship(source, target, type, data);
    } else {
      try {
        if (type.startsWith('custom_') && resolvedCustomLabel && window.currentNetworkId) {
          const storedType = await apiCreateCustomRelationshipType(window.currentNetworkId, {
            key: type,
            label: resolvedCustomLabel,
            source_type: broadEntityType(source.entityType),
            target_type: broadEntityType(target.entityType),
            direction_behavior: 'directed',
          });
          currentCustomRelationshipTypes = currentCustomRelationshipTypes
            .filter(item => !(item.key === storedType.key && item.source_type === storedType.source_type && item.target_type === storedType.target_type))
            .concat(storedType);
        }
        rel = await apiCreateRelationship(
          source.id, source.entityType,
          target.id, target.entityType,
          type, data
        );
      } catch (err) {
        if (!isTemporaryApiError(err)) throw err;
        localOnlyWarning('Relationship creation', err);
        rel = createTemporaryRelationship(source, target, type, data);
      }
    }
    graph.addLink({
      id: rel.id,
      sourceId: source.id,
      targetId: target.id,
      sourceType: source.entityType,
      targetType: target.entityType,
      type: rel.type,
      customLabel: rel.custom_label,
      role: rel.role,
      startDate: rel.start_date,
      endDate: rel.end_date,
      current: rel.current,
      notes: rel.notes,
    });
    closeLinkModal();
  } catch (err) {
    console.error('Link save failed:', err);
    alert('Could not create relationship.');
  } finally {
    el('link-modal-save').disabled = false;
  }
}

function closeLinkModal() {
  el('link-modal-overlay').classList.add('modal-hidden');
  linkSourceNodeId = null;
  linkTargetNodeId = null;
  graph.setLinkSource(null);
}

// ---- Search ----

function initSearch() {
  const input = el('search-input');
  const wrap = qs('.search-wrap');
  let resultsEl = null;
  let searchTimer = null;

  input.addEventListener('input', () => {
    clearTimeout(searchTimer);
    const q = input.value.trim();
    if (q.length < 2) {
      removeSearchResults();
      return;
    }
    searchTimer = setTimeout(async () => {
      const results = await apiSearchAll(q).catch(() => null);
      if (!results) return;
      showSearchResults(results);
    }, 250);
  });

  input.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      input.value = '';
      removeSearchResults();
    }
  });

  document.addEventListener('click', (e) => {
    if (!wrap.contains(e.target)) removeSearchResults();
  });

  function showSearchResults(data) {
    removeSearchResults();
    const all = [];
    (data.persons || []).forEach(p => all.push({ id: p.id, name: p.name, type: 'person' }));
    (data.organizations || []).forEach(o => all.push({ id: o.id, name: o.name, type: o.type }));
    (data.locations || []).forEach(l => all.push({ id: l.id, name: l.name, type: 'location' }));
    (data.tags || []).forEach(t => all.push({ id: t.id, name: t.name, type: 'tag' }));
    if (!all.length) return;

    resultsEl = createEl('div', '');
    resultsEl.id = 'search-results';

    all.forEach(item => {
      const row = createEl('div', 'search-result-item');
      const dot = createEl('span', 'search-result-dot');
      dot.style.background = getNodeColor(item.type);
      const name = createEl('span', 'search-result-name');
      name.textContent = item.name;
      const type = createEl('span', 'search-result-type');
      type.textContent = ENTITY_LABELS[item.type] || item.type;
      row.appendChild(dot);
      row.appendChild(name);
      row.appendChild(type);
      row.addEventListener('click', () => {
        const node = graph.getNode(item.id);
        if (node) {
          graph.selectNode(node.id);
          canvas.fitToNodes([node]);
        }
        input.value = '';
        removeSearchResults();
      });
      resultsEl.appendChild(row);
    });
    wrap.appendChild(resultsEl);
  }

  function removeSearchResults() {
    if (resultsEl) { resultsEl.remove(); resultsEl = null; }
  }
}

// ---- Keyboard shortcuts ----

function initKeyboard() {
  document.addEventListener('keydown', (e) => {
    if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') return;

    if ((e.key === 'Delete' || e.key === 'Backspace') && graph.selectedNodeId) {
      e.preventDefault();
      const nodeId = graph.selectedNodeId;
      const node = graph.getNode(nodeId);
      const action = destructiveActionLabel(node ? node.entityType : '');
      if (!confirm(`${action} "${node ? node.label : '?'}"?`)) return;
      const deletePromise = isTemporaryGraphMode()
        ? Promise.resolve()
        : apiDelete(node.entityType, nodeId).catch(err => {
          if (!isTemporaryApiError(err)) throw err;
          localOnlyWarning(`Node ${action.toLowerCase()}`, err);
        });

      deletePromise.then(() => {
        removeNodeEl(nodeId);
        graph.removeNode(nodeId);
      }).catch(err => {
        console.error(`${action} failed:`, err);
        alert(`Could not ${action.toLowerCase()}.`);
      });
    }

    if (e.key === 'Escape') {
      graph.setLinkSource(null);
      graph.selectNode(null);
      cancelAddMode();
    }
  });
}

// ---- Start ----

document.addEventListener('DOMContentLoaded', boot);
