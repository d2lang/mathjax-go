# SPDX-License-Identifier: Apache-2.0
# Regenerate only from D2's frozen original. Usage:
# python3 testdata/generate_matrix_fence_spacing.py PINNED_ASSETS --node NODE
import argparse,gzip,hashlib,json,pathlib,subprocess
HERE=pathlib.Path(__file__).resolve().parent
p=argparse.ArgumentParser()
p.add_argument('assets',type=pathlib.Path)
p.add_argument('--node',default='node')
a=p.parse_args()
hashes={'polyfills.js':'7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01','mathjax.js':'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869','setup.js':'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881'}
for name,want in hashes.items():
 if hashlib.sha256((a.assets/name).read_bytes()).hexdigest()!=want:raise ValueError('unverified '+name)
cases=[]; seen=set()
def add(group,name,tex):
 for display in [False,True]:
  key=(tex,display)
  if key in seen:continue
  seen.add(key);cases.append(dict(group=group,name=f'{group}-{name}-'+('display' if display else 'inline'),tex=tex,display=display))
commands=['matrix','array','pmatrix','cases','eqalign','displaylines','eqalignno','leqalignno']
def matrix(name,body=None):
 if body is None:body=r'x&if x>0\cr 0&otherwise' if name=='cases' else r'a&b\cr c&d'
 return '\\'+name+'{'+body+'}'
neighbors={'text':r'\text{Matrix: }','textplain':r'\text{label}','ord':'x','number':'2','bin':'+','rel':'=','op':r'\sum','function':r'\sin','open':'(','close':')','punct':',','inner':r'\mathinner{z}','forced-ord':r'\mathord{z}','forced-op':r'\mathop{z}','forced-bin':r'\mathbin{z}','forced-rel':r'\mathrel{z}','forced-open':r'\mathopen{(}','forced-close':r'\mathclose{)}','forced-punct':r'\mathpunct{,}','space':r'\quad','fraction':r'\frac{p}{q}'}
for name in commands:
 m=matrix(name)
 add('bare',name,m)
 for label,n in neighbors.items():
  add('preceding',name+'-'+label,n+m)
  add('following',name+'-'+label,m+n)
 for label,left,right in [('text',r'\text{Matrix: }',r'\text{ done}'),('ord','x','y'),('bin','x+','+y'),('rel','x=','=y'),('op',r'\sum',r'\sum'),('function',r'\sin',r'\cos'),('inner',r'\mathinner{z}',r'\mathinner{w}')]:add('surrounding',name+'-'+label,left+m+right)
 for label,tex in [('mathsf',r'\mathsf{x'+m+'y}'),('mathbf',r'\mathbf{x'+m+'y}'),('mathit',r'\mathit{x'+m+'y}'),('boldsymbol',r'\boldsymbol{x'+m+'y}'),('sup',r'X^{x'+m+'y}'),('sub',r'X_{x'+m+'y}'),('denominator',r'\frac{z}{x'+m+'y}'),('root',r'\sqrt{x'+m+'y}'),('postscript',r'x'+m+r'_i^jy'),('braced',r'x{'+m+'}y'),('scriptstyle',r'\scriptstyle x'+m+'y'),('textstyle',r'\textstyle x'+m+'y')]:add('scope',name+'-'+label,tex)
 for label,body in [('single','a'),('empty',''),('fraction',r'\frac{a}{b}&\sqrt{c}\cr g&h'),('font',r'\mathbf{a}&\mathsf{b}\cr c&d')]:
  if name=='cases' and label in ['fraction','font']:continue
  add('body',name+'-'+label,r'x'+matrix(name,body)+'y')
for env in ['matrix','pmatrix','bmatrix','Bmatrix','vmatrix','Vmatrix','cases','smallmatrix','array','aligned']:
 m='\\begin{'+env+'}'+('{cc}' if env=='array' else '')+r'a&b\\c&d'+'\\end{'+env+'}'
 for label,tex in [('bare',m),('text',r'\text{Matrix: }'+m),('surrounding','x'+m+'y'),('operator',r'\sin'+m+r'\cos'),('font',r'\mathbf{x'+m+'y}'),('script',r'X^{x'+m+'y}')]:add('environment',env+'-'+label,tex)
for name in commands:
 for label,tex in [('missing','\\'+name),('close','\\'+name+'{a'),('unbraced-script','X^\\'+name+'{a}'),('extra-close','\\'+name+'{a}}')]:add('diagnostic',name+'-'+label,tex)

records=[];qualified=[];runtime=[]
for start in range(0,len(cases),24):
 batch=cases[start:start+24]
 inputs=''.join(json.dumps({'tex':c['tex'],'options':{'Display':c['display'],'Em':16,'Ex':8}})+'\n' for c in batch)
 run=subprocess.run([a.node,'--jitless',str(HERE/'differential/oracle.mjs'),'--asset-dir',str(a.assets)],input=inputs,text=True,capture_output=True,check=True)
 outputs=[json.loads(line) for line in run.stdout.rstrip('\n').split('\n')]
 if len(outputs)!=len(batch):raise ValueError('incomplete original output')
 for c,original in zip(batch,outputs):
  if original.get('error'):runtime.append({**c,'original':original});continue
  svg=original['svg'];valid='data-mml-node="merror"' not in svg
  if r'data-mjx-error="Undefined control sequence \boldsymbol"' in svg:
   if valid:raise ValueError('qualified diagnostic is valid')
   qualified.append({**c,'original':original,'qualification':'Frozen D2 original does not register \\boldsymbol, while Go accepts it; excluded from exact SVG assertions and parity credit.'})
  else:records.append({**c,'originalValid':valid,'svg':svg})
 print('fresh original',min(start+24,len(cases)),'/',len(cases),flush=True)
fixture={'mathjaxGitCommit':'ad8f5c21cb810236551da8c6512ba733e67357ee','oracleAssetsSHA256':hashes,'freshOriginalPerCase':True,'counts':{'total':len(cases),'fullSVGAssertions':len(records),'valid':sum(c['originalValid'] for c in records),'diagnostics':sum(not c['originalValid'] for c in records),'qualifiedDiagnostics':len(qualified),'originalRuntimeErrors':len(runtime)},'cases':records,'qualifiedDiagnostics':qualified,'originalRuntimeErrors':runtime}
(HERE/'matrix_fence_spacing_mathjax_3_2_2.json.gz').write_bytes(gzip.compress((json.dumps(fixture,ensure_ascii=False,indent=2)+'\n').encode(),mtime=0))
