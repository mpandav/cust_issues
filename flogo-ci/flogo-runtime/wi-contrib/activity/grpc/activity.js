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
    },
  __param =
    (this && this.__param) ||
    function (e, t) {
      return function (r, i) {
        t(r, i, e);
      };
    };
Object.defineProperty(exports, "__esModule", { value: !0 });
var oldProtoFileContent,
  _a,
  http_1 = require("@angular/http"),
  core_1 = require("@angular/core"),
  protobufjs = require("protobufjs"),
  wi_contrib_1 = require("wi-studio/app/contrib/wi-contrib"),
  flogo_contrib_sdk_1 = require("@tibco/flogo-contrib-sdk"),
  emptyArray = [],
  protoMap = new Map(),
  serviceNames = [],
  camelCaseRe = /_([a-z])/g,
  SPEC_PREFIX = "spec://",
  gRPCInvokeCustomerActivityContribution = (function (e) {
    function t(t, r, i) {
      var o = e.call(this, t, r) || this;
      return (
        (o.http = r),
        (o.appSpecsService = i),
        (o.value = function (e, t) {
          var r,
            i,
            n = t.getField("protoFile").value;
          if (n && "" !== n)
            if ((i = "string" == typeof n && n.startsWith(SPEC_PREFIX))) {
              var a = n.replace(SPEC_PREFIX, "");
              o.appSpecsService
                .getAppSpecById(a)
                .take(1)
                .subscribe(function (e) {
                  e && (r = e);
                });
            } else r = n;
          var s = t.getField("serviceName").value,
            c = t.getField("methodName").value;
          switch (e) {
            case "operatingMode":
              return "grpc-to-grpc";
            case "protoName":
              if (
                (console.log("Checking content now..."), r && "" !== r.content)
              )
                return i ? r.name : r.filename;
            case "serviceName":
              if (r && "" !== r.content) {
                var l = r.content;
                if (oldProtoFileContent !== l)
                  if (((oldProtoFileContent = l), protoMap.has(l))) {
                    var u = protoMap.get(l);
                    serviceNames = u.success
                      ? Object.keys(u.services)
                      : emptyArray;
                  } else {
                    u = o.parseProtoFile(r);
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
            case "grpcMthdParamtrs":
              return r && "" !== r.content && s && "" !== s && c && "" !== c
                ? (u = protoMap.get(r.content)).services[s].methods[c].inputs
                : null;
            case "body":
              return r && "" !== r.content && s && "" !== s && c && "" !== c
                ? (u = protoMap.get(r.content)).services[s].methods[c].outputs
                : null;
            case "enableTLS":
            default:
              return null;
          }
        }),
        (o.validate = function (e, t) {
          var r,
            i = t.getField("protoFile").value;
          if (i && "" !== i)
            if ("string" == typeof i && i.startsWith(SPEC_PREFIX)) {
              var n = i.replace(SPEC_PREFIX, "");
              o.appSpecsService
                .getAppSpecById(n)
                .take(1)
                .subscribe(function (e) {
                  e && (r = e);
                });
            } else r = i;
          if ("protoFile" === e) {
            if (!r)
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
            if ("grpcMthdParamtrs" === e || "body" === e)
              return wi_contrib_1.ValidationResult.newValidationResult().setReadOnly(
                !0
              );
            if ("protoName" === e)
              return wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                !1
              );
            if ("enableMTLS" === e)
              return t.getField("enableTLS").value
                ? ((l =
                    wi_contrib_1.ValidationResult.newValidationResult()).setVisible(
                    !0
                  ),
                  l)
                : wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !1
                  );
            if ("clientCert2" === e) {
              var s = t.getField("enableMTLS").value,
                c = t.getField("enableTLS").value;
              return s && c
                ? ((l =
                    wi_contrib_1.ValidationResult.newValidationResult()).setVisible(
                    !0
                  ),
                  t.getField("clientCert2").value ||
                    l.setError(
                      "Required",
                      "A Client Certificate file must be configured."
                    ),
                  l)
                : wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !1
                  );
            }
            if ("clientCert" === e)
              return t.getField("enableTLS").value
                ? ((l =
                    wi_contrib_1.ValidationResult.newValidationResult()).setVisible(
                    !0
                  ),
                  t.getField("clientCert").value ||
                    l.setError(
                      "Required",
                      "A CA or Server Certificate file must be configured."
                    ),
                  l)
                : wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !1
                  );
            if ("clientKey" === e) {
              var l;
              (s = t.getField("enableMTLS").value),
                (c = t.getField("enableTLS").value);
              return s && c
                ? ((l =
                    wi_contrib_1.ValidationResult.newValidationResult()).setVisible(
                    !0
                  ),
                  t.getField("clientKey").value ||
                    l.setError(
                      "Required",
                      "A private key file must be configured."
                    ),
                  l)
                : wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                    !1
                  );
            }
            if (
              "operatingMode" === e ||
              "params" === e ||
              "queryParams" === e ||
              "content" === e ||
              "pathParams" === e
            )
              return wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                !1
              );
            if ("header" === e)
              return wi_contrib_1.ValidationResult.newValidationResult().setVisible(
                !0
              );
          }
          return null;
        }),
        o
      );
    }
    return (
      __extends(t, e),
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
                  var c = function (e) {
                      var t = { methods: {} },
                        o = n.root.lookupService(a + s[e]);
                      o.methodsArray.forEach(function (e) {
                        var i = {};
                        (i.name = r.convertToCamelCase(e.name)),
                          (i.inputs = r.parseProtoMethod(e, "input")),
                          (i.outputs = r.parseProtoMethod(e, "output")),
                          (t.methods[i.name] = i);
                      }),
                        (i.services[l.convertToCamelCase(o.name)] = t);
                    },
                    l = this,
                    u = 0;
                  u < s.length;
                  u++
                )
                  c(u);
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
(gRPCInvokeCustomerActivityContribution = __decorate(
  [
    wi_contrib_1.WiContrib({}),
    core_1.Injectable(),
    __param(0, core_1.Inject(core_1.Injector)),
    __metadata("design:paramtypes", [
      Object,
      http_1.Http,
      ("function" ==
        typeof (_a = (
          void 0 !== flogo_contrib_sdk_1.default && flogo_contrib_sdk_1.default
        ).AppSpecsService) &&
        _a) ||
        Object,
    ]),
  ],
  gRPCInvokeCustomerActivityContribution
)),
  (exports.gRPCInvokeCustomerActivityContribution =
    gRPCInvokeCustomerActivityContribution);
//# sourceMappingURL=activity.js.map
