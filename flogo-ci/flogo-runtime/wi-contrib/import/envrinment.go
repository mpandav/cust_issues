package main

import (
	"os"

	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/engine"
)

func init() {
	os.Setenv(engine.EnvEnableSchemaSupport, "true")
	os.Setenv("GOLANG_PROTOBUF_REGISTRATION_CONFLICT", "ignore")
	os.Setenv("FLOGO_TASK_PRIORITIZE_EXPR_LINK", "true")
	os.Setenv("FLOGO_TASK_PROPAGATE_SKIP", "false")
	//Enable schema
	schema.Enable()
	if engine.IsSchemaValidationEnabled() {
		schema.ValidationEnabled()
	} else {
		schema.DisableValidation()
	}
}
