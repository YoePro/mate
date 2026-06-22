// Router: view management
// Currently graph view only; extended in future versions.

const router = (() => {
  function navigate(view, params) {
    if (view === 'graph') showGraphView();
  }

  function showGraphView() {
    // graph view is the default and only view in this version
  }

  return { navigate };
})();
