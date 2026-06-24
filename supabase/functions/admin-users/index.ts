import { createClient } from 'npm:@supabase/supabase-js@2';

const corsHeaders = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, PATCH, DELETE, OPTIONS',
  'Access-Control-Allow-Headers': 'Content-Type, Authorization, X-Client-Info, Apikey',
};

function json(data: unknown, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: { ...corsHeaders, 'Content-Type': 'application/json' },
  });
}

Deno.serve(async (req: Request) => {
  if (req.method === 'OPTIONS') {
    return new Response(null, { status: 200, headers: corsHeaders });
  }

  try {
    const authHeader = req.headers.get('Authorization') ?? '';
    const token = authHeader.replace('Bearer ', '');
    if (!token) return json({ error: 'Missing authorization' }, 401);

    const supabaseUrl = Deno.env.get('SUPABASE_URL')!;
    const serviceKey  = Deno.env.get('SUPABASE_SERVICE_ROLE_KEY')!;
    const anonKey     = Deno.env.get('SUPABASE_ANON_KEY')!;

    const admin = createClient(supabaseUrl, serviceKey, {
      auth: { autoRefreshToken: false, persistSession: false },
    });

    const userClient = createClient(supabaseUrl, anonKey);
    const { data: { user }, error: jwtErr } = await userClient.auth.getUser(token);
    if (jwtErr || !user) return json({ error: 'Invalid token' }, 401);

    const { data: callerProfile } = await admin
      .from('user_profiles')
      .select('role')
      .eq('id', user.id)
      .maybeSingle();

    if (!callerProfile || callerProfile.role !== 'admin') {
      return json({ error: 'Forbidden — admin role required' }, 403);
    }

    const url      = new URL(req.url);
    const segments = url.pathname.split('/').filter(Boolean);
    const targetId = segments[1] ?? null;

    if (req.method === 'GET') {
      const { data: { users }, error } = await admin.auth.admin.listUsers({ perPage: 1000 });
      if (error) throw error;
      const { data: profiles } = await admin.from('user_profiles').select('*');
      const byId = Object.fromEntries((profiles ?? []).map((p: Record<string, unknown>) => [p.id, p]));
      const merged = users.map((u) => ({
        id:              u.id,
        email:           u.email ?? '',
        created_at:      u.created_at,
        last_sign_in_at: u.last_sign_in_at ?? null,
        name:            byId[u.id]?.name      ?? null,
        role:            byId[u.id]?.role      ?? 'editor',
        read_only:       byId[u.id]?.read_only ?? false,
      }));
      return json(merged);
    }

    if (!targetId) return json({ error: 'Missing user id' }, 400);

    if (req.method === 'PATCH') {
      const body = await req.json() as Record<string, unknown>;
      const authUpdates: Record<string, unknown> = {};
      if (body.email)    authUpdates.email    = body.email;
      if (body.password) authUpdates.password = body.password;
      if (Object.keys(authUpdates).length) {
        const { error } = await admin.auth.admin.updateUserById(targetId, authUpdates);
        if (error) throw error;
      }
      const profilePatch: Record<string, unknown> = { id: targetId };
      if (body.name      !== undefined) profilePatch.name      = body.name;
      if (body.role      !== undefined) profilePatch.role      = body.role;
      if (body.read_only !== undefined) profilePatch.read_only = body.read_only;
      const { error: pErr } = await admin.from('user_profiles').upsert(profilePatch, { onConflict: 'id' });
      if (pErr) throw pErr;
      return json({ success: true });
    }

    if (req.method === 'DELETE') {
      if (targetId === user.id) return json({ error: 'Cannot delete your own account' }, 400);
      const { error } = await admin.auth.admin.deleteUser(targetId);
      if (error) throw error;
      return json({ success: true });
    }

    return json({ error: 'Method not allowed' }, 405);

  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err);
    return json({ error: msg }, 500);
  }
});
