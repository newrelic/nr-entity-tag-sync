package sync

import (
	"fmt"
	"strings"

	"github.com/newrelic/newrelic-client-go/pkg/common"
	"github.com/newrelic/newrelic-client-go/pkg/entities"
	"github.com/newrelic/nr-entity-tag-sync/pkg/interop"
)

type entityProcessorResult int

const (
  ENTITY_NO_MATCH     entityProcessorResult = iota
  ENTITY_UPDATE_OK
  ENTITY_UPDATE_NONE
  ENTITY_UPDATE_ERR
)

type EntityOutline struct {
  Guid              common.EntityGUID
  Name              string
  AccountID         int
  EntityType        entities.EntityType
  Domain            string
  Type              string
  Tags              []entities.EntityTag
}

const (
  entitySearchResultQuery = `
    entities {
      guid
      name
      accountId
      entityType
      domain
      type
      tags {
        key
        values
      }
    }
    nextCursor
  `
  getEntitySearchByQuery = `query(
    $query: String,
  ) { actor { entitySearch(
    query: $query,
  ) {
    count
    results {` + entitySearchResultQuery + "} } } }"
  getEntitySearchByQueryWithCursor = `query(
    $query: String,
    $cursor: String,
  ) { actor { entitySearch(
    query: $query,
  ) {
    count
    results(
      cursor: $cursor,
    ) {` + entitySearchResultQuery + "} } } }"
)

type entitySearchResponse struct {
  Actor struct {
    EntitySearch struct {
      Count int
      Results struct {
        Entities []EntityOutline
        NextCursor string
      }
    }
  }
}

type entityProcessingResult struct {
  totalEntities             int
  totalEntitiesScanned      int
  totalEntitiesMatched      int
  totalEntitiesNoMatch      int
  totalEntitiesUpdated      int
  totalEntitiesWithErrors   int
  totalEntitiesSkipped      int
}

type entityProcessorFn func (
  i                 *interop.Interop,
  mapping           *MappingConfig,
  entity            *EntityOutline,
) (entityProcessorResult, []error)

func processEntities(
  i *interop.Interop,
  mapping *MappingConfig,
  entityProcessor entityProcessorFn,
) (*entityProcessingResult, error) {
  processingResult := &entityProcessingResult{}
  query := buildQuery(&mapping.EntityQuery)
  nextCursor := ""

  i.Logger.Debugf("fetching New Relic entities for query: \"%s\"", query)

  for done := false; !done; {
    resp, err := getEntities(i, query, nextCursor)
    if err != nil {
      return processingResult,
        fmt.Errorf("graphql error fetching entities: %s", err)
    }

    entitySearch := resp.Actor.EntitySearch
    processingResult.totalEntities = entitySearch.Count

    i.Logger.Tracef(
      "scanning %d New Relic entities",
      len(entitySearch.Results.Entities),
    )

    for _, entityOutline := range entitySearch.Results.Entities {
      i.Logger.Tracef(
        "processing New Relic entity %s (%s)",
        entityOutline.Name,
        entityOutline.Guid,
      )

      result, errors := entityProcessor(
        i,
        mapping,
        &entityOutline,
      )

      i.Logger.Tracef(
        "result of processing entity %s (%s): %d",
        entityOutline.Name,
        entityOutline.Guid,
        result,
      )

      if result == ENTITY_NO_MATCH {
        processingResult.totalEntitiesNoMatch += 1
      } else {
        processingResult.totalEntitiesMatched += 1
      }

      if result == ENTITY_UPDATE_ERR {
        processingResult.totalEntitiesWithErrors += 1

        i.Logger.Warnf(
          "errors while updating entity %s (%s): see below for errors",
          entityOutline.Name,
          entityOutline.Guid,
        )

        for _, err := range errors {
          i.Logger.Warnf("error while updating entity: %s", err)
        }
      } else if result == ENTITY_UPDATE_NONE {
        processingResult.totalEntitiesSkipped += 1
      } else if result == ENTITY_UPDATE_OK {
        processingResult.totalEntitiesUpdated += 1
      }
    }

    processingResult.totalEntitiesScanned += len(entitySearch.Results.Entities)
    nextCursor = entitySearch.Results.NextCursor

    if nextCursor == "" {
      done = true
    }
  }

  return processingResult, nil
}

func buildQuery(entityQuery *EntityQuery) string {
  if entityQuery.Query != "" {
    return entityQuery.Query
  }

  var parts []string

  if len(entityQuery.Domain) > 0 {
    parts = append(
      parts,
      fmt.Sprintf("domain IN ('%s')", strings.Join(entityQuery.Domain, "','")),
    )
  }

  if len(entityQuery.Type) > 0 {
    parts = append(
      parts,
      fmt.Sprintf("type IN ('%s')", strings.Join(entityQuery.Type, "','")),
    )
  }

  if entityQuery.Name != "" {
    parts = append(
      parts,
      fmt.Sprintf("name LIKE '%s'", entityQuery.Name),
    )
  }

  if entityQuery.AccountId != 0 {
    parts = append(
      parts,
      fmt.Sprintf("tags.`accountId` = %d", entityQuery.AccountId),
    )
  }

  if len(entityQuery.Tags) > 0 {
    var tags []string

    for _, tag := range entityQuery.Tags {
      tags = append(
        tags,
        fmt.Sprintf(
          "tags.`%s` IN ('%s')",
          tag.Key,
          strings.Join(tag.Values, "','"),
        ),
      )
    }

    parts = append(
      parts,
      strings.Join(tags, " AND "),
    )
  }

  return strings.Join(parts, " AND ")
}

func getEntities(
  i                 *interop.Interop,
  query             string,
  cursor            string,
) (*entitySearchResponse, error) {
  var resp entitySearchResponse

  vars := map[string]interface{}{
    "query":   query,
  }

  if cursor != "" {
    vars["cursor"] = cursor

    i.Logger.Tracef("running query using cursor: %s", cursor)

    if err := i.NrClient.NerdGraph.QueryWithResponse(
      getEntitySearchByQueryWithCursor,
      vars,
      &resp,
    ); err != nil {
      return nil, err
    }

    return &resp, nil
  }

  if err := i.NrClient.NerdGraph.QueryWithResponse(
    getEntitySearchByQuery,
    vars,
    &resp,
  ); err != nil {
    return nil, err
  }

  return &resp, nil
}
