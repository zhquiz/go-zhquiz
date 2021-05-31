(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["vocab"],{"1e17":function(t,e,n){},"5c02":function(t,e,n){"use strict";n("1e17")},c36c:function(t,e,n){"use strict";n.r(e);var u=function(){var t=this,e=t.$createElement,n=t._self._c||e;return n("section",[n("div",{staticClass:"VocabPage"},[n("form",{staticClass:"field",on:{submit:function(e){e.preventDefault(),t.q=t.q0}}},[n("div",{staticClass:"control"},[n("input",{directives:[{name:"model",rawName:"v-model",value:t.q0,expression:"q0"}],staticClass:"input",attrs:{type:"search",name:"q",placeholder:"Type here to search.","aria-label":"search"},domProps:{value:t.q0},on:{input:function(e){e.target.composing||(t.q0=e.target.value)}}})])]),n("div",{staticClass:"columns"},[n("div",{staticClass:"column is-6 entry-display"},[n("div",{class:t.simplified.length>3?"smaller-vocab-display":"vocab-display"},[n("div",{staticClass:"clickable text-center font-zh-simp",on:{contextmenu:function(e){return e.preventDefault(),function(e){return t.openContext(e,t.simplified,"vocab")}(e)}}},[t._v(" "+t._s(t.simplified)+" ")])]),n("div",{staticClass:"buttons has-addons"},[n("button",{staticClass:"button",attrs:{disabled:t.i<1},on:{click:function(e){t.i--},keypress:function(e){t.i--}}},[t._v(" Previous ")]),n("button",{staticClass:"button",attrs:{disabled:t.i>t.entries.length-2},on:{click:function(e){t.i++},keypress:function(e){t.i++}}},[t._v(" Next ")])])]),n("div",{staticClass:"column is-6"},[n("b-collapse",{staticClass:"card",staticStyle:{"margin-bottom":"1em"},attrs:{animation:"slide",open:"object"===typeof t.current},scopedSlots:t._u([{key:"trigger",fn:function(e){return n("div",{staticClass:"card-header",attrs:{role:"button"}},[n("h2",{staticClass:"card-header-title"},[t._v("Reading")]),n("a",{staticClass:"card-header-icon",attrs:{role:"button"}},[n("fontawesome",{attrs:{icon:e.open?"caret-down":"caret-up"}})],1)])}}])},[n("div",{staticClass:"card-content"},[n("span",[t._v(t._s(t.current.pinyin))])])]),n("b-collapse",{staticClass:"card",attrs:{animation:"slide",open:!!t.current.traditional},scopedSlots:t._u([{key:"trigger",fn:function(e){return n("div",{staticClass:"card-header",attrs:{role:"button"}},[n("h2",{staticClass:"card-header-title"},[t._v("Traditional")]),n("a",{staticClass:"card-header-icon",attrs:{role:"button"}},[n("fontawesome",{attrs:{icon:e.open?"caret-down":"caret-up"}})],1)])}}])},[n("div",{staticClass:"card-content"},[n("div",{staticClass:"font-zh-trad clickable",on:{contextmenu:function(e){return e.preventDefault(),function(e){return t.openContext(e,t.current.traditional,"vocab")}(e)}}},[t._v(" "+t._s(t.current.traditional)+" ")])])]),n("b-collapse",{staticClass:"card",attrs:{animation:"slide",open:!!t.current.english},scopedSlots:t._u([{key:"trigger",fn:function(e){return n("div",{staticClass:"card-header",attrs:{role:"button"}},[n("h2",{staticClass:"card-header-title"},[t._v("English")]),n("a",{staticClass:"card-header-icon",attrs:{role:"button"}},[n("fontawesome",{attrs:{icon:e.open?"caret-down":"caret-up"}})],1)])}}])},[n("div",{staticClass:"card-content"},[n("span",[t._v(t._s(t.current.english))])])]),n("b-collapse",{key:t.sentenceKey,staticClass:"card",attrs:{animation:"slide",open:!!t.sentences().length},scopedSlots:t._u([{key:"trigger",fn:function(e){return n("div",{staticClass:"card-header",attrs:{role:"button"}},[n("h2",{staticClass:"card-header-title"},[t._v("Sentences")]),n("a",{staticClass:"card-header-icon",attrs:{role:"button"}},[n("fontawesome",{attrs:{icon:e.open?"caret-down":"caret-up"}})],1)])}}])},[n("div",{staticClass:"card-content"},t._l(t.sentences(),(function(e,u){return n("div",{key:u,staticClass:"sentence-entry"},[n("span",{staticClass:"clickable",on:{contextmenu:function(n){return n.preventDefault(),function(n){return t.openContext(n,e.chinese,"sentence")}(n)}}},[t._v(" "+t._s(e.chinese)+" ")]),n("span",[t._v(t._s(e.english))])])})),0)])],1)])]),n("ContextMenu",{ref:"context",attrs:{entry:t.selected.entry,type:t.selected.type,additional:t.additionalContext,pinyin:t.sentenceDef.pinyin,english:t.sentenceDef.english}})],1)},s=[],a=(n("99af"),n("4de4"),n("7db0"),n("c975"),n("fb6a"),n("d3b7"),n("ddb0"),n("ade3")),i=n("2909"),r=(n("96cf"),n("1da1")),c=n("d4ec"),o=n("bee2"),D=n("262e"),l=n("2caf"),d=n("9ab4"),F=n("1b40"),h=n("5962"),p=n.n(h),f=n("6825"),C=n("02ef"),b=n("d8bb"),v=function(t){Object(D["a"])(n,t);var e=Object(l["a"])(n);function n(){var t;return Object(c["a"])(this,n),t=e.apply(this,arguments),t.entries=[],t.i=0,t.selected={entry:"",type:""},t.q0="",t.sentenceKey=0,t}return Object(o["a"])(n,[{key:"sentences",value:function(){return b["c"].find({chinese:{$containsString:"string"===typeof this.current?this.current:this.current.simplified}}).slice(0,10)}},{key:"created",value:function(){var t=Object(r["a"])(regeneratorRuntime.mark((function t(){var e,n,u,s,a,i;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:if(e=this.$route.query.entry,e||this.q){t.next=7;break}return t.next=4,C["a"].get("/api/vocab/random",{params:{levelMin:this.$accessor.levelMin,level:this.$accessor.level}});case 4:n=t.sent,u=n.data.result,e=u;case 7:if(this.q0=e||this.q,!e){t.next=16;break}return s=window,a=s.frameElement,a&&(i=parseInt(a.getAttribute("data-id")||""),window.parent.setName(i,(e?e+" - ":"")+"Vocab")),this.entries=[e],t.next=14,this.loadContent();case 14:t.next=18;break;case 16:return t.next=18,this.onQChange(this.q0);case 18:case"end":return t.stop()}}),t,this)})));function e(){return t.apply(this,arguments)}return e}()},{key:"openContext",value:function(t){var e=arguments.length>1&&void 0!==arguments[1]?arguments[1]:this.selected.entry,n=arguments.length>2&&void 0!==arguments[2]?arguments[2]:this.selected.type;this.selected={entry:e,type:n},this.context.open(t)}},{key:"onQChange",value:function(){var t=Object(r["a"])(regeneratorRuntime.mark((function t(e){var n,u,s,a;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:if(n=window,u=n.frameElement,u&&(s=parseInt(u.getAttribute("data-id")||""),window.parent.setName(s,(e?e+" - ":"")+"Vocab")),!/(?:[\u2E80-\u2E99\u2E9B-\u2EF3\u2F00-\u2FD5\u3005\u3007\u3021-\u3029\u3038-\u303B\u3400-\u4DBF\u4E00-\u9FFC\uF900-\uFA6D\uFA70-\uFAD9]|\uD81B[\uDFF0\uDFF1]|[\uD840-\uD868\uD86A-\uD86C\uD86F-\uD872\uD874-\uD879\uD880-\uD883][\uDC00-\uDFFF]|\uD869[\uDC00-\uDEDD\uDF00-\uDFFF]|\uD86D[\uDC00-\uDF34\uDF40-\uDFFF]|\uD86E[\uDC00-\uDC1D\uDC20-\uDFFF]|\uD873[\uDC00-\uDEA1\uDEB0-\uDFFF]|\uD87A[\uDC00-\uDFE0]|\uD87E[\uDC00-\uDE1D]|\uD884[\uDC00-\uDF4A])+/.test(e)){t.next=10;break}return t.next=5,C["a"].get("/api/chinese/jieba",{params:{q:e}}).then((function(t){return t.data.result}));case 5:a=t.sent,a=a.filter((function(t){return/(?:[\u2E80-\u2E99\u2E9B-\u2EF3\u2F00-\u2FD5\u3005\u3007\u3021-\u3029\u3038-\u303B\u3400-\u4DBF\u4E00-\u9FFC\uF900-\uFA6D\uFA70-\uFAD9]|\uD81B[\uDFF0\uDFF1]|[\uD840-\uD868\uD86A-\uD86C\uD86F-\uD872\uD874-\uD879\uD880-\uD883][\uDC00-\uDFFF]|\uD869[\uDC00-\uDEDD\uDF00-\uDFFF]|\uD86D[\uDC00-\uDF34\uDF40-\uDFFF]|\uD86E[\uDC00-\uDC1D\uDC20-\uDFFF]|\uD873[\uDC00-\uDEA1\uDEB0-\uDFFF]|\uD87A[\uDC00-\uDFE0]|\uD87E[\uDC00-\uDE1D]|\uD884[\uDC00-\uDF4A])+/.test(t)})).filter((function(t,e,n){return n.indexOf(t)===e})),this.entries=a,t.next=11;break;case 10:this.entries=[e];case 11:return t.next=13,this.loadContent();case 13:this.i=0;case 14:case"end":return t.stop()}}),t,this)})));function e(e){return t.apply(this,arguments)}return e}()},{key:"loadContent",value:function(){var t=Object(r["a"])(regeneratorRuntime.mark((function t(){var e,n,u,s,a;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:if(e=this.entries[this.i],e){t.next=3;break}return t.abrupt("return");case 3:if("string"!==typeof e){t.next=18;break}if(!/(?:[\u2E80-\u2E99\u2E9B-\u2EF3\u2F00-\u2FD5\u3005\u3007\u3021-\u3029\u3038-\u303B\u3400-\u4DBF\u4E00-\u9FFC\uF900-\uFA6D\uFA70-\uFAD9]|\uD81B[\uDFF0\uDFF1]|[\uD840-\uD868\uD86A-\uD86C\uD86F-\uD872\uD874-\uD879\uD880-\uD883][\uDC00-\uDFFF]|\uD869[\uDC00-\uDEDD\uDF00-\uDFFF]|\uD86D[\uDC00-\uDF34\uDF40-\uDFFF]|\uD86E[\uDC00-\uDC1D\uDC20-\uDFFF]|\uD873[\uDC00-\uDEA1\uDEB0-\uDFFF]|\uD87A[\uDC00-\uDFE0]|\uD87E[\uDC00-\uDE1D]|\uD884[\uDC00-\uDF4A])/.test(e)){t.next=12;break}return t.next=7,C["a"].get("/api/vocab",{params:{entry:e}});case 7:n=t.sent,u=n.data.result,u.length>0?(e=u[0].simplified,this.entries=[].concat(Object(i["a"])(this.entries.slice(0,this.i)),Object(i["a"])(u),Object(i["a"])(this.entries.slice(this.i+1)))):this.entries=[].concat(Object(i["a"])(this.entries.slice(0,this.i)),[{simplified:e,pinyin:p()(e,{keepRest:!0,toneToNumber:!0})}],Object(i["a"])(this.entries.slice(this.i+1))),t.next=18;break;case 12:return t.next=14,C["a"].get("/api/vocab/q",{params:{q:e}});case 14:s=t.sent,a=s.data.result,e=a.length>0?a[0].simplified:"",this.entries=[].concat(Object(i["a"])(this.entries.slice(0,this.i)),Object(i["a"])(a),Object(i["a"])(this.entries.slice(this.i+1)));case 18:if(e){t.next=20;break}return t.abrupt("return");case 20:return t.next=22,Object(b["a"])("string"===typeof e?e:e.simplified,10);case 22:if(!t.sent){t.next=24;break}this.sentenceKey=Math.random();case 24:case"end":return t.stop()}}),t,this)})));function e(){return t.apply(this,arguments)}return e}()},{key:"sentenceDef",get:function(){if("sentence"!==this.selected.type)return{};var t=b["c"].findOne({chinese:this.selected.entry});return t?{pinyin:Object(a["a"])({},this.selected.entry,t.pinyin),english:Object(a["a"])({},this.selected.entry,t.english)}:{}}},{key:"q",get:function(){var t=this.$route.query.q;return(Array.isArray(t)?t[0]:t)||""},set:function(t){this.$router.push({query:{q:t}})}},{key:"current",get:function(){var t=this.entries[this.i];return"string"===typeof t?/(?:[\u2E80-\u2E99\u2E9B-\u2EF3\u2F00-\u2FD5\u3005\u3007\u3021-\u3029\u3038-\u303B\u3400-\u4DBF\u4E00-\u9FFC\uF900-\uFA6D\uFA70-\uFAD9]|\uD81B[\uDFF0\uDFF1]|[\uD840-\uD868\uD86A-\uD86C\uD86F-\uD872\uD874-\uD879\uD880-\uD883][\uDC00-\uDFFF]|\uD869[\uDC00-\uDEDD\uDF00-\uDFFF]|\uD86D[\uDC00-\uDF34\uDF40-\uDFFF]|\uD86E[\uDC00-\uDC1D\uDC20-\uDFFF]|\uD873[\uDC00-\uDEA1\uDEB0-\uDFFF]|\uD87A[\uDC00-\uDFE0]|\uD87E[\uDC00-\uDE1D]|\uD884[\uDC00-\uDF4A])/.test(t)?t:"":t||""}},{key:"simplified",get:function(){return"string"===typeof this.current?this.current:this.current.simplified}},{key:"additionalContext",get:function(){var t=this;return this.q?[]:[{name:"Reload",handler:function(){var e=Object(r["a"])(regeneratorRuntime.mark((function e(){var n,u;return regeneratorRuntime.wrap((function(e){while(1)switch(e.prev=e.next){case 0:return e.next=2,C["a"].get("/api/vocab/random");case 2:n=e.sent,u=n.data.result,t.q0=u;case 5:case"end":return e.stop()}}),e)})));function n(){return e.apply(this,arguments)}return n}()}]}}]),n}(F["d"]);Object(d["a"])([Object(F["c"])()],v.prototype,"context",void 0),v=Object(d["a"])([Object(F["a"])({components:{ContextMenu:f["a"]},watch:{q:function(){this.onQChange(this.q)},current:function(){this.loadContent()}}})],v);var y=v,m=y,g=(n("5c02"),n("2877")),E=Object(g["a"])(m,u,s,!1,null,"19b820a5",null);e["default"]=E.exports}}]);
//# sourceMappingURL=vocab.96f5802b.js.map