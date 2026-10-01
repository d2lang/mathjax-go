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
inventories = [('border_colors',
  [('rgb-spaced', 'border:2px solid rgb(255, 0, 0)'),
   ('rgb-tight', 'border:2px solid rgb(255,0,0)'),
   ('rgba-spaced', 'border:2px solid rgba(0, 128, 255, .5)'),
   ('rgba-tight', 'border:2px solid rgba(0,128,255,.5)'),
   ('hsl-spaced', 'border:2px solid hsl(120, 50%, 50%)'),
   ('hsl-tight', 'border:2px solid hsl(120,50%,50%)'),
   ('hsla-spaced', 'border:2px solid hsla(240, 50%, 50%, .5)'),
   ('colour-first', 'border:rgb(255, 0, 0) solid 2px'),
   ('colour-middle', 'border:solid rgb(255, 0, 0) 2px'),
   ('global-colour', 'border-width:2px; border-style:solid; border-color:rgb(255, 0, 0)'),
   ('global-pair',
    'border-width:2px; border-style:solid; border-color:rgb(255, 0, 0) rgb(0, 0, 255)'),
   ('side-component',
    'border-bottom-width:2px; border-bottom-style:solid; border-bottom-color:rgb(255, 0, 0)'),
   ('side-shorthand', 'border-bottom:2px solid rgb(255, 0, 0)'),
   ('named-control', 'border:2px solid red'),
   ('hex-control', 'border:2px solid #8000ff'),
   ('unit-control', 'border:.2em solid red'),
   ('witness', 'border:rgb(255, 0, 0) solid 12px')]),
 ('border_paint',
  [('none', 'border:2px none red'),
   ('hidden', 'border:2px hidden red'),
   ('top-none', 'border-top:2px none red'),
   ('right-hidden', 'border-right:2px hidden red'),
   ('bottom-none', 'border-bottom:2px none red'),
   ('left-hidden', 'border-left:2px hidden red'),
   ('global-none', 'border-width:2px; border-style:none; border-color:red'),
   ('global-hidden', 'border-width:2px; border-style:hidden; border-color:red'),
   ('none-override', 'border:2px none red; border-top-style:dashed'),
   ('hidden-override', 'border:2px hidden red; border-bottom-style:dotted'),
   ('none-zero', 'border:0px none red'),
   ('hidden-zero', 'border:0px hidden red'),
   ('double-control', 'border:2px double red'),
   ('default-control', 'border-width:2px; border-color:red')])]
for name, styles in inventories:
    cases = []
    for kind, (label, style), display in itertools.product(['mi', 'mtext'], styles, [True, False]):
        cases.append({'name': kind + '-' + label + ('-display' if display else '-inline'),
                      'tex': r"\mmlToken{" + kind + "}[style='" + style + "']{x}", 'display': display})
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
    (here / (name + '_mathjax_3_2_2.json')).write_text(json.dumps(fixture, indent=2) + '\n')
