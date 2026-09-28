// SPDX-License-Identifier: Apache-2.0
// Capture unmodified D2 MathJax 3.2.2 output in a fresh runtime per expression.
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const crypto = require('node:crypto');
const assets = process.argv[2];
if (!assets) throw new Error('usage: node generate_horizontal_position.cjs PINNED_ASSETS');
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
    "moveleft-positive",
    "a\\moveleft1em{X}b"
  ],
  [
    "moveleft-negative",
    "a\\moveleft-1em{X}b"
  ],
  [
    "moveleft-plus",
    "a\\moveleft+1em{X}b"
  ],
  [
    "moveleft-zero",
    "a\\moveleft0pt{X}b"
  ],
  [
    "moveleft-negative-zero",
    "a\\moveleft-0em{X}b"
  ],
  [
    "moveleft-comma",
    "a\\moveleft,5em{X}b"
  ],
  [
    "moveleft-signed-comma",
    "a\\moveleft-,5em{X}b"
  ],
  [
    "moveleft-mu",
    "a\\moveleft3mu{X}b"
  ],
  [
    "moveleft-physical",
    "a\\moveleft2mm{X}b"
  ],
  [
    "moveleft-braced",
    "a\\moveleft{1.2ex}{X}b"
  ],
  [
    "moveleft-spaced",
    "a\\moveleft 1 em {X}b"
  ],
  [
    "moveleft-negative-braced",
    "a\\moveleft{-2mu}{X}b"
  ],
  [
    "moveleft-operand-char",
    "a\\moveleft1em X"
  ],
  [
    "moveleft-operand-group",
    "a\\moveleft1em {XY}"
  ],
  [
    "moveleft-operand-empty-group",
    "a\\moveleft1em {}"
  ],
  [
    "moveleft-operand-fraction",
    "a\\moveleft1em \\frac{a}{b}"
  ],
  [
    "moveleft-operand-sqrt",
    "a\\moveleft1em \\sqrt{x}"
  ],
  [
    "moveleft-operand-left",
    "a\\moveleft1em \\left(x\\middle|y\\right)"
  ],
  [
    "moveleft-operand-font",
    "a\\moveleft1em \\bf XY"
  ],
  [
    "moveleft-operand-style",
    "a\\moveleft1em \\scriptstyle XY"
  ],
  [
    "moveleft-operand-size",
    "a\\moveleft1em \\Large XY"
  ],
  [
    "moveleft-operand-color",
    "a\\moveleft1em \\color{red}XY"
  ],
  [
    "moveleft-operand-textcolor",
    "a\\moveleft1em \\textcolor{red}{X}Y"
  ],
  [
    "moveleft-operand-text",
    "a\\moveleft1em \\text{abc}"
  ],
  [
    "moveleft-operand-fn",
    "a\\moveleft1em \\sin x"
  ],
  [
    "moveleft-operand-fn-script",
    "a\\moveleft1em \\sin^2 x"
  ],
  [
    "moveleft-operand-dots",
    "a\\moveleft1em \\dots x"
  ],
  [
    "moveleft-operand-not",
    "a\\moveleft1em \\not= x"
  ],
  [
    "moveleft-operand-prime",
    "a\\moveleft1em x' y"
  ],
  [
    "moveleft-operand-phantom",
    "a\\moveleft1em \\phantom{X}"
  ],
  [
    "moveleft-operand-multi-result",
    "a\\moveleft1em \\mod{x}"
  ],
  [
    "moveleft-operand-differential",
    "a\\moveleft1em \\dd{x}"
  ],
  [
    "moveleft-operand-autoopen",
    "a\\moveleft1em \\dv(x)"
  ],
  [
    "moveleft-operand-braket",
    "a\\moveleft1em \\Braket{x|y}"
  ],
  [
    "moveleft-operand-rule",
    "a\\moveleft1em \\rule[-.2em]{1em}{.3em}"
  ],
  [
    "moveleft-operand-Space",
    "a\\moveleft1em \\Space{1em}{.2em}{.1em}"
  ],
  [
    "moveleft-operand-matrix",
    "a\\moveleft1em \\matrix{a&b\\\\c&d}"
  ],
  [
    "moveleft-operand-superscript",
    "a\\moveleft1em X^2"
  ],
  [
    "moveleft-operand-subscript",
    "a\\moveleft1em X_i"
  ],
  [
    "moveleft-operand-missing",
    "a\\moveleft1em "
  ],
  [
    "moveleft-operand-missing-after-relax",
    "a\\moveleft1em \\relax"
  ],
  [
    "moveleft-operand-missing-after-font",
    "a\\moveleft1em \\bf"
  ],
  [
    "moveleft-operand-empty-style",
    "a\\moveleft1em \\color{red}"
  ],
  [
    "moveleft-close-right",
    "\\left(x\\moveleft1em \\right)"
  ],
  [
    "moveleft-close-middle",
    "\\left(x\\moveleft1em \\middle|y\\right)"
  ],
  [
    "moveleft-close-over",
    "\\left(x\\moveleft1em \\over y\\right)"
  ],
  [
    "moveleft-close-close",
    "\\left(x\\moveleft1em }"
  ],
  [
    "moveleft-close-linebreak",
    "\\left(x\\moveleft1em \\\\"
  ],
  [
    "moveleft-close-linebreak-bad",
    "\\left(x\\moveleft1em \\\\[bad]"
  ],
  [
    "moveleft-close-alignment",
    "\\left(x\\moveleft1em &y"
  ],
  [
    "moveleft-close-bad-delimiter",
    "\\left(x\\moveleft1em \\right\\foo"
  ],
  [
    "moveleft-dimension-bad",
    "\\moveleftbad X"
  ],
  [
    "moveleft-dimension-unitless",
    "\\moveleft{2}X"
  ],
  [
    "moveleft-dimension-missing",
    "\\moveleft"
  ],
  [
    "moveleft-dimension-unclosed",
    "\\moveleft{2em"
  ],
  [
    "moveleft-dimension-bad-braced",
    "\\moveleft{bad}X"
  ],
  [
    "moveleft-context-script",
    "x^\\moveleft1em X"
  ],
  [
    "moveleft-context-braced-script",
    "x^{\\moveleft1em X}"
  ],
  [
    "moveleft-context-numerator",
    "\\frac{\\moveleft1em X}{y}"
  ],
  [
    "moveleft-context-color-outside",
    "\\color{red}\\moveleft1em X"
  ],
  [
    "moveleft-context-font-outside",
    "\\bf\\moveleft1em XY"
  ],
  [
    "moveleft-context-subgroup",
    "{\\moveleft1em X}Y"
  ],
  [
    "moveleft-context-middle-group",
    "\\left(\\moveleft1em\\color{red}X\\middle|Y\\right)"
  ],
  [
    "moveleft-context-override",
    "\\DeclareMathOperator{\\moveleft}{M}\\moveleft1em X"
  ],
  [
    "moveright-positive",
    "a\\moveright1em{X}b"
  ],
  [
    "moveright-negative",
    "a\\moveright-1em{X}b"
  ],
  [
    "moveright-plus",
    "a\\moveright+1em{X}b"
  ],
  [
    "moveright-zero",
    "a\\moveright0pt{X}b"
  ],
  [
    "moveright-negative-zero",
    "a\\moveright-0em{X}b"
  ],
  [
    "moveright-comma",
    "a\\moveright,5em{X}b"
  ],
  [
    "moveright-signed-comma",
    "a\\moveright-,5em{X}b"
  ],
  [
    "moveright-mu",
    "a\\moveright3mu{X}b"
  ],
  [
    "moveright-physical",
    "a\\moveright2mm{X}b"
  ],
  [
    "moveright-braced",
    "a\\moveright{1.2ex}{X}b"
  ],
  [
    "moveright-spaced",
    "a\\moveright 1 em {X}b"
  ],
  [
    "moveright-negative-braced",
    "a\\moveright{-2mu}{X}b"
  ],
  [
    "moveright-operand-char",
    "a\\moveright1em X"
  ],
  [
    "moveright-operand-group",
    "a\\moveright1em {XY}"
  ],
  [
    "moveright-operand-empty-group",
    "a\\moveright1em {}"
  ],
  [
    "moveright-operand-fraction",
    "a\\moveright1em \\frac{a}{b}"
  ],
  [
    "moveright-operand-sqrt",
    "a\\moveright1em \\sqrt{x}"
  ],
  [
    "moveright-operand-left",
    "a\\moveright1em \\left(x\\middle|y\\right)"
  ],
  [
    "moveright-operand-font",
    "a\\moveright1em \\bf XY"
  ],
  [
    "moveright-operand-style",
    "a\\moveright1em \\scriptstyle XY"
  ],
  [
    "moveright-operand-size",
    "a\\moveright1em \\Large XY"
  ],
  [
    "moveright-operand-color",
    "a\\moveright1em \\color{red}XY"
  ],
  [
    "moveright-operand-textcolor",
    "a\\moveright1em \\textcolor{red}{X}Y"
  ],
  [
    "moveright-operand-text",
    "a\\moveright1em \\text{abc}"
  ],
  [
    "moveright-operand-fn",
    "a\\moveright1em \\sin x"
  ],
  [
    "moveright-operand-fn-script",
    "a\\moveright1em \\sin^2 x"
  ],
  [
    "moveright-operand-dots",
    "a\\moveright1em \\dots x"
  ],
  [
    "moveright-operand-not",
    "a\\moveright1em \\not= x"
  ],
  [
    "moveright-operand-prime",
    "a\\moveright1em x' y"
  ],
  [
    "moveright-operand-phantom",
    "a\\moveright1em \\phantom{X}"
  ],
  [
    "moveright-operand-multi-result",
    "a\\moveright1em \\mod{x}"
  ],
  [
    "moveright-operand-differential",
    "a\\moveright1em \\dd{x}"
  ],
  [
    "moveright-operand-autoopen",
    "a\\moveright1em \\dv(x)"
  ],
  [
    "moveright-operand-braket",
    "a\\moveright1em \\Braket{x|y}"
  ],
  [
    "moveright-operand-rule",
    "a\\moveright1em \\rule[-.2em]{1em}{.3em}"
  ],
  [
    "moveright-operand-Space",
    "a\\moveright1em \\Space{1em}{.2em}{.1em}"
  ],
  [
    "moveright-operand-matrix",
    "a\\moveright1em \\matrix{a&b\\\\c&d}"
  ],
  [
    "moveright-operand-superscript",
    "a\\moveright1em X^2"
  ],
  [
    "moveright-operand-subscript",
    "a\\moveright1em X_i"
  ],
  [
    "moveright-operand-missing",
    "a\\moveright1em "
  ],
  [
    "moveright-operand-missing-after-relax",
    "a\\moveright1em \\relax"
  ],
  [
    "moveright-operand-missing-after-font",
    "a\\moveright1em \\bf"
  ],
  [
    "moveright-operand-empty-style",
    "a\\moveright1em \\color{red}"
  ],
  [
    "moveright-close-right",
    "\\left(x\\moveright1em \\right)"
  ],
  [
    "moveright-close-middle",
    "\\left(x\\moveright1em \\middle|y\\right)"
  ],
  [
    "moveright-close-over",
    "\\left(x\\moveright1em \\over y\\right)"
  ],
  [
    "moveright-close-close",
    "\\left(x\\moveright1em }"
  ],
  [
    "moveright-close-linebreak",
    "\\left(x\\moveright1em \\\\"
  ],
  [
    "moveright-close-linebreak-bad",
    "\\left(x\\moveright1em \\\\[bad]"
  ],
  [
    "moveright-close-alignment",
    "\\left(x\\moveright1em &y"
  ],
  [
    "moveright-close-bad-delimiter",
    "\\left(x\\moveright1em \\right\\foo"
  ],
  [
    "moveright-dimension-bad",
    "\\moverightbad X"
  ],
  [
    "moveright-dimension-unitless",
    "\\moveright{2}X"
  ],
  [
    "moveright-dimension-missing",
    "\\moveright"
  ],
  [
    "moveright-dimension-unclosed",
    "\\moveright{2em"
  ],
  [
    "moveright-dimension-bad-braced",
    "\\moveright{bad}X"
  ],
  [
    "moveright-context-script",
    "x^\\moveright1em X"
  ],
  [
    "moveright-context-braced-script",
    "x^{\\moveright1em X}"
  ],
  [
    "moveright-context-numerator",
    "\\frac{\\moveright1em X}{y}"
  ],
  [
    "moveright-context-color-outside",
    "\\color{red}\\moveright1em X"
  ],
  [
    "moveright-context-font-outside",
    "\\bf\\moveright1em XY"
  ],
  [
    "moveright-context-subgroup",
    "{\\moveright1em X}Y"
  ],
  [
    "moveright-context-middle-group",
    "\\left(\\moveright1em\\color{red}X\\middle|Y\\right)"
  ],
  [
    "moveright-context-override",
    "\\DeclareMathOperator{\\moveright}{M}\\moveright1em X"
  ],
  [
    "pair-l-l-char",
    "a\\moveleft1em\\moveleft1em Xb"
  ],
  [
    "pair-l-l-group",
    "a\\moveleft1em\\moveleft1em {XY}b"
  ],
  [
    "pair-l-l-style",
    "a\\moveleft1em\\moveleft1em \\color{red}XYb"
  ],
  [
    "pair-l-l-fn",
    "a\\moveleft1em\\moveleft1em \\sin xb"
  ],
  [
    "pair-l-l-script",
    "a\\moveleft1em\\moveleft1em X^2b"
  ],
  [
    "triple-lll",
    "a\\moveleft1em\\moveleft1em\\moveleft1em Xb"
  ],
  [
    "triple-llr",
    "a\\moveleft1em\\moveleft1em\\moveright.5em Xb"
  ],
  [
    "triple-llup",
    "a\\moveleft1em\\moveleft1em\\raise.7em Xb"
  ],
  [
    "triple-lldown",
    "a\\moveleft1em\\moveleft1em\\lower.3em Xb"
  ],
  [
    "pair-l-r-char",
    "a\\moveleft1em\\moveright.5em Xb"
  ],
  [
    "pair-l-r-group",
    "a\\moveleft1em\\moveright.5em {XY}b"
  ],
  [
    "pair-l-r-style",
    "a\\moveleft1em\\moveright.5em \\color{red}XYb"
  ],
  [
    "pair-l-r-fn",
    "a\\moveleft1em\\moveright.5em \\sin xb"
  ],
  [
    "pair-l-r-script",
    "a\\moveleft1em\\moveright.5em X^2b"
  ],
  [
    "triple-lrl",
    "a\\moveleft1em\\moveright.5em\\moveleft1em Xb"
  ],
  [
    "triple-lrr",
    "a\\moveleft1em\\moveright.5em\\moveright.5em Xb"
  ],
  [
    "triple-lrup",
    "a\\moveleft1em\\moveright.5em\\raise.7em Xb"
  ],
  [
    "triple-lrdown",
    "a\\moveleft1em\\moveright.5em\\lower.3em Xb"
  ],
  [
    "pair-l-up-char",
    "a\\moveleft1em\\raise.7em Xb"
  ],
  [
    "pair-l-up-group",
    "a\\moveleft1em\\raise.7em {XY}b"
  ],
  [
    "pair-l-up-style",
    "a\\moveleft1em\\raise.7em \\color{red}XYb"
  ],
  [
    "pair-l-up-fn",
    "a\\moveleft1em\\raise.7em \\sin xb"
  ],
  [
    "pair-l-up-script",
    "a\\moveleft1em\\raise.7em X^2b"
  ],
  [
    "triple-lupl",
    "a\\moveleft1em\\raise.7em\\moveleft1em Xb"
  ],
  [
    "triple-lupr",
    "a\\moveleft1em\\raise.7em\\moveright.5em Xb"
  ],
  [
    "triple-lupup",
    "a\\moveleft1em\\raise.7em\\raise.7em Xb"
  ],
  [
    "triple-lupdown",
    "a\\moveleft1em\\raise.7em\\lower.3em Xb"
  ],
  [
    "pair-l-down-char",
    "a\\moveleft1em\\lower.3em Xb"
  ],
  [
    "pair-l-down-group",
    "a\\moveleft1em\\lower.3em {XY}b"
  ],
  [
    "pair-l-down-style",
    "a\\moveleft1em\\lower.3em \\color{red}XYb"
  ],
  [
    "pair-l-down-fn",
    "a\\moveleft1em\\lower.3em \\sin xb"
  ],
  [
    "pair-l-down-script",
    "a\\moveleft1em\\lower.3em X^2b"
  ],
  [
    "triple-ldownl",
    "a\\moveleft1em\\lower.3em\\moveleft1em Xb"
  ],
  [
    "triple-ldownr",
    "a\\moveleft1em\\lower.3em\\moveright.5em Xb"
  ],
  [
    "triple-ldownup",
    "a\\moveleft1em\\lower.3em\\raise.7em Xb"
  ],
  [
    "triple-ldowndown",
    "a\\moveleft1em\\lower.3em\\lower.3em Xb"
  ],
  [
    "pair-r-l-char",
    "a\\moveright.5em\\moveleft1em Xb"
  ],
  [
    "pair-r-l-group",
    "a\\moveright.5em\\moveleft1em {XY}b"
  ],
  [
    "pair-r-l-style",
    "a\\moveright.5em\\moveleft1em \\color{red}XYb"
  ],
  [
    "pair-r-l-fn",
    "a\\moveright.5em\\moveleft1em \\sin xb"
  ],
  [
    "pair-r-l-script",
    "a\\moveright.5em\\moveleft1em X^2b"
  ],
  [
    "triple-rll",
    "a\\moveright.5em\\moveleft1em\\moveleft1em Xb"
  ],
  [
    "triple-rlr",
    "a\\moveright.5em\\moveleft1em\\moveright.5em Xb"
  ],
  [
    "triple-rlup",
    "a\\moveright.5em\\moveleft1em\\raise.7em Xb"
  ],
  [
    "triple-rldown",
    "a\\moveright.5em\\moveleft1em\\lower.3em Xb"
  ],
  [
    "pair-r-r-char",
    "a\\moveright.5em\\moveright.5em Xb"
  ],
  [
    "pair-r-r-group",
    "a\\moveright.5em\\moveright.5em {XY}b"
  ],
  [
    "pair-r-r-style",
    "a\\moveright.5em\\moveright.5em \\color{red}XYb"
  ],
  [
    "pair-r-r-fn",
    "a\\moveright.5em\\moveright.5em \\sin xb"
  ],
  [
    "pair-r-r-script",
    "a\\moveright.5em\\moveright.5em X^2b"
  ],
  [
    "triple-rrl",
    "a\\moveright.5em\\moveright.5em\\moveleft1em Xb"
  ],
  [
    "triple-rrr",
    "a\\moveright.5em\\moveright.5em\\moveright.5em Xb"
  ],
  [
    "triple-rrup",
    "a\\moveright.5em\\moveright.5em\\raise.7em Xb"
  ],
  [
    "triple-rrdown",
    "a\\moveright.5em\\moveright.5em\\lower.3em Xb"
  ],
  [
    "pair-r-up-char",
    "a\\moveright.5em\\raise.7em Xb"
  ],
  [
    "pair-r-up-group",
    "a\\moveright.5em\\raise.7em {XY}b"
  ],
  [
    "pair-r-up-style",
    "a\\moveright.5em\\raise.7em \\color{red}XYb"
  ],
  [
    "pair-r-up-fn",
    "a\\moveright.5em\\raise.7em \\sin xb"
  ],
  [
    "pair-r-up-script",
    "a\\moveright.5em\\raise.7em X^2b"
  ],
  [
    "triple-rupl",
    "a\\moveright.5em\\raise.7em\\moveleft1em Xb"
  ],
  [
    "triple-rupr",
    "a\\moveright.5em\\raise.7em\\moveright.5em Xb"
  ],
  [
    "triple-rupup",
    "a\\moveright.5em\\raise.7em\\raise.7em Xb"
  ],
  [
    "triple-rupdown",
    "a\\moveright.5em\\raise.7em\\lower.3em Xb"
  ],
  [
    "pair-r-down-char",
    "a\\moveright.5em\\lower.3em Xb"
  ],
  [
    "pair-r-down-group",
    "a\\moveright.5em\\lower.3em {XY}b"
  ],
  [
    "pair-r-down-style",
    "a\\moveright.5em\\lower.3em \\color{red}XYb"
  ],
  [
    "pair-r-down-fn",
    "a\\moveright.5em\\lower.3em \\sin xb"
  ],
  [
    "pair-r-down-script",
    "a\\moveright.5em\\lower.3em X^2b"
  ],
  [
    "triple-rdownl",
    "a\\moveright.5em\\lower.3em\\moveleft1em Xb"
  ],
  [
    "triple-rdownr",
    "a\\moveright.5em\\lower.3em\\moveright.5em Xb"
  ],
  [
    "triple-rdownup",
    "a\\moveright.5em\\lower.3em\\raise.7em Xb"
  ],
  [
    "triple-rdowndown",
    "a\\moveright.5em\\lower.3em\\lower.3em Xb"
  ],
  [
    "pair-up-l-char",
    "a\\raise.7em\\moveleft1em Xb"
  ],
  [
    "pair-up-l-group",
    "a\\raise.7em\\moveleft1em {XY}b"
  ],
  [
    "pair-up-l-style",
    "a\\raise.7em\\moveleft1em \\color{red}XYb"
  ],
  [
    "pair-up-l-fn",
    "a\\raise.7em\\moveleft1em \\sin xb"
  ],
  [
    "pair-up-l-script",
    "a\\raise.7em\\moveleft1em X^2b"
  ],
  [
    "triple-upll",
    "a\\raise.7em\\moveleft1em\\moveleft1em Xb"
  ],
  [
    "triple-uplr",
    "a\\raise.7em\\moveleft1em\\moveright.5em Xb"
  ],
  [
    "triple-uplup",
    "a\\raise.7em\\moveleft1em\\raise.7em Xb"
  ],
  [
    "triple-upldown",
    "a\\raise.7em\\moveleft1em\\lower.3em Xb"
  ],
  [
    "pair-up-r-char",
    "a\\raise.7em\\moveright.5em Xb"
  ],
  [
    "pair-up-r-group",
    "a\\raise.7em\\moveright.5em {XY}b"
  ],
  [
    "pair-up-r-style",
    "a\\raise.7em\\moveright.5em \\color{red}XYb"
  ],
  [
    "pair-up-r-fn",
    "a\\raise.7em\\moveright.5em \\sin xb"
  ],
  [
    "pair-up-r-script",
    "a\\raise.7em\\moveright.5em X^2b"
  ],
  [
    "triple-uprl",
    "a\\raise.7em\\moveright.5em\\moveleft1em Xb"
  ],
  [
    "triple-uprr",
    "a\\raise.7em\\moveright.5em\\moveright.5em Xb"
  ],
  [
    "triple-uprup",
    "a\\raise.7em\\moveright.5em\\raise.7em Xb"
  ],
  [
    "triple-uprdown",
    "a\\raise.7em\\moveright.5em\\lower.3em Xb"
  ],
  [
    "pair-up-up-char",
    "a\\raise.7em\\raise.7em Xb"
  ],
  [
    "pair-up-up-group",
    "a\\raise.7em\\raise.7em {XY}b"
  ],
  [
    "pair-up-up-style",
    "a\\raise.7em\\raise.7em \\color{red}XYb"
  ],
  [
    "pair-up-up-fn",
    "a\\raise.7em\\raise.7em \\sin xb"
  ],
  [
    "pair-up-up-script",
    "a\\raise.7em\\raise.7em X^2b"
  ],
  [
    "triple-upupl",
    "a\\raise.7em\\raise.7em\\moveleft1em Xb"
  ],
  [
    "triple-upupr",
    "a\\raise.7em\\raise.7em\\moveright.5em Xb"
  ],
  [
    "triple-upupup",
    "a\\raise.7em\\raise.7em\\raise.7em Xb"
  ],
  [
    "triple-upupdown",
    "a\\raise.7em\\raise.7em\\lower.3em Xb"
  ],
  [
    "pair-up-down-char",
    "a\\raise.7em\\lower.3em Xb"
  ],
  [
    "pair-up-down-group",
    "a\\raise.7em\\lower.3em {XY}b"
  ],
  [
    "pair-up-down-style",
    "a\\raise.7em\\lower.3em \\color{red}XYb"
  ],
  [
    "pair-up-down-fn",
    "a\\raise.7em\\lower.3em \\sin xb"
  ],
  [
    "pair-up-down-script",
    "a\\raise.7em\\lower.3em X^2b"
  ],
  [
    "triple-updownl",
    "a\\raise.7em\\lower.3em\\moveleft1em Xb"
  ],
  [
    "triple-updownr",
    "a\\raise.7em\\lower.3em\\moveright.5em Xb"
  ],
  [
    "triple-updownup",
    "a\\raise.7em\\lower.3em\\raise.7em Xb"
  ],
  [
    "triple-updowndown",
    "a\\raise.7em\\lower.3em\\lower.3em Xb"
  ],
  [
    "pair-down-l-char",
    "a\\lower.3em\\moveleft1em Xb"
  ],
  [
    "pair-down-l-group",
    "a\\lower.3em\\moveleft1em {XY}b"
  ],
  [
    "pair-down-l-style",
    "a\\lower.3em\\moveleft1em \\color{red}XYb"
  ],
  [
    "pair-down-l-fn",
    "a\\lower.3em\\moveleft1em \\sin xb"
  ],
  [
    "pair-down-l-script",
    "a\\lower.3em\\moveleft1em X^2b"
  ],
  [
    "triple-downll",
    "a\\lower.3em\\moveleft1em\\moveleft1em Xb"
  ],
  [
    "triple-downlr",
    "a\\lower.3em\\moveleft1em\\moveright.5em Xb"
  ],
  [
    "triple-downlup",
    "a\\lower.3em\\moveleft1em\\raise.7em Xb"
  ],
  [
    "triple-downldown",
    "a\\lower.3em\\moveleft1em\\lower.3em Xb"
  ],
  [
    "pair-down-r-char",
    "a\\lower.3em\\moveright.5em Xb"
  ],
  [
    "pair-down-r-group",
    "a\\lower.3em\\moveright.5em {XY}b"
  ],
  [
    "pair-down-r-style",
    "a\\lower.3em\\moveright.5em \\color{red}XYb"
  ],
  [
    "pair-down-r-fn",
    "a\\lower.3em\\moveright.5em \\sin xb"
  ],
  [
    "pair-down-r-script",
    "a\\lower.3em\\moveright.5em X^2b"
  ],
  [
    "triple-downrl",
    "a\\lower.3em\\moveright.5em\\moveleft1em Xb"
  ],
  [
    "triple-downrr",
    "a\\lower.3em\\moveright.5em\\moveright.5em Xb"
  ],
  [
    "triple-downrup",
    "a\\lower.3em\\moveright.5em\\raise.7em Xb"
  ],
  [
    "triple-downrdown",
    "a\\lower.3em\\moveright.5em\\lower.3em Xb"
  ],
  [
    "pair-down-up-char",
    "a\\lower.3em\\raise.7em Xb"
  ],
  [
    "pair-down-up-group",
    "a\\lower.3em\\raise.7em {XY}b"
  ],
  [
    "pair-down-up-style",
    "a\\lower.3em\\raise.7em \\color{red}XYb"
  ],
  [
    "pair-down-up-fn",
    "a\\lower.3em\\raise.7em \\sin xb"
  ],
  [
    "pair-down-up-script",
    "a\\lower.3em\\raise.7em X^2b"
  ],
  [
    "triple-downupl",
    "a\\lower.3em\\raise.7em\\moveleft1em Xb"
  ],
  [
    "triple-downupr",
    "a\\lower.3em\\raise.7em\\moveright.5em Xb"
  ],
  [
    "triple-downupup",
    "a\\lower.3em\\raise.7em\\raise.7em Xb"
  ],
  [
    "triple-downupdown",
    "a\\lower.3em\\raise.7em\\lower.3em Xb"
  ],
  [
    "pair-down-down-char",
    "a\\lower.3em\\lower.3em Xb"
  ],
  [
    "pair-down-down-group",
    "a\\lower.3em\\lower.3em {XY}b"
  ],
  [
    "pair-down-down-style",
    "a\\lower.3em\\lower.3em \\color{red}XYb"
  ],
  [
    "pair-down-down-fn",
    "a\\lower.3em\\lower.3em \\sin xb"
  ],
  [
    "pair-down-down-script",
    "a\\lower.3em\\lower.3em X^2b"
  ],
  [
    "triple-downdownl",
    "a\\lower.3em\\lower.3em\\moveleft1em Xb"
  ],
  [
    "triple-downdownr",
    "a\\lower.3em\\lower.3em\\moveright.5em Xb"
  ],
  [
    "triple-downdownup",
    "a\\lower.3em\\lower.3em\\raise.7em Xb"
  ],
  [
    "triple-downdowndown",
    "a\\lower.3em\\lower.3em\\lower.3em Xb"
  ],
  [
    "prefix-\\not-moveleft",
    "\\not\\moveleft1em X"
  ],
  [
    "prefix-\\not-moveright",
    "\\not\\moveright1em X"
  ],
  [
    "prefix-\\dots-moveleft",
    "\\dots\\moveleft1em X"
  ],
  [
    "prefix-\\dots-moveright",
    "\\dots\\moveright1em X"
  ],
  [
    "prefix-\\sin-moveleft",
    "\\sin\\moveleft1em X"
  ],
  [
    "prefix-\\sin-moveright",
    "\\sin\\moveright1em X"
  ],
  [
    "prefix-x'-moveleft",
    "x'\\moveleft1em X"
  ],
  [
    "prefix-x'-moveright",
    "x'\\moveright1em X"
  ],
  [
    "prefix-\\color{red}-moveleft",
    "\\color{red}\\moveleft1em X"
  ],
  [
    "prefix-\\color{red}-moveright",
    "\\color{red}\\moveright1em X"
  ],
  [
    "prefix-\\Large-moveleft",
    "\\Large\\moveleft1em X"
  ],
  [
    "prefix-\\Large-moveright",
    "\\Large\\moveright1em X"
  ],
  [
    "prefix-\\bf-moveleft",
    "\\bf\\moveleft1em X"
  ],
  [
    "prefix-\\bf-moveright",
    "\\bf\\moveright1em X"
  ],
  [
    "additional-moveleft-prime",
    "a\\moveleft1em X'Y"
  ],
  [
    "additional-moveleft-double-prime",
    "a\\moveleft1em X''_aY"
  ],
  [
    "additional-moveleft-fn-close",
    "a\\moveleft1em \\sin"
  ],
  [
    "additional-moveleft-not-close",
    "a\\moveleft1em \\not"
  ],
  [
    "additional-moveleft-dots-close",
    "a\\moveleft1em \\dots"
  ],
  [
    "additional-moveleft-not-prime",
    "a\\moveleft1em \\not'X"
  ],
  [
    "additional-moveleft-relax-then-box",
    "a\\moveleft1em \\relax X"
  ],
  [
    "additional-moveleft-empty-font",
    "a\\moveleft1em \\bf{}Y"
  ],
  [
    "additional-moveleft-operator-limits",
    "a\\moveleft1em \\sum\\limits_a^b X"
  ],
  [
    "additional-moveleft-operatorname",
    "a\\moveleft1em \\operatorname{abc}X"
  ],
  [
    "additional-moveleft-nested-styles",
    "a\\moveleft1em \\color{red}\\Large X\\color{blue}Y"
  ],
  [
    "additional-moveleft-position-between-styles",
    "a\\moveleft1em \\color{red}\\raise.3em\\Large X"
  ],
  [
    "additional-moveleft-font-after-style",
    "a\\moveleft1em \\color{red}X\\bf Y"
  ],
  [
    "additional-moveleft-empty-scope",
    "a\\moveleft1em {}X"
  ],
  [
    "additional-moveleft-dotspace",
    "a\\moveleft1em \\,X"
  ],
  [
    "additional-moveleft-null-fence",
    "a\\moveleft1em \\left. X\\right|Y"
  ],
  [
    "additional-moveleft-horizontal-twice",
    "a\\moveleft1em \\moveleft1em X"
  ],
  [
    "additional-moveleft-physics-qty",
    "a\\moveleft1em \\qty(x)Y"
  ],
  [
    "additional-moveleft-physics-dv-star",
    "a\\moveleft1em \\dv*(x)Y"
  ],
  [
    "additional-moveleft-braket-single",
    "a\\moveleft1em \\bra X"
  ],
  [
    "additional-moveleft-lapped",
    "a\\moveleft1em \\llap{X}Y"
  ],
  [
    "additional-moveleft-alone-cases",
    "\\moveleft1em \\cases{x&if positive\\\\0&otherwise}"
  ],
  [
    "additional-moveleft-grouped-cases",
    "a\\moveleft1em {\\cases{x&if positive\\\\0&otherwise}}"
  ],
  [
    "additional-moveleft-internal-math",
    "\\text{$\\moveleft1em X$}"
  ],
  [
    "additional-moveleft-root-index",
    "\\sqrt[\\moveleft1em n]{x}"
  ],
  [
    "additional-moveleft-matrix-cell",
    "\\matrix{\\moveleft1em X&Y\\\\Z&W}"
  ],
  [
    "additional-moveleft-cases-math",
    "\\cases{\\moveleft1em X&if positive}"
  ],
  [
    "additional-moveleft-cases-text-math",
    "\\cases{x&if $\\moveleft1em X$}"
  ],
  [
    "additional-moveleft-boxed",
    "\\boxed{\\moveleft1em X}"
  ],
  [
    "additional-moveleft-stackrel",
    "\\stackrel{\\moveleft1em X}{Y}"
  ],
  [
    "additional-moveleft-leading-subscript",
    "_a\\moveleft1em X"
  ],
  [
    "additional-moveleft-pending-subscript",
    "X_\\moveleft1em Y"
  ],
  [
    "additional-moveleft-pending-script-bad-dimension",
    "X_\\moveleft{bad}Y"
  ],
  [
    "additional-moveleft-double-script",
    "X_a_b\\moveleft1em Y"
  ],
  [
    "additional-moveleft-over-right",
    "\\left(\\moveleft1em \\over X\\right)"
  ],
  [
    "additional-moveleft-group",
    "\\left(\\moveleft1em }"
  ],
  [
    "additional-moveleft-left-missing",
    "\\left(\\moveleft1em \\left(X"
  ],
  [
    "additional-moveleft-linebreak-invalid",
    "\\left(\\moveleft1em \\\\[-bad]"
  ],
  [
    "additional-moveleft-middle-invalid",
    "\\left(\\moveleft1em \\middle a"
  ],
  [
    "additional-moveleft-relax-right",
    "\\left(\\moveleft1em \\relax\\right)"
  ],
  [
    "additional-moveleft-style-right",
    "\\left(\\moveleft1em \\color{red}\\right)"
  ],
  [
    "additional-moveleft-style-empty-middle",
    "\\left(\\moveleft1em \\color{red}\\middle|X\\right)"
  ],
  [
    "additional-moveright-prime",
    "a\\moveright1em X'Y"
  ],
  [
    "additional-moveright-double-prime",
    "a\\moveright1em X''_aY"
  ],
  [
    "additional-moveright-fn-close",
    "a\\moveright1em \\sin"
  ],
  [
    "additional-moveright-not-close",
    "a\\moveright1em \\not"
  ],
  [
    "additional-moveright-dots-close",
    "a\\moveright1em \\dots"
  ],
  [
    "additional-moveright-not-prime",
    "a\\moveright1em \\not'X"
  ],
  [
    "additional-moveright-relax-then-box",
    "a\\moveright1em \\relax X"
  ],
  [
    "additional-moveright-empty-font",
    "a\\moveright1em \\bf{}Y"
  ],
  [
    "additional-moveright-operator-limits",
    "a\\moveright1em \\sum\\limits_a^b X"
  ],
  [
    "additional-moveright-operatorname",
    "a\\moveright1em \\operatorname{abc}X"
  ],
  [
    "additional-moveright-nested-styles",
    "a\\moveright1em \\color{red}\\Large X\\color{blue}Y"
  ],
  [
    "additional-moveright-position-between-styles",
    "a\\moveright1em \\color{red}\\raise.3em\\Large X"
  ],
  [
    "additional-moveright-font-after-style",
    "a\\moveright1em \\color{red}X\\bf Y"
  ],
  [
    "additional-moveright-empty-scope",
    "a\\moveright1em {}X"
  ],
  [
    "additional-moveright-dotspace",
    "a\\moveright1em \\,X"
  ],
  [
    "additional-moveright-null-fence",
    "a\\moveright1em \\left. X\\right|Y"
  ],
  [
    "additional-moveright-horizontal-twice",
    "a\\moveright1em \\moveright1em X"
  ],
  [
    "additional-moveright-physics-qty",
    "a\\moveright1em \\qty(x)Y"
  ],
  [
    "additional-moveright-physics-dv-star",
    "a\\moveright1em \\dv*(x)Y"
  ],
  [
    "additional-moveright-braket-single",
    "a\\moveright1em \\bra X"
  ],
  [
    "additional-moveright-lapped",
    "a\\moveright1em \\llap{X}Y"
  ],
  [
    "additional-moveright-alone-cases",
    "\\moveright1em \\cases{x&if positive\\\\0&otherwise}"
  ],
  [
    "additional-moveright-grouped-cases",
    "a\\moveright1em {\\cases{x&if positive\\\\0&otherwise}}"
  ],
  [
    "additional-moveright-internal-math",
    "\\text{$\\moveright1em X$}"
  ],
  [
    "additional-moveright-root-index",
    "\\sqrt[\\moveright1em n]{x}"
  ],
  [
    "additional-moveright-matrix-cell",
    "\\matrix{\\moveright1em X&Y\\\\Z&W}"
  ],
  [
    "additional-moveright-cases-math",
    "\\cases{\\moveright1em X&if positive}"
  ],
  [
    "additional-moveright-cases-text-math",
    "\\cases{x&if $\\moveright1em X$}"
  ],
  [
    "additional-moveright-boxed",
    "\\boxed{\\moveright1em X}"
  ],
  [
    "additional-moveright-stackrel",
    "\\stackrel{\\moveright1em X}{Y}"
  ],
  [
    "additional-moveright-leading-subscript",
    "_a\\moveright1em X"
  ],
  [
    "additional-moveright-pending-subscript",
    "X_\\moveright1em Y"
  ],
  [
    "additional-moveright-pending-script-bad-dimension",
    "X_\\moveright{bad}Y"
  ],
  [
    "additional-moveright-double-script",
    "X_a_b\\moveright1em Y"
  ],
  [
    "additional-moveright-over-right",
    "\\left(\\moveright1em \\over X\\right)"
  ],
  [
    "additional-moveright-group",
    "\\left(\\moveright1em }"
  ],
  [
    "additional-moveright-left-missing",
    "\\left(\\moveright1em \\left(X"
  ],
  [
    "additional-moveright-linebreak-invalid",
    "\\left(\\moveright1em \\\\[-bad]"
  ],
  [
    "additional-moveright-middle-invalid",
    "\\left(\\moveright1em \\middle a"
  ],
  [
    "additional-moveright-relax-right",
    "\\left(\\moveright1em \\relax\\right)"
  ],
  [
    "additional-moveright-style-right",
    "\\left(\\moveright1em \\color{red}\\right)"
  ],
  [
    "additional-moveright-style-empty-middle",
    "\\left(\\moveright1em \\color{red}\\middle|X\\right)"
  ],
  [
    "additional-vertical1",
    "A\\raise.3em\\lower.2em X"
  ],
  [
    "additional-vertical2",
    "A\\raise.3em\\lower.2em\\color{red}X"
  ],
  [
    "additional-ordinary-kerns",
    "A\\kern-1em X\\kern1em B"
  ],
  [
    "additional-ordinary-hskip",
    "A\\hskip1em X\\hskip-1em B"
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
fs.writeFileSync(path.join(__dirname, 'horizontal_position_mathjax_3_2_2.json'), JSON.stringify({
  mathjaxGitCommit: 'ad8f5c21cb810236551da8c6512ba733e67357ee',
  assets: hashes,
  scope: 'Complete unmodified SVGs from the frozen D2 MathJax bundle; each case uses a fresh runtime.',
  cases,
}, null, 2) + '\n');
