// Router: view management

const router = (() => {
  let _currentView = 'graph';

  function navigate(view, params) {
    if (view === 'profile' && params && params.personId) {
      _showProfile(params.personId);
    } else {
      _showGraph();
    }
  }

  function _showProfile(personId) {
    _currentView = 'profile';
    el('main').style.display = 'none';
    el('profile-view').classList.remove('view-hidden');
    el('header').classList.add('profile-mode');
    profile.load(personId);
  }

  function _showGraph() {
    _currentView = 'graph';
    el('profile-view').classList.add('view-hidden');
    el('main').style.display = '';
    el('header').classList.remove('profile-mode');
  }

  function currentView() { return _currentView; }

  return { navigate, currentView };
})();
