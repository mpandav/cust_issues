package put

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/sftp"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	flogosftp "github.com/tibco/flogo-sftp/src/app/SFTP/activity"
	"github.com/tibco/flogo-sftp/src/app/SFTP/connector/connection"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&MyActivity{}, New)
}

// New creates a new activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MyActivity{logger: log.ChildLogger(ctx.Logger(), "SFTP-activity-put"), activityName: "put"}, nil
}

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
	logger       log.Logger
	activityName string
}

// Metadata implements activity.Activity.Metadata
func (*MyActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (activity *MyActivity) Eval(context activity.Context) (done bool, err error) {
	activity.logger.Infof(flogosftp.GetMessage(flogosftp.ActivityStart, activity.activityName))

	input := &Input{}
	output := &Output{}

	//Get Input Object
	err = context.GetInputObject(input)
	if err != nil {
		return false, flogosftp.GetError(flogosftp.FailedInputObject, activity.activityName, err.Error())
	}

	//Get sftp client from input connection
	sc := input.Connection.GetConnection().(*sftp.Client)

	//Get data from input mapper
	inputData := input.Input
	inputParams, err := flogosftp.GetInputData(inputData, activity.logger)
	if err != nil {
		return false, flogosftp.GetError(flogosftp.FailedInputProcess, activity.activityName, err.Error())
	}

	//put operation
	var putOperationOutput *flogosftp.Output
	processData := input.ProcessData
	overwrite := input.Overwrite
	if processData {
		putOperationOutput, err = flogosftp.ProcessDataPutOperation(sc, inputParams, overwrite, input.Binary, activity.activityName)
	} else {
		putOperationOutput, err = flogosftp.FileTransferPutOperation(sc, inputParams, overwrite, activity.activityName)
	}

	if err != nil {
		//check if error is due to connection close
		if flogosftp.IsConnectionLost(err) || !flogosftp.IsConnectionAlive(sc) {
			activity.logger.Infof("Error is : %s.", err.Error())
			err := input.Connection.(*connection.SftpSharedConfigManager).Reconnect()
			if err != nil {
				return false, flogosftp.GetRetriableError(flogosftp.DefaultError, activity.activityName, err.Error())
			} else {
				sc = input.Connection.GetConnection().(*sftp.Client)
				// Retry put operation after reconnection
				if processData {
					putOperationOutput, _ = flogosftp.ProcessDataPutOperation(sc, inputParams, overwrite, input.Binary, activity.activityName)
				} else {
					putOperationOutput, _ = flogosftp.FileTransferPutOperation(sc, inputParams, overwrite, activity.activityName)
				}
			}
		} else {
			return false, flogosftp.GetError(flogosftp.DefaultError, activity.activityName, err.Error())
		}
	}
	activity.logger.Infof(flogosftp.GetMessage(flogosftp.ActivityEnd, activity.activityName))

	//Logging output at debug level
	out, err := coerce.ToString(putOperationOutput)
	if err != nil {
		return false, flogosftp.GetError(flogosftp.DefaultError, activity.activityName, err.Error())
	}
	activity.logger.Debugf(flogosftp.GetMessage(flogosftp.ActivityOutput, out))

	//set the output
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(putOperationOutput)
	err = json.Unmarshal(reqBodyBytes.Bytes(), &output.Output)
	if err != nil {
		return false, flogosftp.GetError(flogosftp.DefaultError, activity.activityName, err.Error())
	}
	err = context.SetOutputObject(output)
	if err != nil {
		return false, flogosftp.GetError(flogosftp.DefaultError, activity.activityName, err.Error())
	}

	return true, nil
}
