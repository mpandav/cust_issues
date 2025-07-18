package sleep

import (
	"github.com/project-flogo/core/data/coerce"
)

type Input struct {
	IntervalTypeSetting string `md:"IntervalTypeSetting"`
	IntervalSetting     int64  `md:"IntervalSetting"`
	IntervalType        string `md:"Interval Type"`
	Interval            int64  `md:"Interval"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"IntervalTypeSetting": i.IntervalTypeSetting,
		"IntervalSetting":     i.IntervalSetting,
		"Interval Type":       i.IntervalType,
		"Interval":            i.Interval,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.IntervalTypeSetting, err = coerce.ToString(values["IntervalTypeSetting"])
	if err != nil {
		return err
	}

	i.IntervalSetting, err = coerce.ToInt64(values["IntervalSetting"])
	if err != nil {
		return err
	}

	i.IntervalType, err = coerce.ToString(values["Interval Type"])
	if err != nil {
		return err
	}

	i.Interval, err = coerce.ToInt64(values["Interval"])
	if err != nil {
		return err
	}

	return nil
}
