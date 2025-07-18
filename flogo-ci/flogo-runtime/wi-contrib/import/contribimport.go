package main

import (
	_ "github.com/tibco/wi-contrib/cloudelements"
	_ "github.com/tibco/wi-contrib/connection/aws"
	_ "github.com/tibco/wi-contrib/connection/generic"
	_ "github.com/tibco/wi-contrib/connection/google"
	_ "github.com/tibco/wi-contrib/engine/eventflowcontrol"
	_ "github.com/tibco/wi-contrib/environment"
	_ "github.com/tibco/wi-contrib/integrations/aws"
	_ "github.com/tibco/wi-contrib/integrations/azure"
	_ "github.com/tibco/wi-contrib/integrations/consul"
	_ "github.com/tibco/wi-contrib/integrations/hybridmonitoring"
	_ "github.com/tibco/wi-contrib/integrations/hybridmonitoring/types"
	_ "github.com/tibco/wi-contrib/integrations/jaeger"
	_ "github.com/tibco/wi-contrib/integrations/k8ssecret"
	_ "github.com/tibco/wi-contrib/integrations/monitoring"
	_ "github.com/tibco/wi-contrib/integrations/opentelemetry"
	_ "github.com/tibco/wi-contrib/integrations/prometheus"
	_ "github.com/tibco/wi-contrib/integrations/springcloud"
	_ "github.com/tibco/wi-contrib/metrics"

	//TODO Maybe move to correct folder but for now init json schema and expr
	_ "github.com/project-flogo/core/app/propertyresolver"
	_ "github.com/project-flogo/core/data/expression/script"
	_ "github.com/project-flogo/core/data/schema/json"
	_ "github.com/tibco/wi-contrib/engine/stateful"
	_ "github.com/tibco/wi-contrib/engine/unittest"
)
