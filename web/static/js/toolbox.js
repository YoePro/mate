// Toolbox: node-type palette and drag-to-add behaviour

let pendingEntityType = null;
let pendingEntityDefaults = null;
let dragPhantom = null;

function initToolbox() {
  qsa('[data-tool-action]').forEach(item => {
    item.addEventListener('click', () => runToolAction(item.dataset.toolAction));
  });

  qsa('[data-entity]').forEach(item => {
    const entityType = item.dataset.entity;
    if (!entityType) return;

    item.addEventListener('click', () => {
      const defaults = entityDefaultsFromElement(item);
      if (pendingEntityType === entityType && sameDefaults(pendingEntityDefaults, defaults)) {
        cancelAddMode();
        return;
      }
      enterAddMode(entityType, defaults);
    });

    item.addEventListener('dragstart', (e) => {
      const defaults = entityDefaultsFromElement(item);
      e.dataTransfer.setData('text/entity-type', entityType);
      if (defaults && defaults.gender) e.dataTransfer.setData('text/default-gender', defaults.gender);
      e.dataTransfer.effectAllowed = 'copy';
      createDragPhantom(entityType, e.clientX, e.clientY, defaults);
    });

    item.addEventListener('dragend', () => {
      removeDragPhantom();
    });
  });

  const workspace = el('workspace');

  workspace.addEventListener('dragover', (e) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
    if (dragPhantom) {
      dragPhantom.style.left = e.clientX + 'px';
      dragPhantom.style.top  = e.clientY + 'px';
    }
  });

  workspace.addEventListener('drop', (e) => {
    e.preventDefault();
    const entityType = e.dataTransfer.getData('text/entity-type');
    if (!entityType) return;
    const gender = e.dataTransfer.getData('text/default-gender');
    const defaults = gender ? { gender } : null;
    removeDragPhantom();
    const pos = canvas.screenToWorld(e.clientX, e.clientY);
    openAddModal(entityType, pos.x, pos.y, defaults);
  });

  workspace.addEventListener('click', (e) => {
    if (!pendingEntityType) return;
    if (e.target !== workspace && !e.target.closest('#workspace-empty')) return;
    const pos = canvas.screenToWorld(e.clientX, e.clientY);
    const type = pendingEntityType;
    const defaults = pendingEntityDefaults;
    cancelAddMode();
    openAddModal(type, pos.x, pos.y, defaults);
  });
}

function entityDefaultsFromElement(item) {
  const defaults = {};
  if (item.dataset.defaultGender) defaults.gender = item.dataset.defaultGender;
  return Object.keys(defaults).length ? defaults : null;
}

function sameDefaults(a, b) {
  const left = a || {};
  const right = b || {};
  return left.gender === right.gender;
}

function runToolAction(action) {
  switch (action) {
  case 'fit-all':
    canvas.fitToNodes(graph.nodes);
    break;
  case 'fit-selection':
    {
      const nodes = selectedNodes();
      if (nodes.length) canvas.fitToNodes(nodes);
    }
    break;
  case 'zoom-in':
    canvas.zoomBy(1.2);
    break;
  case 'zoom-out':
    canvas.zoomBy(0.8);
    break;
  case 'reset-zoom':
    canvas.resetZoom();
    break;
  case 'select-connected':
    selectConnectedNodes();
    break;
  case 'clear-selection':
    graph.selectNode(null);
    break;
  case 'hide-selected':
    hideSelectedNodes();
    break;
  case 'show-hidden':
    graph.showHiddenNodes();
    break;
  case 'auto-layout':
    autoLayoutVisibleNodes();
    break;
  }
}

function selectedNodes() {
  return graph.selectedNodeIds
    .map(id => graph.getNode(id))
    .filter(Boolean);
}

function selectConnectedNodes() {
  const baseIds = graph.selectedNodeIds.length ? graph.selectedNodeIds : (graph.selectedNodeId ? [graph.selectedNodeId] : []);
  if (!baseIds.length) return;
  const base = new Set(baseIds);
  const ids = new Set(baseIds);
  graph.links.forEach(link => {
    if (base.has(link.sourceId)) ids.add(link.targetId);
    if (base.has(link.targetId)) ids.add(link.sourceId);
  });
  graph.selectNodes(Array.from(ids));
  canvas.fitToNodes(Array.from(ids).map(id => graph.getNode(id)).filter(Boolean));
}

function hideSelectedNodes() {
  const ids = graph.selectedNodeIds.length ? graph.selectedNodeIds : (graph.selectedNodeId ? [graph.selectedNodeId] : []);
  if (!ids.length) return;
  graph.hideNodes(ids);
}

function visibleNodes() {
  return graph.nodes.filter(node => !graph.isNodeHidden(node.id));
}

function autoLayoutVisibleNodes() {
  const nodes = visibleNodes();
  if (!nodes.length) return;
  const cols = Math.ceil(Math.sqrt(nodes.length));
  const spacingX = 170;
  const spacingY = 140;
  const startX = 80;
  const startY = 80;

  nodes.forEach((node, index) => {
    const x = startX + (index % cols) * spacingX;
    const y = startY + Math.floor(index / cols) * spacingY;
    graph.updateNodePosition(node.id, x, y);
    if (!isTemporaryGraphMode()) {
      apiSavePosition(node.id, node.entityType, x, y);
    }
  });
  renderAllNodes();
  renderAllLinks();
  canvas.fitToNodes(nodes);
}

function enterAddMode(entityType, defaults) {
  pendingEntityType = entityType;
  pendingEntityDefaults = defaults || null;
  qsa('[data-entity]').forEach(i => {
    i.classList.toggle('active', i.dataset.entity === entityType && sameDefaults(entityDefaultsFromElement(i), pendingEntityDefaults));
  });
  el('workspace').classList.add('adding-node');
}

function cancelAddMode() {
  pendingEntityType = null;
  pendingEntityDefaults = null;
  qsa('[data-entity]').forEach(i => i.classList.remove('active'));
  el('workspace').classList.remove('adding-node');
}

function createDragPhantom(entityType, x, y, defaults) {
  removeDragPhantom();
  dragPhantom = createEl('div', 'drag-phantom');
  dragPhantom.style.left = x + 'px';
  dragPhantom.style.top  = y + 'px';

  const dot = createEl('div', 'node-body');
  dot.style.setProperty('--node-color', getNodeColor(entityType, defaults));
  dot.style.width  = '40px';
  dot.style.height = '40px';
  dot.innerHTML = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" style="width:18px;height:18px">${NODE_ICONS[entityType] || ''}</svg>`;
  dragPhantom.appendChild(dot);
  document.body.appendChild(dragPhantom);
}

function removeDragPhantom() {
  if (dragPhantom) {
    dragPhantom.remove();
    dragPhantom = null;
  }
}
