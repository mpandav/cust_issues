package dfpipeline

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Input Structure
type Input struct {
	Connection     connection.Manager `md:"Connection,required"`
	Operation      string             `md:"operation"`
	SubscriptionId string             `md:"subscriptionId"`
	ResourceGroup  string             `md:"resourceGroup"`
	DataFactories  string             `md:"dataFactories"`
	DfPipeline     string             `md:"dfPipeline"`
	Input          interface{}        `md:"input"`
}

// ToMap Input interface
func (o *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection":     o.Connection,
		"Operation":      o.Operation,
		"SubscriptionId": o.SubscriptionId,
		"DataFactories":  o.DataFactories,
		"ResourceGroup":  o.ResourceGroup,
		"DfPipeline":     o.DfPipeline,
		"Input":          o.Input,
	}
}

// FromMap Input interface
func (o *Input) FromMap(values map[string]interface{}) error {
	var err error
	o.Operation, err = coerce.ToString(values["operation"])
	if err != nil {
		return err
	}

	o.SubscriptionId, err = coerce.ToString(values["subscriptionId"])
	if err != nil {
		return err
	}
	o.DataFactories, err = coerce.ToString(values["dataFactories"])
	if err != nil {
		return err
	}

	o.ResourceGroup, err = coerce.ToString(values["resourceGroup"])
	if err != nil {
		return err
	}
	o.DfPipeline, err = coerce.ToString(values["dfPipeline"])
	if err != nil {
		return err
	}
	o.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}

	o.Connection, err = coerce.ToConnection(values["Connection"])
	if err != nil {
		return err
	}

	return nil

}

// Output struct
type Output struct {
	Output map[string]interface{} `md:"output"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}
	return err
}
