package poller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
	"github.com/radovskyb/watcher"
)

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

// Trigger File Poller struct
type Trigger struct {
	id             string
	logger         log.Logger
	pollerWatchers []*PollerWatcher
}

type PollerWatcher struct {
	logger   log.Logger
	watch    *watcher.Watcher
	settings *HandlerSettings
	handler  trigger.Handler
}

type (
	EventOutput struct {
		Action       string        `json:"action"`
		FileMetadata *FileMetadata `json:"fileMetadata"`
	}

	FileMetadata struct {
		FullPath string `json:"fullPath"`
		Name     string `json:"name"`
		OldPath  string `json:"oldPath"`
		Size     int64  `json:"size"`
		Mode     string `json:"mode"`
		ModTime  string `json:"modTime"`
		IsDir    bool   `json:"isDir"`
	}
)

func (t *Trigger) Initialize(ctx trigger.InitContext) (err error) {

	t.logger = ctx.Logger()

	// Init handlers
	for _, handler := range ctx.GetHandlers() {
		handlerSettings := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSettings, true)
		if err != nil {
			return err
		}

		t.logger.Infof("Input Handler Settings - [Polling Directory: %s], [Include Sub-Directories: %t], [File Filter: %s], [Polling Interval : %d], [Mode : %s] , [Poll File Events : %s]", handlerSettings.PollingDirectory, handlerSettings.Recursive, handlerSettings.FileFilter, handlerSettings.PollingInterval, handlerSettings.Mode, handlerSettings.FileEvents)

		pollerWatcher := &PollerWatcher{}
		pollerWatcher.logger = t.logger
		pollerWatcher.settings = handlerSettings
		pollerWatcher.handler = handler
		pollerWatcher.watch = watcher.New()

		t.pollerWatchers = append(t.pollerWatchers, pollerWatcher)
	}
	return nil
}

func (t *Trigger) Start() error {
	t.logger.Infof("In File Poller Start method")
	for _, pollerWatcher := range t.pollerWatchers {

		if strings.Contains(pollerWatcher.settings.PollingDirectory, "..") {
			errMessage := fmt.Sprintf("\"..\" is not allowed in the field \"Polling Directory\"")
			return fmt.Errorf(errMessage)
		}

		rootDir := os.Getenv("FLOGO_FILES_ROOT_DIR")
		if rootDir != "" && !strings.HasPrefix(pollerWatcher.settings.PollingDirectory, "/") {
			err := os.MkdirAll(rootDir, os.ModePerm)
			if err != nil {
				return err
			}
			pollerWatcher.settings.PollingDirectory = filepath.Join(rootDir, pollerWatcher.settings.PollingDirectory)
		}

		// Convert the stringified array to a slice of string
		var events []string
		err := json.Unmarshal([]byte(pollerWatcher.settings.FileEvents), &events)
		if err != nil {
			t.logger.Errorf("Failed to unmarshal due to error - %s", err.Error())
			return err
		}

		ops := make([]watcher.Op, 0, 5)
		for _, value := range events {
			switch value {
			case "Create":
				ops = append(ops, watcher.Create)

			case "Write":
				ops = append(ops, watcher.Write)

			case "Rename":
				ops = append(ops, watcher.Rename)

			case "Remove":
				ops = append(ops, watcher.Remove)

			case "Move":
				ops = append(ops, watcher.Move)
			}
		}

		pollerWatcher.watch.FilterOps(ops...)

		// Only files that match the regular expression during file listings will be watched.
		if pollerWatcher.settings.FileFilter != "" {
			r := regexp.MustCompile(pollerWatcher.settings.FileFilter)
			pollerWatcher.watch.AddFilterHook(watcher.RegexFilterHook(r, false))
		}

		if !pollerWatcher.settings.Recursive {
			// Watch the folder for changes
			if err := pollerWatcher.watch.Add(pollerWatcher.settings.PollingDirectory); err != nil {
				t.logger.Errorf("Failed to add polling directory to watch for file events - %s", err.Error())
				return err
			}
		} else {
			// Watch the folder recursively for changes
			if err := pollerWatcher.watch.AddRecursive(pollerWatcher.settings.PollingDirectory); err != nil {
				t.logger.Errorf("Failed to add polling directory to watch recursively for file events - %s", err.Error())
				return err
			}
		}

		go func(pw *PollerWatcher) {
			if err := pw.watch.Start(time.Millisecond * time.Duration(pw.settings.PollingInterval)); err != nil {
				t.logger.Errorf("Failed to start the polling cycle - %s", err.Error())
				panic(err.Error())
			}
		}(pollerWatcher)

		go pollerWatcher.ListenEvents(pollerWatcher.settings.Mode)
	}

	return nil
}

func (pollerWatcher *PollerWatcher) ListenEvents(mode string) {
	for {
		select {
		case event := <-pollerWatcher.watch.Event:
			if accept(mode, event.IsDir()) {
				output := &Output{}

				var fileMetadata FileMetadata
				fileMetadata.FullPath = event.Path
				fileMetadata.Name = event.Name()
				fileMetadata.OldPath = event.OldPath
				fileMetadata.Size = event.Size()
				fileMetadata.Mode = event.Mode().String()
				fileMetadata.ModTime = event.ModTime().String()
				fileMetadata.IsDir = event.IsDir()

				var eventOutput EventOutput
				eventOutput.Action = event.Op.String()
				eventOutput.FileMetadata = &fileMetadata

				//set the output
				reqBodyBytes := new(bytes.Buffer)
				json.NewEncoder(reqBodyBytes).Encode(eventOutput)
				err := json.Unmarshal(reqBodyBytes.Bytes(), &output.Output)
				if err != nil {
					pollerWatcher.logger.Errorf("Failed to unmarshal due to error - %s", err.Error())
					panic(err.Error())
				}

				eventId := fmt.Sprintf("%s#%s#%d", pollerWatcher.settings.PollingDirectory, pollerWatcher.settings.FileFilter, pollerWatcher.settings.PollingInterval)
				ctx := context.Background()
				if eventId != "" {
					ctx = trigger.NewContextWithEventId(ctx, eventId)
				}

				_, err = pollerWatcher.handler.Handle(ctx, output)
				if err != nil {
					pollerWatcher.logger.Errorf("Failed to set output, error : %s", err.Error())
					return
				}
			}
		case err := <-pollerWatcher.watch.Error:
			pollerWatcher.logger.Errorf("Poller watch error : %s", err.Error())
		case <-pollerWatcher.watch.Closed:
			return
		}
	}
}

// Stop implements util.Managed.Stop
func (t *Trigger) Stop() error {
	t.logger.Infof("In File Poller Stop method")
	for _, pollerWatcher := range t.pollerWatchers {
		pollerWatcher.watch.Close()
	}
	return nil
}

func accept(mode string, isDir bool) bool {
	return (mode == "Only Directories" && isDir) || (mode == "Only Files" && !isDir) || (mode == "Files and Directories")
}
