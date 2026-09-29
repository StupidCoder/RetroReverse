import {platforms} from './platforms.js';
import {presentation} from './ui-platforms.js';

// Only trusted, checked-in presentation metadata is interpolated here.
export function shellMarkup(platform) {
  const config=platforms[platform], view=presentation[platform], dos=platform==='dos';
  if(!config||!view)throw new Error('Unknown emulator platform');
  return `<header><a href="../../">RetroReverse</a><a href="../">All emulators</a><span>${config.name}</span></header>
<main>
  <div class="workspace-heading"><h1>${config.name}</h1><nav id="workspace-nav"></nav></div>
  <p id="status" class="status" role="status">Select a local game ${dos?'folder':'image'} to begin. Nothing is uploaded.</p>
  <section id="play-workspace">
    <div class="files"><label for="files">Game ${dos?'folder':'image'}</label><input type="file" id="files" multiple ${dos?'webkitdirectory':''}>
      ${dos?'<label for="program">Executable</label><select id="program" aria-label="DOS executable"></select>':''}
      <button id="load" class="primary">Load ${dos?'executable':'image'}</button></div>
    <div class="play-layout">
      <section class="play-stage">
        <div class="monitor"><canvas id="screen" width="${view.width}" height="${view.height}" tabindex="0" aria-label="${config.name} display and keyboard controls"></canvas></div>
        <div class="transport"><button id="run" disabled>Run</button><button id="pause" disabled>Pause</button><button id="reset" disabled>Reset</button><button id="step" disabled>${dos?'Advance':'Next frame'}</button><label><input id="turbo" type="checkbox"> Fast forward</label><button id="fullscreen">Full screen</button></div>
        <div id="pad" class="pad"></div>
        ${platform==='amiga'?'<label class="mouse-speed">Mouse speed <select id="mouse-speed"><option value="1">1× — games</option><option value="2">2×</option><option value="4" selected>4× — Workbench</option></select></label>':''}
      </section>
      <aside class="play-inspector">
        <p>Pause to capture the rendering, then explore it in Render.</p>
        <div class="state-controls"><button id="save" disabled>Save state</button><label>Load state <input id="statefile" type="file" accept=".rrstate"></label></div>
        <p id="capture-note" class="metrics">${dos?'Pause to trace RAM rendering and its copies to VGA.':'Pause to record the next complete display interval.'}</p>
        <button id="cancelcapture" hidden>Cancel capture</button><p id="metrics" class="metrics">No machine running</p>
        ${view.firmwareHTML||''}
        <div id="tape" ${platform==='c64'?'':'hidden'}><button id="tapeplay">Play tape</button><button id="tapestop">Stop tape</button><p>Type LOAD, press Enter, then play the tape. Type RUN after loading.</p></div>
        <details><summary>Controls</summary><p id="help"></p><p id="device" class="metrics">Keyboard and on-screen controls</p></details>
        <details><summary>Performance</summary><p id="profile-note">Run the machine to measure subsystem timings.</p><table class="stats"><thead><tr><th>Subsystem</th><th>ms</th><th>%</th></tr></thead><tbody id="profile"></tbody></table></details>
        <details><summary>Compatibility</summary><p id="compat"></p>${view.compatibilityHTML}</details>
      </aside>
    </div>
  </section>
  <section id="render-workspace" hidden>
    <div class="render-heading"><div><h2>Building the screen</h2><p id="render-position">Capture a display to inspect it.</p></div><button id="render-resume">Resume game</button></div>
    <div id="render-toolbar" class="render-toolbar" hidden></div>
    <div class="render-layout">
      <div class="render-stage"><div class="render-buffers"><div id="render-auxiliary" class="render-auxiliary" hidden></div><div id="render-output-slot" class="render-output-slot"></div></div><div id="render-timeline"></div></div>
      <aside id="render-details" class="render-inspector" aria-label="Render inspector"></aside>
    </div>
  </section>
</main>`;
}

export function mountShell(platform) {
  const view=presentation[platform];
  document.getElementById('emulator-app').innerHTML=shellMarkup(platform);
  document.body.style.setProperty('--display-aspect',view.aspect);
  document.body.style.setProperty('--display-ratio',String(Number(view.aspect.split('/')[0])/Number(view.aspect.split('/')[1])));
  return view;
}
