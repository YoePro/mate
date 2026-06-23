// auth.js — session-based authentication for the Go backend
// Bolt/Supabase version uses Supabase auth; this version uses the Go /api/v1/auth endpoints.

const auth = (() => {
  let _mode = 'signin';

  function init() {
    _initTabs();
    _initForm();
    _initSignOut();
    _checkSession();
  }

  async function _checkSession() {
    try {
      const user = await apiFetch('/auth/me');
      if (user) {
        _onSignedIn(user);
      } else {
        _onSignedOut();
      }
    } catch {
      _onSignedOut();
    }
  }

  function _onSignedIn(user) {
    el('login-screen').style.display = 'none';
    el('app').classList.remove('app-hidden');
    const emailEl = el('user-email-display');
    if (emailEl) emailEl.textContent = user.email || user.username || '';
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
        const btn = el('auth-submit');
        btn.textContent = _mode === 'signin' ? 'Sign in' : 'Create account';
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
        const endpoint = _mode === 'signin' ? '/auth/signin' : '/auth/signup';
        const user = await apiFetch(endpoint, {
          method: 'POST',
          body: JSON.stringify({ email, password }),
        });
        if (_mode === 'signup') {
          msg.classList.add('success');
          msg.textContent = 'Account created — signing you in...';
        }
        _onSignedIn(user);
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
    el('btn-signout').addEventListener('click', async () => {
      await apiFetch('/auth/signout', { method: 'POST' }).catch(() => {});
      _onSignedOut();
    });
  }

  return { init };
})();
