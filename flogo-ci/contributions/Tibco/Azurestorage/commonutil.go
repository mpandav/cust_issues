package commonutil

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/project-flogo/core/support/log"

	azstorage "github.com/tibco/wi-azstorage/src/app/Azurestorage/connector/connection"
)

// ReadSeekCloser is a composite interface that groups io.Reader, io.Seeker, and io.Closer.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// NewReadSeekCloserFromBytes creates a ReadSeekCloser from a byte slice.
func NewReadSeekCloserFromBytes(data []byte) ReadSeekCloser {
	return &readSeekCloser{bytes.NewReader(data)}
}

type readSeekCloser struct {
	*bytes.Reader
}

func (rsc *readSeekCloser) Close() error {
	return nil // No-op, as there's nothing to close for a byte slice
}

func CheckForRenewal(err error, activityLog log.Logger, conn *azstorage.AzStorageSharedConfigManager) error {
	if !strings.Contains(err.Error(), "403") {
		return err
	}
	if conn.Config.AuthMode != azstorage.AuthModeSAS || !conn.Config.RegenerateFlag {
		return err
	}

	activityLog.Errorf("Error while executing the activity: %s", err.Error())
	activityLog.Infof("Retrying the operation with new SAS token")
	err = conn.RenewToken()
	if err != nil {

		return fmt.Errorf("Error while renewing the token: %s", err.Error())
	}
	return nil
}
