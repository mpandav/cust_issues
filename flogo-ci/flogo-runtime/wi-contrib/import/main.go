package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	feEngine "github.com/tibco/wi-contrib/engine"
	"github.com/tibco/wi-contrib/engine/cmd"
	"github.com/tibco/wi-contrib/environment"
)

var cfgJson string

func main() {

	waitTibtunnelReady()

	var e engine.Engine
	var err error

	err = feEngine.StartProfile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// This for app command which avalible in base engine as well as binary
	if IsBaseEngine() {
		feEngine.AddSharedData("isBaseEngine", true)
	}

	if len(os.Args) >= 2 && !feEngine.HasProfile() {
		err := cmd.HandleCommandline(os.Args[1:], cfgJson)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {

		if IsBaseEngine() {
			if len(os.Args) < 2 || feEngine.HasProfile() {
				e, err = feEngine.CreateEngine(cfgJson)
				if err != nil {
					log.RootLogger().Error(err)
					os.Exit(1)
				}

				appJson, err := feEngine.GetAppJsonFromCurrentDir()
				if err != nil {
					log.RootLogger().Error(err)
					os.Exit(1)
				}
				cmd.PrintContributions(appJson)
			}

		}

		/** %%AWS_MARKETPLACE%% - to be removed at build time

		// Register usage with AWS Metering service
		registerUsage()

		%%AWS_MARKETPLACE%% - to be removed at build time **/

		startTime := time.Now()
		println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
		println("     Starting TIBCO Flogo® Runtime")
		println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")

		e, err = feEngine.CreateEngine(cfgJson)
		if err != nil {
			log.RootLogger().Error(err)
			os.Exit(1)
		}

		err = feEngine.StartEngine(e)
		if err != nil {
			log.RootLogger().Error(err)
			os.Exit(1)
		}

		err = feEngine.StartOthers(e, startTime)
		if err != nil {
			log.RootLogger().Error(err)
			os.Exit(1)
		}

	}
}

func waitTibtunnelReady() {
	//Only do for TCI env
	if environment.IsTCIEnv() {
		//Make sure access key configured and not in Flogo Tester User app
		testerAppEnv, ok := os.LookupEnv("FLOGO_TESTER_USER_APP")
		if ok && len(testerAppEnv) > 0 {
			if enabled, _ := coerce.ToBool(testerAppEnv); enabled {
				return
			}
		}
		accessKey, ok := os.LookupEnv("TIBCO_INTERNAL_TCI_TIBTUNNEL_ACCESS_KEY")
		if ok && len(accessKey) > 0 {
			waitSecs := getWaitTunnelReadySec()
			log.RootLogger().Infof("Application [%s] is configured with Accesskey [%s], waiting [%d] seconds for connectivity with TIBCO Cloud™ - Proxy Agent", environment.GetTCIAppName(), accessKey, waitSecs)
			//Sleep 5s
			time.Sleep(time.Duration(waitSecs) * time.Second)

			//log.RootLogger().Infof("Application [%s] is configured with Accesskey [%s]. Application will not start if connection with TIBCO Cloud™ - Proxy Agent is not established within next 3 minutes.", environment.GetTCIAppName(), accessKey)
			//Check to see if port is expose in env var
			//p, ok := os.LookupEnv("TIBCO_INTERNAL_CONTAINER_AGENT_HTTP_PORT")
			//if ok && len(p) > 0 {
			//	//Check tibtunne
			//	var done = make(chan bool, 1)
			//	tunnelUrl := "http://localhost:" + p + "/v1/tunnel/status"
			//	go checkTibtunnel(tunnelUrl, done)
			//	for {
			//		select {
			//		case <-done:
			//			log.RootLogger().Info("Successfully connected to TIBCO Cloud™ - Proxy Agent.")
			//			return
			//		}
			//	}
			//}
		}
	}
}

func getWaitTunnelReadySec() int {
	v, ok := os.LookupEnv("TCI_FLOGO_TIBTUNNEL_WAIT_DURATION")
	if !ok {
		return 20
	}
	i, _ := strconv.Atoi(v)
	if i > 0 {
		return i
	}
	return 20
}

func checkTibtunnel(url string, done chan bool) {
	var attempts = 0
	var maxDelay = 15
	client := http.DefaultClient
	for running := true; running; {
		dur := time.Duration(attempts) * time.Second
		if dur > time.Duration(maxDelay)*time.Second {
			dur = time.Duration(maxDelay) * time.Second
		}
		log.RootLogger().Info("Waiting for connection from TIBCO Cloud™ - Proxy Agent....")
		res, err := client.Get(url)
		if err != nil {
			attempts = attempts + 2
			log.RootLogger().Warn("Checking TIBCO Cloud™ - Proxy Agent status failed, due to [%s]", err.Error())
			time.Sleep(dur)
			continue
		}
		if res.StatusCode == 200 {
			b, _ := ioutil.ReadAll(res.Body)
			data := make(map[string]interface{})
			err = json.Unmarshal(b, &data)
			if err != nil {
				continue
			}
			connected, _ := coerce.ToBool(data["isConnected"])
			if connected {
				running = false
				done <- true
				return
			}
		}

		time.Sleep(dur)
		attempts = attempts + 2
	}

}
