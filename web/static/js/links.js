// SVG link rendering

const SVG_OFFSET = 4000;

function getLinkPath(source, target, offset = { x: 0, y: 0 }) {
  const sx = source.x + SVG_OFFSET;
  const sy = source.y + SVG_OFFSET;
  const tx = target.x + SVG_OFFSET;
  const ty = target.y + SVG_OFFSET;
  const dx = tx - sx;
  const dy = ty - sy;
  const cx = dx / 2;
  const cy = dy / 2;
  return `M ${sx} ${sy} C ${sx + cx + offset.x} ${sy + cy + offset.y}, ${tx - cx + offset.x} ${ty - cy + offset.y}, ${tx} ${ty}`;
}

function getLinkMidpoint(source, target, offset = { x: 0, y: 0 }) {
  return {
    x: (source.x + target.x) / 2 + SVG_OFFSET + offset.x,
    y: (source.y + target.y) / 2 + SVG_OFFSET + offset.y,
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
  next: 'next',
  yes: 'yes',
  no: 'no',
  loop: 'loop',
  error: 'error',
};

function relationshipLabel(link) {
  if (link && link.customLabel) return link.customLabel;
  return REL_LABELS[link.type] || link.type;
}

function relationshipStyleClass(link) {
  const type = link && link.type ? String(link.type) : '';
  if (['next', 'yes', 'no', 'loop', 'error'].includes(type)) return `link-type-${type}`;
  return '';
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

function linkOffsetVector(link, source, target) {
  const offset = linkOffset(link);
  if (!offset) return { x: 0, y: 0 };

  const ids = [link.sourceId, link.targetId].sort();
  const canonicalSource = graph.getNode(ids[0]);
  const canonicalTarget = graph.getNode(ids[1]);
  const from = canonicalSource || source;
  const to = canonicalTarget || target;
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const length = Math.sqrt(dx * dx + dy * dy) || 1;
  return {
    x: (-dy / length) * offset,
    y: (dx / length) * offset,
  };
}

function renderLink(link) {
  const source = graph.getNode(link.sourceId);
  const target = graph.getNode(link.targetId);
  if (!source || !target) return;
  if (graph.isNodeHidden(source.id) || graph.isNodeHidden(target.id)) return;

  const layer = el('links-layer');
  const offset = linkOffsetVector(link, source, target);

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

  const styleClass = relationshipStyleClass(link);
  const g = svgEl('g', {
    'data-link-group': link.id,
    'class': styleClass,
  });
  g.appendChild(path);
  g.appendChild(text);

  g.addEventListener('click', () => {
    if (!canWriteData()) return;
    openRelationshipEditModal(link.id);
  });

  g.addEventListener('contextmenu', (e) => {
    e.preventDefault();
    e.stopPropagation();
    openRelationshipContextMenu(link.id, e.clientX, e.clientY);
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

async function promptDeleteLink(linkId) {
  if (!canWriteData()) return;
  const link = graph.getLink(linkId);
  if (!link) return;
  const source = graph.getNode(link.sourceId);
  const target = graph.getNode(link.targetId);
  const label = relationshipLabel(link);
  const ok = await openConfirmDialog({
    title: 'Delete Relationship',
    message: `Remove "${label}" between ${source ? source.label : '?'} and ${target ? target.label : '?'}?`,
    confirmLabel: 'Delete',
  });
  if (!ok) return;

  try {
    if (!isTemporaryGraphMode()) {
      try {
        await apiDeleteRelationship(linkId);
      } catch (err) {
        if (!isTemporaryApiError(err)) throw err;
        localOnlyWarning('Relationship delete', err);
      }
    }
    graph.removeLink(linkId);
  } catch (err) {
    console.error('Failed to delete relationship:', err);
    await openMessageDialog('Delete Failed', 'Could not delete relationship.');
  }
}
