package sync

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/nr-entity-tag-sync/internal/provider"
	"github.com/newrelic/nr-entity-tag-sync/pkg/interop"
)

func updateTags(
  i                 *interop.Interop,
  mapping           Mapping,
  extEntity         *provider.Entity,
  entity            *EntityOutline,
) (entityProcessorResult, []error) {
  tagChanges := 0
  tagsToDelete := []string{}
  tagsToAdd := []entities.TaggingTagInput{}
  tagsToUpdate := []entities.TaggingTagInput{}

  for extEntityKeyName, entityTagName := range mapping {
    extEntityKeyValue, extEntityKeyExists := getExtEntityKeyValue(
      i,
      extEntity,
      extEntityKeyName,
    )
    entityTagValues, entityTagExists := getEntityTagValues(
      i,
      entity.Tags,
      entityTagName,
    )

    if !extEntityKeyExists || extEntityKeyValue == "" {
      // ext entity key - no
      if entityTagExists {
        // entity key - yes, delete tag
        tagsToDelete = append(tagsToDelete, entityTagName)
        tagChanges += 1
      }
    } else if extEntityKeyExists {
      // ext entity key - yes
      if !entityTagExists {
        // entity key - no, add tag
        tagsToAdd = append(
          tagsToAdd,
          entities.TaggingTagInput{
            Key: entityTagName,
            Values: []string {extEntityKeyValue},
          },
        )
        tagChanges += 1
      } else if !stringSliceContains(entityTagValues, extEntityKeyValue){
        // entity key - yes, tag values contain ext entity value - no, replace
        for _, tag := range entity.Tags {
          if tag.Key == entityTagName {
            values := tag.Values
            values = append(values, extEntityKeyValue)
            tagsToUpdate = append(
              tagsToUpdate,
              entities.TaggingTagInput{
                Key: tag.Key,
                Values: values,
              },
            )
            continue
          }

          tagsToUpdate = append(
            tagsToUpdate,
            entities.TaggingTagInput(tag),
          )
        }
        tagChanges += 1
      }
    }
  }

  if tagChanges == 0 {
    return ENTITY_UPDATE_NONE, nil
  }

  errors := applyUpdates(
    i,
    entity,
    tagsToDelete,
    tagsToAdd,
    tagsToUpdate,
  )

  if len(errors) > 0 {
    return ENTITY_UPDATE_ERR, errors
  }

  return ENTITY_UPDATE_OK, errors
}

func applyUpdates(
  i                 *interop.Interop,
  entity            *EntityOutline,
  tagsToDelete      []string,
  tagsToAdd         []entities.TaggingTagInput,
  tagsToUpdate      []entities.TaggingTagInput,
) []error {
  nrClient := i.NrClient
  errors := []error{}

  if len(tagsToDelete) > 0 {
    taggingMutationResult, err := nrClient.Entities.TaggingDeleteTagFromEntity(
      entity.Guid,
      tagsToDelete,
    )
    if err != nil {
      errors = append(
        errors,
        fmt.Errorf(
          "deleting tags on entity %s (%s) failed: %v",
          entity.Name,
          entity.Guid,
          err,
        ),
      )
    } else if len(taggingMutationResult.Errors) > 0 {
      errors = append(
        errors,
        fmt.Errorf(
          "deleting tags on entity %s (%s) failed: %s",
          entity.Name,
          entity.Guid,
          buildTaggingMutationErrorMessage(taggingMutationResult.Errors),
        ),
      )
    }
  }

  if len(tagsToAdd) > 0 {
    taggingMutationResult, err := nrClient.Entities.TaggingAddTagsToEntity(
      entity.Guid,
      tagsToAdd,
    )
    if err != nil {
      errors = append(
        errors,
        fmt.Errorf(
          "adding tags on entity %s (%s) failed: %v",
          entity.Name,
          entity.Guid,
          err,
        ),
      )
    } else if len(taggingMutationResult.Errors) > 0 {
      errors = append(
        errors,
        fmt.Errorf(
          "adding tags on entity %s (%s) failed: %s",
          entity.Name,
          entity.Guid,
          buildTaggingMutationErrorMessage(taggingMutationResult.Errors),
        ),
      )
    }
  }

  if len(tagsToUpdate) > 0 {
    taggingMutationResult, err := nrClient.Entities.TaggingReplaceTagsOnEntity(
      entity.Guid,
      tagsToUpdate,
    )
    if err != nil {
      errors = append(
        errors,
        fmt.Errorf(
          "updating tags on entity %s (%s) failed: %v",
          entity.Name,
          entity.Guid,
          err,
        ),
      )
    } else if len(taggingMutationResult.Errors) > 0 {
      errors = append(
        errors,
        fmt.Errorf(
          "updating tags on entity %s (%s) failed: %s",
          entity.Name,
          entity.Guid,
          buildTaggingMutationErrorMessage(taggingMutationResult.Errors),
        ),
      )
    }
  }

  return errors
}

func buildTaggingMutationErrorMessage(
  errors []entities.TaggingMutationError,
) string {
  messages := []string{}
  for _, err := range errors {
    messages = append(messages, err.Message)
  }
  return strings.Join(messages, ";")
}

 func getEntityTagValues(
  i                 *interop.Interop,
  tags              []entities.EntityTag,
  tagName           string,
) ([]string, bool) {
  for _, tag := range tags {
    if tag.Key == tagName {
      return tag.Values, true
    }
  }

  return nil, false
}

func getEntityKeyValue(
  i                 *interop.Interop,
  entity            *EntityOutline,
  entityKey         string,
) (string, bool) {
  if strings.EqualFold(entityKey, "name") {
    return entity.Name, true
  } else if strings.EqualFold(entityKey, "guid") {
    return string(entity.Guid), true
  } else if strings.EqualFold(entityKey, "accountId") {
    return strconv.Itoa(entity.AccountID), true
  }

  if v, ok := getEntityTagValues(i, entity.Tags, entityKey); ok {
    return v[0], true
  }

  return "", false
}

func getExtEntityKeyValue(
  i *interop.Interop,
  extEntity *provider.Entity,
  keyName string,
) (string, bool) {
  return getNestedKeyValue(keyName, extEntity.Tags)
}
