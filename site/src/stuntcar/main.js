import { TrackViewer } from './viewer.js';

const viewport = document.getElementById('viewport');
const select = document.getElementById('track');
const viewer = new TrackViewer(viewport);

fetch('public/stuntcar/tracks.json')
  .then(r => r.json())
  .then(tracks => {
    select.innerHTML = '';
    tracks.forEach((t, i) => {
      const o = document.createElement('option');
      o.value = i;
      o.textContent = `${i + 1}. ${t.name} (${t.sections} sections)`;
      select.appendChild(o);
    });
    const pick = () => viewer.show(tracks[+select.value]);
    select.addEventListener('change', pick);
    pick();
  })
  .catch(err => {
    viewport.innerHTML = `<p style="padding:1em;color:#c66">Could not load tracks.json: ${err}</p>`;
  });
