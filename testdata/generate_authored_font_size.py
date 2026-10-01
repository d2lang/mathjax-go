#!/usr/bin/env python3
# Copyright 2017-2022 The MathJax Consortium
# SPDX-License-Identifier: Apache-2.0
# Regenerate with D2's frozen assets; oracle.mjs creates a fresh runtime per case.

import argparse
import itertools
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--asset-dir', required=True)
parser.add_argument('--node', default='node')
args = parser.parse_args()
here = Path(__file__).resolve().parent
attributes = [
    ('css-percent', "style='font-size:200%'"),
    ('css-em', "style='font-size:2em'"),
    ('css-absolute', "style='font-size:20px'"),
    ('css-named', "style='font-size:small'"),
    ('fontsize-percent', "fontsize='150%'"),
    ('fontsize-em', "fontsize='2em'"),
    ('attribute-precedes-css', "fontsize='150%',style='font-size:2em'"),
    ('authored-mathsize-precedes-css', "mathsize='125%',style='font-size:2em'"),
    ('authored-mathsize-precedes-fontsize', "mathsize='125%',fontsize='2em'"),
    ('css-with-color', "style='font-size:200%; color:red'"),
    ('duplicate-css', "style='font-size:150%; font-size:2em'"),
    ('unscaled-control', ''),
]
contexts = [('plain', '%s'), ('subscript', r'x_{%s}'),
            ('huge', r'\Huge{%s}'), ('fraction', r'\frac{%s}{z}')]
cases = []
for kind, (label, attrs), (context, form), display in itertools.product(
        ['mi', 'mn', 'mo', 'mtext', 'ms'], attributes, contexts, [True, False]):
    token = r'\mmlToken{' + kind + '}' + ('[' + attrs + ']' if attrs else '') + '{x}'
    cases.append({'name': '-'.join([kind, label, context, 'display' if display else 'inline']),
                  'tex': form % token, 'display': display})
requests = ''.join(json.dumps({'tex': c['tex'], 'options': {'Display': c['display']}}) + '\n'
                   for c in cases)
result = subprocess.run([args.node, str(here / 'differential/oracle.mjs'),
                         '--asset-dir', args.asset_dir], input=requests, text=True,
                        capture_output=True, check=True)
references = [json.loads(line) for line in result.stdout.splitlines()]
if len(references) != len(cases):
    raise RuntimeError('oracle reference count differs')
for case, reference in zip(cases, references):
    if 'error' in reference or 'data-mml-node="merror"' in reference.get('svg', ''):
        raise RuntimeError(f"oracle rejected {case['name']}: {reference}")
    case['svg'] = reference['svg']
fixture = {'mathjaxGitCommit': 'ad8f5c21cb810236551da8c6512ba733e67357ee',
           'oracle': 'D2 frozen MathJax 3.2.2, fresh runtime for each case', 'cases': cases}
(here / 'authored_font_size_mathjax_3_2_2.json').write_text(
    json.dumps(fixture, ensure_ascii=False, indent=2) + '\n')
