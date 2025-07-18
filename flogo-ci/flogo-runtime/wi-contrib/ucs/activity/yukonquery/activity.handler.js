"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
var __assign = (this && this.__assign) || Object.assign || function(t) {
    for (var s, i = 1, n = arguments.length; i < n; i++) {
        s = arguments[i];
        for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
            t[p] = s[p];
    }
    return t;
};
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var http_1 = require("@angular/http");
var rxjs_1 = require("rxjs");
var Observable_1 = require("rxjs/Observable");
var wi_contrib_1 = require("wi-studio/app/contrib/wi-contrib");
var UCS_API_URL = "/wistudio/v1/connections/universal";
var Endpoints = {
    ListConnections: function () { return UCS_API_URL + "/connections"; },
    GetConnection: function (connectionId) {
        return UCS_API_URL + "/connections/" + connectionId;
    },
    ListSchemas: function (connectionId) {
        return UCS_API_URL + "/connections/" + connectionId + "/schemas";
    },
    GetSchema: function (p) {
        return UCS_API_URL + "/connections/" + p.connectionId + "/schemas/" + p.schemaName;
    },
};
var CACHE_TIME_MS = 20 * 1000;
var PropertyNames = {
    Settings: {
        connection: "connection",
        action: "action",
        requiresLookupCondition: "requiresLookupCondition",
        requiresInputData: "requiresInputData",
    },
    Inputs: {
        dataObject: "dataObject",
        filter: "filter",
    },
    Outputs: {
        output: "output",
        fieldSelection: "fieldSelection",
    },
};
var fieldValueExtractor = function (context) {
    return function getFieldValue(fieldName) {
        var field = context.getField(fieldName);
        if (field) {
            return field.value;
        }
    };
};
var parseJsonAsArray = function (value) {
    if (value === null || value === undefined) {
        return value;
    }
    try {
        return JSON.parse(value);
    }
    catch (e) {
        return [];
    }
};
var UcsQueryActivityHandler = (function (_super) {
    __extends(UcsQueryActivityHandler, _super);
    function UcsQueryActivityHandler(injector, http, ucsConnectionsService) {
        var _this = _super.call(this, injector, http) || this;
        _this.injector = injector;
        _this.http = http;
        _this.ucsConnectionsService = ucsConnectionsService;
        _this.activityData = new Map();
        _this.value = function (fieldName, context) {
            var activityUniqueId = wi_contrib_1.WiContributionUtils.getUniqueId(context);
            var activityContext = context;
            var getFieldValue = fieldValueExtractor(activityContext);
            var connectionId = getFieldValue(PropertyNames.Settings.connection);
            var objectName = getFieldValue(PropertyNames.Inputs.dataObject);
            var actionName = getFieldValue(PropertyNames.Settings.action);
            var existingFields = parseJsonAsArray(getFieldValue(PropertyNames.Outputs.fieldSelection));
            switch (fieldName) {
                case PropertyNames.Inputs.dataObject:
                    if (connectionId) {
                        return _this.loadSchemaNames({ connectionId: connectionId, actionName: actionName });
                    }
                    break;
                case PropertyNames.Outputs.fieldSelection:
                    if (connectionId && objectName) {
                        return _this.loadFields({
                            activityUniqueId: activityUniqueId,
                            connectionId: connectionId,
                            objectName: objectName,
                            existingFields: existingFields,
                        }).map(function (fields) { return JSON.stringify(fields); });
                    }
                    break;
                case PropertyNames.Inputs.filter:
                    return Observable_1.Observable.of(getFilterSchema());
                case PropertyNames.Outputs.output:
                    var activityContext_1 = _this.getActivityData(activityUniqueId);
                    return activityContext_1.schema$
                        .asObservable()
                        .filter(function (v) { return !!v && v.objectName === objectName; })
                        .first()
                        .map(function (v) {
                        var schema = v ? v.schema : undefined;
                        return getOutputJsonSchema(extractSelectedFields(schema, existingFields));
                    });
            }
            return null;
        };
        _this.validate = function (fieldName, context) {
            var activityContext = context;
            var getFieldValue = fieldValueExtractor(activityContext);
            var connectionId = getFieldValue(PropertyNames.Settings.connection);
            var objectName = getFieldValue(PropertyNames.Inputs.dataObject);
            switch (fieldName) {
                case PropertyNames.Inputs.dataObject:
                    return Observable_1.Observable.of(wi_contrib_1.ValidationResult.newValidationResult().setVisible(Boolean(connectionId)));
                case PropertyNames.Inputs.filter:
                    return Observable_1.Observable.of(wi_contrib_1.ValidationResult.newValidationResult().setVisible(true));
                case PropertyNames.Outputs.fieldSelection:
                    return Observable_1.Observable.of(wi_contrib_1.ValidationResult.newValidationResult().setVisible(Boolean(connectionId && objectName)));
            }
            return null;
        };
        return _this;
    }
    UcsQueryActivityHandler.prototype.loadFields = function (_a) {
        var _this = this;
        var activityUniqueId = _a.activityUniqueId, connectionId = _a.connectionId, objectName = _a.objectName, existingFields = _a.existingFields;
        return this.waitForConnectionReady(connectionId)
            .switchMap(function () {
            var cachedSchema = _this.getCachedSchema({
                activityUniqueId: activityUniqueId,
                connectionId: connectionId,
                objectName: objectName,
            });
            if (cachedSchema) {
                return Observable_1.Observable.of(cachedSchema.schema);
            }
            else {
                return _this.getSchemaDetails(connectionId, objectName).do(function (schema) {
                    _this.setCachedSchema({
                        activityUniqueId: activityUniqueId,
                        connectionId: connectionId,
                        objectName: objectName,
                        schema: schema,
                    });
                });
            }
        })
            .do(function (schema) {
            _this.getActivityData(activityUniqueId).schema$.next({
                objectName: objectName,
                schema: schema,
            });
        })
            .map(function (schema) {
            return _this.mergeSchemaAndFields(schema, existingFields);
        });
    };
    UcsQueryActivityHandler.prototype.mergeSchemaAndFields = function (schema, existingFields) {
        var selectableFields = [];
        var schemaPropNames = new Set(Object.keys(schema.properties));
        existingFields = existingFields || [];
        for (var _i = 0, existingFields_1 = existingFields; _i < existingFields_1.length; _i++) {
            var existingField = existingFields_1[_i];
            if (schemaPropNames.has(existingField.FieldName)) {
                selectableFields.push(existingField);
                schemaPropNames.delete(existingField.FieldName);
            }
        }
        for (var _a = 0, _b = Array.from(schemaPropNames.values()); _a < _b.length; _a++) {
            var propName = _b[_a];
            selectableFields.push({
                FieldName: propName,
                Selected: "false",
            });
        }
        selectableFields.sort(function (f1, f2) {
            return f1.FieldName.localeCompare(f2.FieldName);
        });
        return selectableFields;
    };
    UcsQueryActivityHandler.prototype.loadSchemaNames = function (_a) {
        var _this = this;
        var connectionId = _a.connectionId, actionName = _a.actionName;
        return this.waitForConnectionReady(connectionId).switchMap(function () {
            return _this.listSchemasNames(connectionId, actionName);
        });
    };
    UcsQueryActivityHandler.prototype.waitForConnectionReady = function (connectionId) {
        return Observable_1.Observable.from(this.ucsConnectionsService.observeConnectionUntilReady(connectionId)).skipWhile(function (conn) { return conn.connectionStatus !== "ready"; });
    };
    UcsQueryActivityHandler.prototype.listSchemasNames = function (connectionId, actionName) {
        var searchParams = new http_1.URLSearchParams();
        searchParams.set("actionName", actionName);
        return this.http
            .get(Endpoints.ListSchemas(connectionId), { search: searchParams })
            .map(function (response) { return response.json(); });
    };
    UcsQueryActivityHandler.prototype.getSchemaDetails = function (connectionId, schemaName) {
        return this.http
            .get(Endpoints.GetSchema({ connectionId: connectionId, schemaName: schemaName }))
            .map(function (response) { return response.json(); });
    };
    UcsQueryActivityHandler.prototype.getActivityData = function (activityUniqueId) {
        if (!this.activityData.has(activityUniqueId)) {
            this.activityData.set(activityUniqueId, {
                uniqueId: activityUniqueId,
                schema$: new rxjs_1.BehaviorSubject(undefined),
            });
        }
        return this.activityData.get(activityUniqueId);
    };
    UcsQueryActivityHandler.prototype.getCachedSchema = function (params) {
        var activityData = this.getActivityData(params.activityUniqueId);
        var cachedSchema = activityData.cachedSchema;
        if (!cachedSchema) {
            return null;
        }
        if (Date.now() - cachedSchema.createdAt >= CACHE_TIME_MS) {
            activityData.cachedSchema = undefined;
            return null;
        }
        if (cachedSchema.connectionId === params.connectionId &&
            cachedSchema.objectName === params.objectName) {
            return cachedSchema;
        }
        return null;
    };
    UcsQueryActivityHandler.prototype.setCachedSchema = function (params) {
        var activityData = this.getActivityData(params.activityUniqueId);
        activityData.cachedSchema = {
            connectionId: params.connectionId,
            objectName: params.objectName,
            schema: params.schema,
            createdAt: Date.now(),
        };
    };
    return UcsQueryActivityHandler;
}(wi_contrib_1.WiServiceHandlerContribution));
UcsQueryActivityHandler = __decorate([
    wi_contrib_1.WiContrib({}),
    core_1.Injectable(),
    __metadata("design:paramtypes", [core_1.Injector,
        http_1.Http,
        wi_contrib_1.UcsConnectionsService])
], UcsQueryActivityHandler);
exports.UcsQueryActivityHandler = UcsQueryActivityHandler;
function getFilterSchema() {
    return JSON.stringify({
        $schema: "http://json-schema.org/draft-07/schema",
        type: "object",
        required: ["lookupCondition"],
        properties: {
            lookupCondition: {
                type: "object",
                title: "Lookup condition",
            },
        },
    });
}
function getOutputJsonSchema(outputDataSchema) {
    return JSON.stringify({
        $schema: "http://json-schema.org/draft-07/schema",
        type: "object",
        properties: {
            action: {
                type: "string",
                title: "Action",
                description: "Operation name",
            },
            dataObject: {
                type: "string",
                title: "Data object",
            },
            results: {
                type: "array",
                title: "Operation results",
                additionalItems: true,
                items: {
                    type: "object",
                    properties: {
                        outputData: outputDataSchema
                            ? __assign({}, outputDataSchema) : {
                            type: "any",
                        },
                    },
                    additionalProperties: false,
                },
            },
        },
        additionalProperties: false,
    });
}
function extractSelectedFields(schema, existingFields) {
    if (!schema) {
        return undefined;
    }
    var outputSchema = {
        type: "object",
        properties: {},
    };
    if (Array.isArray(existingFields)) {
        for (var _i = 0, existingFields_2 = existingFields; _i < existingFields_2.length; _i++) {
            var field = existingFields_2[_i];
            if (field.Selected === "true" && schema.properties) {
                outputSchema.properties[field.FieldName] =
                    schema.properties[field.FieldName];
            }
        }
    }
    return outputSchema;
}
//# sourceMappingURL=activity.handler.js.map