// Canvas: pan, zoom, and coordinate transforms for the workspace

const canvas = (() => {
  let tx = 0, ty = 0, scale = 1;
  let isPanning = false;
  let panStart = { x: 0, y: 0 };
  let spaceDown = false;

  const workspace = el('workspace');
  const inner = el('workspace-inner');

  function applyTransform() {
    inner.style.transform = `translate(${tx}px, ${ty}px) scale(${scale})`;
  }

  function screenToWorld(sx, sy) {
    const rect = workspace.getBoundingClientRect();
    return {
      x: (sx - rect.left - tx) / scale,
      y: (sy - rect.top  - ty) / scale,
    };
  }

  function worldToScreen(wx, wy) {
    const rect = workspace.getBoundingClientRect();
    return {
      x: wx * scale + tx + rect.left,
      y: wy * scale + ty + rect.top,
    };
  }

  function fitToNodes(nodes) {
    if (!nodes.length) {
      tx = 0; ty = 0; scale = 1;
      applyTransform();
      return;
    }
    const rect = workspace.getBoundingClientRect();
    const xs = nodes.map(n => n.x);
    const ys = nodes.map(n => n.y);
    const minX = Math.min(...xs) - 80;
    const maxX = Math.max(...xs) + 80;
    const minY = Math.min(...ys) - 80;
    const maxY = Math.max(...ys) + 80;
    const contentW = maxX - minX;
    const contentH = maxY - minY;
    scale = Math.min(
      rect.width  / contentW,
      rect.height / contentH,
      1.2
    ) * 0.9;
    tx = (rect.width  - contentW * scale) / 2 - minX * scale;
    ty = (rect.height - contentH * scale) / 2 - minY * scale;
    applyTransform();
  }

  function startPan(e) {
    isPanning = true;
    panStart = { x: e.clientX - tx, y: e.clientY - ty };
    workspace.classList.add('panning');
  }

  function onMouseMove(e) {
    if (!isPanning) return;
    tx = e.clientX - panStart.x;
    ty = e.clientY - panStart.y;
    applyTransform();
  }

  function stopPan() {
    if (!isPanning) return;
    isPanning = false;
    workspace.classList.remove('panning');
  }

  workspace.addEventListener('wheel', (e) => {
    e.preventDefault();
    const rect = workspace.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    const delta = e.deltaY > 0 ? 0.9 : 1.1;
    const newScale = Math.max(0.2, Math.min(3, scale * delta));
    tx = mx - (mx - tx) * (newScale / scale);
    ty = my - (my - ty) * (newScale / scale);
    scale = newScale;
    applyTransform();
  }, { passive: false });

  workspace.addEventListener('mousedown', (e) => {
    if (e.button === 1 || spaceDown) {
      e.preventDefault();
      startPan(e);
    }
  });

  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', stopPan);

  document.addEventListener('keydown', (e) => {
    if (e.code === 'Space' && e.target === document.body) {
      e.preventDefault();
      spaceDown = true;
      workspace.style.cursor = 'grab';
    }
  });

  document.addEventListener('keyup', (e) => {
    if (e.code === 'Space') {
      spaceDown = false;
      workspace.style.cursor = '';
    }
  });

  applyTransform();

  return { screenToWorld, worldToScreen, fitToNodes, applyTransform, getScale: () => scale };
})();
