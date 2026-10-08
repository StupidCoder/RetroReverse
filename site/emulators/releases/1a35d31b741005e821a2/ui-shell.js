import {platforms} from './platforms.js';
import {presentation} from './ui-platforms.js';

const icon=(path)=>`<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" focusable="false" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${path}</svg>`;
const legacyURL=new URL(import.meta.url.includes('/releases/')?'../../c64/legacy.html':'./c64/legacy.html',import.meta.url).href;
const icons={load:icon('<path d="M3 7h7l2 2h9v11H3zM3 7V4h7l2 3"/>'),run:icon('<path d="m7 4 14 8-14 8z"/>'),pause:icon('<path d="M8 4v16M16 4v16"/>'),step:icon('<path d="m4 4 12 8-12 8zM20 4v16"/>'),fullscreen:icon('<path d="M8 3H3v5M16 3h5v5M21 16v5h-5M8 21H3v-5"/>')};
const iconButton=(id,label,graphic,disabled=false)=>`<button id="${id}" class="icon-button" aria-label="${label}" title="${label}" ${disabled?'disabled':''}>${graphic}</button>`;

// Only trusted, checked-in presentation metadata is interpolated here.
export function shellMarkup(platform) {
  const config=platforms[platform], view=presentation[platform], dos=platform==='dos';
  if(!config||!view)throw new Error('Unknown emulator platform');
  return `<header class="emulator-topbar"><a href="../../">RetroReverse</a><label class="sr-only" for="system-select">System</label><select id="system-select">${Object.entries(platforms).map(([id,p])=>`<option value="${id}">${p.name}</option>`).join('')}</select>${iconButton("open-media","Load game",icons.load)}<label class="sr-only" for="workspace-nav">Layout</label><select id="workspace-nav"></select><button id="restore-layout">Reset layout</button><span id="game-name">No game loaded</span><div id="global-transport"></div></header>
<main><h1 class="sr-only">${config.name} emulator</h1>
  <p id="status" class="status" role="status">Select a local game ${dos?'folder':'image'} to begin. Nothing is uploaded.</p>
  <div id="viewport-root"></div><div id="panel-warehouse" hidden>
  <section id="play-workspace">
    <div class="files"><label for="files">Game ${dos?'folder':'image'}</label><input type="file" id="files" multiple ${dos?'webkitdirectory':''}>
      ${dos?'<label for="program">Executable</label><select id="program" aria-label="DOS executable"></select>':''}
      <button id="load" class="primary">Load ${dos?'executable':'image'}</button><button id="inspect-storage">Inspect storage</button></div>
    <div class="play-layout">
      <section class="play-stage">
        <div class="monitor"><canvas id="screen" width="${view.width}" height="${view.height}" tabindex="0" aria-label="${config.name} display and keyboard controls"></canvas></div>
        <div class="transport">${iconButton("run","Run",icons.run,true)}${iconButton("pause","Pause",icons.pause,true)}<button id="reset" disabled>Reset</button>${iconButton("step",dos?"Advance":"Next frame",icons.step,true)}<label><input id="turbo" type="checkbox"> Fast forward</label>${iconButton("fullscreen","Full screen",icons.fullscreen)}</div>
        <div id="pad" class="pad"></div>
        ${platform==='amiga'?'<label class="mouse-speed">Mouse speed <select id="mouse-speed"><option value="1">1× — games</option><option value="2">2×</option><option value="4" selected>4× — Workbench</option></select></label>':''}
      </section>
      <aside class="play-inspector">
        ${platform==='3ds'?'<label for="execution-mode">Execution mode</label><select id="execution-mode" disabled aria-describedby="execution-note"><option value="reference">Reference</option><option value="experimental" disabled>Experimental</option></select><p id="execution-note" class="metrics" role="status">ARM interpreter · Software PICA. Experimental rendering is not available in this build.</p>':''}
        <p>Open Memory to inspect storage, or Render to capture a display.</p>
        ${platform==='c64'?`<p class="hint">Checkpoints are core-specific. <a href="${legacyURL}">Open the legacy core for older checkpoints</a>.</p>`:''}
        <div class="state-controls"><button id="save" disabled>Save state</button><label>Load state <input id="statefile" type="file" accept=".rrstate"></label></div>
        <p id="capture-note" class="metrics">${dos?'Open Render to trace RAM rendering and its copies to VGA.':'Use Capture next display in Render to record a complete interval.'}</p>
        <p id="metrics" class="metrics">No machine running</p>
        ${view.firmwareHTML||''}
        <div id="tape" ${platform==='c64'?'':'hidden'}><button id="tapeplay">Play tape</button><button id="tapestop">Stop tape</button><p>Type LOAD, press Enter, then play the tape. Type RUN after loading.</p></div>
        <details><summary>Controls</summary><p id="help"></p><p id="device" class="metrics">Keyboard and on-screen controls</p></details>
        <details id="performance-panel"><summary>Performance</summary><p id="profile-note">Run the machine to measure subsystem timings.</p>${platform==='3ds'?'<p id="graphics-timing" class="metrics" hidden></p>':''}<table class="stats"><thead><tr><th>Subsystem</th><th>ms</th><th>%</th></tr></thead><tbody id="profile"></tbody></table></details>
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
