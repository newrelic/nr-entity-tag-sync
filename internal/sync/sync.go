package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

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

  return &Syncer{i, i.Logger, mappings, p}, nil
}

func (s *Syncer) Sync() error {
  for _, mappingConfig := range s.mappings {
    s.log.Debugf("reading all external entities from provider")

    extEntityTags := []string { mappingConfig.Match.ExtEntityKey }
    extEntityTags = append(extEntityTags, getKeys(mappingConfig.Mapping)...)

    extEntities, err := s.provider.GetEntities(
      mappingConfig.ExtEntityQuery,
      extEntityTags,
      time.Now(),
    )
    if err != nil {
      return err
    }

    s.log.Debugf("read %d entities from provider", len(extEntities))

    errMessage := ""
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
    if err != nil {
      errMessage := fmt.Sprintf("error while processing entities, sync results may be incomplete: %s", err)
      s.log.Error(errMessage)
    }

    if s.i.IsEventsEnabled() {
      event := struct {
        EventType                 string  `json:"eventType"`
        Action                    string  `json:"action"`
        Error                     bool    `json:"error"`
        ErrorMessage              string  `json:"errorMessage"`
        TotalExtEntities          int     `json:"totalExternalEntities"`
        TotalEntities             int     `json:"totalEntities"`
        TotalEntitiesScanned      int     `json:"totalEntitiesScanned"`
        TotalEntitiesMatched      int     `json:"totalEntitiesMatched"`
        TotalEntitiesNoMatch      int     `json:"totalEntitiesNoMatch"`
        TotalEntitiesSkipped      int     `json:"totalEntitiesSkipped"`
        TotalEntityUpdates        int     `json:"totalEntityUpdates"`
        TotalEntityUpdateErrors   int     `json:"totalEntityUpdateErrors"`
      } {
        EventType:                  "MyEventType",
        Action:                     "mapping_complete",
        Error:                      err != nil,
        ErrorMessage:               errMessage,
        TotalExtEntities:           len(extEntities),
        TotalEntities:              processingResults.totalEntities,
        TotalEntitiesScanned:       processingResults.totalEntitiesScanned,
        TotalEntitiesMatched:       processingResults.totalEntitiesMatched,
        TotalEntitiesNoMatch:       processingResults.totalEntitiesNoMatch,
        TotalEntitiesSkipped:       processingResults.totalEntitiesSkipped,
        TotalEntityUpdates:         processingResults.totalEntitiesUpdated,
        TotalEntityUpdateErrors:    processingResults.totalEntitiesWithErrors,
      }

      if err := s.i.NrClient.Events.EnqueueEvent(context.Background(), event);
        err != nil {
        s.log.Warnf("failed to push event: %s", err)
      }

      s.i.Logger.Debugf(
        "read %d total New Relic entities, %d scanned, %d matched, %d skipped, %d updated, %d updates with errors",
        processingResults.totalEntities,
        processingResults.totalEntitiesScanned,
        processingResults.totalEntitiesMatched,
        processingResults.totalEntitiesSkipped,
        processingResults.totalEntitiesUpdated,
        processingResults.totalEntitiesWithErrors,
      )
    }
  }

  return nil
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
