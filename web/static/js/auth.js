// auth.js — Supabase authentication

const auth = (() => {
  let _mode = 'signin';

  function init() {
    _initTabs();
    _initForm();
    _initSignOut();

    // Auth state listener — never call Supabase methods inside this callback
    db.auth.onAuthStateChange((event, session) => {
      if (event === 'SIGNED_IN' || (event === 'INITIAL_SESSION' && session)) {
        _onSignedIn(session.user);
      } else if (event === 'SIGNED_OUT' || (event === 'INITIAL_SESSION' && !session)) {
        _onSignedOut();
      }
    });
  }

  function _onSignedIn(user) {
    el('login-screen').style.display = 'none';
    el('app').classList.remove('app-hidden');
    const emailEl = el('user-email-display');
    if (emailEl) emailEl.textContent = user.email;
    window.dispatchEvent(new CustomEvent('mate:authed', { detail: { user } }));
  }

  function _onSignedOut() {
    el('app').classList.add('app-hidden');
    el('login-screen').style.display = '';
    const emailEl = el('user-email-display');
    if (emailEl) emailEl.textContent = '';
    window.dispatchEvent(new CustomEvent('mate:signed-out'));
  }

  function _initTabs() {
    qsa('.login-tab').forEach(tab => {
      tab.addEventListener('click', () => {
        qsa('.login-tab').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        _mode = tab.dataset.tab;
        el('auth-submit').textContent = _mode === 'signin' ? 'Sign in' : 'Create account';
        const msg = el('auth-message');
        msg.textContent = '';
        msg.classList.remove('success');
      });
    });
  }

  function _initForm() {
    const submit = el('auth-submit');
    const msg = el('auth-message');

    submit.addEventListener('click', async () => {
      const email = el('auth-email').value.trim();
      const password = el('auth-password').value;
      if (!email || !password) {
        msg.textContent = 'Please enter your email and password.';
        return;
      }
      submit.disabled = true;
      submit.textContent = _mode === 'signin' ? 'Signing in...' : 'Creating account...';
      msg.textContent = '';
      msg.classList.remove('success');

      try {
        if (_mode === 'signin') {
          const { error } = await db.auth.signInWithPassword({ email, password });
          if (error) throw error;
        } else {
          const { error } = await db.auth.signUp({ email, password });
          if (error) throw error;
          msg.classList.add('success');
          msg.textContent = 'Account created — signing you in...';
        }
      } catch (err) {
        msg.textContent = err.message || 'Authentication failed.';
        msg.classList.remove('success');
      } finally {
        submit.disabled = false;
        submit.textContent = _mode === 'signin' ? 'Sign in' : 'Create account';
      }
    });

    [el('auth-email'), el('auth-password')].forEach(input => {
      input.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') submit.click();
      });
    });
  }

  function _initSignOut() {
    const btn = el('btn-signout');
    if (btn) btn.addEventListener('click', () => db.auth.signOut());
  }

  return { init };
})();
