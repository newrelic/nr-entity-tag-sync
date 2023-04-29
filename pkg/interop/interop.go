package interop

import (
	"fmt"
	"os"
	"time"

	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/nrlogrus"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/newrelic/nr-entity-tag-sync/pkg/nerdgraph"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Interop struct {
  App               *newrelic.Application
  ApiKey            string
  ApiURL            string
  Logger            *log.Logger
  Nerdgraph         *nerdgraph.NerdgraphClient
}

func NewInteroperability() (*Interop, error) {
  app, err := newrelic.NewApplication(
    newrelic.ConfigAppName("New Relic Entity CMDB Sync"),
    newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
  )
  if err != nil {
		return nil, err
  }

  logger := log.New()

  logger.SetLevel(log.WarnLevel)
  logger.SetFormatter(nrlogrus.NewFormatter(app, &log.TextFormatter{}))

  viper.SetConfigName("config")
  viper.AddConfigPath("configs")
  viper.AddConfigPath(".")

  err = viper.ReadInConfig()
  if err != nil {
    return nil, err
  }

  setupLogging(logger)

  apiKey := viper.GetString("apiKey")
  if apiKey == "" {
    apiKey = os.Getenv("NEW_RELIC_API_KEY")
    if apiKey == "" {
      return nil, fmt.Errorf("missing New Relic API key")
    }
  }

  apiUrl := viper.GetString("apiUrl")
  if apiUrl == "" {
    apiUrl = os.Getenv("NEW_RELIC_API_URL")
    if apiUrl == "" {
      apiUrl = "https://api.newrelic.com"
    }
  }

  nerdgraph := &nerdgraph.NerdgraphClient{
    ApiURL: apiUrl,
    ApiKey: apiKey,
    Logger: logger,
  }

  return &Interop{app, apiKey, apiUrl, logger, nerdgraph}, nil
}

func (i *Interop) Shutdown() {
  i.App.Shutdown(time.Second * 3)
}

func setupLogging(logger *log.Logger) {
  logLevel := viper.GetString("log.level")
  if logLevel != "" {
    level, err := log.ParseLevel(logLevel)
    if err != nil {
      log.Infof("failed to parse log level, default will be used: %s", err)
    } else {
      logger.SetLevel(level)
    }
  }

  if viper.IsSet("log.fileName") {
    file, err := os.OpenFile(
      viper.GetString("log.fileName"),
      os.O_CREATE|os.O_WRONLY|os.O_APPEND,
      0666,
    )
    if err != nil {
      log.Infof("failed to log to file, using default stderr: %s", err)
    } else {
      logger.Out = file
    }
  }
}
