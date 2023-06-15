package servicenow

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/newrelic/nr-entity-tag-sync/internal/provider"
	"github.com/newrelic/nr-entity-tag-sync/pkg/interop"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type ServiceNowProvider struct {
  Interop           *interop.Interop
  ApiURL            string
  ApiUser           string
  ApiPassword       string
  PageSize          int
}

var (
  dateRE            *regexp.Regexp
  timeRE            *regexp.Regexp
)

func init() {
  provider.RegisterProvider("servicenow", New)
  dateRE = regexp.MustCompile(`(?i)\${lastUpdateDate}`)
  timeRE = regexp.MustCompile(`(?i)\${lastUpdateTime}`)
}

func New(i *interop.Interop, v *viper.Viper) (provider.Provider, error) {
  v.AutomaticEnv()
  v.SetEnvPrefix("NR_CMDB_SNOW")

  apiUrl := v.GetString("apiUrl")
  if apiUrl == "" {
    return nil, fmt.Errorf("missing servicenow api url")
  }

  apiUser := v.GetString("apiUser")
  if apiUser =="" {
    return nil, fmt.Errorf("missing servicenow api user")
  }

  apiPassword := v.GetString("apiPassword")
  if apiPassword == "" {
    return nil, fmt.Errorf("missing servicenow api password")
  }

  pageSize := v.GetInt("pageSize")
  if pageSize <= 0 {
    pageSize = 10000
  }

  return &ServiceNowProvider{
    i,
    apiUrl,
    apiUser,
    apiPassword,
    pageSize,
  }, nil
}

func (snp *ServiceNowProvider) GetEntities(
  config            map[string]interface{},
  tags              []string,
  lastUpdate        *time.Time,
) (
  []provider.Entity,
  error,
) {
  ciType := cast.ToString(config["type"])
  if ciType == "" {
    return nil, fmt.Errorf("missing CI type field")
  }

  var err error

  ciQuery := cast.ToString(config["query"])
  if ciQuery != "" && lastUpdate != nil {
    ciQuery, err = subsDateTime(ciQuery, config, lastUpdate)
    if err != nil {
      return nil, fmt.Errorf("query datetime substitution failed: %v", err)
    }
  }

  newTags := []string{}

  for _, tag := range tags {
    if index := strings.Index(tag, "."); index > 0 {
      newTags = append(newTags, tag[0:index])
      continue
    }
    newTags = append(newTags, tag)
  }

  items, err := snp.getRecords(ciType, ciQuery, newTags)
  if err != nil {
    return nil, fmt.Errorf("get records failed: %s", err)
  }

  var entities []provider.Entity

  for _, item := range items {
    v, ok := item["sys_id"]
    if !ok {
      snp.Interop.Logger.Warn("skipping CI with no sys_id")
      continue
    }

    sysId, ok := v.(string)
    if !ok {
      snp.Interop.Logger.Warn("skipping CI with non-string sys_id")
      continue
    }

    entities = append(
      entities,
      provider.Entity{ID: sysId, Tags: item},
    )
  }

  return entities, nil
}

func subsDateTime(
  query             string,
  config            map[string]interface{},
  lastUpdate        *time.Time,
 ) (string, error) {
  loc, err := getServerTimezone(config)
  if err != nil {
    return "", err
  }

  lastUpdateDate := lastUpdate.In(loc).Format("2006-01-02")
  lastUpdateTime := lastUpdate.In(loc).Format("15:04:05")

  s := dateRE.ReplaceAllLiteralString(query, lastUpdateDate)

  return timeRE.ReplaceAllLiteralString(s, lastUpdateTime), nil
}

func getServerTimezone(config map[string]interface{}) (*time.Location, error) {
  val, ok := config["servertimezone"]
  if !ok {
    return time.LoadLocation("")
  }

  timezone, ok:= val.(string)
  if !ok {
    return time.LoadLocation("")
  }

  return time.LoadLocation(timezone)
}
