package services

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/cloudspannerecosystem/dynamodb-adapter/config"
	"github.com/cloudspannerecosystem/dynamodb-adapter/models"
	"github.com/cloudspannerecosystem/dynamodb-adapter/pkg/logger"
	uuid "github.com/satori/go.uuid"
)

var pubsubClient *pubsub.Client
var mClients = map[string]*pubsub.Topic{}
var mux = &sync.Mutex{}

func InitStream() {
	var err error
	pubsubClient, err = pubsub.NewClient(context.Background(), config.ConfigurationMap.GOOGLE_PROJECT_ID)
	if err != nil {
		logger.LogFatal(err)
	}
}

func StreamDataToThirdParty(oldImage, newImage map[string]interface{}, tableName string) {
	if !IsMyStreamEnabled(tableName) {
		return
	}
	if (oldImage == nil || len(oldImage) == 0) && newImage == nil || len(newImage) == 0 {
		return
	}
	streamObj := models.StreamDataModel{}
	tableConf, err := config.GetTableConf(tableName)
	if err == nil {
		streamObj.Keys = map[string]interface{}{}
		if oldImage != nil && len(oldImage) > 0 {
			streamObj.Keys[tableConf.PartitionKey] = oldImage[tableConf.PartitionKey]
			if tableConf.SortKey != "" {
				streamObj.Keys[tableConf.SortKey] = oldImage[tableConf.SortKey]
			}
		} else {
			streamObj.Keys[tableConf.PartitionKey] = newImage[tableConf.PartitionKey]
			if tableConf.SortKey != "" {
				streamObj.Keys[tableConf.SortKey] = newImage[tableConf.SortKey]
			}
		}
	}
	streamObj.EventID = uuid.NewV1().String()
	streamObj.EventSourceArn = "arn:aws:dynamodb:us-east-2:123456789012:table/" + tableName
	streamObj.OldImage = oldImage
	streamObj.NewImage = newImage
	streamObj.Timestamp = time.Now().UnixNano()
	streamObj.SequenceNumber = streamObj.Timestamp
	streamObj.Table = tableName
	if oldImage == nil || len(oldImage) == 0 {
		streamObj.EventName = "INSERT"
	} else if newImage == nil || len(newImage) == 0 {
		streamObj.EventName = "REMOVE"
	} else {
		streamObj.EventName = "MODIFY"
	}
	connectors(&streamObj)
}

func connectors(streamObj *models.StreamDataModel) {
	go pubsubPublish(streamObj)
}

func pubsubPublish(streamObj *models.StreamDataModel) {
	topicName, status := IsPubSubAllowed(streamObj.Table)
	if !status {
		return
	}
	mux.Lock()
	defer mux.Unlock()
	topic, ok := mClients[topicName]
	if !ok {
		topic = pubsubClient.
			TopicInProject(topicName, config.ConfigurationMap.GOOGLE_PROJECT_ID)
		mClients[topicName] = topic
	}
	message := &pubsub.Message{}
	message.Data, _ = json.Marshal(streamObj)
	_, err := topic.Publish(context.Background(), message).Get(ctx)
	if err != nil {
		logger.LogError(err)
	}
}
