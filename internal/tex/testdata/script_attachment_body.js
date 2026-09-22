(()=>{
 const f=html.inputJax[0].mmlFactory,nf=html.inputJax[0].parseOptions.nodeFactory;
 const methods=MathJax._.input.tex.base.BaseMethods.default,NodeUtil=MathJax._.input.tex.NodeUtil.default,rows=[];
 const types=['msub','msup','msubsup','munder','mover','munderover','msubsup-no-under','msubsup-no-over','munderover-no-under','munderover-no-over'];
 for(const type of types)for(const marker of ['_','^'])for(const moves of [null,false,true])for(const allowed of [false,true]){
  const kind=type.split('-')[0], token=t=>f.create('mi',{},[f.create('text').setText(t)]);
  const core=token('x'),under=token('i'),over=token('n'),script=token('a');
  const kids=kind==='msub'||kind==='munder'?[core,under]:kind==='msup'||kind==='mover'?[core,over]:type.endsWith('no-under')?[core]:type.endsWith('no-over')?[core,under]:[core,under,over];
  const base=f.create(kind,{id:'authored'},kids);if(type.endsWith('no-under'))NodeUtil.setChild(base,2,over);base.setProperty('kept',17);if(moves!==null)base.setProperty('movesupsub',moves);if(allowed)base.setProperty('subsupOK',true);
  let item;const p={GetNext:()=>'',stack:{Top:()=>({isKind:()=>false}),Prev:()=>base},create:(...a)=>nf.create(...a),itemFactory:{create:(name,node)=>({name,node,setProperties(props){this.props=props;return this}})},Push:i=>item=i};
  const record={type,kind,marker,moves,allowed};
  try{
   methods[marker==='_'?'Subscript':'Superscript'](p,marker);
   const out=item.node,position=item.props.position;
   // Complete the selected source slot exactly as the SubsupItem receives its
   // argument; this probe does not invoke a renderer or alter the source method.
   NodeUtil.setChild(out,position,script);
   const isSide=['msub','msup','msubsup'].includes(out.kind),u=out.childNodes[out.sub??out.under],o=out.childNodes[out.sup??out.over];
   const ids=new Map([[base,'base'],[core,'core'],[under,'under'],[over,'over'],[script,'script']]);
   let compactKind=isSide?'msubsup':'munderover',children=[out.childNodes[0],u,o];
   if(!u){compactKind=isSide?'msup':'mover';children=[out.childNodes[0],o]}
   else if(!o){compactKind=isSide?'msub':'munder';children=[out.childNodes[0],u]}
   Object.assign(record,{error:null,sourceKind:out.kind,position,reused:out===base,compactKind,children:children.map(n=>ids.get(n)||'unknown'),childParents:children.map(n=>n.parent===out),attributes:out.attributes.getAllAttributes(),properties:out.getAllProperties(),baseRetainedWhole:children[0]===base});
  }catch(error){record.error=error.message}
  rows.push(record);
 }
 return {subscriptSource:methods.Subscript.toString(),superscriptSource:methods.Superscript.toString(),rows};
})()
