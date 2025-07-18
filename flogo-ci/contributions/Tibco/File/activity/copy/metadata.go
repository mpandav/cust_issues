package copy

import (
	"github.com/project-flogo/core/data/coerce"
)

// Input corresponds to activity.json inputs
type Input struct {
	CreateNonExistingDir  bool                   `md:"createNonExistingDir,required"`
	IncludeSubDirectories bool                   `md:"includeSubDirectories,required"`
	Input                 map[string]interface{} `md:"input"`
}

// ToMap converts Input struct to map
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"createNonExistingDir":  i.CreateNonExistingDir,
		"includeSubDirectories": i.IncludeSubDirectories,
		"input":                 i.Input,
	}
}

// FromMap converts a map to Input struct
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.CreateNonExistingDir, err = coerce.ToBool(values["createNonExistingDir"])
	if err != nil {
		return err
	}

	i.IncludeSubDirectories, err = coerce.ToBool(values["includeSubDirectories"])
	if err != nil {
		return err
	}

	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}

	return nil
}
