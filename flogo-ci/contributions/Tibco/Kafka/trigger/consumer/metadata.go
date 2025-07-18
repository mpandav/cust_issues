package consumer

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Settings struct {
	Connection connection.Manager `md:"kafkaConnection,required"`
}

type HandlerSettings struct {
	FetchMinBytes     int    `md:"fetchMinBytes"`
	CommitInterval    int    `md:"commitInterval"`
	FetchMaxWait      int    `md:"fetchMaxWait"`
	HeartbeatInterval int    `md:"heartbeatInterval"`
	SessionTimeout    int    `md:"sessionTimeout"`
	Topic             string `md:"topic"`
	ConsumerGroup     string `md:"consumerGroup,required"`
	TopicPattern      string `md:"topicPattern"`
	ValueType         string `md:"valueType,allowed(String,JSON,Avro)"`
	InitialOffset     string `md:"initialOffset,allowed(Newest,Oldest,Seek By Offset,Seek By Timestamp)"`
	PartitionId       int32  `md:"partitionId"`
	CustomPartition   bool   `md:"customPartition"`
	Offset            int64  `md:"seekOffset"` // For Seek By Offset
	TimeStamp         string `md:"timeStamp"`  // For Seek By Timestamp
}

type Output struct {
	Partition   int                    `md:"partition,required"`
	Offset      int64                  `md:"offset,required"`
	Key         string                 `md:"key"`
	StringValue string                 `md:"stringValue"`
	Topic       string                 `md:"topic,required"`
	JsonValue   map[string]interface{} `md:"jsonValue"`

	AvroData map[string]interface{} `md:"avroData"`
	Headers  map[string]interface{} `md:"headers"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"partition":   o.Partition,
		"offset":      o.Offset,
		"key":         o.Key,
		"headers":     o.Headers,
		"stringValue": o.StringValue,
		"jsonValue":   o.JsonValue,

		"avroData": o.AvroData,
		"topic":    o.Topic,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.Topic, err = coerce.ToString(values["topic"])
	if err != nil {
		return err
	}

	o.Key, err = coerce.ToString(values["key"])
	if err != nil {
		return err
	}

	o.Headers, err = coerce.ToObject(values["headers"])
	if err != nil {
		return err
	}

	o.StringValue, err = coerce.ToString(values["stringValue"])
	if err != nil {
		return err
	}

	o.JsonValue, err = coerce.ToObject(values["jsonValue"])
	if err != nil {
		return err
	}

	o.AvroData, err = coerce.ToObject(values["avroData"])
	if err != nil {
		return err
	}

	o.Offset, err = coerce.ToInt64(values["offset"])
	if err != nil {
		return err
	}

	o.Partition, err = coerce.ToInt(values["partition"])
	if err != nil {
		return err
	}

	return nil
}
