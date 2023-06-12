package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/newrelic/newrelic-client-go/pkg/nrdb"
)

type auditEvent map[string]interface{}

type eventsConfig struct {
  Enabled           bool
  AccountId         int
  EventType         string
}

func (s *Syncer) newAuditEvent(
  uuid uuid.UUID,
  action string,
  err error,
) auditEvent {
  event := auditEvent{}

  event["eventType"] = s.eventsConfig.EventType
  event["id"] = uuid.String()
  event["action"] = action
  event["error"] = err != nil
  if err != nil {
    event["errorMessage"] = err.Error()
  }

  return event
}

func (s *Syncer ) pushEvent(event auditEvent) {
  if err := s.i.NrClient.Events.EnqueueEvent(context.Background(), event);
    err != nil {
    s.log.Warnf("failed to push event: %s", err)
  }
}

func (s *Syncer) getLastUpdateTimestamp() (*time.Time, error) {
  if !s.useLastUpdate {
    return nil, nil
  }

  if !s.eventsConfig.Enabled {
    return nil, fmt.Errorf("events must be enabled to use last timestamp")
  }

  s.log.Tracef(
    "querying for latest timestamp for event %s",
    s.eventsConfig.EventType,
  )

  result, err := s.i.NrClient.Nrdb.Query(
    s.eventsConfig.AccountId,
    nrdb.NRQL(fmt.Sprintf(
      "SELECT latest(timestamp) FROM %s WHERE action = 'sync_end' SINCE 1 MONTH AGO",
      s.eventsConfig.EventType,
    )),
  )
  if err != nil {
    return nil, fmt.Errorf("query for last update failed: %s", err)
  }

  if len(result.Results) == 0 {
    s.log.Warn("no results found searching for last update timetaamp")
    return nil, nil
  }

  row := result.Results[0]
  val, ok := row["latest.timestamp"]
  if !ok {
    s.log.Warn("no timestamp attribute found in result")
    return nil, nil
  }

  latestTimestamp, ok := val.(float64)
  if !ok {
    s.log.Warn("timestamp attribute found in result is not an integer")
    return nil, nil
  }

  s.log.Tracef("found latest timestamp %f", latestTimestamp)
  t := time.UnixMilli(int64(latestTimestamp))

  return &t, nil
}
