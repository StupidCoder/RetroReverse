// Logical sector geometry: image files do not preserve physical gaps or sector angles.
export function diskTracks(geometry){
 let sector=0;return geometry.counts.map((count,index)=>{const t={index,cylinder:Math.floor(index/geometry.sides),side:index%geometry.sides,start:sector,count};sector+=count;return t;});
}
const outer=246,inner=54,size=512;
export function diskPoint(geometry,side,x,y,tracks=diskTracks(geometry)){
 const radius=Math.hypot(x-256,y-256);if(radius<inner||radius>=outer)return null;
 const cylinder=Math.floor((outer-radius)/(outer-inner)*geometry.cylinders),track=tracks[cylinder*geometry.sides+side];
 if(!track)return null;
 const angle=(Math.atan2(y-256,x-256)+Math.PI/2+Math.PI*2)%(Math.PI*2);
 return {track,sector:track.start+Math.floor(angle/(Math.PI*2)*track.count),fraction:angle/(Math.PI*2)};
}
export function createDiskAtlas({root,onSector}){
 root.innerHTML='<div class="disk-surfaces"></div><p class="disk-legend"></p>';
 const surfaces=root.querySelector('div'),legend=root.querySelector('p');let geometry=null,tracks=[],canvases=[],backgrounds=[],selection=0,head=null;
 function draw(){if(!geometry)return;const pitch=(outer-inner)/geometry.cylinders;
  canvases.forEach((canvas,side)=>{const ctx=canvas.getContext('2d');ctx.putImageData(backgrounds[side],0,0);
   const selected=tracks.find(t=>selection>=t.start&&selection<t.start+t.count&&t.side===side);
   if(selected){ctx.strokeStyle='#1678c8';ctx.lineWidth=1;ctx.setLineDash([3,3]);ctx.beginPath();ctx.arc(256,256,outer-(selected.cylinder+.5)*pitch,0,Math.PI*2);ctx.stroke();ctx.setLineDash([]);const angle=(selection-selected.start)/selected.count*Math.PI*2-Math.PI/2;ctx.strokeStyle='#1678c8';ctx.lineWidth=Math.max(2,pitch);ctx.beginPath();ctx.arc(256,256,outer-(selected.cylinder+.5)*pitch,angle,angle+Math.PI*2/selected.count);ctx.stroke();}
   if(Number.isInteger(head)&&tracks[head]?.side===side){ctx.strokeStyle='#ff7b24';ctx.lineWidth=2;ctx.beginPath();ctx.arc(256,256,outer-(tracks[head].cylinder+.5)*pitch,0,Math.PI*2);ctx.stroke();}
  });
  legend.textContent='Darker bytes = 0, lighter bytes = 255. Outer to inner tracks; logical sectors clockwise from the top (angular placement is schematic). Blue = selected track and sector. '+(Number.isInteger(head)&&tracks[head]?`Orange = live head, track ${tracks[head].cylinder+geometry.firstTrack}, side ${tracks[head].side}.`:'Live head position unavailable for this image.');
 }
 return {
  load(g,bytes,sectorSize){geometry=g;tracks=diskTracks(g);head=null;selection=0;surfaces.replaceChildren();canvases=[];backgrounds=[];
   for(let side=0;side<g.sides;side++){
    const figure=document.createElement('figure'),caption=document.createElement('figcaption'),canvas=document.createElement('canvas');caption.textContent=g.sides>1?'Side '+side:'Disk';canvas.width=canvas.height=size;canvas.setAttribute('aria-label',`Disk side ${side}; grayscale bytes by track and sector; select a sector with the sector controls or click the disk`);figure.append(canvas,caption);surfaces.append(figure);canvases.push(canvas);
    const ctx=canvas.getContext('2d'),im=ctx.createImageData(size,size);
    for(let y=0;y<size;y++)for(let x=0;x<size;x++){
     const p=diskPoint(g,side,x+.5,y+.5,tracks);if(!p)continue;
     const at=p.track.start*sectorSize+Math.floor(p.fraction*p.track.count*sectorSize),v=bytes[at]??0;im.data.set([v,v,v,255],(y*size+x)*4);
    }
    backgrounds.push(im);canvas.onclick=e=>{const b=canvas.getBoundingClientRect(),p=diskPoint(g,side,(e.clientX-b.left)*size/b.width,(e.clientY-b.top)*size/b.height);if(p)onSector(p.sector);};
   }draw();
  },select(sector){selection=sector;draw();},head(track){if(head===track)return;head=track;draw();},clear(){geometry=null;surfaces.replaceChildren();legend.textContent='';}
 };
}
