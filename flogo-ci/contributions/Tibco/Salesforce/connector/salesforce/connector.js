"use strict";var __extends=this&&this.__extends||function(){var e=function(t,n){return(e=Object.setPrototypeOf||{__proto__:[]}instanceof Array&&function(e,t){e.__proto__=t}||function(e,t){for(var n in t)Object.prototype.hasOwnProperty.call(t,n)&&(e[n]=t[n])})(t,n)};return function(t,n){if("function"!=typeof n&&null!==n)throw new TypeError("Class extends value "+String(n)+" is not a constructor or null");function i(){this.constructor=t}e(t,n),t.prototype=null===n?Object.create(n):(i.prototype=n.prototype,new i)}}(),__decorate=this&&this.__decorate||function(e,t,n,i){var o,r=arguments.length,a=r<3?t:null===i?i=Object.getOwnPropertyDescriptor(t,n):i;if("object"==typeof Reflect&&"function"==typeof Reflect.decorate)a=Reflect.decorate(e,t,n,i);else for(var s=e.length-1;s>=0;s--)(o=e[s])&&(a=(r<3?o(a):r>3?o(t,n,a):o(t,n))||a);return r>3&&a&&Object.defineProperty(t,n,a),a},__metadata=this&&this.__metadata||function(e,t){if("object"==typeof Reflect&&"function"==typeof Reflect.metadata)return Reflect.metadata(e,t)},__param=this&&this.__param||function(e,t){return function(n,i){t(n,i,e)}};Object.defineProperty(exports,"__esModule",{value:!0}),exports.SalesforceConnectorService=void 0;var core_1=require("@angular/core"),http_1=require("@angular/http"),wi_contrib_1=require("wi-studio/app/contrib/wi-contrib"),validation_1=require("wi-studio/common/models/validation"),contrib_1=require("wi-studio/common/models/contrib"),Observable_1=require("rxjs/Observable"),jose=require("jose"),lodash=require("lodash"),wi_contrib_internal_1=require("wi-studio/app/contrib/wi-contrib-internal"),AUTH_ENDPOINT="https://login.salesforce.com/services/oauth2/authorize",TOKEN_ENDPOINT="https://login.salesforce.com/services/oauth2/token",SANDBOX_AUTH_ENDPOINT="https://test.salesforce.com/services/oauth2/authorize",SANDBOX_TOKEN_ENDPOINT="https://test.salesforce.com/services/oauth2/token",SalesforceConnectorService=function(e){function t(t,i){var o=e.call(this,t,i)||this;return o.http=i,o.value=function(e,t){if("authType"===e)return"OAUTH2"===t.getField("authType").value?"OAuth 2.0 Web Server Flow":null;if("customOAuth2Credentials"===e)return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getAppConfig(o.http).subscribe(function(t){t.deployment===wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE?e.next(!0):e.next(null)},function(t){e.next(null)},function(){return e.complete()})});if("WI_STUDIO_OAUTH_CONNECTOR_INFO"===e)for(var i=function(e){if("WI_STUDIO_OAUTH_CONNECTOR_INFO"===e.name)try{var i=JSON.parse(e.value),r=n.authCodeMap.has(wi_contrib_1.WiContributionUtils.getUniqueId(t))?n.authCodeMap.get(wi_contrib_1.WiContributionUtils.getUniqueId(t)):"";if(i.code&&r!==i.code)return n.authCodeMap.set(wi_contrib_1.WiContributionUtils.getUniqueId(t),i.code),{value:Observable_1.Observable.create(function(e){o.getOauthConfig(t).subscribe(function(n){o.getAccessToken(t,i.code).subscribe(function(t){var i={client_id:n.client_id};i=lodash.assign(i,t);var o=btoa(JSON.stringify(i));e.next(o)},function(t){console.log("Auth error: ",t),e.next("")})})})}}catch(e){return{value:null}}},r=0,a=t.settings;r<a.length;r++){var s=i(a[r]);if("object"==typeof s)return s.value}return null},o.validate=function(e,t){var n=t.getField("authType").value;if("WI_STUDIO_OAUTH_CONNECTOR_INFO"===e){var i=o.checkForTokens(t);return"OAuth 2.0 JWT Bearer Flow"===n&&i.setVisible(!1),i}if("environment"===e){for(var r=0,a=t.settings;r<a.length;r++){var s=a[r];if("WI_STUDIO_OAUTH_CONNECTOR_INFO"===s.name)if(s.value)try{var l=JSON.parse(s.value);l.access_token&&l.refresh_token&&!0}catch(e){!1}else!1}return validation_1.ValidationResult.newValidationResult().setVisible(!0)}if("customOAuth2Credentials"===e)return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getAppConfig(o.http).subscribe(function(t){t.deployment===wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE||"OAuth 2.0 JWT Bearer Flow"===n?e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1)):e.next(validation_1.ValidationResult.newValidationResult().setVisible(!0))},function(t){e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(){return e.complete()})});if("generateJwt"===e)return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getAppConfig(o.http).subscribe(function(t){"OAuth 2.0 JWT Bearer Flow"===n?e.next(validation_1.ValidationResult.newValidationResult().setVisible(!0)):e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(t){e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(){return e.complete()})});if("clientId"===e){var u=t.getField("customOAuth2Credentials"),c=t.getField("generateJwt");return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getAppConfig(o.http).subscribe(function(t){"OAuth 2.0 JWT Bearer Flow"!==n&&(t.deployment===wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE||u&&u.value&&!0===u.value)||"OAuth 2.0 JWT Bearer Flow"===n&&!0===c.value?e.next(validation_1.ValidationResult.newValidationResult().setVisible(!0)):e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(t){e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(){return e.complete()})})}if("clientSecret"===e){var _=t.getField("customOAuth2Credentials");return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getAppConfig(o.http).subscribe(function(t){"OAuth 2.0 JWT Bearer Flow"!==n&&(t.deployment===wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE||_&&_.value&&!0===_.value)?e.next(validation_1.ValidationResult.newValidationResult().setVisible(!0)):e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(t){e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(){return e.complete()})})}if("jwt"===e){var b=t.getField("generateJwt");return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getAppConfig(o.http).subscribe(function(){"OAuth 2.0 JWT Bearer Flow"===n&&!1===b.value?e.next(validation_1.ValidationResult.newValidationResult().setVisible(!0)):e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(){return e.complete()})})}if("subject"===e||"jwtExpiry"===e||"clientKey"===e){var d=t.getField("generateJwt");return Observable_1.Observable.create(function(e){wi_contrib_1.WiContributionUtils.getAppConfig(o.http).subscribe(function(){"OAuth 2.0 JWT Bearer Flow"===n&&!0===d.value?e.next(validation_1.ValidationResult.newValidationResult().setVisible(!0)):e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))},function(){return e.complete()})})}return"Login"===e?"OAuth 2.0 JWT Bearer Flow"!==n?validation_1.ValidationResult.newValidationResult().setVisible(!0):validation_1.ValidationResult.newValidationResult().setVisible(!1):"Save"===e?"OAuth 2.0 JWT Bearer Flow"===n?validation_1.ValidationResult.newValidationResult().setVisible(!0):validation_1.ValidationResult.newValidationResult().setVisible(!1):null},o.action=function(e,t){if("Login"===e||"Save"===e)return Observable_1.Observable.create(function(e){for(var n="",i="",r=!1,a=0;a<t.settings.length;a++)"name"===t.settings[a].name?n=t.settings[a].value:"authType"===t.settings[a].name?i=t.settings[a].value:"generateJwt"===t.settings[a].name&&(r=t.settings[a].value);var s=!1;wi_contrib_1.WiContributionUtils.getConnections(o.http,"Salesforce").subscribe(function(a){for(var l=0,u=a;l<u.length;l++)for(var c=u[l],_=0;_<c.settings.length;_++){if("name"===c.settings[_].name)if(c.settings[_].value===n&&wi_contrib_1.WiContributionUtils.getUniqueId(c)!==wi_contrib_1.WiContributionUtils.getUniqueId(t)){s=!0;break}}if(s)e.next(contrib_1.ActionResult.newActionResult().setSuccess(!1).setResult(new validation_1.ValidationError("SALESFORCE-1000","Connection name already exist !!!"))),e.complete();else{for(_=0;_<t.settings.length;_++)if("WI_STUDIO_OAUTH_CONNECTOR_INFO"===t.settings[_].name){t.settings[_].value;break}"OAuth 2.0 JWT Bearer Flow"!==i?o.getOauthConfig(t).subscribe(function(n){var i={context:t,authType:wi_contrib_1.AUTHENTICATION_TYPE.OAUTH2,authData:{authLoginUrl:o.getAuthLoginUrl(t,n),authStateQueryParam:"state"}};e.next(contrib_1.ActionResult.newActionResult().setResult(i))}):"OAuth 2.0 JWT Bearer Flow"!==i||r?o.getJSONWebToken(t).subscribe(function(n){for(var i=0;i<t.settings.length;i++)if("jwt"===t.settings[i].name){t.settings[i].value=n;break}o.getAccessTokenUsingJWT(t).subscribe(function(n){for(var i=0;i<t.settings.length;i++)if("WI_STUDIO_OAUTH_CONNECTOR_INFO"===t.settings[i].name){t.settings[i].value=btoa(JSON.stringify(n));break}var o={context:t,authType:wi_contrib_1.AUTHENTICATION_TYPE.BASIC,authData:{}};e.next(contrib_1.ActionResult.newActionResult().setResult(o))},function(t){console.log("Error: ",t),e.next(contrib_1.ActionResult.newActionResult().setSuccess(!1).setResult(new validation_1.ValidationError("Salesforce-1001","This connection is invalid as failed to get the required tokens from the credentials provided.")))})},function(e){console.log("Err: ",e)}):o.getAccessTokenUsingJWT(t).subscribe(function(n){for(var i=0;i<t.settings.length;i++)if("WI_STUDIO_OAUTH_CONNECTOR_INFO"===t.settings[i].name){t.settings[i].value=btoa(JSON.stringify(n));break}var o={context:t,authType:wi_contrib_1.AUTHENTICATION_TYPE.BASIC,authData:{}};e.next(contrib_1.ActionResult.newActionResult().setResult(o))},function(t){console.log("Error: ",t),e.next(contrib_1.ActionResult.newActionResult().setSuccess(!1).setResult(new validation_1.ValidationError("Salesforce-1001","This connection is invalid as failed to get the required tokens from the credentials provided.")))})}})})},o}var n;return __extends(t,e),n=t,t.prototype.getAuthLoginUrl=function(e,t){for(var n=t.client_id,i=t.callback_url,o="",r=0;r<e.settings.length;r++)if("environment"===e.settings[r].name){o=e.settings[r].value;break}var a=AUTH_ENDPOINT;return"Sandbox"===o&&(a=SANDBOX_AUTH_ENDPOINT),encodeURI(a+"?response_type=code&client_id="+n+"&redirect_uri="+i+"&display=popup&prompt=login consent")},t.prototype.getJSONWebToken=function(e){for(var t="",n="",i="",o=3,r="",a=0;a<e.settings.length;a++)"environment"===e.settings[a].name&&(t=e.settings[a].value),"clientId"===e.settings[a].name&&(n=e.settings[a].value),"subject"===e.settings[a].name&&(i=e.settings[a].value),"jwtExpiry"===e.settings[a].name&&(o=e.settings[a].value),"clientKey"===e.settings[a].name&&(r=e.settings[a].value.content,r=atob(r.split("base64,")[1]));var s={iss:n,aud:"Production"===t?"https://login.salesforce.com":"https://test.salesforce.com",sub:i,exp:Math.floor(Date.now()/1e3)+60*o};return Observable_1.Observable.from(jose.importPKCS8(r,"RS256")).switchMap(function(e){var t=new jose.SignJWT(s).setProtectedHeader({alg:"RS256"});return Observable_1.Observable.from(t.sign(e))})},t.prototype.getAccessTokenUsingJWT=function(e){for(var t=this,n="",i="",o=0;o<e.settings.length;o++)"environment"===e.settings[o].name&&(n=e.settings[o].value),"jwt"===e.settings[o].name&&(i=e.settings[o].value);var r="Production"===n?TOKEN_ENDPOINT:SANDBOX_TOKEN_ENDPOINT,a=new http_1.URLSearchParams;return a.set("grant_type","urn:ietf:params:oauth:grant-type:jwt-bearer"),a.set("assertion",i),Observable_1.Observable.create(function(e){wi_contrib_1.WiProxyCORSUtils.createRequest(t.http,r).addMethod("POST").addBody(a.toString()).addHeader("Content-Type","application/x-www-form-urlencoded").send().subscribe(function(t){if(200===t.status){var n=t.json();e.next(n)}else e.error({code:t.status,message:"Failed to get access token using jwt",details:t.json()})},function(t){console.log(t),e.error({code:t.status,message:"Failed to get access token using jwt.",details:t.json()})})})},t.prototype.getAccessToken=function(e,t){var n=this,i=e.getField("customOAuth2Credentials");return Observable_1.Observable.create(function(o){n.getOauthConfig(e).subscribe(function(r){var a=new http_1.URLSearchParams;a.set("grant_type","authorization_code"),a.set("code",decodeURIComponent(t)),a.set("client_id",r.client_id),r.deployEnv===wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE||i&&i.value&&!0===i.value?a.set("client_secret",r.client_secret):a.set("client_secret","SALESFORCE_CLIENT_SECRET"),a.set("redirect_uri",r.callback_url);for(var s="",l=0;l<e.settings.length;l++)if("environment"===e.settings[l].name){s=e.settings[l].value;break}var u=TOKEN_ENDPOINT;("Sandbox"===s&&(u=SANDBOX_TOKEN_ENDPOINT),r.deployEnv===wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE||i&&i.value&&!0===i.value)?wi_contrib_1.WiProxyCORSUtils.createRequest(n.http,u).addBody(a.toString()).addMethod("POST").addHeader("Content-Type","application/x-www-form-urlencoded").send().subscribe(function(e){if(200===e.status){var t=e.json();o.next(t)}else o.error({code:e.status,message:"Failed to get access token",details:e.json()})},function(e){console.log(e),o.error({code:e.status,message:"Failed to get access token",details:e.json()})}):wi_contrib_internal_1.WiInternalProxyCORSUtils.createRequest(n.http,u,"SALESFORCE").addBody(a.toString()).addMethod("POST").addHeader("Content-Type","application/x-www-form-urlencoded").send().subscribe(function(e){if(200===e.status){var t=e.json();o.next(t)}else o.error({code:e.status,message:"Failed to get access token",details:e.json()})},function(e){console.log(e),o.error({code:e.status,message:"Failed to get access token",details:e.json()})})})})},t.prototype.getOauthConfig=function(e){var t=this,n=e.getField("customOAuth2Credentials");return Observable_1.Observable.create(function(i){wi_contrib_1.WiContributionUtils.getAppConfig(t.http).subscribe(function(o){o.deployment===wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE||n&&n.value&&!0===n.value?Observable_1.Observable.zip(t.getClientId(e),t.getClientSecret(e),wi_contrib_1.WiContributionUtils.getEnvironment(t.http,"OAUTH_REDIRECT_URL"),t.getDeployment(o.deployment),function(e,t,n,i){return{client_id:e.value,client_secret:t.value,callback_url:n.value,deployEnv:i.value}}).subscribe(function(e){i.next(e)}):Observable_1.Observable.zip(wi_contrib_1.WiContributionUtils.getEnvironment(t.http,"WISTUDIO_SALESFORCE_CLIENT_ID"),wi_contrib_1.WiContributionUtils.getEnvironment(t.http,"OAUTH_REDIRECT_URL"),t.getDeployment(o.deployment),function(e,t,n){return{client_id:e.value,callback_url:t.value,deployEnv:n.value}}).subscribe(function(e){i.next(e)})},function(e){Observable_1.Observable.zip(wi_contrib_1.WiContributionUtils.getEnvironment(t.http,"WISTUDIO_SALESFORCE_CLIENT_ID"),wi_contrib_1.WiContributionUtils.getEnvironment(t.http,"OAUTH_REDIRECT_URL"),t.getDeployment(wi_contrib_1.APP_DEPLOYMENT.CLOUD),function(e,t,n){return{client_id:e.value,callback_url:t.value,deployEnv:n.value}}).subscribe(function(e){i.next(e)})})})},t.prototype.checkForTokens=function(e){for(var t=0,n=e.settings;t<n.length;t++){var i=n[t];if("WI_STUDIO_OAUTH_CONNECTOR_INFO"===i.name){var o=validation_1.ValidationResult.newValidationResult().setReadOnly(!0);if(!i.value)return o.setValid(!1),o;try{var r=void 0;return(r=i.value.startsWith("{")?JSON.parse(i.value):JSON.parse(atob(i.value))).access_token&&r.refresh_token||o.setValid(!1),o.setVisible(!0)}catch(e){return o.setValid(!1),o}}}},t.prototype.getDeployment=function(e){return Observable_1.Observable.create(function(t){t.next({value:e})})},t.prototype.getClientId=function(e){return Observable_1.Observable.create(function(t){for(var n="",i=0;i<e.settings.length;i++)if("clientId"===e.settings[i].name){n=e.settings[i].value;break}t.next({value:n})})},t.prototype.getClientSecret=function(e){return Observable_1.Observable.create(function(t){for(var n="",i=0;i<e.settings.length;i++)if("clientSecret"===e.settings[i].name){n=e.settings[i].value;break}t.next({value:n})})},t.authCodeMap=new Map,t=n=__decorate([core_1.Injectable(),wi_contrib_1.WiContrib({}),__param(0,core_1.Inject(core_1.Injector)),__metadata("design:paramtypes",[Object,http_1.Http])],t)}(wi_contrib_1.WiServiceHandlerContribution);exports.SalesforceConnectorService=SalesforceConnectorService;
//# sourceMappingURL=connector.js.map
