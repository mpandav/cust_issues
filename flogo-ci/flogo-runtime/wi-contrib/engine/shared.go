package engine

import "fmt"

var sharedData = make(map[string]interface{})

func AddSharedData(key string, value interface{}) error {
	_, ok := sharedData[key]
	if ok {
		return fmt.Errorf("kye [%s] already registered", key)
	}
	sharedData[key] = value
	return nil
}

func GetSharedData(key string) interface{} {
	return sharedData[key]
}
