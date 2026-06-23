// profile.js — Person profile page

const PROP_TYPES = {
  title: {
    label: 'Title / Role',
    color: '#3b82f6',
    fields: [
      { key: 'label', label: 'Title / Role', placeholder: 'CEO, Director, Advisor...', required: true },
      { key: 'value', label: 'Organization', placeholder: 'Company or organization' },
    ],
    metaFields: [
      { key: 'year_from', label: 'From', type: 'number', placeholder: '2020' },
      { key: 'year_to', label: 'To', type: 'number', placeholder: 'Leave blank if current' },
    ],
    display: (p) => ({ main: p.label, sub: p.value || '', meta: _yearRange(p.meta) }),
  },
  education: {
    label: 'Education',
    color: '#10b981',
    fields: [
      { key: 'label', label: 'Degree / Program', placeholder: 'B.Sc., MBA, Ph.D....', required: true },
      { key: 'value', label: 'Institution', placeholder: 'University or school', required: true },
    ],
    metaFields: [
      { key: 'field', label: 'Field of Study', type: 'text', placeholder: 'Computer Science, Law...' },
      { key: 'year_from', label: 'From', type: 'number', placeholder: '2010' },
      { key: 'year_to', label: 'Graduated', type: 'number', placeholder: '2014' },
    ],
    display: (p) => ({
      main: p.label,
      sub: [p.value, p.meta && p.meta.field].filter(Boolean).join(' \u00b7 '),
      meta: _yearRange(p.meta),
    }),
  },
  certificate: {
    label: 'Certificate',
    color: '#f59e0b',
    fields: [
      { key: 'label', label: 'Certificate Name', placeholder: 'AWS Solutions Architect...', required: true },
      { key: 'value', label: 'Issuing Organization', placeholder: 'Amazon, Google...' },
    ],
    metaFields: [
      { key: 'date_issued', label: 'Date Issued', type: 'date' },
      { key: 'date_expires', label: 'Expiry Date', type: 'date' },
    ],
    display: (p) => ({ main: p.label, sub: p.value || '', meta: _dateRange(p.meta) }),
  },
  award: {
    label: 'Award',
    color: '#ef4444',
    fields: [
      { key: 'label', label: 'Award Name', placeholder: 'Best Innovation Award...', required: true },
      { key: 'value', label: 'Organization / Event', placeholder: 'Awarded by...' },
    ],
    metaFields: [
      { key: 'year', label: 'Year', type: 'number', placeholder: '2023' },
    ],
    display: (p) => ({ main: p.label, sub: p.value || '', meta: p.meta && p.meta.year ? String(p.meta.year) : '' }),
  },
  membership: {
    label: 'Membership',
    color: '#06b6d4',
    fields: [
      { key: 'value', label: 'Organization', placeholder: 'ACM, IEEE, Club...', required: true },
      { key: 'label', label: 'Role / Status', placeholder: 'Member, Fellow, Chair...' },
    ],
    metaFields: [
      { key: 'year_from', label: 'Since', type: 'number', placeholder: '2015' },
      { key: 'year_to', label: 'Until', type: 'number', placeholder: 'Leave blank if active' },
    ],
    display: (p) => ({ main: p.value, sub: p.label || '', meta: _yearRange(p.meta) }),
  },
  board: {
    label: 'Board Position',
    color: '#8b5cf6',
    fields: [
      { key: 'value', label: 'Organization', placeholder: 'Company or foundation', required: true },
      { key: 'label', label: 'Role', placeholder: 'Board Member, Chair, Observer...' },
    ],
    metaFields: [
      { key: 'year_from', label: 'From', type: 'number', placeholder: '2020' },
      { key: 'year_to', label: 'To', type: 'number', placeholder: 'Leave blank if current' },
    ],
    display: (p) => ({ main: p.label || 'Board Member', sub: p.value || '', meta: _yearRange(p.meta) }),
  },
  competition: {
    label: 'Competition',
    color: '#f97316',
    fields: [
      { key: 'label', label: 'Event / Competition', placeholder: 'National Hackathon...', required: true },
      { key: 'value', label: 'Result / Place', placeholder: '1st Place, Finalist, Gold...' },
    ],
    metaFields: [
      { key: 'year', label: 'Year', type: 'number', placeholder: '2023' },
    ],
    display: (p) => ({ main: p.label, sub: p.value || '', meta: p.meta && p.meta.year ? String(p.meta.year) : '' }),
  },
  custom: {
    label: 'Custom',
    color: '#6b7280',
    fields: [
      { key: 'label', label: 'Property Name', placeholder: 'Languages, Skills, Interests...', required: true },
      { key: 'value', label: 'Value', placeholder: 'The property value' },
    ],
    metaFields: [],
    display: (p) => ({ main: p.label, sub: p.value || '', meta: '' }),
  },
};

function _yearRange(meta) {
  if (!meta) return '';
  const from = meta.year_from;
  const to = meta.year_to;
  if (from && to) return `${from} \u2013 ${to}`;
  if (from) return `${from} \u2013 present`;
  if (to) return `Until ${to}`;
  return '';
}

function _dateRange(meta) {
  if (!meta) return '';
  const issued = meta.date_issued;
  const expires = meta.date_expires;
  if (issued && expires) return `Issued ${issued} \u00b7 Expires ${expires}`;
  if (issued) return `Issued ${issued}`;
  if (expires) return `Expires ${expires}`;
  return '';
}

const GENDER_LABELS = { m: 'Male', f: 'Female', o: 'Other' };

const profile = (() => {
  let currentPersonId = null;
  let personData = null;
  let propsData = [];
  let propModalId = null;

  async function load(id) {
    currentPersonId = id;
    personData = null;
    propsData = [];
    _showLoading();
    try {
      const [person, props, rels] = await Promise.all([
        apiGetPerson(id),
        apiGetPersonProperties(id),
        apiGetPersonRelationships(id),
      ]);
      personData = person;
      propsData = props || [];
      _renderHero();
      _renderProperties();
      _renderConnections(rels || []);
    } catch (err) {
      console.error('Profile load failed:', err);
    }
  }

  function _showLoading() {
    el('profile-name').textContent = 'Loading\u2026';
    el('profile-nickname').style.display = 'none';
    el('profile-title-text').style.display = 'none';
    clearChildren(el('profile-badges'));
    el('profile-notes-card').style.display = 'none';
    clearChildren(el('profile-properties'));
    clearChildren(el('profile-connections'));
    const nameEl = el('profile-header-name');
    if (nameEl) nameEl.textContent = '';
  }

  function _renderHero() {
    const d = personData;
    const nameEl = el('profile-header-name');
    if (nameEl) nameEl.textContent = d.name || '?';
    el('profile-name').textContent = d.name || '?';

    const nickEl = el('profile-nickname');
    if (d.nickname) {
      nickEl.textContent = `\u201c${d.nickname}\u201d`;
      nickEl.style.display = '';
    } else {
      nickEl.style.display = 'none';
    }

    const titleEl = el('profile-title-text');
    if (d.title) {
      titleEl.textContent = d.title;
      titleEl.style.display = '';
    } else {
      titleEl.style.display = 'none';
    }

    const badgesEl = el('profile-badges');
    clearChildren(badgesEl);
    if (d.gender) {
      const b = createEl('span', `profile-badge badge-gender-${d.gender}`);
      b.textContent = GENDER_LABELS[d.gender] || d.gender.toUpperCase();
      badgesEl.appendChild(b);
    }
    if (d.deceased) {
      const b = createEl('span', 'profile-badge badge-deceased');
      b.textContent = '\u2020 Deceased';
      badgesEl.appendChild(b);
    }

    const avatar = el('profile-avatar');
    avatar.style.background = d.deceased ? '#374151' : 'var(--node-person, #3b82f6)';

    const notesCard = el('profile-notes-card');
    if (d.notes) {
      el('profile-notes-text').textContent = d.notes;
      notesCard.style.display = '';
    } else {
      notesCard.style.display = 'none';
    }

    el('profile-edit-btn').onclick = () => openEditModal(currentPersonId);
  }

  function _renderProperties() {
    const list = el('profile-properties');
    clearChildren(list);
    if (!propsData.length) {
      const empty = createEl('div', 'properties-empty');
      empty.textContent = 'No properties yet. Click \u201c+ Add property\u201d to add titles, education, certificates, and more.';
      list.appendChild(empty);
      return;
    }
    const grouped = {};
    Object.keys(PROP_TYPES).forEach(t => { grouped[t] = []; });
    propsData.forEach(p => {
      if (!grouped[p.type]) grouped[p.type] = [];
      grouped[p.type].push(p);
    });
    Object.entries(grouped).forEach(([type, items]) => {
      if (!items.length) return;
      const typeDef = PROP_TYPES[type] || PROP_TYPES.custom;
      const group = createEl('div', 'prop-group');
      const groupHeader = createEl('div', 'prop-group-header');
      const dot = createEl('span', 'prop-group-dot');
      dot.style.background = typeDef.color;
      const label = createEl('span', 'prop-group-label');
      label.textContent = typeDef.label;
      label.style.color = typeDef.color;
      groupHeader.appendChild(dot);
      groupHeader.appendChild(label);
      group.appendChild(groupHeader);
      items.forEach(prop => group.appendChild(_buildPropCard(prop)));
      list.appendChild(group);
    });
  }

  function _buildPropCard(prop) {
    const typeDef = PROP_TYPES[prop.type] || PROP_TYPES.custom;
    const display = typeDef.display(prop);
    const card = createEl('div', 'property-card');
    card.style.borderLeftColor = typeDef.color;
    const actions = createEl('div', 'property-card-actions');
    const editBtn = createEl('button', 'prop-action-btn');
    editBtn.title = 'Edit';
    editBtn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="13" height="13"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>';
    editBtn.addEventListener('click', () => _openPropModal(prop.type, prop));
    const delBtn = createEl('button', 'prop-action-btn danger');
    delBtn.title = 'Delete';
    delBtn.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="13" height="13"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6M14 11v6"/><path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/></svg>';
    delBtn.addEventListener('click', () => _deleteProp(prop.id));
    actions.appendChild(editBtn);
    actions.appendChild(delBtn);
    card.appendChild(actions);
    if (display.main) { const m = createEl('div', 'property-card-main'); m.textContent = display.main; card.appendChild(m); }
    if (display.sub)  { const s = createEl('div', 'property-card-sub');  s.textContent = display.sub;  card.appendChild(s); }
    if (display.meta) { const t = createEl('div', 'property-card-meta'); t.textContent = display.meta; card.appendChild(t); }
    return card;
  }

  function _renderConnections(rels) {
    const container = el('profile-connections');
    clearChildren(container);
    if (!rels.length) {
      const empty = createEl('p', 'profile-conn-empty');
      empty.textContent = 'No connections yet.';
      container.appendChild(empty);
      return;
    }
    rels.forEach(rel => {
      const otherId = rel.source_id === currentPersonId ? rel.target_id : rel.source_id;
      const other = graph.getNode(otherId);
      if (!other) return;
      const item = createEl('div', 'profile-conn-item');
      const dot = createEl('div', 'profile-conn-dot');
      dot.style.background = getNodeColor(other.entityType);
      const info = createEl('div', 'profile-conn-info');
      const name = createEl('div', 'profile-conn-name');
      name.textContent = other.label || '?';
      const relType = createEl('div', 'profile-conn-rel');
      relType.textContent = REL_LABELS[rel.type] || rel.type;
      info.appendChild(name); info.appendChild(relType);
      item.appendChild(dot); item.appendChild(info);
      item.addEventListener('click', () => {
        if (other.entityType === 'person') {
          router.navigate('profile', { personId: other.id });
        } else {
          router.navigate('graph');
          setTimeout(() => graph.selectNode(other.id), 80);
        }
      });
      container.appendChild(item);
    });
  }

  function initPropModal() {
    const select = el('prop-type-select');
    clearChildren(select);
    Object.entries(PROP_TYPES).forEach(([key, def]) => {
      const opt = createEl('option', '');
      opt.value = key;
      opt.textContent = def.label;
      select.appendChild(opt);
    });
    select.addEventListener('change', (e) => { _renderPropFields(e.target.value, null); });
    el('prop-modal-close').addEventListener('click', _closePropModal);
    el('prop-modal-cancel').addEventListener('click', _closePropModal);
    el('prop-modal-overlay').addEventListener('click', (e) => {
      if (e.target === el('prop-modal-overlay')) _closePropModal();
    });
    el('prop-modal-save').addEventListener('click', _saveProp);
    el('prop-modal-delete').addEventListener('click', () => { if (propModalId) _deleteProp(propModalId); });
    el('profile-add-prop-btn').addEventListener('click', () => { _openPropModal('title', null); });
  }

  function _openPropModal(type, existingProp) {
    propModalId = existingProp ? existingProp.id : null;
    el('prop-modal-title').textContent = existingProp ? 'Edit Property' : 'Add Property';
    el('prop-modal-delete').style.display = existingProp ? 'inline-flex' : 'none';
    const select = el('prop-type-select');
    select.value = type || 'title';
    select.disabled = !!existingProp;
    _renderPropFields(type || 'title', existingProp);
    el('prop-modal-overlay').classList.remove('modal-hidden');
    const first = qs('.form-input', el('prop-dynamic-fields'));
    if (first) first.focus();
  }

  function _closePropModal() {
    el('prop-modal-overlay').classList.add('modal-hidden');
    propModalId = null;
    el('prop-type-select').disabled = false;
  }

  function _renderPropFields(type, data) {
    const container = el('prop-dynamic-fields');
    clearChildren(container);
    const typeDef = PROP_TYPES[type] || PROP_TYPES.custom;
    typeDef.fields.forEach(field => {
      const group = createEl('div', 'form-group');
      const label = createEl('label', 'form-label');
      label.textContent = field.label + (field.required ? ' *' : '');
      const input = createEl('input', 'form-input');
      input.type = field.type || 'text';
      input.placeholder = field.placeholder || '';
      input.dataset.propField = field.key;
      if (data && data[field.key] != null) input.value = data[field.key];
      group.appendChild(label); group.appendChild(input); container.appendChild(group);
    });
    if (!typeDef.metaFields || !typeDef.metaFields.length) return;
    let i = 0;
    while (i < typeDef.metaFields.length) {
      const f = typeDef.metaFields[i];
      const next = typeDef.metaFields[i + 1];
      if (f.type === 'number' && next && next.type === 'number') {
        const row = createEl('div', 'form-row');
        [f, next].forEach(mf => {
          const group = createEl('div', 'form-group');
          const label = createEl('label', 'form-label');
          label.textContent = mf.label;
          const input = createEl('input', 'form-input');
          input.type = 'number'; input.placeholder = mf.placeholder || '';
          input.dataset.propMeta = mf.key;
          if (data && data.meta && data.meta[mf.key] != null) input.value = data.meta[mf.key];
          group.appendChild(label); group.appendChild(input); row.appendChild(group);
        });
        container.appendChild(row); i += 2;
      } else {
        const group = createEl('div', 'form-group');
        const label = createEl('label', 'form-label');
        label.textContent = f.label;
        const input = createEl('input', 'form-input');
        input.type = f.type || 'text'; input.placeholder = f.placeholder || '';
        input.dataset.propMeta = f.key;
        if (data && data.meta && data.meta[f.key] != null) input.value = data.meta[f.key];
        group.appendChild(label); group.appendChild(input); container.appendChild(group); i++;
      }
    }
  }

  async function _saveProp() {
    const type = el('prop-type-select').value;
    const typeDef = PROP_TYPES[type] || PROP_TYPES.custom;
    const data = { label: null, value: null, meta: {} };
    qsa('[data-prop-field]', el('prop-dynamic-fields')).forEach(input => {
      data[input.dataset.propField] = input.value.trim() || null;
    });
    qsa('[data-prop-meta]', el('prop-dynamic-fields')).forEach(input => {
      const v = input.value.trim();
      if (v) data.meta[input.dataset.propMeta] = v;
    });
    if (!Object.keys(data.meta).length) data.meta = null;
    const reqField = typeDef.fields.find(f => f.required);
    const reqKey = reqField ? reqField.key : null;
    if (reqKey && !data[reqKey]) {
      const input = qs(`[data-prop-field="${reqKey}"]`, el('prop-dynamic-fields'));
      if (input) { input.focus(); input.style.borderColor = 'var(--error)'; setTimeout(() => { input.style.borderColor = ''; }, 1200); }
      return;
    }
    el('prop-modal-save').disabled = true;
    try {
      if (propModalId) {
        const result = await apiUpdatePersonProperty(propModalId, { label: data.label, value: data.value, meta: data.meta });
        const idx = propsData.findIndex(p => p.id === propModalId);
        if (idx !== -1) propsData[idx] = result;
      } else {
        const result = await apiCreatePersonProperty(currentPersonId, type, data.label, data.value, data.meta);
        propsData.push(result);
      }
      _renderProperties(); _closePropModal();
    } catch (err) {
      console.error('Property save failed:', err);
      alert('Could not save property.');
    } finally {
      el('prop-modal-save').disabled = false;
    }
  }

  async function _deleteProp(propId) {
    if (!confirm('Delete this property?')) return;
    try {
      await apiDeletePersonProperty(propId);
      propsData = propsData.filter(p => p.id !== propId);
      _renderProperties(); _closePropModal();
    } catch (err) {
      console.error('Property delete failed:', err);
      alert('Could not delete property.');
    }
  }

  function onPersonUpdated(nodeId) {
    if (nodeId !== currentPersonId) return;
    const node = graph.getNode(nodeId);
    if (node) personData = node.data;
    _renderHero();
  }

  return { load, initPropModal, onPersonUpdated };
})();
