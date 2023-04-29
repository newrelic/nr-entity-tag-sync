package sync

import (
	"strconv"
	"strings"

	"github.com/newrelic/nr-entity-tag-sync/internal/provider"
	"github.com/newrelic/nr-entity-tag-sync/pkg/interop"
	"github.com/newrelic/nr-entity-tag-sync/pkg/nerdgraph"
	"github.com/spf13/viper"
)

func getNestedHelper(
  path []string,
  m map[string]interface{},
  index int,
) string {
  if index >= len(path) {
    return ""
  }

  v, ok := m[path[index]]
  if !ok {
    return ""
  }

  switch u := v.(type) {
  case string:
    if index + 1 == len(path) {
      return u
    }
    return ""

  case map[string]interface{}:
    return getNestedHelper(path, u, index + 1)
  }

  return ""
}

func getNestedKeyValue(path string, m map[string]interface{}) string {
  return getNestedHelper(strings.Split(path, "."), m, 0)
}

func getKeys(m map[string]string) []string {
  keys := make([]string, len(m))

  i := 0
  for k := range m {
    keys[i] = k
    i += 1
  }

  return keys
}

func getEntityTagValue(
  i *interop.Interop,
  entity *nerdgraph.EntityOutline,
  tagName string,
) string {
  for _, tag := range entity.Tags {
    if tag.Key == tagName {
      // TODO: for now use first value
      return tag.Values[0]
    }
  }

  return ""
}

func getEntityKeyValue(
  entityKey string,
  entity *nerdgraph.EntityOutline,
) string {
  if strings.EqualFold(entityKey, "name") {
    return entity.Name
  } else if strings.EqualFold(entityKey, "guid") {
    return entity.Guid
  } else if strings.EqualFold(entityKey, "accountId") {
    return strconv.Itoa(entity.AccountId)
  }

  for _, tag := range entity.Tags {
    if strings.EqualFold(entityKey, tag.Key) && len(tag.Values) > 0 {
      return tag.Values[0]
    }
  }

  return ""
}

func getExtEntityTagValue(
  i *interop.Interop,
  extEntity *provider.Entity,
  tagName string,
) string {
  return getNestedKeyValue(tagName, extEntity.Tags)
}

func getEntities(i *interop.Interop, mapping *MappingConfig) (
  []nerdgraph.EntityOutline,
  error,
) {
  nrEntityQuery := &nerdgraph.EntityQuery{
    Type: mapping.EntityQuery.Type,
    Domain: mapping.EntityQuery.Domain,
    Name: mapping.EntityQuery.Name,
    AccountId: mapping.EntityQuery.AccountId,
    Tags: nil,
    Query: mapping.EntityQuery.Query,
  }

  if len(mapping.EntityQuery.Tags) > 0 {
    for _, tag := range mapping.EntityQuery.Tags {
      nrEntityQuery.Tags = append(
        nrEntityQuery.Tags,
        nerdgraph.Tag{Key: tag.Key, Values: tag.Values},
      )
    }
  }

  return i.Nerdgraph.GetEntities(nrEntityQuery)
}

func getMatchingEntity(
  i *interop.Interop,
  entity *nerdgraph.EntityOutline,
  match *Match,
  extEntities []provider.Entity,
) *provider.Entity {
  entityKeyValue := getEntityKeyValue(match.EntityKey, entity)
  if entityKeyValue == "" {
    // This entity does not have a value for the entity match key
    i.Logger.Debugf(
      "skipping entity %s because it does not have the match key %s or the match key is not a string value",
      entity.Guid,
      match.EntityKey,
    )
    return nil
  }

  for _, extEntity := range extEntities {
    extEntityKeyValue := getExtEntityTagValue(i, &extEntity, match.ExtEntityKey)
    if extEntityKeyValue == "" {
      // This external entity does not have a value for the external entity
      // match key or the value is not a string
      i.Logger.Debugf(
        "skipping external entity %s because it does not have the match key %s or the match key is not a string value",
        extEntity.ID,
        match.ExtEntityKey,
      )
      continue
    }

    i.Logger.Debugf(
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

func updateTags(
  i *interop.Interop,
  mapping Mapping,
  extEntity *provider.Entity,
  entity *nerdgraph.EntityOutline,
) error {
  tagsToDelete := []string{}
  tagsToAdd := []nerdgraph.Tag{}

  for extEntityTagName, entityTagName := range mapping {
    extEntityTagValue := getExtEntityTagValue(i, extEntity, extEntityTagName)
    entityTagValue := getEntityTagValue(i, entity, entityTagName)

    if extEntityTagValue == "" && entityTagValue != "" {
      // No tag on external entity but internal entity has one, delete on entity
      tagsToDelete = append(tagsToDelete, entityTagName)
    } else if extEntityTagValue != "" && entityTagValue == "" {
      // Tag on external entity but not on internal entity, add to entity
      tagsToAdd = append(tagsToAdd, nerdgraph.Tag{
        Key: entityTagName,
        Values: []string {extEntityTagValue},
      })
    } else if
      (extEntityTagValue != "" && entityTagValue != "") &&
      (extEntityTagValue != entityTagValue) {
      tagsToDelete = append(tagsToDelete, entityTagName)
      tagsToAdd = append(tagsToAdd, nerdgraph.Tag{
        Key: entityTagName,
        Values: []string {extEntityTagValue},
      })
    }
  }

  if len(tagsToDelete) > 0 {
    err := i.Nerdgraph.DeleteTags(entity, tagsToDelete)
    if err != nil {
      return err
    }
  }

  if len(tagsToAdd) > 0 {
    err := i.Nerdgraph.AddTags(entity, tagsToAdd)
    if err != nil {
      return err
    }
  }

  return nil
}

func Sync(i *interop.Interop) error {
  mappings := Mappings{}

  err := viper.UnmarshalKey("mappings", &mappings)
  if err != nil {
    return err
  }

  p, err := provider.GetProvider(i)
  if err != nil {
    return err
  }

  //updates := []EntityUpdate

  for _, mappingConfig := range mappings {
    i.Logger.Debugf("reading all external entities from provider")

    extEntityTags := []string { mappingConfig.Match.ExtEntityKey }
    extEntityTags = append(extEntityTags, getKeys(mappingConfig.Mapping)...)

    extEntities, err := p.GetEntities(
      mappingConfig.ExtEntityQuery,
      extEntityTags,
    )
    if err != nil {
      return err
    }

    i.Logger.Debugf("read %d entities from provider", len(extEntities))

    entities, err := getEntities(i, &mappingConfig)
    if err != nil {
      return err
    }

    i.Logger.Debugf("found %d entities in New Relic", len(entities))

    for _, entity := range entities {
      extEntity := getMatchingEntity(
        i,
        &entity,
        &mappingConfig.Match,
        extEntities,
      )
      if extEntity == nil {
        // No entity with a value for extEntityKey that maches an entity with a
        // value for entityKey
        continue
      }

      i.Logger.Debugf(
        "external entity %s matches entity %s",
        extEntity.ID,
        entity.Guid,
      )

      err = updateTags(i, mappingConfig.Mapping, extEntity, &entity)
      if err != nil {
        return err
      }
    }
  }

  return nil
}
