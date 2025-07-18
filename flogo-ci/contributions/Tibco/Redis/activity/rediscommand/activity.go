package rediscommand

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	redis "github.com/redis/go-redis/v9"
	"github.com/tibco/wi-redis/src/app/Redis"
	redisconnection "github.com/tibco/wi-redis/src/app/Redis/connector/connection"
)

// Oss upgrade--
var RedislogCache = log.ChildLogger(log.RootLogger(), "redis.activity.rediscommand")
var activityMd = activity.ToMetadata(&Input{}, &Output{}) //duplicate input type existing over here need to chhange it.

func init() {
	_ = activity.Register(&RedisCommandActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &RedisCommandActivity{}, nil
}

type RedisCommandActivity struct {
}

func (*RedisCommandActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *RedisCommandActivity) Eval(context activity.Context) (done bool, err error) {

	logger := context.Logger()
	logger.Info("Executing  Redis Command Activity")
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	connection, _ := input.RedisConnection.(*redisconnection.RedisSharedConfigManager)

	connectionUpdated, err := Redis.GetConnection(connection, input.Input)
	if err != nil {
		return false, fmt.Errorf("Error getting Redis connection %s", err.Error())
	}

	command := input.Command
	if command == "" {
		return false, activity.NewError("Redis Command is not configured", "REDIS-OPERATION-ERROR-2001", nil)
	}

	resp, err := ExecuteCommand(connectionUpdated.RedisClient, command, input.Input)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "connection refused") || strings.Contains(strings.ToLower(err.Error()), "server misbehaving") || strings.Contains(strings.ToLower(err.Error()), "network is unreachable") {
			return false, activity.NewRetriableError(fmt.Sprintf("Failed to execute query [%s] on REDIS server due to error - {%s}.", command, err.Error()), "Redis-4001", nil)
		}

		return false, fmt.Errorf("Redis command failed, %s", err.Error())
	}

	respMap := make(map[string]interface{})
	respMap[command+"Response"] = resp

	output := &Output{}
	output.Output = respMap
	logger.Debugf("Output is %s", output)
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	return true, nil

	/*


		outputComplex := &data.ComplexObject{Metadata: "", Value: respMap}

		RedislogCache.Debugf("Output is %s", outputComplex)
		context.SetOutput(outputProperty, outputComplex)
		return true, nil */
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// ExecuteCommand executes command
func ExecuteCommand(client *redis.Client, command string, inputParams map[string]interface{}) (interface{}, error) {
	//log.Info("Executing command: " + command)
	RedislogCache.Info("Executing command: " + command)
	ctx := context.Background()
	switch command {
	case "APPEND":
		result, err := client.Append(ctx, inputParams["key"].(string), inputParams["value"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "GET":
		result, err := client.Get(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "SET":
		key := inputParams["key"].(string)
		val := inputParams["value"]
		var duration time.Duration
		if inputParams["EX"] != nil && ToFloat64(inputParams["EX"]) != 0 {
			duration = time.Duration(ToFloat64(inputParams["EX"])) * time.Second
		} else if inputParams["PX"] != nil && ToFloat64(inputParams["PX"]) != 0 {
			duration = time.Duration(ToFloat64(inputParams["PX"])) * time.Millisecond
		}
		var result interface{}
		var err error
		if inputParams["NX|XX"] != nil {
			if inputParams["NX|XX"].(string) == "NX" {
				result, err = client.SetNX(ctx, key, val, duration).Result()
			} else if inputParams["NX|XX"].(string) == "XX" {
				result, err = client.SetXX(ctx, key, val, duration).Result()
			}
			if result.(bool) {
				result = "OK"
			} else {
				result = nil
			}
		} else {
			result, err = client.Set(ctx, key, val, duration).Result()
		}
		if err != nil {
			return nil, err
		}

		return result, nil
	case "MGET":
		/* Input for MGET:
		 * {
		 * 	 "keys":[{
		 *	 	"key": "testKey1"
		 *	 }]
		 * }
		 *
		 * Output for MGET:
		 * {
		 *	 "keyvalues": [{
		 *		"key": "",
		 *	 	"value": ""
		 *	}]
		 * }
		 */
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			inputkey := v.(map[string]interface{})
			inputKeys[i] = inputkey["key"].(string)
		}
		result, err := client.MGet(ctx, inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"key": inputKeys[i], "value": v}
		}

		return map[string]interface{}{"keyvalues": outputResult}, nil

	case "MSET":
		/* Input for MSET:
		 * {
		 *		"keyvalues": [{
		 *			"key": "",
		 *		 	"value": ""
		 *		}]
		 *	}
		 */
		keyValues := inputParams["keyvalues"].([]interface{})
		inputKeyValues := make([]interface{}, len(keyValues)*2)
		for i, keyval := range keyValues {
			keyvalPair := keyval.(map[string]interface{})
			inputKeyValues[2*i] = keyvalPair["key"]
			inputKeyValues[2*i+1] = keyvalPair["value"]
		}
		result, err := client.MSet(ctx, inputKeyValues...).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "GETSET":
		result, err := client.GetSet(ctx, inputParams["key"].(string), inputParams["value"]).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "STRLEN":
		result, err := client.StrLen(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "GETRANGE":
		result, err := client.GetRange(ctx, inputParams["key"].(string), ToInt64(inputParams["start"]), ToInt64(inputParams["end"])).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "SETRANGE":
		result, err := client.SetRange(ctx, inputParams["key"].(string), ToInt64(inputParams["offset"]), inputParams["value"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "GETDEL":
		result, err := client.GetDel(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "INCR":
		result, err := client.Incr(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "DECR":
		result, err := client.Decr(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "INCRBY":
		IncrByFloat := ToFloat64(inputParams["increment"])
		if IncrByFloat != float64(ToInt64(inputParams["increment"])) {
			return nil, fmt.Errorf("ERR value is not an integer or out of range")
		}

		result, err := client.IncrBy(ctx, inputParams["key"].(string), ToInt64(inputParams["increment"])).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "DECRBY":
		result, err := client.DecrBy(ctx, inputParams["key"].(string), ToInt64(inputParams["decrement"])).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "INCRBYFLOAT":
		result, err := client.IncrByFloat(ctx, inputParams["key"].(string), ToFloat64(inputParams["increment"])).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "MSETNX":
		keyValues := inputParams["keyvalues"].([]interface{})
		inputKeyValues := make([]interface{}, len(keyValues)*2)
		for i, keyval := range keyValues {
			keyvalPair := keyval.(map[string]interface{})
			inputKeyValues[2*i] = keyvalPair["key"].(string)
			inputKeyValues[2*i+1] = keyvalPair["value"]
		}
		tempResult, err := client.MSetNX(ctx, inputKeyValues...).Result()
		if err != nil {
			return nil, err
		}
		var result int
		if tempResult == true {
			result = 1
		} else {
			result = 0
		}
		return result, nil
	case "GETEX":
		key := inputParams["key"].(string)
		var duration time.Duration
		inputDuration := ToInt64(inputParams["time"])
		if inputDuration <= 0 {
			return nil, errors.New("ERR invalid expire time in getex")
		}
		if inputParams["EX|PX|EXAT|PXAT"] == "EX" {
			duration = time.Duration(inputDuration) * time.Second
		} else if inputParams["EX|PX|EXAT|PXAT"] == "PX" {
			duration = time.Duration(inputDuration) * time.Millisecond
		} else if inputParams["EX|PX|EXAT|PXAT"] == "EXAT" {
			inputDuration = inputDuration - (ToInt64(time.Now().Unix()))
			//if difference between given unix time and current time is negative or zero
			// we are sending 1 nanosecond to GETEX considering its the smallest value we can send to expire the key in past
			if inputDuration <= 0 {
				duration = time.Duration(1)
			} else {
				duration = time.Duration(inputDuration) * time.Second
			}
		} else if inputParams["EX|PX|EXAT|PXAT"] == "PXAT" {
			inputDuration = inputDuration - (time.Now().UnixNano() / int64(time.Millisecond))
			if inputDuration <= 0 {
				duration = time.Duration(1)
			} else {
				duration = time.Duration(inputDuration) * time.Millisecond
			}
		}

		result, err := client.GetEx(ctx, key, duration).Result()

		if err != nil {
			return nil, err
		}
		return result, nil
	case "SETNX":
		var duration time.Duration
		result, err := client.SetNX(ctx, inputParams["key"].(string), inputParams["value"], duration).Result()
		if err != nil {
			return nil, err
		}
		return BoolToInt(result), nil
	case "SETEX":
		duration := time.Duration(ToFloat64(inputParams["seconds"])) * time.Second
		// changed from SetEX to SetEx in v9
		result, err := client.SetEx(ctx, inputParams["key"].(string), inputParams["value"], duration).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "PSETEX":
		duration := time.Duration(ToFloat64(inputParams["milliseconds"])) * time.Millisecond

		result, err := client.SetEx(ctx, inputParams["key"].(string), inputParams["value"], duration).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "LPOP":
		var count int = 1
		if inputParams["count"] != nil {
			count = ToInt(inputParams["count"])
		}

		result, err := client.LPopCount(ctx, inputParams["key"].(string), count).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"value": v}
		}
		return map[string]interface{}{"values": outputResult}, nil

	case "RPOP":
		var count int = 1
		if inputParams["count"] != nil {
			count = ToInt(inputParams["count"])
		}
		result, err := client.RPopCount(ctx, inputParams["key"].(string), count).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"value": v}
		}
		return map[string]interface{}{"values": outputResult}, nil

	case "LPUSH":
		inputValuesArray := inputParams["values"].([]interface{})
		inputValues := make([]interface{}, len(inputValuesArray))
		for i, v := range inputValuesArray {
			val := v.(map[string]interface{})
			inputValues[i] = val["value"]
		}
		result, err := client.LPush(ctx, inputParams["key"].(string), inputValues...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "RPUSH":
		inputValuesArray := inputParams["values"].([]interface{})
		inputValues := make([]interface{}, len(inputValuesArray))
		for i, v := range inputValuesArray {
			val := v.(map[string]interface{})
			inputValues[i] = val["value"]
		}
		result, err := client.RPush(ctx, inputParams["key"].(string), inputValues...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "LINDEX":
		result, err := client.LIndex(ctx, inputParams["key"].(string), ToInt64(inputParams["index"])).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "LSET":
		result, err := client.LSet(ctx, inputParams["key"].(string), ToInt64(inputParams["index"]), inputParams["value"]).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "LREM":
		result, err := client.LRem(ctx, inputParams["key"].(string), ToInt64(inputParams["count"]), inputParams["value"]).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "LINSERT":
		result, err := client.LInsert(ctx, inputParams["key"].(string), inputParams["BEFORE|AFTER"].(string), inputParams["pivot"].(string), inputParams["value"]).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "LRANGE":
		result, err := client.LRange(ctx, inputParams["key"].(string), ToInt64(inputParams["start"]), ToInt64(inputParams["stop"])).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"value": v}
		}
		return map[string]interface{}{"values": outputResult}, nil
	case "BLPOP":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.BLPop(ctx, time.Duration(ToInt64(inputParams["timeout"])), inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"element": v}
		}
		return map[string]interface{}{"elements": outputResult}, nil
	case "BRPOP":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.BRPop(ctx, time.Duration(ToInt64(inputParams["timeout"])), inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"element": v}
		}
		return map[string]interface{}{"elements": outputResult}, nil
	case "LPUSHX":
		inputValuesArray := inputParams["values"].([]interface{})
		inputValues := make([]interface{}, len(inputValuesArray))
		for i, v := range inputValuesArray {
			val := v.(map[string]interface{})
			inputValues[i] = val["value"]
		}
		result, err := client.LPushX(ctx, inputParams["key"].(string), inputValues...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "RPUSHX":
		inputValuesArray := inputParams["values"].([]interface{})
		inputValues := make([]interface{}, len(inputValuesArray))
		for i, v := range inputValuesArray {
			val := v.(map[string]interface{})
			inputValues[i] = val["value"]
		}
		result, err := client.RPushX(ctx, inputParams["key"].(string), inputValues...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "LLEN":
		result, err := client.LLen(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "LMOVE":
		result, err := client.LMove(ctx, inputParams["SourceList"].(string), inputParams["DestinationList"].(string), inputParams["SrcPos LEFT|RIGHT"].(string), inputParams["DestPos LEFT|RIGHT"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "BLMOVE":
		result, err := client.BLMove(ctx, inputParams["SourceList"].(string), inputParams["DestinationList"].(string), inputParams["SrcPos LEFT|RIGHT"].(string), inputParams["DestPos LEFT|RIGHT"].(string), time.Duration(ToInt64(inputParams["timeout"]))).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "LPOS":
		var result []int64
		var err error
		var rank64, maxLen64 int64

		if val, ok := inputParams["rank"]; ok {
			rank64 = ToInt64(val)
		} else {
			rank64 = 1
		}
		if val, ok := inputParams["maxlen"]; ok {
			maxLen64 = ToInt64(val)
		} else {
			maxLen64 = 0
		}

		args := redis.LPosArgs{rank64, maxLen64}

		if val, ok := inputParams["count"]; ok {
			result, err = client.LPosCount(ctx, inputParams["key"].(string), inputParams["value"].(string), ToInt64(val), args).Result()
		} else {
			var res int64
			res, err = client.LPos(ctx, inputParams["key"].(string), inputParams["value"].(string), args).Result()
			result = []int64{res}
		}
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"index": v}
		}
		return map[string]interface{}{"indexes": outputResult}, nil
	case "LTRIM":
		result, err := client.LTrim(ctx, inputParams["key"].(string), ToInt64(inputParams["start"]), ToInt64(inputParams["stop"])).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "RPOPLPUSH":
		result, err := client.RPopLPush(ctx, inputParams["SourceList"].(string), inputParams["DestinationList"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "BRPOPLPUSH":
		result, err := client.BRPopLPush(ctx, inputParams["SourceList"].(string), inputParams["DestinationList"].(string), time.Duration(ToInt64(inputParams["timeout"]))).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "SADD":
		inputMembersArray := inputParams["members"].([]interface{})
		inputMembers := make([]interface{}, len(inputMembersArray))
		for i, v := range inputMembersArray {
			val := v.(map[string]interface{})
			inputMembers[i] = val["member"]
		}
		result, err := client.SAdd(ctx, inputParams["key"].(string), inputMembers...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "SREM":
		inputMembersArray := inputParams["members"].([]interface{})
		inputMembers := make([]interface{}, len(inputMembersArray))
		for i, v := range inputMembersArray {
			val := v.(map[string]interface{})
			inputMembers[i] = val["member"]
		}
		result, err := client.SRem(ctx, inputParams["key"].(string), inputMembers...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "SPOP":
		var result []string
		var err error
		if val, ok := inputParams["count"]; ok {
			result, err = client.SPopN(ctx, inputParams["key"].(string), ToInt64(val)).Result()
		} else {
			var res string
			res, err = client.SPop(ctx, inputParams["key"].(string)).Result()
			result = []string{res}
		}
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "SCARD":
		result, err := client.SCard(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "SMEMBERS":
		result, err := client.SMembers(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "SISMEMBER":
		result, err := client.SIsMember(ctx, inputParams["key"].(string), inputParams["member"]).Result()
		if err != nil {
			return nil, err
		}
		if result {
			return 1, nil
		} else {
			return 0, nil
		}

	case "SMOVE":
		tempResult, err := client.SMove(ctx, inputParams["source"].(string), inputParams["destination"].(string), inputParams["member"]).Result()
		if err != nil {
			return nil, err
		}
		var result int
		if tempResult == true {
			result = 1
		} else {
			result = 0
		}
		return result, nil
	case "SDIFF":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.SDiff(ctx, inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "SDIFFSTORE":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.SDiffStore(ctx, inputParams["destination"].(string), inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "SINTER":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.SInter(ctx, inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "SINTERSTORE":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.SInterStore(ctx, inputParams["destination"].(string), inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "SUNION":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.SUnion(ctx, inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "SUNIONSTORE":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.SUnionStore(ctx, inputParams["destination"].(string), inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil
	case "SMISMEMBER":
		inputMembersArray := inputParams["members"].([]interface{})
		inputMembers := make([]interface{}, len(inputMembersArray))
		for i, v := range inputMembersArray {
			val := v.(map[string]interface{})
			inputMembers[i] = val["member"]
		}
		result, err := client.SMIsMember(ctx, inputParams["key"].(string), inputMembers...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": BoolToInt(v)}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "SRANDMEMBER":
		var result []string
		var err error
		if val, ok := inputParams["count"]; ok {
			result, err = client.SRandMemberN(ctx, inputParams["key"].(string), ToInt64(val)).Result()
		} else {
			var res string
			res, err = client.SRandMember(ctx, inputParams["key"].(string)).Result()
			result = []string{res}
		}
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "SSCAN":
		var result []string
		var match string
		var count int64
		cursorSigned := ToInt64(inputParams["cursor"])
		var cursorUnSigned uint64
		//if value provided by user is negative, setting it as zero unsigned value because cursor will never negative
		if cursorSigned <= 0 {
			cursorUnSigned = uint64(0)
		} else {
			cursorUnSigned = uint64(cursorSigned)
		}
		//if count is not set by user default value as per RedisDocs is 10
		if val, ok := inputParams["count"]; ok {
			count = ToInt64(val)
		} else {
			count = int64(10)
		}
		if val, ok := inputParams["match"]; ok {
			match = val.(string)
		} else {
			match = "*"
		}
		result, cursor, err := client.SScan(ctx, inputParams["key"].(string), cursorUnSigned, match, count).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"cursor": cursor, "members": outputResult}, nil

	case "HDEL":
		inputFieldArray := inputParams["fields"].([]interface{})
		inputFields := make([]string, len(inputFieldArray))
		for i, v := range inputFieldArray {
			inputField := v.(map[string]interface{})
			inputFields[i] = inputField["field"].(string)
		}
		result, err := client.HDel(ctx, inputParams["key"].(string), inputFields...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "HEXISTS":
		result, err := client.HExists(ctx, inputParams["key"].(string), inputParams["field"].(string)).Result()
		if err != nil {
			return nil, err
		}
		if result {
			return 1, nil
		} else {
			return 0, nil
		}

	case "HGET":
		result, err := client.HGet(ctx, inputParams["key"].(string), inputParams["field"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "HGETALL":
		result, err := client.HGetAll(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		index := 0
		for i, v := range result {
			outputResult[index] = map[string]interface{}{"field": i, "value": v}
			index++
		}
		return map[string]interface{}{"fieldvalues": outputResult}, nil

	case "HKEYS":
		result, err := client.HKeys(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"field": v}
		}
		return map[string]interface{}{"fields": outputResult}, nil

	case "HLEN":
		result, err := client.HLen(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "HMGET":
		inputFieldArray := inputParams["fields"].([]interface{})
		inputFields := make([]string, len(inputFieldArray))
		for i, v := range inputFieldArray {
			inputfield := v.(map[string]interface{})
			inputFields[i] = inputfield["field"].(string)
		}
		result, err := client.HMGet(ctx, inputParams["key"].(string), inputFields...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"field": inputFields[i], "value": v}
		}

		return map[string]interface{}{"key": inputParams["key"].(string), "fieldvalues": outputResult}, nil

	case "HMSET":
		fieldValues := inputParams["fieldvalues"].([]interface{})
		inputFieldValues := make([]interface{}, len(fieldValues)*2)
		for i, fieldVal := range fieldValues {
			fieldValPair := fieldVal.(map[string]interface{})
			inputFieldValues[2*i] = fieldValPair["field"].(string)
			inputFieldValues[2*i+1] = fieldValPair["value"]
		}
		tempResult, err := client.HMSet(ctx, inputParams["key"].(string), inputFieldValues).Result()
		if err != nil {
			return nil, err
		}
		var result string
		if tempResult == true {
			result = "OK"
		} else {
			result = "failed"
		}
		return result, nil

	case "HSET":

		fieldValues := inputParams["fieldvalues"].([]interface{})
		inputFieldValues := make([]interface{}, len(fieldValues)*2)
		for i, fieldVal := range fieldValues {
			fieldValPair := fieldVal.(map[string]interface{})
			inputFieldValues[2*i] = fieldValPair["field"].(string)
			inputFieldValues[2*i+1] = fieldValPair["value"]
		}
		result, err := client.HSet(ctx, inputParams["key"].(string), inputFieldValues).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "HVALS":
		result, err := client.HVals(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"value": v}
		}
		return map[string]interface{}{"values": outputResult}, nil

	case "ZCARD":
		result, err := client.ZCard(ctx, inputParams["key"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "ZREM":
		inputMembersArray := inputParams["members"].([]interface{})
		inputMembers := make([]interface{}, len(inputMembersArray))
		for i, v := range inputMembersArray {
			val := v.(map[string]interface{})
			inputMembers[i] = val["member"]
		}
		result, err := client.ZRem(ctx, inputParams["key"].(string), inputMembers...).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "ZRANGE":
		var outputResult []map[string]interface{}
		withscores, byscore, bylex, rev := false, false, false, false
		if val, ok := inputParams["WITHSCORES"]; ok {
			withscores = val.(bool)
		}
		if val, ok := inputParams["BYSCORE"]; ok {
			byscore = val.(bool)
		}
		if val, ok := inputParams["BYLEX"]; ok {
			bylex = val.(bool)
		}
		if val, ok := inputParams["REV"]; ok {
			rev = val.(bool)
		}

		// validations
		if byscore && bylex {
			return nil, fmt.Errorf("cannot select both BYSCORE and BYLEX")
		} else if bylex && withscores {
			return nil, fmt.Errorf("WITHSCORES not supported in combination with BYLEX")
		}

		var zrb *redis.ZRangeBy
		zrb = new(redis.ZRangeBy)
		zrb.Min = fmt.Sprintf("%v", inputParams["start"])
		zrb.Max = fmt.Sprintf("%v", inputParams["stop"])
		zrb.Offset = 0
		if val, ok := inputParams["offset"]; ok {
			zrb.Offset = ToInt64(val)
		}
		zrb.Count = 0
		if val, ok := inputParams["count"]; ok {
			zrb.Count = ToInt64(val)
		}
		var err error
		if withscores {
			var result []redis.Z
			if byscore {
				if rev {
					result, err = client.ZRevRangeByScoreWithScores(ctx, inputParams["key"].(string), zrb).Result()
				} else {
					result, err = client.ZRangeByScoreWithScores(ctx, inputParams["key"].(string), zrb).Result()
				}
			} else {
				start_int, err := strconv.ParseInt(inputParams["start"].(string), 10, 64)
				if err != nil {
					return nil, err
				}
				stop_int, err := strconv.ParseInt(inputParams["stop"].(string), 10, 64)
				if err != nil {
					return nil, err
				}
				if rev {
					result, err = client.ZRevRangeWithScores(ctx, inputParams["key"].(string), start_int, stop_int).Result()
				} else {
					result, err = client.ZRangeWithScores(ctx, inputParams["key"].(string), start_int, stop_int).Result()
				}
			}
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
			}
		} else {
			var result []string
			if byscore {
				if rev {
					result, err = client.ZRevRangeByScore(ctx, inputParams["key"].(string), zrb).Result()
				} else {
					result, err = client.ZRangeByScore(ctx, inputParams["key"].(string), zrb).Result()
				}
			} else if bylex {
				if rev {
					result, err = client.ZRevRangeByLex(ctx, inputParams["key"].(string), zrb).Result()
				} else {
					result, err = client.ZRangeByLex(ctx, inputParams["key"].(string), zrb).Result()
				}
			} else {
				start_int, err := strconv.ParseInt(inputParams["start"].(string), 10, 64)
				if err != nil {
					return nil, err
				}
				stop_int, err := strconv.ParseInt(inputParams["stop"].(string), 10, 64)
				if err != nil {
					return nil, err
				}
				if rev {
					result, err = client.ZRevRange(ctx, inputParams["key"].(string), start_int, stop_int).Result()
				} else {
					result, err = client.ZRange(ctx, inputParams["key"].(string), start_int, stop_int).Result()
				}
			}
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v}
			}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "ZCOUNT":
		result, err := client.ZCount(ctx, inputParams["key"].(string), inputParams["min"].(string), inputParams["max"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "ZRANK":
		result, err := client.ZRank(ctx, inputParams["key"].(string), inputParams["member"].(string)).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "ZREMRANGEBYRANK":
		result, err := client.ZRemRangeByRank(ctx, inputParams["key"].(string), ToInt64(inputParams["start"]), ToInt64(inputParams["stop"])).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "ZREMRANGEBYSCORE":
		result, err := client.ZRemRangeByScore(ctx, inputParams["key"].(string), (inputParams["min"].(string)), (inputParams["max"].(string))).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "ZADD":
		/*
			type ZAddArgs struct {
				NX      bool
				XX      bool
				LT      bool
				GT      bool
				Ch      bool
				Members []Z
			}
			The GT, LT and NX options are mutually exclusive
		*/
		nx, xx, lt, gt, incr, ch := false, false, false, false, false, false
		inputMembersArray := inputParams["memberswithscores"].([]interface{})
		zArray := make([]redis.Z, len(inputMembersArray))

		for i, v := range inputMembersArray {
			memberScorePair := v.(map[string]interface{})
			zArray[i].Score = memberScorePair["score"].(float64)
			zArray[i].Member = memberScorePair["member"]
		}

		if inputParams["INCR"] != nil && inputParams["INCR"] == "INCR" {
			incr = true
		}
		if val, ok := inputParams["CH"]; ok {
			ch = val.(bool)
		}

		if inputParams["NX|XX"] != nil {
			if inputParams["NX|XX"] == "NX" {
				nx = true
			} else if inputParams["NX|XX"] == "XX" {
				xx = true
			}
		}
		if inputParams["LT|GT"] != nil {
			if inputParams["LT|GT"] == "LT" {
				if nx {
					return nil, errors.New("ERR LT and NX options at the same time are not compatible")
				}
				lt = true
			} else if inputParams["LT|GT"] == "GT" {
				if nx {
					return nil, errors.New("ERR GT and NX options at the same time are not compatible")
				}
				gt = true
			}
		}
		args := redis.ZAddArgs{
			NX:      nx,
			XX:      xx,
			LT:      lt,
			GT:      gt,
			Ch:      ch,
			Members: zArray}

		if incr {
			result, err := client.ZAddArgsIncr(ctx, inputParams["key"].(string), args).Result()
			if err != nil {
				return nil, err
			}
			return result, nil
		}
		result, err := client.ZAddArgs(ctx, inputParams["key"].(string), args).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "ZSCORE":
		result, err := client.ZScore(ctx, inputParams["key"].(string), inputParams["member"].(string)).Result()
		if err != nil {
			return nil, err
		}
		//converted float to string
		res := strconv.FormatFloat(result, 'f', -1, 64)
		return res, nil

	case "ZUNIONSTORE":
		/*
			type ZStore struct {
				Keys    []string
				Weights []float64
				// Can be SUM, MIN or MAX.
				Aggregate string
			}
		*/

		var numkeys int64
		numkeys = ToInt64(inputParams["numkeys"])
		if numkeys <= 0 {
			return nil, errors.New("numkeys : At least 1 input key is needed for this Command")
		}
		inputKeysArray := inputParams["keys"].([]interface{})
		if len(inputKeysArray) != int(numkeys) {
			return nil, errors.New("number of keys are mismatching")
		}
		inputKeys := make([]string, numkeys)
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		inputWeights := make([]float64, numkeys)
		if val, ok := inputParams["weights"]; ok {
			inputWeightsArray := val.([]interface{})
			if len(inputWeightsArray) != int(numkeys) {
				return nil, errors.New("number of weights are mismatching")
			}
			for i, v := range inputWeightsArray {
				val := v.(map[string]interface{})
				inputWeights[i] = ToFloat64(val["weight"])
			}
		} else {
			for i := 0; i < int(numkeys); i++ {
				inputWeights[i] = ToFloat64(1)
			}
		}

		var aggregate string
		if val, ok := inputParams["AGGREGATE SUM|MIN|MAX"]; ok {
			aggregate = val.(string)
		} else {
			aggregate = "SUM"
		}

		store := &redis.ZStore{Keys: inputKeys, Weights: inputWeights, Aggregate: aggregate}
		result, err := client.ZUnionStore(ctx, inputParams["destination"].(string), store).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "ZUNION":
		var outputResult []map[string]interface{}
		withscores := false
		if val, ok := inputParams["WITHSCORES"]; ok {
			withscores = val.(bool)
		}

		var numkeys int64
		numkeys = ToInt64(inputParams["numkeys"])
		if numkeys <= 0 {
			return nil, errors.New("numkeys : At least 1 input key is needed for this Command")
		}
		inputKeysArray := inputParams["keys"].([]interface{})
		if len(inputKeysArray) != int(numkeys) {
			return nil, errors.New("number of keys are mismatching")
		}
		inputKeys := make([]string, numkeys)
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		inputWeights := make([]float64, numkeys)
		if val, ok := inputParams["weights"]; ok {
			inputWeightsArray := val.([]interface{})
			if len(inputWeightsArray) != int(numkeys) {
				return nil, errors.New("number of weights are mismatching")
			}
			for i, v := range inputWeightsArray {
				val := v.(map[string]interface{})
				inputWeights[i] = ToFloat64(val["weight"])
			}
		} else {
			for i := 0; i < int(numkeys); i++ {
				inputWeights[i] = ToFloat64(1)
			}
		}

		var aggregate string
		if val, ok := inputParams["AGGREGATE SUM|MIN|MAX"]; ok {
			aggregate = val.(string)
		} else {
			aggregate = "SUM"
		}

		store := redis.ZStore{Keys: inputKeys, Weights: inputWeights, Aggregate: aggregate}

		if withscores {
			result, err := client.ZUnionWithScores(ctx, store).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
			}
		} else {
			result, err := client.ZUnion(ctx, store).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v}
			}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "ZPOPMAX":
		count := make([]int64, 1)
		if val, ok := inputParams["count"]; ok {
			count[0] = ToInt64(val)
		} else {
			count[0] = int64(1)
		}
		result, err := client.ZPopMax(ctx, inputParams["key"].(string), count...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "ZPOPMIN":
		var count int64
		if val, ok := inputParams["count"]; ok {
			count = ToInt64(val)
		} else {
			count = int64(1)
		}
		result, err := client.ZPopMin(ctx, inputParams["key"].(string), count).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "BZPOPMAX":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.BZPopMax(ctx, time.Duration(ToInt64(inputParams["timeout"])), inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, 3)
		outputResult[0] = map[string]interface{}{"element": result.Key}
		outputResult[1] = map[string]interface{}{"element": result.Z.Member}
		outputResult[2] = map[string]interface{}{"element": result.Z.Score}
		return map[string]interface{}{"elements": outputResult}, nil
	case "BZPOPMIN":
		inputKeysArray := inputParams["keys"].([]interface{})
		inputKeys := make([]string, len(inputKeysArray))
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		result, err := client.BZPopMin(ctx, time.Duration(ToInt64(inputParams["timeout"])), inputKeys...).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, 3)
		outputResult[0] = map[string]interface{}{"element": result.Key}
		outputResult[1] = map[string]interface{}{"element": result.Z.Member}
		outputResult[2] = map[string]interface{}{"element": result.Z.Score}
		return map[string]interface{}{"elements": outputResult}, nil
	case "ZDIFFSTORE":
		var numkeys int64
		numkeys = ToInt64(inputParams["numkeys"])
		if numkeys <= 0 {
			return nil, errors.New("numkeys : At least 1 input key is needed for this Command")
		}
		inputKeysArray := inputParams["keys"].([]interface{})
		if len(inputKeysArray) != int(numkeys) {
			return nil, errors.New("number of keys are mismatching")
		}
		inputKeys := make([]string, numkeys)
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}

		result, err := client.ZDiffStore(ctx, inputParams["destination"].(string), inputKeys...).Result()
		if err != nil {
			return nil, err
		}

		return result, nil
	case "ZDIFF":
		var outputResult []map[string]interface{}
		withscores := false
		if val, ok := inputParams["WITHSCORES"]; ok {
			withscores = val.(bool)
		}

		var numkeys int64
		numkeys = ToInt64(inputParams["numkeys"])
		if numkeys <= 0 {
			return nil, errors.New("numkeys : At least 1 input key is needed for this Command")
		}
		inputKeysArray := inputParams["keys"].([]interface{})
		if len(inputKeysArray) != int(numkeys) {
			return nil, errors.New("number of keys are mismatching")
		}
		inputKeys := make([]string, numkeys)
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}

		if withscores {
			result, err := client.ZDiffWithScores(ctx, inputKeys...).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
			}
		} else {
			result, err := client.ZDiff(ctx, inputKeys...).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v}
			}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "ZINCRBY":
		result, err := client.ZIncrBy(ctx, inputParams["key"].(string), ToFloat64(inputParams["increment"]), inputParams["member"].(string)).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "ZINTER":
		var outputResult []map[string]interface{}
		withscores := false
		if val, ok := inputParams["WITHSCORES"]; ok {
			withscores = val.(bool)
		}

		var numkeys int64
		numkeys = ToInt64(inputParams["numkeys"])
		if numkeys <= 0 {
			return nil, errors.New("numkeys : At least 1 input key is needed for this Command")
		}
		inputKeysArray := inputParams["keys"].([]interface{})
		if len(inputKeysArray) != int(numkeys) {
			return nil, errors.New("number of keys are mismatching")
		}
		inputKeys := make([]string, numkeys)
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		inputWeights := make([]float64, numkeys)
		if val, ok := inputParams["weights"]; ok {
			inputWeightsArray := val.([]interface{})
			if len(inputWeightsArray) != int(numkeys) {
				return nil, errors.New("number of weights are mismatching")
			}
			for i, v := range inputWeightsArray {
				val := v.(map[string]interface{})
				inputWeights[i] = ToFloat64(val["weight"])
			}
		} else {
			for i := 0; i < int(numkeys); i++ {
				inputWeights[i] = ToFloat64(1)
			}
		}

		var aggregate string
		if val, ok := inputParams["AGGREGATE SUM|MIN|MAX"]; ok {
			aggregate = val.(string)
		} else {
			aggregate = "SUM"
		}

		store := &redis.ZStore{Keys: inputKeys, Weights: inputWeights, Aggregate: aggregate}

		if withscores {
			result, err := client.ZInterWithScores(ctx, store).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
			}
		} else {
			result, err := client.ZInter(ctx, store).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v}
			}
		}
		return map[string]interface{}{"members": outputResult}, nil
	case "ZINTERSTORE":
		/*
			type ZStore struct {
				Keys    []string
				Weights []float64
				// Can be SUM, MIN or MAX.
				Aggregate string
			}
		*/

		var numkeys int64
		numkeys = ToInt64(inputParams["numkeys"])
		if numkeys <= 0 {
			return nil, errors.New("numkeys : At least 1 input key is needed for this Command")
		}
		inputKeysArray := inputParams["keys"].([]interface{})
		if len(inputKeysArray) != int(numkeys) {
			return nil, errors.New("number of keys are mismatching")
		}
		inputKeys := make([]string, numkeys)
		for i, v := range inputKeysArray {
			val := v.(map[string]interface{})
			inputKeys[i] = val["key"].(string)
		}
		inputWeights := make([]float64, numkeys)
		if val, ok := inputParams["weights"]; ok {
			inputWeightsArray := val.([]interface{})
			if len(inputWeightsArray) != int(numkeys) {
				return nil, errors.New("number of weights are mismatching")
			}
			for i, v := range inputWeightsArray {
				val := v.(map[string]interface{})
				inputWeights[i] = ToFloat64(val["weight"])
			}
		} else {
			for i := 0; i < int(numkeys); i++ {
				inputWeights[i] = ToFloat64(1)
			}
		}

		var aggregate string
		if val, ok := inputParams["AGGREGATE SUM|MIN|MAX"]; ok {
			aggregate = val.(string)
		} else {
			aggregate = "SUM"
		}
		store := &redis.ZStore{Keys: inputKeys, Weights: inputWeights, Aggregate: aggregate}
		result, err := client.ZInterStore(ctx, inputParams["destination"].(string), store).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "ZMSCORE":
		var outputResult []map[string]interface{}

		inputMembersArray := inputParams["members"].([]interface{})
		inputMembers := make([]string, len(inputMembersArray))
		for i, v := range inputMembersArray {
			val := v.(map[string]interface{})
			inputMembers[i] = val["member"].(string)
		}
		result, err := client.ZMScore(ctx, inputParams["key"].(string), inputMembers...).Result()
		if err != nil {
			return nil, err
		}
		outputResult = make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"score": v}
		}

		return map[string]interface{}{"scores": outputResult}, nil

	case "ZRANDMEMBER":
		var outputResult []map[string]interface{}
		count := int(1)
		if val, ok := inputParams["count"]; ok {
			count = ToInt(val)
		}
		withscores := false
		if val, ok := inputParams["WITHSCORES"]; ok {
			withscores = val.(bool)
		}
		if withscores {
			// from redis v9 we have ZRandMemberWithScores which returns []redis.Z where Z(member,score)
			result, err := client.ZRandMemberWithScores(ctx, inputParams["key"].(string), count).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
			}
		} else {
			result, err := client.ZRandMember(ctx, inputParams["key"].(string), count).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v}
			}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "ZSCAN":
		var result []string
		var match string
		var count int64
		cursorSigned := ToInt64(inputParams["cursor"])
		var cursorUnSigned uint64
		//if value provided by user is negative, setting it as zero unsigned value because cursor will never negative
		if cursorSigned <= 0 {
			cursorUnSigned = uint64(0)
		} else {
			cursorUnSigned = uint64(cursorSigned)
		}
		//if count is not set by user default value as per RedisDocs is 10
		if val, ok := inputParams["count"]; ok {
			count = ToInt64(val)
		} else {
			count = 10
		}
		if val, ok := inputParams["match"]; ok {
			match = val.(string)
		} else {
			match = "*"
		}
		result, cursor, err := client.ZScan(ctx, inputParams["key"].(string), cursorUnSigned, match, count).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, (len(result) / 2))
		for i, j := 0, 0; i < len(result); i, j = i+2, j+1 {
			outputResult[j] = map[string]interface{}{"member": result[i], "score": result[i+1]}
		}
		return map[string]interface{}{"cursor": cursor, "members": outputResult}, nil

	case "HINCRBY":
		IncrByFloat := ToFloat64(inputParams["increment"])
		if IncrByFloat != float64(ToInt64(inputParams["increment"])) {
			return nil, fmt.Errorf("ERR value is not an integer or out of range")
		}

		incr := ToInt64(inputParams["increment"])
		result, err := client.HIncrBy(ctx, inputParams["key"].(string), inputParams["field"].(string), incr).Result()
		if err != nil {
			return nil, err
		}

		return result, nil

	case "HINCRBYFLOAT":
		var err error

		result, err := client.HIncrByFloat(ctx, inputParams["key"].(string), inputParams["field"].(string), ToFloat64(inputParams["increment"])).Result()
		if err != nil {
			return nil, err
		}

		//converted float to string
		res := strconv.FormatFloat(result, 'f', -1, 64)
		return res, nil

	case "HSCAN":

		cursor := ToInt64(inputParams["cursor"])
		var pattern string = ""
		if val, ok := inputParams["pattern"]; ok {
			pattern = val.(string)
		}

		var count int64 = 0
		if val, ok := inputParams["count"]; ok {
			count = ToInt64(val)
		}
		result, crs, err := client.HScan(ctx, inputParams["key"].(string), uint64(cursor), pattern, count).Result()
		if err != nil {
			return nil, err
		}

		// converting string array to fieldvalue pairs
		outputResult := make([]map[string]interface{}, len(result)/2)
		j := 0
		for i := 0; i < len(result); i += 2 {
			outputResult[j] = map[string]interface{}{"field": result[i], "value": result[i+1]}
			j++
		}
		return map[string]interface{}{"cursor": crs, "fieldvalues": outputResult}, nil

	case "HRANDFIELD":

		count := 1
		if val, ok := inputParams["count"]; ok {
			count = int(ToInt64(val))
		}

		withvalues := false
		if val, ok := inputParams["withvalues"]; ok {
			withvalues = val.(bool)
		}
		var outputResult []map[string]interface{}

		if withvalues {
			result, err := client.HRandFieldWithValues(ctx, inputParams["key"].(string), count).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"field": v.Key, "value": v.Value}
			}
		} else {
			result, err := client.HRandField(ctx, inputParams["key"].(string), count).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"field": v}
			}
		}

		return map[string]interface{}{"fieldvalues": outputResult}, nil

	case "HSETNX":
		result, err := client.HSetNX(ctx, inputParams["key"].(string), inputParams["field"].(string), inputParams["value"]).Result()
		if err != nil {
			return nil, err
		}
		// output change
		var op int
		if result == true {
			op = 1
		} else {
			op = 0
		}
		return op, nil

	case "ZREVRANK":
		result, err := client.ZRevRank(ctx, inputParams["key"].(string), inputParams["member"].(string)).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "ZRANGEBYLEX":
		var zrb *redis.ZRangeBy
		zrb = new(redis.ZRangeBy)
		zrb.Min = inputParams["min"].(string)
		zrb.Max = inputParams["max"].(string)
		zrb.Offset = 0
		if val, ok := inputParams["offset"]; ok {
			zrb.Offset = ToInt64(val)
		}
		zrb.Count = 0
		if val, ok := inputParams["count"]; ok {
			zrb.Count = ToInt64(val)
		}

		result, err := client.ZRangeByLex(ctx, inputParams["key"].(string), zrb).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "ZRANGEBYSCORE":
		var outputResult []map[string]interface{}
		withscores := false
		if val, ok := inputParams["WITHSCORES"]; ok {
			withscores = val.(bool)
		}

		var zrb *redis.ZRangeBy
		zrb = new(redis.ZRangeBy)
		zrb.Min = fmt.Sprintf("%v", inputParams["min"])
		zrb.Max = fmt.Sprintf("%v", inputParams["max"])
		zrb.Offset = 0
		if val, ok := inputParams["offset"]; ok {
			zrb.Offset = ToInt64(val)
		}
		zrb.Count = 0
		if val, ok := inputParams["count"]; ok {
			zrb.Count = ToInt64(val)
		}

		if withscores {
			result, err := client.ZRangeByScoreWithScores(ctx, inputParams["key"].(string), zrb).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
			}

		} else {
			result, err := client.ZRangeByScore(ctx, inputParams["key"].(string), zrb).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v}
			}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "ZREVRANGE":
		var outputResult []map[string]interface{}
		withscores := false
		if val, ok := inputParams["WITHSCORES"]; ok {
			withscores = val.(bool)
		}
		if withscores {
			result, err := client.ZRevRangeWithScores(ctx, inputParams["key"].(string), ToInt64(inputParams["start"]), ToInt64(inputParams["stop"])).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
			}
		} else {
			result, err := client.ZRevRange(ctx, inputParams["key"].(string), ToInt64(inputParams["start"]), ToInt64(inputParams["stop"])).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v}
			}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "ZREVRANGEBYLEX":

		var zrb *redis.ZRangeBy
		zrb = new(redis.ZRangeBy)
		zrb.Min = inputParams["min"].(string)
		zrb.Max = inputParams["max"].(string)
		zrb.Offset = 0
		if val, ok := inputParams["offset"]; ok {
			zrb.Offset = ToInt64(val)
		}
		zrb.Count = 0
		if val, ok := inputParams["count"]; ok {
			zrb.Count = ToInt64(val)
		}

		result, err := client.ZRevRangeByLex(ctx, inputParams["key"].(string), zrb).Result()
		if err != nil {
			return nil, err
		}
		outputResult := make([]map[string]interface{}, len(result))
		for i, v := range result {
			outputResult[i] = map[string]interface{}{"member": v}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "ZREVRANGEBYSCORE":
		var outputResult []map[string]interface{}
		withscores := false
		if val, ok := inputParams["WITHSCORES"]; ok {
			withscores = val.(bool)
		}
		var zrb *redis.ZRangeBy
		zrb = new(redis.ZRangeBy)
		zrb.Min = fmt.Sprintf("%v", inputParams["min"])
		zrb.Max = fmt.Sprintf("%v", inputParams["max"])
		zrb.Offset = 0
		if val, ok := inputParams["offset"]; ok {
			zrb.Offset = ToInt64(val)
		}
		zrb.Count = 0
		if val, ok := inputParams["count"]; ok {
			zrb.Count = ToInt64(val)
		}
		if withscores {
			result, err := client.ZRevRangeByScoreWithScores(ctx, inputParams["key"].(string), zrb).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v.Member, "score": v.Score}
			}
		} else {
			result, err := client.ZRevRangeByScore(ctx, inputParams["key"].(string), zrb).Result()
			if err != nil {
				return nil, err
			}
			outputResult = make([]map[string]interface{}, len(result))
			for i, v := range result {
				outputResult[i] = map[string]interface{}{"member": v}
			}
		}
		return map[string]interface{}{"members": outputResult}, nil

	case "ZRANGESTORE":

		var zrargs *redis.ZRangeArgs
		zrargs = new(redis.ZRangeArgs)
		zrargs.Key = inputParams["src"].(string)
		zrargs.Start = fmt.Sprintf("%v", inputParams["min"])
		zrargs.Stop = fmt.Sprintf("%v", inputParams["max"])
		zrargs.ByScore = false
		if val, ok := inputParams["byScore"]; ok {
			zrargs.ByScore = val.(bool)
		}
		zrargs.ByLex = false
		if val, ok := inputParams["byLex"]; ok {
			zrargs.ByLex = val.(bool)
		}
		zrargs.Rev = false
		if val, ok := inputParams["rev"]; ok {
			zrargs.Rev = val.(bool)
		}
		zrargs.Offset = 0
		if val, ok := inputParams["offset"]; ok {
			zrargs.Offset = ToInt64(val)
		}
		zrargs.Count = 0
		if val, ok := inputParams["count"]; ok {
			zrargs.Count = ToInt64(val)
		}

		result, err := client.ZRangeStore(ctx, inputParams["dst"].(string), *zrargs).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "ZREMRANGEBYLEX":
		result, err := client.ZRemRangeByLex(ctx, inputParams["key"].(string), inputParams["min"].(string), inputParams["max"].(string)).Result()
		if err != nil {

			return result, nil
		}
		return result, nil

	case "ZLEXCOUNT":
		result, err := client.ZLexCount(ctx, inputParams["key"].(string), inputParams["min"].(string), inputParams["max"].(string)).Result()
		if err != nil {
			return nil, err
		}
		return result, nil

	case "JSON.SET":
		key := inputParams["key"].(string)
		path := inputParams["path"].(string)
		val := inputParams["value"]
		mode := ""
		if inputParams["NX|XX"] != nil {
			if inputParams["NX|XX"].(string) == "NX" {
				mode = "NX"
			} else if inputParams["NX|XX"].(string) == "XX" {
				mode = "XX"
			}
		}
		if mode == "" {
			result, err := client.JSONSet(ctx, key, path, val).Result()
			if err != nil {
				return checkForRedisNilError(err)
			}
			return result, nil
		}
		result, err := client.JSONSetMode(ctx, key, path, val, mode).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.MSET":
		inputParamArray := inputParams["keyPathValues"]
		byteArr, err := json.Marshal(inputParamArray)
		if err != nil {
			return nil, err
		}
		// args ex: [{key:"key1",path:"path1",value:{key:"val"}}]
		var args []map[string]interface{}
		var jsonValArg jsonValueArg
		var stringValArg stringValueArg
		docs := []redis.JSONSetArgs{}
		json.Unmarshal(byteArr, &args)
		for _, v := range args {
			var nArg redis.JSONSetArgs
			val := v["value"]
			key := v["key"].(string)
			path := v["path"].(string)
			switch val.(type) {
			case map[string]interface{}:
				jsonValArg = val.(map[string]interface{})
				nArg = redis.JSONSetArgs{Key: key, Path: path, Value: jsonValArg}
			case string:
				stringValArg = stringValueArg(fmt.Sprint(val))
				nArg = redis.JSONSetArgs{Key: key, Path: path, Value: stringValArg}
			default:
				nArg = redis.JSONSetArgs{Key: key, Path: path, Value: val}
			}
			docs = append(docs, nArg)
		}
		result, err := client.JSONMSetArgs(ctx, docs).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.GET":
		indent, newline, space := "", "", ""
		if inputParams["indent"] != nil {
			indent = inputParams["indent"].(string)
		}
		if inputParams["newline"] != nil {
			newline = inputParams["newline"].(string)
		}
		if inputParams["space"] != nil {
			space = inputParams["space"].(string)
		}
		options := redis.JSONGetArgs{
			Indent:  indent,
			Newline: newline,
			Space:   space,
		}
		var paths []string
		if inputParams["paths"] != nil {
			inputPathArray := inputParams["paths"].([]interface{})
			paths = make([]string, len(inputPathArray))
			for i, v := range inputPathArray {
				paths[i] = v.(string)
			}
		}

		result, err := client.JSONGetWithArgs(ctx, inputParams["key"].(string), &options, paths...).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.MGET":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		inputkey := inputParams["key"].([]interface{})
		var keys []string
		for _, v := range inputkey {
			keys = append(keys, v.(string))
		}
		result, err := client.JSONMGet(ctx, path, keys...).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.DEL":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONDel(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.CLEAR":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONClear(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.TYPE":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONType(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.TOGGLE":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONToggle(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.STRLEN":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONStrLen(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.STRAPPEND":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONStrAppend(ctx, inputParams["key"].(string), path, inputParams["value"].(string)).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.ARRAPPEND":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		var arrValues arrValueArg
		for _, v := range inputParams["value"].([]interface{}) {
			arrValues = append(arrValues, v)
		}
		result, err := client.JSONArrAppend(ctx, inputParams["key"].(string), path, arrValues).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.ARRINDEX":
		start, stop := 0, 0
		if inputParams["start"] != nil {
			start = ToInt(inputParams["start"])
		}
		if inputParams["stop"] != nil {
			stop = ToInt(inputParams["stop"])
		}
		inputValues := inputParams["value"].([]interface{})
		var arrValues arrValueArg

		for _, v := range inputValues {
			arrValues = append(arrValues, v)
		}
		if inputParams["stop"] != nil || inputParams["start"] != nil {
			indexArgs := &redis.JSONArrIndexArgs{
				Start: start,
				Stop:  &stop,
			}
			result, err := client.JSONArrIndexWithArgs(ctx, inputParams["key"].(string), inputParams["path"].(string), indexArgs, arrValues...).Result()
			if err != nil {
				return checkForRedisNilError(err)
			}
			return result, nil
		}
		result, err := client.JSONArrIndex(ctx, inputParams["key"].(string), inputParams["path"].(string), arrValues...).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.ARRINSERT":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		inputValues := inputParams["value"].([]interface{})
		var arrValues arrValueArg

		for _, v := range inputValues {
			arrValues = append(arrValues, v)
		}
		result, err := client.JSONArrInsert(ctx, inputParams["key"].(string), path, ToInt64(inputParams["index"]), arrValues...).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.ARRLEN":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONArrLen(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.ARRPOP":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONArrPop(ctx, inputParams["key"].(string), path, ToInt(inputParams["index"])).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.ARRTRIM":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		start, stop := 0, 0
		if inputParams["start"] != nil {
			start = ToInt(inputParams["start"])
		}
		if inputParams["stop"] != nil {
			stop = ToInt(inputParams["stop"])
		}
		if inputParams["stop"] != nil || inputParams["start"] != nil {
			options := &redis.JSONArrTrimArgs{
				Start: start,
				Stop:  &stop,
			}
			result, err := client.JSONArrTrimWithArgs(ctx, inputParams["key"].(string), path, options).Result()
			if err != nil {
				return checkForRedisNilError(err)
			}
			return result, nil
		}
		result, err := client.JSONArrTrim(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.MERGE":
		result, err := client.JSONMerge(ctx, inputParams["key"].(string), inputParams["path"].(string), ToString(inputParams["value"])).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.NUMINCRBY":
		result, err := client.JSONNumIncrBy(ctx, inputParams["key"].(string), inputParams["path"].(string), ToFloat64(inputParams["value"])).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil

	case "JSON.OBJKEYS":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		//result is [][]string
		result, err := client.JSONObjKeys(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		var out []interface{}
		for _, v := range result {
			if v == nil {
				out = append(out, nil)
				continue
			}
			keys := v.([]interface{})
			outkeys := make([]string, len(keys))
			for i, k := range keys {
				outkeys[i] = k.(string)
			}
			out = append(out, outkeys)
		}
		return out, nil

	case "JSON.OBJLEN":
		path := "$"
		if inputParams["path"] != nil {
			path = inputParams["path"].(string)
		}
		result, err := client.JSONObjLen(ctx, inputParams["key"].(string), path).Result()
		if err != nil {
			return checkForRedisNilError(err)
		}
		return result, nil
	}

	return nil, nil
}

// Declared jsonValueArg,stringValueArg needs to implement MarshalBinary method for redis.JSONSetArgs
type jsonValueArg map[string]interface{}

func (i jsonValueArg) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

type stringValueArg string

func (i stringValueArg) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

type arrValueArg []interface{}

func (i arrValueArg) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}
func ToFloat64(val interface{}) float64 {
	switch val.(type) {
	case int:
		return float64(val.(int))
	case int64:
		return float64(val.(int64))
	case float64:
		return val.(float64)
	}
	return 0
}

func ToInt64(val interface{}) int64 {
	switch val.(type) {
	case int:
		return int64(val.(int))
	case int64:
		return val.(int64)
	case float64:
		return int64(val.(float64))
	}
	return 0
}
func ToInt(val interface{}) int {
	switch val.(type) {
	case int:
		return val.(int)
	case int64:
		return int(val.(int64))
	case float64:
		return int(val.(float64))
	}
	return 0
}
func BoolToInt(val bool) int {
	if val {
		return 1
	}
	return 0
}
func ToString(value interface{}) string {
	switch v := value.(type) {
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		return v
	case []interface{}:
		result := "["
		for i, item := range v {
			if i > 0 {
				result += ", "
			}
			result += ToString(item)
		}
		result += "]"
		return result
	case map[interface{}]interface{}:
		result := "{"
		first := true
		for key, val := range v {
			if !first {
				result += ", "
			}
			result += ToString(key) + ": " + ToString(val)
			first = false
		}
		result += "}"
		return result
	default:
		// Handle other types using reflection
		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.Slice, reflect.Array:
			result := "["
			for i := 0; i < val.Len(); i++ {
				if i > 0 {
					result += ", "
				}
				result += ToString(val.Index(i).Interface())
			}
			result += "]"
			return result
		case reflect.Map:
			result := "{"
			keys := val.MapKeys()
			for i, key := range keys {
				if i > 0 {
					result += ", "
				}
				result += ToString(key.Interface()) + ": " + ToString(val.MapIndex(key).Interface())
			}
			result += "}"
			return result
		default:
			// Fallback to fmt.Sprintf with %v
			return fmt.Sprintf("%v", value)
		}
	}
}
func checkForRedisNilError(err error) (interface{}, error) {
	if err.Error() == "redis: nil" {
		return "redis: nil", nil
	}
	return nil, err
}
