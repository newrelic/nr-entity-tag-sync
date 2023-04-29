package nerdgraph

import (
	"context"
	"fmt"
)

func (n *NerdgraphClient) AddTags(
  entity *EntityOutline,
  tags []Tag,
) error {
  var gql struct {
    TaggingAddTagsToEntity struct {
      Errors []struct {
        Message string
        Type string
      }
    } `graphql:"taggingAddTagsToEntity(guid: $guid, tags: $tags)"`
  }

  tagInput := []TaggingTagInput{}

  for _, tag := range tags {
    tagInput = append(tagInput, TaggingTagInput(tag))
  }

  variables := map[string]interface{} {
    "guid": EntityGuid(entity.Guid),
    "tags": tagInput,
  }

  err := n.Mutate(context.Background(), &gql, variables)
  if err != nil {
    return err
  }

  if len(gql.TaggingAddTagsToEntity.Errors) > 0 {
    return fmt.Errorf("error adding tags to entity: %v", gql.TaggingAddTagsToEntity.Errors)
  }

  return nil
}

func (n *NerdgraphClient) DeleteTags(
  entity *EntityOutline,
  tagNames []string,
) error {
  var gql struct {
    TaggingDeleteTagFromEntity struct {
      Errors []struct {
        Message string
        Type string
      }
    } `graphql:"taggingDeleteTagFromEntity(guid: $guid, tagKeys: $tags)"`
  }

  variables := map[string]interface{} {
    "guid": EntityGuid(entity.Guid),
    "tags": tagNames,
  }

  err := n.Mutate(context.Background(), &gql, variables)
  if err != nil {
    return err
  }

  if len(gql.TaggingDeleteTagFromEntity.Errors) > 0 {
    return fmt.Errorf("error adding tags to entity: %v", gql.TaggingDeleteTagFromEntity.Errors)
  }

  return nil
}
