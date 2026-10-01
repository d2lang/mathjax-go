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
attributes = [('top-width-override', "style='border:2px solid red; border-top-width:4px'"),
 ('top-width-reset', "style='border:2px solid red; border-top-width:'"),
 ('top-width-fresh', "style='border-top-width:4px'"),
 ('top-style-override', "style='border:2px solid red; border-top-style:dotted'"),
 ('top-style-reset', "style='border:2px solid red; border-top-style:'"),
 ('top-style-fresh', "style='border-top-style:dotted'"),
 ('top-color-override', "style='border:2px solid red; border-top-color:blue'"),
 ('top-color-reset', "style='border:2px solid red; border-top-color:'"),
 ('top-color-fresh', "style='border-top-color:blue'"),
 ('right-width-override', "style='border:2px solid red; border-right-width:4px'"),
 ('right-width-reset', "style='border:2px solid red; border-right-width:'"),
 ('right-width-fresh', "style='border-right-width:4px'"),
 ('right-style-override', "style='border:2px solid red; border-right-style:dotted'"),
 ('right-style-reset', "style='border:2px solid red; border-right-style:'"),
 ('right-style-fresh', "style='border-right-style:dotted'"),
 ('right-color-override', "style='border:2px solid red; border-right-color:blue'"),
 ('right-color-reset', "style='border:2px solid red; border-right-color:'"),
 ('right-color-fresh', "style='border-right-color:blue'"),
 ('bottom-width-override', "style='border:2px solid red; border-bottom-width:4px'"),
 ('bottom-width-reset', "style='border:2px solid red; border-bottom-width:'"),
 ('bottom-width-fresh', "style='border-bottom-width:4px'"),
 ('bottom-style-override', "style='border:2px solid red; border-bottom-style:dotted'"),
 ('bottom-style-reset', "style='border:2px solid red; border-bottom-style:'"),
 ('bottom-style-fresh', "style='border-bottom-style:dotted'"),
 ('bottom-color-override', "style='border:2px solid red; border-bottom-color:blue'"),
 ('bottom-color-reset', "style='border:2px solid red; border-bottom-color:'"),
 ('bottom-color-fresh', "style='border-bottom-color:blue'"),
 ('left-width-override', "style='border:2px solid red; border-left-width:4px'"),
 ('left-width-reset', "style='border:2px solid red; border-left-width:'"),
 ('left-width-fresh', "style='border-left-width:4px'"),
 ('left-style-override', "style='border:2px solid red; border-left-style:dotted'"),
 ('left-style-reset', "style='border:2px solid red; border-left-style:'"),
 ('left-style-fresh', "style='border-left-style:dotted'"),
 ('left-color-override', "style='border:2px solid red; border-left-color:blue'"),
 ('left-color-reset', "style='border:2px solid red; border-left-color:'"),
 ('left-color-fresh', "style='border-left-color:blue'"),
 ('witness', "style='border:2px solid red; border-bottom-width:12px'"),
 ('later-shorthand',
  "style='border:2px solid red; border-bottom-width:12px; border:1px solid blue'"),
 ('later-components',
  "style='border:2px solid red; border-bottom-width:12px; border-width:1px'"),
 ('control', "style='border:2px solid red'")]
contexts = [('plain', '%s'), ('subscript', r'x_{%s}')]
cases = []
for kind, (label, attrs), (context, form), display in itertools.product(
        ['mi', 'mtext'], attributes, contexts, [True, False]):
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
(here / 'border_components_mathjax_3_2_2.json').write_text(
    json.dumps(fixture, ensure_ascii=False, indent=2) + '\n')
