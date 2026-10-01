import {platforms} from './platforms.js';
import {presentation} from './ui-platforms.js';

// Only trusted, checked-in presentation metadata is interpolated here.
export function shellMarkup(platform) {
  const config=platforms[platform], view=presentation[platform], dos=platform==='dos';
  if(!config||!view)throw new Error('Unknown emulator platform');
  return `<header class="emulator-topbar"><a href="../../">RetroReverse</a><label class="sr-only" for="system-select">System</label><select id="system-select">${Object.entries(platforms).map(([id,p])=>`<option value="${id}">${p.name}</option>`).join('')}</select><button id="open-media">Load game…</button><span id="game-name">No game loaded</span><div id="global-transport"></div><button id="session-settings">Session</button></header>
<main><h1 class="sr-only">${config.name} emulator</h1>
  <div class="workspace-heading"><nav id="workspace-nav"></nav><button id="restore-layout">Reset layout</button></div>
  <p id="status" class="status" role="status">Select a local game ${dos?'folder':'image'} to begin. Nothing is uploaded.</p>
  <div id="viewport-root"></div><div id="panel-warehouse" hidden>
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
        <p>Open Memory to inspect storage, or Render to capture a display.</p>
        <div class="state-controls"><button id="save" disabled>Save state</button><label>Load state <input id="statefile" type="file" accept=".rrstate"></label></div>
        <p id="capture-note" class="metrics">${dos?'Open Render to trace RAM rendering and its copies to VGA.':'Use Capture next display in Render to record a complete interval.'}</p>
        <p id="metrics" class="metrics">No machine running</p>
        ${view.firmwareHTML||''}
        <div id="tape" ${platform==='c64'?'':'hidden'}><button id="tapeplay">Play tape</button><button id="tapestop">Stop tape</button><p>Type LOAD, press Enter, then play the tape. Type RUN after loading.</p></div>
        <details><summary>Controls</summary><p id="help"></p><p id="device" class="metrics">Keyboard and on-screen controls</p></details>
        <details><summary>Performance</summary><p id="profile-note">Run the machine to measure subsystem timings.</p><table class="stats"><thead><tr><th>Subsystem</th><th>ms</th><th>%</th></tr></thead><tbody id="profile"></tbody></table></details>
        <details><summary>Compatibility</summary><p id="compat"></p>${view.compatibilityHTML}</details>
      </aside>
    </div>
  </section>
  <section id="memory-workspace" hidden></section>
  <section id="code-workspace" hidden></section>
  <section id="render-workspace" hidden>
    <div class="render-heading"><div><h2>Building the screen</h2><p id="render-position">Capture a display to inspect it.</p></div><button id="render-capture">Capture next display</button><button id="cancelcapture" hidden>Cancel capture</button><button id="render-resume">Resume game</button></div>
    <div id="render-toolbar" class="render-toolbar" hidden></div>
    <div class="render-layout">
      <div class="render-stage"><div class="render-buffers"><div id="render-auxiliary" class="render-auxiliary" hidden></div><div id="render-output-slot" class="render-output-slot"></div></div><div id="render-timeline"></div></div>
      <aside id="render-details" class="render-inspector" aria-label="Render inspector"></aside>
    </div>
  </section>
</div><dialog id="media-dialog"><div class="dialog-heading"><h2>Load a game</h2><button id="close-media">Close</button></div><div id="media-controls"></div></dialog></main>`;
}

export function mountShell(platform) {
  const view=presentation[platform];
  document.getElementById('emulator-app').innerHTML=shellMarkup(platform);
  const $=id=>document.getElementById(id);
  $('global-transport').append(document.querySelector('.transport'));
  const options=document.querySelector('.play-inspector');options.id='session-options';
  $('media-controls').append(document.querySelector('.files'));
  if(options.querySelector('details')){const firmware=options.querySelector('details');if(firmware.textContent.includes('firmware'))$('media-controls').append(firmware);}
  document.body.style.setProperty('--display-aspect',view.aspect);
  document.body.style.setProperty('--display-ratio',String(Number(view.aspect.split('/')[0])/Number(view.aspect.split('/')[1])));
  return view;
}
