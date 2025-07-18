package timer

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/carlescere/scheduler"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
	"github.com/robfig/cron/v3"
)

var triggerMd = trigger.NewMetadata(&HandlerSettings{})

func init() {
	_ = trigger.Register(&TimerTrigger{}, &TimerFactory{})
}

type TimerTrigger struct {
	timerJobs []*TimerJob
	config    *trigger.Config
	name      string
	timers    map[string]*scheduler.Job
	logger    log.Logger
}
type TimerJob struct {
	handler      trigger.Handler
	startTime    string
	repeating    bool
	timeInterval int
	intervalUnit string
	schedulerOpt string
	cronExp      string
	cronJob      *cron.Cron
	addDelay     bool
}

// TimerFactory Timer Trigger factory
type TimerFactory struct {
}

// Metadata implements trigger.Trigger.Metadata
func (t *TimerFactory) Metadata() *trigger.Metadata {
	return triggerMd
}

//New Creates a new trigger instance for a given id
func (t *TimerFactory) New(config *trigger.Config) (trigger.Trigger, error) {
	return &TimerTrigger{name: config.Id, config: config}, nil
}

// Init implements ext.Trigger.Init
func (t *TimerTrigger) Initialize(ctx trigger.InitContext) error {
	t.logger = ctx.Logger()

	t.logger.Debugf("Initializing %s", t.config.Id)

	for _, handler := range ctx.GetHandlers() {

		handlerSettings := &HandlerSettings{}

		err := metadata.MapToStruct(handler.Settings(), handlerSettings, true)

		if err != nil {
			return fmt.Errorf("error - %s", err.Error())
		}
		timerJob := &TimerJob{
			handler:      handler,
			startTime:    handlerSettings.StartTime,
			repeating:    handlerSettings.Repeating,
			timeInterval: handlerSettings.TimeInterval,
			intervalUnit: handlerSettings.IntervalUnit,
			schedulerOpt: handlerSettings.SchedulerOpt,
			cronExp:      handlerSettings.CronExp,
			addDelay:     handlerSettings.AddDelay,
		}
		t.logger.Debugf("timerJob: %+v\n", timerJob)
		t.timerJobs = append(t.timerJobs, timerJob)

	}
	return nil
}

func (t *TimerTrigger) Start() error {

	t.logger.Infof("Starting %s...", t.name)
	t.timers = make(map[string]*scheduler.Job)

	for _, timerJob := range t.timerJobs {

		t.logger.Debugf("Processing Handler: %s", timerJob.handler.Name())

		if timerJob.schedulerOpt == "Cron Job" {
			err := t.scheduleCronJob(timerJob)
			if err != nil {
				return err
			}
		} else {
			if timerJob.repeating {
				err := t.scheduleRepeating(timerJob)
				if err != nil {
					return err
				}
			} else {
				err := t.scheduleOnce(timerJob)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Stop implements ext.Trigger.Stop
func (t *TimerTrigger) Stop() error {

	t.logger.Debugf("Stopping %s", t.name)

	for k, v := range t.timers {
		t.logger.Debug("Stopping timer for: ", k)
		v.Quit <- true
	}

	for _, timerJob := range t.timerJobs {
		if timerJob.cronJob != nil {
			t.logger.Debug("Stopping timer with cron expression: ", timerJob.cronExp)
			timerJob.cronJob.Stop()
		}

	}
	return nil
}

//commenting pause resume since its not implemented completely for FlowLimit
// func (t *TimerTrigger) Pause() error {
// 	return nil
// }

// func (t *TimerTrigger) Resume() error {
// 	return nil
// }

func (t *TimerTrigger) scheduleOnce(tj *TimerJob) error {
	t.logger.Debug("Scheduling a run once job")

	seconds, err := getInitialStartInSeconds(tj.startTime, t.logger)
	if err != nil {
		return err
	}
	t.logger.Debug("Seconds till trigger fires: ", seconds)

	var timerJob *scheduler.Job

	if seconds != 0 {
		t.logger.Infof("Start time is configured, executing handler after %d seconds", seconds)
		timerJob = scheduler.Every(int(seconds))
		if timerJob == nil {
			err := fmt.Errorf("Failed to create timer job")
			t.logger.Error(err.Error())
			return err
		}
	} else {
		if tj.startTime != "" {
			t.logger.Info("Start time has already passed, executing handler now")
		} else {
			t.logger.Info("Start time is not configured, executing handler now")
		}
	}

	fn := func() {
		t.logger.Debug("Starting \"Once\" timer process")
		tags := make(map[string]string, 1)
		tags["repeating"] = "false"
		evtContext := trigger.AppendEventDataToContext(context.Background(), tags)
		_, err := tj.handler.Handle(evtContext, nil)
		if err != nil {
			t.logger.Error("Error starting action: ", err.Error())
		}
		if timerJob != nil {
			timerJob.Quit <- true
		}
	}

	if seconds == 0 {
		//Run now
		go fn()
	} else {
		timerJob, err := timerJob.Seconds().NotImmediately().Run(fn)
		if err != nil {
			t.logger.Error("Error scheduleOnce flow err: ", err.Error())
			return err
		}
		t.timers["r:"+tj.handler.Name()] = timerJob

	}

	return nil
}

func (t *TimerTrigger) scheduleRepeating(tj *TimerJob) error {
	t.logger.Debug("Scheduling a repeating job")

	seconds, err := getInitialStartInSeconds(tj.startTime, t.logger)
	if err != nil {
		return err
	}
	t.logger.Debug("Seconds till trigger fires: ", seconds)

	fn := func() {
		t.logger.Debug("Starting \"Repeating\" timer process")
		go func() { //Next scheduler for time interval
			tags := make(map[string]string, 3)
			tags["repeating"] = "true"
			tags["interval"] = strconv.Itoa(tj.timeInterval)
			tags["interval_unit"] = tj.intervalUnit
			evtContext := trigger.AppendEventDataToContext(context.Background(), tags)
			_, err := tj.handler.Handle(evtContext, nil)
			if err != nil {
				t.logger.Error("Error starting flow: ", err.Error())
			}
		}()
	}

	if seconds > 0 {
		//Start from specific time
		timerJobForStartTime := scheduler.Every(seconds).Seconds()

		fn2 := func() { //Initial scheduler for start time
			t.logger.Debug("Starting \"Repeating\" timer process.")
			go func() {

				if timerJobForStartTime != nil {
					timerJobForStartTime.Quit <- true
				}

				timerJobForInterval, err := t.scheduleJobEverySecond(tj, tj.addDelay, fn)
				if err != nil {
					t.logger.Error("Error occured : ", err.Error())
				}
				t.timers["r:"+tj.handler.Name()] = timerJobForInterval

			}()
		}
		t.logger.Infof("Start time is configured, executing handler after %d seconds", seconds)
		timerJobForStartTime, err := timerJobForStartTime.NotImmediately().Run(fn2) //Actual run for StartTime using 'seconds'
		if err != nil {
			t.logger.Error("Error scheduling delayed start repeating timer: ", err.Error())
		}
	} else {
		if tj.startTime != "" {
			t.logger.Info("Start time has already passed, executing handler now")
		} else {
			t.logger.Info("Start time is not configured, executing handler now")
		}
		timerJobForInterval, err := t.scheduleJobEverySecond(tj, tj.addDelay, fn)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		t.timers["r:"+tj.handler.Name()] = timerJobForInterval
	}
	return nil

}

func (t *TimerTrigger) scheduleJobEverySecond(tj *TimerJob, notImmediately bool, fn func()) (*scheduler.Job, error) {

	var interval int = tj.timeInterval
	var err error

	if interval < 0 {
		err = fmt.Errorf("Invalid Time Interval [%d]. Must not be a negative value. ", interval)
		t.logger.Error(err.Error())
		return nil, err
	}

	intervalUnit := tj.intervalUnit

	switch intervalUnit {
	case "Second":
		interval = interval * 1
	case "Minute":
		interval = interval * 60
	case "Hour":
		interval = interval * 60 * 60
	case "Day":
		interval = interval * 60 * 60 * 24
	case "Week":
		interval = interval * 60 * 60 * 24 * 7
	default:
		return nil, fmt.Errorf("Invalid interval unit %s", intervalUnit)
	}

	t.logger.Debugf("Time interval: %d seconds ", interval)
	// schedule repeating
	var timerJob *scheduler.Job
	if notImmediately {
		timerJob, err = scheduler.Every(interval).Seconds().NotImmediately().Run(fn)
		if err != nil {
			t.logger.Error("Error scheduleRepeating (repeat seconds) flow err: ", err.Error())
		}
	} else {
		timerJob, err = scheduler.Every(interval).Seconds().Run(fn)
		if err != nil {
			t.logger.Error("Error scheduleRepeating (repeat seconds) flow err: ", err.Error())
		}
	}

	if timerJob == nil {
		t.logger.Error("timerJob is nil")
	}
	t.timers["r:"+tj.handler.Name()] = timerJob

	return timerJob, nil
}

func (t *TimerTrigger) scheduleCronJob(tj *TimerJob) error {
	//Function calling the correspoing flow in trigger
	fn := func() {
		t.logger.Debug("Starting Cron job")
		tags := make(map[string]string, 1)
		tags["repeating"] = "false"
		evtContext := trigger.AppendEventDataToContext(context.Background(), tags)
		_, err := tj.handler.Handle(evtContext, nil)
		if err != nil {
			t.logger.Error("Error starting action: ", err.Error())
		}
	}

	c := cron.New()
	_, err := c.AddFunc(tj.cronExp, fn)
	//Check to validate the cron expression
	if err != nil {
		t.logger.Error("Invalid cron expression ", tj.cronExp)
		return err
	}
	tj.cronJob = c
	tj.cronJob.Start()

	return nil
}

func getInitialStartInSeconds(startTime string, logger log.Logger) (int, error) {

	layout := time.RFC3339

	if startTime == "" {
		return 0, nil
	}

	// idx := strings.LastIndex(startTime, "Z")
	// timeZone := startTime[idx+1:]
	// logger.Debug("Time Zone: ", timeZone)
	// startTime = strings.TrimSuffix(startTime, timeZone)
	logger.Debug("Start Time: ", startTime)

	// is timezone negative
	// isNegative := strings.HasPrefix(timeZone, "-")
	// remove sign
	// timeZone = strings.TrimPrefix(timeZone, "-")

	triggerTime, err := time.Parse(layout, startTime)
	if err != nil {
		logger.Error("Failed to parse time due to error: ", err.Error())
		return 0, err
	}
	logger.Debug("Time parsed from settings: ", triggerTime)

	// var hour int
	// var minutes int

	// sliceArray := strings.Split(timeZone, ":")
	// if len(sliceArray) != 2 {
	// 	logger.Error("Time zone has wrong format: ", timeZone)
	// } else {
	// 	hour, _ = strconv.Atoi(sliceArray[0])
	// 	minutes, _ = strconv.Atoi(sliceArray[1])

	// 	logger.Debug("Duration hour: ", time.Duration(hour)*time.Hour)
	// 	logger.Debug("Duration minutes: ", time.Duration(minutes)*time.Minute)
	// }

	// hours, _ := strconv.Atoi(timeZone)
	// logger.Debug("hours: ", hours)
	// if isNegative {
	// 	logger.Debug("Adding to triggerTime")
	// 	triggerTime = triggerTime.Add(time.Duration(hour) * time.Hour)
	// 	triggerTime = triggerTime.Add(time.Duration(minutes) * time.Minute)
	// } else {
	// 	logger.Debug("Subtracting to triggerTime")
	// 	triggerTime = triggerTime.Add(time.Duration(hour * -1))
	// 	triggerTime = triggerTime.Add(time.Duration(minutes))
	// }

	currentTime := time.Now().UTC()
	logger.Debug("Current time: ", currentTime)
	//logger.Debug("Setting start time: ", triggerTime)
	duration := time.Since(triggerTime)
	durSeconds := duration.Seconds()
	if durSeconds < 0 {
		//Future date
		return int(math.Abs(durSeconds)), nil
	} else {
		// Past date
		return 0, nil
	}
}
