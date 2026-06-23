// Node rendering and drag behaviour

const NODE_COLORS = {
  person:      'var(--node-person)',
  company:     'var(--node-company)',
  association: 'var(--node-association)',
  school:      'var(--node-school)',
  location:    'var(--node-location)',
  tag:         'var(--node-tag)',
};

const NODE_ICONS = {
  person:      '<circle cx="12" cy="8" r="4"/><path d="M4 20c0-4 3.6-7 8-7s8 3 8 7"/>',
  company:     '<rect x="2" y="7" width="20" height="15" rx="1"/><path d="M16 22V12h-8v10"/><path d="M9 7V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v3"/>',
  association: '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>',
  school:      '<path d="M22 10v6M2 10l10-5 10 5-10 5z"/><path d="M6 12v5c3 3 9 3 12 0v-5"/>',
  location:    '<path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/>',
  tag:         '<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/>',
};

const ENTITY_LABELS = {
  person: 'Person',
  company: 'Company',
  association: 'Association',
  school: 'School',
  location: 'Location',
  tag: 'Tag',
};

function getNodeColor(entityType) {
  return NODE_COLORS[entityType] || 'var(--text-muted)';
}

function buildNodeEl(node) {
  const wrap = createEl('div', 'graph-node');
  wrap.id = `node-${node.id}`;
  wrap.dataset.id = node.id;
  wrap.style.left = node.x + 'px';
  wrap.style.top  = node.y + 'px';
  wrap.style.setProperty('--node-color', getNodeColor(node.entityType));

  if (node.data && node.data.deceased) wrap.classList.add('node-deceased');

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
}

function handleNodeClick(id, e) {
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

function updateNodeSelection(selectedId) {
  qsa('.graph-node').forEach(n => {
    n.classList.toggle('selected', n.dataset.id === selectedId);
  });
}

function updateLinkSource(sourceId) {
  qsa('.graph-node').forEach(n => {
    n.classList.toggle('link-source', n.dataset.id === sourceId);
  });
}

function updateEmptyState() {
  const empty = el('workspace-empty');
  if (graph.nodes.length > 0) {
    empty.classList.add('hidden');
  } else {
    empty.classList.remove('hidden');
  }
}
