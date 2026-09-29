// The Go Dreamcast core supports a single raw BIN with a CUE/TOC track table.
// Resolve companions locally and anchor each MODE1 track from its sector header.
export async function selectDreamcastMedia(files) {
 const cues=files.filter(f=>/\.cue$/i.test(f.name));
 if(cues.length!==1)throw Error('Select one Dreamcast CUE sheet and its BIN companion');
 if(cues[0].size>1024*1024)throw Error('CUE sheet exceeds 1 MiB');
 const text=await cues[0].text(),cdrdao=/^\s*TRACK\s+(MODE1_RAW|AUDIO)\s*$/mi.test(text);
 const tracks=[];let name='',track;
 for(const raw of text.split(/\r?\n/)){
  const line=raw.replace(/\s*\/\/.*$/,'').trim();let m;
  if(cdrdao){
   if(m=/^TRACK\s+(\S+)$/i.exec(line)){if(!['MODE1_RAW','AUDIO'].includes(m[1]))throw Error('Only raw Mode 1 and audio tracks are supported');track={number:tracks.length+1,data:m[1]==='MODE1_RAW',offset:0};tracks.push(track);}
   else if(m=/^(?:DATAFILE|FILE|AUDIOFILE)\s+"([^"]+)"(.*)$/i.exec(line)){if(!track)throw Error('CUE file before track');if(name&&name!==m[1])throw Error('Dreamcast currently requires a single combined BIN');name=m[1];const off=/#(\d+)/.exec(m[2]);track.offset=off?Number(off[1]):0;}
  }else{
   if(m=/^FILE\s+"([^"]+)"\s+BINARY$/i.exec(line)){if(name&&name!==m[1])throw Error('Dreamcast currently requires a single combined BIN');name=m[1];}
   else if(m=/^TRACK\s+(\d+)\s+(\S+)$/i.exec(line)){if(!['MODE1/2352','AUDIO'].includes(m[2]))throw Error('Dreamcast requires raw Mode 1 sectors');track={number:Number(m[1]),data:m[2]==='MODE1/2352',offset:null};tracks.push(track);}
   else if(m=/^INDEX\s+01\s+(\d+):(\d+):(\d+)$/i.exec(line)){if(!track||+m[2]>=60||+m[3]>=75)throw Error('Invalid CUE index');track.offset=((+m[1]*60+ +m[2])*75+ +m[3])*2352;}
   else if(/^(PREGAP|POSTGAP)\b/i.test(line))throw Error('CUE gaps require an explicitly laid out raw BIN');
  }
 }
 name=name.replaceAll('\\','/').split('/').pop();const matches=files.filter(f=>f.name.toLowerCase()===name.toLowerCase());
 if(matches.length!==1)throw Error('Missing or ambiguous CUE companion: '+name);const file=matches[0];
 if(files.length!==2||!tracks.length||tracks.length>99||!tracks.some(t=>t.data))throw Error('Select exactly the CUE and its single combined BIN');
 for(let i=0;i<tracks.length;i++){
  const t=tracks[i];t.length=(tracks[i+1]?.offset??file.size)-t.offset;t.lba=-1;
  if(!Number.isSafeInteger(t.offset)||t.offset<0||!Number.isSafeInteger(t.length)||t.length<=0||t.offset+t.length>file.size||t.length%2352||t.number<1||t.number>99)throw Error('Invalid raw track extent');
  if(t.data){const h=new Uint8Array(await file.slice(t.offset,t.offset+16).arrayBuffer());if(h.length!==16||h[0]!==0||h[11]!==0||h[15]!==1||!h.slice(1,11).every(v=>v===255))throw Error('Invalid Mode 1 sector header');const bcd=b=>{if((b>>4)>9||(b&15)>9)throw Error('Invalid BCD track address');return (b>>4)*10+(b&15);};const m=bcd(h[12]),s=bcd(h[13]),f=bcd(h[14]);if(s>=60||f>=75)throw Error('Invalid track address');t.lba=(m*60+s)*75+f-150;if(t.lba<0)throw Error('Negative data-track LBA');}
 }
 return {file,tracks};
}
