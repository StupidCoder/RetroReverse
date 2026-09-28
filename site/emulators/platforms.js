export const platforms = {
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
