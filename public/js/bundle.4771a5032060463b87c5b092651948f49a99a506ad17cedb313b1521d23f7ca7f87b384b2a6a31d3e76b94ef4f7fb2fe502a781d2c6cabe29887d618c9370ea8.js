/*! highlight.js v9.17.1 | BSD3 License | git.io/hljslicense */
!function(e){var n="object"==typeof window&&window||"object"==typeof self&&self;"undefined"==typeof exports||exports.nodeType?n&&(n.hljs=e({}),"function"==typeof define&&define.amd&&define([],function(){return n.hljs})):e(exports)}(function(a){var f=[],o=Object.keys,N={},g={},_=!0,n=/^(no-?highlight|plain|text)$/i,E=/\blang(?:uage)?-([\w-]+)\b/i,t=/((^(<[^>]+>|\t|)+|(?:\n)))/gm,r={case_insensitive:"cI",lexemes:"l",contains:"c",keywords:"k",subLanguage:"sL",className:"cN",begin:"b",beginKeywords:"bK",end:"e",endsWithParent:"eW",illegal:"i",excludeBegin:"eB",excludeEnd:"eE",returnBegin:"rB",returnEnd:"rE",variants:"v",IDENT_RE:"IR",UNDERSCORE_IDENT_RE:"UIR",NUMBER_RE:"NR",C_NUMBER_RE:"CNR",BINARY_NUMBER_RE:"BNR",RE_STARTERS_RE:"RSR",BACKSLASH_ESCAPE:"BE",APOS_STRING_MODE:"ASM",QUOTE_STRING_MODE:"QSM",PHRASAL_WORDS_MODE:"PWM",C_LINE_COMMENT_MODE:"CLCM",C_BLOCK_COMMENT_MODE:"CBCM",HASH_COMMENT_MODE:"HCM",NUMBER_MODE:"NM",C_NUMBER_MODE:"CNM",BINARY_NUMBER_MODE:"BNM",CSS_NUMBER_MODE:"CSSNM",REGEXP_MODE:"RM",TITLE_MODE:"TM",UNDERSCORE_TITLE_MODE:"UTM",COMMENT:"C",beginRe:"bR",endRe:"eR",illegalRe:"iR",lexemesRe:"lR",terminators:"t",terminator_end:"tE"},C="</span>",m="Could not find the language '{}', did you forget to load/include a language module?",O={classPrefix:"hljs-",tabReplace:null,useBR:!1,languages:void 0},c="of and for in not or if then".split(" ");function B(e){return e.replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;")}function d(e){return e.nodeName.toLowerCase()}function R(e){return n.test(e)}function i(e){var n,t={},r=Array.prototype.slice.call(arguments,1);for(n in e)t[n]=e[n];return r.forEach(function(e){for(n in e)t[n]=e[n]}),t}function p(e){var a=[];return function e(n,t){for(var r=n.firstChild;r;r=r.nextSibling)3===r.nodeType?t+=r.nodeValue.length:1===r.nodeType&&(a.push({event:"start",offset:t,node:r}),t=e(r,t),d(r).match(/br|hr|img|input/)||a.push({event:"stop",offset:t,node:r}));return t}(e,0),a}function v(e,n,t){var r=0,a="",i=[];function o(){return e.length&&n.length?e[0].offset!==n[0].offset?e[0].offset<n[0].offset?e:n:"start"===n[0].event?e:n:e.length?e:n}function c(e){a+="<"+d(e)+f.map.call(e.attributes,function(e){return" "+e.nodeName+'="'+B(e.value).replace(/"/g,"&quot;")+'"'}).join("")+">"}function l(e){a+="</"+d(e)+">"}function u(e){("start"===e.event?c:l)(e.node)}for(;e.length||n.length;){var s=o();if(a+=B(t.substring(r,s[0].offset)),r=s[0].offset,s===e){for(i.reverse().forEach(l);u(s.splice(0,1)[0]),(s=o())===e&&s.length&&s[0].offset===r;);i.reverse().forEach(c)}else"start"===s[0].event?i.push(s[0].node):i.pop(),u(s.splice(0,1)[0])}return a+B(t.substr(r))}function l(n){return n.v&&!n.cached_variants&&(n.cached_variants=n.v.map(function(e){return i(n,{v:null},e)})),n.cached_variants?n.cached_variants:function e(n){return!!n&&(n.eW||e(n.starts))}(n)?[i(n,{starts:n.starts?i(n.starts):null})]:Object.isFrozen(n)?[i(n)]:[n]}function u(e){if(r&&!e.langApiRestored){for(var n in e.langApiRestored=!0,r)e[n]&&(e[r[n]]=e[n]);(e.c||[]).concat(e.v||[]).forEach(u)}}function M(n,t){var i={};return"string"==typeof n?r("keyword",n):o(n).forEach(function(e){r(e,n[e])}),i;function r(a,e){t&&(e=e.toLowerCase()),e.split(" ").forEach(function(e){var n,t,r=e.split("|");i[r[0]]=[a,(n=r[0],(t=r[1])?Number(t):function(e){return-1!=c.indexOf(e.toLowerCase())}(n)?0:1)]})}}function x(r){function s(e){return e&&e.source||e}function f(e,n){return new RegExp(s(e),"m"+(r.cI?"i":"")+(n?"g":""))}function a(a){var i,e,o={},c=[],l={},t=1;function n(e,n){o[t]=e,c.push([e,n]),t+=new RegExp(n.toString()+"|").exec("").length-1+1}for(var r=0;r<a.c.length;r++){n(e=a.c[r],e.bK?"\\.?(?:"+e.b+")\\.?":e.b)}a.tE&&n("end",a.tE),a.i&&n("illegal",a.i);var u=c.map(function(e){return e[1]});return i=f(function(e,n){for(var t=/\[(?:[^\\\]]|\\.)*\]|\(\??|\\([1-9][0-9]*)|\\./,r=0,a="",i=0;i<e.length;i++){var o=r+=1,c=s(e[i]);for(0<i&&(a+=n),a+="(";0<c.length;){var l=t.exec(c);if(null==l){a+=c;break}a+=c.substring(0,l.index),c=c.substring(l.index+l[0].length),"\\"==l[0][0]&&l[1]?a+="\\"+String(Number(l[1])+o):(a+=l[0],"("==l[0]&&r++)}a+=")"}return a}(u,"|"),!0),l.lastIndex=0,l.exec=function(e){var n;if(0===c.length)return null;i.lastIndex=l.lastIndex;var t=i.exec(e);if(!t)return null;for(var r=0;r<t.length;r++)if(null!=t[r]&&null!=o[""+r]){n=o[""+r];break}return"string"==typeof n?(t.type=n,t.extra=[a.i,a.tE]):(t.type="begin",t.rule=n),t},l}if(r.c&&-1!=r.c.indexOf("self")){if(!_)throw new Error("ERR: contains `self` is not supported at the top-level of a language.  See documentation.");r.c=r.c.filter(function(e){return"self"!=e})}!function n(t,e){t.compiled||(t.compiled=!0,t.k=t.k||t.bK,t.k&&(t.k=M(t.k,r.cI)),t.lR=f(t.l||/\w+/,!0),e&&(t.bK&&(t.b="\\b("+t.bK.split(" ").join("|")+")\\b"),t.b||(t.b=/\B|\b/),t.bR=f(t.b),t.endSameAsBegin&&(t.e=t.b),t.e||t.eW||(t.e=/\B|\b/),t.e&&(t.eR=f(t.e)),t.tE=s(t.e)||"",t.eW&&e.tE&&(t.tE+=(t.e?"|":"")+e.tE)),t.i&&(t.iR=f(t.i)),null==t.relevance&&(t.relevance=1),t.c||(t.c=[]),t.c=Array.prototype.concat.apply([],t.c.map(function(e){return l("self"===e?t:e)})),t.c.forEach(function(e){n(e,t)}),t.starts&&n(t.starts,e),t.t=a(t))}(r)}function S(n,i,a,e){function c(e,n,t,r){if(!t&&""===n)return"";if(!e)return n;var a='<span class="'+(r?"":O.classPrefix);return(a+=e+'">')+n+(t?"":C)}function o(){R+=null!=E.sL?function(){var e="string"==typeof E.sL;if(e&&!N[E.sL])return B(p);var n=e?S(E.sL,p,!0,d[E.sL]):T(p,E.sL.length?E.sL:void 0);return 0<E.relevance&&(v+=n.relevance),e&&(d[E.sL]=n.top),c(n.language,n.value,!1,!0)}():function(){var e,n,t,r,a,i,o;if(!E.k)return B(p);for(r="",n=0,E.lR.lastIndex=0,t=E.lR.exec(p);t;)r+=B(p.substring(n,t.index)),a=E,i=t,o=g.cI?i[0].toLowerCase():i[0],(e=a.k.hasOwnProperty(o)&&a.k[o])?(v+=e[1],r+=c(e[0],B(t[0]))):r+=B(t[0]),n=E.lR.lastIndex,t=E.lR.exec(p);return r+B(p.substr(n))}(),p=""}function l(e){R+=e.cN?c(e.cN,"",!0):"",E=Object.create(e,{parent:{value:E}})}function u(e){var n=e[0],t=e.rule;return t&&t.endSameAsBegin&&(t.eR=new RegExp(n.replace(/[-\/\\^$*+?.()|[\]{}]/g,"\\$&"),"m")),t.skip?p+=n:(t.eB&&(p+=n),o(),t.rB||t.eB||(p=n)),l(t),t.rB?0:n.length}function s(e){var n=e[0],t=i.substr(e.index),r=function e(n,t){if(r=n.eR,a=t,(i=r&&r.exec(a))&&0===i.index){for(;n.endsParent&&n.parent;)n=n.parent;return n}var r,a,i;if(n.eW)return e(n.parent,t)}(E,t);if(r){var a=E;for(a.skip?p+=n:(a.rE||a.eE||(p+=n),o(),a.eE&&(p=n));E.cN&&(R+=C),E.skip||E.sL||(v+=E.relevance),(E=E.parent)!==r.parent;);return r.starts&&(r.endSameAsBegin&&(r.starts.eR=r.eR),l(r.starts)),a.rE?0:n.length}}var f={};function t(e,n){var t=n&&n[0];if(p+=e,null==t)return o(),0;if("begin"==f.type&&"end"==n.type&&f.index==n.index&&""===t)return p+=i.slice(n.index,n.index+1),1;if("begin"===(f=n).type)return u(n);if("illegal"===n.type&&!a)throw new Error('Illegal lexeme "'+t+'" for mode "'+(E.cN||"<unnamed>")+'"');if("end"===n.type){var r=s(n);if(null!=r)return r}return p+=t,t.length}var g=D(n);if(!g)throw console.error(m.replace("{}",n)),new Error('Unknown language: "'+n+'"');x(g);var r,E=e||g,d={},R="";for(r=E;r!==g;r=r.parent)r.cN&&(R=c(r.cN,"",!0)+R);var p="",v=0;try{for(var M,b,h=0;E.t.lastIndex=h,M=E.t.exec(i);)b=t(i.substring(h,M.index),M),h=M.index+b;for(t(i.substr(h)),r=E;r.parent;r=r.parent)r.cN&&(R+=C);return{relevance:v,value:R,i:!1,language:n,top:E}}catch(e){if(e.message&&-1!==e.message.indexOf("Illegal"))return{i:!0,relevance:0,value:B(i)};if(_)return{relevance:0,value:B(i),language:n,top:E,errorRaised:e};throw e}}function T(t,e){e=e||O.languages||o(N);var r={relevance:0,value:B(t)},a=r;return e.filter(D).filter(L).forEach(function(e){var n=S(e,t,!1);n.language=e,n.relevance>a.relevance&&(a=n),n.relevance>r.relevance&&(a=r,r=n)}),a.language&&(r.second_best=a),r}function b(e){return O.tabReplace||O.useBR?e.replace(t,function(e,n){return O.useBR&&"\n"===e?"<br>":O.tabReplace?n.replace(/\t/g,O.tabReplace):""}):e}function s(e){var n,t,r,a,i,o,c,l,u,s,f=function(e){var n,t,r,a,i=e.className+" ";if(i+=e.parentNode?e.parentNode.className:"",t=E.exec(i)){var o=D(t[1]);return o||(console.warn(m.replace("{}",t[1])),console.warn("Falling back to no-highlight mode for this block.",e)),o?t[1]:"no-highlight"}for(n=0,r=(i=i.split(/\s+/)).length;n<r;n++)if(R(a=i[n])||D(a))return a}(e);R(f)||(O.useBR?(n=document.createElement("div")).innerHTML=e.innerHTML.replace(/\n/g,"").replace(/<br[ \/]*>/g,"\n"):n=e,i=n.textContent,r=f?S(f,i,!0):T(i),(t=p(n)).length&&((a=document.createElement("div")).innerHTML=r.value,r.value=v(t,p(a),i)),r.value=b(r.value),e.innerHTML=r.value,e.className=(o=e.className,c=f,l=r.language,u=c?g[c]:l,s=[o.trim()],o.match(/\bhljs\b/)||s.push("hljs"),-1===o.indexOf(u)&&s.push(u),s.join(" ").trim()),e.result={language:r.language,re:r.relevance},r.second_best&&(e.second_best={language:r.second_best.language,re:r.second_best.relevance}))}function h(){if(!h.called){h.called=!0;var e=document.querySelectorAll("pre code");f.forEach.call(e,s)}}var w={disableAutodetect:!0};function D(e){return e=(e||"").toLowerCase(),N[e]||N[g[e]]}function L(e){var n=D(e);return n&&!n.disableAutodetect}return a.highlight=S,a.highlightAuto=T,a.fixMarkup=b,a.highlightBlock=s,a.configure=function(e){O=i(O,e)},a.initHighlighting=h,a.initHighlightingOnLoad=function(){window.addEventListener("DOMContentLoaded",h,!1),window.addEventListener("load",h,!1)},a.registerLanguage=function(n,e){var t;try{t=e(a)}catch(e){if(console.error("Language definition for '{}' could not be registered.".replace("{}",n)),!_)throw e;console.error(e),t=w}u(N[n]=t),t.rawDefinition=e.bind(null,a),t.aliases&&t.aliases.forEach(function(e){g[e]=n})},a.listLanguages=function(){return o(N)},a.getLanguage=D,a.requireLanguage=function(e){var n=D(e);if(n)return n;throw new Error("The '{}' language is required, but not loaded.".replace("{}",e))},a.autoDetection=L,a.inherit=i,a.debugMode=function(){_=!1},a.IR=a.IDENT_RE="[a-zA-Z]\\w*",a.UIR=a.UNDERSCORE_IDENT_RE="[a-zA-Z_]\\w*",a.NR=a.NUMBER_RE="\\b\\d+(\\.\\d+)?",a.CNR=a.C_NUMBER_RE="(-?)(\\b0[xX][a-fA-F0-9]+|(\\b\\d+(\\.\\d*)?|\\.\\d+)([eE][-+]?\\d+)?)",a.BNR=a.BINARY_NUMBER_RE="\\b(0b[01]+)",a.RSR=a.RE_STARTERS_RE="!|!=|!==|%|%=|&|&&|&=|\\*|\\*=|\\+|\\+=|,|-|-=|/=|/|:|;|<<|<<=|<=|<|===|==|=|>>>=|>>=|>=|>>>|>>|>|\\?|\\[|\\{|\\(|\\^|\\^=|\\||\\|=|\\|\\||~",a.BE=a.BACKSLASH_ESCAPE={b:"\\\\[\\s\\S]",relevance:0},a.ASM=a.APOS_STRING_MODE={cN:"string",b:"'",e:"'",i:"\\n",c:[a.BE]},a.QSM=a.QUOTE_STRING_MODE={cN:"string",b:'"',e:'"',i:"\\n",c:[a.BE]},a.PWM=a.PHRASAL_WORDS_MODE={b:/\b(a|an|the|are|I'm|isn't|don't|doesn't|won't|but|just|should|pretty|simply|enough|gonna|going|wtf|so|such|will|you|your|they|like|more)\b/},a.C=a.COMMENT=function(e,n,t){var r=a.inherit({cN:"comment",b:e,e:n,c:[]},t||{});return r.c.push(a.PWM),r.c.push({cN:"doctag",b:"(?:TODO|FIXME|NOTE|BUG|XXX):",relevance:0}),r},a.CLCM=a.C_LINE_COMMENT_MODE=a.C("//","$"),a.CBCM=a.C_BLOCK_COMMENT_MODE=a.C("/\\*","\\*/"),a.HCM=a.HASH_COMMENT_MODE=a.C("#","$"),a.NM=a.NUMBER_MODE={cN:"number",b:a.NR,relevance:0},a.CNM=a.C_NUMBER_MODE={cN:"number",b:a.CNR,relevance:0},a.BNM=a.BINARY_NUMBER_MODE={cN:"number",b:a.BNR,relevance:0},a.CSSNM=a.CSS_NUMBER_MODE={cN:"number",b:a.NR+"(%|em|ex|ch|rem|vw|vh|vmin|vmax|cm|mm|in|pt|pc|px|deg|grad|rad|turn|s|ms|Hz|kHz|dpi|dpcm|dppx)?",relevance:0},a.RM=a.REGEXP_MODE={cN:"regexp",b:/\//,e:/\/[gimuy]*/,i:/\n/,c:[a.BE,{b:/\[/,e:/\]/,relevance:0,c:[a.BE]}]},a.TM=a.TITLE_MODE={cN:"title",b:a.IR,relevance:0},a.UTM=a.UNDERSCORE_TITLE_MODE={cN:"title",b:a.UIR,relevance:0},a.METHOD_GUARD={b:"\\.\\s*"+a.UIR,relevance:0},[a.BE,a.ASM,a.QSM,a.PWM,a.C,a.CLCM,a.CBCM,a.HCM,a.NM,a.CNM,a.BNM,a.CSSNM,a.RM,a.TM,a.UTM,a.METHOD_GUARD].forEach(function(e){!function n(t){Object.freeze(t);var r="function"==typeof t;Object.getOwnPropertyNames(t).forEach(function(e){!t.hasOwnProperty(e)||null===t[e]||"object"!=typeof t[e]&&"function"!=typeof t[e]||r&&("caller"===e||"callee"===e||"arguments"===e)||Object.isFrozen(t[e])||n(t[e])});return t}(e)}),a});hljs.registerLanguage("bash",function(e){var t={cN:"variable",v:[{b:/\$[\w\d#@][\w\d_]*/},{b:/\$\{(.*?)}/}]},a={cN:"string",b:/"/,e:/"/,c:[e.BE,t,{cN:"variable",b:/\$\(/,e:/\)/,c:[e.BE]}]};return{aliases:["sh","zsh"],l:/\b-?[a-z\._]+\b/,k:{keyword:"if then else elif fi for while in do done case esac function",literal:"true false",built_in:"break cd continue eval exec exit export getopts hash pwd readonly return shift test times trap umask unset alias bind builtin caller command declare echo enable help let local logout mapfile printf read readarray source type typeset ulimit unalias set shopt autoload bg bindkey bye cap chdir clone comparguments compcall compctl compdescribe compfiles compgroups compquote comptags comptry compvalues dirs disable disown echotc echoti emulate fc fg float functions getcap getln history integer jobs kill limit log noglob popd print pushd pushln rehash sched setcap setopt stat suspend ttyctl unfunction unhash unlimit unsetopt vared wait whence where which zcompile zformat zftp zle zmodload zparseopts zprof zpty zregexparse zsocket zstyle ztcp",_:"-ne -eq -lt -gt -f -d -e -s -l -a"},c:[{cN:"meta",b:/^#![^\n]+sh\s*$/,relevance:10},{cN:"function",b:/\w[\w\d_]*\s*\(\s*\)\s*\{/,rB:!0,c:[e.inherit(e.TM,{b:/\w[\w\d_]*/})],relevance:0},e.HCM,a,{cN:"",b:/\\"/},{cN:"string",b:/'/,e:/'/},t]}});hljs.registerLanguage("shell",function(s){return{aliases:["console"],c:[{cN:"meta",b:"^\\s{0,3}[/\\w\\d\\[\\]()@-]*[>%$#]",starts:{e:"$",sL:"bash"}}]}});hljs.registerLanguage("ruby",function(e){var c="[a-zA-Z_]\\w*[!?=]?|[-+~]\\@|<<|>>|=~|===?|<=>|[<>]=?|\\*\\*|[-/+%^&*~`|]|\\[\\]=?",b={keyword:"and then defined module in return redo if BEGIN retry end for self when next until do begin unless END rescue else break undef not super class case require yield alias while ensure elsif or include attr_reader attr_writer attr_accessor",literal:"true false nil"},r={cN:"doctag",b:"@[A-Za-z]+"},a={b:"#<",e:">"},n=[e.C("#","$",{c:[r]}),e.C("^\\=begin","^\\=end",{c:[r],relevance:10}),e.C("^__END__","\\n$")],s={cN:"subst",b:"#\\{",e:"}",k:b},t={cN:"string",c:[e.BE,s],v:[{b:/'/,e:/'/},{b:/"/,e:/"/},{b:/`/,e:/`/},{b:"%[qQwWx]?\\(",e:"\\)"},{b:"%[qQwWx]?\\[",e:"\\]"},{b:"%[qQwWx]?{",e:"}"},{b:"%[qQwWx]?<",e:">"},{b:"%[qQwWx]?/",e:"/"},{b:"%[qQwWx]?%",e:"%"},{b:"%[qQwWx]?-",e:"-"},{b:"%[qQwWx]?\\|",e:"\\|"},{b:/\B\?(\\\d{1,3}|\\x[A-Fa-f0-9]{1,2}|\\u[A-Fa-f0-9]{4}|\\?\S)\b/},{b:/<<[-~]?'?(\w+)(?:.|\n)*?\n\s*\1\b/,rB:!0,c:[{b:/<<[-~]?'?/},{b:/\w+/,endSameAsBegin:!0,c:[e.BE,s]}]}]},i={cN:"params",b:"\\(",e:"\\)",endsParent:!0,k:b},l=[t,a,{cN:"class",bK:"class module",e:"$|;",i:/=/,c:[e.inherit(e.TM,{b:"[A-Za-z_]\\w*(::\\w+)*(\\?|\\!)?"}),{b:"<\\s*",c:[{b:"("+e.IR+"::)?"+e.IR}]}].concat(n)},{cN:"function",bK:"def",e:"$|;",c:[e.inherit(e.TM,{b:c}),i].concat(n)},{b:e.IR+"::"},{cN:"symbol",b:e.UIR+"(\\!|\\?)?:",relevance:0},{cN:"symbol",b:":(?!\\s)",c:[t,{b:c}],relevance:0},{cN:"number",b:"(\\b0[0-7_]+)|(\\b0x[0-9a-fA-F_]+)|(\\b[1-9][0-9_]*(\\.[0-9_]+)?)|[0_]\\b",relevance:0},{b:"(\\$\\W)|((\\$|\\@\\@?)(\\w+))"},{cN:"params",b:/\|/,e:/\|/,k:b},{b:"("+e.RSR+"|unless)\\s*",k:"unless",c:[a,{cN:"regexp",c:[e.BE,s],i:/\n/,v:[{b:"/",e:"/[a-z]*"},{b:"%r{",e:"}[a-z]*"},{b:"%r\\(",e:"\\)[a-z]*"},{b:"%r!",e:"![a-z]*"},{b:"%r\\[",e:"\\][a-z]*"}]}].concat(n),relevance:0}].concat(n);s.c=l;var d=[{b:/^\s*=>/,starts:{e:"$",c:i.c=l}},{cN:"meta",b:"^([>?]>|[\\w#]+\\(\\w+\\):\\d+:\\d+>|(\\w+-)?\\d+\\.\\d+\\.\\d(p\\d+)?[^>]+>)",starts:{e:"$",c:l}}];return{aliases:["rb","gemspec","podspec","thor","irb"],k:b,i:/\/\*/,c:n.concat(d).concat(l)}});hljs.registerLanguage("yaml",function(e){var b="true false yes no null",a={cN:"string",relevance:0,v:[{b:/'/,e:/'/},{b:/"/,e:/"/},{b:/\S+/}],c:[e.BE,{cN:"template-variable",v:[{b:",e:"},{b:"%{",e:"}"}]}]};return{cI:!0,aliases:["yml","YAML","yaml"],c:[{cN:"attr",v:[{b:"\\w[\\w :\\/.-]*:(?=[ \t]|$)"},{b:'"\\w[\\w :\\/.-]*":(?=[ \t]|$)'},{b:"'\\w[\\w :\\/.-]*':(?=[ \t]|$)"}]},{cN:"meta",b:"^---s*$",relevance:10},{cN:"string",b:"[\\|>]([0-9]?[+-])?[ ]*\\n( *)[\\S ]+\\n(\\2[\\S ]+\\n?)*"},{b:"<%[%=-]?",e:"[%-]?%>",sL:"ruby",eB:!0,eE:!0,relevance:0},{cN:"type",b:"!"+e.UIR},{cN:"type",b:"!!"+e.UIR},{cN:"meta",b:"&"+e.UIR+"$"},{cN:"meta",b:"\\*"+e.UIR+"$"},{cN:"bullet",b:"\\-(?=[ ]|$)",relevance:0},e.HCM,{bK:b,k:{literal:b}},{cN:"number",b:e.CNR+"\\b"},a]}});

;
// global variables
const doc = document.documentElement;
const inline = ":inline";
// variables read from your hugo configuration
const parentURL = 'https://example.com/';
let showImagePosition = "false";

const showImagePositionLabel = 'Figure';

function isObj(obj) {
  return (obj && typeof obj === 'object' && obj !== null) ? true : false;
}

function createEl(element = 'div') {
  return document.createElement(element);
}

function elem(selector, parent = document){
  let elem = parent.querySelector(selector);
  return elem != false ? elem : false;
}

function elems(selector, parent = document) {
  let elems = parent.querySelectorAll(selector);
  return elems.length ? elems : false;
}

function pushClass(el, targetClass) {
  if (isObj(el) && targetClass) {
    elClass = el.classList;
    elClass.contains(targetClass) ? false : elClass.add(targetClass);
  }
}

function hasClasses(el) {
  if(isObj(el)) {
    const classes = el.classList;
    return classes.length
  }
}

(function markInlineCodeTags(){
  const codeBlocks = elems('code');
  if(codeBlocks) {
    codeBlocks.forEach(function(codeBlock){
      hasClasses(codeBlock) ? false: pushClass(codeBlock, 'noClass');
    });
  }
})();

function deleteClass(el, targetClass) {
  if (isObj(el) && targetClass) {
    elClass = el.classList;
    elClass.contains(targetClass) ? elClass.remove(targetClass) : false;
  }
}

function modifyClass(el, targetClass) {
  if (isObj(el) && targetClass) {
    elClass = el.classList;
    elClass.contains(targetClass) ? elClass.remove(targetClass) : elClass.add(targetClass);
  }
}

function containsClass(el, targetClass) {
  if (isObj(el) && targetClass && el !== document ) {
    return el.classList.contains(targetClass) ? true : false;
  }
}

function elemAttribute(elem, attr, value = null) {
  if (value) {
    elem.setAttribute(attr, value);
  } else {
    value = elem.getAttribute(attr);
    return value ? value : false;
  }
}

function wrapEl(el, wrapper) {
  el.parentNode.insertBefore(wrapper, el);
  wrapper.appendChild(el);
}

function deleteChars(str, subs) {
  let newStr = str;
  if (Array.isArray(subs)) {
    for (let i = 0; i < subs.length; i++) {
      newStr = newStr.replace(subs[i], '');
    }
  } else {
    newStr = newStr.replace(subs, '');
  }
  return newStr;
}

function isBlank(str) {
  return (!str || str.trim().length === 0);
}

function isMatch(element, selectors) {
  if(isObj(element)) {
    if(selectors.isArray) {
      let matching = selectors.map(function(selector){
        return element.matches(selector)
      })
      return matching.includes(true);
    }
    return element.matches(selectors)
  }
}

function copyToClipboard(str) {
  let copy, selection, selected;
  copy = createEl('textarea');
  copy.value = str;
  copy.setAttribute('readonly', '');
  copy.style.position = 'absolute';
  copy.style.left = '-9999px';
  selection = document.getSelection();
  doc.appendChild(copy);
  // check if there is any selected content
  selected = selection.rangeCount > 0 ? selection.getRangeAt(0) : false;
  copy.select();
  document.execCommand('copy');
  doc.removeChild(copy);
  if (selected) { // if a selection existed before copying
    selection.removeAllRanges(); // unselect existing selection
    selection.addRange(selected); // restore the original selection
  }
}

function loadSvg(file, parent, path = 'icons/') {
  const link = `${parentURL}${path}${file}.svg`;
  fetch(link)
  .then((response) => {
    return response.text();
  })
  .then((data) => {
    parent.innerHTML = data;
  });
}

function getMobileOperatingSystem() {
  let userAgent = navigator.userAgent || navigator.vendor || window.opera;
  
  if (/android/i.test(userAgent)) {
    return "Android";
  }
  
  if (/iPad|iPhone|iPod/.test(userAgent) && !window.MSStream) {
    return "iOS";
  }
  
  return "unknown";
}

function horizontalSwipe(element, func, direction) {
  // call func if result of swipeDirection() 👇🏻 is equal to direction
  
  let touchstartX = 0;
  let touchendX = 0;
  let swipeDirection = null;

  function handleGesure() {
    return (touchendX + 50 < touchstartX) ? 'left' : (touchendX < touchstartX + 50) ? 'right' : false;
  }

  element.addEventListener('touchstart', e => {
    touchstartX = e.changedTouches[0].screenX
  });

  element.addEventListener('touchend', e => {
    touchendX = e.changedTouches[0].screenX
    swipeDirection = handleGesure()
    swipeDirection === direction ? func() : false;
  });

}

function parseBoolean(string) {
  let bool;
  string = string.trim().toLowerCase();
  switch (string) {
    case 'true':
      return true;
    case 'false':
      return false;
    default:
      return undefined;
  }
};

(function() {
  const bodyElement = elem('body');
  const platform = navigator.platform.toLowerCase();
  if(platform.includes("win")) {
    pushClass(bodyElement, 'windows');
  }
})();
;
const codeActionButtons = [
  {
    icon: 'copy', 
    id: 'copy',
    title: 'Copy Code',
    show: true
  },
  {
    icon: 'order',
    id: 'lines',
    title: 'Toggle Line Numbers',
    show: true 
  },
  {
    icon: 'carly',
    id: 'wrap',
    title: 'Toggle Line Wrap',
    show: false
  },
  {
    icon: 'expand',
    id: 'expand',
    title: 'Toggle code block expand',
    show: false 
  }
];

const body = elem('body');
const maxLines = parseInt(body.dataset.code);
const copyId = 'panel_copy';
const wrapId = 'panel_wrap';
const linesId = 'panel_lines';
const panelExpand = 'panel_expand';
const panelExpanded = 'panel_expanded';
const panelHide = 'panel_hide';
const panelFrom = 'panel_from';
const panelBox = 'panel_box';
const fullHeight = '100vh';
const highlightWrap = 'highlight_wrap'

function codeBlocks() {
  const markedCodeBlocks = elems('code');
  const blocks = Array.from(markedCodeBlocks).filter(function(block){
    return hasClasses(block) && !Array.from(block.classList).includes('noClass');
  }).map(function(block){
    return block
  });
  return blocks;
}

function codeBlockFits(block) {
  // return false if codeblock overflows
  const blockWidth = block.offsetWidth;
  const highlightBlockWidth = block.parentNode.parentNode.offsetWidth;
  return blockWidth <= highlightBlockWidth ? true : false;
}

function maxHeightIsSet(elem) {
  let maxHeight = elem.style.maxHeight;
  return maxHeight.includes('px')
}

function restrainCodeBlockHeight(lines) {
  const lastLine = lines[maxLines-1];
  let maxCodeBlockHeight = fullHeight;
  if(lastLine) {
    const lastLinePos = lastLine.offsetTop;
    if(lastLinePos !== 0) {
      maxCodeBlockHeight = `${lastLinePos}px`;
      const codeBlock = lines[0].parentNode;
      const outerBlock = codeBlock.closest('.highlight');
      const isExpanded = containsClass(outerBlock, panelExpanded);
      if(!isExpanded) {
        codeBlock.dataset.height = maxCodeBlockHeight;
        codeBlock.style.maxHeight = maxCodeBlockHeight;
      }
    }
  }
}

const blocks = codeBlocks();

function collapseCodeBlock(block) {
  const lines = elems('.ln', block);
  const codeLines = lines.length;
  if (codeLines > maxLines) {
    const expandDot = createEl()
    pushClass(expandDot, panelExpand);
    pushClass(expandDot, panelFrom);
    expandDot.title = "Toggle code block expand";
    expandDot.textContent = "...";
    const outerBlock = block.closest('.highlight');
    window.setTimeout(function(){
      const expandIcon = outerBlock.nextElementSibling.lastElementChild;
      deleteClass(expandIcon, panelHide);
    }, 150)

    restrainCodeBlockHeight(lines);
    const highlightElement = block.parentNode.parentNode;
    highlightElement.appendChild(expandDot);
  }
}

blocks.forEach(function(block){
  collapseCodeBlock(block);
})

function actionPanel() {
  const panel = createEl();
  panel.className = panelBox;

  codeActionButtons.forEach(function(button) {
    // create button
    const btn = createEl('a');
    btn.href = '#';
    btn.title = button.title;
    btn.className = `icon panel_icon panel_${button.id}`;
    button.show ? false : pushClass(btn, panelHide);
    // load icon inside button
    loadSvg(button.icon, btn);
    // append button on panel
    panel.appendChild(btn);
  });

  return panel;
}

function toggleLineNumbers(elems) {
  elems.forEach(function (elem, index) {
    // mark the code element when there are no lines
    modifyClass(elem, 'pre_nolines')
  });
  restrainCodeBlockHeight(elems);
}

function toggleLineWrap(elem) {
  modifyClass(elem, 'pre_wrap');
  // retain max number of code lines on line wrap
  const lines = elems('.ln', elem);
  restrainCodeBlockHeight(lines);
}

function copyCode(codeElement) {
  lineNumbers = elems('.ln', codeElement);
  // remove line numbers before copying
  if(lineNumbers.length) {
    lineNumbers.forEach(function(line){
      line.remove();
    });
  }

  const codeToCopy = codeElement.textContent;
  // copy code
  copyToClipboard(codeToCopy);
}

function disableCodeLineNumbers(block){
  const lines = elems('.ln', block)
  toggleLineNumbers(lines);
}

(function codeActions(){
  const blocks = codeBlocks();

  const highlightWrapId = highlightWrap;
  blocks.forEach(function(block){
    // disable line numbers if disabled globally
    const showLines = elem('body').dataset.lines;
    parseBoolean(showLines) === false ? disableCodeLineNumbers(block) : false;

    const highlightElement = block.parentNode.parentNode;
    // wrap code block in a div
    const highlightWrapper = createEl();
    highlightWrapper.className = highlightWrapId;
    wrapEl(highlightElement, highlightWrapper);

    const panel = actionPanel();
    // show wrap icon only if the code block needs wrapping
    const wrapIcon = elem(`.${wrapId}`, panel);
    codeBlockFits(block) ? false : deleteClass(wrapIcon, panelHide);

    // append buttons 
    highlightWrapper.appendChild(panel);
  });

  function isItem(target, id) {
    // if is item or within item
    return target.matches(`.${id}`) || target.closest(`.${id}`);
  }

  function showActive(target, targetClass,activeClass = 'active') {
    const active = activeClass;
    const targetElement = target.matches(`.${targetClass}`) ? target : target.closest(`.${targetClass}`);

    deleteClass(targetElement, active);
    setTimeout(function() {
      modifyClass(targetElement, active)
    }, 50)
  }

  doc.addEventListener('click', function(event){
    // copy code block
    const target = event.target;
    const isCopyIcon = isItem(target, copyId);
    const isWrapIcon = isItem(target, wrapId);
    const isLinesIcon = isItem(target, linesId);
    const isExpandIcon = isItem(target, panelExpand);
    const isActionable = isCopyIcon || isWrapIcon || isLinesIcon || isExpandIcon;

    if(isActionable) {
      event.preventDefault();
      showActive(target, 'icon');
      const codeElement = target.closest(`.${highlightWrapId}`).firstElementChild.firstElementChild;
      let lineNumbers = elems('.ln', codeElement);

      isWrapIcon ? toggleLineWrap(codeElement) : false;

      isLinesIcon ? toggleLineNumbers(lineNumbers) : false;

      if (isExpandIcon) {
        let thisCodeBlock = codeElement.firstElementChild;
        const outerBlock = thisCodeBlock.closest('.highlight');
        if(maxHeightIsSet(thisCodeBlock)) {
          thisCodeBlock.style.maxHeight = `100vh`;
          // mark code block as expanded
          pushClass(outerBlock, panelExpanded)
        } else {
          thisCodeBlock.style.maxHeight = thisCodeBlock.dataset.height;
          // unmark code block as expanded
          deleteClass(outerBlock, panelExpanded)
        }
      }

      if(isCopyIcon) {
        // clone code element
        const codeElementClone = codeElement.cloneNode(true);
        copyCode(codeElementClone);
      }
    }
  });

  (function addLangLabel() {
    const blocks = codeBlocks();
    blocks.forEach(function(block){
      let label = block.dataset.lang;
      label = label === 'sh' ? 'bash' : label;
      if(label !== "fallback") {
        const labelEl = createEl();
        labelEl.textContent = label;
        pushClass(labelEl, 'lang');
        block.closest(`.${highlightWrap}`).appendChild(labelEl);
      }
    });
  })();
})();

;
(function toggleColorModes(){
  const light = 'lit';
  const dark = 'dim';
  const storageKey = 'colorMode';
  const key = '--color-mode';
  const data = 'data-mode';
  const bank = window.localStorage;
  
  function currentMode() {
    let acceptableChars = light + dark;
    acceptableChars = [...acceptableChars];
    let mode = getComputedStyle(doc).getPropertyValue(key).replace(/\"/g, '').trim();
    
    mode = [...mode].filter(function(letter){
      return acceptableChars.includes(letter);
    });
    
    return mode.join('');
  }
  
  function changeMode(isDarkMode) {
    if(isDarkMode) {
      bank.setItem(storageKey, light)
      elemAttribute(doc, data, light);
    } else {
      bank.setItem(storageKey, dark);
      elemAttribute(doc, data, dark);
    }
  }
  
  function setUserColorMode(mode = false) {
    const isDarkMode = currentMode() == dark;
    const storedMode = bank.getItem(storageKey);
    if(storedMode) {
      if(mode) {
        changeMode(isDarkMode);
      } else {
        elemAttribute(doc, data, storedMode);
      }
    } else {
      if(mode === true) {
        changeMode(isDarkMode) 
      }
    }
  }
  
  setUserColorMode();
  
  doc.addEventListener('click', function(event) {
    let target = event.target;
    let modeClass = 'color_choice';
    let animateClass = 'color_animate';
    let isModeToggle = containsClass(target, modeClass);
    if(isModeToggle) {
      pushClass(target, animateClass);
      setUserColorMode(true);        
    }
  });
})();

function fileClosure(){ 
  // everything in this file should be declared within this closure (function).
  
  (function updateDate() {
    var date = new Date();
    var year = date.getFullYear();
    elem('.year').innerHTML = year;
  })();
  
  (function makeExternalLinks(){
    let links = elems('a');
    if(links) {
      Array.from(links).forEach(function(link){
        let target, rel, blank, noopener, attr1, attr2, url, isExternal;
        url = elemAttribute(link, 'href');
        isExternal = (url && typeof url == 'string' && url.startsWith('http')) && !url.startsWith(parentURL) ? true : false;
        if(isExternal) {
          target = 'target';
          rel = 'rel';
          blank = '_blank';
          noopener = 'noopener';
          attr1 = elemAttribute(link, target);
          attr2 = elemAttribute(link, noopener);
          
          attr1 ? false : elemAttribute(link, target, blank);
          attr2 ? false : elemAttribute(link, rel, noopener);
        }
      });
    }
  })();
  
  let headingNodes = [], results, link, icon, current, id,
  tags = ['h2', 'h3', 'h4', 'h5', 'h6'];
  
  current = document.URL;
  
  tags.forEach(function(tag){
    const article = elem('.post_content');
    if (article) {
      results = article.getElementsByTagName(tag);
      Array.prototype.push.apply(headingNodes, results);
    }
  });
  
  headingNodes.forEach(function(node){
    link = createEl('a');
    loadSvg('link', link);
    link.className = 'link icon';
    id = node.getAttribute('id');
    if(id) {
      link.href = `${current}#${id}`;
      node.appendChild(link);
      pushClass(node, 'link_owner');
    }
  });
  
  let inlineListItems = elems('ol li');
  if(inlineListItems) {
    inlineListItems.forEach(function(listItem){
      let firstChild = listItem.children[0]
      let containsHeading = isMatch(firstChild, tags);
      containsHeading ? pushClass(listItem, 'align') : false;
    })
  }
  
  function copyFeedback(parent) {
    const copyText = document.createElement('div');
    const yanked = 'link_yanked';
    copyText.classList.add(yanked);
    copyText.innerText = 'Link Copied';
    if(!elem(`.${yanked}`, parent)) {
      parent.appendChild(copyText);
      setTimeout(function() { 
        parent.removeChild(copyText)
      }, 3000);
    }
  }
  
  (function copyHeadingLink() {
    let deeplink, deeplinks, newLink, parent, target;
    deeplink = 'link';
    deeplinks = elems(`.${deeplink}`);
    if(deeplinks) {
      document.addEventListener('click', function(event)
      {
        target = event.target;
        parent = target.parentNode;
        if (target && containsClass(target, deeplink) || containsClass(parent, deeplink)) {
          event.preventDefault();
          newLink = target.href != undefined ? target.href : target.parentNode.href;
          copyToClipboard(newLink);
          target.href != undefined ?  copyFeedback(target) : copyFeedback(target.parentNode);
        }
      });
    }
  })();

  (function copyLinkToShare() {
    let  copy, copied, excerpt, isCopyIcon, isInExcerpt, link, postCopy, postLink, target;
    copy = 'copy';
    copied = 'copy_done';
    excerpt = 'excerpt';
    postCopy = 'post_copy';
    postLink = 'post_card';
    
    doc.addEventListener('click', function(event) {
      target = event.target;
      isCopyIcon = containsClass(target, copy);
      let isWithinCopyIcon = target.closest(`.${copy}`);
      if (isCopyIcon || isWithinCopyIcon) {
        let icon = isCopyIcon ? isCopyIcon : isWithinCopyIcon;
        isInExcerpt =  containsClass(icon, postCopy);
        if (isInExcerpt) {
          link = target.closest(`.${excerpt}`).previousElementSibling;
          link = containsClass(link, postLink)? elemAttribute(link, 'href') : false;
        } else {
          link = window.location.href;
        }
        if(link) {
          copyToClipboard(link);
          pushClass(icon, copied);
        }
      }
      const yankLink = '.link_yank';
      const isCopyLink = target.matches(yankLink);
      const isCopyLinkIcon = target.closest(yankLink);
      
      if(isCopyLink || isCopyLinkIcon) {
        event.preventDefault();
        const yankContent = isCopyLinkIcon ? elemAttribute(target.closest(yankLink), 'href') : elemAttribute(target, 'href');
        copyToClipboard(yankContent);
        isCopyLink ?  copyFeedback(target) : copyFeedback(target.parentNode);
      }
      
    });
  })();
  
  (function hideAside(){
    let aside, title, posts;
    aside = elem('.aside');
    title = aside ? aside.previousElementSibling : null;
    if(aside && title.nodeName.toLowerCase() === 'h3') {
      posts = Array.from(aside.children);
      posts.length < 1 ? title.remove() : false;
    }
  })();
  
  (function goBack() {
    let backBtn = elem('.btn_back');
    let history = window.history;
    if (backBtn) {
      backBtn.addEventListener('click', function(){
        history.back();
      });
    }
  })();
  
  function showingImagePosition(){
    // whether or not to track image position for non-linear images within the article body element.
    const thisPage = document.documentElement;
    let showImagePositionOnPage = thisPage.dataset.figures;
    
    if(showImagePositionOnPage) {
      showImagePosition = showImagePositionOnPage;
    }
    return showImagePosition === "true" ? true : false;
  }
  
  function populateAlt(images) {
    let imagePosition = 0;
    images.forEach((image) => {
      let alt = image.alt;
      image.loading = "lazy";
      const modifiers = [':left', ':right'];
      
      modifiers.forEach(function(modifier){
        const canModify = alt.includes(modifier);
        if(canModify) {
          pushClass(image, `float_${modifier.replace(":", "")}`);
          
          alt = alt.replace(modifier, "");
        }
      });
      
      const isInline = alt.includes(inline);
      alt = alt.replace(inline, "");
      
      // wait for position to load and a caption if the image is not online and has an alt attribute
      if (alt.length > 0 && !containsClass(image, 'alt' && !isInline)) {
        imagePosition += 1;
        image.dataset.pos = imagePosition;
        image.addEventListener('load', function() {
          const showImagePosition = showingImagePosition();
          
          let desc = document.createElement('p');
          desc.classList.add('img_alt');
          let imageAlt = image.alt;
          
          const thisImgPos = image.dataset.pos;
          // modify image caption is necessary
          imageAlt = showImagePosition ? `${showImagePositionLabel} ${thisImgPos}: ${imageAlt}` : imageAlt;
          desc.textContent = imageAlt;
          image.insertAdjacentHTML('afterend', desc.outerHTML);
        })
      }
      
      if(isInline) {
        modifyClass(image, 'inline');
      }
      
    });
    
    hljs.initHighlightingOnLoad();
  }
  
  function largeImages(baseParent, images = []) {
    if(images) {
      images.forEach(function(image) {
        
        image.addEventListener('load', function(){
          
          let actualWidth = image.naturalWidth;
          
          let parentWidth = baseParent.offsetWidth;
          
          let actionableRatio = actualWidth / parentWidth;
          
          if (actionableRatio > 1) {
            pushClass(image, "image-scalable");
            image.dataset.scale = actionableRatio;
            let figure = createEl('figure');
            
            wrapEl(image, figure)
          }
          
        });
      })
    }
  }
  
  (function AltImage() {
    let post = elem('.post_content');
    let images = post ? post.querySelectorAll('img') : false;
    images ? populateAlt(images) : false;
    largeImages(post, images);
  })();
  
  
  doc.addEventListener('click', function(event) {
    let target = event.target;
    isClickableImage = target.matches('.image-scalable');
    
    let isFigure = target.matches('figure');
    
    if(isFigure) {
      let hasClickableImage = containsClass(target.children[0], 'image-scalable');
      if(hasClickableImage) {
        modifyClass(target, 'image-scale');
      }
    }
    
    if(isClickableImage) {
      let figure = target.parentNode;
      modifyClass(figure, 'image-scale');
    }
    
  });
  
  const tables = elems('table');
  if (tables) {
    const scrollable = 'scrollable';
    tables.forEach(function(table) {
      const wrapper = createEl();
      wrapper.className = scrollable;
      wrapEl(table, wrapper);
    });
  }

  // track if there's an expanded tags' widget
  let tagWidgetOpen = false;
    
  function toggleTags(target = null) {
    const tagsButtonClass = 'post_tags_toggle';
    const tagsButtonClass2 = 'tags_hide';
    const tagsShowClass = 'jswidgetopen';
    const postTagsWrapper = elem(`.${tagsShowClass}`);
    target = target === null ? postTagsWrapper : target;
    const showingAllTags = target.matches(`.${tagsShowClass}`);
    const isExandButton = target.matches(`.${tagsButtonClass}`);
    const isCloseButton = target.matches(`.${tagsButtonClass2}`) || target.closest(`.${tagsButtonClass2}`);
    const isButton =  isExandButton || isCloseButton;
    const isActionable = isButton || showingAllTags;
    
    if(isActionable) {
      if(isButton) {
        if(isExandButton) {
          let allTagsWrapper = target.nextElementSibling 
          pushClass(allTagsWrapper, tagsShowClass); 
        } else {
          deleteClass(postTagsWrapper, tagsShowClass);
        }
      } else {
        isActionable ? deleteClass(target, tagsShowClass) : false;
      }
    }
  }
  
  (function showAllPostTags(){
    doc.addEventListener('click', function(event){
      const target = event.target;
      toggleTags(target)
    });
    
    horizontalSwipe(doc, toggleTags, 'left');
  })();
  
  (function navToggle() {
    doc.addEventListener('click', function(event){
      const target = event.target;
      const open = 'jsopen';
      const isNavToggle = target.matches('.nav_close') || target.closest('.nav_close');
      const targetIcon = elem('.nav_close_icon');
      const harmburgerIcon = 'https://example.com/icons/bar.svg';
      const closeIcon = 'https://example.com/icons/cancel.svg';
      if(isNavToggle) {
        event.preventDefault();
        modifyClass(doc, open);
        let navIsOpen = containsClass(doc, open);
        targetIcon.src = navIsOpen  ? closeIcon : harmburgerIcon;
      }
      
      if(!target.closest('.nav') && elem(`.${open}`)) {
        modifyClass(doc, open);
        let navIsOpen = containsClass(doc, open);
        targetIcon.src = navIsOpen  ? closeIcon : harmburgerIcon;
      }
      
      const navItem = 'nav_item';
      const navSub = 'nav_sub';
      const showSub = 'nav_open';
      const isNavItem = target.matches(`.${navItem}`);
      const isNavItemIcon = target.closest(`.${navItem}`)
      
      if(isNavItem || isNavItemIcon) {
        const thisItem = isNavItem ? target : isNavItemIcon;
        const hasNext = thisItem.nextElementSibling
        const hasSubNav = hasNext ? hasNext.matches(`.${navSub}`) : null;
        if (hasSubNav) {
          event.preventDefault();
          modifyClass(thisItem, showSub);
        } 
      }
      
    });
  })();

  function isMobileDevice() {
    const agent = navigator.userAgent.toLowerCase();
    const isMobile = agent.includes('android') || agent.includes('iphone');
    return  isMobile;
  };

  (function ifiOS(){
    // modify backto top button
    const backToTopButton = elem('.to_top');
    const thisOS = getMobileOperatingSystem();
    const ios = 'ios';
    if(backToTopButton && thisOS === 'iOS') {
      pushClass(backToTopButton, ios);
    }
    // precisely position back to top button on large screens
    const buttonParentWidth = backToTopButton.parentNode.offsetWidth;
    const docWidth = doc.offsetWidth;
    let leftOffset = (docWidth - buttonParentWidth) / 2;
    const buttonWidth = backToTopButton.offsetWidth;
    leftOffset = leftOffset + buttonParentWidth - buttonWidth;
    if(!isMobileDevice()){
      backToTopButton.style.left = `${leftOffset}px`;
    } 
  })();

  (function sortTags() {
    doc.addEventListener('click', function(event){
      const active = 'active';
      const target = event.target;
      const isSortButton = target.matches('.tags_sort') || target.matches('.tags_sort span');
      if(isSortButton) {
        const tagsList = target.closest('.tags_list');
        const sortButton = elem('.tags_sort', tagsList);
        modifyClass(sortButton, 'sorted');
        const tags = elems('.post_tag', tagsList);
        Array.from(tags).forEach(function(tag){
          const order = tag.dataset.position;
          const reverseSorting = containsClass(tag, active);
          tag.style.order = reverseSorting ? 0 : -order;
          modifyClass(tag, active);
        })
      }
    })
  })();

  // add new code above this line
}

window.addEventListener('load', fileClosure());
