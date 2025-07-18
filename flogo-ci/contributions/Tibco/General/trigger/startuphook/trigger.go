package startuphook

import (
	"context"
	"strconv"

	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

type HandlerSettings struct {
}

var triggerMd = trigger.NewMetadata(&HandlerSettings{})

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

// Trigger REST trigger struct
type Trigger struct {
	id              string
	startupHandlers []startupHandler
	logger          log.Logger
}

type startupHandler struct {
	handler trigger.Handler
	name    string
}

func (t *Trigger) Initialize(ctx trigger.InitContext) error {

	// Init handlers
	t.logger = ctx.Logger()
	for i, handler := range ctx.GetHandlers() {
		name := t.id + "-" + handler.Name()
		if name == "" {
			name = t.id + "-handler-" + strconv.Itoa(i)
		}
		t.startupHandlers = append(t.startupHandlers, startupHandler{handler: handler, name: name})
	}
	return nil
}

func (t *Trigger) Start() error {
	t.logger.Infof("Trigger '%s' started", t.id)
	return nil
}

// Stop implements util.Managed.Stop
func (t *Trigger) Stop() error {
	t.logger.Infof("Trigger '%s' stopped", t.id)
	return nil
}

func (t *Trigger) OnStartup() error {
	t.logger.Infof("Executing app startup handlers for trigger '%s'", t.id)
	for _, sHandler := range t.startupHandlers {
		_, err := sHandler.handler.Handle(context.Background(), nil)
		if err != nil {
			sHandler.handler.Logger().Debugf("Error in app startup handler [%s]: %s", sHandler.name, err.Error())
			return err
		}
	}
	return nil
}

func (t *Trigger) OnShutdown() error {
	return nil
}
