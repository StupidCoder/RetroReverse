// Presentation metadata only. Execution and input capabilities stay in platforms.js.
export const presentation = {
  "3do": {
    "width": 320,
    "height": 240,
    "aspect": "4/3",
    "render": "commands",
    "compatibilityHTML": "<label><input type=\"checkbox\" id=\"compatprofile\" checked> Apply known-image compatibility fixes on load</label><p>Cel, PIXC and framebuffer history reflects the current model. Packed-source decoding is captured as immutable source bytes; this is not general CPU dataflow tracing.</p>"
  },
  "3ds": {
    "width": 400,
    "height": 480,
    "aspect": "5/6",
    "render": "commands",
    "compatibilityHTML": "<p>Capture follows PICA register writes, raster writes and GX transfers into the displayed buffers. Captured RAM uses packed offsets; command metadata retains guest addresses. Texture-sampler ancestry is not yet recorded.</p>"
  },
  "amiga": {
    "width": 640,
    "height": 256,
    "aspect": "5/4",
    "render": "amiga",
    "compatibilityHTML": "<p>Scanout samples bitplane memory per line; palette writes retain their modeled horizontal position. Blitter DMA timing remains approximate. BOBs are drawn into bitplanes; they are separate from hardware sprites.</p><p>68000 CPU: <a href=\"../cores/amiga/LICENSE.txt\">Musashi license and attribution</a>.</p>",
    "firmwareHTML": "<details class=\"firmware\"><summary>Optional Kickstart override</summary><label>256 or 512 KiB ROM <input id=\"kickstart\" type=\"file\" accept=\".rom,.bin\"></label><p>Kickstart 1.2 is included.</p></details>"
  },
  "c64": {
    "width": 392,
    "height": 272,
    "aspect": "392/272",
    "render": "raster",
    "compatibilityHTML": "<p>Inspection uses the recorded state at each scanline, with memory and registers held fixed in the source previews. Actual output retains cycle-level VIC fetches, priority decisions and changes within a scanline. The previews do not reproduce the CPU’s later writes. Character ROM, RAM, color RAM and I/O writers remain distinct.</p>",
    "firmwareHTML": "<details><summary>Optional firmware override</summary><p>Leave empty to use the hosted ROMs.</p><label>BASIC ROM <input id=\"basic\" type=\"file\"></label><label>KERNAL ROM <input id=\"kernal\" type=\"file\"></label><label>Character ROM <input id=\"chargen\" type=\"file\"></label></details>"
  },
  "dc": {
    "width": 640,
    "height": 480,
    "aspect": "4/3",
    "render": "commands",
    "compatibilityHTML": "<p>Capture records PowerVR parameters, framebuffer clears and actual VRAM writes. Three video fields include double-buffer context. Replay uses the final scanout mapping. Rejected fragments and texture-fetch ancestry are not recorded.</p>"
  },
  "dos": {
    "width": 640,
    "height": 480,
    "aspect": "4/3",
    "render": "commands",
    "compatibilityHTML": "<p>Capture discovers RAM render buffers through literal MOVS copies to VGA and records their CPU writes. Select a pixel to see renderer PCs, follow earlier buffer copies, or inspect the final VGA transfer. Reveal buffer writes dims pixels until they are written; this projection is not historical monitor output. Register-mediated or transformed copies currently fall back to VGA history. Display intervals use an instruction budget because the original PC model has no cycle-accurate VGA scanout. The known Quake executable profile restores a DJGPP base-address global; other executables are left untouched.</p><label><input type=\"checkbox\" id=\"compatprofile\" checked> Enable known executable profiles</label>"
  },
  "ds": {
    "width": 256,
    "height": 384,
    "aspect": "2/3",
    "render": "commands",
    "compatibilityHTML": "<p>Captured surfaces track 2D composition and GX polygon contributions. Texture-fetch ancestry is not yet recorded. Addresses refer to capture-local surfaces.</p>"
  },
  "gb": {
    "width": 160,
    "height": 144,
    "aspect": "10/9",
    "render": "raster",
    "compatibilityHTML": "<p>Inspection uses the recorded state at each scanline, including the window row counter. Source panels are frozen-state previews; the output retains actual completed lines. Effects within a scanline remain limited by this core's scanline-level renderer. Memory addresses use DMG VRAM, OAM and palette-register addresses.</p>"
  },
  "gba": {
    "width": 240,
    "height": 160,
    "aspect": "3/2",
    "render": "commands",
    "compatibilityHTML": "<p>Capture records PPU background and object samples, final priority/window/blending composition, and historical VRAM/palette/OAM writes. Render replay follows the recorded scanline work.</p>"
  },
  "gc": {
    "width": 640,
    "height": 480,
    "aspect": "4/3",
    "render": "commands",
    "compatibilityHTML": "<p>Capture follows GX commands, EFB color writes, clears and pixel-engine display copies. Scrubbing shows the EFB being built, then the YUY2 display buffer. Pixel inspection explains the EFB value just before its display copy. Register snapshots are available; rejected fragments, texture-fetch and CPU memory-writer ancestry are not yet recorded.</p>"
  },
  "gg": {
    "width": 160,
    "height": 144,
    "aspect": "10/9",
    "render": "raster",
    "compatibilityHTML": "<p>Inspection uses the recorded state at each scanline, including latched scrolling and palette colors. Source panels are frozen-state previews; the output retains actual completed lines. Effects within a scanline remain limited by this core's scanline-level renderer. The LCD shows VDP columns 48–207 and lines 24–167. Source addresses distinguish VRAM, color RAM and VDP registers.</p>"
  },
  "n64": {
    "width": 320,
    "height": 240,
    "aspect": "4/3",
    "render": "commands",
    "compatibilityHTML": "<p>RDP texture, depth and blending evidence reflects the current model. Coverage and VI scanout remain simplified. A TMEM location may be the last tap of a filtered sample.</p>"
  },
  "ps1": {
    "width": 320,
    "height": 240,
    "aspect": "4/3",
    "render": "commands",
    "compatibilityHTML": "<p>Trace reflects this GPU model. Semi-transparency blending and VRAM mask-bit effects are not implemented; command submission PCs are not data-construction PCs.</p>"
  },
  "ps2": {
    "width": 640,
    "height": 448,
    "aspect": "4/3",
    "render": "commands",
    "compatibilityHTML": "<p>Capture follows GS primitives and transfers, recording actual color and depth writes in 4 MiB GS VRAM. Scrubbing reconstructs the final display buffer in command order. Command details include GS register snapshots. Rejected fragments, VU instruction history and texture-fetch ancestry are not yet recorded.</p>",
    "firmwareHTML": "<details class=\"firmware\"><summary>Optional PS2 firmware</summary><label>BIOS ROM <input id=\"bios\" type=\"file\" accept=\".bin,.rom\"></label><p>Some games need IOP modules from a 4 or 8 MiB BIOS. This file stays on your computer.</p></details>"
  },
  "psp": {
    "width": 480,
    "height": 272,
    "aspect": "30/17",
    "render": "commands",
    "compatibilityHTML": "<p>Capture records GE commands and framebuffer writes, including rejected fragments. Replay shows the captured display buffer as those writes occur. Addresses are capture-local: VRAM starts at 0, main RAM at 0x200000, scratchpad at 0x2200000. Texture-fetch ancestry is not yet recorded.</p>"
  },
  "xbox": {
    "width": 640,
    "height": 480,
    "aspect": "4/3",
    "render": "commands",
    "compatibilityHTML": "<p>Capture records actual RAM writes and NV2A draw/clear register snapshots. Replay uses the final presented surface mapping. Anti-aliased pixels expose all sample addresses; the initial history is for the first sample. Rejected fragments and texture-fetch ancestry are not yet recorded.</p>"
  }
};
