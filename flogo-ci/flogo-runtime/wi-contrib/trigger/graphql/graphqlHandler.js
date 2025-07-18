"use strict";
var __extends = this && this.__extends || function() {
        var e = Object.setPrototypeOf || {
            __proto__: []
        }
        instanceof Array && function(e, t) {
            e.__proto__ = t
        } || function(e, t) {
            for (var i in t) t.hasOwnProperty(i) && (e[i] = t[i])
        };
        return function(t, i) {
            function r() {
                this.constructor = t
            }
            e(t, i), t.prototype = null === i ? Object.create(i) : (r.prototype = i.prototype, new r)
        }
    }(),
    __decorate = this && this.__decorate || function(e, t, i, r) {
        var n, a = arguments.length,
            o = a < 3 ? t : null === r ? r = Object.getOwnPropertyDescriptor(t, i) : r;
        if ("object" == typeof Reflect && "function" == typeof Reflect.decorate) o = Reflect.decorate(e, t, i, r);
        else
            for (var l = e.length - 1; l >= 0; l--)(n = e[l]) && (o = (a < 3 ? n(o) : a > 3 ? n(t, i, o) : n(t, i)) || o);
        return a > 3 && o && Object.defineProperty(t, i, o), o
    },
    __metadata = this && this.__metadata || function(e, t) {
        if ("object" == typeof Reflect && "function" == typeof Reflect.metadata) return Reflect.metadata(e, t)
    };
Object.defineProperty(exports, "__esModule", {
    value: !0
});
var oldSchema, _a, _b, _c, core_1 = require("@angular/core"),
    http_1 = require("@angular/http"),
    Observable_1 = require("rxjs/Observable"),
    wi_contrib_1 = require("wi-studio/app/contrib/wi-contrib"),
    flogo_contrib_sdk_1 = require("@tibco/flogo-contrib-sdk"),
    validation_1 = require("wi-studio/common/models/validation"),
    contrib_1 = require("wi-studio/common/models/contrib"),
    lodash = require("lodash"),
    flogo_contrib_sdk_2 = require("@tibco/flogo-contrib-sdk"),
    gqlMap = new Map,
    mutationFields = [],
    queryFields = [],
    operationTypes = [{
        unique_id: "Query",
        name: "Query"
    }, {
        unique_id: "Mutation",
        name: "Mutation"
    }],
    SPEC_PREFIX = "spec://",
    emptyArray = [],
    graphqlHandler = function(e) {
        function t(t, i, r, n, a, o) {
            var l = e.call(this, t, i, r) || this;
            return l.injector = t, l.http = i, l.contribModelService = r, l.parsingService = n, l.configurationService = a, l.appSpecsService = o, l.value = function(e, t) {
                var i = t.getField("operation").value,
                    r = t.getField("resolverFor").value,
                    n = t.getField("graphqlSchema").value,
                    a = t.getField("schemaFile").value;
                if ("string" == typeof a && a.startsWith(SPEC_PREFIX)) {
                    var o = a.replace(SPEC_PREFIX, "");
                    l.appSpecsService.getAppSpecById(o).take(1).subscribe(function(e) {
                        e && (a = e)
                    })
                }
                switch (a && (n = l.decodeFileContent(a)), e) {
                    case "graphqlSchema":
                        if (a) return n;
                        break;
                    case "resolverFor":
                        var s;
                        return i ? (oldSchema != n ? s = l.parseResolverFor(n, i) : (s = l.cachedResolverFor(btoa(n), i), l.messaging.emit("GraphQL-arguments", "notify"), l.messaging.emit("GraphQL-data", "notify")), s) : emptyArray;
                    case "operation":
                        return n && "" != n ? operationTypes : emptyArray;
                    case "secureConnection":
                        return Observable_1.Observable.create(function(e) {
                            l.configurationService.getAppConfig().subscribe(function(t) {
                                t.deployment === wi_contrib_1.APP_DEPLOYMENT.CLOUD ? e.next(!1) : e.next(null)
                            }, function(t) {
                                e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))
                            }, function() {
                                return e.complete()
                            })
                        });
                    case contrib_1.METADATA.MANIFEST:
                        return l.generateEndpoints();
                    case "arguments":
                        return r ? Observable_1.Observable.create(function(e) {
                            l.messaging.on("GraphQL-arguments", function(t) {
                                if (null != t) {
                                    l.messaging.off("GraphQL-arguments");
                                    for (var a = 0, o = gqlMap.get(btoa(n)).flows; a < o.length; a++) {
                                        var s = o[a];
                                        s.type == i && s.name == r && e.next(s.inputs)
                                    }
                                }
                            })
                        }) : null;
                    case "data":
                        return r ? Observable_1.Observable.create(function(e) {
                            l.messaging.on("GraphQL-data", function(t) {
                                if (null != t) {
                                    l.messaging.off("GraphQL-data");
                                    for (var a = 0, o = gqlMap.get(btoa(n)).flows; a < o.length; a++) {
                                        var s = o[a];
                                        s.type == i && s.name == r && e.next(s.outputs)
                                    }
                                }
                            })
                        }) : null;
                    default:
                        return null
                }
                return null
            }, l.validate = function(e, t) {
                var i = t.getField("graphqlSchema").value,
                    r = t.getField("schemaFile").value;
                if ("string" == typeof r && r.startsWith(SPEC_PREFIX)) {
                    var n = r.replace(SPEC_PREFIX, "");
                    l.appSpecsService.getAppSpecById(n).take(1).subscribe(function(e) {
                        e && (r = e)
                    })
                }
                if ("path" === e) {
                    var a = t.getField("path").value;
                    if ("" == a || !a.startsWith("/")) return validation_1.ValidationResult.newValidationResult().setError("PATH_ERROR", 'Path required and should start with "/" like "/graphql"')
                } else {
                    if ("port" === e) return validation_1.ValidationResult.newValidationResult().setVisible(!0);
                    if ("secureConnection" === e) return Observable_1.Observable.create(function(e) {
                        l.configurationService.getAppConfig().subscribe(function(t) {
                            t.deployment === wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE ? e.next(validation_1.ValidationResult.newValidationResult().setVisible(!0)) : e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))
                        }, function(t) {
                            e.next(validation_1.ValidationResult.newValidationResult().setVisible(!1))
                        }, function() {
                            return e.complete()
                        })
                    });
                    if ("operation" === e || "resolverFor" === e) return i && "" != i && !r ? validation_1.ValidationResult.newValidationResult().setReadOnly(!0) : validation_1.ValidationResult.newValidationResult().setReadOnly(!1);
                    if ("graphqlSchema" === e) return validation_1.ValidationResult.newValidationResult().setVisible(!1);
                    if ("schemaFile" === e) {
                        if (r) {
                            var o = "" != r.content ? r.content.split(",")[1] : "";
                            if (gqlMap && gqlMap.has(o)) {
                                var s = gqlMap.get(o);
                                if (!s.success) return validation_1.ValidationResult.newValidationResult().setError("GraphQLError", s.error)
                            }
                        } else if (!i || "" == i) return validation_1.ValidationResult.newValidationResult().setError("Required", "A graphql schema file must be configured.")
                    } else {
                        if ("caCertificate" === e) {
                            if (t.getField("secureConnection").value) {
                                var c = validation_1.ValidationResult.newValidationResult().setVisible(!0);
                                return t.getField("caCertificate").value || c.setError("Required", "A CA or Server Certificate file must be configured."), c
                            }
                            return validation_1.ValidationResult.newValidationResult().setVisible(!1)
                        }
                        if ("serverKey" === e) {
                            if (t.getField("secureConnection").value) {
                                c = validation_1.ValidationResult.newValidationResult().setVisible(!0);
                                return t.getField("serverKey").value || c.setError("Required", "A Server Key file must be configured."), c
                            }
                            return validation_1.ValidationResult.newValidationResult().setVisible(!1)
                        }
                    }
                }
                return null
            }, l.action = function(e, t) {
                var i = l.getModelService(),
                    r = wi_contrib_1.CreateFlowActionResult.newActionResult(),
                    n = wi_contrib_1.ActionResult.newActionResult(),
                    a = t.getField("schemaFile").value;
                if ("string" == typeof a && a.startsWith(SPEC_PREFIX)) {
                    var o = a.replace(SPEC_PREFIX, "");
                    l.appSpecsService.getAppSpecById(o).take(1).subscribe(function(e) {
                        e && (a = e)
                    })
                }
                var s = l.decodeFileContent(a),
                    c = [];
                return Observable_1.Observable.create(function(e) {
                    l.parsingService.parseGraphqlSchema(s).subscribe(function(a) {
                        if (gqlMap.set(btoa(s), a), !1 === a.success) n.setSuccess(!1), n.setResult(validation_1.ValidationResult.newValidationResult().setError("GraphQLError", a.error)), e.next(n);
                        else {
                            if (c = a.flows, t.getMode() === contrib_1.MODE.SERVERLESS_FLOW) {
                                for (var o = t.getField("operation").value, u = t.getField("resolverFor").value, p = void 0, d = 0, g = c; d < g.length; d++) {
                                    var f = g[d];
                                    if (f.type === o && f.name === u) {
                                        p = f;
                                        break
                                    }
                                }
                                var v = l.doTriggerConfiguration(t, s, p),
                                    _ = i.createFlow(t.getFlowName(), t.getFlowDescription(), !1);
                                r = r.addTriggerFlowMapping(lodash.cloneDeep(v), lodash.cloneDeep(_))
                            } else t.getMode() === contrib_1.MODE.UPLOAD && c.map(function(e) {
                                var n = l.doTriggerConfiguration(t, s, e),
                                    a = i.createFlow(e.type + "_" + e.name, e.description, !1),
                                    o = i.createFlowElement("Default/flogo-return"),
                                    c = a.addFlowElement(o);
                                r = r.addTriggerFlowMapping(lodash.cloneDeep(n), lodash.cloneDeep(c))
                            });
                            n.setSuccess(!0).setResult(r), e.next(n)
                        }
                    }, function(t) {
                        e.next(null)
                    }, function() {
                        return e.complete()
                    })
                })
            }, l.messaging = new wi_contrib_1.WiContribMessaging, l
        }
        return __extends(t, e), t.prototype.doTriggerConfiguration = function(e, t, i) {
            var r = e.getField("schemaFile").value,
                n = e.getField("port").value,
                a = e.getField("path").value,
                o = e.getField("secureConnection").value,
                l = e.getField("serverKey").value,
                s = e.getField("caCertificate").value,
                c = this.getModelService().createTriggerElement("Default/tibco-graphql");
            return c && c.settings && c.settings.length > 0 && c.settings.map(function(e) {
                "path" === e.name ? e.value = a : "port" === e.name ? e.value = n : "graphqlSchema" === e.name ? e.value = t : "schemaFile" === e.name ? e.value = r : "secureConnection" === e.name ? e.value = o : "serverKey" === e.name ? e.value = l : "caCertificate" === e.name && (e.value = s)
            }), c && c.handler && c.handler.settings && c.handler.settings.length > 0 && c.handler.settings.map(function(e) {
                "operation" === e.name ? e.value = i.type : "resolverFor" === e.name && (e.value = i.name)
            }), c && c.outputs && c.outputs.length > 0 && "arguments" === c.outputs[0].name && (c.outputs[0].value = i.inputs), c && c.reply && c.reply.length > 0 && "data" === c.reply[0].name && (c.reply[0].value = i.outputs), this.doTriggerMapping(c), c
        }, t.prototype.doTriggerMapping = function(e) {
            var t = this.contribModelService.createMapping();
            t.addMapping("$INPUT['arguments']", this.contribModelService.createMapExpression().setExpression("$trigger.arguments")), t.addMapping("$INPUT['headers']", this.contribModelService.createMapExpression().setExpression("$trigger.headers")), t.addMapping("$INPUT['fields']", this.contribModelService.createMapExpression().setExpression("$trigger.fields")), e.inputMappings = t;
            var i = this.contribModelService.createMapping();
            i.addMapping("$INPUT['data']", this.contribModelService.createMapExpression().setExpression("isDefined($flow.data) ? $flow.data : coerce.toObject('{}')")), i.addMapping("$INPUT['error']", this.contribModelService.createMapExpression().setExpression("isDefined($flow.error) ? $flow.error : ''")), e.outputMappings = i
        }, t.prototype.cachedResolverFor = function(e, t) {
            return gqlMap.has(e) && gqlMap.get(e).success ? "Query" === t ? queryFields : mutationFields : emptyArray
        }, t.prototype.parseResolverFor = function(e, t, i) {
            var r = this;
            return Observable_1.Observable.create(function(i) {
                r.parsingService.parseGraphqlSchema(e).subscribe(function(n) {
                    if (gqlMap.set(btoa(e), n), r.messaging.emit("GraphQL-arguments", "notify"), r.messaging.emit("GraphQL-data", "notify"), oldSchema = e, !n.success) return i.next(emptyArray);
                    mutationFields.length = 0, queryFields.length = 0, n.flows.map(function(e) {
                        "Query" === e.type ? queryFields.push(e.name) : mutationFields.push(e.name)
                    }), "Query" === t ? i.next(queryFields) : i.next(mutationFields)
                }, function(e) {
                    i.next(emptyArray)
                }, function() {
                    i.complete()
                })
            })
        }, t.prototype.decodeFileContent = function(e) {
            if (e && "" != e.content) {
                var t = e.content.split(",")[1];
                return null == t ? atob(e.content) : atob(t)
            }
            return ""
        }, t.prototype.generateEndpoints = function() {
            for (var e = this, t = [], i = this.getModelService().getApplication().getTriggerFlowModelMaps(), r = new Map, n = 0, a = i; n < a.length; n++) {
                var o = a[n];
                "github.com/project-flogo/graphql/trigger/graphql" === o.getTriggerElement().ref && r.set(o.getTriggerElement().getId(), o.getTriggerElement())
            }
            return r.forEach(function(i) {
                var r = {
                    title: i.getId(),
                    pingable: !0,
                    protocol: "http",
                    port: i.settings.port.value + "",
                    specType: "graphql",
                    spec: {
                        name: e.getModelService().getApplication().getName(),
                        version: "1.1.0"
                    }
                };
                t.push(r)
            }), JSON.stringify(t)
        }, t
    }(wi_contrib_1.WiServiceHandlerContribution);
graphqlHandler = __decorate([wi_contrib_1.WiContrib({}), core_1.Injectable(), __metadata("design:paramtypes", [core_1.Injector, http_1.Http, wi_contrib_1.WiContribModelService, "function" == typeof(_a = void 0 !== flogo_contrib_sdk_1.ParsingService && flogo_contrib_sdk_1.ParsingService) && _a || Object, "function" == typeof(_b = void 0 !== flogo_contrib_sdk_1.ConfigurationService && flogo_contrib_sdk_1.ConfigurationService) && _b || Object, "function" == typeof(_c = (void 0 !== flogo_contrib_sdk_2.default && flogo_contrib_sdk_2.default).AppSpecsService) && _c || Object])], graphqlHandler), exports.graphqlHandler = graphqlHandler;
//# sourceMappingURL=graphqlHandler.js.map