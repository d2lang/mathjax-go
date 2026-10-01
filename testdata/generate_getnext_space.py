#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Capture unchanged frozen originals for active GetNext whitespace boundaries."""
import gzip
import hashlib
import json
import pathlib
import subprocess
import sys

directory = pathlib.Path(__file__).resolve().parent
assets, node = pathlib.Path(sys.argv[1]), sys.argv[2]
hashes = {
    'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
    'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
    'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
}
for name, expected in hashes.items():
    assert hashlib.sha256((assets / name).read_bytes()).hexdigest() == expected, name

points = [9, 10, 11, 12, 13, 32, 160, 5760, *range(8192, 8203),
          8232, 8233, 8239, 8287, 12288, 65279]
assert len(points) == 25
forms = []
for command in ['matrix', 'pmatrix', 'cases', 'eqalign', 'eqalignno', 'leqalignno', 'displaylines']:
    body = r'a&b\cr c&d'
    if command == 'cases':
        body = r'a&x>0\cr b&x<0'
    elif command in ['eqalignno', 'leqalignno']:
        body = r'a&b&1\cr c&d&2'
    elif command == 'displaylines':
        body = r'a\cr b'
    forms.append((command, 'Base.Matrix', '\\' + command + '{SPACE}{' + body + '}'))
for command in ['Braket', 'Set', 'set']:
    forms.append((command, 'Braket.Braket', '\\' + command + r'{SPACE}{\frac{a}{b}|x}'))
for command in ['dv', 'derivative', 'pdv', 'pderivative', 'partialderivative',
                'fdv', 'fderivative', 'functionalderivative']:
    tail = '{f}{SPACE}{x}'
    if command in ['pdv', 'pderivative', 'partialderivative']:
        tail += '{SPACE}{y}'
    forms.append((command, 'Physics.Derivative', '\\' + command + tail))
for command in ['dd', 'differential', 'var', 'variation']:
    forms.append((command + '-parentheses', 'Physics.Differential', '\\' + command + r'[2]{SPACE}(\frac{a}{b})'))
    forms.append((command + '-braces', 'Physics.Differential', 'a+\\' + command + '[2]{SPACE}{x}z'))
for command in ['det', 'Pr', 'tr', 'trace', 'Tr', 'Trace', 'exp', 'erf', 'sin', 'ln']:
    forms.append((command, 'Physics.Expression', '\\' + command + r'{SPACE}(\frac{a}{b})'))
for command in ['dmat', 'admat', 'diagonalmatrix', 'antidiagonalmatrix']:
    forms.append((command, 'Physics.DiagonalMatrix', r'\mqty{\CMD{SPACE}{1,2}}'.replace('CMD', command)))

pending, seen = [], set()
def add(name, tex, family):
    for display in [False, True]:
        if (tex, display) in seen:
            continue
        seen.add((tex, display))
        pending.append({'name': name + ('-display' if display else '-inline'),
                        'tex': tex, 'display': display, 'family': family})

for name, family, form in forms:
    for point in points + [133, 6158, 8203]:
        # With a non-whitespace unbraced matrix argument, alignment tokens in
        # the remaining authored group produce an unrelated Misplaced & error.
        # A one-cell group keeps the exclusion control a valid original input.
        selected = form
        if family == 'Base.Matrix' and point in [6158, 8203]:
            selected = '\\' + name + '{SPACE}{x}'
        add(name + f'-U+{point:04X}', selected.replace('{SPACE}', chr(point)), family)

witnesses = [
    ('matrix', 'Base.Matrix', '\\pmatrix\ufeff{a&b\\cr c&d}'),
    ('braket', 'Braket.Braket', '\\Braket\ufeff{\\frac{a}{b}|x}'),
    ('derivative', 'Physics.Derivative', '\\dv{f}\ufeff{x}'),
    ('differential', 'Physics.Differential', '\\dd[2]\ufeff(\\frac{a}{b})'),
    ('expression', 'Physics.Expression', '\\det\ufeff(\\frac{a}{b})'),
    ('diagonal', 'Physics.DiagonalMatrix', '\\mqty{\\dmat\ufeff{1,2}}'),
]
for name, family, tex in witnesses:
    add('d2-' + name, tex, family)
for name, family, form in [forms[0], forms[7], forms[10], forms[18], forms[26], forms[36]]:
    for point in [32, 65279, 133, 6158, 8203]:
        selected = form
        if family == 'Base.Matrix' and point in [6158, 8203]:
            selected = '\\' + name + '{SPACE}{x}'
        tex = selected.replace('{SPACE}', chr(point))
        for context_name, wrap in [('sup', lambda s: 'Q^{' + s + '}'),
                                   ('font-color', lambda s: r'\color{red}\mathbf{' + s + '}')]:
            add(name + f'-context-{context_name}-U+{point:04X}', wrap(tex), family)

wire = ''.join(json.dumps({'tex': c['tex'], 'options': {'Display': c['display']}}, ensure_ascii=True) + '\n' for c in pending)
process = subprocess.run([node, '--jitless', str(directory / 'differential/oracle.mjs'), '--asset-dir', str(assets)],
                         input=wire, text=True, capture_output=True, check=True)
responses = [json.loads(line) for line in process.stdout.split('\n') if line]
assert len(responses) == len(pending)
cases, runtime_cases, diagnostic_cases = [], [], []
for request, original in zip(pending, responses):
    case = {**request, 'original': original}
    if set(original) == {'error'}:
        assert original['error'].startswith("TypeError: Cannot read properties of null (reading '4')\n"), request
        assert 'at Object.y [as item] (mathjax.js:4:328295)' in original['error'], request
        runtime_cases.append(case)
    else:
        assert set(original) == {'svg'}, request
        (diagnostic_cases if 'data-mjx-error' in original['svg'] else cases).append(case)
source = subprocess.run([node, '--jitless', str(directory / 'generate_getnext_bindings.cjs'), str(assets)],
                        text=True, capture_output=True, check=True)
bindings = json.loads(source.stdout)
payload = {'mathjaxGitCommit': 'ad8f5c21cb810236551da8c6512ba733e67357ee', 'primaryAssets': hashes,
           'sourceBindings': bindings, 'javascriptWhitespace': points,
           'nonWhitespaceControls': [133, 6158, 8203],
           'scope': 'Complete unchanged original APIs in fresh frozen runtimes; original runtime exceptions receive no SVG parity credit.',
           'cases': cases, 'diagnosticCases': diagnostic_cases, 'runtimeCases': runtime_cases}
with (directory / 'getnext_space_mathjax_3_2_2.json.gz').open('wb') as output:
    with gzip.GzipFile(fileobj=output, mode='wb', mtime=0, filename='') as compressed:
        compressed.write((json.dumps(payload, ensure_ascii=True, indent=2) + '\n').encode())
print('Valid SVGs:', len(cases), 'Diagnostic SVGs:', len(diagnostic_cases), 'Original runtime exceptions:', len(runtime_cases))
