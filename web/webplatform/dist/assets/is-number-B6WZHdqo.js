import"./@descope-QDsTNvKh.js";(function(){try{var r=typeof window<"u"?window:typeof global<"u"?global:typeof self<"u"?self:{},e=new r.Error().stack;e&&(r._sentryDebugIds=r._sentryDebugIds||{},r._sentryDebugIds[e]="01a211a7-caf1-4901-b570-d396cd76fe75",r._sentryDebugIdIdentifier="sentry-dbid-01a211a7-caf1-4901-b570-d396cd76fe75")}catch{}})();/*!
 * is-number <https://github.com/jonschlinkert/is-number>
 *
 * Copyright (c) 2014-2017, Jon Schlinkert.
 * Released under the MIT License.
 */var n=function(e){var t=typeof e;if(t==="string"||e instanceof String){if(!e.trim())return!1}else if(t!=="number"&&!(e instanceof Number))return!1;return e-e+1>=0};export{n as i};
