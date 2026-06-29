// Toolbox: node-type palette and drag-to-add behaviour

let pendingEntityType = null;
let pendingEntityDefaults = null;
let dragPhantom = null;
let selectionMarquee = null;

function initToolbox() {
  initToolboxSections();
  graph.on('selection-changed', updateDataToolState);
  graph.on('nodes-changed', updateDataToolState);

  qsa('[data-tool-action]').forEach(item => {
    item.addEventListener('click', () => runToolAction(item.dataset.toolAction));
  });
  updateToolboxAccess();
  updateDataToolState();

  qsa('[data-entity]').forEach(item => {
    const entityType = item.dataset.entity;
    if (!entityType) return;

    item.addEventListener('click', () => {
      if (!canWriteData()) return;
      const defaults = entityDefaultsFromElement(item);
      if (pendingEntityType === entityType && sameDefaults(pendingEntityDefaults, defaults)) {
        cancelAddMode();
        return;
      }
      enterAddMode(entityType, defaults);
    });

    item.addEventListener('dragstart', (e) => {
      if (!canWriteData()) {
        e.preventDefault();
        return;
      }
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

  workspace.addEventListener('mousedown', startSelectionMarquee);

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
    if (!canWriteData()) return;
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
    if (!canWriteData()) {
      cancelAddMode();
      return;
    }
    if (e.target !== workspace && !e.target.closest('#workspace-empty')) return;
    const pos = canvas.screenToWorld(e.clientX, e.clientY);
    const type = pendingEntityType;
    const defaults = pendingEntityDefaults;
    cancelAddMode();
    openAddModal(type, pos.x, pos.y, defaults);
  });
}

function initToolboxSections() {
  qsa('[data-toolbox-section]').forEach(section => {
    const key = section.dataset.toolboxSection;
    const toggle = qs('.toolbox-section-toggle', section);
    if (!key || !toggle) return;

    const collapsed = window.localStorage.getItem(`mate.toolbox.section.${key}`) === 'collapsed';
    setToolboxSectionCollapsed(section, collapsed);

    toggle.addEventListener('click', () => {
      setToolboxSectionCollapsed(section, !section.classList.contains('is-collapsed'));
    });
  });
}

function setToolboxSectionCollapsed(section, collapsed) {
  const key = section.dataset.toolboxSection;
  const toggle = qs('.toolbox-section-toggle', section);
  section.classList.toggle('is-collapsed', collapsed);
  if (toggle) toggle.setAttribute('aria-expanded', collapsed ? 'false' : 'true');
  if (key) {
    window.localStorage.setItem(`mate.toolbox.section.${key}`, collapsed ? 'collapsed' : 'expanded');
  }
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
  case 'select-same-type':
    selectSameNodeType();
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
    if (!canWriteData()) return;
    autoLayoutVisibleNodes();
    break;
  case 'lock-selected':
    lockSelectedNodes();
    break;
  case 'unlock-selected':
    unlockSelectedNodes();
    break;
  case 'edit-selected':
    if (!canWriteData()) return;
    editSelectedNode();
    break;
  case 'delete-selected':
    if (!canWriteData()) return;
    deleteSelectedNode();
    break;
  }
}

function updateToolboxAccess() {
  const canWrite = canWriteData();
  qsa('[data-entity]').forEach(item => {
    item.disabled = !canWrite || item.classList.contains('tool-icon-disabled');
  });
  const autoLayout = qs('[data-tool-action="auto-layout"]');
  if (autoLayout) autoLayout.disabled = !canWrite;
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

function nodeSelectionType(node) {
  if (!node) return '';
  if (typeof isOrganizationEntityType === 'function' && isOrganizationEntityType(node.entityType)) return 'organization';
  return node.entityType;
}

function selectSameNodeType(anchorId) {
  const anchors = anchorId
    ? [graph.getNode(anchorId)].filter(Boolean)
    : selectedNodes();
  if (!anchors.length) return;
  const typeKeys = new Set(anchors.map(nodeSelectionType).filter(Boolean));
  const ids = visibleNodes()
    .filter(node => typeKeys.has(nodeSelectionType(node)))
    .map(node => node.id);
  graph.selectNodes(ids);
  if (ids.length) canvas.fitToNodes(ids.map(id => graph.getNode(id)).filter(Boolean));
}

function hideSelectedNodes() {
  const ids = graph.selectedNodeIds.length ? graph.selectedNodeIds : (graph.selectedNodeId ? [graph.selectedNodeId] : []);
  if (!ids.length) return;
  graph.hideNodes(ids);
}

function selectedNodeIds() {
  return graph.selectedNodeIds.length ? graph.selectedNodeIds : (graph.selectedNodeId ? [graph.selectedNodeId] : []);
}

function primarySelectedNodeId() {
  const ids = selectedNodeIds();
  return ids.length ? ids[0] : null;
}

function updateDataToolState() {
  const canWrite = canWriteData();
  const hasSelection = Boolean(primarySelectedNodeId());
  qsa('[data-requires-selection]').forEach(item => {
    item.disabled = !canWrite || !hasSelection;
  });
}

function editSelectedNode() {
  const id = primarySelectedNodeId();
  if (!id) return;
  openEditModal(id);
}

function deleteSelectedNode() {
  const id = primarySelectedNodeId();
  if (!id) return;
  deleteNodeWithDialog(id);
}

function lockSelectedNodes() {
  const ids = selectedNodeIds();
  if (!ids.length) return;
  graph.lockNodes(ids);
}

function unlockSelectedNodes() {
  const ids = selectedNodeIds();
  if (!ids.length) return;
  graph.unlockNodes(ids);
}

function visibleNodes() {
  return graph.nodes.filter(node => !graph.isNodeHidden(node.id));
}

function startSelectionMarquee(e) {
  if (e.button !== 0 || pendingEntityType || graph.linkSourceId) return;
  if (e.target !== el('workspace') && !e.target.closest('#workspace-empty')) return;
  if (e.shiftKey) return;

  const workspace = el('workspace');
  const rect = workspace.getBoundingClientRect();
  const startX = e.clientX;
  const startY = e.clientY;
  let active = false;

  function ensureMarquee() {
    if (selectionMarquee) return;
    selectionMarquee = createEl('div', 'selection-marquee');
    workspace.appendChild(selectionMarquee);
  }

  function updateMarquee(ev) {
    const dx = ev.clientX - startX;
    const dy = ev.clientY - startY;
    if (!active && Math.hypot(dx, dy) < 5) return;
    active = true;
    ensureMarquee();
    const left = Math.min(startX, ev.clientX) - rect.left;
    const top = Math.min(startY, ev.clientY) - rect.top;
    const width = Math.abs(dx);
    const height = Math.abs(dy);
    selectionMarquee.style.left = left + 'px';
    selectionMarquee.style.top = top + 'px';
    selectionMarquee.style.width = width + 'px';
    selectionMarquee.style.height = height + 'px';
  }

  function finishMarquee(ev) {
    document.removeEventListener('mousemove', updateMarquee);
    document.removeEventListener('mouseup', finishMarquee);
    if (selectionMarquee) {
      selectionMarquee.remove();
      selectionMarquee = null;
    }
    if (!active) return;

    const a = canvas.screenToWorld(startX, startY);
    const b = canvas.screenToWorld(ev.clientX, ev.clientY);
    const minX = Math.min(a.x, b.x);
    const maxX = Math.max(a.x, b.x);
    const minY = Math.min(a.y, b.y);
    const maxY = Math.max(a.y, b.y);
    const ids = visibleNodes()
      .filter(node => node.x >= minX && node.x <= maxX && node.y >= minY && node.y <= maxY)
      .map(node => node.id);

    if (e.ctrlKey || e.metaKey) {
      graph.selectNodes(Array.from(new Set(graph.selectedNodeIds.concat(ids))));
    } else {
      graph.selectNodes(ids);
    }
  }

  document.addEventListener('mousemove', updateMarquee);
  document.addEventListener('mouseup', finishMarquee);
}

function autoLayoutVisibleNodes() {
  const nodes = visibleNodes().filter(node => !graph.isNodeLocked(node.id));
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
