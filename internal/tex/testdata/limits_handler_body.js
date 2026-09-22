(()=>{
 const f=html.inputJax[0].mmlFactory,nf=html.inputJax[0].parseOptions.nodeFactory;
 const Limits=MathJax._.input.tex.base.BaseMethods.default.Limits;
 const rows=[];
 for(const kind of ['msub','msup','msubsup','munder','mover','munderover']) for(const enabled of [false,true]) {
  const base=f.create('mo', {movablelimits:true},[f.create('text').setText('∑')]);base.texClass=1;
  const kids=[base,f.create('mi',{},[f.create('text').setText('i')])];if(['msubsup','munderover'].includes(kind))kids.push(f.create('mi',{},[f.create('text').setText('n')]));
  const original=f.create(kind,{id:'old-wrapper',movablelimits:true},kids);original.setProperty('movesupsub',true);original.setProperty('oldProperty','keep?');
  const top={Last:original};const calls=[];
  const p={stack:{Prev:()=>top.Last,Top:()=>top},currentCS:enabled?'\\limits':'\\nolimits',create:(...a)=>{calls.push(a);return nf.create(...a)}};
  Limits(p,p.currentCS,enabled?1:0);
  const out=top.Last;
  rows.push({kind,enabled,createCalls:calls,resultKind:out.kind,replaced:out!==original,childCount:out.childNodes.length,children:out.childNodes.map((n,i)=>({kind:n.kind,text:n.getText?.(),same:n===kids[i],newParent:n.parent===out})),attributes:out.attributes.getAllAttributes(),properties:out.getAllProperties(),coreKind:out.coreMO().kind,coreAttrs:out.coreMO().attributes.getAllAttributes(),coreProperties:out.coreMO().getAllProperties(),oldChildrenRetained:original.childNodes.every((n,i)=>n===kids[i])});
 }
 return {limitsSource:Limits.toString(),rows};
})()
