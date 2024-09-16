import{g as D}from"./hoist-non-react-statics-DTUYskP9.js";import"./@descope-QDsTNvKh.js";import{r as m}from"./react-BA7A5rMr.js";(function(){try{var e=typeof window<"u"?window:typeof global<"u"?global:typeof self<"u"?self:{},t=new e.Error().stack;t&&(e._sentryDebugIds=e._sentryDebugIds||{},e._sentryDebugIds[t]="1d5fa3a1-6a66-43a3-9966-71fb93793b9b",e._sentryDebugIdIdentifier="sentry-dbid-1d5fa3a1-6a66-43a3-9966-71fb93793b9b")}catch{}})();var E={exports:{}},w={},g={exports:{}},_={};/**
 * @license React
 * use-sync-external-store-shim.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var f=m;function $(e,t){return e===t&&(e!==0||1/e===1/t)||e!==e&&t!==t}var b=typeof Object.is=="function"?Object.is:$,j=f.useState,I=f.useEffect,V=f.useLayoutEffect,O=f.useDebugValue;function k(e,t){var r=t(),i=j({inst:{value:r,getSnapshot:t}}),n=i[0].inst,u=i[1];return V(function(){n.value=r,n.getSnapshot=t,p(n)&&u({inst:n})},[e,r,t]),I(function(){return p(n)&&u({inst:n}),e(function(){p(n)&&u({inst:n})})},[e]),O(r),r}function p(e){var t=e.getSnapshot;e=e.value;try{var r=t();return!b(e,r)}catch{return!0}}function q(e,t){return t()}var C=typeof window>"u"||typeof window.document>"u"||typeof window.document.createElement>"u"?q:k;_.useSyncExternalStore=f.useSyncExternalStore!==void 0?f.useSyncExternalStore:C;g.exports=_;var F=g.exports;/**
 * @license React
 * use-sync-external-store-shim/with-selector.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var d=m,L=F;function M(e,t){return e===t&&(e!==0||1/e===1/t)||e!==e&&t!==t}var R=typeof Object.is=="function"?Object.is:M,W=L.useSyncExternalStore,z=d.useRef,A=d.useEffect,B=d.useMemo,G=d.useDebugValue;w.useSyncExternalStoreWithSelector=function(e,t,r,i,n){var u=z(null);if(u.current===null){var s={hasValue:!1,value:null};u.current=s}else s=u.current;u=B(function(){function y(o){if(!S){if(S=!0,v=o,o=i(o),n!==void 0&&s.hasValue){var c=s.value;if(n(c,o))return l=c}return l=o}if(c=l,R(v,o))return c;var h=i(o);return n!==void 0&&n(c,h)?c:(v=o,l=h)}var S=!1,v,l,x=r===void 0?null:r;return[function(){return y(t())},x===null?void 0:function(){return y(x())}]},[t,r,i,n]);var a=W(e,u[0],u[1]);return A(function(){s.hasValue=!0,s.value=a},[a]),G(a),a};E.exports=w;var H=E.exports;const P=D(H);export{P as u};
