// Central graph state

const graph = (() => {
  const nodes = [];
  const links = [];
  const hiddenNodeIds = new Set();
  const lockedNodeIds = new Set();
  let selectedNodeId = null;
  const selectedNodeIds = new Set();
  let linkSourceId = null;

  const listeners = {};

  function on(event, fn) {
    if (!listeners[event]) listeners[event] = [];
    listeners[event].push(fn);
  }

  function emit(event, data) {
    (listeners[event] || []).forEach(fn => fn(data));
  }

  function addNode(node) {
    nodes.push(node);
    emit('nodes-changed', nodes);
  }

  function removeNode(id) {
    const idx = nodes.findIndex(n => n.id === id);
    if (idx === -1) return;
    nodes.splice(idx, 1);
    const toRemove = links.filter(l => l.sourceId === id || l.targetId === id);
    toRemove.forEach(l => {
      const li = links.indexOf(l);
      if (li !== -1) links.splice(li, 1);
    });
    if (selectedNodeId === id) selectedNodeId = null;
    selectedNodeIds.delete(id);
    hiddenNodeIds.delete(id);
    lockedNodeIds.delete(id);
    if (linkSourceId === id) linkSourceId = null;
    emit('nodes-changed', nodes);
    emit('links-changed', links);
  }

  function updateNodePosition(id, x, y) {
    const node = nodes.find(n => n.id === id);
    if (!node) return;
    node.x = x;
    node.y = y;
    emit('node-moved', node);
  }

  function updateNodeData(id, data) {
    const node = nodes.find(n => n.id === id);
    if (!node) return;
    Object.assign(node.data, data);
    node.label = data.name || data.nickname || node.label;
    emit('node-updated', node);
  }

  function addLink(link) {
    links.push(link);
    emit('links-changed', links);
  }

  function updateLink(id, data) {
    const link = links.find(l => l.id === id);
    if (!link) return;
    Object.assign(link, data);
    emit('links-changed', links);
  }

  function removeLink(id) {
    const idx = links.findIndex(l => l.id === id);
    if (idx !== -1) links.splice(idx, 1);
    emit('links-changed', links);
  }

  function selectNode(id) {
    selectedNodeIds.clear();
    if (id) selectedNodeIds.add(id);
    selectedNodeId = id || null;
    emit('selection-changed', id);
  }

  function toggleNodeSelection(id) {
    if (!id) return;
    if (selectedNodeIds.has(id)) {
      selectedNodeIds.delete(id);
    } else {
      selectedNodeIds.add(id);
    }
    selectedNodeId = selectedNodeIds.size === 1 ? Array.from(selectedNodeIds)[0] : null;
    emit('selection-changed', selectedNodeId);
  }

  function selectNodes(ids) {
    selectedNodeIds.clear();
    (ids || []).forEach(id => {
      if (id) selectedNodeIds.add(id);
    });
    selectedNodeId = selectedNodeIds.size === 1 ? Array.from(selectedNodeIds)[0] : null;
    emit('selection-changed', selectedNodeId);
  }

  function isNodeSelected(id) {
    return selectedNodeIds.has(id);
  }

  function setLinkSource(id) {
    linkSourceId = id;
    emit('link-source-changed', id);
  }

  function getNode(id) { return nodes.find(n => n.id === id); }
  function getLink(id) { return links.find(l => l.id === id); }
  function getNodeLinks(id) { return links.filter(l => l.sourceId === id || l.targetId === id); }

  function clear() {
    nodes.length = 0;
    links.length = 0;
    hiddenNodeIds.clear();
    lockedNodeIds.clear();
    selectedNodeId = null;
    selectedNodeIds.clear();
    linkSourceId = null;
  }

  function hideNodes(ids) {
    (ids || []).forEach(id => {
      if (id) hiddenNodeIds.add(id);
      selectedNodeIds.delete(id);
      if (selectedNodeId === id) selectedNodeId = null;
    });
    emit('selection-changed', selectedNodeId);
    emit('nodes-changed', nodes);
    emit('links-changed', links);
  }

  function showHiddenNodes() {
    hiddenNodeIds.clear();
    emit('nodes-changed', nodes);
    emit('links-changed', links);
  }

  function isNodeHidden(id) {
    return hiddenNodeIds.has(id);
  }

  function lockNodes(ids) {
    (ids || []).forEach(id => {
      if (id) lockedNodeIds.add(id);
    });
    emit('nodes-changed', nodes);
  }

  function unlockNodes(ids) {
    (ids || []).forEach(id => lockedNodeIds.delete(id));
    emit('nodes-changed', nodes);
  }

  function isNodeLocked(id) {
    return lockedNodeIds.has(id);
  }

  function load(data) {
    clear();
    const posMap = {};
    (data.positions || []).forEach(p => {
      posMap[`${p.node_id}:${p.node_type}`] = { x: p.x, y: p.y };
    });

    function nextPos(i) {
      const cols = 5;
      const spacing = 160;
      const ox = 120, oy = 120;
      return { x: ox + (i % cols) * spacing, y: oy + Math.floor(i / cols) * spacing };
    }

    let idx = 0;
    (data.persons || []).forEach(p => {
      const pos = posMap[`${p.id}:person`] || nextPos(idx++);
      nodes.push({ id: p.id, entityType: 'person', label: p.name || p.nickname, x: pos.x, y: pos.y, data: p });
    });

    (data.organizations || []).forEach(o => {
      const pos = posMap[`${o.id}:${o.type}`] || nextPos(idx++);
      nodes.push({ id: o.id, entityType: o.type, label: o.name, x: pos.x, y: pos.y, data: o });
    });

    (data.projects || []).forEach(p => {
      const pos = posMap[`${p.id}:project`] || nextPos(idx++);
      nodes.push({ id: p.id, entityType: 'project', label: p.name, x: pos.x, y: pos.y, data: p });
    });

    (data.diagram_nodes || []).forEach(d => {
      const pos = posMap[`${d.id}:${d.type}`] || nextPos(idx++);
      nodes.push({ id: d.id, entityType: d.type, label: d.name, x: pos.x, y: pos.y, data: d });
    });

    (data.locations || []).forEach(l => {
      const pos = posMap[`${l.id}:location`] || nextPos(idx++);
      nodes.push({ id: l.id, entityType: 'location', label: l.name, x: pos.x, y: pos.y, data: l });
    });

    (data.tags || []).forEach(t => {
      const pos = posMap[`${t.id}:tag`] || nextPos(idx++);
      nodes.push({ id: t.id, entityType: 'tag', label: t.name, x: pos.x, y: pos.y, data: t });
    });

    (data.relationships || []).forEach(r => {
      links.push({
        id: r.id,
        sourceId: r.source_id,
        targetId: r.target_id,
        sourceType: r.source_type,
        targetType: r.target_type,
        type: r.type,
        customLabel: r.custom_label,
        role: r.role,
        startDate: r.start_date,
        endDate: r.end_date,
        current: r.current,
        notes: r.notes,
      });
    });
  }

  return {
    get nodes() { return nodes; },
    get links() { return links; },
    get selectedNodeId() { return selectedNodeId; },
    get selectedNodeIds() { return Array.from(selectedNodeIds); },
    get hiddenNodeIds() { return Array.from(hiddenNodeIds); },
    get lockedNodeIds() { return Array.from(lockedNodeIds); },
    get linkSourceId() { return linkSourceId; },
    on, addNode, removeNode, updateNodePosition, updateNodeData,
    addLink, updateLink, removeLink, selectNode, toggleNodeSelection, selectNodes, isNodeSelected, setLinkSource,
    getNode, getLink, getNodeLinks, hideNodes, showHiddenNodes, isNodeHidden, lockNodes, unlockNodes, isNodeLocked, load, clear,
  };
})();
