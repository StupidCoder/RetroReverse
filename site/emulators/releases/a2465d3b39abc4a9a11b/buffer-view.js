import {pixelCoordinates} from './inspector.js';

// One presentation and coordinate mapping for output and auxiliary buffers.
export function createBufferView(parent,{id,title,description='',width=320,height=240,aspect='4/3',fit='stretch',output=false}={}) {
  const figure=document.createElement('figure');figure.className='buffer-view'+(output?' output-buffer':'');
  const caption=document.createElement('figcaption'),heading=document.createElement('strong'),subheading=document.createElement('span');
  const picture=document.createElement('div');picture.className='buffer-picture';picture.style.aspectRatio=aspect;
  const canvas=document.createElement('canvas');canvas.id=id;canvas.width=width;canvas.height=height;canvas.style.objectFit=fit==='contain'?'contain':'fill';
  caption.append(heading,subheading);picture.append(canvas);figure.append(caption,picture);parent.append(figure);
  let onInspect=null;
  function labels(title,description=''){heading.textContent=title;subheading.textContent=description;subheading.hidden=!description;canvas.setAttribute('aria-label',title);}
  labels(title,description);
  function resize(w,h){if(canvas.width!==w)canvas.width=w;if(canvas.height!==h)canvas.height=h;}
  function draw(pixels,w=canvas.width,h=canvas.height){resize(w,h);canvas.getContext('2d').putImageData(new ImageData(new Uint8ClampedArray(pixels),w,h),0,0);}
  function copy(source){resize(source.width,source.height);canvas.getContext('2d').drawImage(source,0,0);}
  function clear(){canvas.getContext('2d').clearRect(0,0,canvas.width,canvas.height);}
  canvas.addEventListener('click',e=>{const p=pixelCoordinates(canvas.getBoundingClientRect(),canvas.width,canvas.height,e.clientX,e.clientY,fit!=='contain');if(p)onInspect?.(p.x,p.y);});
  return {figure,canvas,picture,labels,resize,draw,copy,clear,setInspect(fn){onInspect=fn;canvas.classList.toggle('inspectable',!!fn);}};
}
