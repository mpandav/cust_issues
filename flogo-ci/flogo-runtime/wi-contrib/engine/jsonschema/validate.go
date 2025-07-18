package jsonschema

import (
	"strings"

	"fmt"
	"os"
	"strconv"

	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/project-flogo/core/support/log"
	"github.com/xeipuuv/gojsonschema"
)

var validatelog = log.ChildLogger(log.RootLogger(), "jsonschema-validate")

func validate(schema, data string) (bool, []string) {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	jsonDataLoader := gojsonschema.NewStringLoader(data)
	errors := []string{}
	result, err := gojsonschema.Validate(schemaLoader, jsonDataLoader)

	if err != nil {
		errors = append(errors, err.Error())
		return false, errors
	}

	if result.Valid() {
		validatelog.Info("The document is valid")
		return true, nil
	} else {
		validatelog.Info("The document is not valid")
		for _, desc := range result.Errors() {
			errString := desc.String()
			validatelog.Error(errString)
			errString = strings.Replace(errString, ":", "->", -1)
			errors = append(errors, errString)
		}
		return false, errors
	}
}

func ValidateFromObject(schema string, data interface{}) error {
	validation, ok := os.LookupEnv("WI_ENBALE_VALIDATION")
	if ok {
		b, err := strconv.ParseBool(validation)
		if err != nil {
			logger.Errorf("Parsing WI_ENBALE_VALIDATION env variable error [%s]", err.Error())
			b = false
		}
		if b {
			if schema != "" {
				ok, errs := validateFromObject(schema, data)
				if !ok {
					for _, v := range errs {
						validatelog.Errorf("Json schema validation error [%s]", v)
					}
					return fmt.Errorf("Output validate failed")
				}

			}
		}
	}
	return nil
}

func ForceValidateFromObject(schema string, data interface{}) error {
	if schema != "" {
		ok, errs := validateFromObject(schema, data)
		if !ok {
			for _, v := range errs {
				validatelog.Errorf("Json schema validation error [%s]", v)
			}
			return fmt.Errorf("validation failed, %v", errs)
		}

	}

	return nil
}

func Validate(schema string, data string) error {
	validation, ok := os.LookupEnv("WI_ENBALE_VALIDATION")
	if ok {
		b, err := strconv.ParseBool(validation)
		if err != nil {
			logger.Errorf("Parsing WI_ENBALE_VALIDATION env variable error [%s]", err.Error())
			b = false
		}
		if b {
			if schema != "" {
				ok, errs := validate(schema, data)
				if !ok {
					for _, v := range errs {
						validatelog.Errorf("Json schema validation error [%s]", v)
					}
					return fmt.Errorf("Output validate failed")
				}

			}
		}
	}
	return nil
}

func ForceValidate(schema string, data string) error {
	if schema != "" {
		ok, errs := validate(schema, data)
		if !ok {
			for _, v := range errs {
				validatelog.Errorf("Json schema validation error [%s]", v)
			}
			return fmt.Errorf("validation failed, %v", errs)
		}

	}
	return nil
}

func validateFromObject(schema string, data interface{}) (bool, []string) {
	schemaLoader := gojsonschema.NewStringLoader(schema)
	jsonDataLoader := gojsonschema.NewGoLoader(data)
	errors := []string{}
	result, err := gojsonschema.Validate(schemaLoader, jsonDataLoader)

	if err != nil {
		errors = append(errors, err.Error())
		return false, errors
	}

	if result.Valid() {
		validatelog.Info("The document is valid")
		return true, nil
	} else {
		validatelog.Info("The document is not valid")
		for _, desc := range result.Errors() {
			errString := desc.String()
			validatelog.Error(errString)
			errString = strings.Replace(errString, ":", "->", -1)
			errors = append(errors, errString)
		}
		return false, errors
	}
}
