export const platforms = {
 dos:{name:'MS-DOS PC',accept:'',help:'Choose the game folder and its EXE. Click the display to type with the PC keyboard. Mouse input is available for real-mode DOS games. Gamepads map to arrows, Enter, Space and Ctrl.',compat:'Shared x86 interpreter with x87, MMX and SSE. Real-mode MZ and DJGPP go32/COFF executable loading with DOS, BIOS and DPMI HLE; VGA 320×200. No bootable OS, paging, general DOS extenders or sound output. Display intervals are synthetic. Unknown executables can be selected; compatibility varies.',buttons:[['↑',1],['↓',2],['←',4],['→',8],['Enter',16],['Space',32],['Ctrl',64],['Esc',128]],keys:{},hz:70},
 xbox:{name:'Xbox',accept:'.iso,.xiso',help:'Arrows steer the left stick; X / Z are A / B. Enter is Start, Space is Back. Q / E operate the triggers. A standard gamepad is supported.',compat:'Pentium III-class x86 interpreter with x87, MMX and SSE, Xbox kernel HLE, USB gamepad and software NV2A. XDVDFS ISO/XISO discs. No system BIOS needed; audio DSP and full hardware compatibility remain incomplete.',buttons:[['←','left'],['→','right'],['↑','up'],['↓','down'],['A',256],['B',512],['Start',16],['LT',16384],['RT',32768]],keys:{ArrowLeft:'left',ArrowRight:'right',ArrowUp:'up',ArrowDown:'down',x:256,z:512,c:1024,v:2048,Enter:16,' ':32,q:16384,e:32768},hz:60},
  gba: {
    name:'Game Boy Advance',accept:'.gba',hz:16777216/280896,
    help:'Arrows move. X / Z are A / B. Enter is Start, Shift is Select, Q / E are L / R. Pause to inspect PPU layers, priorities and color effects.',
    compat:'ARM7TDMI, BIOS HLE, scanline PPU modes 0–5, affine backgrounds/objects, windows, blending, DMA, timers and EEPROM. No firmware needed. Mosaic, SRAM/Flash save chips and link cable are incomplete. Timing follows the existing instruction-budget model. Audio is emulated but not played. Browser states preserve the session.',
    buttons:[['↑',64],['↓',128],['←',32],['→',16],['A',1],['B',2],['Start',8],['Select',4],['L',512],['R',256]],
    keys:{ArrowUp:64,ArrowDown:128,ArrowLeft:32,ArrowRight:16,x:1,z:2,Enter:8,Shift:4,q:512,e:256}
  },
  dc: {
    name:'Dreamcast',accept:'.cue,.bin',hz:60,
    help:'Select the CUE and its combined BIN together. Arrows steer the analog stick; I J K L operate the D-pad. X / Z / C / V are A / B / X / Y. Enter is Start. Q / E are left / right triggers (brake / accelerate in Crazy Taxi).',
    compat:'SH-4, AICA ARM7 and synthesis, Maple controller, BIOS/GD-ROM HLE and software PowerVR rendering. Single combined raw Mode 1 BIN with cdrdao TOC or standard CUE. GDI, CDI, CHD and separate track files are not supported. No firmware needed. Translucent sorting and some PowerVR features remain incomplete. VMU, disc audio and audio output are not exposed. Compatibility inherits the experimental Go core.',
    buttons:[['↑','up'],['↓','down'],['←','left'],['→','right'],['A',4],['B',2],['X',1024],['Y',512],['Start',8],['L',65536],['R',131072]],
    keys:{ArrowUp:'up',ArrowDown:'down',ArrowLeft:'left',ArrowRight:'right',i:16,k:32,j:64,l:128,x:4,z:2,c:1024,v:512,Enter:8,q:65536,e:131072}
  },


  "ps2": {
    "name": "PlayStation 2",
    "accept": ".iso,.bin,.cue",
    "hz": 60,
    "help": "Arrows operate the D-pad. X / Z / C / V are Cross / Square / Circle / Triangle. Enter is Start, Shift is Select. Q / E are L1 / R1, 1 / 3 are L2 / R2. I J K L move the left analog stick.",
    "compat": "Experimental Emotion Engine, IOP, VU0/VU1 and software GS, ported from the project’s Go core. Boots the disc executable with kernel HLE. Some games need a local PS2 BIOS for IOP modules; Jak and Daxter carries its own IOP image. ISO and single data-track BIN/CUE accepted. Compatibility and rendering limits inherit the Go core. Audio output, memory-card files and right-stick input are not exposed.",
    "buttons": [
      [
        "↑",
        16
      ],
      [
        "↓",
        64
      ],
      [
        "←",
        128
      ],
      [
        "→",
        32
      ],
      [
        "Cross",
        16384
      ],
      [
        "Square",
        32768
      ],
      [
        "Circle",
        8192
      ],
      [
        "Triangle",
        4096
      ],
      [
        "Start",
        8
      ],
      [
        "Select",
        1
      ],
      [
        "L1",
        1024
      ],
      [
        "R1",
        2048
      ],
      [
        "L2",
        256
      ],
      [
        "R2",
        512
      ]
    ],
    "keys": {
      "ArrowUp": 16,
      "ArrowDown": 64,
      "ArrowLeft": 128,
      "ArrowRight": 32,
      "x": 16384,
      "z": 32768,
      "c": 8192,
      "v": 4096,
      "Enter": 8,
      "Shift": 1,
      "q": 1024,
      "e": 2048,
      "i": "up",
      "k": "down",
      "j": "left",
      "l": "right",
      "1": 256,
      "3": 512
    }
  },
  "gc": {
    "name": "Nintendo GameCube",
    "accept": ".iso,.gcm",
    "hz": 60,
    "help": "Arrows move the main analog stick. X / Z are A / B, C / V are X / Y. Enter is Start, Q / E are L / R, Space is Z. I J K L operate the D-pad. Gamepads provide the main stick and D-pad separately.",
    "compat": "Experimental Gekko, Flipper GX/TEV, DSP and devices, ported from the project’s Go core. Runs the disc’s apploader with an IPL setup substitute; no firmware required. Raw uncompressed ISO/GCM only. Compatibility, nearest-neighbor texture sampling and incomplete effects inherit the Go core. Audio output, memory-card files and C-stick input are not exposed.",
    "buttons": [
      [
        "↑",
        "up"
      ],
      [
        "↓",
        "down"
      ],
      [
        "←",
        "left"
      ],
      [
        "→",
        "right"
      ],
      [
        "A",
        256
      ],
      [
        "B",
        512
      ],
      [
        "X",
        1024
      ],
      [
        "Y",
        2048
      ],
      [
        "Start",
        4096
      ],
      [
        "L",
        64
      ],
      [
        "R",
        32
      ],
      [
        "Z",
        16
      ]
    ],
    "keys": {
      "ArrowUp": "up",
      "ArrowDown": "down",
      "ArrowLeft": "left",
      "ArrowRight": "right",
      "x": 256,
      "z": 512,
      "c": 1024,
      "v": 2048,
      "Enter": 4096,
      "q": 64,
      "e": 32,
      "i": 8,
      "k": 4,
      "j": 1,
      "l": 2,
      " ": 16
    }
  }
,
  amiga: {
    name:'Amiga 500',accept:'.adf',hz:7093790/(312*454),
    help:'Arrows and Space control joystick port 2. Move the mouse over the display; left/right click operate the Amiga mouse. The Mouse speed control helps with Workbench, which moves more slowly than most games. Click the display to use the keyboard. Marble Madness: double-click the disk, then its game icon, then select GO. Turrican: left-click to leave the intro; Fire starts the game.',
    compat:'Experimental PAL 68000 / OCS machine with 512 KiB chip RAM and 512 KiB slow RAM. Kickstart 1.2 loads automatically. One standard 880 KiB ADF in DF0. Marble Madness and Turrican reach gameplay using their original loaders. Copper, blitter, bitplanes, sprites, CIA timers and floppy DMA are modeled. Chipset arbitration, scanline effects and blitter timing are approximate. Disk writing, extended ADF/IPF, AGA, multi-disk swapping and audio output are not implemented.',
    buttons:[['↑',1],['↓',2],['←',4],['→',8],['Fire',16],['Mouse left',32],['Mouse right',64]],
    keys:{ArrowUp:1,ArrowDown:2,ArrowLeft:4,ArrowRight:8,' ':16}
  },
  gb: {
    name:'Game Boy',accept:'.gb',hz:4194304/70224,
    help:'Arrows move. X / Z are A / B. Enter is Start and Shift is Select. Pause to inspect LCD tile, window and object rendering.',
    compat:'Original monochrome DMG. ROM-only and MBC1 cartridges, 32 KiB–2 MiB. Starts after the boot ROM; no firmware required. Scanline rendering includes the window, sprites and priorities. Pixel timing, DMA timing and timer edge cases are approximate. Game Boy Color-only and other cartridge mappers are not implemented. Audio output is not yet connected.',
    buttons:[['↑',4],['↓',8],['←',2],['→',1],['A',16],['B',32],['Start',128],['Select',64]],
    keys:{ArrowUp:4,ArrowDown:8,ArrowLeft:2,ArrowRight:1,x:16,z:32,Enter:128,Shift:64}
  },
  gg: {
    name:'Game Gear',accept:'.gg',hz:3579545/(228*262),
    help:'Arrows move. Z / X are buttons 1 / 2. Enter is Start. Pause to inspect VDP tile and sprite rendering.',
    compat:'Sega mapper cartridges, 16 KiB–4 MiB, with optional 512-byte copier header. No firmware required. 192-line Mode 4 tiles, sprites, scrolling and the 160×144 LCD viewport. Z80 instruction-cycle timing, line/frame interrupts and scroll/palette latches are modeled. Rendering samples each scanline; VDP fetch timing within a line, the external H-counter latch, alternate mappers, cartridge SRAM and link cable are not implemented. Audio output is not yet connected.',
    buttons:[['↑',1],['↓',2],['←',4],['→',8],['1',16],['2',32],['Start',128]],
    keys:{ArrowUp:1,ArrowDown:2,ArrowLeft:4,ArrowRight:8,z:16,x:32,Enter:128}
  },
  psp: {
    name:'PlayStation Portable',accept:'.iso,.cso',hz:60,
    help:'Arrows navigate the D-pad. X / Z / C / V are Cross / Square / Circle / Triangle. Enter is Start, Shift is Select, Q / E are L / R. I J K L move the analog stick. Gamepads provide the analog stick and D-pad separately.',
    compat:'ISO and CSO v1 UMD images with 2048-byte sectors, up to 4 GiB. Allegrex / VFPU, kernel HLE and software GE rendering. No system firmware required. PRX encryption support covers the tags used by LocoRoco and Burnout Legends. Other games may stop at unsupported imports or graphics features. MPEG video and audio playback are not implemented. Video intervals can appear black until they finish or the game accepts a skip button. Use browser save states to preserve progress.',
    buttons:[['↑',16],['↓',64],['←',128],['→',32],['Cross',16384],['Square',32768],['Circle',8192],['Triangle',4096],['Start',8],['Select',1],['L',256],['R',512]],
    keys:{ArrowUp:16,ArrowDown:64,ArrowLeft:128,ArrowRight:32,x:16384,z:32768,c:8192,v:4096,Enter:8,Shift:1,q:256,e:512,i:'up',k:'down',j:'left',l:'right'}
  },
  "3ds": {
    name:'Nintendo 3DS',accept:'.cci,.3ds',hz:60,
    help:'Arrows move the circle pad and D-pad. X / Z are A / B, C / V are X / Y. Enter is Start, Shift is Select, Q / E are L / R. Click or drag the lower screen to use the stylus. Pause to inspect either screen.',
    compat:'Decrypted CCI / NCSD cartridge images, at most 1 GiB. ARM11 + Horizon HLE and software PICA200 rendering. No system firmware required. Encrypted images, CIA packages and unsupported GPU features are rejected. Compatibility inherits the experimental Go core. Audio output and stereoscopic display are not exposed.',
    buttons:[['↑',64],['↓',128],['←',32],['→',16],['A',1],['B',2],['X',1024],['Y',2048],['Start',8],['Select',4],['L',512],['R',256]],
    keys:{ArrowUp:64,ArrowDown:128,ArrowLeft:32,ArrowRight:16,x:1,z:2,c:1024,v:2048,Enter:8,Shift:4,q:512,e:256}
  },
  ds: {
    name:'Nintendo DS',accept:'.nds',hz:60,
    help:'Arrows move. X / Z are A / B, C / V are X / Y. Enter is Start, Shift is Select, Q / E are L / R. Click or drag the lower screen to use the stylus. Pause to inspect either screen.',
    compat:'ARM9 / ARM7, synthetic firmware and BIOS HLE, two 2D engines and software 3D. Compatibility inherits the experimental Go core; audio, display capture, fog and edge marking remain incomplete.',
    buttons:[['↑',64],['↓',128],['←',32],['→',16],['A',1],['B',2],['X',1024],['Y',2048],['Start',8],['Select',4],['L',512],['R',256]],
    keys:{ArrowUp:64,ArrowDown:128,ArrowLeft:32,ArrowRight:16,x:1,z:2,c:1024,v:2048,Enter:8,Shift:4,q:512,e:256}
  },
  c64 : {
    name : 'Commodore 64',
    accept : '.tap',
    help :
        'Click the display to type on the C64 keyboard. Arrow keys and Space control joystick port 2. Enter is Return; Backspace is Delete. F1–F8 are supported; Escape is Run/Stop and Page Up is Restore.',
    compat :
        'PAL C64, TAP v0/v1. BASIC, KERNAL and character ROMs load automatically. Disk-drive formats are not implemented.',
    buttons :
        [ [ '↑', 1 ], [ '↓', 2 ], [ '←', 4 ], [ '→', 8 ], [ 'Fire', 16 ] ],
    keys :
        {ArrowUp : 1, ArrowDown : 2, ArrowLeft : 4, ArrowRight : 8, ' ' : 16},
    hz : 50
  },
  ps1 : {
    name : 'PlayStation',
    accept : '.bin,.iso,.cue',
    help :
        'Arrows steer/navigate. X / Z / C / V are Cross / Square / Circle / Triangle. Enter is Start. Q / E are L1 / R1.',
    compat :
        'PS-X EXE boot using BIOS HLE. The current core has incomplete BIOS/device coverage; importing an unknown game is not a compatibility guarantee.',
    buttons : [
      [ '↑', 16 ], [ '↓', 64 ], [ '←', 128 ], [ '→', 32 ], [ 'Cross', 16384 ],
      [ 'Square', 32768 ], [ 'Circle', 8192 ], [ 'Triangle', 4096 ],
      [ 'Start', 8 ]
    ],
    keys : {
      ArrowUp : 16,
      ArrowDown : 64,
      ArrowLeft : 128,
      ArrowRight : 32,
      x : 16384,
      z : 32768,
      c : 8192,
      v : 4096,
      Enter : 8,
      q : 1024,
      e : 2048
    },
    hz : 60
  },
  n64 : {
    name : 'Nintendo 64',
    accept : '.z64,.v64,.n64',
    help :
        'Arrows operate the analog stick. X / Z are A / B. Enter is Start, Space is Z, Q / E are L / R, and I J K L are the C buttons.',
    compat :
        'VR4300, low-level RSP and software RDP. Boot chip support and graphics accuracy are limited by the current core.',
    buttons : [
      [ '←', 'left' ], [ '→', 'right' ], [ '↑', 'up' ], [ '↓', 'down' ],
      [ 'A', 32768 ], [ 'B', 16384 ], [ 'Z', 8192 ], [ 'Start', 4096 ]
    ],
    keys : {
      x : 32768,
      z : 16384,
      Enter : 4096,
      ' ' : 8192,
      q : 32,
      e : 16,
      i : 8,
      k : 4,
      j : 2,
      l : 1,
      ArrowLeft : 'left',
      ArrowRight : 'right',
      ArrowUp : 'up',
      ArrowDown : 'down'
    },
    hz : 60
  },
  '3do' : {
    name : '3DO',
    accept : '.bin,.iso,.cue',
    help :
        'Arrows steer/navigate. X / Z / C are A / B / C. Enter is Start, Space is X. Q / E shift down / up in Need for Speed.',
    compat :
        'ARM60, Portfolio OS HLE and software cel rendering. The known Need for Speed profile uses Cinepak movie HLE; generic images retain the native player. OS coverage and game compatibility remain experimental.',
    buttons : [
      [ '↑', 0x40000000 ], [ '↓', 0x80000000 ], [ '←', 0x10000000 ],
      [ '→', 0x20000000 ], [ 'A', 0x08000000 ], [ 'B', 0x04000000 ],
      [ 'C', 0x02000000 ], [ 'Start', 0x01000000 ], [ 'L', 0x00200000 ],
      [ 'R', 0x00400000 ]
    ],
    keys : {
      ArrowUp : 0x40000000,
      ArrowDown : 0x80000000,
      ArrowLeft : 0x10000000,
      ArrowRight : 0x20000000,
      x : 0x08000000,
      z : 0x04000000,
      c : 0x02000000,
      Enter : 0x01000000,
      ' ' : 0x00800000,
      q : 0x00200000,
      e : 0x00400000
    },
    hz : 30
  }
};
