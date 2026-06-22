// DOM helper utilities

function el(id) {
  return document.getElementById(id);
}

function qs(selector, parent) {
  return (parent || document).querySelector(selector);
}

function qsa(selector, parent) {
  return Array.from((parent || document).querySelectorAll(selector));
}

function show(element) {
  element.classList.remove('hidden', 'modal-hidden', 'inspector-hidden');
}

function hide(element) {
  element.classList.add('hidden');
}

function clearChildren(element) {
  element.innerHTML = '';
}

function createEl(tag, classes, attrs) {
  const elem = document.createElement(tag);
  if (classes) {
    classes.split(' ').filter(Boolean).forEach(c => elem.classList.add(c));
  }
  if (attrs) {
    Object.entries(attrs).forEach(([k, v]) => {
      if (k === 'text') {
        elem.textContent = v;
      } else {
        elem.setAttribute(k, v);
      }
    });
  }
  return elem;
}

function svgEl(tag, attrs) {
  const elem = document.createElementNS('http://www.w3.org/2000/svg', tag);
  if (attrs) {
    Object.entries(attrs).forEach(([k, v]) => elem.setAttribute(k, v));
  }
  return elem;
}
