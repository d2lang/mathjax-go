// SPDX-License-Identifier: Apache-2.0
// Save an unmodified original SVG, then observe controlled bbox lifecycle calls.
import {readFileSync} from 'node:fs';
import {createContext, Script} from 'node:vm';
import {createInterface} from 'node:readline';
import {createHash} from 'node:crypto';
import path from 'node:path';

const assets = process.argv[2];
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const scripts = Object.entries(hashes).map(([name, hash]) => {
  const bytes = readFileSync(path.join(assets, name));
  if (createHash('sha256').update(bytes).digest('hex') !== hash) throw Error(`unverified ${name}`);
  return new Script(bytes.toString(), {filename: name});
});
for await (const line of createInterface({input: process.stdin, crlfDelay: Infinity})) {
  if (!line.trim()) continue;
  const request = JSON.parse(line);
  const context = createContext({console, request});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  const result = new Script(`(() => {
    const proto = MathJax._.output.svg.Wrappers.menclose.SVGmenclose.prototype;
    const ids=new WeakMap(),targets=[];let next=0;
    const id=w=>{if(!ids.has(w)){ids.set(w,++next);targets.push(w);}return ids.get(w)};
    const box=b=>({w:b.w,h:b.h,d:b.d,scale:b.scale,rscale:b.rscale,l:b.L,r:b.R,pwidth:b.pwidth});
    // Collect existing wrapper objects without requesting extra measurements.
    const get=proto.getBBox, compute=proto.computeBBox;
    proto.getBBox=function(...args) {id(this);return get.apply(this,args)};
    proto.computeBBox=function(...args) {id(this);return compute.apply(this,args)};
    const svg=adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}));
    const lifecycle = targets.map(w => {
      const states=[];
      const capture=(event,result=null)=>states.push({event,cached:w.bboxComputed,owned:box(w.bbox),result:result?box(result):null,resultIsOwned:result?result===w.bbox:null});
      capture('initial');
      capture('cached-temporary',w.getBBox(false));
      w.invalidateBBox();capture('invalidate');
      capture('temporary',w.getBBox(false));
      capture('saved',w.getBBox(true));
      w.invalidateBBox();capture('second-invalidate');
      capture('second-saved',w.getBBox(true));
      return {id:id(w),states};
    });
    return {svg,lifecycle};
  })()`).runInContext(context);
  process.stdout.write(JSON.stringify(result) + '\n');
}
