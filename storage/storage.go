package storage

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	spannerAdmin "cloud.google.com/go/spanner/apiv1"

	"cloud.google.com/go/spanner"
	"github.com/cloudspannerecosystem/dynamodb-adapter/config"
	"github.com/cloudspannerecosystem/dynamodb-adapter/models"
	"github.com/cloudspannerecosystem/dynamodb-adapter/pkg/logger"
	"github.com/tidwall/gjson"
)

var serviceName string = "DYNAMODB-ADAPTER"

var hostName, _ = os.Hostname()

func init() {
	hostName = strings.ReplaceAll(hostName, ".", "-")
	tokens := strings.Split(hostName, "-")
	if len(tokens) >= 3 {
		sb := strings.Builder{}
		length := len(tokens) - 2
		for i := 0; i < length; i++ {
			sb.WriteString(tokens[i])
			if i != length-1 {
				sb.WriteString("-")
			}
		}
		serviceName = sb.String()
	}
}

type Storage struct {
	spannerClient map[string]*spanner.Client
}

// storage - global instance of storage
var storage *Storage

func initSpannerDriver(instance string, m map[string]*gjson.Result) *spanner.Client {
	conf := spanner.ClientConfig{}

	str := "projects/" + config.ConfigurationMap.GOOGLE_PROJECT_ID + "/instances/" + instance + "/databases/" + config.ConfigurationMap.SPANNER_DB
	Client, err := spanner.NewClientWithConfig(context.Background(), str, conf)
	if err != nil {
		logger.LogFatal(err)
	}
	return Client
}

var admin *spannerAdmin.Client

// InitliazeDriver - this will Initliaze databases object in global map
func InitliazeDriver() {
	var err error
	admin, err = spannerAdmin.NewClient(context.Background())
	if err != nil {
		logger.LogFatal(err)
	}
	storage = new(Storage)
	storage.spannerClient = make(map[string]*spanner.Client)
	config := map[string]*gjson.Result{}
	for _, v := range models.SpannerTableMap {
		if _, ok := storage.spannerClient[v]; !ok {
			storage.spannerClient[v] = initSpannerDriver(v, config)
		}
	}
}

// Close - This gracefully returns the session pool objects, when driver gets exit signal
func (s Storage) Close() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-shutdown
	logger.LogDebug("Connection Shutdown start")
	for _, v := range s.spannerClient {
		v.Close()
	}
	logger.LogDebug("Connection shutted down")
}

var once sync.Once

// GetStorageInstance - return storage instance to call db functionalities
func GetStorageInstance() *Storage {
	once.Do(func() {
		if storage == nil {
			InitliazeDriver()
		}
	})

	return storage
}

func (s Storage) getSpannerClient(tableName string) *spanner.Client {
	return s.spannerClient[models.SpannerTableMap[changeTableNameForSP(tableName)]]
}
