// Real-time credit is measured in emulated time, not execution batch duration.
export class FrameClock {
 constructor(wall,emulated,turbo=false){this.reset(wall,emulated,turbo);}
 reset(wall,emulated,turbo){this.wall=wall;this.emulated=emulated;this.turbo=turbo;}
 delay(wall,emulated,turbo){
  let lead=(emulated-this.emulated)*1000-(wall-this.wall);
  if(turbo!==this.turbo||lead < -100){this.reset(wall,emulated,turbo);lead=0;}
  return turbo?0:Math.max(0,lead);
 }
}
