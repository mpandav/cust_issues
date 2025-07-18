package httpservice

import (
	"fmt"
	"net/http"
	"sync"
)

var mangementHandlers = make(map[string]http.Handler)
var log sync.Mutex

func RegisterHandler(name string, handler http.Handler) {
	_, ok := mangementHandlers[name]
	if ok {
		panic(fmt.Sprintf("Management handler [%s] already registed", name))
	}

	log.Lock()
	mangementHandlers[name] = handler
	log.Unlock()
}

func HasHandler() bool {
	return mangementHandlers != nil && len(mangementHandlers) > 0
}

func AllHandlers() map[string]http.Handler {
	return mangementHandlers
}
