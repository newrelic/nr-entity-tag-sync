package interop

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/nrlogrus"
	"github.com/newrelic/go-agent/v3/newrelic"
	nrClient "github.com/newrelic/newrelic-client-go/newrelic"
	"github.com/newrelic/newrelic-client-go/pkg/config"
	"github.com/newrelic/newrelic-client-go/pkg/logging"
	"github.com/newrelic/newrelic-client-go/pkg/region"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Interop struct {
  App               *newrelic.Application
  Logger            *log.Logger
  NrClient          *nrClient.NewRelic
  eventsEnabled     bool
}

func ConfigLicenseKey(licenseKey string) nrClient.ConfigOption {
	return func(cfg *config.Config) error {
		cfg.LicenseKey = licenseKey
		return nil
	}
}

func NewInteroperability() (*Interop, error) {
  app, err := newrelic.NewApplication(
    newrelic.ConfigAppName("New Relic Entity Tag Sync"),
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
  viper.SetDefault("events.enabled", true)

  err = viper.ReadInConfig()
  if err != nil {
    return nil, err
  }

  i := &Interop{}
  i.App = app

  setupLogging(i, logger)

  err = setupClient(i)
  if err != nil {
    return nil, err
  }

  return i, nil
}

func (i *Interop) Shutdown() {
  if i.eventsEnabled {
    err := i.sendEventsAndWait()
    if err != nil {
      i.Logger.Warnf("flush event queue to New Relic failed: %v", err)
    }
  }

  i.App.Shutdown(time.Second * 3)
}

func (i *Interop) EnableEvents(accountID int) error {
  // Start batch mode
  if err := i.NrClient.Events.BatchMode(
    context.Background(),
    accountID,
  ); err != nil {
    return fmt.Errorf("error starting batch events mode: %v", err)
  }

  i.eventsEnabled = true

  return nil
}

func setupLogging(i *Interop, logger *log.Logger) {
  logLevel := viper.GetString("log.level")
  if logLevel != "" {
    level, err := log.ParseLevel(logLevel)
    if err != nil {
      log.Warnf("failed to parse log level, default will be used: %s", err)
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
      log.Warnf("failed to log to file, using default stderr: %s", err)
    } else {
      logger.Out = file
    }
  }

  i.Logger = logger
}

func setupClient(i *Interop) error {
  licenseKey := viper.GetString("licenseKey")
  if licenseKey == "" {
    licenseKey = os.Getenv("NEW_RELIC_LICENSE_KEY")
    if licenseKey == "" {
      return fmt.Errorf("missing New Relic license key")
    }
  }

  apiKey := viper.GetString("apiKey")
  if apiKey == "" {
    apiKey = os.Getenv("NEW_RELIC_API_KEY")
    if apiKey == "" {
      return fmt.Errorf("missing New Relic API key")
    }
  }

  nrRegion := viper.GetString("region")
  if nrRegion == "" {
    nrRegion = os.Getenv("NEW_RELIC_REGION")
    if nrRegion == "" {
      nrRegion = string(region.Default)
    }
  }

  // Initialize the New Relic Go Client
  client, err := nrClient.New(
    ConfigLicenseKey(licenseKey),
    nrClient.ConfigPersonalAPIKey(apiKey),
    nrClient.ConfigRegion(nrRegion),
    nrClient.ConfigLogger(
      logging.NewLogrusLogger(logging.ConfigLoggerInstance(i.Logger)),
    ),
  )
  if err != nil {
    return fmt.Errorf("error creating New Relic client: %v", err)
  }

  i.NrClient = client

  return nil
}

func (i *Interop) sendEventsAndWait() error {
  err := i.NrClient.Events.Flush()
  if err != nil {
    return err
  }

  <-time.After(3 * time.Second)

  return nil
}
