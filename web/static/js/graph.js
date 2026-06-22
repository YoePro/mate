// Central graph state

const graph = (() => {
  const nodes = [];
  const links = [];
  let selectedNodeId = null;
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

  function removeLink(id) {
    const idx = links.findIndex(l => l.id === id);
    if (idx !== -1) links.splice(idx, 1);
    emit('links-changed', links);
  }

  function selectNode(id) {
    selectedNodeId = id;
    emit('selection-changed', id);
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
    selectedNodeId = null;
    linkSourceId = null;
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
        notes: r.notes,
      });
    });
  }

  return {
    get nodes() { return nodes; },
    get links() { return links; },
    get selectedNodeId() { return selectedNodeId; },
    get linkSourceId() { return linkSourceId; },
    on, addNode, removeNode, updateNodePosition, updateNodeData,
    addLink, removeLink, selectNode, setLinkSource,
    getNode, getLink, getNodeLinks, load, clear,
  };
})();
