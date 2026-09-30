// SPDX-License-Identifier: Apache-2.0
// node --jitless testdata/generate_paired_tag_priority.cjs PINNED_ASSETS SCRATCH_OUTPUT
// Only the frozen original supplies outputs. No Go process is invoked.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const zlib = require('node:zlib'), vm = require('node:vm'), {spawnSync} = require('node:child_process');
const expected = {
  inputs: 1446, originalSVGs: 1432, originalRuntime: 14,
  passiveInputs: 168, bracketPassiveInputs: 72, literalBindings: 126, literalInputs: 12
};
const sha = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const assets = process.argv[2], output = process.argv[3];
if (!assets || !output || path.resolve(output) === path.resolve(__dirname)) {
  throw Error('use PINNED_ASSETS and a separate SCRATCH_OUTPUT');
}
fs.mkdirSync(output, {recursive: true});
const name = 'paired_tag_priority_mathjax_3_2_2.json.gz';
const input = fs.readFileSync(path.join(__dirname, name));
const fixture = JSON.parse(zlib.gunzipSync(input));
if (fixture.mathjaxGitCommit !== 'ad8f5c21cb810236551da8c6512ba733e67357ee' ||
    fixture.cases.length !== expected.inputs) throw Error('unbound original input union');
const scripts = Object.entries(fixture.originalAssets).map(([name, hash]) => {
  const bytes = fs.readFileSync(path.join(assets, name));
  if (sha(bytes) !== hash) throw Error('unverified source asset: ' + name);
  return new vm.Script(bytes.toString(), {filename: name});
});
const firstLine = "TypeError: Cannot read properties of null (reading '4')";
const freshByInput = new Map(), runtimeObservations = [];
let svgCount = 0;
for (let start = 0; start < fixture.cases.length; start += 24) {
  const cases = fixture.cases.slice(start, start + 24);
  const wire = cases.map(c => JSON.stringify({tex: c.tex, display: c.display,
    options: {display: c.display, Display: c.display}})
    .replace(/\u2028/g, '\\u2028').replace(/\u2029/g, '\\u2029')).join('\n') + '\n';
  const run = spawnSync(process.execPath, ['--jitless', path.join(__dirname, 'differential/oracle.mjs'),
    '--asset-dir', assets], {input: wire, encoding: 'utf8', maxBuffer: 64 * 1024 * 1024});
  if (run.status !== 0) throw Error('original process failed: ' + run.stderr);
  const rows = run.stdout.replace(/\n$/, '').split('\n').map(JSON.parse);
  if (rows.length !== cases.length) throw Error('incomplete original capture');
  rows.forEach((fresh, i) => {
    const c = cases[i], key = JSON.stringify([c.tex, c.display]);
    if (freshByInput.has(key)) throw Error('duplicate original input');
    freshByInput.set(key, fresh);
    if (c.original.svg) {
      // Regenerate every complete original SVG in the public union.
      if (c.partition !== 'strict-svg' ||
          JSON.stringify(fresh) !== JSON.stringify(c.original)) {
        throw Error('complete original SVG changed: ' + c.name);
      }
      c.original = fresh;
      svgCount++;
    } else {
      if (c.partition !== 'runtime-bounded-nel' ||
          !c.tex.includes('\u0085') || fresh.svg || typeof fresh.error !== 'string' ||
          c.original.error.split('\n')[0] !== firstLine || fresh.error.split('\n')[0] !== firstLine) {
        throw Error('original runtime contract changed: ' + c.name);
      }
      // Keep complete historical errors, including their original stack paths.
      runtimeObservations.push({name: c.name, tex: c.tex, display: c.display,
        preservedErrorSHA256: sha(c.original.error), fresh});
    }
  });
}
if (svgCount !== expected.originalSVGs || runtimeObservations.length !== expected.originalRuntime ||
    freshByInput.size !== expected.inputs) throw Error('original partition changed');
const regenerated = zlib.gzipSync(JSON.stringify(fixture, null, 2) + '\n', {level: 9});
fs.writeFileSync(path.join(output, name), regenerated);
const reports = [{file: name, inputs: expected.inputs, fullOriginalSVGs: svgCount,
  runtimeObjectsPreserved: runtimeObservations.length, inputSHA256: sha(input),
  regeneratedSHA256: sha(regenerated), byteIdentical: input.equals(regenerated)}];

const passiveFile = 'paired_tag_priority_observations.json.gz';
const passiveBytes = fs.readFileSync(path.join(__dirname, passiveFile));
const passive = JSON.parse(zlib.gunzipSync(passiveBytes));
if (passive.mathjaxGitCommit !== fixture.mathjaxGitCommit || passive.cases.length !== expected.passiveInputs ||
    JSON.stringify(passive.assetSHA256) !== JSON.stringify(fixture.originalAssets)) throw Error('unbound passive observations');
const passiveInputs = new Set(), selectedMapEvents = {};
for (const request of passive.cases) {
  const key = JSON.stringify([request.tex, request.display]);
  if (passiveInputs.has(key)) throw Error('duplicate passive input');
  passiveInputs.add(key);
  const context = vm.createContext({console, request, events: []});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  new vm.Script(`(() => {
    const names = new Set(['tag', 'notag', 'nonumber', 'label', 'ref', 'refeq', 'eqref', 'wrap']);
    const proto = MathJax._.input.tex.MapHandler.SubHandler.prototype, original = proto.parse;
    proto.parse = function(input) {
      const [parser, symbol] = input;
      if (names.has(symbol)) {
        const matches = Array.from(this._configuration).filter(x => x.item.contains(symbol))
          .map(x => ({name: x.item.name, priority: x.priority}));
        if (matches.length) events.push({symbol, selected: matches[0], matches,
          source: parser.string, cursor: parser.i, currentCS: parser.currentCS});
      }
      return original.apply(this, arguments);
    };
  })()`).runInContext(context);
  const observed = new vm.Script('({output:{svg:adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))},events})').runInContext(context);
  const fresh = freshByInput.get(key);
  if (!fresh || JSON.stringify(fresh) !== JSON.stringify(request.original) ||
      JSON.stringify(observed.output) !== JSON.stringify(fresh) ||
      JSON.stringify(observed) !== JSON.stringify(request.observed)) throw Error('passive full source record changed: ' + request.tex);
  for (const e of observed.events) {
    if (JSON.stringify(e.selected) !== JSON.stringify(e.matches[0])) throw Error('selected map changed');
    selectedMapEvents[e.selected.name] = (selectedMapEvents[e.selected.name] || 0) + 1;
  }
}
const expectedEvents = {'mathtools-paired-delims':154, 'AMSmath-macros':12, 'macros':12, 'mathtools-macros':4};
if (Object.keys(selectedMapEvents).length !== Object.keys(expectedEvents).length ||
    Object.entries(expectedEvents).some(([name, count]) => selectedMapEvents[name] !== count)) throw Error('priority event inventory changed');
fs.writeFileSync(path.join(output, passiveFile), passiveBytes);
reports.push({file: passiveFile, observations: passiveInputs.size, additionalPublicInputs: 0,
  inputSHA256: sha(passiveBytes), regeneratedSHA256: sha(passiveBytes), byteIdentical: true});

const bracketFile = 'paired_tag_priority_bracket_observations.json.gz';
const bracketBytes = fs.readFileSync(path.join(__dirname, bracketFile));
const brackets = JSON.parse(zlib.gunzipSync(bracketBytes));
if (brackets.sourcePin !== fixture.mathjaxGitCommit || brackets.cases.length !== expected.bracketPassiveInputs ||
    JSON.stringify(brackets.assetSHA256) !== JSON.stringify(fixture.originalAssets)) throw Error('unbound bracket observations');
for (const request of brackets.cases) {
  const key = JSON.stringify([request.tex, request.display]);
  if (passiveInputs.has(key)) throw Error('overlapping or duplicate passive input');
  passiveInputs.add(key);
  const context = vm.createContext({console, request, events: []});
  context.globalThis = context;
  for (const script of scripts) script.runInContext(context);
  new vm.Script(`(() => {
    const proto = MathJax._.input.tex.TexParser.default.prototype, original = proto.GetBrackets;
    proto.GetBrackets = function(name, defaultValue) {
      events.push({name, currentCS: this.currentCS, source: this.string, cursor: this.i});
      return original.apply(this, arguments);
    };
  })()`).runInContext(context);
  const observed = new vm.Script('({output:{svg:adaptor.innerHTML(html.convert(request.tex,{display:request.display,em:16,ex:8}))},events})').runInContext(context);
  const fresh = freshByInput.get(key);
  if (!fresh || JSON.stringify(fresh) !== JSON.stringify(request.original) ||
      JSON.stringify(observed.output) !== JSON.stringify(fresh) ||
      JSON.stringify(observed) !== JSON.stringify(request.observed)) throw Error('bracket full source record changed: ' + request.tex);
  const calls = observed.events.filter(e => e.name === '\\' + request.name);
  if (calls.length !== 1 || calls[0].currentCS !== '\\' + request.name) throw Error('invoked command context changed');
}
if (passiveInputs.size !== expected.passiveInputs + expected.bracketPassiveInputs) throw Error('passive input union changed');
fs.writeFileSync(path.join(output, bracketFile), bracketBytes);
reports.push({file: bracketFile, observations: expected.bracketPassiveInputs, additionalPublicInputs: 0,
  inputSHA256: sha(bracketBytes), regeneratedSHA256: sha(bracketBytes), byteIdentical: true});

const literalFile = 'paired_tag_priority_literal_bindings.json.gz';
const literalBytes = fs.readFileSync(path.join(__dirname, literalFile));
const literals = JSON.parse(zlib.gunzipSync(literalBytes)), literalInputs = new Set(), bindingKeys = new Set();
if (literals.cases.length !== expected.literalBindings) throw Error('literal binding count changed');
for (const c of literals.cases) {
  const targetKey = JSON.stringify([c.targetTex, c.display]), literalKey = JSON.stringify([c.literalTex, c.display]);
  const bindingKey = JSON.stringify([c.targetTex, c.literalTex, c.display]);
  if (bindingKeys.has(bindingKey)) throw Error('duplicate literal binding');
  bindingKeys.add(bindingKey); literalInputs.add(literalKey);
  const target = freshByInput.get(targetKey), literal = freshByInput.get(literalKey);
  if (!target || !literal || !c.fullObjectEqual || JSON.stringify(target) !== JSON.stringify(c.originalTarget) ||
      JSON.stringify(literal) !== JSON.stringify(c.originalLiteral) ||
      JSON.stringify(target) !== JSON.stringify(literal)) throw Error('full source literal binding changed');
}
if (literalInputs.size !== expected.literalInputs) throw Error('literal input union changed');
fs.writeFileSync(path.join(output, literalFile), literalBytes);
reports.push({file: literalFile, bindings: bindingKeys.size, distinctLiteralInputsAlreadyPublic: literalInputs.size,
  additionalPublicInputs: 0, inputSHA256: sha(literalBytes), regeneratedSHA256: sha(literalBytes), byteIdentical: true});
fs.writeFileSync(path.join(output, 'paired_tag_priority_regeneration.json.gz'), zlib.gzipSync(JSON.stringify({
  mathjaxGitCommit: fixture.mathjaxGitCommit, files: reports, uniquePublicInputs: expected.inputs,
  distinctPassiveInputsAlreadyInPublicFixture: passiveInputs.size, selectedMapEvents,
  literalBindings: bindingKeys.size, distinctLiteralInputsAlreadyInPublicFixture: literalInputs.size,
  source: 'frozen original only', runtimePolicy: 'historical complete errors preserved; fresh full observations below', runtimeObservations
}, null, 2) + '\n', {level: 9}));
console.log(JSON.stringify(reports, null, 2));
if (reports.some(r => !r.byteIdentical)) throw Error('original fixture bytes changed');
