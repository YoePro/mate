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

function destructiveActionLabel(entityType) {
  return 'Delete';
}

function trashIconMarkup() {
  return '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/></svg>';
}

function setDeleteButton(button, label) {
  if (!button) return;
  button.innerHTML = `${trashIconMarkup()}<span>${label || 'Delete'}</span>`;
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
let pendingDisableAccountId = null;
let networkModalMode = 'create';
let pendingNetworkArchive = false;
let confirmDialogResolve = null;
let duplicateDialogResolve = null;
let duplicateSelectedPersonId = null;
let duplicateMatches = [];
let activeContextMenu = null;
let lastLoadedNetworkId = null;
window.currentNetworkId = null;

const ACCOUNT_ROLES = [
  ['owner', 'Owner'],
  ['admin', 'Admin'],
  ['editor', 'Editor'],
  ['viewer', 'Viewer'],
];

const COLOR_PREFERENCES = [
  ['--node-person-male', 'Male person', '#3b82f6'],
  ['--node-person-female', 'Female person', '#ec4899'],
  ['--node-person-other', 'Other person', '#64748b'],
  ['--node-company', 'Company', '#10b981'],
  ['--node-association', 'Association', '#f59e0b'],
  ['--node-school', 'School', '#ef4444'],
  ['--node-government', 'Government', '#6366f1'],
  ['--node-project', 'Project', '#22c55e'],
];

async function boot() {
  currentAccount = await requireAccount();
  if (!currentAccount) return;
  applyStoredColorPreferences();

  await initNetworks();
  initToolbox();
  initInspector();
  initConfirmDialog();
  initDuplicateDialog();
  initContextMenus();
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
  const networkId = window.currentNetworkId;
  const networkChanged = networkId !== lastLoadedNetworkId;
  setStatus('loading');
  try {
    if (!networkId) {
      currentCustomRelationshipTypes = [];
      graph.clear();
      renderAllNodes();
      renderAllLinks();
      if (networkChanged || !canvas.hasUserViewport()) canvas.resetZoom({ user: false });
      lastLoadedNetworkId = networkId;
      setStatus('ok');
      return;
    }
    const data = await apiLoadAll();
    currentCustomRelationshipTypes = data.custom_relationship_types || [];
    graph.load(data);
    renderAllNodes();
    renderAllLinks();
    frameLoadedGraph(networkChanged);
    lastLoadedNetworkId = networkId;
    setStatus('ok');
  } catch (err) {
    console.error('Failed to load graph:', err);
    setStatus('error');
  }
}

function frameLoadedGraph(networkChanged) {
  const visible = graph.nodes.filter(node => !graph.isNodeHidden(node.id));
  if (!networkChanged && canvas.hasUserViewport()) return;
  if (!visible.length) {
    canvas.resetZoom({ user: false });
    return;
  }
  window.requestAnimationFrame(() => {
    canvas.fitToNodes(visible, { user: false });
  });
}

async function initNetworks() {
  const select = el('network-select');
  const createBtn = el('btn-new-network');
  const renameBtn = el('btn-rename-network');
  createBtn.addEventListener('click', () => openNetworkModal('create'));
  renameBtn.addEventListener('click', () => openNetworkModal('rename'));
  initNetworkActionsMenu();
  initNetworkModal();
  select.addEventListener('change', async () => {
    const selected = ownedNetworks.find(n => n.id === select.value);
    setCurrentNetwork(selected || null);
    await loadGraph();
  });
  initNetworkSearch();
  await refreshNetworks();
}

function initNetworkActionsMenu() {
  const trigger = el('btn-network-actions');
  const menu = el('network-actions-menu');
  if (!trigger || !menu) return;

  trigger.addEventListener('click', (event) => {
    event.stopPropagation();
    const open = menu.classList.contains('hidden');
    menu.classList.toggle('hidden', !open);
    trigger.setAttribute('aria-expanded', open ? 'true' : 'false');
    updateNetworkActions();
  });

  menu.addEventListener('click', (event) => {
    const action = event.target && event.target.dataset ? event.target.dataset.networkAction : '';
    if (!action || event.target.disabled) return;
    menu.classList.add('hidden');
    trigger.setAttribute('aria-expanded', 'false');
    if (action === 'create') openNetworkModal('create');
    if (action === 'rename') openNetworkModal('rename');
    if (action === 'metadata') openNetworkModal('metadata');
    if (action === 'archive') openNetworkModal('archive');
  });

  document.addEventListener('click', (event) => {
    if (!menu.classList.contains('hidden') && !menu.contains(event.target) && event.target !== trigger) {
      menu.classList.add('hidden');
      trigger.setAttribute('aria-expanded', 'false');
    }
  });
}

function initNetworkModal() {
  const overlay = el('network-modal-overlay');
  if (!overlay) return;
  el('network-modal-close').addEventListener('click', closeNetworkModal);
  el('network-modal-cancel').addEventListener('click', closeNetworkModal);
  el('network-modal-save').addEventListener('click', handleNetworkModalSave);
  el('network-modal-danger').addEventListener('click', handleNetworkModalArchive);
  overlay.addEventListener('click', (event) => {
    if (event.target === overlay) closeNetworkModal();
  });
}

function openNetworkModal(mode) {
  if (currentAccount && currentAccount.role === 'viewer') return;
  if ((mode === 'rename' || mode === 'metadata' || mode === 'archive') && !currentNetwork) return;

  networkModalMode = mode;
  pendingNetworkArchive = false;
  const overlay = el('network-modal-overlay');
  const title = el('network-modal-title');
  const name = el('network-modal-name');
  const description = el('network-modal-description');
  const save = el('network-modal-save');
  const danger = el('network-modal-danger');
  const fields = el('network-form-fields');

  setNetworkModalMessage('');
  danger.style.display = 'none';
  save.style.display = '';
  fields.style.display = '';
  name.disabled = false;
  description.disabled = false;

  if (mode === 'create') {
    title.textContent = 'Create Network';
    name.value = '';
    description.value = '';
    save.textContent = 'Create';
  } else if (mode === 'archive') {
    title.textContent = 'Delete Network';
    name.value = currentNetwork.name || '';
    description.value = currentNetwork.description || '';
    name.disabled = true;
    description.disabled = true;
    save.style.display = 'none';
    danger.style.display = '';
    setDeleteButton(danger, 'Delete');
    setNetworkModalMessage('Deleting a network hides it from your network list. Global person identities are not deleted.', 'error');
  } else {
    title.textContent = mode === 'metadata' ? 'Edit Network Metadata' : 'Rename Network';
    name.value = currentNetwork.name || '';
    description.value = currentNetwork.description || '';
    save.textContent = 'Save';
  }

  overlay.classList.remove('modal-hidden');
  if (mode !== 'archive') name.focus();
}

function closeNetworkModal() {
  el('network-modal-overlay').classList.add('modal-hidden');
  pendingNetworkArchive = false;
}

function setNetworkModalMessage(message, type) {
  const box = el('network-modal-message');
  if (!box) return;
  box.textContent = message || '';
  box.classList.toggle('is-error', type === 'error');
  box.classList.toggle('is-success', type === 'success');
}

async function handleNetworkModalSave() {
  const name = el('network-modal-name').value.trim();
  const description = el('network-modal-description').value.trim();
  if (!name) {
    setNetworkModalMessage('Network name is required.', 'error');
    el('network-modal-name').focus();
    return;
  }

  el('network-modal-save').disabled = true;
  try {
    if (networkModalMode === 'create') {
      const network = await apiCreateNetwork({ name, description });
      ownedNetworks = await apiListNetworks();
      renderNetworkOptions(ownedNetworks, network.id);
      setCurrentNetwork(network);
      graph.selectNode(null);
      await loadGraph();
      closeNetworkModal();
      return;
    }

    if (!currentNetwork) return;
    const updated = await apiUpdateNetwork(currentNetwork.id, { name, description });
    ownedNetworks = ownedNetworks.map(network => network.id === updated.id ? updated : network);
    renderNetworkOptions(ownedNetworks, updated.id);
    setCurrentNetwork(updated);
    closeNetworkModal();
  } catch (err) {
    console.error('Network save failed:', err);
    setNetworkModalMessage('Could not save network.', 'error');
  } finally {
    el('network-modal-save').disabled = false;
  }
}

async function handleNetworkModalArchive() {
  if (!currentNetwork) return;
  if (!pendingNetworkArchive) {
    pendingNetworkArchive = true;
    setDeleteButton(el('network-modal-danger'), 'Confirm delete');
    setNetworkModalMessage(`Click Confirm delete to delete "${currentNetwork.name}". Global person identities will remain.`, 'error');
    return;
  }

  el('network-modal-danger').disabled = true;
  try {
    await apiArchiveNetwork(currentNetwork.id);
    window.localStorage.removeItem('mate.currentNetworkId');
    await refreshNetworks();
    graph.selectNode(null);
    await loadGraph();
    closeNetworkModal();
  } catch (err) {
    console.error('Network archive failed:', err);
    setNetworkModalMessage('Could not delete network.', 'error');
  } finally {
    el('network-modal-danger').disabled = false;
  }
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
  const actionsBtn = el('btn-network-actions');
  const canWrite = currentAccount && currentAccount.role !== 'viewer';
  if (createBtn) createBtn.disabled = !canWrite;
  if (renameBtn) renameBtn.disabled = !canWrite || !currentNetwork;
  if (actionsBtn) actionsBtn.disabled = !canWrite;
  qsa('[data-network-action]').forEach(item => {
    const action = item.dataset.networkAction;
    const needsCurrent = action === 'rename' || action === 'metadata' || action === 'archive';
    if (action === 'manage') return;
    item.disabled = !canWrite || (needsCurrent && !currentNetwork);
  });
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

function initConfirmDialog() {
  el('confirm-modal-close').addEventListener('click', () => closeConfirmDialog(false));
  el('confirm-modal-cancel').addEventListener('click', () => closeConfirmDialog(false));
  el('confirm-modal-confirm').addEventListener('click', () => closeConfirmDialog(true));
  el('confirm-modal-overlay').addEventListener('click', (event) => {
    if (event.target === el('confirm-modal-overlay')) closeConfirmDialog(false);
  });
}

function openConfirmDialog(options) {
  const overlay = el('confirm-modal-overlay');
  const confirmButton = el('confirm-modal-confirm');
  const cancelButton = el('confirm-modal-cancel');
  el('confirm-modal-title').textContent = options.title || 'Confirm';
  el('confirm-modal-message').textContent = options.message || '';
  const confirmLabel = options.confirmLabel || 'Confirm';
  if (confirmLabel.toLowerCase().includes('delete')) {
    setDeleteButton(confirmButton, confirmLabel);
  } else {
    confirmButton.textContent = confirmLabel;
  }
  cancelButton.textContent = options.cancelLabel || 'Cancel';
  cancelButton.style.display = options.showCancel === false ? 'none' : '';
  confirmButton.classList.toggle('btn-danger', options.danger !== false);
  confirmButton.classList.toggle('btn-primary', options.danger === false);
  overlay.classList.remove('modal-hidden');
  confirmButton.focus();
  return new Promise(resolve => {
    confirmDialogResolve = resolve;
  });
}

function closeConfirmDialog(result) {
  const overlay = el('confirm-modal-overlay');
  overlay.classList.add('modal-hidden');
  if (confirmDialogResolve) {
    const resolve = confirmDialogResolve;
    confirmDialogResolve = null;
    resolve(Boolean(result));
  }
}

async function openMessageDialog(title, message) {
  await openConfirmDialog({
    title,
    message,
    confirmLabel: 'OK',
    danger: false,
    showCancel: false,
  });
}

function initDuplicateDialog() {
  el('duplicate-modal-close').addEventListener('click', () => closeDuplicateDialog({ cancelled: true }));
  el('duplicate-create-new').addEventListener('click', () => closeDuplicateDialog({ createNew: true }));
  el('duplicate-use-selected').addEventListener('click', () => {
    const match = duplicateMatches.find(item => item.person && item.person.id === duplicateSelectedPersonId);
    closeDuplicateDialog(match ? { person: match.person } : { cancelled: true });
  });
  el('duplicate-modal-overlay').addEventListener('click', (event) => {
    if (event.target === el('duplicate-modal-overlay')) closeDuplicateDialog({ cancelled: true });
  });
}

function openDuplicateDialog(matches) {
  duplicateMatches = matches || [];
  duplicateSelectedPersonId = duplicateMatches[0] && duplicateMatches[0].person ? duplicateMatches[0].person.id : null;
  renderDuplicateMatches();
  el('duplicate-use-selected').disabled = !duplicateSelectedPersonId;
  el('duplicate-modal-overlay').classList.remove('modal-hidden');
  el('duplicate-use-selected').focus();
  return new Promise(resolve => {
    duplicateDialogResolve = resolve;
  });
}

function closeDuplicateDialog(result) {
  el('duplicate-modal-overlay').classList.add('modal-hidden');
  duplicateMatches = [];
  duplicateSelectedPersonId = null;
  if (duplicateDialogResolve) {
    const resolve = duplicateDialogResolve;
    duplicateDialogResolve = null;
    resolve(result || { cancelled: true });
  }
}

function renderDuplicateMatches() {
  const list = el('duplicate-match-list');
  clearChildren(list);
  duplicateMatches.forEach(match => {
    const person = match.person || {};
    const card = createEl('button', 'duplicate-match-card', { type: 'button' });
    card.classList.toggle('selected', person.id === duplicateSelectedPersonId);
    card.addEventListener('click', () => {
      duplicateSelectedPersonId = person.id;
      el('duplicate-use-selected').disabled = !duplicateSelectedPersonId;
      renderDuplicateMatches();
    });

    const head = createEl('div', 'duplicate-match-head');
    const name = createEl('div', 'duplicate-match-name', { text: person.name || person.nickname || 'Unnamed person' });
    const confidence = createEl('div', 'duplicate-match-confidence', { text: `${Math.round((match.confidence || 0) * 100)}% match` });
    head.appendChild(name);
    head.appendChild(confidence);
    card.appendChild(head);

    const meta = createEl('div', 'duplicate-match-meta');
    [
      ['Nickname', person.nickname],
      ['Gender', person.gender],
      ['Title', person.title],
      ['Status', person.deceased ? 'Deceased' : ''],
      ['Tags', (person.tags || []).join(', ')],
      ['ID', person.id],
    ].forEach(([label, value]) => {
      if (!value) return;
      meta.appendChild(createEl('span', '', { text: `${label}: ${value}` }));
    });
    if (meta.children.length) card.appendChild(meta);

    if (person.description) {
      card.appendChild(createEl('div', 'duplicate-match-description', { text: person.description }));
    }
    const reasons = (match.reasons || []).join(', ');
    if (reasons) {
      card.appendChild(createEl('div', 'duplicate-match-reasons', { text: `Reasons: ${reasons}` }));
    }
    list.appendChild(card);
  });
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

function initContextMenus() {
  el('workspace').addEventListener('contextmenu', (event) => {
    if (event.target.closest('.graph-node') || event.target.closest('[data-link-group]')) return;
    event.preventDefault();
    openCanvasContextMenu(event.clientX, event.clientY);
  });

  document.addEventListener('click', closeContextMenu);
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') closeContextMenu();
  });
}

function openContextMenu(items, x, y) {
  closeContextMenu();
  const menu = createEl('div', 'context-menu');
  items.forEach(item => {
    if (item.separator) {
      menu.appendChild(createEl('div', 'context-menu-separator'));
      return;
    }
    const button = createEl('button', 'context-menu-item', { type: 'button' });
    if (item.icon === 'trash') {
      button.innerHTML = `${trashIconMarkup()}<span>${item.label}</span>`;
    } else {
      button.textContent = item.label;
    }
    button.disabled = Boolean(item.disabled);
    button.classList.toggle('danger', Boolean(item.danger));
    button.addEventListener('click', async (event) => {
      event.stopPropagation();
      closeContextMenu();
      if (!item.disabled && item.action) await item.action();
    });
    menu.appendChild(button);
  });
  document.body.appendChild(menu);

  const rect = menu.getBoundingClientRect();
  const left = Math.min(x, window.innerWidth - rect.width - 8);
  const top = Math.min(y, window.innerHeight - rect.height - 8);
  menu.style.left = Math.max(8, left) + 'px';
  menu.style.top = Math.max(8, top) + 'px';
  activeContextMenu = menu;
}

function closeContextMenu() {
  if (activeContextMenu) {
    activeContextMenu.remove();
    activeContextMenu = null;
  }
}

function openNodeContextMenu(nodeId, x, y) {
  const node = graph.getNode(nodeId);
  if (!node) return;
  graph.selectNode(nodeId);
  const action = destructiveActionLabel(node.entityType);
  const locked = graph.isNodeLocked(nodeId);
  openContextMenu([
    { label: 'Edit', action: () => openEditModal(nodeId) },
    { label: 'Create relationship from node', action: () => graph.setLinkSource(nodeId) },
    { label: 'Fit selection', action: () => canvas.fitToNodes([node]) },
    { label: 'Select same type', action: () => selectSameNodeType(nodeId) },
    { separator: true },
    { label: 'Hide', action: () => graph.hideNodes([nodeId]) },
    { label: 'Duplicate', disabled: true },
    { label: 'Lock', disabled: locked, action: () => graph.lockNodes([nodeId]) },
    { label: 'Unlock', disabled: !locked, action: () => graph.unlockNodes([nodeId]) },
    { separator: true },
    { label: action, icon: 'trash', danger: true, action: () => deleteNodeWithDialog(nodeId) },
  ], x, y);
}

function openRelationshipContextMenu(linkId, x, y) {
  const link = graph.getLink(linkId);
  if (!link) return;
  const source = graph.getNode(link.sourceId);
  const target = graph.getNode(link.targetId);
  openContextMenu([
    { label: 'Edit relationship', action: () => openRelationshipEditModal(linkId) },
    { label: 'Select connected nodes', action: () => {
      graph.selectNodes([link.sourceId, link.targetId]);
      canvas.fitToNodes([source, target].filter(Boolean));
    } },
    { label: 'Show connected nodes', action: () => {
      graph.selectNodes([link.sourceId, link.targetId]);
      canvas.fitToNodes([source, target].filter(Boolean));
    } },
    { separator: true },
    { label: 'Delete relationship', icon: 'trash', danger: true, action: () => promptDeleteLink(linkId) },
  ], x, y);
}

function openCanvasContextMenu(x, y) {
  const pos = canvas.screenToWorld(x, y);
  openContextMenu([
    { label: 'Create person here', action: () => openAddModal('person', pos.x, pos.y) },
    { label: 'Create organization here', disabled: true },
    { label: 'Paste', disabled: true },
    { separator: true },
    { label: 'Fit all nodes', action: () => canvas.fitToNodes(graph.nodes.filter(node => !graph.isNodeHidden(node.id))) },
    { label: 'Reset zoom', action: () => canvas.resetZoom() },
    { label: 'Show hidden', action: () => graph.showHiddenNodes() },
    { label: 'Clear selection', action: () => graph.selectNode(null) },
  ], x, y);
}

function canOpenAccountAdmin() {
  return currentAccount && ['owner', 'admin'].includes(currentAccount.role);
}

function accountRoleOptionsForActor() {
  if (currentAccount && currentAccount.role === 'owner') return ACCOUNT_ROLES;
  return ACCOUNT_ROLES.filter(([role]) => role === 'editor' || role === 'viewer');
}

function canActorManageAccount(account) {
  if (!currentAccount || !account) return false;
  if (currentAccount.role === 'owner') return true;
  if (currentAccount.role === 'admin') return account.role === 'editor' || account.role === 'viewer';
  return false;
}

async function handleLogout() {
  const trigger = el('account-menu-trigger');
  if (trigger) trigger.disabled = true;
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
  initAccountMenu();
  initAccountAdmin();
  initPreferencesModal();

  el('btn-fit').addEventListener('click', () => {
    canvas.fitToNodes(graph.nodes);
  });

  el('btn-reload').addEventListener('click', async () => {
    graph.selectNode(null);
    await loadGraph();
  });
}

function initAccountMenu() {
  const trigger = el('account-menu-trigger');
  const menu = el('account-menu');
  if (!trigger || !menu) return;

  qsa('.admin-only', menu).forEach(item => item.classList.toggle('hidden', !canOpenAccountAdmin()));
  qsa('.owner-only', menu).forEach(item => item.classList.toggle('hidden', !currentAccount || currentAccount.role !== 'owner'));

  trigger.addEventListener('click', (event) => {
    event.stopPropagation();
    const open = menu.classList.contains('hidden');
    menu.classList.toggle('hidden', !open);
    trigger.setAttribute('aria-expanded', open ? 'true' : 'false');
  });

  menu.addEventListener('click', async (event) => {
    const action = event.target && event.target.dataset ? event.target.dataset.accountAction : '';
    if (!action) return;
    menu.classList.add('hidden');
    trigger.setAttribute('aria-expanded', 'false');
    if (action === 'logout') {
      await handleLogout();
    } else if (action === 'admin') {
      openAccountAdmin();
    } else if (action === 'preferences') {
      openPreferencesModal();
    } else if (action === 'profile' || action === 'system') {
      setAccountAdminMessage('This area is planned for a later 0.13 step.', 'success');
    }
  });

  document.addEventListener('click', (event) => {
    if (!menu.classList.contains('hidden') && !menu.contains(event.target) && event.target !== trigger) {
      menu.classList.add('hidden');
      trigger.setAttribute('aria-expanded', 'false');
    }
  });
}

function colorPreferenceStorageKey() {
  const accountKey = currentAccount && (currentAccount.id || currentAccount.email || currentAccount.display_name);
  return `mate.colorPreferences.${accountKey || 'default'}`;
}

function storedColorPreferences() {
  try {
    return JSON.parse(window.localStorage.getItem(colorPreferenceStorageKey()) || '{}');
  } catch (err) {
    console.warn('Could not parse color preferences:', err.message);
    return {};
  }
}

function applyColorPreferences(values) {
  const root = document.documentElement;
  COLOR_PREFERENCES.forEach(([variable, , fallback]) => {
    root.style.setProperty(variable, values && values[variable] ? values[variable] : fallback);
  });
  renderAllNodes();
}

function applyStoredColorPreferences() {
  applyColorPreferences(storedColorPreferences());
}

function initPreferencesModal() {
  const overlay = el('preferences-overlay');
  if (!overlay) return;
  el('preferences-close').addEventListener('click', closePreferencesModal);
  el('preferences-cancel').addEventListener('click', closePreferencesModal);
  el('preferences-save').addEventListener('click', savePreferencesModal);
  el('preferences-reset').addEventListener('click', resetPreferencesModal);
  overlay.addEventListener('click', (event) => {
    if (event.target === overlay) closePreferencesModal();
  });
}

function openPreferencesModal() {
  const list = el('color-preference-list');
  clearChildren(list);
  const stored = storedColorPreferences();
  COLOR_PREFERENCES.forEach(([variable, label, fallback]) => {
    const row = createEl('div', 'color-preference-row');
    const labelWrap = createEl('label', 'color-preference-label');
    const swatch = createEl('span', 'color-preference-swatch');
    const text = createEl('span', '');
    const input = createEl('input', 'color-preference-input');
    const value = stored[variable] || fallback;
    input.type = 'color';
    input.value = value;
    input.dataset.colorVariable = variable;
    swatch.style.setProperty('--swatch-color', value);
    text.textContent = label;
    input.addEventListener('input', () => {
      swatch.style.setProperty('--swatch-color', input.value);
    });
    labelWrap.appendChild(swatch);
    labelWrap.appendChild(text);
    row.appendChild(labelWrap);
    row.appendChild(input);
    list.appendChild(row);
  });
  el('preferences-overlay').classList.remove('modal-hidden');
}

function closePreferencesModal() {
  el('preferences-overlay').classList.add('modal-hidden');
}

function savePreferencesModal() {
  const values = {};
  qsa('[data-color-variable]', el('color-preference-list')).forEach(input => {
    values[input.dataset.colorVariable] = input.value;
  });
  window.localStorage.setItem(colorPreferenceStorageKey(), JSON.stringify(values));
  applyColorPreferences(values);
  closePreferencesModal();
}

function resetPreferencesModal() {
  window.localStorage.removeItem(colorPreferenceStorageKey());
  applyColorPreferences({});
  openPreferencesModal();
}

function initAccountAdmin() {
  const overlay = el('account-admin-overlay');
  if (!overlay) return;
  el('account-admin-close').addEventListener('click', closeAccountAdmin);
  el('account-refresh').addEventListener('click', refreshAccountAdmin);
  el('account-create-submit').addEventListener('click', handleAccountCreate);
  overlay.addEventListener('click', (event) => {
    if (event.target === overlay) closeAccountAdmin();
  });
  renderAccountCreateRoles();
}

function openAccountAdmin() {
  if (!canOpenAccountAdmin()) return;
  pendingDisableAccountId = null;
  renderAccountCreateRoles();
  setAccountAdminMessage('');
  el('account-admin-overlay').classList.remove('modal-hidden');
  refreshAccountAdmin();
}

function closeAccountAdmin() {
  el('account-admin-overlay').classList.add('modal-hidden');
  pendingDisableAccountId = null;
}

function renderAccountCreateRoles() {
  const select = el('account-create-role');
  if (!select) return;
  clearChildren(select);
  accountRoleOptionsForActor().forEach(([value, label]) => {
    const option = createEl('option', '');
    option.value = value;
    option.textContent = label;
    select.appendChild(option);
  });
  if (!select.value) select.value = 'viewer';
}

function setAccountAdminMessage(message, type) {
  const box = el('account-admin-message');
  if (!box) return;
  box.textContent = message || '';
  box.classList.toggle('is-error', type === 'error');
  box.classList.toggle('is-success', type === 'success');
}

async function refreshAccountAdmin() {
  if (!canOpenAccountAdmin()) return;
  const body = el('account-list-body');
  clearChildren(body);
  const row = createEl('tr');
  const cell = createEl('td', '', { text: 'Loading accounts...' });
  cell.colSpan = 4;
  row.appendChild(cell);
  body.appendChild(row);

  try {
    const accounts = await apiListAccounts();
    renderAccountRows(accounts || []);
    setAccountAdminMessage(accounts && accounts.length ? '' : 'No accounts found.');
  } catch (err) {
    clearChildren(body);
    setAccountAdminMessage('Could not load accounts.', 'error');
    console.error('Account list failed:', err);
  }
}

function renderAccountRows(accounts) {
  const body = el('account-list-body');
  clearChildren(body);
  accounts.forEach(account => {
    const row = createEl('tr');
    row.appendChild(accountNameCell(account));
    row.appendChild(accountRoleCell(account));
    row.appendChild(accountStatusCell(account));
    row.appendChild(accountActionsCell(account));
    body.appendChild(row);
  });
}

function accountNameCell(account) {
  const cell = createEl('td');
  const wrap = createEl('div', 'account-name-cell');
  wrap.appendChild(createEl('span', 'account-display-name', { text: account.display_name || account.email }));
  wrap.appendChild(createEl('span', 'account-email', { text: account.email }));
  cell.appendChild(wrap);
  return cell;
}

function accountRoleCell(account) {
  const cell = createEl('td');
  const select = createEl('select', 'form-select');
  accountRoleOptionsForActor().forEach(([value, label]) => {
    const option = createEl('option', '');
    option.value = value;
    option.textContent = label;
    option.selected = account.role === value;
    select.appendChild(option);
  });
  if (!Array.from(select.options).some(option => option.value === account.role)) {
    const option = createEl('option', '');
    option.value = account.role;
    option.textContent = account.role;
    option.selected = true;
    select.appendChild(option);
  }
  select.disabled = account.disabled || !canActorManageAccount(account);
  select.addEventListener('change', async () => {
    try {
      const updated = await apiUpdateAccountRole(account.id, select.value);
      setAccountAdminMessage(`Updated ${updated.email}.`, 'success');
      await refreshAccountAdmin();
    } catch (err) {
      setAccountAdminMessage('Could not update role.', 'error');
      console.error('Account role update failed:', err);
      await refreshAccountAdmin();
    }
  });
  cell.appendChild(select);
  return cell;
}

function accountStatusCell(account) {
  const cell = createEl('td');
  const status = createEl('span', 'account-status', { text: account.disabled ? 'Disabled' : 'Active' });
  status.classList.toggle('disabled', Boolean(account.disabled));
  cell.appendChild(status);
  return cell;
}

function accountActionsCell(account) {
  const cell = createEl('td');
  const wrap = createEl('div', 'account-row-actions');
  const disable = createEl('button', 'btn btn-danger', { type: 'button' });
  disable.textContent = pendingDisableAccountId === account.id ? 'Confirm' : 'Disable';
  disable.disabled = account.disabled || account.id === currentAccount.id || !canActorManageAccount(account);
  disable.addEventListener('click', async () => {
    if (pendingDisableAccountId !== account.id) {
      pendingDisableAccountId = account.id;
      setAccountAdminMessage(`Click Confirm to disable ${account.email}.`, 'error');
      await refreshAccountAdmin();
      return;
    }
    try {
      const updated = await apiDisableAccount(account.id);
      pendingDisableAccountId = null;
      setAccountAdminMessage(`Disabled ${updated.email}.`, 'success');
      await refreshAccountAdmin();
    } catch (err) {
      pendingDisableAccountId = null;
      setAccountAdminMessage('Could not disable account.', 'error');
      console.error('Account disable failed:', err);
      await refreshAccountAdmin();
    }
  });
  wrap.appendChild(disable);
  cell.appendChild(wrap);
  return cell;
}

async function handleAccountCreate() {
  const email = el('account-create-email').value.trim();
  const displayName = el('account-create-display').value.trim();
  const password = el('account-create-password').value;
  const role = el('account-create-role').value;
  if (!email || !displayName || !password) {
    setAccountAdminMessage('Account name, display name, and password are required.', 'error');
    return;
  }
  try {
    const created = await apiCreateAccount({
      email,
      display_name: displayName,
      password,
      role,
    });
    el('account-create-email').value = '';
    el('account-create-display').value = '';
    el('account-create-password').value = '';
    setAccountAdminMessage(`Created ${created.email}.`, 'success');
    await refreshAccountAdmin();
  } catch (err) {
    setAccountAdminMessage('Could not create account.', 'error');
    console.error('Account create failed:', err);
  }
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
  list.appendChild(buildAttributeState('loading', 'Loading attributes...'));
  sec.appendChild(title);
  sec.appendChild(list);
  body.appendChild(sec);

  try {
    const attributes = await apiListAttributes(ownerType, ownerId);
    clearChildren(list);
    if (!attributes.length) {
      list.appendChild(buildAttributeState('empty', 'No attributes yet.'));
      return;
    }
    attributes.forEach(attribute => {
      list.appendChild(buildAttributeItem(attribute));
    });
  } catch (err) {
    clearChildren(list);
    list.appendChild(buildAttributeState('error', 'Could not load attributes.'));
    console.error('Attribute load failed:', err);
  }
}

function buildAttributeState(type, message) {
  const state = createEl('div', `inspector-attribute-state is-${type}`);
  const dot = createEl('span', 'inspector-attribute-state-dot');
  const text = createEl('span', 'inspector-attribute-state-text');
  text.textContent = message;
  state.appendChild(dot);
  state.appendChild(text);
  return state;
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
  setDeleteButton(el('modal-delete'), destructiveActionLabel(node.entityType));
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
            el('modal-overlay').classList.add('modal-hidden');
            const duplicateResult = await maybeUseExistingPerson(data);
            if (duplicateResult && duplicateResult.cancelled) {
              el('modal-overlay').classList.remove('modal-hidden');
              return;
            }
            if (duplicateResult && duplicateResult.person) {
              data.id = duplicateResult.person.id;
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
    await openMessageDialog('Save Failed', 'Could not save. Please try again.');
  } finally {
    el('modal-save').disabled = false;
    el('modal-save').textContent = 'Save';
  }
}

async function maybeUseExistingPerson(data) {
  if (!data.name && !data.nickname) return null;
  const matches = await apiPersonMatches({ name: data.name || '', nickname: data.nickname || '' });
  if (!matches.length) return null;
  return openDuplicateDialog(matches);
}

async function handleModalDelete() {
  if (!modalEntityId) return;
  await deleteNodeWithDialog(modalEntityId, { closeNodeModal: true });
}

async function deleteNodeWithDialog(nodeId, options) {
  const node = graph.getNode(nodeId);
  if (!node) return;
  const action = destructiveActionLabel(node.entityType);
  const ok = await openConfirmDialog({
    title: `${action} Node`,
    message: `${action} "${node.label || '?'}"?`,
    confirmLabel: action,
  });
  if (!ok) return;
  try {
    if (!isTemporaryGraphMode()) {
      try {
        await apiDelete(node.entityType, nodeId);
      } catch (err) {
        if (!isTemporaryApiError(err)) throw err;
        localOnlyWarning(`Node ${action.toLowerCase()}`, err);
      }
    }
    removeNodeEl(nodeId);
    graph.removeNode(nodeId);
    if (options && options.closeNodeModal) closeModal();
  } catch (err) {
    console.error(`${action} failed:`, err);
    await openMessageDialog(`${action} Failed`, `Could not ${action.toLowerCase()}. Please try again.`);
  }
}

function closeModal() {
  el('modal-overlay').classList.add('modal-hidden');
  modalEntityType = null;
  modalEntityId = null;
}

function validDateInputValue(value) {
  if (!value) return true;
  if (!/^\d{4}(-\d{2})?(-\d{2})?$/.test(value)) return false;
  const parts = value.split('-').map(part => Number(part));
  if (parts.length >= 2 && (parts[1] < 1 || parts[1] > 12)) return false;
  if (parts.length === 3 && (parts[2] < 1 || parts[2] > 31)) return false;
  return true;
}

async function validateDateInputs(inputs) {
  for (const input of inputs) {
    const value = input.value.trim();
    if (validDateInputValue(value)) continue;
    input.focus();
    input.style.borderColor = 'var(--error)';
    setTimeout(() => { input.style.borderColor = ''; }, 1200);
    await openMessageDialog('Invalid Date', 'Use YYYY, YYYY-MM, or YYYY-MM-DD.');
    return false;
  }
  return true;
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

  if (!await validateDateInputs([el('attribute-start'), el('attribute-end')])) return;

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
    await openMessageDialog('Attribute Save Failed', 'Could not save attribute.');
  } finally {
    el('attribute-modal-save').disabled = false;
    el('attribute-modal-save').textContent = 'Save';
  }
}

// ---- Link modal ----

let linkSourceNodeId = null;
let linkTargetNodeId = null;
let linkEditingId = null;

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
  const custom = selected === '__custom__' || (linkEditingId && selected.startsWith('custom_'));
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
  el('link-modal-delete').addEventListener('click', handleLinkDelete);
}

function openLinkModal(sourceId, targetId) {
  linkEditingId = null;
  linkSourceNodeId = sourceId;
  linkTargetNodeId = targetId;

  const source = graph.getNode(sourceId);
  const target = graph.getNode(targetId);
  if (!source || !target) return;

  renderLinkPreview(source, target);
  el('link-modal-title').textContent = 'Create Relationship';
  el('link-modal-save').textContent = 'Create';
  el('link-modal-delete').style.display = 'none';
  el('link-type-select').disabled = false;
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

function openRelationshipEditModal(linkId) {
  const link = graph.getLink(linkId);
  if (!link) return;
  const source = graph.getNode(link.sourceId);
  const target = graph.getNode(link.targetId);
  if (!source || !target) return;

  linkEditingId = linkId;
  linkSourceNodeId = link.sourceId;
  linkTargetNodeId = link.targetId;
  renderLinkPreview(source, target);
  el('link-modal-title').textContent = 'Edit Relationship';
  el('link-modal-save').textContent = 'Save';
  el('link-modal-delete').style.display = 'inline-flex';
  configureLinkTypeOptions(source, target);
  ensureLinkTypeOption(link.type, relationshipLabel(link));
  el('link-type-select').value = link.type;
  el('link-type-select').disabled = false;
  el('link-custom-label').value = link.customLabel || customRelationshipLabel(link.type) || '';
  el('link-role').value = link.role || '';
  el('link-start').value = link.startDate || '';
  el('link-end').value = link.endDate || '';
  el('link-current').checked = Boolean(link.current);
  el('link-notes').value = link.notes || '';
  updateLinkContextFields();
  el('link-modal-overlay').classList.remove('modal-hidden');
}

function renderLinkPreview(source, target) {
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
}

function ensureLinkTypeOption(value, label) {
  const select = el('link-type-select');
  if (Array.from(select.options).some(option => option.value === value)) return;
  const option = createEl('option', '');
  option.value = value;
  option.textContent = label || value;
  select.appendChild(option);
}

function relationshipFromAPI(rel) {
  return {
    id: rel.id,
    sourceId: rel.source_id || rel.sourceId,
    targetId: rel.target_id || rel.targetId,
    sourceType: rel.source_type || rel.sourceType,
    targetType: rel.target_type || rel.targetType,
    type: rel.type,
    customLabel: rel.custom_label || rel.customLabel,
    role: rel.role || '',
    startDate: rel.start_date || rel.startDate || '',
    endDate: rel.end_date || rel.endDate || '',
    current: rel.current,
    notes: rel.notes || '',
  };
}

async function handleLinkSave() {
  let type = el('link-type-select').value;
  const customLabel = el('link-custom-label').value.trim();
  let resolvedCustomLabel = customRelationshipLabel(type);
  if (linkEditingId && type.startsWith('custom_')) {
    resolvedCustomLabel = customLabel;
  }
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
  if (!await validateDateInputs([el('link-start'), el('link-end')])) return;
  const data = {
    type,
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
    if (!isTemporaryGraphMode() && type.startsWith('custom_') && resolvedCustomLabel && window.currentNetworkId) {
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
    if (linkEditingId) {
      if (isTemporaryGraphMode()) {
        rel = Object.assign({}, graph.getLink(linkEditingId), {
          custom_label: data.custom_label,
          role: data.role,
          start_date: data.start_date,
          end_date: data.end_date,
          current: data.current,
          notes: data.notes,
          type,
        });
      } else {
        rel = await apiUpdateRelationship(linkEditingId, data);
      }
      graph.updateLink(linkEditingId, relationshipFromAPI(rel));
    } else if (isTemporaryGraphMode()) {
      rel = createTemporaryRelationship(source, target, type, data);
    } else {
      try {
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
    if (!linkEditingId) graph.addLink(relationshipFromAPI(rel));
    closeLinkModal();
  } catch (err) {
    console.error('Link save failed:', err);
    await openMessageDialog('Relationship Save Failed', 'Could not save relationship.');
  } finally {
    el('link-modal-save').disabled = false;
  }
}

async function handleLinkDelete() {
  if (!linkEditingId) return;
  await promptDeleteLink(linkEditingId);
  if (!graph.getLink(linkEditingId)) closeLinkModal();
}

function closeLinkModal() {
  el('link-modal-overlay').classList.add('modal-hidden');
  linkSourceNodeId = null;
  linkTargetNodeId = null;
  linkEditingId = null;
  el('link-type-select').disabled = false;
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
    if (!el('confirm-modal-overlay').classList.contains('modal-hidden')) {
      if (e.key === 'Escape') closeConfirmDialog(false);
      return;
    }
    if (!el('duplicate-modal-overlay').classList.contains('modal-hidden')) {
      if (e.key === 'Escape') closeDuplicateDialog({ cancelled: true });
      return;
    }
    if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') return;

    if ((e.key === 'Delete' || e.key === 'Backspace') && graph.selectedNodeId) {
      e.preventDefault();
      deleteNodeWithDialog(graph.selectedNodeId);
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
