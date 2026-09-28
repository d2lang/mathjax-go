// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_middle_scope.cjs PINNED_ASSETS');
const hashes = {
  'polyfills.js': '7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01',
  'mathjax.js': 'cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869',
  'setup.js': 'a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881',
};
const scripts = Object.entries(hashes).map(([file, hash]) => {
  const bytes = fs.readFileSync(path.join(assets, file));
  if (crypto.createHash('sha256').update(bytes).digest('hex') !== hash) throw new Error(`unverified ${file}`);
  return new vm.Script(bytes.toString(), {filename: file});
});
const inputs = [
  [
    "stack--plain",
    "\\left(x\\middle|y\\right)q"
  ],
  [
    "stack--empty",
    "\\left(x\\middle|\\right)q"
  ],
  [
    "stack--new-color",
    "\\left(x\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack--new-font",
    "\\left(x\\middle|\\rm y\\right)q"
  ],
  [
    "stack--infix",
    "\\left(x\\middle|y\\over z\\right)q"
  ],
  [
    "stack--nested",
    "\\left(x\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack--middle",
    "\\left(x\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack--color-middle",
    "\\left(x\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-color-plain",
    "\\left(\\color{red}x\\middle|y\\right)q"
  ],
  [
    "stack-color-empty",
    "\\left(\\color{red}x\\middle|\\right)q"
  ],
  [
    "stack-color-new-color",
    "\\left(\\color{red}x\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-color-new-font",
    "\\left(\\color{red}x\\middle|\\rm y\\right)q"
  ],
  [
    "stack-color-infix",
    "\\left(\\color{red}x\\middle|y\\over z\\right)q"
  ],
  [
    "stack-color-nested",
    "\\left(\\color{red}x\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-color-middle",
    "\\left(\\color{red}x\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-color-color-middle",
    "\\left(\\color{red}x\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-color-empty-plain",
    "\\left(\\color{red}\\middle|y\\right)q"
  ],
  [
    "stack-color-empty-empty",
    "\\left(\\color{red}\\middle|\\right)q"
  ],
  [
    "stack-color-empty-new-color",
    "\\left(\\color{red}\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-color-empty-new-font",
    "\\left(\\color{red}\\middle|\\rm y\\right)q"
  ],
  [
    "stack-color-empty-infix",
    "\\left(\\color{red}\\middle|y\\over z\\right)q"
  ],
  [
    "stack-color-empty-nested",
    "\\left(\\color{red}\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-color-empty-middle",
    "\\left(\\color{red}\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-color-empty-color-middle",
    "\\left(\\color{red}\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-double-color-plain",
    "\\left(\\color{red}x\\color{blue}y\\middle|y\\right)q"
  ],
  [
    "stack-double-color-empty",
    "\\left(\\color{red}x\\color{blue}y\\middle|\\right)q"
  ],
  [
    "stack-double-color-new-color",
    "\\left(\\color{red}x\\color{blue}y\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-double-color-new-font",
    "\\left(\\color{red}x\\color{blue}y\\middle|\\rm y\\right)q"
  ],
  [
    "stack-double-color-infix",
    "\\left(\\color{red}x\\color{blue}y\\middle|y\\over z\\right)q"
  ],
  [
    "stack-double-color-nested",
    "\\left(\\color{red}x\\color{blue}y\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-double-color-middle",
    "\\left(\\color{red}x\\color{blue}y\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-double-color-color-middle",
    "\\left(\\color{red}x\\color{blue}y\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-size-plain",
    "\\left(\\Huge x\\middle|y\\right)q"
  ],
  [
    "stack-size-empty",
    "\\left(\\Huge x\\middle|\\right)q"
  ],
  [
    "stack-size-new-color",
    "\\left(\\Huge x\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-size-new-font",
    "\\left(\\Huge x\\middle|\\rm y\\right)q"
  ],
  [
    "stack-size-infix",
    "\\left(\\Huge x\\middle|y\\over z\\right)q"
  ],
  [
    "stack-size-nested",
    "\\left(\\Huge x\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-size-middle",
    "\\left(\\Huge x\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-size-color-middle",
    "\\left(\\Huge x\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-style-plain",
    "\\left(\\scriptstyle x\\middle|y\\right)q"
  ],
  [
    "stack-style-empty",
    "\\left(\\scriptstyle x\\middle|\\right)q"
  ],
  [
    "stack-style-new-color",
    "\\left(\\scriptstyle x\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-style-new-font",
    "\\left(\\scriptstyle x\\middle|\\rm y\\right)q"
  ],
  [
    "stack-style-infix",
    "\\left(\\scriptstyle x\\middle|y\\over z\\right)q"
  ],
  [
    "stack-style-nested",
    "\\left(\\scriptstyle x\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-style-middle",
    "\\left(\\scriptstyle x\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-style-color-middle",
    "\\left(\\scriptstyle x\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-font-plain",
    "\\left(\\bf x\\middle|y\\right)q"
  ],
  [
    "stack-font-empty",
    "\\left(\\bf x\\middle|\\right)q"
  ],
  [
    "stack-font-new-color",
    "\\left(\\bf x\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-font-new-font",
    "\\left(\\bf x\\middle|\\rm y\\right)q"
  ],
  [
    "stack-font-infix",
    "\\left(\\bf x\\middle|y\\over z\\right)q"
  ],
  [
    "stack-font-nested",
    "\\left(\\bf x\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-font-middle",
    "\\left(\\bf x\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-font-color-middle",
    "\\left(\\bf x\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-color-font-plain",
    "\\left(\\color{red}\\bf x\\middle|y\\right)q"
  ],
  [
    "stack-color-font-empty",
    "\\left(\\color{red}\\bf x\\middle|\\right)q"
  ],
  [
    "stack-color-font-new-color",
    "\\left(\\color{red}\\bf x\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-color-font-new-font",
    "\\left(\\color{red}\\bf x\\middle|\\rm y\\right)q"
  ],
  [
    "stack-color-font-infix",
    "\\left(\\color{red}\\bf x\\middle|y\\over z\\right)q"
  ],
  [
    "stack-color-font-nested",
    "\\left(\\color{red}\\bf x\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-color-font-middle",
    "\\left(\\color{red}\\bf x\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-color-font-color-middle",
    "\\left(\\color{red}\\bf x\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-font-color-plain",
    "\\left(\\bf\\color{red}x\\middle|y\\right)q"
  ],
  [
    "stack-font-color-empty",
    "\\left(\\bf\\color{red}x\\middle|\\right)q"
  ],
  [
    "stack-font-color-new-color",
    "\\left(\\bf\\color{red}x\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-font-color-new-font",
    "\\left(\\bf\\color{red}x\\middle|\\rm y\\right)q"
  ],
  [
    "stack-font-color-infix",
    "\\left(\\bf\\color{red}x\\middle|y\\over z\\right)q"
  ],
  [
    "stack-font-color-nested",
    "\\left(\\bf\\color{red}x\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-font-color-middle-2",
    "\\left(\\bf\\color{red}x\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-font-color-color-middle",
    "\\left(\\bf\\color{red}x\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-nested-color-plain",
    "\\left({\\color{red}x}\\middle|y\\right)q"
  ],
  [
    "stack-nested-color-empty",
    "\\left({\\color{red}x}\\middle|\\right)q"
  ],
  [
    "stack-nested-color-new-color",
    "\\left({\\color{red}x}\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-nested-color-new-font",
    "\\left({\\color{red}x}\\middle|\\rm y\\right)q"
  ],
  [
    "stack-nested-color-infix",
    "\\left({\\color{red}x}\\middle|y\\over z\\right)q"
  ],
  [
    "stack-nested-color-nested",
    "\\left({\\color{red}x}\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-nested-color-middle",
    "\\left({\\color{red}x}\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-nested-color-color-middle",
    "\\left({\\color{red}x}\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-prime-plain",
    "\\left(x'\\middle|y\\right)q"
  ],
  [
    "stack-prime-empty",
    "\\left(x'\\middle|\\right)q"
  ],
  [
    "stack-prime-new-color",
    "\\left(x'\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-prime-new-font",
    "\\left(x'\\middle|\\rm y\\right)q"
  ],
  [
    "stack-prime-infix",
    "\\left(x'\\middle|y\\over z\\right)q"
  ],
  [
    "stack-prime-nested",
    "\\left(x'\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-prime-middle",
    "\\left(x'\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-prime-color-middle",
    "\\left(x'\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-fn-plain",
    "\\left(\\sin\\middle|y\\right)q"
  ],
  [
    "stack-fn-empty",
    "\\left(\\sin\\middle|\\right)q"
  ],
  [
    "stack-fn-new-color",
    "\\left(\\sin\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-fn-new-font",
    "\\left(\\sin\\middle|\\rm y\\right)q"
  ],
  [
    "stack-fn-infix",
    "\\left(\\sin\\middle|y\\over z\\right)q"
  ],
  [
    "stack-fn-nested",
    "\\left(\\sin\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-fn-middle",
    "\\left(\\sin\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-fn-color-middle",
    "\\left(\\sin\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-not-plain",
    "\\left(\\not\\middle|y\\right)q"
  ],
  [
    "stack-not-empty",
    "\\left(\\not\\middle|\\right)q"
  ],
  [
    "stack-not-new-color",
    "\\left(\\not\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-not-new-font",
    "\\left(\\not\\middle|\\rm y\\right)q"
  ],
  [
    "stack-not-infix",
    "\\left(\\not\\middle|y\\over z\\right)q"
  ],
  [
    "stack-not-nested",
    "\\left(\\not\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-not-middle",
    "\\left(\\not\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-not-color-middle",
    "\\left(\\not\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-dots-plain",
    "\\left(\\dots\\middle|y\\right)q"
  ],
  [
    "stack-dots-empty",
    "\\left(\\dots\\middle|\\right)q"
  ],
  [
    "stack-dots-new-color",
    "\\left(\\dots\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-dots-new-font",
    "\\left(\\dots\\middle|\\rm y\\right)q"
  ],
  [
    "stack-dots-infix",
    "\\left(\\dots\\middle|y\\over z\\right)q"
  ],
  [
    "stack-dots-nested",
    "\\left(\\dots\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-dots-middle",
    "\\left(\\dots\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-dots-color-middle",
    "\\left(\\dots\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-position-plain",
    "\\left(\\raise1em\\middle|y\\right)q"
  ],
  [
    "stack-position-empty",
    "\\left(\\raise1em\\middle|\\right)q"
  ],
  [
    "stack-position-new-color",
    "\\left(\\raise1em\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-position-new-font",
    "\\left(\\raise1em\\middle|\\rm y\\right)q"
  ],
  [
    "stack-position-infix",
    "\\left(\\raise1em\\middle|y\\over z\\right)q"
  ],
  [
    "stack-position-nested",
    "\\left(\\raise1em\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-position-middle",
    "\\left(\\raise1em\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-position-color-middle",
    "\\left(\\raise1em\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-position-style-plain",
    "\\left(\\raise1em\\color{red}x\\middle|y\\right)q"
  ],
  [
    "stack-position-style-empty",
    "\\left(\\raise1em\\color{red}x\\middle|\\right)q"
  ],
  [
    "stack-position-style-new-color",
    "\\left(\\raise1em\\color{red}x\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-position-style-new-font",
    "\\left(\\raise1em\\color{red}x\\middle|\\rm y\\right)q"
  ],
  [
    "stack-position-style-infix",
    "\\left(\\raise1em\\color{red}x\\middle|y\\over z\\right)q"
  ],
  [
    "stack-position-style-nested",
    "\\left(\\raise1em\\color{red}x\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-position-style-middle",
    "\\left(\\raise1em\\color{red}x\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-position-style-color-middle",
    "\\left(\\raise1em\\color{red}x\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-style-position-plain",
    "\\left(\\color{red}\\raise1em\\middle|y\\right)q"
  ],
  [
    "stack-style-position-empty",
    "\\left(\\color{red}\\raise1em\\middle|\\right)q"
  ],
  [
    "stack-style-position-new-color",
    "\\left(\\color{red}\\raise1em\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-style-position-new-font",
    "\\left(\\color{red}\\raise1em\\middle|\\rm y\\right)q"
  ],
  [
    "stack-style-position-infix",
    "\\left(\\color{red}\\raise1em\\middle|y\\over z\\right)q"
  ],
  [
    "stack-style-position-nested",
    "\\left(\\color{red}\\raise1em\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-style-position-middle",
    "\\left(\\color{red}\\raise1em\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-style-position-color-middle",
    "\\left(\\color{red}\\raise1em\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-position-empty-style-plain",
    "\\left(\\raise1em\\color{red}\\middle|y\\right)q"
  ],
  [
    "stack-position-empty-style-empty",
    "\\left(\\raise1em\\color{red}\\middle|\\right)q"
  ],
  [
    "stack-position-empty-style-new-color",
    "\\left(\\raise1em\\color{red}\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-position-empty-style-new-font",
    "\\left(\\raise1em\\color{red}\\middle|\\rm y\\right)q"
  ],
  [
    "stack-position-empty-style-infix",
    "\\left(\\raise1em\\color{red}\\middle|y\\over z\\right)q"
  ],
  [
    "stack-position-empty-style-nested",
    "\\left(\\raise1em\\color{red}\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-position-empty-style-middle",
    "\\left(\\raise1em\\color{red}\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-position-empty-style-color-middle",
    "\\left(\\raise1em\\color{red}\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-infix-plain",
    "\\left(a\\over b\\middle|y\\right)q"
  ],
  [
    "stack-infix-empty",
    "\\left(a\\over b\\middle|\\right)q"
  ],
  [
    "stack-infix-new-color",
    "\\left(a\\over b\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-infix-new-font",
    "\\left(a\\over b\\middle|\\rm y\\right)q"
  ],
  [
    "stack-infix-infix",
    "\\left(a\\over b\\middle|y\\over z\\right)q"
  ],
  [
    "stack-infix-nested",
    "\\left(a\\over b\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-infix-middle",
    "\\left(a\\over b\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-infix-color-middle",
    "\\left(a\\over b\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-colored-infix-plain",
    "\\left(\\color{red}a\\over\\color{blue}b\\middle|y\\right)q"
  ],
  [
    "stack-colored-infix-empty",
    "\\left(\\color{red}a\\over\\color{blue}b\\middle|\\right)q"
  ],
  [
    "stack-colored-infix-new-color",
    "\\left(\\color{red}a\\over\\color{blue}b\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-colored-infix-new-font",
    "\\left(\\color{red}a\\over\\color{blue}b\\middle|\\rm y\\right)q"
  ],
  [
    "stack-colored-infix-infix",
    "\\left(\\color{red}a\\over\\color{blue}b\\middle|y\\over z\\right)q"
  ],
  [
    "stack-colored-infix-nested",
    "\\left(\\color{red}a\\over\\color{blue}b\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-colored-infix-middle",
    "\\left(\\color{red}a\\over\\color{blue}b\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-colored-infix-color-middle",
    "\\left(\\color{red}a\\over\\color{blue}b\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-font-infix-plain",
    "\\left(\\bf a\\over b\\middle|y\\right)q"
  ],
  [
    "stack-font-infix-empty",
    "\\left(\\bf a\\over b\\middle|\\right)q"
  ],
  [
    "stack-font-infix-new-color",
    "\\left(\\bf a\\over b\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-font-infix-new-font",
    "\\left(\\bf a\\over b\\middle|\\rm y\\right)q"
  ],
  [
    "stack-font-infix-infix",
    "\\left(\\bf a\\over b\\middle|y\\over z\\right)q"
  ],
  [
    "stack-font-infix-nested",
    "\\left(\\bf a\\over b\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-font-infix-middle",
    "\\left(\\bf a\\over b\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-font-infix-color-middle",
    "\\left(\\bf a\\over b\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "stack-infix-position-style-plain",
    "\\left(a\\over\\raise1em\\color{red}b\\middle|y\\right)q"
  ],
  [
    "stack-infix-position-style-empty",
    "\\left(a\\over\\raise1em\\color{red}b\\middle|\\right)q"
  ],
  [
    "stack-infix-position-style-new-color",
    "\\left(a\\over\\raise1em\\color{red}b\\middle|\\color{blue}y\\right)q"
  ],
  [
    "stack-infix-position-style-new-font",
    "\\left(a\\over\\raise1em\\color{red}b\\middle|\\rm y\\right)q"
  ],
  [
    "stack-infix-position-style-infix",
    "\\left(a\\over\\raise1em\\color{red}b\\middle|y\\over z\\right)q"
  ],
  [
    "stack-infix-position-style-nested",
    "\\left(a\\over\\raise1em\\color{red}b\\middle|\\left[\\color{green}y\\right]\\right)q"
  ],
  [
    "stack-infix-position-style-middle",
    "\\left(a\\over\\raise1em\\color{red}b\\middle|y\\middle\\Vert z\\right)q"
  ],
  [
    "stack-infix-position-style-color-middle",
    "\\left(a\\over\\raise1em\\color{red}b\\middle|\\color{blue}y\\middle.\\color{green}z\\right)q"
  ],
  [
    "plain-outer-color",
    "\\color{green}\\left(x\\middle|y\\right)"
  ],
  [
    "plain-outer-font",
    "\\sf \\left(x\\middle|y\\right)"
  ],
  [
    "plain-outer-color-font",
    "\\color{green}\\sf \\left(x\\middle|y\\right)"
  ],
  [
    "plain-outer-size",
    "\\Huge \\left(x\\middle|y\\right)"
  ],
  [
    "plain-font-argument",
    "\\mathbf{\\left(x\\middle|y\\right)}"
  ],
  [
    "plain-vector",
    "\\vb{\\left(x\\middle|y\\right)}"
  ],
  [
    "plain-operator",
    "\\operatorname{\\left(x\\middle|y\\right)}"
  ],
  [
    "plain-fraction",
    "\\frac{\\left(x\\middle|y\\right)}{z}"
  ],
  [
    "plain-subscript",
    "a_{\\left(x\\middle|y\\right)}"
  ],
  [
    "plain-nested-script",
    "a^{b^{\\left(x\\middle|y\\right)}}"
  ],
  [
    "plain-root",
    "\\sqrt[\\left(x\\middle|y\\right)]{x}"
  ],
  [
    "plain-boxed",
    "\\boxed{\\left(x\\middle|y\\right)}"
  ],
  [
    "plain-matrix",
    "\\begin{matrix}\\left(x\\middle|y\\right)&z\\end{matrix}"
  ],
  [
    "plain-internal",
    "\\text{$\\left(x\\middle|y\\right)$}"
  ],
  [
    "plain-textcolor",
    "\\textcolor{blue}{\\left(x\\middle|y\\right)}"
  ],
  [
    "red-outer-color",
    "\\color{green}\\left(\\color{red}x\\middle|y\\right)"
  ],
  [
    "red-outer-font",
    "\\sf \\left(\\color{red}x\\middle|y\\right)"
  ],
  [
    "red-outer-color-font",
    "\\color{green}\\sf \\left(\\color{red}x\\middle|y\\right)"
  ],
  [
    "red-outer-size",
    "\\Huge \\left(\\color{red}x\\middle|y\\right)"
  ],
  [
    "red-font-argument",
    "\\mathbf{\\left(\\color{red}x\\middle|y\\right)}"
  ],
  [
    "red-vector",
    "\\vb{\\left(\\color{red}x\\middle|y\\right)}"
  ],
  [
    "red-operator",
    "\\operatorname{\\left(\\color{red}x\\middle|y\\right)}"
  ],
  [
    "red-fraction",
    "\\frac{\\left(\\color{red}x\\middle|y\\right)}{z}"
  ],
  [
    "red-subscript",
    "a_{\\left(\\color{red}x\\middle|y\\right)}"
  ],
  [
    "red-nested-script",
    "a^{b^{\\left(\\color{red}x\\middle|y\\right)}}"
  ],
  [
    "red-root",
    "\\sqrt[\\left(\\color{red}x\\middle|y\\right)]{x}"
  ],
  [
    "red-boxed",
    "\\boxed{\\left(\\color{red}x\\middle|y\\right)}"
  ],
  [
    "red-matrix",
    "\\begin{matrix}\\left(\\color{red}x\\middle|y\\right)&z\\end{matrix}"
  ],
  [
    "red-internal",
    "\\text{$\\left(\\color{red}x\\middle|y\\right)$}"
  ],
  [
    "red-textcolor",
    "\\textcolor{blue}{\\left(\\color{red}x\\middle|y\\right)}"
  ],
  [
    "font-outer-color",
    "\\color{green}\\left(\\bf x\\middle|y\\right)"
  ],
  [
    "font-outer-font",
    "\\sf \\left(\\bf x\\middle|y\\right)"
  ],
  [
    "font-outer-color-font",
    "\\color{green}\\sf \\left(\\bf x\\middle|y\\right)"
  ],
  [
    "font-outer-size",
    "\\Huge \\left(\\bf x\\middle|y\\right)"
  ],
  [
    "font-font-argument",
    "\\mathbf{\\left(\\bf x\\middle|y\\right)}"
  ],
  [
    "font-vector",
    "\\vb{\\left(\\bf x\\middle|y\\right)}"
  ],
  [
    "font-operator",
    "\\operatorname{\\left(\\bf x\\middle|y\\right)}"
  ],
  [
    "font-fraction",
    "\\frac{\\left(\\bf x\\middle|y\\right)}{z}"
  ],
  [
    "font-subscript",
    "a_{\\left(\\bf x\\middle|y\\right)}"
  ],
  [
    "font-nested-script",
    "a^{b^{\\left(\\bf x\\middle|y\\right)}}"
  ],
  [
    "font-root",
    "\\sqrt[\\left(\\bf x\\middle|y\\right)]{x}"
  ],
  [
    "font-boxed",
    "\\boxed{\\left(\\bf x\\middle|y\\right)}"
  ],
  [
    "font-matrix",
    "\\begin{matrix}\\left(\\bf x\\middle|y\\right)&z\\end{matrix}"
  ],
  [
    "font-internal",
    "\\text{$\\left(\\bf x\\middle|y\\right)$}"
  ],
  [
    "font-textcolor",
    "\\textcolor{blue}{\\left(\\bf x\\middle|y\\right)}"
  ],
  [
    "font-color-outer-color",
    "\\color{green}\\left(\\color{red}\\bf x\\middle|y\\right)"
  ],
  [
    "font-color-outer-font",
    "\\sf \\left(\\color{red}\\bf x\\middle|y\\right)"
  ],
  [
    "font-color-outer-color-font",
    "\\color{green}\\sf \\left(\\color{red}\\bf x\\middle|y\\right)"
  ],
  [
    "font-color-outer-size",
    "\\Huge \\left(\\color{red}\\bf x\\middle|y\\right)"
  ],
  [
    "font-color-font-argument",
    "\\mathbf{\\left(\\color{red}\\bf x\\middle|y\\right)}"
  ],
  [
    "font-color-vector",
    "\\vb{\\left(\\color{red}\\bf x\\middle|y\\right)}"
  ],
  [
    "font-color-operator",
    "\\operatorname{\\left(\\color{red}\\bf x\\middle|y\\right)}"
  ],
  [
    "font-color-fraction",
    "\\frac{\\left(\\color{red}\\bf x\\middle|y\\right)}{z}"
  ],
  [
    "font-color-subscript",
    "a_{\\left(\\color{red}\\bf x\\middle|y\\right)}"
  ],
  [
    "font-color-nested-script",
    "a^{b^{\\left(\\color{red}\\bf x\\middle|y\\right)}}"
  ],
  [
    "font-color-root",
    "\\sqrt[\\left(\\color{red}\\bf x\\middle|y\\right)]{x}"
  ],
  [
    "font-color-boxed",
    "\\boxed{\\left(\\color{red}\\bf x\\middle|y\\right)}"
  ],
  [
    "font-color-matrix",
    "\\begin{matrix}\\left(\\color{red}\\bf x\\middle|y\\right)&z\\end{matrix}"
  ],
  [
    "font-color-internal",
    "\\text{$\\left(\\color{red}\\bf x\\middle|y\\right)$}"
  ],
  [
    "font-color-textcolor",
    "\\textcolor{blue}{\\left(\\color{red}\\bf x\\middle|y\\right)}"
  ],
  [
    "nested-outer-color",
    "\\color{green}\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)"
  ],
  [
    "nested-outer-font",
    "\\sf \\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)"
  ],
  [
    "nested-outer-color-font",
    "\\color{green}\\sf \\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)"
  ],
  [
    "nested-outer-size",
    "\\Huge \\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)"
  ],
  [
    "nested-font-argument",
    "\\mathbf{\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)}"
  ],
  [
    "nested-vector",
    "\\vb{\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)}"
  ],
  [
    "nested-operator",
    "\\operatorname{\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)}"
  ],
  [
    "nested-fraction",
    "\\frac{\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)}{z}"
  ],
  [
    "nested-subscript",
    "a_{\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)}"
  ],
  [
    "nested-nested-script",
    "a^{b^{\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)}}"
  ],
  [
    "nested-root",
    "\\sqrt[\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)]{x}"
  ],
  [
    "nested-boxed",
    "\\boxed{\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)}"
  ],
  [
    "nested-matrix",
    "\\begin{matrix}\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)&z\\end{matrix}"
  ],
  [
    "nested-internal",
    "\\text{$\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)$}"
  ],
  [
    "nested-textcolor",
    "\\textcolor{blue}{\\left(\\left[\\color{blue}x\\middle|y\\right]\\middle|z\\right)}"
  ],
  [
    "infix-outer-color",
    "\\color{green}\\left(\\color{red}x\\over y\\middle|z\\over w\\right)"
  ],
  [
    "infix-outer-font",
    "\\sf \\left(\\color{red}x\\over y\\middle|z\\over w\\right)"
  ],
  [
    "infix-outer-color-font",
    "\\color{green}\\sf \\left(\\color{red}x\\over y\\middle|z\\over w\\right)"
  ],
  [
    "infix-outer-size",
    "\\Huge \\left(\\color{red}x\\over y\\middle|z\\over w\\right)"
  ],
  [
    "infix-font-argument",
    "\\mathbf{\\left(\\color{red}x\\over y\\middle|z\\over w\\right)}"
  ],
  [
    "infix-vector",
    "\\vb{\\left(\\color{red}x\\over y\\middle|z\\over w\\right)}"
  ],
  [
    "infix-operator",
    "\\operatorname{\\left(\\color{red}x\\over y\\middle|z\\over w\\right)}"
  ],
  [
    "infix-fraction",
    "\\frac{\\left(\\color{red}x\\over y\\middle|z\\over w\\right)}{z}"
  ],
  [
    "infix-subscript",
    "a_{\\left(\\color{red}x\\over y\\middle|z\\over w\\right)}"
  ],
  [
    "infix-nested-script",
    "a^{b^{\\left(\\color{red}x\\over y\\middle|z\\over w\\right)}}"
  ],
  [
    "infix-root",
    "\\sqrt[\\left(\\color{red}x\\over y\\middle|z\\over w\\right)]{x}"
  ],
  [
    "infix-boxed",
    "\\boxed{\\left(\\color{red}x\\over y\\middle|z\\over w\\right)}"
  ],
  [
    "infix-matrix",
    "\\begin{matrix}\\left(\\color{red}x\\over y\\middle|z\\over w\\right)&z\\end{matrix}"
  ],
  [
    "infix-internal",
    "\\text{$\\left(\\color{red}x\\over y\\middle|z\\over w\\right)$}"
  ],
  [
    "infix-textcolor",
    "\\textcolor{blue}{\\left(\\color{red}x\\over y\\middle|z\\over w\\right)}"
  ],
  [
    "unowned",
    "a\\middle|b"
  ],
  [
    "unowned-colored",
    "\\color{red}a\\middle|b"
  ],
  [
    "unowned-position",
    "\\raise1em\\middle|"
  ],
  [
    "unowned-infix",
    "a\\over b\\middle|c"
  ],
  [
    "unowned-font",
    "\\bf x\\middle|y"
  ],
  [
    "group-boundary",
    "\\left({x\\middle|y}\\right)"
  ],
  [
    "group-no-delim",
    "\\left({x\\middle}\\right)"
  ],
  [
    "script",
    "\\left(x^\\middle|y\\right)"
  ],
  [
    "script-missing-delim",
    "\\left(x^\\middle\\right)"
  ],
  [
    "script-macro",
    "\\DeclareMathOperator{\\middle}{M}\\left(x^\\middle y\\right)"
  ],
  [
    "macro",
    "\\DeclareMathOperator{\\middle}{M}\\left(x\\middle|y\\right)"
  ],
  [
    "right-macro",
    "\\DeclareMathOperator{\\right}{R}\\right(x)"
  ],
  [
    "right-macro-left",
    "\\DeclareMathOperator{\\right}{R}\\left(x\\right)"
  ],
  [
    "missing-delim",
    "\\left(x\\middle"
  ],
  [
    "invalid-delim",
    "\\left(x\\middle a y\\right)"
  ],
  [
    "invalid-command-delim",
    "\\left(x\\middle\\foo y\\right)"
  ],
  [
    "position-invalid-delim",
    "\\left(x\\raise1em\\middle\\foo y\\right)"
  ],
  [
    "extra-close",
    "\\left(x\\middle|y}"
  ],
  [
    "missing-right",
    "\\left(x\\middle|y"
  ],
  [
    "bare-dot",
    "\\left.x\\middle.y\\right."
  ],
  [
    "middle-brace",
    "\\left(x\\middle\\{y\\right)"
  ],
  [
    "middle-relation",
    "\\left(x\\middle\\rightarrow y\\right)"
  ],
  [
    "middle-angle",
    "\\left(x\\middle\\langle y\\right)"
  ],
  [
    "middle-backslash",
    "\\left(x\\middle\\backslash y\\right)"
  ],
  [
    "script-after-middle",
    "\\left(x\\middle|_a y\\right)"
  ],
  [
    "prime-after-middle",
    "\\left(x\\middle|'y\\right)"
  ],
  [
    "limits-after-middle",
    "\\left(x\\middle|\\limits y\\right)"
  ],
  [
    "rule-env",
    "\\color{green}\\left(\\color{red}\\Rule{1em}{1em}{0em}\\middle|\\Rule{1em}{1em}{0em}\\right)"
  ],
  [
    "root-env",
    "\\sqrt[\\left(\\uproot2 n\\middle|\\uproot3 m\\right)]{x}"
  ],
  [
    "multiple-over",
    "\\left(a\\over b\\middle|c\\over d\\middle|e\\over f\\right)"
  ],
  [
    "consecutive-middle",
    "\\left(\\color{red}\\middle|\\middle|\\middle|\\right)"
  ],
  [
    "auto-open-middle",
    "\\dv(x\\middle|y)"
  ],
  [
    "braket-middle",
    "\\Braket{x\\middle|y}"
  ],
  [
    "matrix-middle",
    "\\matrix{x\\middle|y}"
  ],
  [
    "fraction-middle",
    "\\frac{x\\middle|y}{z}"
  ],
  [
    "infix-before-middle-script",
    "\\left(a\\over b\\middle|_x y\\right)"
  ],
  [
    "ordinary-middle-control",
    "\\left\\langle\\frac{x}{y}\\middle|z\\right\\rangle"
  ],
  [
    "matrix-owner-matrix-middle",
    "\\matrix{x\\middle|y}"
  ],
  [
    "matrix-owner-matrix-right",
    "\\matrix{x\\right)}"
  ],
  [
    "matrix-owner-matrix-font",
    "\\matrix{\\bf x\\middle|y}"
  ],
  [
    "matrix-owner-matrix-style",
    "\\matrix{\\color{red}x\\middle|y}"
  ],
  [
    "matrix-owner-matrix-infix",
    "\\matrix{x\\over y\\middle|z}"
  ],
  [
    "matrix-owner-matrix-group",
    "\\matrix{{x\\middle|y}}"
  ],
  [
    "matrix-owner-matrix-group-right",
    "\\matrix{{x\\right)}}"
  ],
  [
    "matrix-owner-matrix-nested-left",
    "\\matrix{\\left(x\\middle|y\\right)}"
  ],
  [
    "matrix-owner-matrix-infix-left",
    "\\matrix{\\left(x\\middle|y\\over z\\right)}"
  ],
  [
    "matrix-owner-matrix-position",
    "\\matrix{\\raise1em\\middle|}"
  ],
  [
    "matrix-owner-matrix-position-style",
    "\\matrix{\\raise1em\\color{red}x\\middle|y}"
  ],
  [
    "matrix-owner-matrix-script",
    "\\matrix{x^\\middle|y}"
  ],
  [
    "matrix-owner-matrix-auto-open",
    "\\matrix{\\dv(x\\middle|y)}"
  ],
  [
    "matrix-owner-matrix-bad-delimiter",
    "\\matrix{x\\middle\\foo y}"
  ],
  [
    "matrix-owner-array-middle",
    "\\array{x\\middle|y}"
  ],
  [
    "matrix-owner-array-right",
    "\\array{x\\right)}"
  ],
  [
    "matrix-owner-array-font",
    "\\array{\\bf x\\middle|y}"
  ],
  [
    "matrix-owner-array-style",
    "\\array{\\color{red}x\\middle|y}"
  ],
  [
    "matrix-owner-array-infix",
    "\\array{x\\over y\\middle|z}"
  ],
  [
    "matrix-owner-array-group",
    "\\array{{x\\middle|y}}"
  ],
  [
    "matrix-owner-array-group-right",
    "\\array{{x\\right)}}"
  ],
  [
    "matrix-owner-array-nested-left",
    "\\array{\\left(x\\middle|y\\right)}"
  ],
  [
    "matrix-owner-array-infix-left",
    "\\array{\\left(x\\middle|y\\over z\\right)}"
  ],
  [
    "matrix-owner-array-position",
    "\\array{\\raise1em\\middle|}"
  ],
  [
    "matrix-owner-array-position-style",
    "\\array{\\raise1em\\color{red}x\\middle|y}"
  ],
  [
    "matrix-owner-array-script",
    "\\array{x^\\middle|y}"
  ],
  [
    "matrix-owner-array-auto-open",
    "\\array{\\dv(x\\middle|y)}"
  ],
  [
    "matrix-owner-array-bad-delimiter",
    "\\array{x\\middle\\foo y}"
  ],
  [
    "matrix-owner-pmatrix-middle",
    "\\pmatrix{x\\middle|y}"
  ],
  [
    "matrix-owner-pmatrix-right",
    "\\pmatrix{x\\right)}"
  ],
  [
    "matrix-owner-pmatrix-font",
    "\\pmatrix{\\bf x\\middle|y}"
  ],
  [
    "matrix-owner-pmatrix-style",
    "\\pmatrix{\\color{red}x\\middle|y}"
  ],
  [
    "matrix-owner-pmatrix-infix",
    "\\pmatrix{x\\over y\\middle|z}"
  ],
  [
    "matrix-owner-pmatrix-group",
    "\\pmatrix{{x\\middle|y}}"
  ],
  [
    "matrix-owner-pmatrix-group-right",
    "\\pmatrix{{x\\right)}}"
  ],
  [
    "matrix-owner-pmatrix-nested-left",
    "\\pmatrix{\\left(x\\middle|y\\right)}"
  ],
  [
    "matrix-owner-pmatrix-infix-left",
    "\\pmatrix{\\left(x\\middle|y\\over z\\right)}"
  ],
  [
    "matrix-owner-pmatrix-position",
    "\\pmatrix{\\raise1em\\middle|}"
  ],
  [
    "matrix-owner-pmatrix-position-style",
    "\\pmatrix{\\raise1em\\color{red}x\\middle|y}"
  ],
  [
    "matrix-owner-pmatrix-script",
    "\\pmatrix{x^\\middle|y}"
  ],
  [
    "matrix-owner-pmatrix-auto-open",
    "\\pmatrix{\\dv(x\\middle|y)}"
  ],
  [
    "matrix-owner-pmatrix-bad-delimiter",
    "\\pmatrix{x\\middle\\foo y}"
  ],
  [
    "matrix-owner-eqalign-middle",
    "\\eqalign{x\\middle|y}"
  ],
  [
    "matrix-owner-eqalign-right",
    "\\eqalign{x\\right)}"
  ],
  [
    "matrix-owner-eqalign-font",
    "\\eqalign{\\bf x\\middle|y}"
  ],
  [
    "matrix-owner-eqalign-style",
    "\\eqalign{\\color{red}x\\middle|y}"
  ],
  [
    "matrix-owner-eqalign-infix",
    "\\eqalign{x\\over y\\middle|z}"
  ],
  [
    "matrix-owner-eqalign-group",
    "\\eqalign{{x\\middle|y}}"
  ],
  [
    "matrix-owner-eqalign-group-right",
    "\\eqalign{{x\\right)}}"
  ],
  [
    "matrix-owner-eqalign-nested-left",
    "\\eqalign{\\left(x\\middle|y\\right)}"
  ],
  [
    "matrix-owner-eqalign-infix-left",
    "\\eqalign{\\left(x\\middle|y\\over z\\right)}"
  ],
  [
    "matrix-owner-eqalign-position",
    "\\eqalign{\\raise1em\\middle|}"
  ],
  [
    "matrix-owner-eqalign-position-style",
    "\\eqalign{\\raise1em\\color{red}x\\middle|y}"
  ],
  [
    "matrix-owner-eqalign-script",
    "\\eqalign{x^\\middle|y}"
  ],
  [
    "matrix-owner-eqalign-auto-open",
    "\\eqalign{\\dv(x\\middle|y)}"
  ],
  [
    "matrix-owner-eqalign-bad-delimiter",
    "\\eqalign{x\\middle\\foo y}"
  ]
];
const cases = [];
for (const [name, tex] of inputs) {
  for (const display of [false, true]) {
    const context = vm.createContext({console});
    context.globalThis = context;
    for (const script of scripts) script.runInContext(context);
    const svg = context.adaptor.innerHTML(context.html.convert(tex, {em: 16, ex: 8, display}));
    const width = Math.ceil(Number(svg.match(/ width="([\d.]+)ex"/)[1]) * 8);
    const height = Math.ceil(Number(svg.match(/ height="([\d.]+)ex"/)[1]) * 8);
    cases.push({name: `${name}-${display ? 'display' : 'inline'}`, tex, display, svg, width, height});
  }
}
fs.writeFileSync(path.join(__dirname, 'middle_scope_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n');
