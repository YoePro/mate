// API layer — calls the Go backend at /api/v1
// window.MATE_CONFIG is loaded from /static/js/config.js served by the backend.

const BASE = (window.MATE_CONFIG && window.MATE_CONFIG.apiBasePath) || '/api/v1';

async function apiFetch(path, options) {
  const res = await fetch(BASE + path, Object.assign({
    headers: { 'Content-Type': 'application/json' },
  }, options));
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`${res.status} ${text}`);
  }
  if (res.status === 204) return null;
  return res.json();
}

// Map entity type to API resource path
function entityPath(entityType) {
  if (entityType === 'person')                          return 'persons';
  if (entityType === 'company' || entityType === 'association' || entityType === 'school') return 'organizations';
  if (entityType === 'location')                        return 'locations';
  if (entityType === 'tag')                             return 'tags';
  return entityType + 's';
}

async function apiCreate(entityType, data) {
  const body = Object.assign({}, data);
  if (entityType === 'company' || entityType === 'association' || entityType === 'school') {
    body.type = entityType;
  }
  return apiFetch('/' + entityPath(entityType), { method: 'POST', body: JSON.stringify(body) });
}

async function apiUpdate(entityType, id, data) {
  return apiFetch('/' + entityPath(entityType) + '/' + id, { method: 'PUT', body: JSON.stringify(data) });
}

async function apiDelete(entityType, id) {
  await apiFetch('/' + entityPath(entityType) + '/' + id, { method: 'DELETE' });
}

async function apiLoadAll() {
  const [graph] = await Promise.all([
    apiFetch('/graph'),
  ]);
  return graph;
}

async function apiSavePosition(nodeId, nodeType, x, y) {
  apiFetch('/positions', {
    method: 'POST',
    body: JSON.stringify({ node_id: nodeId, node_type: nodeType, x, y }),
  }).catch(err => console.warn('Position save failed:', err.message));
}

async function apiCreateRelationship(sourceId, sourceType, targetId, targetType, type, notes) {
  return apiFetch('/relationships', {
    method: 'POST',
    body: JSON.stringify({ source_id: sourceId, source_type: sourceType, target_id: targetId, target_type: targetType, type, notes: notes || null }),
  });
}

async function apiDeleteRelationship(id) {
  return apiFetch('/relationships/' + id, { method: 'DELETE' });
}

async function apiSearchAll(query) {
  return apiFetch('/search?q=' + encodeURIComponent(query));
}

async function apiListPersonAttributes(personId) {
  return apiFetch('/persons/' + personId + '/attributes');
}

async function apiCreatePersonAttribute(personId, data) {
  return apiFetch('/persons/' + personId + '/attributes', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

async function apiListAttributes(ownerType, ownerId) {
  const path = ownerType === 'person' ? 'persons' : 'organizations';
  return apiFetch('/' + path + '/' + ownerId + '/attributes');
}

async function apiCreateAttribute(ownerType, ownerId, data) {
  const path = ownerType === 'person' ? 'persons' : 'organizations';
  return apiFetch('/' + path + '/' + ownerId + '/attributes', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}
