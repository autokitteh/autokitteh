import{e as F}from"./react-BA7A5rMr.js";(function(){try{var e=typeof window<"u"?window:typeof global<"u"?global:typeof self<"u"?self:{},t=new e.Error().stack;t&&(e._sentryDebugIds=e._sentryDebugIds||{},e._sentryDebugIds[t]="3ccd2aea-87de-4b4c-8cc1-5d0706fe0115",e._sentryDebugIdIdentifier="sentry-dbid-3ccd2aea-87de-4b4c-8cc1-5d0706fe0115")}catch{}})();var _e=e=>e.type==="checkbox",ne=e=>e instanceof Date,W=e=>e==null;const fr=e=>typeof e=="object";var T=e=>!W(e)&&!Array.isArray(e)&&fr(e)&&!ne(e),dr=e=>T(e)&&e.target?_e(e.target)?e.target.checked:e.target.value:e,Lr=e=>e.substring(0,e.search(/\.\d+(\.|$)/))||e,yr=(e,t)=>e.has(Lr(t)),Rr=e=>{const t=e.constructor&&e.constructor.prototype;return T(t)&&t.hasOwnProperty("isPrototypeOf")},He=typeof window<"u"&&typeof window.HTMLElement<"u"&&typeof document<"u";function I(e){let t;const r=Array.isArray(e);if(e instanceof Date)t=new Date(e);else if(e instanceof Set)t=new Set(e);else if(!(He&&(e instanceof Blob||e instanceof FileList))&&(r||T(e)))if(t=r?[]:{},!r&&!Rr(e))t=e;else for(const i in e)e.hasOwnProperty(i)&&(t[i]=I(e[i]));else return e;return t}var ve=e=>Array.isArray(e)?e.filter(Boolean):[],E=e=>e===void 0,c=(e,t,r)=>{if(!t||!T(e))return r;const i=ve(t.split(/[,[\].]+?/)).reduce((u,n)=>W(u)?u:u[n],e);return E(i)||i===e?E(e[t])?r:e[t]:i},G=e=>typeof e=="boolean",$e=e=>/^\w*$/.test(e),gr=e=>ve(e.replace(/["|']|\]/g,"").split(/\.|\[/)),k=(e,t,r)=>{let i=-1;const u=$e(t)?[t]:gr(t),n=u.length,y=n-1;for(;++i<n;){const _=u[i];let m=r;if(i!==y){const O=e[_];m=T(O)||Array.isArray(O)?O:isNaN(+u[i+1])?{}:[]}if(_==="__proto__")return;e[_]=m,e=e[_]}return e};const Ae={BLUR:"blur",FOCUS_OUT:"focusout",CHANGE:"change"},Y={onBlur:"onBlur",onChange:"onChange",onSubmit:"onSubmit",onTouched:"onTouched",all:"all"},re={max:"max",min:"min",maxLength:"maxLength",minLength:"minLength",pattern:"pattern",required:"required",validate:"validate"},_r=F.createContext(null),we=()=>F.useContext(_r),rt=e=>{const{children:t,...r}=e;return F.createElement(_r.Provider,{value:r},t)};var vr=(e,t,r,i=!0)=>{const u={defaultValues:t._defaultValues};for(const n in e)Object.defineProperty(u,n,{get:()=>{const y=n;return t._proxyFormState[y]!==Y.all&&(t._proxyFormState[y]=!i||Y.all),r&&(r[y]=!0),e[y]}});return u},j=e=>T(e)&&!Object.keys(e).length,hr=(e,t,r,i)=>{r(e);const{name:u,...n}=e;return j(n)||Object.keys(n).length>=Object.keys(t).length||Object.keys(n).find(y=>t[y]===(!i||Y.all))},q=e=>Array.isArray(e)?e:[e],br=(e,t,r)=>!e||!t||e===t||q(e).some(i=>i&&(r?i===t:i.startsWith(t)||t.startsWith(i)));function De(e){const t=F.useRef(e);t.current=e,F.useEffect(()=>{const r=!e.disabled&&t.current.subject&&t.current.subject.subscribe({next:t.current.next});return()=>{r&&r.unsubscribe()}},[e.disabled])}function Ir(e){const t=we(),{control:r=t.control,disabled:i,name:u,exact:n}=e||{},[y,_]=F.useState(r._formState),m=F.useRef(!0),O=F.useRef({isDirty:!1,isLoading:!1,dirtyFields:!1,touchedFields:!1,validatingFields:!1,isValidating:!1,isValid:!1,errors:!1}),A=F.useRef(u);return A.current=u,De({disabled:i,next:b=>m.current&&br(A.current,b.name,n)&&hr(b,O.current,r._updateFormState)&&_({...r._formState,...b}),subject:r._subjects.state}),F.useEffect(()=>(m.current=!0,O.current.isValid&&r._updateValid(!0),()=>{m.current=!1}),[r]),vr(y,r,O.current,!1)}var ee=e=>typeof e=="string",Fr=(e,t,r,i,u)=>ee(e)?(i&&t.watch.add(e),c(r,e,u)):Array.isArray(e)?e.map(n=>(i&&t.watch.add(n),c(r,n))):(i&&(t.watchAll=!0),r);function Mr(e){const t=we(),{control:r=t.control,name:i,defaultValue:u,disabled:n,exact:y}=e||{},_=F.useRef(i);_.current=i,De({disabled:n,subject:r._subjects.values,next:A=>{br(_.current,A.name,y)&&O(I(Fr(_.current,r._names,A.values||r._formValues,!1,u)))}});const[m,O]=F.useState(r._getWatch(i,u));return F.useEffect(()=>r._removeUnmounted()),m}function Br(e){const t=we(),{name:r,disabled:i,control:u=t.control,shouldUnregister:n}=e,y=yr(u._names.array,r),_=Mr({control:u,name:r,defaultValue:c(u._formValues,r,c(u._defaultValues,r,e.defaultValue)),exact:!0}),m=Ir({control:u,name:r,exact:!0}),O=F.useRef(u.register(r,{...e.rules,value:_,...G(e.disabled)?{disabled:e.disabled}:{}}));return F.useEffect(()=>{const A=u._options.shouldUnregister||n,b=(S,H)=>{const L=c(u._fields,S);L&&L._f&&(L._f.mount=H)};if(b(r,!0),A){const S=I(c(u._options.defaultValues,r));k(u._defaultValues,r,S),E(c(u._formValues,r))&&k(u._formValues,r,S)}return()=>{(y?A&&!u._state.action:A)?u.unregister(r):b(r,!1)}},[r,u,y,n]),F.useEffect(()=>{c(u._fields,r)&&u._updateDisabledField({disabled:i,fields:u._fields,name:r,value:c(u._fields,r)._f.value})},[i,r,u]),{field:{name:r,value:_,...G(i)||m.disabled?{disabled:m.disabled||i}:{},onChange:F.useCallback(A=>O.current.onChange({target:{value:dr(A),name:r},type:Ae.CHANGE}),[r]),onBlur:F.useCallback(()=>O.current.onBlur({target:{value:c(u._formValues,r),name:r},type:Ae.BLUR}),[r,u]),ref:F.useCallback(A=>{const b=c(u._fields,r);b&&A&&(b._f.ref={focus:()=>A.focus(),select:()=>A.select(),setCustomValidity:S=>A.setCustomValidity(S),reportValidity:()=>A.reportValidity()})},[u._fields,r])},formState:m,fieldState:Object.defineProperties({},{invalid:{enumerable:!0,get:()=>!!c(m.errors,r)},isDirty:{enumerable:!0,get:()=>!!c(m.dirtyFields,r)},isTouched:{enumerable:!0,get:()=>!!c(m.touchedFields,r)},isValidating:{enumerable:!0,get:()=>!!c(m.validatingFields,r)},error:{enumerable:!0,get:()=>c(m.errors,r)}})}}const tt=e=>e.render(Br(e));var Nr=(e,t,r,i,u)=>t?{...r[e],types:{...r[e]&&r[e].types?r[e].types:{},[i]:u||!0}}:{},se=()=>{const e=typeof performance>"u"?Date.now():performance.now()*1e3;return"xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g,t=>{const r=(Math.random()*16+e)%16|0;return(t=="x"?r:r&3|8).toString(16)})},Oe=(e,t,r={})=>r.shouldFocus||E(r.shouldFocus)?r.focusName||`${e}.${E(r.focusIndex)?t:r.focusIndex}.`:"",ge=e=>({isOnSubmit:!e||e===Y.onSubmit,isOnBlur:e===Y.onBlur,isOnChange:e===Y.onChange,isOnAll:e===Y.all,isOnTouch:e===Y.onTouched}),We=(e,t,r)=>!r&&(t.watchAll||t.watch.has(e)||[...t.watch].some(i=>e.startsWith(i)&&/^\.\w+/.test(e.slice(i.length))));const le=(e,t,r,i)=>{for(const u of r||Object.keys(e)){const n=c(e,u);if(n){const{_f:y,..._}=n;if(y){if(y.refs&&y.refs[0]&&t(y.refs[0],u)&&!i)return!0;if(y.ref&&t(y.ref,y.name)&&!i)return!0;if(le(_,t))break}else if(T(_)&&le(_,t))break}}};var Ar=(e,t,r)=>{const i=q(c(e,r));return k(i,"root",t[r]),k(e,r,i),e},Ke=e=>e.type==="file",te=e=>typeof e=="function",me=e=>{if(!He)return!1;const t=e?e.ownerDocument:0;return e instanceof(t&&t.defaultView?t.defaultView.HTMLElement:HTMLElement)},Fe=e=>ee(e),Ge=e=>e.type==="radio",Ve=e=>e instanceof RegExp;const ir={value:!1,isValid:!1},ar={value:!0,isValid:!0};var mr=e=>{if(Array.isArray(e)){if(e.length>1){const t=e.filter(r=>r&&r.checked&&!r.disabled).map(r=>r.value);return{value:t,isValid:!!t.length}}return e[0].checked&&!e[0].disabled?e[0].attributes&&!E(e[0].attributes.value)?E(e[0].value)||e[0].value===""?ar:{value:e[0].value,isValid:!0}:ar:ir}return ir};const ur={isValid:!1,value:null};var Vr=e=>Array.isArray(e)?e.reduce((t,r)=>r&&r.checked&&!r.disabled?{isValid:!0,value:r.value}:t,ur):ur;function nr(e,t,r="validate"){if(Fe(e)||Array.isArray(e)&&e.every(Fe)||G(e)&&!e)return{type:r,message:Fe(e)?e:"",ref:t}}var ue=e=>T(e)&&!Ve(e)?e:{value:e,message:""},qe=async(e,t,r,i,u)=>{const{ref:n,refs:y,required:_,maxLength:m,minLength:O,min:A,max:b,pattern:S,validate:H,name:L,valueAsNumber:oe,mount:z,disabled:Z}=e._f,p=c(t,L);if(!z||Z)return{};const J=y?y[0]:n,Q=x=>{i&&J.reportValidity&&(J.setCustomValidity(G(x)?"":x||""),J.reportValidity())},g={},h=Ge(n),V=_e(n),C=h||V,$=(oe||Ke(n))&&E(n.value)&&E(p)||me(n)&&n.value===""||p===""||Array.isArray(p)&&!p.length,P=Nr.bind(null,L,r,g),he=(x,w,R,N=re.maxLength,X=re.minLength)=>{const K=x?w:R;g[L]={type:x?N:X,message:K,ref:n,...P(x?N:X,K)}};if(u?!Array.isArray(p)||!p.length:_&&(!C&&($||W(p))||G(p)&&!p||V&&!mr(y).isValid||h&&!Vr(y).isValid)){const{value:x,message:w}=Fe(_)?{value:!!_,message:_}:ue(_);if(x&&(g[L]={type:re.required,message:w,ref:J,...P(re.required,w)},!r))return Q(w),g}if(!$&&(!W(A)||!W(b))){let x,w;const R=ue(b),N=ue(A);if(!W(p)&&!isNaN(p)){const X=n.valueAsNumber||p&&+p;W(R.value)||(x=X>R.value),W(N.value)||(w=X<N.value)}else{const X=n.valueAsDate||new Date(p),K=de=>new Date(new Date().toDateString()+" "+de),ce=n.type=="time",fe=n.type=="week";ee(R.value)&&p&&(x=ce?K(p)>K(R.value):fe?p>R.value:X>new Date(R.value)),ee(N.value)&&p&&(w=ce?K(p)<K(N.value):fe?p<N.value:X<new Date(N.value))}if((x||w)&&(he(!!x,R.message,N.message,re.max,re.min),!r))return Q(g[L].message),g}if((m||O)&&!$&&(ee(p)||u&&Array.isArray(p))){const x=ue(m),w=ue(O),R=!W(x.value)&&p.length>+x.value,N=!W(w.value)&&p.length<+w.value;if((R||N)&&(he(R,x.message,w.message),!r))return Q(g[L].message),g}if(S&&!$&&ee(p)){const{value:x,message:w}=ue(S);if(Ve(x)&&!p.match(x)&&(g[L]={type:re.pattern,message:w,ref:n,...P(re.pattern,w)},!r))return Q(w),g}if(H){if(te(H)){const x=await H(p,t),w=nr(x,J);if(w&&(g[L]={...w,...P(re.validate,w.message)},!r))return Q(w.message),g}else if(T(H)){let x={};for(const w in H){if(!j(x)&&!r)break;const R=nr(await H[w](p,t),J,w);R&&(x={...R,...P(w,R.message)},Q(R.message),r&&(g[L]=x))}if(!j(x)&&(g[L]={ref:J,...x},!r))return g}}return Q(!0),g},Ue=(e,t)=>[...e,...q(t)],Te=e=>Array.isArray(e)?e.map(()=>{}):void 0;function Le(e,t,r){return[...e.slice(0,t),...q(r),...e.slice(t)]}var Re=(e,t,r)=>Array.isArray(e)?(E(e[r])&&(e[r]=void 0),e.splice(r,0,e.splice(t,1)[0]),e):[],Ie=(e,t)=>[...q(t),...q(e)];function Pr(e,t){let r=0;const i=[...e];for(const u of t)i.splice(u-r,1),r++;return ve(i).length?i:[]}var Me=(e,t)=>E(t)?[]:Pr(e,q(t).sort((r,i)=>r-i)),Be=(e,t,r)=>{[e[t],e[r]]=[e[r],e[t]]};function jr(e,t){const r=t.slice(0,-1).length;let i=0;for(;i<r;)e=E(e)?i++:e[t[i++]];return e}function Wr(e){for(const t in e)if(e.hasOwnProperty(t)&&!E(e[t]))return!1;return!0}function U(e,t){const r=Array.isArray(t)?t:$e(t)?[t]:gr(t),i=r.length===1?e:jr(e,r),u=r.length-1,n=r[u];return i&&delete i[n],u!==0&&(T(i)&&j(i)||Array.isArray(i)&&Wr(i))&&U(e,r.slice(0,-1)),e}var lr=(e,t,r)=>(e[t]=r,e);function st(e){const t=we(),{control:r=t.control,name:i,keyName:u="id",shouldUnregister:n}=e,[y,_]=F.useState(r._getFieldArray(i)),m=F.useRef(r._getFieldArray(i).map(se)),O=F.useRef(y),A=F.useRef(i),b=F.useRef(!1);A.current=i,O.current=y,r._names.array.add(i),e.rules&&r.register(i,e.rules),De({next:({values:g,name:h})=>{if(h===A.current||!h){const V=c(g,A.current);Array.isArray(V)&&(_(V),m.current=V.map(se))}},subject:r._subjects.array});const S=F.useCallback(g=>{b.current=!0,r._updateFieldArray(i,g)},[r,i]),H=(g,h)=>{const V=q(I(g)),C=Ue(r._getFieldArray(i),V);r._names.focus=Oe(i,C.length-1,h),m.current=Ue(m.current,V.map(se)),S(C),_(C),r._updateFieldArray(i,C,Ue,{argA:Te(g)})},L=(g,h)=>{const V=q(I(g)),C=Ie(r._getFieldArray(i),V);r._names.focus=Oe(i,0,h),m.current=Ie(m.current,V.map(se)),S(C),_(C),r._updateFieldArray(i,C,Ie,{argA:Te(g)})},oe=g=>{const h=Me(r._getFieldArray(i),g);m.current=Me(m.current,g),S(h),_(h),r._updateFieldArray(i,h,Me,{argA:g})},z=(g,h,V)=>{const C=q(I(h)),$=Le(r._getFieldArray(i),g,C);r._names.focus=Oe(i,g,V),m.current=Le(m.current,g,C.map(se)),S($),_($),r._updateFieldArray(i,$,Le,{argA:g,argB:Te(h)})},Z=(g,h)=>{const V=r._getFieldArray(i);Be(V,g,h),Be(m.current,g,h),S(V),_(V),r._updateFieldArray(i,V,Be,{argA:g,argB:h},!1)},p=(g,h)=>{const V=r._getFieldArray(i);Re(V,g,h),Re(m.current,g,h),S(V),_(V),r._updateFieldArray(i,V,Re,{argA:g,argB:h},!1)},J=(g,h)=>{const V=I(h),C=lr(r._getFieldArray(i),g,V);m.current=[...C].map(($,P)=>!$||P===g?se():m.current[P]),S(C),_([...C]),r._updateFieldArray(i,C,lr,{argA:g,argB:V},!0,!1)},Q=g=>{const h=q(I(g));m.current=h.map(se),S([...h]),_([...h]),r._updateFieldArray(i,[...h],V=>V,{},!0,!1)};return F.useEffect(()=>{if(r._state.action=!1,We(i,r._names)&&r._subjects.state.next({...r._formState}),b.current&&(!ge(r._options.mode).isOnSubmit||r._formState.isSubmitted))if(r._options.resolver)r._executeSchema([i]).then(g=>{const h=c(g.errors,i),V=c(r._formState.errors,i);(V?!h&&V.type||h&&(V.type!==h.type||V.message!==h.message):h&&h.type)&&(h?k(r._formState.errors,i,h):U(r._formState.errors,i),r._subjects.state.next({errors:r._formState.errors}))});else{const g=c(r._fields,i);g&&g._f&&!(ge(r._options.reValidateMode).isOnSubmit&&ge(r._options.mode).isOnSubmit)&&qe(g,r._formValues,r._options.criteriaMode===Y.all,r._options.shouldUseNativeValidation,!0).then(h=>!j(h)&&r._subjects.state.next({errors:Ar(r._formState.errors,h,i)}))}r._subjects.values.next({name:i,values:{...r._formValues}}),r._names.focus&&le(r._fields,(g,h)=>{if(r._names.focus&&h.startsWith(r._names.focus)&&g.focus)return g.focus(),1}),r._names.focus="",r._updateValid(),b.current=!1},[y,i,r]),F.useEffect(()=>(!c(r._formValues,i)&&r._updateFieldArray(i),()=>{(r._options.shouldUnregister||n)&&r.unregister(i)}),[i,r,u,n]),{swap:F.useCallback(Z,[S,i,r]),move:F.useCallback(p,[S,i,r]),prepend:F.useCallback(L,[S,i,r]),append:F.useCallback(H,[S,i,r]),remove:F.useCallback(oe,[S,i,r]),insert:F.useCallback(z,[S,i,r]),update:F.useCallback(J,[S,i,r]),replace:F.useCallback(Q,[S,i,r]),fields:F.useMemo(()=>y.map((g,h)=>({...g,[u]:m.current[h]||se()})),[y,u])}}var Ne=()=>{let e=[];return{get observers(){return e},next:u=>{for(const n of e)n.next&&n.next(u)},subscribe:u=>(e.push(u),{unsubscribe:()=>{e=e.filter(n=>n!==u)}}),unsubscribe:()=>{e=[]}}},xe=e=>W(e)||!fr(e);function ie(e,t){if(xe(e)||xe(t))return e===t;if(ne(e)&&ne(t))return e.getTime()===t.getTime();const r=Object.keys(e),i=Object.keys(t);if(r.length!==i.length)return!1;for(const u of r){const n=e[u];if(!i.includes(u))return!1;if(u!=="ref"){const y=t[u];if(ne(n)&&ne(y)||T(n)&&T(y)||Array.isArray(n)&&Array.isArray(y)?!ie(n,y):n!==y)return!1}}return!0}var xr=e=>e.type==="select-multiple",qr=e=>Ge(e)||_e(e),Pe=e=>me(e)&&e.isConnected,pr=e=>{for(const t in e)if(te(e[t]))return!0;return!1};function pe(e,t={}){const r=Array.isArray(e);if(T(e)||r)for(const i in e)Array.isArray(e[i])||T(e[i])&&!pr(e[i])?(t[i]=Array.isArray(e[i])?[]:{},pe(e[i],t[i])):W(e[i])||(t[i]=!0);return t}function wr(e,t,r){const i=Array.isArray(e);if(T(e)||i)for(const u in e)Array.isArray(e[u])||T(e[u])&&!pr(e[u])?E(t)||xe(r[u])?r[u]=Array.isArray(e[u])?pe(e[u],[]):{...pe(e[u])}:wr(e[u],W(t)?{}:t[u],r[u]):r[u]=!ie(e[u],t[u]);return r}var be=(e,t)=>wr(e,t,pe(t)),Dr=(e,{valueAsNumber:t,valueAsDate:r,setValueAs:i})=>E(e)?e:t?e===""?NaN:e&&+e:r&&ee(e)?new Date(e):i?i(e):e;function je(e){const t=e.ref;if(!(e.refs?e.refs.every(r=>r.disabled):t.disabled))return Ke(t)?t.files:Ge(t)?Vr(e.refs).value:xr(t)?[...t.selectedOptions].map(({value:r})=>r):_e(t)?mr(e.refs).value:Dr(E(t.value)?e.ref.value:t.value,e)}var Hr=(e,t,r,i)=>{const u={};for(const n of e){const y=c(t,n);y&&k(u,n,y._f)}return{criteriaMode:r,names:[...e],fields:u,shouldUseNativeValidation:i}},ye=e=>E(e)?e:Ve(e)?e.source:T(e)?Ve(e.value)?e.value.source:e.value:e;const or="AsyncFunction";var $r=e=>(!e||!e.validate)&&!!(te(e.validate)&&e.validate.constructor.name===or||T(e.validate)&&Object.values(e.validate).find(t=>t.constructor.name===or)),Kr=e=>e.mount&&(e.required||e.min||e.max||e.maxLength||e.minLength||e.pattern||e.validate);function cr(e,t,r){const i=c(e,r);if(i||$e(r))return{error:i,name:r};const u=r.split(".");for(;u.length;){const n=u.join("."),y=c(t,n),_=c(e,n);if(y&&!Array.isArray(y)&&r!==n)return{name:r};if(_&&_.type)return{name:n,error:_};u.pop()}return{name:r}}var Gr=(e,t,r,i,u)=>u.isOnAll?!1:!r&&u.isOnTouch?!(t||e):(r?i.isOnBlur:u.isOnBlur)?!e:(r?i.isOnChange:u.isOnChange)?e:!0,Yr=(e,t)=>!ve(c(e,t)).length&&U(e,t);const zr={mode:Y.onSubmit,reValidateMode:Y.onChange,shouldFocusError:!0};function Jr(e={}){let t={...zr,...e},r={submitCount:0,isDirty:!1,isLoading:te(t.defaultValues),isValidating:!1,isSubmitted:!1,isSubmitting:!1,isSubmitSuccessful:!1,isValid:!1,touchedFields:{},dirtyFields:{},validatingFields:{},errors:t.errors||{},disabled:t.disabled||!1},i={},u=T(t.defaultValues)||T(t.values)?I(t.defaultValues||t.values)||{}:{},n=t.shouldUnregister?{}:I(u),y={action:!1,mount:!1,watch:!1},_={mount:new Set,unMount:new Set,array:new Set,watch:new Set},m,O=0;const A={isDirty:!1,dirtyFields:!1,validatingFields:!1,touchedFields:!1,isValidating:!1,isValid:!1,errors:!1},b={values:Ne(),array:Ne(),state:Ne()},S=ge(t.mode),H=ge(t.reValidateMode),L=t.criteriaMode===Y.all,oe=s=>a=>{clearTimeout(O),O=setTimeout(s,a)},z=async s=>{if(A.isValid||s){const a=t.resolver?j((await C()).errors):await P(i,!0);a!==r.isValid&&b.state.next({isValid:a})}},Z=(s,a)=>{(A.isValidating||A.validatingFields)&&((s||Array.from(_.mount)).forEach(l=>{l&&(a?k(r.validatingFields,l,a):U(r.validatingFields,l))}),b.state.next({validatingFields:r.validatingFields,isValidating:!j(r.validatingFields)}))},p=(s,a=[],l,d,f=!0,o=!0)=>{if(d&&l){if(y.action=!0,o&&Array.isArray(c(i,s))){const v=l(c(i,s),d.argA,d.argB);f&&k(i,s,v)}if(o&&Array.isArray(c(r.errors,s))){const v=l(c(r.errors,s),d.argA,d.argB);f&&k(r.errors,s,v),Yr(r.errors,s)}if(A.touchedFields&&o&&Array.isArray(c(r.touchedFields,s))){const v=l(c(r.touchedFields,s),d.argA,d.argB);f&&k(r.touchedFields,s,v)}A.dirtyFields&&(r.dirtyFields=be(u,n)),b.state.next({name:s,isDirty:x(s,a),dirtyFields:r.dirtyFields,errors:r.errors,isValid:r.isValid})}else k(n,s,a)},J=(s,a)=>{k(r.errors,s,a),b.state.next({errors:r.errors})},Q=s=>{r.errors=s,b.state.next({errors:r.errors,isValid:!1})},g=(s,a,l,d)=>{const f=c(i,s);if(f){const o=c(n,s,E(l)?c(u,s):l);E(o)||d&&d.defaultChecked||a?k(n,s,a?o:je(f._f)):N(s,o),y.mount&&z()}},h=(s,a,l,d,f)=>{let o=!1,v=!1;const D={name:s},M=!!(c(i,s)&&c(i,s)._f&&c(i,s)._f.disabled);if(!l||d){A.isDirty&&(v=r.isDirty,r.isDirty=D.isDirty=x(),o=v!==D.isDirty);const B=M||ie(c(u,s),a);v=!!(!M&&c(r.dirtyFields,s)),B||M?U(r.dirtyFields,s):k(r.dirtyFields,s,!0),D.dirtyFields=r.dirtyFields,o=o||A.dirtyFields&&v!==!B}if(l){const B=c(r.touchedFields,s);B||(k(r.touchedFields,s,l),D.touchedFields=r.touchedFields,o=o||A.touchedFields&&B!==l)}return o&&f&&b.state.next(D),o?D:{}},V=(s,a,l,d)=>{const f=c(r.errors,s),o=A.isValid&&G(a)&&r.isValid!==a;if(e.delayError&&l?(m=oe(()=>J(s,l)),m(e.delayError)):(clearTimeout(O),m=null,l?k(r.errors,s,l):U(r.errors,s)),(l?!ie(f,l):f)||!j(d)||o){const v={...d,...o&&G(a)?{isValid:a}:{},errors:r.errors,name:s};r={...r,...v},b.state.next(v)}},C=async s=>{Z(s,!0);const a=await t.resolver(n,t.context,Hr(s||_.mount,i,t.criteriaMode,t.shouldUseNativeValidation));return Z(s),a},$=async s=>{const{errors:a}=await C(s);if(s)for(const l of s){const d=c(a,l);d?k(r.errors,l,d):U(r.errors,l)}else r.errors=a;return a},P=async(s,a,l={valid:!0})=>{for(const d in s){const f=s[d];if(f){const{_f:o,...v}=f;if(o){const D=_.array.has(o.name),M=f._f&&$r(f._f);M&&A.validatingFields&&Z([d],!0);const B=await qe(f,n,L,t.shouldUseNativeValidation&&!a,D);if(M&&A.validatingFields&&Z([d]),B[o.name]&&(l.valid=!1,a))break;!a&&(c(B,o.name)?D?Ar(r.errors,B,o.name):k(r.errors,o.name,B[o.name]):U(r.errors,o.name))}!j(v)&&await P(v,a,l)}}return l.valid},he=()=>{for(const s of _.unMount){const a=c(i,s);a&&(a._f.refs?a._f.refs.every(l=>!Pe(l)):!Pe(a._f.ref))&&Se(s)}_.unMount=new Set},x=(s,a)=>(s&&a&&k(n,s,a),!ie(Ye(),u)),w=(s,a,l)=>Fr(s,_,{...y.mount?n:E(a)?u:ee(s)?{[s]:a}:a},l,a),R=s=>ve(c(y.mount?n:u,s,e.shouldUnregister?c(u,s,[]):[])),N=(s,a,l={})=>{const d=c(i,s);let f=a;if(d){const o=d._f;o&&(!o.disabled&&k(n,s,Dr(a,o)),f=me(o.ref)&&W(a)?"":a,xr(o.ref)?[...o.ref.options].forEach(v=>v.selected=f.includes(v.value)):o.refs?_e(o.ref)?o.refs.length>1?o.refs.forEach(v=>(!v.defaultChecked||!v.disabled)&&(v.checked=Array.isArray(f)?!!f.find(D=>D===v.value):f===v.value)):o.refs[0]&&(o.refs[0].checked=!!f):o.refs.forEach(v=>v.checked=v.value===f):Ke(o.ref)?o.ref.value="":(o.ref.value=f,o.ref.type||b.values.next({name:s,values:{...n}})))}(l.shouldDirty||l.shouldTouch)&&h(s,f,l.shouldTouch,l.shouldDirty,!0),l.shouldValidate&&de(s)},X=(s,a,l)=>{for(const d in a){const f=a[d],o=`${s}.${d}`,v=c(i,o);(_.array.has(s)||!xe(f)||v&&!v._f)&&!ne(f)?X(o,f,l):N(o,f,l)}},K=(s,a,l={})=>{const d=c(i,s),f=_.array.has(s),o=I(a);k(n,s,o),f?(b.array.next({name:s,values:{...n}}),(A.isDirty||A.dirtyFields)&&l.shouldDirty&&b.state.next({name:s,dirtyFields:be(u,n),isDirty:x(s,o)})):d&&!d._f&&!W(o)?X(s,o,l):N(s,o,l),We(s,_)&&b.state.next({...r}),b.values.next({name:y.mount?s:void 0,values:{...n}})},ce=async s=>{y.mount=!0;const a=s.target;let l=a.name,d=!0;const f=c(i,l),o=()=>a.type?je(f._f):dr(s),v=D=>{d=Number.isNaN(D)||ie(D,c(n,l,D))};if(f){let D,M;const B=o(),ae=s.type===Ae.BLUR||s.type===Ae.FOCUS_OUT,Or=!Kr(f._f)&&!t.resolver&&!c(r.errors,l)&&!f._f.deps||Gr(ae,c(r.touchedFields,l),r.isSubmitted,H,S),Ee=We(l,_,ae);k(n,l,B),ae?(f._f.onBlur&&f._f.onBlur(s),m&&m(0)):f._f.onChange&&f._f.onChange(s);const Ce=h(l,B,ae,!1),Ur=!j(Ce)||Ee;if(!ae&&b.values.next({name:l,type:s.type,values:{...n}}),Or)return A.isValid&&(e.mode==="onBlur"?ae&&z():z()),Ur&&b.state.next({name:l,...Ee?{}:Ce});if(!ae&&Ee&&b.state.next({...r}),t.resolver){const{errors:tr}=await C([l]);if(v(B),d){const Tr=cr(r.errors,i,l),sr=cr(tr,i,Tr.name||l);D=sr.error,l=sr.name,M=j(tr)}}else Z([l],!0),D=(await qe(f,n,L,t.shouldUseNativeValidation))[l],Z([l]),v(B),d&&(D?M=!1:A.isValid&&(M=await P(i,!0)));d&&(f._f.deps&&de(f._f.deps),V(l,M,D,Ce))}},fe=(s,a)=>{if(c(r.errors,a)&&s.focus)return s.focus(),1},de=async(s,a={})=>{let l,d;const f=q(s);if(t.resolver){const o=await $(E(s)?s:f);l=j(o),d=s?!f.some(v=>c(o,v)):l}else s?(d=(await Promise.all(f.map(async o=>{const v=c(i,o);return await P(v&&v._f?{[o]:v}:v)}))).every(Boolean),!(!d&&!r.isValid)&&z()):d=l=await P(i);return b.state.next({...!ee(s)||A.isValid&&l!==r.isValid?{}:{name:s},...t.resolver||!s?{isValid:l}:{},errors:r.errors}),a.shouldFocus&&!d&&le(i,fe,s?f:_.mount),d},Ye=s=>{const a={...y.mount?n:u};return E(s)?a:ee(s)?c(a,s):s.map(l=>c(a,l))},ze=(s,a)=>({invalid:!!c((a||r).errors,s),isDirty:!!c((a||r).dirtyFields,s),error:c((a||r).errors,s),isValidating:!!c(r.validatingFields,s),isTouched:!!c((a||r).touchedFields,s)}),Sr=s=>{s&&q(s).forEach(a=>U(r.errors,a)),b.state.next({errors:s?r.errors:{}})},Je=(s,a,l)=>{const d=(c(i,s,{_f:{}})._f||{}).ref,f=c(r.errors,s)||{},{ref:o,message:v,type:D,...M}=f;k(r.errors,s,{...M,...a,ref:d}),b.state.next({name:s,errors:r.errors,isValid:!1}),l&&l.shouldFocus&&d&&d.focus&&d.focus()},kr=(s,a)=>te(s)?b.values.subscribe({next:l=>s(w(void 0,a),l)}):w(s,a,!0),Se=(s,a={})=>{for(const l of s?q(s):_.mount)_.mount.delete(l),_.array.delete(l),a.keepValue||(U(i,l),U(n,l)),!a.keepError&&U(r.errors,l),!a.keepDirty&&U(r.dirtyFields,l),!a.keepTouched&&U(r.touchedFields,l),!a.keepIsValidating&&U(r.validatingFields,l),!t.shouldUnregister&&!a.keepDefaultValue&&U(u,l);b.values.next({values:{...n}}),b.state.next({...r,...a.keepDirty?{isDirty:x()}:{}}),!a.keepIsValid&&z()},Qe=({disabled:s,name:a,field:l,fields:d,value:f})=>{if(G(s)&&y.mount||s){const o=s?void 0:E(f)?je(l?l._f:c(d,a)._f):f;k(n,a,o),h(a,o,!1,!1,!0)}},ke=(s,a={})=>{let l=c(i,s);const d=G(a.disabled)||G(e.disabled);return k(i,s,{...l||{},_f:{...l&&l._f?l._f:{ref:{name:s}},name:s,mount:!0,...a}}),_.mount.add(s),l?Qe({field:l,disabled:G(a.disabled)?a.disabled:e.disabled,name:s,value:a.value}):g(s,!0,a.value),{...d?{disabled:a.disabled||e.disabled}:{},...t.progressive?{required:!!a.required,min:ye(a.min),max:ye(a.max),minLength:ye(a.minLength),maxLength:ye(a.maxLength),pattern:ye(a.pattern)}:{},name:s,onChange:ce,onBlur:ce,ref:f=>{if(f){ke(s,a),l=c(i,s);const o=E(f.value)&&f.querySelectorAll&&f.querySelectorAll("input,select,textarea")[0]||f,v=qr(o),D=l._f.refs||[];if(v?D.find(M=>M===o):o===l._f.ref)return;k(i,s,{_f:{...l._f,...v?{refs:[...D.filter(Pe),o,...Array.isArray(c(u,s))?[{}]:[]],ref:{type:o.type,name:s}}:{ref:o}}}),g(s,!1,void 0,o)}else l=c(i,s,{}),l._f&&(l._f.mount=!1),(t.shouldUnregister||a.shouldUnregister)&&!(yr(_.array,s)&&y.action)&&_.unMount.add(s)}}},Xe=()=>t.shouldFocusError&&le(i,fe,_.mount),Er=s=>{G(s)&&(b.state.next({disabled:s}),le(i,(a,l)=>{const d=c(i,l);d&&(a.disabled=d._f.disabled||s,Array.isArray(d._f.refs)&&d._f.refs.forEach(f=>{f.disabled=d._f.disabled||s}))},0,!1))},Ze=(s,a)=>async l=>{let d;l&&(l.preventDefault&&l.preventDefault(),l.persist&&l.persist());let f=I(n);if(b.state.next({isSubmitting:!0}),t.resolver){const{errors:o,values:v}=await C();r.errors=o,f=v}else await P(i);if(U(r.errors,"root"),j(r.errors)){b.state.next({errors:{}});try{await s(f,l)}catch(o){d=o}}else a&&await a({...r.errors},l),Xe(),setTimeout(Xe);if(b.state.next({isSubmitted:!0,isSubmitting:!1,isSubmitSuccessful:j(r.errors)&&!d,submitCount:r.submitCount+1,errors:r.errors}),d)throw d},Cr=(s,a={})=>{c(i,s)&&(E(a.defaultValue)?K(s,I(c(u,s))):(K(s,a.defaultValue),k(u,s,I(a.defaultValue))),a.keepTouched||U(r.touchedFields,s),a.keepDirty||(U(r.dirtyFields,s),r.isDirty=a.defaultValue?x(s,I(c(u,s))):x()),a.keepError||(U(r.errors,s),A.isValid&&z()),b.state.next({...r}))},er=(s,a={})=>{const l=s?I(s):u,d=I(l),f=j(s),o=f?u:d;if(a.keepDefaultValues||(u=l),!a.keepValues){if(a.keepDirtyValues)for(const v of _.mount)c(r.dirtyFields,v)?k(o,v,c(n,v)):K(v,c(o,v));else{if(He&&E(s))for(const v of _.mount){const D=c(i,v);if(D&&D._f){const M=Array.isArray(D._f.refs)?D._f.refs[0]:D._f.ref;if(me(M)){const B=M.closest("form");if(B){B.reset();break}}}}i={}}n=e.shouldUnregister?a.keepDefaultValues?I(u):{}:I(o),b.array.next({values:{...o}}),b.values.next({values:{...o}})}_={mount:a.keepDirtyValues?_.mount:new Set,unMount:new Set,array:new Set,watch:new Set,watchAll:!1,focus:""},y.mount=!A.isValid||!!a.keepIsValid||!!a.keepDirtyValues,y.watch=!!e.shouldUnregister,b.state.next({submitCount:a.keepSubmitCount?r.submitCount:0,isDirty:f?!1:a.keepDirty?r.isDirty:!!(a.keepDefaultValues&&!ie(s,u)),isSubmitted:a.keepIsSubmitted?r.isSubmitted:!1,dirtyFields:f?{}:a.keepDirtyValues?a.keepDefaultValues&&n?be(u,n):r.dirtyFields:a.keepDefaultValues&&s?be(u,s):a.keepDirty?r.dirtyFields:{},touchedFields:a.keepTouched?r.touchedFields:{},errors:a.keepErrors?r.errors:{},isSubmitSuccessful:a.keepIsSubmitSuccessful?r.isSubmitSuccessful:!1,isSubmitting:!1})},rr=(s,a)=>er(te(s)?s(n):s,a);return{control:{register:ke,unregister:Se,getFieldState:ze,handleSubmit:Ze,setError:Je,_executeSchema:C,_getWatch:w,_getDirty:x,_updateValid:z,_removeUnmounted:he,_updateFieldArray:p,_updateDisabledField:Qe,_getFieldArray:R,_reset:er,_resetDefaultValues:()=>te(t.defaultValues)&&t.defaultValues().then(s=>{rr(s,t.resetOptions),b.state.next({isLoading:!1})}),_updateFormState:s=>{r={...r,...s}},_disableForm:Er,_subjects:b,_proxyFormState:A,_setErrors:Q,get _fields(){return i},get _formValues(){return n},get _state(){return y},set _state(s){y=s},get _defaultValues(){return u},get _names(){return _},set _names(s){_=s},get _formState(){return r},set _formState(s){r=s},get _options(){return t},set _options(s){t={...t,...s}}},trigger:de,register:ke,handleSubmit:Ze,watch:kr,setValue:K,getValues:Ye,reset:rr,resetField:Cr,clearErrors:Sr,unregister:Se,setError:Je,setFocus:(s,a={})=>{const l=c(i,s),d=l&&l._f;if(d){const f=d.refs?d.refs[0]:d.ref;f.focus&&(f.focus(),a.shouldSelect&&f.select())}},getFieldState:ze}}function it(e={}){const t=F.useRef(),r=F.useRef(),[i,u]=F.useState({isDirty:!1,isValidating:!1,isLoading:te(e.defaultValues),isSubmitted:!1,isSubmitting:!1,isSubmitSuccessful:!1,isValid:!1,submitCount:0,dirtyFields:{},touchedFields:{},validatingFields:{},errors:e.errors||{},disabled:e.disabled||!1,defaultValues:te(e.defaultValues)?void 0:e.defaultValues});t.current||(t.current={...Jr(e),formState:i});const n=t.current.control;return n._options=e,De({subject:n._subjects.state,next:y=>{hr(y,n._proxyFormState,n._updateFormState,!0)&&u({...n._formState})}}),F.useEffect(()=>n._disableForm(e.disabled),[n,e.disabled]),F.useEffect(()=>{if(n._proxyFormState.isDirty){const y=n._getDirty();y!==i.isDirty&&n._subjects.state.next({isDirty:y})}},[n,i.isDirty]),F.useEffect(()=>{e.values&&!ie(e.values,r.current)?(n._reset(e.values,n._options.resetOptions),r.current=e.values,u(y=>({...y}))):n._resetDefaultValues()},[e.values,n]),F.useEffect(()=>{e.errors&&n._setErrors(e.errors)},[e.errors,n]),F.useEffect(()=>{n._state.mount||(n._updateValid(),n._state.mount=!0),n._state.watch&&(n._state.watch=!1,n._subjects.state.next({...n._formState})),n._removeUnmounted()}),F.useEffect(()=>{e.shouldUnregister&&n._subjects.values.next({values:n._getWatch()})},[e.shouldUnregister,n]),t.current.formState=vr(i,n),t.current}export{tt as C,rt as F,Nr as a,we as b,st as c,Mr as d,c as g,k as s,it as u};
