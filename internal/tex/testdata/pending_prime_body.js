(() => {
 const f=html.inputJax[0].mmlFactory,nf=html.inputJax[0].parseOptions.nodeFactory;
 const methods=MathJax._.input.tex.base.BaseMethods.default;
 const items=MathJax._.input.tex.base.BaseItems,NodeUtil=MathJax._.input.tex.NodeUtil.default;
 const token=(kind,text)=>f.create(kind,{},[f.create('text').setText(text)]);
 const types=['mi','msup','msub','msubsup-no-over','msubsup-full','mover','munder','munderover-no-over','munderover-full'];
 const rows=[];
 function setup(type,moves,allowed){
  const core=token('mi','x'),under=token('mi','i'),over=token('mi','n'),script=token('mi','a');
  const kind=type.split('-')[0];
  const kids=kind==='mi'?[]:kind==='msup'||kind==='mover'?[core,over]:kind==='msub'||kind==='munder'||type.endsWith('no-over')?[core,under]:[core,under,over];
  const base=kind==='mi'?core:f.create(kind,{id:'authored'},kids);
  base.setProperty('kept',17);if(moves!==null)base.setProperty('movesupsub',moves);if(allowed)base.setProperty('subsupOK',true);
  const ids=new Map([[base,'base'],[core,'core'],[under,'under'],[over,'over'],[script,'script']]);ids.set(base,'base');
  let pair;const p={i:1,GetNext:()=>'',stack:{Prev:()=>base},create:(...a)=>nf.create(...a),itemFactory:{create:(name,...nodes)=>({name,nodes})},Push:item=>pair=item};
  methods.Prime(p,"'");const prime=pair.nodes[1];ids.set(prime,'prime');return {base,core,under,over,script,prime,ids};
 }
 const describe=(n,ids)=>({identity:ids.get(n)||'new',kind:n.kind,text:n.kind==='text'?n.getText():null,attributes:n.attributes?{...n.attributes.getAllAttributes()}:{},properties:{...n.getAllProperties()},children:n.childNodes.map(c=>c?describe(c,ids):null)});
 for(const type of types)for(const moves of [null,false,true])for(const allowed of [false,true])for(const action of ['finalize','_','^']){
  const row={type,moves,allowed,action};let s;
  try{
   s=setup(type,moves,allowed);const{base,prime,script,ids}=s;
   if(action==='finalize'){
    const next={kind:'mml'};const r=items.PrimeItem.prototype.checkItem.call({Peek:()=>[base,prime],create:(...a)=>nf.create(...a)},next);
    const out=r[0][0];Object.assign(row,{error:null,reused:out===base,keptNext:r[0][1]===next,success:r[1],tree:describe(out,ids),parentChecks:out.childNodes.filter(Boolean).map(n=>n.parent===out)});
   }else{
    let selected,popped=false;const p={GetNext:()=>'',stack:{Top:()=>({isKind:k=>k==='prime',Peek:()=>[base,prime]}),Pop:()=>{popped=true},Prev:()=>{throw Error('Pending pair was not selected')}},create:(...a)=>nf.create(...a),itemFactory:{create:(name,node)=>({name,node,setProperties(props){this.props=props;return this}})},Push:item=>selected=item};
    methods[action==='_'?'Subscript':'Superscript'](p,action);
    const event={First:script,isKind:k=>k==='mml'};
    const ctx={First:selected.node,getProperty:n=>selected.props[n],create:(...a)=>nf.create(...a),factory:{create:(name,node)=>({name,First:node})}};
    const r=items.SubsupItem.prototype.checkItem.call(ctx,event),out=r[0][0].First;
    Object.assign(row,{error:null,pendingPopped:popped,position:selected.props.position,reused:out===base,success:r[1],tree:describe(out,ids),primeProperties:{...prime.getAllProperties()},primeAttributes:{...prime.attributes.getAllAttributes()},parentChecks:out.childNodes.filter(Boolean).map(n=>n.parent===out)});
   }
  }catch(error){row.error=error.message;row.stage=s?'consume':'prime'}
  rows.push(row);
 }
 return{primeSource:methods.Prime.toString(),subscriptSource:methods.Subscript.toString(),superscriptSource:methods.Superscript.toString(),primeItemSource:items.PrimeItem.prototype.checkItem.toString(),subsupItemSource:items.SubsupItem.prototype.checkItem.toString(),rows};
})()
