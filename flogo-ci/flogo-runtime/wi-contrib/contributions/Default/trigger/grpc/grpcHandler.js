"use strict";
var __extends =
    (this && this.__extends) ||
    (function () {
      var e =
        Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array &&
          function (e, t) {
            e.__proto__ = t;
          }) ||
        function (e, t) {
          for (var r in t) t.hasOwnProperty(r) && (e[r] = t[r]);
        };
      return function (t, r) {
        function i() {
          this.constructor = t;
        }
        e(t, r),
          (t.prototype =
            null === r
              ? Object.create(r)
              : ((i.prototype = r.prototype), new i()));
      };
    })(),
  __decorate =
    (this && this.__decorate) ||
    function (e, t, r, i) {
      var o,
        n = arguments.length,
        a =
          n < 3
            ? t
            : null === i
            ? (i = Object.getOwnPropertyDescriptor(t, r))
            : i;
      if ("object" == typeof Reflect && "function" == typeof Reflect.decorate)
        a = Reflect.decorate(e, t, r, i);
      else
        for (var s = e.length - 1; s >= 0; s--)
          (o = e[s]) &&
            (a = (n < 3 ? o(a) : n > 3 ? o(t, r, a) : o(t, r)) || a);
      return n > 3 && a && Object.defineProperty(t, r, a), a;
    },
  __metadata =
    (this && this.__metadata) ||
    function (e, t) {
      if ("object" == typeof Reflect && "function" == typeof Reflect.metadata)
        return Reflect.metadata(e, t);
    };
Object.defineProperty(exports, "__esModule", { value: !0 });
var oldProtoFileContent,
  _a,
  core_1 = require("@angular/core"),
  http_1 = require("@angular/http"),
  Observable_1 = require("rxjs/Observable"),
  protobufjs = require("protobufjs"),
  wi_contrib_1 = require("wi-studio/app/contrib/wi-contrib"),
  contrib_1 = require("wi-studio/common/models/contrib"),
  lodash = require("lodash"),
  flogo_contrib_sdk_1 = require("@tibco/flogo-contrib-sdk"),
  emptyArray = [],
  protoMap = new Map(),
  serviceNames = [],
  camelCaseRe = /_([a-z])/g,
  SPEC_PREFIX = "spec://",
  grpcHandler = (function (e) {
    function t(t, r, i, o) {
      var n = e.call(this, t, r, i) || this;
      return (
        (n.injector = t),
        (n.http = r),
        (n.contribModelService = i),
        (n.appSpecsService = o),
        (n.value = function (e, t) {
          var r,
            i,
            o = t.getField("protoFile").value;
          if (o && "" !== o)
            if ((i = "string" == typeof o && o.startsWith(SPEC_PREFIX))) {
              var a = o.replace(SPEC_PREFIX, "");
              n.appSpecsService
                .getAppSpecById(a)
                .take(1)
                .subscribe(function (e) {
                  e && (r = e);
                });
            } else r = o;
          var s = t.getField("serviceName").value,
            l = t.getField("methodName").value;
          switch (e) {
            case "protoName":
              if (r && "" !== r.content) return i ? r.name : r.filename;
            case "serviceName":
              if (
                (console.log("Checking content now.."), r && "" !== r.content)
              ) {
                var c = r.content;
                if (oldProtoFileContent !== c)
                  if (((oldProtoFileContent = c), protoMap.has(c))) {
                    var u = protoMap.get(c);
                    serviceNames = u.success
                      ? Object.keys(u.services)
                      : emptyArray;
                  } else {
                    u = n.parseProtoFile(r);
                    serviceNames = u.success
                      ? Object.keys(u.services)
                      : emptyArray;
                  }
                return serviceNames;
              }
              return emptyArray;
            case "methodName":
              if (r && "" !== r.content && s && "" !== s) {
                u = protoMap.get(r.content);
                return Object.keys(u.services[s].methods);
              }
              return emptyArray;
            case "params":
              return r && "" !== r.content && s && "" !== s && l && "" !== l
                ? (u = protoMap.get(r.content)).services[s].methods[l].inputs
                : null;
            case "data":
              return r && "" !== r.content && s && "" !== s && l && "" !== l
                ? (u = protoMap.get(r.content)).services[s].methods[l].outputs
                : null;
            case "enableTLS":
            case "enableMTLS":
              return Observable_1.Observable.create(function (e) {
                wi_contrib_1.WiContributionUtils.getAppConfig(n.http).subscribe(
                  function (t) {
                    t.deployment === wi_contrib_1.APP_DEPLOYMENT.CLOUD
                      ? e.next(!1)
                      : e.next(null);
                  },
                  function (t) {
                    e.next(
                      wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                        !1
                      )
                    );
                  },
                  function () {
                    return e.complete();
                  }
                );
              });
            default:
              return null;
          }
        }),
        (n.validate = function (e, t) {
          if ("port" === e || "enableTLS" === e)
            return Observable_1.Observable.create(function (e) {
              wi_contrib_1.WiContributionUtils.getAppConfig(n.http).subscribe(
                function (t) {
                  t.deployment === wi_contrib_1.APP_DEPLOYMENT.ON_PREMISE
                    ? e.next(
                        wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                          !0
                        )
                      )
                    : e.next(
                        wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                          !1
                        )
                      );
                },
                function (t) {
                  e.next(
                    wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                      !1
                    )
                  );
                },
                function () {
                  return e.complete();
                }
              );
            });
          if ("protoFile" === e) {
            var r,
              i = t.getField("protoFile").value;
            if (i && "" !== i)
              if ("string" == typeof i && i.startsWith(SPEC_PREFIX)) {
                var o = i.replace(SPEC_PREFIX, "");
                n.appSpecsService
                  .getAppSpecById(o)
                  .take(1)
                  .subscribe(function (e) {
                    e && (r = e);
                  });
              } else r = i;
            if ((console.log("protoFile", r), !r || !r.content))
              return wi_contrib_1.ValidationResult.newValidationResult().setError(
                "Required",
                "A proto file must be configured."
              );
            if (protoMap && protoMap.has(r.content)) {
              var a = protoMap.get(r.content);
              if (!a.success)
                return "RangeError: Maximum call stack size exceeded" ===
                  a.error
                  ? wi_contrib_1.ValidationResult.newValidationResult().setError(
                      "gRPCError",
                      "Proto file should not have any cyclic dependency."
                    )
                  : wi_contrib_1.ValidationResult.newValidationResult().setError(
                      "gRPCError",
                      a.error
                    );
            }
          } else {
            if ("protoName" === e)
              return wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                !1
              );
            if ("enableMTLS" === e)
              return (l = t.getField("enableTLS").value)
                ? wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !0
                  )
                : wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !1
                  );
            if ("rootCA" === e) {
              var s,
                l = t.getField("enableMTLS").value,
                c = t.getField("enableTLS").value;
              return l && c
                ? ((s =
                    wi_contrib_1.ValidationResult.newValidationResult()).setVisible(
                    !0
                  ),
                  t.getField("rootCA").value ||
                    s.setError(
                      "Required",
                      "CA Certificate must be configured."
                    ),
                  s)
                : wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !1
                  );
            }
            if ("serverCert" === e)
              return (l = t.getField("enableTLS").value)
                ? ((s =
                    wi_contrib_1.ValidationResult.newValidationResult()).setVisible(
                    !0
                  ),
                  t.getField("serverCert").value ||
                    s.setError(
                      "Required",
                      "Server certificate must be configured."
                    ),
                  s)
                : wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !1
                  );
            if ("serverKey" === e)
              return (l = t.getField("enableTLS").value)
                ? ((s =
                    wi_contrib_1.ValidationResult.newValidationResult()).setVisible(
                    !0
                  ),
                  t.getField("serverKey").value ||
                    s.setError(
                      "Required",
                      "Server private key must be configured."
                    ),
                  s)
                : wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !1
                  );
            if ("content" === e || "grpcData" === e || "code" === e)
              return wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                !1
              );
          }
          return null;
        }),
        (n.action = function (e, t) {
          var r,
            i = n.getModelService(),
            o = wi_contrib_1.CreateFlowActionResult.newActionResult(),
            a = wi_contrib_1.ActionResult.newActionResult(),
            s = t.getField("protoFile").value;
          if (s && "" !== s)
            if ("string" == typeof s && s.startsWith(SPEC_PREFIX)) {
              var l = s.replace(SPEC_PREFIX, "");
              n.appSpecsService
                .getAppSpecById(l)
                .take(1)
                .subscribe(function (e) {
                  e && (r = e);
                });
            } else r = s;
          return Observable_1.Observable.create(function (e) {
            var s = n.parseProtoFile(r);
            if (s.success) {
              if (t.getMode() === contrib_1.MODE.SERVERLESS_FLOW) {
                var l = t.getField("serviceName").value,
                  c = t.getField("methodName").value,
                  u = n.doTriggerConfiguration(t, s, l, c),
                  p = i.createFlow(t.getFlowName(), t.getFlowDescription(), !1);
                o = o.addTriggerFlowMapping(
                  lodash.cloneDeep(u),
                  lodash.cloneDeep(p)
                );
              } else
                t.getMode() === contrib_1.MODE.UPLOAD &&
                  Object.keys(s.services).map(function (e) {
                    Object.keys(s.services[e].methods).map(function (r) {
                      var a = n.doTriggerConfiguration(t, s, e, r),
                        l = i.createFlow(e + "_" + r, "", !1),
                        c = i.createFlowElement("Default/flogo-return"),
                        u = l.addFlowElement(c);
                      o = o.addTriggerFlowMapping(
                        lodash.cloneDeep(a),
                        lodash.cloneDeep(u)
                      );
                    });
                  });
              a.setSuccess(!0).setResult(o);
            } else a.setSuccess(!1), a.setResult(wi_contrib_1.ValidationResult.newValidationResult().setError("gRPCError", s.error));
            e.next(a);
          });
        }),
        n
      );
    }
    return (
      __extends(t, e),
      (t.prototype.doTriggerConfiguration = function (e, t, r, i) {
        var o = e.getField("port").value,
          n = e.getField("protoName").value,
          a = e.getField("enableTLS").value,
          s = e.getField("enableMTLS").value,
          l = e.getField("serverCert").value,
          c = e.getField("serverKey").value,
          u = e.getField("protoFile").value,
          p = e.getField("rootCA").value,
          d = this.getModelService().createTriggerElement(
            "Default/grpc-trigger"
          );
        if (
          (d &&
            d.settings &&
            d.settings.length > 0 &&
            d.settings.map(function (e) {
              "port" === e.name
                ? (e.value = o)
                : "protoName" === e.name
                ? (e.value = n)
                : "enableTLS" === e.name
                ? (e.value = a)
                : "serverCert" === e.name
                ? (e.value = l)
                : "serverKey" === e.name
                ? (e.value = c)
                : "protoFile" === e.name
                ? (e.value = u)
                : "enableMTLS" === e.name
                ? (e.value = s)
                : "rootCA" === e.name && (e.value = p);
            }),
          d &&
            d.handler &&
            d.handler.settings &&
            d.handler.settings.length > 0 &&
            d.handler.settings.map(function (e) {
              "serviceName" === e.name
                ? (e.value = r)
                : "methodName" === e.name && (e.value = i);
            }),
          d && d.outputs && d.outputs.length > 0)
        )
          for (var f = 0; f < d.outputs.length; f++)
            if ("params" === d.outputs[f].name) {
              d.outputs[f].value = t.services[r].methods[i].inputs;
              break;
            }
        if (d && d.reply && d.reply.length > 0)
          for (f = 0; f < d.reply.length; f++)
            if ("data" === d.reply[f].name) {
              d.reply[f].value = t.services[r].methods[i].outputs;
              break;
            }
        return this.doTriggerMapping(d), d;
      }),
      (t.prototype.doTriggerMapping = function (e) {
        for (
          var t = this.contribModelService.createMapping(),
            r = this.contribModelService.createMapExpression(),
            i = 0;
          i < e.outputs.length;
          i++
        )
          if ("params" === e.outputs[i].name) {
            t.addMapping(
              "$INPUT['" + e.outputs[i].name + "']",
              r.setExpression("$trigger." + e.outputs[i].name)
            );
            break;
          }
        e.inputMappings = t;
        for (
          var o = this.contribModelService.createMapping(),
            n = this.contribModelService.createMapExpression(),
            a = 0;
          a < e.reply.length;
          a++
        )
          if ("data" === e.reply[a].name) {
            o.addMapping(
              "$INPUT['" + e.reply[a].name + "']",
              n.setExpression("$flow." + e.reply[a].name)
            );
            break;
          }
        e.outputMappings = o;
      }),
      (t.prototype.parseServiceNames = function (e, t) {
        for (
          var r = e.toJSON(), i = t.split("."), o = [], n = 0;
          !((o = Object.keys(r.nested)).length > 1);

        )
          r = r.nested[i[n++]];
        var a = [];
        return (
          o.forEach(function (e) {
            r.nested[e].methods && a.push(e);
          }),
          a
        );
      }),
      (t.prototype.parseProtoFile = function (e) {
        var t,
          r = this,
          i = {},
          o = e.content.split(",")[1];
        t = null == o ? atob(e.content) : atob(o);
        try {
          var n = protobufjs.parse(t, { keepCase: !0 });
          if (n && "proto2" !== n.syntax) {
            (i.success = !0), (i.services = {});
            try {
              var a = "";
              n.package && (a = n.package + ".");
              var s = this.parseServiceNames(n.root, a);
              if (s && s.length > 0)
                for (
                  var l = function (e) {
                      var t = { methods: {} },
                        o = n.root.lookupService(a + s[e]);
                      o.methodsArray.forEach(function (e) {
                        var i = {};
                        (i.name = r.convertToCamelCase(e.name)),
                          (i.inputs = r.parseProtoMethod(e, "input")),
                          (i.outputs = r.parseProtoMethod(e, "output")),
                          (t.methods[i.name] = i);
                      }),
                        (i.services[c.convertToCamelCase(o.name)] = t);
                    },
                    c = this,
                    u = 0;
                  u < s.length;
                  u++
                )
                  l(u);
              else
                (i.success = !1),
                  (i.error =
                    "Error: Could not find any service definition in proto file.");
            } catch (e) {
              (i.success = !1), (i.error = e + "");
            }
          } else
            (i.success = !1),
              (i.error =
                "Error: proto2 syntax is not supported. Please define the proto file using proto3 syntax.");
        } catch (e) {
          (i.success = !1), (i.error = e + "");
        }
        return protoMap.set(e.content, i), i;
      }),
      (t.prototype.parseProtoMethod = function (e, t) {
        var r = this,
          i = { type: "object", properties: {}, required: [] };
        return (
          e.resolved || e.resolve(),
          ("input" === t
            ? e.resolvedRequestType.fieldsArray
            : e.resolvedResponseType.fieldsArray
          ).forEach(function (e) {
            e.resolved || e.resolve(),
              "repeated" === e.rule
                ? ((i.properties[e.name] = { type: "array" }),
                  (i.properties[e.name].items = r.coerceField(e)))
                : (i.properties[e.name] = r.coerceField(e));
          }),
          JSON.stringify(i)
        );
      }),
      (t.prototype.coerceField = function (e) {
        return this.isScalarType(e.type)
          ? this.coerceType(e.type)
          : this.coerceType(e.resolvedType);
      }),
      (t.prototype.isScalarType = function (e) {
        switch (e) {
          case "uint32":
          case "int32":
          case "sint32":
          case "int64":
          case "uint64":
          case "sint64":
          case "fixed32":
          case "sfixed32":
          case "fixed64":
          case "sfixed64":
          case "float":
          case "double":
          case "bool":
          case "bytes":
          case "string":
            return !0;
          default:
            return !1;
        }
      }),
      (t.prototype.coerceType = function (e) {
        var t = this;
        if (e instanceof protobufjs.Enum)
          return (
            e.resolved || e.resolve(),
            { type: "string", enum: Object.keys(e.values) }
          );
        if (e instanceof protobufjs.Type) {
          e.resolved || e.resolve();
          var r = { type: "object", properties: {} };
          return (
            e.fieldsArray.forEach(function (e) {
              e.resolved || e.resolve(),
                "repeated" === e.rule
                  ? ((r.properties[e.name] = { type: "array" }),
                    (r.properties[e.name].items = t.coerceField(e)))
                  : (r.properties[e.name] = t.coerceField(e));
            }),
            r
          );
        }
        switch (e) {
          case "uint32":
          case "int32":
          case "sint32":
          case "int64":
          case "uint64":
          case "sint64":
          case "fixed32":
          case "sfixed32":
          case "fixed64":
          case "sfixed64":
          case "float":
          case "double":
            return { type: "number" };
          case "bool":
            return { type: "boolean" };
          case "bytes":
          case "string":
            return { type: "string" };
          default:
            return { type: "any" };
        }
      }),
      (t.prototype.convertToCamelCase = function (e) {
        return (
          e.substring(0, 1).toUpperCase() +
          e.substring(1).replace(camelCaseRe, function (e, t) {
            return t.toUpperCase();
          })
        );
      }),
      t
    );
  })(wi_contrib_1.WiServiceHandlerContribution);
(grpcHandler = __decorate(
  [
    wi_contrib_1.WiContrib({}),
    core_1.Injectable(),
    __metadata("design:paramtypes", [
      core_1.Injector,
      http_1.Http,
      wi_contrib_1.WiContribModelService,
      ("function" ==
        typeof (_a = (
          void 0 !== flogo_contrib_sdk_1.default && flogo_contrib_sdk_1.default
        ).AppSpecsService) &&
        _a) ||
        Object,
    ]),
  ],
  grpcHandler
)),
  (exports.grpcHandler = grpcHandler);
//# sourceMappingURL=grpcHandler.js.map
