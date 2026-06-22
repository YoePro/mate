// SVG link rendering

const SVG_OFFSET = 4000;

function getLinkPath(source, target) {
  const sx = source.x + SVG_OFFSET;
  const sy = source.y + SVG_OFFSET;
  const tx = target.x + SVG_OFFSET;
  const ty = target.y + SVG_OFFSET;
  const dx = tx - sx;
  const dy = ty - sy;
  const cx = dx / 2;
  const cy = dy / 2;
  return `M ${sx} ${sy} C ${sx + cx} ${sy}, ${tx - cx} ${ty}, ${tx} ${ty}`;
}

function getLinkMidpoint(source, target) {
  return {
    x: (source.x + target.x) / 2 + SVG_OFFSET,
    y: (source.y + target.y) / 2 + SVG_OFFSET,
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
};

function renderAllLinks() {
  const layer = el('links-layer');
  const defs = qs('defs', layer);
  clearChildren(layer);
  layer.appendChild(defs);

  graph.links.forEach(link => {
    renderLink(link);
  });
}

function renderLink(link) {
  const source = graph.getNode(link.sourceId);
  const target = graph.getNode(link.targetId);
  if (!source || !target) return;

  const layer = el('links-layer');

  const path = svgEl('path', {
    'd': getLinkPath(source, target),
    'class': 'graph-link',
    'data-link-id': link.id,
    'marker-end': 'url(#arrow)',
  });

  const mid = getLinkMidpoint(source, target);
  const text = svgEl('text', {
    'x': String(mid.x),
    'y': String(mid.y - 8),
    'class': 'link-label',
  });
  text.textContent = REL_LABELS[link.type] || link.type;

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
  const label = REL_LABELS[link.type] || link.type;
  const msg = `Remove relationship "${label}" between ${source ? source.label : '?'} and ${target ? target.label : '?'}?`;
  if (!confirm(msg)) return;
  apiDeleteRelationship(linkId).then(() => {
    graph.removeLink(linkId);
  }).catch(err => {
    console.error('Failed to delete relationship:', err);
    alert('Could not delete relationship.');
  });
}
