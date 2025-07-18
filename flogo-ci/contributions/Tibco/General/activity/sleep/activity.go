package sleep

import (
	"time"

	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{})

const (
	intervalTypeSecond      = "Second"
	intervalTypeMillisecond = "Millisecond"
	intervalTypeMinute      = "Minute"
)

func init() {
	_ = activity.Register(&SleepActivity{}, New)
}

// New creates new instance of SleepActivity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &SleepActivity{}, nil
}

// SleepActivity is an Activity that is used to create time lag between activities by making it sleep
// inputs : {IntervalType, Interval}
// outputs: none
type SleepActivity struct {
}

//func init() {
//	md := activity.NewMetadata(jsonMetadata)
//	activity.Register(&SleepActivity{metadata: md})
//}

// Metadata returns the activity's metadata
func (a *SleepActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Sleep the Activity
func (a *SleepActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debugf("Executing Sleep activity")

	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	intervalType := ""
	var interval int64 = 0

	if input.IntervalTypeSetting != "" {
		intervalType = input.IntervalTypeSetting
	}

	if input.IntervalType != "" {
		intervalType = input.IntervalType
	}

	if input.IntervalSetting != 0 {
		interval = input.IntervalSetting

	}

	if input.Interval != 0 {
		interval = input.Interval
	}

	context.Logger().Debugf("Activity sleep :-  [Interval : %d] and [type : %s]", interval, intervalType)

	switch intervalType {
	case intervalTypeMillisecond:
		time.Sleep(time.Duration(interval) * time.Millisecond)
	case intervalTypeSecond:
		time.Sleep(time.Duration(interval) * time.Second)
	case intervalTypeMinute:
		time.Sleep(time.Duration(interval) * time.Minute)
	default:
		return false, activity.NewError("Unsupported Interval Type. Supported Types- [Millisecond, Second, Minute]", "", nil)
	}

	context.Logger().Debugf("Sleep activity completed")
	return true, nil
}
