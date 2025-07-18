package timer

type HandlerSettings struct {
	StartTime    string `md:"Start Time"`
	Repeating    bool   `md:"Repeating"`
	TimeInterval int    `md:"Time Interval"`
	IntervalUnit string `md:"Interval Unit"`
	SchedulerOpt string `md:"Scheduler Options"`
	CronExp      string `md:"Cron Expression"`
	AddDelay     bool   `md:"Delayed Start"`
}
