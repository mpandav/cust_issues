package queuereceiver

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Settings struct {
	Connection connection.Manager `md:"azservicebusConnection,required"`
}

type HandlerSettings struct {
	//Connection  connection.Manager `md:"azservicebusConnection"`
	Queue       string `md:"queue"`
	SessionId   string `md:"sessionId"`
	ReceiveMode string `md:"receiveMode"`
	ValueType   string `md:"valueType"`
	Deadletter  bool   `md:"deadletter"`
	Timeout     int    `md:"Timeout"`
	Count       int    `md:"Count"`
	Interval    int    `md:"Interval"`
}

type Output struct {
	Output map[string]interface{} `md:"output"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}
	return nil
}
