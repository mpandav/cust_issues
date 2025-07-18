package metrics

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"

	"github.com/tibco/wi-contrib/httpservice"
)

func init() {
	httpservice.RegisterHandler("/debug/vars", expvar.Handler())
	setDefualtStats()

	// Runtime profiling
	httpservice.RegisterHandler("/debug/pprof/", http.HandlerFunc(pprof.Index))
	httpservice.RegisterHandler("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	httpservice.RegisterHandler("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	httpservice.RegisterHandler("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	httpservice.RegisterHandler("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))

}

func setDefualtStats() {
	// Set default metrix
	expvar.NewInt("processid").Set(int64(os.Getpid()))
	expvar.NewInt("cpus").Set(int64(runtime.NumCPU()))
	expvar.NewInt("goroutines").Set(int64(runtime.NumGoroutine()))
	expvar.NewString("version").Set(runtime.Version())
}
