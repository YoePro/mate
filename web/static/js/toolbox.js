// Toolbox: node-type palette and drag-to-add behaviour

let pendingEntityType = null;
let dragPhantom = null;

function initToolbox() {
  qsa('.tool-item').forEach(item => {
    const entityType = item.dataset.entity;

    item.addEventListener('click', () => {
      if (pendingEntityType === entityType) {
        cancelAddMode();
        return;
      }
      enterAddMode(entityType);
    });

    item.addEventListener('dragstart', (e) => {
      e.dataTransfer.setData('text/entity-type', entityType);
      e.dataTransfer.effectAllowed = 'copy';
      createDragPhantom(entityType, e.clientX, e.clientY);
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
    removeDragPhantom();
    const pos = canvas.screenToWorld(e.clientX, e.clientY);
    openAddModal(entityType, pos.x, pos.y);
  });

  workspace.addEventListener('click', (e) => {
    if (!pendingEntityType) return;
    if (e.target !== workspace && !e.target.closest('#workspace-empty')) return;
    const pos = canvas.screenToWorld(e.clientX, e.clientY);
    const type = pendingEntityType;
    cancelAddMode();
    openAddModal(type, pos.x, pos.y);
  });
}

function enterAddMode(entityType) {
  pendingEntityType = entityType;
  qsa('.tool-item').forEach(i => {
    i.classList.toggle('active', i.dataset.entity === entityType);
  });
  el('workspace').classList.add('adding-node');
}

function cancelAddMode() {
  pendingEntityType = null;
  qsa('.tool-item').forEach(i => i.classList.remove('active'));
  el('workspace').classList.remove('adding-node');
}

function createDragPhantom(entityType, x, y) {
  removeDragPhantom();
  dragPhantom = createEl('div', 'drag-phantom');
  dragPhantom.style.left = x + 'px';
  dragPhantom.style.top  = y + 'px';

  const dot = createEl('div', 'node-body');
  dot.style.setProperty('--node-color', getNodeColor(entityType));
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
