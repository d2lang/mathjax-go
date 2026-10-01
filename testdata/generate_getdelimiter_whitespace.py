#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Capture complete original GetDelimiter results; no Go output is an expectation."""
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys

here = Path(__file__).resolve().parent
assets, node = Path(sys.argv[1]), sys.argv[2]
hashes = {
    "mathjax.js": "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
    "polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
    "setup.js": "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
}
for name, expected in hashes.items():
    assert hashlib.sha256((assets / name).read_bytes()).hexdigest() == expected, name

method_code = r"""
const fs = require('node:fs'), path = require('node:path'), vm = require('node:vm');
const c = vm.createContext({console}); c.globalThis = c;
for (const name of ['polyfills.js','mathjax.js','setup.js'])
  new vm.Script(fs.readFileSync(path.join(process.argv[1], name), 'utf8'), {filename:name}).runInContext(c);
process.stdout.write(JSON.stringify(vm.runInContext(`(()=>{
  const P = MathJax._.input.tex.TexParser.default.prototype;
  return Object.fromEntries(['nextIsSpace','GetNext','GetDelimiter'].map(name=>[name,P[name].toString()]));
})()`, c)));
"""
methods = json.loads(subprocess.check_output([node, '--jitless', '-e', method_code, str(assets)], text=True))
assert 'GetNext' in methods['GetDelimiter'] and '.trim()' in methods['GetDelimiter']
assert 'nextIsSpace' in methods['GetNext'] and r'/\s/' in methods['nextIsSpace']

requests, seen = [], set()
def add(name, tex):
    for display in [False, True]:
        if (tex, display) in seen:
            continue
        seen.add((tex, display))
        requests.append({'name': name + ('-display' if display else '-inline'), 'tex': tex, 'display': display})

spaces = [('none',''), ('space',' '), ('nbsp','\u00a0'), ('bom','\ufeff'),
          ('nel-control','\u0085'), ('line-separator','\u2028'), ('paragraph-separator','\u2029')]
for ws_name, ws in spaces:
    for size in ['big','Big','bigg','Bigg']:
        for suffix in ['', 'l', 'r', 'm']:
            cmd = size + suffix
            for delimiter_name, delimiter in [('raw','('), ('command',r'\langle'), ('brace','{(}'),
                                               ('brace-trim','{' + ws + '(' + ws + '}')]:
                add(f'{cmd}-{ws_name}-{delimiter_name}', '\\' + cmd + ws + delimiter + 'x')
    for label, formula in [
        ('left-right', r'\left' + ws + r'(\frac{a}{b}\right' + ws + ')'),
        ('angle-fence', r'\left' + ws + r'\langle x\right' + ws + r'\rangle'),
        ('middle-leading', r'\left(x\middle' + ws + r'|y\right)'),
        ('middle-adjacent', r'\left(' + ws + r'x\middle' + ws + r'|y\right)'),
        ('mleft-mright', r'\mleft' + ws + r'(\frac{a}{b}\mright' + ws + ')'),
        ('dot-fence', r'\left' + ws + r'.x\right' + ws + '|'),
    ]:
        add(label + '-' + ws_name, formula)

# Full ECMAScript whitespace inventory, plus three characters outside that set.
all_spaces = [
    ('space',' '), ('tab','\t'), ('lf','\n'), ('vt','\v'), ('ff','\f'), ('cr','\r'),
    ('nbsp','\u00a0'), ('ogham','\u1680'), ('en-quad','\u2000'), ('em-quad','\u2001'),
    ('en-space','\u2002'), ('em-space','\u2003'), ('third-em','\u2004'), ('quarter-em','\u2005'),
    ('sixth-em','\u2006'), ('figure-space','\u2007'), ('punctuation-space','\u2008'),
    ('thin-space','\u2009'), ('hair-space','\u200a'), ('ls','\u2028'), ('ps','\u2029'),
    ('narrow-nbsp','\u202f'), ('medium-space','\u205f'), ('ideographic-space','\u3000'),
    ('bom','\ufeff'), ('nel-control','\u0085'), ('mvs-control','\u180e'), ('zwsp-control','\u200b')]
for name, ws in all_spaces:
    add('unicode-leading-' + name, r'\big' + ws + r'\langle x')
    add('unicode-brace-leading-' + name, r'\Big{' + ws + '(}x')
    add('unicode-brace-trailing-' + name, r'\bigg{(' + ws + '}x')
    add('unicode-brace-both-' + name, r'\Bigg{' + ws + '(' + ws + '}x')
    add('unicode-right-' + name, r'\left(x\right' + ws + ')')

contexts = [
    ('group', lambda s: '{' + s + '}'), ('sup', lambda s: 'Q^{' + s + '}'),
    ('sub', lambda s: 'Q_{' + s + '}'), ('color', lambda s: r'\color{red}' + s),
    ('bold', lambda s: r'\mathbf{' + s + '}'),
    ('text', lambda s: r'\text{A $' + s + '$ B}'),
    ('array', lambda s: r'\begin{matrix}' + s + r'&Z\end{matrix}'),
    ('scriptstyle', lambda s: r'\scriptstyle ' + s),
]
for label, formula in [('big-bom', '\\big\ufeff{\ufeff(\ufeff}x'),
                       ('left-bom', '\\left\ufeff(\\frac{a}{b}\\right\ufeff)'),
                       ('ordinary-control', r'\left(\frac{a}{b}\right)')]:
    for context_name, wrap in contexts:
        add(label + '-' + context_name, wrap(formula))
add('d2-bom-witness', 'F(x) = \\left\ufeff(\\frac{a}{b}\\right\ufeff)')
for label, formula in [
    ('empty', r'\big'), ('blank-brace', r'\big{ }'), ('bom-brace', '\\big{\ufeff}'),
    ('bad-command', r'\big\doesnotexist'), ('multiple-delimiters', r'\big{()}'),
    ('unclosed-brace', r'\big{('), ('supplementary-control', '\\big\U0001f642x'),
]:
    add('diagnostic-control-' + label, formula)

wire = ''.join(json.dumps({'tex': c['tex'], 'options': {'Display': c['display']}}, ensure_ascii=True) + '\n' for c in requests)
process = subprocess.run([node, '--jitless', str(here / 'differential/oracle.mjs'), '--asset-dir', str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split('\n') if line]
assert len(responses) == len(requests)
cases, runtime_cases = [], []
for request, response in zip(requests, responses):
    c = {**request, 'original': response}
    if response.get('error'):
        assert set(response) == {'error'}, c
        runtime_cases.append(c)
    else:
        assert set(response) == {'svg'} and response['svg'], c
        cases.append(c)
metadata = {
    'mathjaxGitCommit': 'ad8f5c21cb810236551da8c6512ba733e67357ee', 'primaryAssets': hashes,
    'methods': methods,
    'scope': 'GetDelimiter GetNext whitespace and braced JavaScript trim only; complete fresh original SVGs. Rendered diagnostics retained; runtime failures are separately recorded and excluded from SVG parity credit.',
}
payload = (json.dumps({**metadata, 'cases': cases, 'runtimeCases': runtime_cases}, ensure_ascii=True, indent=2) + '\n').encode()
with open(here / 'getdelimiter_whitespace_mathjax_3_2_2.json.gz', 'wb') as output:
    with gzip.GzipFile(fileobj=output, mode='wb', mtime=0, filename='') as compressed:
        compressed.write(payload)
print('Original SVGs:', len(cases), 'Valid:', sum('data-mjx-error' not in c['original']['svg'] for c in cases),
      'Rendered errors:', sum('data-mjx-error' in c['original']['svg'] for c in cases), 'Runtime failures:', len(runtime_cases))
