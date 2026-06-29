// API layer — calls the Go backend at /api/v1
// window.MATE_CONFIG is loaded from /static/js/config.js served by the backend.

const BASE = (window.MATE_CONFIG && window.MATE_CONFIG.apiBasePath) || '/api/v1';
const API_ORGANIZATION_TYPES = [
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
];

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
  if (API_ORGANIZATION_TYPES.includes(entityType)) return 'organizations';
  if (entityType === 'project')                       return 'projects';
  if (entityType.startsWith('flow_'))                  return 'diagram-nodes';
  if (entityType === 'location')                        return 'locations';
  if (entityType === 'tag')                             return 'tags';
  return entityType + 's';
}

async function apiCreate(entityType, data) {
  const body = Object.assign({}, data);
  if (entityType.startsWith('flow_') && window.currentNetworkId) {
    body.type = entityType;
    return apiFetch('/networks/' + window.currentNetworkId + '/diagram-nodes', {
      method: 'POST',
      body: JSON.stringify(body),
    });
  }
  if (entityType === 'person' && window.currentNetworkId) {
    const result = await apiFetch('/networks/' + window.currentNetworkId + '/persons', {
      method: 'POST',
      body: JSON.stringify({ person: body, context: { notes: body.notes || '' } }),
    });
    return result.person || result;
  }
  if (API_ORGANIZATION_TYPES.includes(entityType)) {
    body.type = entityType;
  }
  return apiFetch('/' + entityPath(entityType), { method: 'POST', body: JSON.stringify(body) });
}

async function apiUpdate(entityType, id, data) {
  if (entityType.startsWith('flow_') && window.currentNetworkId) {
    const body = Object.assign({ type: entityType }, data);
    return apiFetch('/networks/' + window.currentNetworkId + '/diagram-nodes/' + id, { method: 'PUT', body: JSON.stringify(body) });
  }
  return apiFetch('/' + entityPath(entityType) + '/' + id, { method: 'PUT', body: JSON.stringify(data) });
}

async function apiDelete(entityType, id) {
  if (entityType.startsWith('flow_') && window.currentNetworkId) {
    await apiFetch('/networks/' + window.currentNetworkId + '/diagram-nodes/' + id, { method: 'DELETE' });
    return;
  }
  if (entityType === 'person' && window.currentNetworkId) {
    await apiFetch('/networks/' + window.currentNetworkId + '/persons/' + id + '/archive', { method: 'POST' });
    return;
  }
  await apiFetch('/' + entityPath(entityType) + '/' + id, { method: 'DELETE' });
}

async function apiLoadAll() {
  if (window.currentNetworkId) {
    const networkGraph = await apiFetch('/networks/' + window.currentNetworkId + '/graph');
    return {
      persons: (networkGraph.persons || []).map(item => Object.assign({}, item.person, {
        network_context: item.context,
      })),
      organizations: networkGraph.organizations || [],
      projects: networkGraph.projects || [],
      diagram_nodes: networkGraph.diagram_nodes || [],
      locations: [],
      tags: [],
      relationships: networkGraph.relationships || [],
      positions: networkGraph.positions || [],
      custom_relationship_types: networkGraph.custom_relationship_types || [],
      network: networkGraph.network,
    };
  }
  const [graph] = await Promise.all([
    apiFetch('/graph'),
  ]);
  return graph;
}

async function apiSavePosition(nodeId, nodeType, x, y) {
  const path = window.currentNetworkId ? '/networks/' + window.currentNetworkId + '/positions' : '/positions';
  apiFetch(path, {
    method: 'POST',
    body: JSON.stringify({ node_id: nodeId, node_type: nodeType, x, y }),
  }).catch(err => console.warn('Position save failed:', err.message));
}

async function apiCreateRelationship(sourceId, sourceType, targetId, targetType, type, data) {
  const relationship = Object.assign({
    network_id: window.currentNetworkId || '',
    source_id: sourceId,
    source_type: sourceType,
    target_id: targetId,
    target_type: targetType,
    type,
  }, data || {});
  return apiFetch('/relationships', {
    method: 'POST',
    body: JSON.stringify(relationship),
  });
}

async function apiDeleteRelationship(id) {
  return apiFetch('/relationships/' + id, { method: 'DELETE' });
}

async function apiUpdateRelationship(id, data) {
  return apiFetch('/relationships/' + id, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
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

async function apiCurrentAccount() {
  return apiFetch('/auth/me');
}

async function apiLogout() {
  return apiFetch('/auth/logout', { method: 'POST' });
}

async function apiListAccounts() {
  return apiFetch('/accounts');
}

async function apiCreateAccount(data) {
  return apiFetch('/accounts', { method: 'POST', body: JSON.stringify(data) });
}

async function apiUpdateAccountRole(id, role) {
  return apiFetch('/accounts/' + id, {
    method: 'PATCH',
    body: JSON.stringify({ role }),
  });
}

async function apiDisableAccount(id) {
  return apiFetch('/accounts/' + id + '/disable', { method: 'POST' });
}

async function apiListNetworks() {
  return apiFetch('/networks');
}

async function apiCreateNetwork(data) {
  return apiFetch('/networks', { method: 'POST', body: JSON.stringify(data) });
}

async function apiUpdateNetwork(id, data) {
  return apiFetch('/networks/' + id, { method: 'PUT', body: JSON.stringify(data) });
}

async function apiArchiveNetwork(id) {
  return apiFetch('/networks/' + id + '/archive', { method: 'POST' });
}

async function apiSearchNetworks(query) {
  return apiFetch('/networks/search?q=' + encodeURIComponent(query));
}

async function apiPersonMatches(data) {
  return apiFetch('/person-matches', { method: 'POST', body: JSON.stringify(data) });
}

async function apiCreateCustomRelationshipType(networkId, data) {
  return apiFetch('/networks/' + networkId + '/relationship-types', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

async function apiListCustomRelationshipTypes(networkId) {
  return apiFetch('/networks/' + networkId + '/relationship-types');
}
