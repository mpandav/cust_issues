package error

import (
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

type HandlerSettings struct {
}

var triggerMd = trigger.NewMetadata(&HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	return &Trigger{id: config.Id}, nil
}

// Error trigger struct
type Trigger struct {
	id     string
	logger log.Logger
}

func (t *Trigger) Initialize(ctx trigger.InitContext) error {
	return nil
}

func (t *Trigger) Start() error {
	t.logger.Infof("Error handler '%s' is started", t.id)
	return nil
}

func (t *Trigger) Stop() error {
	return nil
}
