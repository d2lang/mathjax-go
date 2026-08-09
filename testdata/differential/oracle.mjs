// Test-only differential oracle for D2's frozen MathJax 3.2.2 component.
// SPDX-License-Identifier: Apache-2.0

import {readFileSync} from 'node:fs';
import {join} from 'node:path';
import {createInterface} from 'node:readline';
import {createContext, Script} from 'node:vm';

function usage() {
  throw new Error('usage: node oracle.mjs --asset-dir /path/to/d2latex-assets');
}

const index = process.argv.indexOf('--asset-dir');
if (index < 0 || !process.argv[index + 1]) usage();
const assetDir = process.argv[index + 1];

const scripts = ['polyfills.js', 'mathjax.js', 'setup.js'].map((asset) =>
  new Script(readFileSync(join(assetDir, asset), 'utf8'), {filename: asset}),
);

function createD2Runtime() {
  // D2 constructs a fresh JavaScript engine, adaptor, input jax, output jax,
  // and MathDocument for every Render call. Reusing setup.js state here would
  // leak tag, macro, handler, and font state across a batch and would therefore
  // cease to be an exact oracle.
  const context = createContext({console});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  return context;
}

const input = createInterface({input: process.stdin, crlfDelay: Infinity});
for await (const line of input) {
  if (!line.trim()) continue;
  try {
    const request = JSON.parse(line);
    const options = request.options ?? {};
    if ((options.FontCache ?? options.fontCache ?? 0) !== 0) {
      throw new Error('only fontCache none is supported');
    }
    const em = options.Em ?? options.em ?? 16;
    const ex = options.Ex ?? options.ex ?? 8;
    const display = options.Display ?? options.display ?? true;
    const context = createD2Runtime();
    const node = context.html.convert(request.tex, {em, ex, display});
    const svg = context.adaptor.innerHTML(node);
    process.stdout.write(JSON.stringify({svg}) + '\n');
  } catch (error) {
    process.stdout.write(JSON.stringify({error: String(error?.stack ?? error)}) + '\n');
  }
}
