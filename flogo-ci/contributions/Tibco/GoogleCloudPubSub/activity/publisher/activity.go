package publisher

import (
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	_ = activity.Register(&MessagePublisher{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})
var permissions = []string{"pubsub.topics.publish"}

type MessagePublisher struct {
	topic         *pubsub.Topic
	dynamicTopics map[string]*pubsub.Topic
	settings      *Settings
	client        *pubsub.Client
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		ctx.Logger().Errorf("Failed to read activity settings due to error - %s", err.Error())
		return nil, err
	}
	client := s.GooglePubSubConnection.GetConnection().(*pubsub.Client)
	act := &MessagePublisher{settings: s, client: client, dynamicTopics: make(map[string]*pubsub.Topic)}
	if s.Topic != "" {
		act.topic, err = act.createTopic(ctx.Logger(), s.Topic)
		if err != nil {
			return nil, err
		}
	}
	return act, nil
}

func (mp *MessagePublisher) createTopic(logger log.Logger, topicId string) (*pubsub.Topic, error) {
	if strings.Contains(topicId, "/") && strings.HasPrefix(topicId, "projects/") {
		logger.Warnf("Topic name '%s' found instead of Topic Id", topicId)
		segments := strings.Split(mp.settings.Topic, "/")
		topicId = segments[len(segments)-1]
	}

	logger.Infof("Topic Id set to '%s'", topicId)
	topic := mp.client.Topic(topicId)
	// Check required permissions
	perms, err := topic.IAM().TestPermissions(context.Background(), permissions)
	if err == nil {
		if len(perms) != 1 {
			// Required permissions not configured
			logger.Error("The IAM role configured in the service account must have ['pubsub.topics.publish'] permissions.")
			return nil, errors.New("insufficient IAM role permissions")
		}
	} else {
		if status.Code(err) == codes.NotFound {
			logger.Errorf("Topic(Id:%s) not found in Google project. Check Topic Id and Google project Id.", topicId)
			return nil, errors.New("topic not found")
		} else {
			logger.Errorf("Unable to validate required permissions due to error - %s", err.Error())
			return nil, errors.New("failed to validate required permissions")
		}
	}
	topic.EnableMessageOrdering = mp.settings.MessageOrdering
	return topic, err
}

func (mp *MessagePublisher) Metadata() *activity.Metadata {
	return activityMd
}

func (mp *MessagePublisher) Eval(ctx activity.Context) (done bool, err error) {
	input := &Input{}

	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}

	ctx.Logger().Debugf("Sending message to Google Cloud Pub/Sub service")

	topic := mp.topic
	if input.TopicId != "" || topic == nil {
		if input.TopicId == "" {
			return false, activity.NewError("Topic Id is not configured", "", nil)
		}
		// Dynamic overriding of topic Id
		topic, _ = mp.dynamicTopics[input.TopicId]
		if topic == nil {
			// cache topics to optimize performance
			topic, err = mp.createTopic(ctx.Logger(), input.TopicId)
			if err != nil {
				return false, err
			}
			mp.dynamicTopics[input.TopicId] = topic
		}
	}

	if input.MessageData == nil {
		return false, activity.NewError("Message data is not configured", "", nil)
	}

	msg := &pubsub.Message{}
	msg.Data, err = coerce.ToBytes(input.MessageData)
	if err != nil {
		ctx.Logger().Errorf("Failed to process input data due to error - %s", err.Error())
		return false, activity.NewError("Failed to process input data", "", nil)
	}
	if input.MessageAttributes != nil {
		msg.Attributes = input.MessageAttributes
	}

	if topic.EnableMessageOrdering && input.MessageOrderingKey != "" {
		//Ordering Key is set
		msg.OrderingKey = input.MessageOrderingKey
	}
	//Send message to Pub/Sub service
	output := &Output{}
	response := topic.Publish(context.Background(), msg)
	output.MessageId, err = response.Get(context.Background())
	if err != nil {
		ctx.Logger().Errorf("Failed to send message to Google Cloud Pub/Sub service due to error - %s", err.Error())
		if status.Code(err) == codes.DeadlineExceeded || status.Code(err) == codes.Unavailable || status.Code(err) == codes.Internal || status.Code(err) == codes.ResourceExhausted {
			// As per https://cloud.google.com/pubsub/docs/reference/error-codes, above are retryable errors
			return false, activity.NewRetriableError("Failed to send message. Retrying the operation.", "", err.Error())
		}
		return false, err
	}
	ctx.Logger().Infof("Message(Id:%s) successfully published on the Topic(Id:%s)", output.MessageId, topic.ID())
	err = ctx.SetOutputObject(output)
	if err != nil {
		return false, fmt.Errorf("error setting output for Activity [%s]: %s", ctx.Name(), err.Error())
	}
	return true, nil
}

func (mp *MessagePublisher) Cleanup() error {
	if mp.topic != nil {
		// Stop topic goroutine
		mp.topic.Stop()
	}
	for _, topic := range mp.dynamicTopics {
		if topic != nil {
			// Stop dynamically created topics
			topic.Stop()
		}
	}
	return nil
}
