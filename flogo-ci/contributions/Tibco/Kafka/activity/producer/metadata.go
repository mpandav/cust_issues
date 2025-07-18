package producer

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	"github.com/tibco/wi-plugins/contributions/kafka/src/app/Kafka/connector/kafka"
)

type Input struct {
	Partition       int32                  `md:"partition"`
	Connection      connection.Manager     `md:"kafkaConnection,required"`
	AckTimeout      int                    `md:"ackTimeout"`
	MaxRequestSize  int                    `md:"maxRequestSize"`
	MaxMessages     int                    `md:"maxMessages"`
	Frequency       int                    `md:"frequency"`
	AckMode         string                 `md:"ackMode,required"`
	Partitioner     string                 `md:"partitioner"`
	CompressionType string                 `md:"compressionType"`
	ValueType       string                 `md:"valueType,allowed(String,JSON,Avro)`
	Key             string                 `md:"key"`
	StringValue     string                 `md:"stringValue,required"`
	Topic           string                 `md:"topic,required"`
	JsonValue       map[string]interface{} `md:"jsonValue,required"`
	AvroData        map[string]interface{} `md:"avroData,required"`

	Headers map[string]interface{} `md:"headers"`
}

type Output struct {
	Partition int32  `md:"partition,required"`
	Offset    int64  `md:"offset,required"`
	Topic     string `md:"topic,required"`
}

func (o *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"partition":       o.Partition,
		"kafkaConnection": o.Connection,
		"topic":           o.Topic,
		"ackTimeout":      o.AckTimeout,
		"maxRequestSize":  o.MaxRequestSize,
		"maxMessages":     o.MaxMessages,
		"frequency":       o.Frequency,
		"ackMode":         o.AckMode,
		"compressionType": o.CompressionType,
		"partitioner":     o.Partitioner,
		"valueType":       o.ValueType,
		"key":             o.Key,
		"stringValue":     o.StringValue,
		"jsonValue":       o.JsonValue,
		"avroData":        o.AvroData,
		"headers":         o.Headers,
	}
}

func (o *Input) FromMap(values map[string]interface{}) error {
	var err error
	o.Topic, err = coerce.ToString(values["topic"])
	if err != nil {
		return err
	}

	o.CompressionType, err = coerce.ToString(values["compressionType"])
	if err != nil {
		return err
	}

	o.Partitioner, _ = coerce.ToString(values["partitioner"])
	if o.Partitioner == "" {
		o.Partitioner = "Hash"
	}

	o.Key, err = coerce.ToString(values["key"])
	if err != nil {
		return err
	}

	o.ValueType, err = coerce.ToString(values["valueType"])
	if err != nil {
		return err
	}

	o.AckMode, err = coerce.ToString(values["ackMode"])
	if err != nil {
		return err
	}

	o.StringValue, err = coerce.ToString(values["stringValue"])
	if err != nil {
		return err
	}

	if values["partition"] == nil {
		o.Partition = -1
	} else {
		o.Partition, err = coerce.ToInt32(values["partition"])
		if err != nil {
			return err
		}
	}

	o.AckTimeout, err = coerce.ToInt(values["ackTimeout"])
	if err != nil {
		return err
	}

	o.Frequency, err = coerce.ToInt(values["frequency"])
	if err != nil {
		return err
	}

	o.MaxMessages, err = coerce.ToInt(values["maxMessages"])
	if err != nil {
		return err
	}

	o.MaxRequestSize, err = coerce.ToInt(values["maxRequestSize"])
	if err != nil {
		return err
	}

	o.Headers, err = coerce.ToObject(values["headers"])
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

	o.Connection, err = kafka.GetSharedConfiguration(values["kafkaConnection"])
	if err != nil {
		return err
	}

	return nil

}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"partition": o.Partition,
		"offset":    o.Offset,
		"topic":     o.Topic,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.Topic, err = coerce.ToString(values["topic"])
	if err != nil {
		return err
	}

	o.Offset, err = coerce.ToInt64(values["offset"])
	if err != nil {
		return err
	}

	o.Partition, err = coerce.ToInt32(values["partition"])
	if err != nil {
		return err
	}

	return nil
}
