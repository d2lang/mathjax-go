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
attributes = [('percent', "style='padding:10%'"),
 ('left-percent', "style='padding-left:25%'"),
 ('right-percent', "style='padding-right:5%'"),
 ('top-percent', "style='padding-top:50%'"),
 ('bottom-percent', "style='padding-bottom:12%'"),
 ('negative', "style='padding:-.1em'"),
 ('negative-left', "style='padding-left:-.2em'"),
 ('zero', "style='padding:0'"),
 ('em-control', "style='padding:1em'"),
 ('px-control', "style='padding:2px'"),
 ('border-percent', "style='border:2px solid red; padding:10%'"),
 ('border-left-px', "style='border-left:1px solid red; padding-left:2px'"),
 ('background-witness', "style='padding:50%; background-color:yellow'")]
contexts = [('plain', '%s'), ('subscript', r'x_{%s}'),
            ('huge', r'\Huge{%s}'), ('fraction', r'\frac{%s}{z}')]
cases = []
for kind, (label, attrs), (context, form), display in itertools.product(
        ['mi', 'mo', 'mtext'], attributes, contexts, [True, False]):
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
(here / 'style_lengths_mathjax_3_2_2.json').write_text(
    json.dumps(fixture, ensure_ascii=False, indent=2) + '\n')
