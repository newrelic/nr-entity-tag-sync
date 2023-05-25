package provider

import (
	"fmt"
	"sync"
	"time"

	"github.com/newrelic/nr-entity-tag-sync/pkg/interop"
	"github.com/spf13/viper"
)

type Entity struct {
  ID        string
  Tags      map[string]interface{}
}

type Provider interface {
  GetEntities(
    config          map[string]interface{},
    tags            []string,
    lastSync        time.Time,
  ) ([]Entity, error)
}

type InitFn func (*interop.Interop, *viper.Viper) (Provider, error)

var (
  initFns map[string]InitFn
  providerLock sync.Mutex
)

func GetProvider(i *interop.Interop) (Provider, error) {
  if !viper.IsSet("provider") {
    return nil, fmt.Errorf("missing provider in config")
  }

  providerType := viper.GetString("provider.type")
  if providerType == "" {
    return nil, fmt.Errorf("missing provider type")
  }

  i.Logger.Debugf("getting provider for type %s...", providerType)

  providerLock.Lock()
  defer providerLock.Unlock()

  fn, ok := initFns[providerType]
  if !ok {
    return nil, fmt.Errorf("invalid provider: %s", providerType)
  }

  i.Logger.Debugf("initializing provider...")
  return fn(i, viper.Sub("provider"))
}

func RegisterProvider(t string, initFn InitFn) {
  providerLock.Lock()
  defer providerLock.Unlock()

  if initFns == nil {
    initFns = make(map[string]InitFn)
  }

  initFns[t] = initFn
}
