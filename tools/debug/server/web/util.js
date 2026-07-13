// Small shared helpers.

export function esc(s) {
  return String(s).replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c]));
}

export function hex(n, width = 8) {
  return (n >>> 0).toString(16).padStart(width, '0');
}
