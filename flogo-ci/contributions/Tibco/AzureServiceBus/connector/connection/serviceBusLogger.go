package azureservicebusconnection

import (
	"strings"

	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/project-flogo/core/support/log"
)

func InitLogger() {

	var logCache = log.ChildLogger(log.RootLogger(), "azureServiceBus-connection")

	azlog.SetListener(func(event azlog.Event, s string) {
		if strings.Contains(s, "Recover connection failure") {
			//logCache.Errorf("[%s] %s\n", event, s)
			logCache.Info("Recover connection failure, retrying ...  ")
		}
	})

	// pick the set of events to log
	azlog.SetEvents(
		// EventConn is used whenever we create a connection or any links (that is, receivers or senders).
		azservicebus.EventConn,
		// // EventAuth is used when we're doing authentication/claims negotiation.
		// azservicebus.EventAuth,
		// // EventReceiver represents operations that happen on receivers.
		// azservicebus.EventReceiver,
		// // EventSender represents operations that happen on senders.
		// azservicebus.EventSender,
		// // EventAdmin is used for operations in the azservicebus/admin.Client
		// azservicebus.EventAdmin,
	)
}
