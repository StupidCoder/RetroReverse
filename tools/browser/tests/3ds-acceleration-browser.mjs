// PLAYWRIGHT_MODULE may point to an installed Playwright package entry point.
import assert from 'node:assert/strict';
import fs from 'node:fs';
const {chromium}=await import(process.env.PLAYWRIGHT_MODULE||'playwright');
const [url='http://127.0.0.1:8790/tools/browser/tests/3ds-acceleration.html',out,media]=process.argv.slice(2);
const browser=await chromium.launch({channel:process.env.CHROME_CHANNEL||'chrome',headless:true});
try{
 const page=await browser.newPage({viewport:{width:1280,height:900}}),errors=[];
 page.on('pageerror',e=>errors.push(e.message));
 await page.goto(url);await page.waitForSelector('#execution-mode',{state:'attached'});
 assert.equal(await page.locator('#execution-mode').inputValue(),'reference');
 assert.equal(await page.locator('#execution-mode option[value=experimental]').isDisabled(),true);
 assert.match(await page.locator('#execution-note').textContent(),/Software PICA/);
 if(media){
  await page.locator('#open-media').click();await page.locator('#files').setInputFiles(media);await page.locator('#load').click();
  await page.waitForFunction(()=>!document.getElementById('run').disabled,{timeout:120000});
  assert.equal(await page.locator('#execution-mode').isDisabled(),false);
  await page.locator('#execution-mode').selectOption('reference');
  await page.waitForFunction(()=>!document.getElementById('execution-mode').disabled);
 }
 const gpu=await page.evaluate(async()=>{if(!navigator.gpu)return {available:false};const a=await navigator.gpu.requestAdapter();return {available:!!a,info:a?{vendor:a.info?.vendor,architecture:a.info?.architecture,device:a.info?.device,description:a.info?.description}:null};});
 assert.deepEqual(errors,[]);
 const report={result:'PASS',browser:await browser.version(),checks:['3DS selector mounts','Reference is default','Unavailable Experimental is disabled','Effective renderer is displayed',...(media?['Cartridge loads','Mode command acknowledged']:[])],gpu};
 if(out)fs.writeFileSync(out,JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report));
}finally{await browser.close();}
