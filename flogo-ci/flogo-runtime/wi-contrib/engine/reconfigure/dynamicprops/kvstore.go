package dynamicprops

import (
	"fmt"
	"os"

	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
)

const (
	ResolverName = "dynamicprops"
)

var dynamicAppPropResolverLogger = log.ChildLogger(log.RootLogger(), fmt.Sprintf("app-props-%s-resolver", ResolverName))

func init() {
	if property.IsPropertyReconfigureEnabled() {
		evnAppProp := os.Getenv(engine.EnvAppPropertyResolvers)
		if evnAppProp == "" {
			os.Setenv(engine.EnvAppPropertyResolvers, ResolverName)
		} else {
			os.Setenv(engine.EnvAppPropertyResolvers, fmt.Sprintf("%s,%s", ResolverName, evnAppProp))
		}
		dynamicAppPropResolver := &DynamicAppPropResolver{}
		dynamicAppPropResolver.store = make(map[string]interface{})
		property.RegisterPropertyResolver(dynamicAppPropResolver)
		dynamicAppPropResolverLogger.Info("Dynamic app property override resolver is registered")
	}

}

type DynamicAppPropResolver struct {
	store map[string]interface{}
}

func (resolver *DynamicAppPropResolver) Name() string {
	return ResolverName
}

func (resolver *DynamicAppPropResolver) LookupValue(toResolve string) (interface{}, bool) {
	dynamicAppPropResolverLogger.Debugf("Resolving key - %s , with resolver %s", toResolve, ResolverName)
	if val, ok := resolver.store[toResolve]; ok {
		return val, true
	}
	return nil, false
}

func (resolver *DynamicAppPropResolver) UpdateStore(properties map[string]interface{}) {
	//Cleaning old values from store, so with new request body, only new values will be available
	resolver.store = make(map[string]interface{})
	for key, value := range properties {
		resolver.store[key] = value
	}
	dynamicAppPropResolverLogger.Debugf("Updated %s store", ResolverName)
}
