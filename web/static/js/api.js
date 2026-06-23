// Supabase API layer

const { createClient } = supabase;
const db = createClient(window.MATE_CONFIG.supabaseUrl, window.MATE_CONFIG.supabaseKey);

const ENTITY_TABLE = {
  person:      'persons',
  company:     'organizations',
  association: 'organizations',
  school:      'organizations',
  location:    'locations',
  tag:         'tags',
};

async function apiCreate(entityType, data) {
  const table = ENTITY_TABLE[entityType];
  const row = Object.assign({}, data);
  if (entityType === 'company' || entityType === 'association' || entityType === 'school') {
    row.type = entityType;
  }
  const { data: result, error } = await db.from(table).insert(row).select().single();
  if (error) throw error;
  return result;
}

async function apiUpdate(entityType, id, data) {
  const table = ENTITY_TABLE[entityType];
  const { data: result, error } = await db.from(table).update(data).eq('id', id).select().single();
  if (error) throw error;
  return result;
}

async function apiDelete(entityType, id) {
  const table = ENTITY_TABLE[entityType];
  const { error } = await db.from(table).delete().eq('id', id);
  if (error) throw error;
  await db.from('node_positions').delete().eq('node_id', id).eq('node_type', entityType);
  await db.from('relationships').delete().or(`source_id.eq.${id},target_id.eq.${id}`);
}

async function apiLoadAll() {
  const [persons, orgs, locations, tags, rels, positions] = await Promise.all([
    db.from('persons').select('*'),
    db.from('organizations').select('*'),
    db.from('locations').select('*'),
    db.from('tags').select('*'),
    db.from('relationships').select('*'),
    db.from('node_positions').select('*'),
  ]);

  const errors = [persons, orgs, locations, tags, rels, positions]
    .map(r => r.error).filter(Boolean);
  if (errors.length > 0) throw errors[0];

  return {
    persons:       persons.data,
    organizations: orgs.data,
    locations:     locations.data,
    tags:          tags.data,
    relationships: rels.data,
    positions:     positions.data,
  };
}

async function apiSavePosition(nodeId, nodeType, x, y) {
  const { error } = await db.from('node_positions')
    .upsert({ node_id: nodeId, node_type: nodeType, x, y }, { onConflict: 'node_id,node_type' });
  if (error) console.warn('Position save failed:', error.message);
}

async function apiCreateRelationship(sourceId, sourceType, targetId, targetType, type, notes) {
  const { data, error } = await db.from('relationships').insert({
    source_id: sourceId, source_type: sourceType,
    target_id: targetId, target_type: targetType,
    type, notes: notes || null,
  }).select().single();
  if (error) throw error;
  return data;
}

async function apiDeleteRelationship(id) {
  const { error } = await db.from('relationships').delete().eq('id', id);
  if (error) throw error;
}

async function apiSearchAll(query) {
  const q = query.toLowerCase();
  const [persons, orgs, locations, tags] = await Promise.all([
    db.from('persons').select('id, name, nickname').ilike('name', `%${q}%`).limit(5),
    db.from('organizations').select('id, name, type').ilike('name', `%${q}%`).limit(5),
    db.from('locations').select('id, name').ilike('name', `%${q}%`).limit(3),
    db.from('tags').select('id, name').ilike('name', `%${q}%`).limit(3),
  ]);
  return { persons: persons.data, organizations: orgs.data, locations: locations.data, tags: tags.data };
}

// ---- Person profile API ----

async function apiGetPerson(id) {
  const { data, error } = await db.from('persons').select('*').eq('id', id).single();
  if (error) throw error;
  return data;
}

async function apiGetPersonRelationships(personId) {
  const { data, error } = await db.from('relationships').select('*')
    .or(`source_id.eq.${personId},target_id.eq.${personId}`);
  if (error) throw error;
  return data;
}

async function apiGetPersonProperties(personId) {
  const { data, error } = await db.from('person_properties').select('*')
    .eq('person_id', personId).order('sort_order').order('created_at');
  if (error) throw error;
  return data;
}

async function apiCreatePersonProperty(personId, type, label, value, meta) {
  const { data, error } = await db.from('person_properties').insert({
    person_id: personId, type,
    label: label || null, value: value || null, meta: meta || null,
  }).select().single();
  if (error) throw error;
  return data;
}

async function apiUpdatePersonProperty(id, updates) {
  const { data, error } = await db.from('person_properties').update(updates).eq('id', id).select().single();
  if (error) throw error;
  return data;
}

async function apiDeletePersonProperty(id) {
  const { error } = await db.from('person_properties').delete().eq('id', id);
  if (error) throw error;
}
