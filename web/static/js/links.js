// SVG link rendering

const SVG_OFFSET = 4000;

function getLinkPath(source, target, offset = 0) {
  const sx = source.x + SVG_OFFSET;
  const sy = source.y + SVG_OFFSET;
  const tx = target.x + SVG_OFFSET;
  const ty = target.y + SVG_OFFSET;
  const dx = tx - sx;
  const dy = ty - sy;
  const cx = dx / 2;
  const cy = dy / 2;
  const length = Math.sqrt(dx * dx + dy * dy) || 1;
  const nx = (-dy / length) * offset;
  const ny = (dx / length) * offset;
  return `M ${sx} ${sy} C ${sx + cx + nx} ${sy + ny}, ${tx - cx + nx} ${ty + ny}, ${tx} ${ty}`;
}

function getLinkMidpoint(source, target, offset = 0) {
  const dx = target.x - source.x;
  const dy = target.y - source.y;
  const length = Math.sqrt(dx * dx + dy * dy) || 1;
  return {
    x: (source.x + target.x) / 2 + SVG_OFFSET + (-dy / length) * offset,
    y: (source.y + target.y) / 2 + SVG_OFFSET + (dx / length) * offset,
  };
}

const REL_LABELS = {
  knows: 'knows',
  spouse_of: 'spouse of',
  parent_of: 'parent of',
  sibling_of: 'sibling of',
  works_at: 'works at',
  member_of: 'member of',
  studied_at: 'studied at',
  lives_in: 'lives in',
  has_tag: 'tagged',
  works_on: 'works on',
  sponsors: 'sponsors',
  partner_of: 'partner of',
  owns: 'owns',
};

function relationshipLabel(link) {
  if (link && link.customLabel) return link.customLabel;
  return REL_LABELS[link.type] || link.type;
}

function renderAllLinks() {
  const layer = el('links-layer');
  const defs = qs('defs', layer);
  clearChildren(layer);
  layer.appendChild(defs);

  graph.links.forEach(link => {
    renderLink(link);
  });
}

function linkPairKey(link) {
  return [link.sourceId, link.targetId].sort().join(':');
}

function linkOffset(link) {
  const pairLinks = graph.links
    .filter(item => linkPairKey(item) === linkPairKey(link))
    .sort((a, b) => String(a.id).localeCompare(String(b.id)));
  if (pairLinks.length <= 1) return 0;
  const index = pairLinks.findIndex(item => item.id === link.id);
  return (index - (pairLinks.length - 1) / 2) * 34;
}

function renderLink(link) {
  const source = graph.getNode(link.sourceId);
  const target = graph.getNode(link.targetId);
  if (!source || !target) return;
  if (graph.isNodeHidden(source.id) || graph.isNodeHidden(target.id)) return;

  const layer = el('links-layer');
  const offset = linkOffset(link);

  const path = svgEl('path', {
    'd': getLinkPath(source, target, offset),
    'class': 'graph-link',
    'data-link-id': link.id,
    'marker-end': 'url(#arrow)',
  });

  const mid = getLinkMidpoint(source, target, offset);
  const text = svgEl('text', {
    'x': String(mid.x),
    'y': String(mid.y - 8),
    'class': 'link-label',
  });
  text.textContent = relationshipLabel(link);

  const g = svgEl('g', { 'data-link-group': link.id });
  g.appendChild(path);
  g.appendChild(text);

  g.addEventListener('click', () => {
    promptDeleteLink(link.id);
  });

  const existing = qs(`[data-link-group="${link.id}"]`, layer);
  if (existing) {
    existing.replaceWith(g);
  } else {
    layer.appendChild(g);
  }
}

function removeLinkEl(id) {
  const layer = el('links-layer');
  const g = qs(`[data-link-group="${id}"]`, layer);
  if (g) g.remove();
}

function updateLinksForNode(nodeId) {
  graph.links
    .filter(l => l.sourceId === nodeId || l.targetId === nodeId)
    .forEach(renderLink);
}

function promptDeleteLink(linkId) {
  const link = graph.getLink(linkId);
  if (!link) return;
  const source = graph.getNode(link.sourceId);
  const target = graph.getNode(link.targetId);
  const label = relationshipLabel(link);
  const msg = `Remove relationship "${label}" between ${source ? source.label : '?'} and ${target ? target.label : '?'}?`;
  if (!confirm(msg)) return;
  const deletePromise = isTemporaryGraphMode()
    ? Promise.resolve()
    : apiDeleteRelationship(linkId).catch(err => {
      if (!isTemporaryApiError(err)) throw err;
      localOnlyWarning('Relationship delete', err);
    });

  deletePromise.then(() => {
    graph.removeLink(linkId);
  }).catch(err => {
    console.error('Failed to delete relationship:', err);
    alert('Could not delete relationship.');
  });
}
