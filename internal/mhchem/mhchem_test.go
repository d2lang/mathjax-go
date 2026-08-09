// Copyright (c) 2015-2021 Martin Hensel
// SPDX-License-Identifier: Apache-2.0

package mhchem

import "testing"

// TestUpstream411 reproduces every assertion in mhchemparser 4.1.1/test/test.html.
func TestUpstream411(t *testing.T) {
	tests := []struct {
		mode        Mode
		input, want string
	}{
		{ModeTeX, "m_{\\ce{H2O}} = \\pu{1.2kg}", "m_{{\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}} = {1.2~\\mathrm{kg}}"},
		{ModeCE, "CO2 + C -> 2 CO", "{\\mathrm{CO}{\\vphantom{A}}_{\\smash[t]{2}} {}+{} \\mathrm{C} {}\\mathrel{\\longrightarrow}{} 2\\,\\mathrm{CO}}"},
		{ModeCE, "Hg^2+ ->[I-] HgI2 ->[I-] [Hg^{II}I4]^2-", "{\\mathrm{Hg}{\\vphantom{A}}^{2+} {}\\mathrel{\\xrightarrow{\\mathrm{I}{\\vphantom{A}}^{-}}}{} \\mathrm{HgI}{\\vphantom{A}}_{\\smash[t]{2}} {}\\mathrel{\\xrightarrow{\\mathrm{I}{\\vphantom{A}}^{-}}}{} [\\mathrm{Hg}{\\vphantom{A}}^{\\mathrm{II}}\\mathrm{I}{\\vphantom{A}}_{\\smash[t]{4}}]{\\vphantom{A}}^{2-}}"},
		{ModeCE, "H2O", "{\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "Sb2O3", "{\\mathrm{Sb}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}{\\vphantom{A}}_{\\smash[t]{3}}}"},
		{ModeCE, "H+", "{\\mathrm{H}{\\vphantom{A}}^{+}}"},
		{ModeCE, "CrO4^2-", "{\\mathrm{CrO}{\\vphantom{A}}_{\\smash[t]{4}}{\\vphantom{A}}^{2-}}"},
		{ModeCE, "[AgCl2]-", "{[\\mathrm{AgCl}{\\vphantom{A}}_{\\smash[t]{2}}]{\\vphantom{A}}^{-}}"},
		{ModeCE, "Y^99+", "{\\mathrm{Y}{\\vphantom{A}}^{99+}}"},
		{ModeCE, "Y^{99+}", "{\\mathrm{Y}{\\vphantom{A}}^{99+}}"},
		{ModeCE, "2 H2O", "{2\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "2H2O", "{2\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "0.5 H2O", "{0.5\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "1/2 H2O", "{\\mathchoice{\\textstyle\\frac{1}{2}}{\\frac{1}{2}}{\\frac{1}{2}}{\\frac{1}{2}}\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "(1/2) H2O", "{(1/2)\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "$n$ H2O", "{n \\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "n H2O", "{n\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "nH2O", "{n\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "n/2 H2O", "{\\mathchoice{\\textstyle\\frac{n}{2}}{\\frac{n}{2}}{\\frac{n}{2}}{\\frac{n}{2}}\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "^{227}_{90}Th+", "{{\\vphantom{A}}^{\\hphantom{227}}_{\\hphantom{90}}\\mkern-1.5mu{\\vphantom{A}}^{\\smash[t]{\\vphantom{2}}\\llap{227}}_{\\vphantom{2}\\llap{\\smash[t]{90}}}\\mathrm{Th}{\\vphantom{A}}^{+}}"},
		{ModeCE, "^227_90Th+", "{{\\vphantom{A}}^{\\hphantom{227}}_{\\hphantom{90}}\\mkern-1.5mu{\\vphantom{A}}^{\\smash[t]{\\vphantom{2}}\\llap{227}}_{\\vphantom{2}\\llap{\\smash[t]{90}}}\\mathrm{Th}{\\vphantom{A}}^{+}}"},
		{ModeCE, "^{0}_{-1}n^{-}", "{{\\vphantom{A}}^{\\hphantom{0}}_{\\hphantom{-1}}\\mkern-1.5mu{\\vphantom{A}}^{\\smash[t]{\\vphantom{2}}\\llap{0}}_{\\vphantom{2}\\llap{\\smash[t]{-1}}}\\mathrm{n}{\\vphantom{A}}^{-}}"},
		{ModeCE, "^0_-1n-", "{{\\vphantom{A}}^{\\hphantom{0}}_{\\hphantom{-1}}\\mkern-1.5mu{\\vphantom{A}}^{\\smash[t]{\\vphantom{2}}\\llap{0}}_{\\vphantom{2}\\llap{\\smash[t]{-1}}}\\mathrm{n}{\\vphantom{A}}^{-}}"},
		{ModeCE, "H{}^3HO", "{\\mathrm{H}\\mkern2mu{\\vphantom{A}}^{\\hphantom{3}}_{\\hphantom{}}\\mkern-1.5mu{\\vphantom{A}}^{\\smash[t]{\\vphantom{2}}\\llap{3}}_{\\vphantom{2}\\llap{\\smash[t]{}}}\\mathrm{HO}}"},
		{ModeCE, "H^3HO", "{\\mathrm{H}\\mkern2mu{\\vphantom{A}}^{\\hphantom{3}}_{\\hphantom{}}\\mkern-1.5mu{\\vphantom{A}}^{\\smash[t]{\\vphantom{2}}\\llap{3}}_{\\vphantom{2}\\llap{\\smash[t]{}}}\\mathrm{HO}}"},
		{ModeCE, "A -> B", "{\\mathrm{A} {}\\mathrel{\\longrightarrow}{} \\mathrm{B}}"},
		{ModeCE, "A <- B", "{\\mathrm{A} {}\\mathrel{\\longleftarrow}{} \\mathrm{B}}"},
		{ModeCE, "A <-> B", "{\\mathrm{A} {}\\mathrel{\\longleftrightarrow}{} \\mathrm{B}}"},
		{ModeCE, "A <--> B", "{\\mathrm{A} {}\\mathrel{\\longleftrightarrows}{} \\mathrm{B}}"},
		{ModeCE, "A <=> B", "{\\mathrm{A} {}\\mathrel{\\longrightleftharpoons}{} \\mathrm{B}}"},
		{ModeCE, "A <=>> B", "{\\mathrm{A} {}\\mathrel{\\longRightleftharpoons}{} \\mathrm{B}}"},
		{ModeCE, "A <<=> B", "{\\mathrm{A} {}\\mathrel{\\longLeftrightharpoons}{} \\mathrm{B}}"},
		{ModeCE, "A ->[H2O] B", "{\\mathrm{A} {}\\mathrel{\\xrightarrow{\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}}{} \\mathrm{B}}"},
		{ModeCE, "A ->[{text above}][{text below}] B", "{\\mathrm{A} {}\\mathrel{\\xrightarrow[{{\\text{text below}}}]{{\\text{text above}}}}{} \\mathrm{B}}"},
		{ModeCE, "A ->[$x$][$x_i$] B", "{\\mathrm{A} {}\\mathrel{\\xrightarrow[{x_i }]{x }}{} \\mathrm{B}}"},
		{ModeCE, "(NH4)2S", "{(\\mathrm{NH}{\\vphantom{A}}_{\\smash[t]{4}}){\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{S}}"},
		{ModeCE, "[\\{(X2)3\\}2]^3+", "{[\\{(\\mathrm{X}{\\vphantom{A}}_{\\smash[t]{2}}){\\vphantom{A}}_{\\smash[t]{3}}\\}{\\vphantom{A}}_{\\smash[t]{2}}]{\\vphantom{A}}^{3+}}"},
		{ModeCE, "CH4 + 2 $\\left( \\ce{O2 + 79/21 N2} \\right)$", "{\\mathrm{CH}{\\vphantom{A}}_{\\smash[t]{4}} {}+{} 2\\,\\left(  \\mathrm{O}{\\vphantom{A}}_{\\smash[t]{2}} {}+{} \\mathchoice{\\textstyle\\frac{79}{21}}{\\frac{79}{21}}{\\frac{79}{21}}{\\frac{79}{21}}\\,\\mathrm{N}{\\vphantom{A}}_{\\smash[t]{2}} \\right) }"},
		{ModeCE, "H2(aq)", "{\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mskip2mu (\\mathrm{aq})}"},
		{ModeCE, "CO3^2-_{(aq)}", "{\\mathrm{CO}{\\vphantom{A}}_{\\smash[t]{3}}{\\vphantom{A}}^{2-}{\\vphantom{A}}_{\\smash[t]{\\mskip1mu (\\mathrm{aq})}}}"},
		{ModeCE, "CO3^2-{}_{(aq)}", "{\\mathrm{CO}{\\vphantom{A}}_{\\smash[t]{3}}{\\vphantom{A}}^{2-}{\\vphantom{A}}_{\\smash[t]{\\mskip1mu (\\mathrm{aq})}}}"},
		{ModeCE, "NaOH(aq,$\\infty$)", "{\\mathrm{NaOH}\\mskip2mu (\\mathrm{aq},\\infty )}"},
		{ModeCE, "ZnS ($c$)", "{\\mathrm{ZnS}\\mskip2mu (c )}"},
		{ModeCE, "ZnS (\\ca$c$)", "{\\mathrm{ZnS}\\mskip2mu ({\\sim}c )}"},
		{ModeCE, "NO_x", "{\\mathrm{NO}{\\vphantom{A}}_{\\smash[t]{x }}}"},
		{ModeCE, "Fe^n+", "{\\mathrm{Fe}{\\vphantom{A}}^{n +}}"},
		// These two upstream inputs use unknown escapes in non-strict JavaScript
		// string literals; \Delta and \mu therefore evaluate to Delta and mu.
		{ModeCE, "x Na(NH4)HPO4 ->[Delta] (NaPO3)_x + x NH3 ^ + x H2O", "{x\\,\\mathrm{Na}(\\mathrm{NH}{\\vphantom{A}}_{\\smash[t]{4}})\\mathrm{HPO}{\\vphantom{A}}_{\\smash[t]{4}} {}\\mathrel{\\xrightarrow{\\mathrm{Delta}}}{} (\\mathrm{NaPO}{\\vphantom{A}}_{\\smash[t]{3}}){\\vphantom{A}}_{\\smash[t]{x }} {}+{} x\\,\\mathrm{NH}{\\vphantom{A}}_{\\smash[t]{3}} \\uparrow{}  {}+{} x\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "mu-Cl", "{\\mathrm{mu}{-}\\mathrm{Cl}}"},
		{ModeCE, "[Pt(\\eta^2-C2H4)Cl3]-", "{[\\mathrm{Pt}(\\mathrm{\\eta}{\\vphantom{A}}^{2}\\text{-}\\mathrm{C}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{4}})\\mathrm{Cl}{\\vphantom{A}}_{\\smash[t]{3}}]{\\vphantom{A}}^{-}}"},
		{ModeCE, "\\beta +", "{\\mathrm{\\beta }{\\vphantom{A}}^{+}}"},
		{ModeCE, "^40_18Ar + \\gamma{} + \\nu_e", "{{\\vphantom{A}}^{\\hphantom{40}}_{\\hphantom{18}}\\mkern-1.5mu{\\vphantom{A}}^{\\smash[t]{\\vphantom{2}}\\llap{40}}_{\\vphantom{2}\\llap{\\smash[t]{18}}}\\mathrm{Ar} {}+{} \\mathrm{\\gamma{}} {}+{} \\mathrm{\\nu}{\\vphantom{A}}_{\\smash[t]{e }}}"},
		{ModeCE, "NaOH(aq,$\\infty$)", "{\\mathrm{NaOH}\\mskip2mu (\\mathrm{aq},\\infty )}"},
		{ModeCE, "Fe(CN)_{$\\frac{6}{2}$}", "{\\mathrm{Fe}(\\mathrm{CN}){\\vphantom{A}}_{\\smash[t]{\\frac{6}{2} }}}"},
		{ModeCE, "X_{$i$}^{$x$}", "{\\mathrm{X}{\\vphantom{A}}_{\\smash[t]{i }}{\\vphantom{A}}^{x }}"},
		{ModeCE, "X_$i$^$x$", "{\\mathrm{X}{\\vphantom{A}}_{\\smash[t]{i }}{\\vphantom{A}}^{x }}"},
		{ModeCE, "$cis${-}[PtCl2(NH3)2]", "{cis {\\text{-}}[\\mathrm{PtCl}{\\vphantom{A}}_{\\smash[t]{2}}(\\mathrm{NH}{\\vphantom{A}}_{\\smash[t]{3}}){\\vphantom{A}}_{\\smash[t]{2}}]}"},
		{ModeCE, "CuS($hP12$)", "{\\mathrm{CuS}(hP12 )}"},
		{ModeCE, "{Gluconic Acid} + H2O2", "{{\\text{Gluconic Acid}} {}+{} \\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}{\\vphantom{A}}_{\\smash[t]{2}}}"},
		{ModeCE, "X_{{red}}", "{\\mathrm{X}{\\vphantom{A}}_{\\smash[t]{\\text{red}}}}"},
		{ModeCE, "{(+)}_589{-}[Co(en)3]Cl3", "{{\\text{(+)}}{\\vphantom{A}}_{\\smash[t]{589}}{\\text{-}}[\\mathrm{Co}(\\mathrm{en}){\\vphantom{A}}_{\\smash[t]{3}}]\\mathrm{Cl}{\\vphantom{A}}_{\\smash[t]{3}}}"},
		{ModeCE, "C6H5-CHO", "{\\mathrm{C}{\\vphantom{A}}_{\\smash[t]{6}}\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{5}}{-}\\mathrm{CHO}}"},
		{ModeCE, "A-B=C#D", "{\\mathrm{A}{-}\\mathrm{B}{=}\\mathrm{C}{\\equiv}\\mathrm{D}}"},
		{ModeCE, "A\\bond{-}B\\bond{=}C\\bond{#}D", "{\\mathrm{A}{-}\\mathrm{B}{=}\\mathrm{C}{\\equiv}\\mathrm{D}}"},
		{ModeCE, "A\\bond{1}B\\bond{2}C\\bond{3}D", "{\\mathrm{A}{-}\\mathrm{B}{=}\\mathrm{C}{\\equiv}\\mathrm{D}}"},
		{ModeCE, "A\\bond{~}B\\bond{~-}C", "{\\mathrm{A}{\\tripledash}\\mathrm{B}{\\rlap{\\lower.1em{-}}\\raise.1em{\\tripledash}}\\mathrm{C}}"},
		{ModeCE, "A\\bond{~--}B\\bond{~=}C\\bond{-~-}D", "{\\mathrm{A}{\\rlap{\\lower.2em{-}}\\rlap{\\raise.2em{\\tripledash}}-}\\mathrm{B}{\\rlap{\\lower.2em{-}}\\rlap{\\raise.2em{\\tripledash}}-}\\mathrm{C}{\\rlap{\\lower.2em{-}}\\rlap{\\raise.2em{-}}\\tripledash}\\mathrm{D}}"},
		{ModeCE, "A\\bond{...}B\\bond{....}C", "{\\mathrm{A}{{\\cdot}{\\cdot}{\\cdot}}\\mathrm{B}{{\\cdot}{\\cdot}{\\cdot}{\\cdot}}\\mathrm{C}}"},
		{ModeCE, "A\\bond{->}B\\bond{<-}C", "{\\mathrm{A}{\\rightarrow}\\mathrm{B}{\\leftarrow}\\mathrm{C}}"},
		{ModeCE, "KCr(SO4)2*12H2O", "{\\mathrm{KCr}(\\mathrm{SO}{\\vphantom{A}}_{\\smash[t]{4}}){\\vphantom{A}}_{\\smash[t]{2}}\\,{\\cdot}\\,12\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "KCr(SO4)2.12H2O", "{\\mathrm{KCr}(\\mathrm{SO}{\\vphantom{A}}_{\\smash[t]{4}}){\\vphantom{A}}_{\\smash[t]{2}}\\,{\\cdot}\\,12\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "KCr(SO4)2 * 12 H2O", "{\\mathrm{KCr}(\\mathrm{SO}{\\vphantom{A}}_{\\smash[t]{4}}){\\vphantom{A}}_{\\smash[t]{2}}\\,{\\cdot}\\,12\\,\\mathrm{H}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}}"},
		{ModeCE, "Fe^{II}Fe^{III}2O4", "{\\mathrm{Fe}{\\vphantom{A}}^{\\mathrm{II}}\\mathrm{Fe}{\\vphantom{A}}^{\\mathrm{III}}{\\vphantom{A}}_{\\smash[t]{2}}\\mathrm{O}{\\vphantom{A}}_{\\smash[t]{4}}}"},
		{ModeCE, "OCO^{.-}", "{\\mathrm{OCO}{\\vphantom{A}}^{\\mkern1mu \\bullet\\mkern1mu -}}"},
		{ModeCE, "NO^{(2.)-}", "{\\mathrm{NO}{\\vphantom{A}}^{(2\\mkern1mu \\bullet\\mkern1mu )-}}"},
		{ModeCE, "Li^x_{Li,1-2x}Mg^._{Li,x}$V$'_{Li,x}Cl^x_{Cl}", "{\\mathrm{Li}{\\vphantom{A}}^{{\\times}}_{\\smash[t]{\\mathrm{Li}{,}\\mkern1mu 1-2x }}\\mathrm{Mg}{\\vphantom{A}}^{\\mkern1mu \\bullet\\mkern1mu }_{\\smash[t]{\\mathrm{Li}{,}\\mkern1mu x }}V {\\vphantom{A}}^{\\prime }_{\\smash[t]{\\mathrm{Li}{,}\\mkern1mu x }}\\mathrm{Cl}{\\vphantom{A}}^{{\\times}}_{\\smash[t]{\\mathrm{Cl}}}}"},
		{ModeCE, "O''_{i,x}", "{\\mathrm{O}{\\vphantom{A}}^{\\prime \\prime }_{\\smash[t]{\\mathrm{i}{,}\\mkern1mu x }}}"},
		{ModeCE, "M^{..}_i", "{\\mathrm{M}{\\vphantom{A}}^{\\mkern1mu \\bullet\\mkern1mu \\mkern1mu \\bullet\\mkern1mu }_{\\smash[t]{\\mathrm{i}}}}"},
		{ModeCE, "$V$^{4'}_{Ti}", "{V {\\vphantom{A}}^{4\\prime }_{\\smash[t]{\\mathrm{Ti}}}}"},
		{ModeCE, "V_{V,1}C_{C,0.8}$V$_{C,0.2}", "{\\mathrm{V}{\\vphantom{A}}_{\\smash[t]{\\mathrm{V}{,}\\mkern1mu 1}}\\mathrm{C}{\\vphantom{A}}_{\\smash[t]{\\mathrm{C}{,}\\mkern1mu 0.8}}V {\\vphantom{A}}_{\\smash[t]{\\mathrm{C}{,}\\mkern1mu 0.2}}}"},
		{ModeCE, "A + B", "{\\mathrm{A} {}+{} \\mathrm{B}}"},
		{ModeCE, "A - B", "{\\mathrm{A} {}-{} \\mathrm{B}}"},
		{ModeCE, "A = B", "{\\mathrm{A} {}={} \\mathrm{B}}"},
		{ModeCE, "A \\pm B", "{\\mathrm{A} {}\\pm{} \\mathrm{B}}"},
		{ModeCE, "SO4^2- + Ba^2+ -> BaSO4 v", "{\\mathrm{SO}{\\vphantom{A}}_{\\smash[t]{4}}{\\vphantom{A}}^{2-} {}+{} \\mathrm{Ba}{\\vphantom{A}}^{2+} {}\\mathrel{\\longrightarrow}{} \\mathrm{BaSO}{\\vphantom{A}}_{\\smash[t]{4}} \\downarrow{} }"},
		{ModeCE, "A v B (v) -> B ^ B (^)", "{\\mathrm{A} \\downarrow{} ~\\mathrm{B} \\downarrow{}  {}\\mathrel{\\longrightarrow}{} \\mathrm{B} \\uparrow{} ~\\mathrm{B} \\uparrow{} }"},
		{ModeCE, "NO^*", "{\\mathrm{NO}{\\vphantom{A}}^{*}}"},
		{ModeCE, "1s^2-N", "{1\\mathrm{s}{\\vphantom{A}}^{2}\\text{-}\\mathrm{N}}"},
		{ModeCE, "n-Pr", "{n \\text{-}\\mathrm{Pr}}"},
		{ModeCE, "iPr", "{\\mathrm{iPr}}"},
		{ModeCE, "\\ca Fe", "{{\\sim}\\mathrm{Fe}}"},
		{ModeCE, "A, B, C; F", "{\\mathrm{A}{,}\\mkern6mu \\mathrm{B}{,}\\mkern6mu \\mathrm{C}{;}\\mkern6mu \\mathrm{F}}"},
		{ModeCE, "{and others}", "{{\\text{and others}}}"},
		{ModeCE, "Zn^2+  <=>[+ 2OH-][+ 2H+]  $\\underset{\\text{amphoteres Hydroxid}}{\\ce{Zn(OH)2 v}}$  <=>[+ 2OH-][+ 2H+]  $\\underset{\\text{Hydroxozikat}}{\\ce{[Zn(OH)4]^2-}}$", "{\\mathrm{Zn}{\\vphantom{A}}^{2+} {}\\mathrel{\\underset{\\lower6mu{ {}+{} 2\\,\\mathrm{H}{\\vphantom{A}}^{+}}}{\\overset{ {}+{} 2\\,\\mathrm{OH}{\\vphantom{A}}^{-}}{\\longrightleftharpoons}}}{} \\underset{\\text{amphoteres Hydroxid}}{\\ce{Zn(OH)2 v}}  {}\\mathrel{\\underset{\\lower6mu{ {}+{} 2\\,\\mathrm{H}{\\vphantom{A}}^{+}}}{\\overset{ {}+{} 2\\,\\mathrm{OH}{\\vphantom{A}}^{-}}{\\longrightleftharpoons}}}{} \\underset{\\text{Hydroxozikat}}{\\ce{[Zn(OH)4]^2-}} }"},
		{ModeCE, "$K = \\frac{[\\ce{Hg^2+}][\\ce{Hg}]}{[\\ce{Hg2^2+}]}$", "{K = \\frac{[\\ce{Hg^2+}][\\ce{Hg}]}{[\\ce{Hg2^2+}]} }"},
		{ModeCE, "$K = \\ce{\\frac{[Hg^2+][Hg]}{[Hg2^2+]}}$", "{K =  \\frac{[\\mathrm{Hg}{\\vphantom{A}}^{2+}][\\mathrm{Hg}]}{[\\mathrm{Hg}{\\vphantom{A}}_{\\smash[t]{2}}{\\vphantom{A}}^{2+}]}}"},
		{ModeCE, "Hg^2+ ->[I-]  $\\underset{\\mathrm{red}}{\\ce{HgI2}}$  ->[I-]  $\\underset{\\mathrm{red}}{\\ce{[Hg^{II}I4]^2-}}$", "{\\mathrm{Hg}{\\vphantom{A}}^{2+} {}\\mathrel{\\xrightarrow{\\mathrm{I}{\\vphantom{A}}^{-}}}{} \\underset{\\mathrm{red}}{\\ce{HgI2}}  {}\\mathrel{\\xrightarrow{\\mathrm{I}{\\vphantom{A}}^{-}}}{} \\underset{\\mathrm{red}}{\\ce{[Hg^{II}I4]^2-}} }"},
		{ModePU, "123 kJ", "{123~\\mathrm{kJ}}"},
		{ModePU, "123 mm2", "{123~\\mathrm{mm^{2}}}"},
		{ModePU, "123 J s", "{123~\\mathrm{J}\\mkern3mu \\mathrm{s}}"},
		{ModePU, "123 J*s", "{123~\\mathrm{J}\\mkern1mu{\\cdot}\\mkern1mu \\mathrm{s}}"},
		{ModePU, "123 kJ/mol", "{123~\\mathrm{kJ}/\\mathrm{mol}}"},
		{ModePU, "123 kJ//mol", "{123~\\mathchoice{\\textstyle\\frac{\\mathrm{kJ}}{\\mathrm{mol}}}{\\frac{\\mathrm{kJ}}{\\mathrm{mol}}}{\\frac{\\mathrm{kJ}}{\\mathrm{mol}}}{\\frac{\\mathrm{kJ}}{\\mathrm{mol}}}}"},
		{ModePU, "123 kJ mol^-1", "{123~\\mathrm{kJ}\\mkern3mu \\mathrm{mol^{-1}}}"},
		{ModePU, "123 kJ*mol-1", "{123~\\mathrm{kJ}\\mkern1mu{\\cdot}\\mkern1mu \\mathrm{mol^{-1}}}"},
		{ModePU, "123 kJ.mol-1", "{123~\\mathrm{kJ}\\mkern1mu{\\cdot}\\mkern1mu \\mathrm{mol^{-1}}}"},
		{ModePU, "1.2e3 kJ", "{1.2\\cdot 10^{3}~\\mathrm{kJ}}"},
		{ModePU, "1,2e3 kJ", "{1{,}2\\cdot 10^{3}~\\mathrm{kJ}}"},
		{ModePU, "1.2E3 kJ", "{1.2\\times 10^{3}~\\mathrm{kJ}}"},
		{ModePU, "1,2E3 kJ", "{1{,}2\\times 10^{3}~\\mathrm{kJ}}"},
		{ModePU, "1234", "{1234}"},
		{ModePU, "12345", "{12\\mkern2mu 345}"},
		{ModePU, "1°C", "{1~\\mathrm{{}^{\\circ}C}}"},
		{ModePU, "23.4782(32) m", "{23.4782(32)~\\mathrm{m}}"},
		{ModePU, "8.00001 \\pm 0.00005 nm", "{8.000\\mkern2mu 01 {}\\pm{} 0.000\\mkern2mu 05~\\mathrm{nm}}"},
		{ModePU, ".25", "{.25}"},
		{ModePU, "1 mol ", "{1~\\mathrm{mol}~}"},
		{ModePU, "123 l//100km", "{123~\\mathchoice{\\textstyle\\frac{\\mathrm{l}}{100~\\mathrm{km}}}{\\frac{\\mathrm{l}}{100~\\mathrm{km}}}{\\frac{\\mathrm{l}}{100~\\mathrm{km}}}{\\frac{\\mathrm{l}}{100~\\mathrm{km}}}}"},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode)+"/"+tt.input, func(t *testing.T) {
			got, err := ToTeX(tt.input, tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("ToTeX() mismatch:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestToTeXRejectsUnknownMode(t *testing.T) {
	if _, err := ToTeX("H2O", Mode("unknown")); err == nil {
		t.Fatal("ToTeX accepted an unknown mode")
	}
}
