// API layer — calls the Go backend at /api/v1

const BASE = '/api/v1';

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

function entityPath(entityType) {
  if (entityType === 'person')                                                    return 'persons';
  if (entityType === 'company' || entityType === 'association' || entityType === 'school') return 'organizations';
  if (entityType === 'location')                                                  return 'locations';
  if (entityType === 'tag')                                                       return 'tags';
  return entityType + 's';
}

async function apiCreate(entityType, data) {
  const body = Object.assign({}, data);
  if (entityType === 'company' || entityType === 'association' || entityType === 'school') body.type = entityType;
  return apiFetch('/' + entityPath(entityType), { method: 'POST', body: JSON.stringify(body) });
}

async function apiUpdate(entityType, id, data) {
  return apiFetch('/' + entityPath(entityType) + '/' + id, { method: 'PUT', body: JSON.stringify(data) });
}

async function apiDelete(entityType, id) {
  await apiFetch('/' + entityPath(entityType) + '/' + id, { method: 'DELETE' });
}

async function apiLoadAll() {
  return apiFetch('/graph');
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

// ---- Person profile API ----

async function apiGetPerson(id) {
  return apiFetch('/persons/' + id);
}

async function apiGetPersonRelationships(personId) {
  return apiFetch('/persons/' + personId + '/relationships');
}

async function apiGetPersonProperties(personId) {
  return apiFetch('/persons/' + personId + '/properties');
}

async function apiCreatePersonProperty(personId, type, label, value, meta) {
  return apiFetch('/persons/' + personId + '/properties', {
    method: 'POST',
    body: JSON.stringify({ type, label: label || null, value: value || null, meta: meta || null }),
  });
}

async function apiUpdatePersonProperty(id, updates) {
  return apiFetch('/person-properties/' + id, {
    method: 'PUT',
    body: JSON.stringify(updates),
  });
}

async function apiDeletePersonProperty(id) {
  await apiFetch('/person-properties/' + id, { method: 'DELETE' });
}
