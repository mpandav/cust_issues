package delete

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	sftp "github.com/tibco/flogo-sftp/src/app/SFTP/connector/connection"
)

// Input corresponds to activity.json inputs
type Input struct {
	Connection connection.Manager     `md:"SFTP Connection,required"`
	Input      map[string]interface{} `md:"input"`
}

// ToMap converts Input struct to map
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"SFTP Connection": i.Connection,
		"input":           i.Input,
	}
}

// FromMap converts a map to Input struct
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Connection, err = sftp.GetSharedConfiguration(values["SFTP Connection"])
	if err != nil {
		return err
	}

	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}

	return nil
}
