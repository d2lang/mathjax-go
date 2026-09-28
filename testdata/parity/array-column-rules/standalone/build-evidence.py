import argparse,gzip,hashlib,html,json,pathlib,subprocess

parser=argparse.ArgumentParser(description="Reproduce the standalone comparison with explicit probes and freshly extracted original assets.")
parser.add_argument('--baseline',required=True,type=pathlib.Path,help='Baseline JSONL probe executable')
parser.add_argument('--candidate',required=True,type=pathlib.Path,help='Candidate JSONL probe executable')
parser.add_argument('--output-dir',required=True,type=pathlib.Path,help='Separate directory already populated by extract-original.cjs')
parser.add_argument('--reference-dir',type=pathlib.Path,default=pathlib.Path(__file__).resolve().parent,help='Preserved evidence directory (default: directory containing this script)')
parser.add_argument('--original-corpus',type=pathlib.Path,help='Optional independent JSON or JSON.gz corpus with cases[].{tex,display,original.svg}')
args=parser.parse_args()
A=args.output_dir.resolve();R=args.reference_dir.resolve()
if A==R or A==pathlib.Path(__file__).resolve().parent:
 parser.error('Use a separate output directory to preserve the published evidence')
if not (A/'original-css-extraction.json').is_file():
 parser.error('Run extract-original.cjs into --output-dir first')
ex=json.loads((A/'original-css-extraction.json').read_text());rows=ex['cases'];hash=lambda b:hashlib.sha256(b).hexdigest()
binaries={'before':args.baseline.resolve(),'after':args.candidate.resolve()}
for side,binary in binaries.items():
 out=subprocess.check_output([str(binary)],input=''.join(json.dumps({'tex':r['tex'],'display':True,'options':{'Display':True}})+'\n'for r in rows),text=True).rstrip('\n').split('\n');assert len(out)==len(rows)
 for r,line in zip(rows,out):
  response=json.loads(line);assert'svg'in response and not response.get('error');(A/(r['id']+'-'+side+'.svg')).write_text(response['svg']);r[side+'SVG_SHA256']=hash(response['svg'].encode())
for r in rows:assert(A/(r['id']+'-after.svg')).read_bytes()==(A/(r['id']+'-original.svg')).read_bytes()
# The preserved evidence, not a Go-generated golden, binds every regenerated
# renderer output and the stylesheet extracted directly from original MathJax.
reference_files=['original-mathjax-svg.css','original-mathjax-svg-style.html','original-css-extraction.json']
reference_files += [r['id']+'-'+side+'.svg' for r in rows for side in ['before','after','original']]
for name in reference_files:
 assert (A/name).read_bytes()==(R/name).read_bytes(), 'Reference bytes changed: '+name
preserved_corpus_matched=None
if args.original_corpus:
 opener=gzip.open if args.original_corpus.suffix=='.gz' else open
 with opener(args.original_corpus,'rt') as stream: corpus=json.load(stream)['cases']
 by={(r['tex'],r['display']):r for r in corpus}
 for r in rows: assert by[r['tex'],True]['original']['svg']==(A/(r['id']+'-original.svg')).read_text()
 preserved_corpus_matched=True
(A/'comparison-layout.css').write_bytes((R/'comparison-layout.css').read_bytes())
text={
'outer-and-dashed':('Restore the outer frames and dashed inner separator','The corrected output matches the original: two outside lines and a dashed internal line.'),
'unmarked-gap':('Keep the unmarked column gap empty','The corrected output matches the original: one internal rule, with no rule in the second gap.'),
}
def section(r):
 title,note=text[r['id']];panels=[]
 for side,label in [('before','BEFORE · Go on PR157'),('after','AFTER · column-rule fix'),('original','ORIGINAL · MathJax 3.2.2')]:
  raw=(A/(r['id']+'-'+side+'.svg')).read_text()
  panels.append('<article class="panel '+side+'"><div class="label">'+label+'</div><div class="math-stage"><mjx-container class="MathJax" jax="SVG" display="true">'+raw+'</mjx-container></div></article>')
 return '<section class="case" id="'+r['id']+'"><h2>'+title+'</h2><div class="source">'+html.escape(r['tex'])+'</div><div class="panels">'+''.join(panels)+'</div><p class="note">'+note+'</p></section>'
def page(selected):return '<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Array column rules — original MathJax stylesheet</title><link rel="stylesheet" href="comparison-layout.css"><link rel="stylesheet" href="original-mathjax-svg.css"></head><body><main><h1>Array column rules · standalone SVG comparison</h1><p>The original frozen MathJax SVG stylesheet is applied equally to all three panels, at the same font size and scale.</p>'+''.join(section(r)for r in selected)+'<p class="footer">Supplemental browser evidence. D2 does not include this MathJax stylesheet, so its internal table rules appear much thinner. These are the unchanged raw renderer SVGs inside the original MathJax container; the fixed and original SVG files are byte-identical. The separate D2 example demonstrates the outside frames.</p></main></body></html>'
(A/'index.html').write_text(page(rows))
for r in rows:(A/(r['id']+'.html')).write_text(page([r]))
for name in ['index.html']+[r['id']+'.html' for r in rows]:
 assert (A/name).read_bytes()==(R/name).read_bytes(), 'HTML changed: '+name
receipt={
 'binaries':{k:{'path':str(p),'sha256':hash(p.read_bytes())}for k,p in binaries.items()},
 'referenceDirectory':str(R),
 'referenceManifestSHA256':hash((R/'manifest.json').read_bytes()),
 'verifiedUnmodifiedFiles':{name:hash((A/name).read_bytes()) for name in reference_files+['comparison-layout.css','index.html']+[r['id']+'.html' for r in rows]},
 'afterMatchesOriginalBytes':True,
 'independentOriginalCorpusMatched':preserved_corpus_matched,
 'note':'This verifies the two evidence outputs against preserved originals; rebuilt probe hashes may differ from the original publication binaries.',
}
(A/'reproduction-verification.json').write_text(json.dumps(receipt,indent=2)+'\n')
print(json.dumps({'cases':len(rows),'allReferenceBytesExact':True,'afterOriginalExact':True,'html':str(A/'index.html'),'receipt':str(A/'reproduction-verification.json')},indent=2))
