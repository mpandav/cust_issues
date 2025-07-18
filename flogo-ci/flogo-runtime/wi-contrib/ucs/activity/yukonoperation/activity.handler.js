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
        inputData: "inputData",
    },
    Outputs: {
        output: "output",
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
var UcsOperationActivityHandler = (function (_super) {
    __extends(UcsOperationActivityHandler, _super);
    function UcsOperationActivityHandler(injector, http, ucsConnectionsService) {
        var _this = _super.call(this, injector, http) || this;
        _this.injector = injector;
        _this.http = http;
        _this.ucsConnectionsService = ucsConnectionsService;
        _this.value = function (fieldName, context) {
            var activityContext = context;
            var getFieldValue = fieldValueExtractor(activityContext);
            var connectionId = getFieldValue(PropertyNames.Settings.connection);
            var objectName = getFieldValue(PropertyNames.Inputs.dataObject);
            var actionName = getFieldValue(PropertyNames.Settings.action);
            switch (fieldName) {
                case PropertyNames.Inputs.dataObject:
                    if (connectionId) {
                        return _this.loadSchemaNames({ connectionId: connectionId, actionName: actionName });
                    }
                    break;
                case PropertyNames.Inputs.inputData:
                    if (connectionId && objectName) {
                        return _this.loadInputSchema({ connectionId: connectionId, objectName: objectName }).map(function (schema) {
                            var actionType = getFieldValue(PropertyNames.Settings.action) || "";
                            if (actionType.toLowerCase() === "update") {
                                delete schema.required;
                            }
                            return JSON.stringify(schema);
                        });
                    }
                    break;
                case PropertyNames.Inputs.filter:
                    return rxjs_1.Observable.of(getFilterSchema());
                case PropertyNames.Outputs.output:
                    return rxjs_1.Observable.of(getOutputJsonSchema());
            }
            return null;
        };
        _this.validate = function (fieldName, context) {
            var activityContext = context;
            var getFieldValue = fieldValueExtractor(activityContext);
            var hasConnection = getFieldValue(PropertyNames.Settings.connection);
            var objectName = getFieldValue(PropertyNames.Inputs.dataObject);
            switch (fieldName) {
                case PropertyNames.Inputs.dataObject:
                    return rxjs_1.Observable.of(wi_contrib_1.ValidationResult.newValidationResult().setVisible(Boolean(hasConnection)));
                case PropertyNames.Inputs.filter:
                    return wi_contrib_1.ValidationResult.newValidationResult().setVisible(getFieldValue(PropertyNames.Settings.requiresLookupCondition) &&
                        hasConnection &&
                        objectName);
                case PropertyNames.Inputs.inputData:
                    return wi_contrib_1.ValidationResult.newValidationResult().setVisible(Boolean(getFieldValue(PropertyNames.Settings.requiresInputData) &&
                        hasConnection &&
                        objectName));
            }
            return null;
        };
        return _this;
    }
    UcsOperationActivityHandler.prototype.loadInputSchema = function (_a) {
        var _this = this;
        var connectionId = _a.connectionId, objectName = _a.objectName;
        return this.waitForConnectionReady(connectionId)
            .switchMap(function () { return _this.getSchemaDetails(connectionId, objectName); })
            .map(schemaToFlogoSchema);
    };
    UcsOperationActivityHandler.prototype.loadSchemaNames = function (_a) {
        var _this = this;
        var connectionId = _a.connectionId, actionName = _a.actionName;
        return this.waitForConnectionReady(connectionId).switchMap(function () {
            return _this.listSchemasNames(connectionId, actionName);
        });
    };
    UcsOperationActivityHandler.prototype.waitForConnectionReady = function (connectionId) {
        return rxjs_1.Observable.from(this.ucsConnectionsService.observeConnectionUntilReady(connectionId)).skipWhile(function (conn) { return conn.connectionStatus !== "ready"; });
    };
    UcsOperationActivityHandler.prototype.listSchemasNames = function (connectionId, actionName) {
        var searchParams = new http_1.URLSearchParams();
        searchParams.set("actionName", actionName);
        return this.http
            .get(Endpoints.ListSchemas(connectionId), { search: searchParams })
            .map(function (response) { return response.json(); });
    };
    UcsOperationActivityHandler.prototype.getSchemaDetails = function (connectionId, schemaName) {
        return this.http
            .get(Endpoints.GetSchema({ connectionId: connectionId, schemaName: schemaName }))
            .map(function (response) { return response.json(); });
    };
    return UcsOperationActivityHandler;
}(wi_contrib_1.WiServiceHandlerContribution));
UcsOperationActivityHandler = __decorate([
    wi_contrib_1.WiContrib({}),
    core_1.Injectable(),
    __metadata("design:paramtypes", [core_1.Injector,
        http_1.Http,
        wi_contrib_1.UcsConnectionsService])
], UcsOperationActivityHandler);
exports.UcsOperationActivityHandler = UcsOperationActivityHandler;
function schemaToFlogoSchema(schema) {
    schema = sortSchemaKeys(schema);
    var root = __assign({ $schema: "http://json-schema.org/draft-04/schema#", type: "object" }, schema);
    return root;
}
var sortStrings = function (str1, str2) { return str1.localeCompare(str2); };
function sortSchemaKeys(obj) {
    var keys = Array.from(Object.keys(obj)).sort(sortStrings);
    var sortedObject = {};
    for (var _i = 0, keys_1 = keys; _i < keys_1.length; _i++) {
        var key = keys_1[_i];
        var value = obj[key];
        if (typeof value === "object" && value !== null) {
            value = sortSchemaKeys(value);
        }
        sortedObject[key] = value;
    }
    return sortedObject;
}
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
function getOutputJsonSchema() {
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
                        error: {
                            type: "object",
                            title: "The error schema",
                            properties: {
                                details: {
                                    type: "string",
                                },
                                message: {
                                    type: "string",
                                },
                                number: {
                                    type: "integer",
                                },
                            },
                            additionalProperties: false,
                        },
                        objectsAffected: {
                            type: "integer",
                            description: "Count of objects affected by this operation",
                        },
                        outputData: {
                            type: "any",
                        },
                        success: {
                            type: "boolean",
                        },
                    },
                    additionalProperties: false,
                },
            },
        },
        additionalProperties: false,
    });
}
//# sourceMappingURL=activity.handler.js.map