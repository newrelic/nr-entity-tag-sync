package sync

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/newrelic/nr-entity-tag-sync/internal/provider"
	"github.com/newrelic/nr-entity-tag-sync/pkg/interop"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Syncer struct {
  i                 *interop.Interop
  log               *logrus.Logger
  mappings          Mappings
  provider          provider.Provider
  useLastUpdate     bool
  eventsConfig      *eventsConfig
}

func New(i *interop.Interop) (*Syncer, error) {
  mappings := Mappings{}

  err := viper.UnmarshalKey("mappings", &mappings)
  if err != nil {
    return nil, err
  }

  p, err := provider.GetProvider(i)
  if err != nil {
    return nil, err
  }

  events := &eventsConfig{}

  err = viper.UnmarshalKey("events", events)
  if err != nil {
    return nil, fmt.Errorf("error parsing events config: %v", err)
  }

  if events.Enabled {
    if err := requireAccountID(events); err != nil {
      return nil, err
    }

    if err = i.EnableEvents(events.AccountId); err != nil {
      return nil, err
    }

    if events.EventType == "" {
      events.EventType = "EntityTagSync"
    }
  }

  return &Syncer{
    i,
    i.Logger,
    mappings,
    p,
    viper.GetBool("provider.useLastUpdate"),
    events,
  }, nil
}

func (s *Syncer) Sync() error {
	cycleId, err := uuid.NewV4()
	if err != nil {
		s.syncFailed(uuid.Nil, err)
	}

  s.syncStarted(cycleId)

  lastUpdateTs, err := s.getLastUpdateTimestamp()
  if err != nil {
    return s.syncFailed(cycleId, err)
  }

  errorCount := 0

  for i, mappingConfig := range s.mappings {
    s.log.Debugf(
      "starting mapping %d; reading all external entities from provider",
      i,
    )

    extEntityTags := []string { mappingConfig.Match.ExtEntityKey }
    extEntityTags = append(extEntityTags, getKeys(mappingConfig.Mapping)...)

    extEntities, err := s.provider.GetEntities(
      mappingConfig.ExtEntityQuery,
      extEntityTags,
      lastUpdateTs,
    )
    if err != nil {
      s.mappingFailed(cycleId, fmt.Errorf("reading entities from provider failed: %v", err))
      errorCount += 1
      continue
    }

    extEntityCount := len(extEntities)

    if extEntityCount == 0 {
      s.mappingSkipped(cycleId)
      continue
    }

    s.log.Debugf("read %d entities from provider", extEntityCount)

    processingResults, err := processEntities(
      s.i,
      &mappingConfig,
      func (
        i                 *interop.Interop,
        mapping           *MappingConfig,
        entity            *EntityOutline,
      ) (entityProcessorResult, []error) {
        extEntity := getMatchingEntity(
          s.i,
          entity,
          &mapping.Match,
          extEntities,
        )
        if extEntity == nil {
          // No entity with a value for extEntityKey that maches an entity with a
          // value for entityKey
          return ENTITY_NO_MATCH, nil
        }

        s.log.Debugf(
          "external entity %s matches New Relic entity %s (%s)",
          extEntity.ID,
          entity.Name,
          entity.Guid,
        )

        return updateTags(
          s.i,
          mappingConfig.Mapping,
          extEntity,
          entity,
        )
      },
    )

    s.mappingComplete(cycleId, extEntityCount, processingResults, err)

    if err != nil || processingResults.totalEntitiesWithErrors > 0 {
      errorCount += 1
    }
  }

  if errorCount > 0 {
    return s.syncFailed(cycleId, fmt.Errorf("sync completed with errors"))
  }

  s.syncComplete(cycleId)

  return nil
}

func (s *Syncer) syncStarted(uuid uuid.UUID) {
  if s.eventsConfig.Enabled {
    startEvent := s.newAuditEvent(uuid, "sync_start", nil)
    s.pushEvent(startEvent)
  }

  s.log.Debugf("sync started")
}

func (s *Syncer) syncFailed(uuid uuid.UUID, err error) error {
  if s.eventsConfig.Enabled {
    endEvent := s.newAuditEvent(uuid, "sync_end", err)
    s.pushEvent(endEvent)
  }

  s.log.Debugf("sync failed")

  return err
}

func (s *Syncer) syncComplete(uuid uuid.UUID) {
  if s.eventsConfig.Enabled {
    endEvent := s.newAuditEvent(uuid, "sync_end", nil)
    s.pushEvent(endEvent)
  }

  s.log.Debugf("sync complete")
}

func (s *Syncer) mappingFailed(uuid uuid.UUID, err error) {
  if s.eventsConfig.Enabled {
    mappingEvent := s.newAuditEvent(uuid, "mapping_complete", err)
    s.pushEvent(mappingEvent)
  }
  s.log.Error(fmt.Sprintf("mapping failed: %v", err))
}

func (s *Syncer) mappingSkipped(uuid uuid.UUID) {
  if s.eventsConfig.Enabled {
    mappingEvent := s.newAuditEvent(uuid, "mapping_complete", nil)

    mappingEvent["extEntityCount"] = 0

    s.pushEvent(mappingEvent)
  }

  s.log.Debug("no entities returned from provider; continuing to next mapping")
}

func (s *Syncer) mappingComplete(
  uuid                uuid.UUID,
  extEntityCount      int,
  processingResults   *entityProcessingResult,
  err                 error,
) {
  if s.eventsConfig.Enabled {
    mappingEvent := s.newAuditEvent(uuid, "mapping_complete", err)

    mappingEvent["extEntityCount"] = extEntityCount
    mappingEvent["totalEntityCount"] = processingResults.totalEntities
    mappingEvent["totalEntitiesScanned"] = processingResults.totalEntitiesScanned
    mappingEvent["totalEntitiesMatched"] = processingResults.totalEntitiesMatched
    mappingEvent["totalEntitiesNoMatch"] = processingResults.totalEntitiesNoMatch
    mappingEvent["totalEntitiesSkipped"] = processingResults.totalEntitiesSkipped
    mappingEvent["totalEntitiesUpdated"] = processingResults.totalEntitiesUpdated
    mappingEvent["totalEntitiesWithErrors"] = processingResults.totalEntitiesWithErrors

    s.pushEvent(mappingEvent)
  }

  if err != nil {
    s.log.Warnf(
      "mapping completed with an error; results may be incomplete; see output for details: %v",
      err,
    )
  }

  s.log.Debugf(
    "read %d external entities, %d total New Relic entities, %d scanned, %d matched, %d skipped, %d updated, %d updates with errors",
    extEntityCount,
    processingResults.totalEntities,
    processingResults.totalEntitiesScanned,
    processingResults.totalEntitiesMatched,
    processingResults.totalEntitiesSkipped,
    processingResults.totalEntitiesUpdated,
    processingResults.totalEntitiesWithErrors,
  )
}

func getMatchingEntity(
  i                 *interop.Interop,
  entity            *EntityOutline,
  match             *Match,
  extEntities       []provider.Entity,
) *provider.Entity {
  entityKeyValue, entityKeyExists := getEntityKeyValue(
    i,
    entity,
    match.EntityKey,
  )
  if !entityKeyExists || entityKeyValue == "" {
    // This entity does not have a value for the entity match key
    i.Logger.Tracef(
      "skipping entity %s (%s) because it does not have the match key %s or the match key is not a string value",
      entity.Name,
      entity.Guid,
      match.EntityKey,
    )
    return nil
  }

  for _, extEntity := range extEntities {
    extEntityKeyValue, extEntityKeyExists := getExtEntityKeyValue(
      i,
      &extEntity,
      match.ExtEntityKey,
    )
    if !extEntityKeyExists || extEntityKeyValue == "" {
      // This external entity does not have a value for the external entity
      // match key or the value is not a string
      i.Logger.Tracef(
        "skipping external entity %s because it does not have the match key %s or the match key is not a string value",
        extEntity.ID,
        match.ExtEntityKey,
      )
      continue
    }

    i.Logger.Tracef(
      "comparing external entity key %s against entity key %s using strategy %s",
      extEntityKeyValue,
      entityKeyValue,
      match.Operator,
    )

    switch match.Operator {
    case "equal":
      if extEntityKeyValue == entityKeyValue {
        return &extEntity
      }

    case "equal-ignore-case":
      if strings.EqualFold(extEntityKeyValue, entityKeyValue) {
        return &extEntity
      }

    case "contains":
      if strings.Contains(extEntityKeyValue, entityKeyValue) {
        return &extEntity
      }

    case "contains-ignore-case":
      if strings.Contains(
        strings.ToLower(extEntityKeyValue),
        strings.ToLower(entityKeyValue),
      ) {
        return &extEntity
      }

    case "inverse-contains-ignore-case":
      if strings.Contains(
        strings.ToLower(entityKeyValue),
        strings.ToLower(extEntityKeyValue),
      ) {
        return &extEntity
      }
    }
  }

  return nil
}

func requireAccountID(events *eventsConfig) error {
  if events.AccountId == 0 {
    eventsAccountID := os.Getenv("NEW_RELIC_ACCOUNT_ID")
    if eventsAccountID == "" {
      return fmt.Errorf("missing New Relic account ID")
    }

    accountID, err := strconv.Atoi(eventsAccountID)
    if err != nil {
      return fmt.Errorf("invalid New Relic account ID %s", eventsAccountID)
    }

    events.AccountId = accountID
  }

  return nil
}
