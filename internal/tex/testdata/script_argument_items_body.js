(() => {
  const options = html.inputJax[0].parseOptions;
  const f = html.inputJax[0].mmlFactory, nf = options.nodeFactory, items = options.itemFactory;
  const methods = MathJax._.input.tex.base.BaseMethods.default;
  const Parser = MathJax._.input.tex.TexParser.default;
  const NodeUtil = MathJax._.input.tex.NodeUtil.default;
  const rows = [];
  const token = text => f.create('mi', {}, [f.create('text').setText(text)]);
  const types = ['mi', 'msub', 'msup', 'msubsup', 'munder', 'mover', 'munderover',
    'msubsup-no-under', 'msubsup-no-over', 'munderover-no-under', 'munderover-no-over'];
  for (const type of types) for (const marker of ['_', '^'])
    for (const moves of [null, false, true]) for (const allowed of [false, true])
      for (const event of ['prime', 'stop']) {
        const kind = type.split('-')[0];
        const core = token('x'), under = token('i'), over = token('n');
        const kids = kind === 'msub' || kind === 'munder' ? [core, under] :
          kind === 'msup' || kind === 'mover' ? [core, over] :
          type.endsWith('no-under') ? [core] : type.endsWith('no-over') ? [core, under] : [core, under, over];
        const base = kind === 'mi' ? core : f.create(kind, {id: 'authored'}, kids);
        if (type.endsWith('no-under')) NodeUtil.setChild(base, 2, over);
        base.setProperty('kept', 17);
        if (moves !== null) base.setProperty('movesupsub', moves);
        if (allowed) base.setProperty('subsupOK', true);
        let selected;
        const p = Object.create(Parser.prototype);
        p._string = event === 'prime' ? "' ’z" : '';
        p.i = 0;
        p.configuration = options;
        p.stack = {Top: () => ({isKind: () => false}), Prev: () => base};
        p.Push = item => { selected = item; };
        const row = {type, kind, marker, moves, allowed, event, source: p.string};
        let stage = 'prepare';
        try {
          methods[marker === '_' ? 'Subscript' : 'Superscript'](p, marker);
          const prepared = selected.First;
          const ids = new Map([[base, 'base'], [core, 'core'], [under, 'under'], [over, 'over']]);
          ids.set(base, 'base');
          row.prepared = {
            kind: prepared.kind, position: selected.getProperty('position'), reused: prepared === base,
            children: prepared.childNodes.map(n => n ? ids.get(n) || 'unknown' : null),
            attributes: {...prepared.attributes.getAllAttributes()}, properties: {...prepared.getAllProperties()}
          };
          stage = 'argument';
          p.stack.Prev = () => prepared;
          p.Push = item => selected.checkItem(item);
          if (event === 'prime') {
            p.i++;
            methods.Prime(p, "'");
          } else {
            p.Push(items.create('stop'));
          }
          row.error = null;
        } catch (error) {
          row.error = {id: error.id, message: error.message};
        }
        row.stage = stage;
        row.cursor = p.i;
        row.remaining = p.string.slice(p.i);
        rows.push(row);
      }
  return {
    superscriptSource: methods.Superscript.toString(), subscriptSource: methods.Subscript.toString(),
    primeSource: methods.Prime.toString(),
    subsupItemSource: MathJax._.input.tex.base.BaseItems.SubsupItem.prototype.checkItem.toString(), rows
  };
})()
