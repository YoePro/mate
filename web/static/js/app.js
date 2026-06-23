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
  return entityType === 'company' || entityType === 'association' || entityType === 'school';
}

function isTemporaryApiError(err) {
  return err && /^(404|501)\b/.test(err.message || '');
}

function createTemporaryEntity(entityType, data) {
  const record = Object.assign({}, data, {
    id: `tmp-${entityType}-${Date.now()}-${temporaryIdCounter++}`,
    temporary: true,
  });

  if (entityType === 'company' || entityType === 'association' || entityType === 'school') {
    record.type = entityType;
  }

  return record;
}

function createTemporaryRelationship(source, target, type, notes) {
  return {
    id: `tmp-rel-${Date.now()}-${temporaryIdCounter++}`,
    source_id: source.id,
    source_type: source.entityType,
    target_id: target.id,
    target_type: target.entityType,
    type,
    notes: notes || null,
    temporary: true,
  };
}

function localOnlyWarning(action, err) {
  console.warn(`${action} is using temporary frontend state:`, err.message);
}

// ---- Boot ----

async function boot() {
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
  graph.on('selection-changed', (id) => {
    updateNodeSelection(id);
    if (id) {
      renderInspector(id);
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
    const data = await apiLoadAll();
    graph.load(data);
    renderAllNodes();
    renderAllLinks();
    setStatus('ok');
  } catch (err) {
    console.error('Failed to load graph:', err);
    setStatus('error');
  }
}

function setStatus(state) {
  const dot = el('status-indicator');
  dot.classList.remove('status-ok', 'status-loading', 'status-error');
  dot.classList.add('status-dot', `status-${state}`);
}

// ---- Header buttons ----

function initHeaderButtons() {
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
      relType.textContent = REL_LABELS[link.type] || link.type;
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

function initModals() {
  el('modal-close').addEventListener('click', closeModal);
  el('modal-cancel').addEventListener('click', closeModal);
  el('modal-overlay').addEventListener('click', (e) => {
    if (e.target === el('modal-overlay')) closeModal();
  });
  el('modal-save').addEventListener('click', handleModalSave);
  el('modal-delete').addEventListener('click', handleModalDelete);
}

function openAddModal(entityType, x, y) {
  modalEntityType = entityType;
  modalEntityId = null;
  modalX = x;
  modalY = y;
  el('modal-title').textContent = `Add ${ENTITY_LABELS[entityType] || entityType}`;
  el('modal-delete').style.display = 'none';
  buildModalForm(entityType, null);
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
  buildModalForm(node.entityType, node.data);
  el('modal-overlay').classList.remove('modal-hidden');
}

function buildModalForm(entityType, data) {
  const form = el('modal-form');
  clearChildren(form);
  const fields = FORM_FIELDS[entityType] || [];

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

async function handleModalDelete() {
  if (!modalEntityId) return;
  const node = graph.getNode(modalEntityId);
  if (!confirm(`Delete "${node ? node.label : '?'}"? This will also remove all relationships.`)) return;

  try {
    if (!isTemporaryGraphMode()) {
      try {
        await apiDelete(modalEntityType, modalEntityId);
      } catch (err) {
        if (!isTemporaryApiError(err)) throw err;
        localOnlyWarning('Node delete', err);
      }
    }
    removeNodeEl(modalEntityId);
    graph.removeNode(modalEntityId);
    closeModal();
  } catch (err) {
    console.error('Delete failed:', err);
    alert('Could not delete. Please try again.');
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

function initLinkModal() {
  el('link-modal-close').addEventListener('click', closeLinkModal);
  el('link-modal-cancel').addEventListener('click', closeLinkModal);
  el('link-modal-overlay').addEventListener('click', (e) => {
    if (e.target === el('link-modal-overlay')) closeLinkModal();
  });
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

  el('link-notes').value = '';
  el('link-modal-overlay').classList.remove('modal-hidden');
  el('link-type-select').focus();

  graph.setLinkSource(null);
}

async function handleLinkSave() {
  const type = el('link-type-select').value;
  const notes = el('link-notes').value.trim();
  const source = graph.getNode(linkSourceNodeId);
  const target = graph.getNode(linkTargetNodeId);
  if (!source || !target) return;

  el('link-modal-save').disabled = true;
  try {
    let rel;
    if (isTemporaryGraphMode()) {
      rel = createTemporaryRelationship(source, target, type, notes);
    } else {
      try {
        rel = await apiCreateRelationship(
          source.id, source.entityType,
          target.id, target.entityType,
          type, notes
        );
      } catch (err) {
        if (!isTemporaryApiError(err)) throw err;
        localOnlyWarning('Relationship creation', err);
        rel = createTemporaryRelationship(source, target, type, notes);
      }
    }
    graph.addLink({
      id: rel.id,
      sourceId: source.id,
      targetId: target.id,
      sourceType: source.entityType,
      targetType: target.entityType,
      type: rel.type,
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
      if (!confirm(`Delete "${node ? node.label : '?'}"?`)) return;
      const deletePromise = isTemporaryGraphMode()
        ? Promise.resolve()
        : apiDelete(node.entityType, nodeId).catch(err => {
          if (!isTemporaryApiError(err)) throw err;
          localOnlyWarning('Node delete', err);
        });

      deletePromise.then(() => {
        removeNodeEl(nodeId);
        graph.removeNode(nodeId);
      }).catch(err => {
        console.error('Delete failed:', err);
        alert('Could not delete.');
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
