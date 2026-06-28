// Node rendering and drag behaviour

const NODE_COLORS = {
  person:      'var(--node-person)',
  personMale:  'var(--node-person-male)',
  personFemale:'var(--node-person-female)',
  personOther: 'var(--node-person-other)',
  company:     'var(--node-company)',
  association: 'var(--node-association)',
  school:      'var(--node-school)',
  government:  'var(--node-government)',
  political_party: 'var(--node-political-party)',
  religious_organization: 'var(--node-religious-organization)',
  sports_club: 'var(--node-sports-club)',
  military_unit: 'var(--node-military-unit)',
  ngo:         'var(--node-ngo)',
  community:   'var(--node-community)',
  project:     'var(--node-project)',
  location:    'var(--node-location)',
  tag:         'var(--node-tag)',
};

const NODE_ICONS = {
  person:      '<circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 3.6-7 8-7s8 3 8 7"/>',
  company:     '<rect x="2" y="7" width="20" height="15" rx="1"/><path d="M16 22V12h-8v10"/><path d="M9 7V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v3"/>',
  association: '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
  school:      '<path d="M22 10v6M2 10l10-5 10 5-10 5z"/><path d="M6 12v5c3 3 9 3 12 0v-5"/>',
  government:  '<path d="M3 21h18"/><path d="M5 21V10h14v11"/><path d="M12 3 3 8h18z"/><path d="M8 10v11M12 10v11M16 10v11"/>',
  political_party: '<path d="M4 21V4"/><path d="M4 5h13l-2 4 2 4H4"/><path d="M9 9h3"/>',
  religious_organization: '<path d="M12 2v20"/><path d="M6 8h12"/><path d="M5 22h14"/><path d="M8 18c1-3 7-3 8 0"/>',
  sports_club: '<circle cx="12" cy="12" r="9"/><path d="M6 12h12"/><path d="M12 3c3 3 3 15 0 18"/><path d="M12 3c-3 3-3 15 0 18"/>',
  military_unit: '<path d="M12 2 4 5v6c0 5 3.5 9 8 11 4.5-2 8-6 8-11V5z"/><path d="M12 7v8"/><path d="M8 11h8"/>',
  ngo:         '<path d="M12 21s-7-4.5-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 11c0 5.5-7 10-7 10z"/><path d="M8 13h8"/>',
  community:   '<circle cx="12" cy="7" r="3"/><circle cx="6" cy="16" r="3"/><circle cx="18" cy="16" r="3"/><path d="M9 14l2-4M15 14l-2-4M9 16h6"/>',
  family:      '<circle cx="8" cy="8" r="3"/><circle cx="16" cy="8" r="3"/><circle cx="12" cy="15" r="3"/><path d="M3 21c0-3 2.2-5 5-5"/><path d="M21 21c0-3-2.2-5-5-5"/><path d="M7 22c0-3 2.2-5 5-5s5 2 5 5"/>',
  project:     '<path d="M4 5h16v14H4z"/><path d="M8 5V3h8v2"/><path d="M8 10h8"/><path d="M8 14h5"/>',
  location:    '<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>',
  tag:         '<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/>',
};

const ENTITY_LABELS = {
  person: 'Person',
  company: 'Company',
  association: 'Association',
  school: 'School',
  government: 'Government',
  political_party: 'Political Party',
  religious_organization: 'Religious Organization',
  sports_club: 'Sports Club',
  military_unit: 'Military Unit',
  ngo: 'NGO',
  community: 'Community',
  family: 'Family',
  project: 'Project',
  location: 'Location',
  tag: 'Tag',
};

function getNodeColor(entityType, data) {
  if (entityType === 'person') {
    if (data && data.gender === 'm') return NODE_COLORS.personMale;
    if (data && data.gender === 'f') return NODE_COLORS.personFemale;
    if (data && data.gender === 'o') return NODE_COLORS.personOther;
  }
  return NODE_COLORS[entityType] || 'var(--text-muted)';
}

function buildNodeEl(node) {
  const wrap = createEl('div', 'graph-node');
  wrap.id = `node-${node.id}`;
  wrap.dataset.id = node.id;
  wrap.style.left = node.x + 'px';
  wrap.style.top  = node.y + 'px';
  wrap.style.setProperty('--node-color', getNodeColor(node.entityType, node.data));

  if (node.data && node.data.deceased) wrap.classList.add('node-deceased');
  if (graph.isNodeLocked(node.id)) wrap.classList.add('node-locked');

  const card = createEl('div', 'node-card');

  const body = createEl('div', 'node-body');
  body.innerHTML = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">${NODE_ICONS[node.entityType] || ''}</svg>`;

  const label = createEl('div', 'node-label');
  label.textContent = node.label || '?';

  card.appendChild(body);
  card.appendChild(label);

  if (node.data && node.data.title) {
    const sub = createEl('div', 'node-sublabel');
    sub.textContent = node.data.title;
    card.appendChild(sub);
  }

  wrap.appendChild(card);
  attachNodeEvents(wrap, node);
  return wrap;
}

function attachNodeEvents(wrap, node) {
  let dragStarted = false;
  let startX, startY, startWorldX, startWorldY;
  let positionSaveTimer = null;

  wrap.addEventListener('mousedown', (e) => {
    if (e.button !== 0) return;
    e.stopPropagation();

    if (e.shiftKey) {
      handleLinkInteraction(node.id);
      return;
    }

    startX = e.clientX;
    startY = e.clientY;
    startWorldX = node.x;
    startWorldY = node.y;
    dragStarted = false;

    function onMove(ev) {
      if (graph.isNodeLocked(node.id)) return;
      const dx = ev.clientX - startX;
      const dy = ev.clientY - startY;
      if (!dragStarted && Math.hypot(dx, dy) < 4) return;
      dragStarted = true;
      const scale = canvas.getScale();
      const newX = startWorldX + dx / scale;
      const newY = startWorldY + dy / scale;
      wrap.style.left = newX + 'px';
      wrap.style.top  = newY + 'px';
      graph.updateNodePosition(node.id, newX, newY);
    }

    function onUp() {
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
      if (!dragStarted) {
        handleNodeClick(node.id, e);
      } else {
        clearTimeout(positionSaveTimer);
        if (!isTemporaryGraphMode()) {
          positionSaveTimer = setTimeout(() => {
            apiSavePosition(node.id, node.entityType, node.x, node.y);
          }, 400);
        }
      }
    }

    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  });

  wrap.addEventListener('dblclick', (e) => {
    e.stopPropagation();
    openEditModal(node.id);
  });

  wrap.addEventListener('contextmenu', (e) => {
    e.preventDefault();
    e.stopPropagation();
    openNodeContextMenu(node.id, e.clientX, e.clientY);
  });
}

function handleNodeClick(id, e) {
  if (e.ctrlKey || e.metaKey) {
    graph.toggleNodeSelection(id);
    return;
  }
  if (graph.linkSourceId && graph.linkSourceId !== id) {
    openLinkModal(graph.linkSourceId, id);
    return;
  }
  graph.selectNode(graph.selectedNodeId === id ? null : id);
}

function handleLinkInteraction(id) {
  if (!graph.linkSourceId) {
    graph.setLinkSource(id);
  } else if (graph.linkSourceId === id) {
    graph.setLinkSource(null);
  } else {
    openLinkModal(graph.linkSourceId, id);
  }
}

function renderAllNodes() {
  const layer = el('nodes-layer');
  clearChildren(layer);
  graph.nodes.forEach(node => {
    if (graph.isNodeHidden(node.id)) return;
    layer.appendChild(buildNodeEl(node));
  });
  updateEmptyState();
}

function renderNode(node) {
  const existing = el(`node-${node.id}`);
  const newEl = buildNodeEl(node);
  if (existing) {
    existing.replaceWith(newEl);
  } else {
    el('nodes-layer').appendChild(newEl);
  }
  updateEmptyState();
}

function removeNodeEl(id) {
  const existing = el(`node-${id}`);
  if (existing) existing.remove();
  updateEmptyState();
}

function updateNodeSelection() {
  qsa('.graph-node').forEach(n => {
    n.classList.toggle('selected', graph.isNodeSelected(n.dataset.id));
  });
}

function updateLinkSource(sourceId) {
  qsa('.graph-node').forEach(n => {
    n.classList.toggle('link-source', n.dataset.id === sourceId);
  });
}

function updateEmptyState() {
  const empty = el('workspace-empty');
  const visibleCount = graph.nodes.filter(node => !graph.isNodeHidden(node.id)).length;
  if (visibleCount > 0) {
    empty.classList.add('hidden');
  } else {
    empty.classList.remove('hidden');
  }
}
