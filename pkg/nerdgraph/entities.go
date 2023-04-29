package nerdgraph

import "context"

func (c *NerdgraphClient) GetEntities(entityQuery *EntityQuery) ([]EntityOutline, error) {
  var entities []EntityOutline

  nextCursor := ""

  for done := false; !done; {
    if nextCursor != "" {
      var gql struct {
        Actor struct {
          EntitySearch struct {
            Count int
            Results struct {
              Entities []EntityOutline
              NextCursor string
            } `graphql:"results(cursor: $c)"`
          } `graphql:"entitySearch(query: $q)"`
        }
      }

      variables := map[string]interface{} {
        "q": buildQuery(entityQuery),
        "c": nextCursor,
      }

      err := c.Query(context.Background(), &gql, variables)
      if err != nil {
        return nil, err
      }

      entities = append(entities, gql.Actor.EntitySearch.Results.Entities...)
      nextCursor = gql.Actor.EntitySearch.Results.NextCursor
      done = (nextCursor == "")

      continue
    }

    var gql struct {
      Actor struct {
        EntitySearch struct {
          Count int
          Results struct {
            Entities []EntityOutline
            NextCursor string
          }
        } `graphql:"entitySearch(query: $q)"`
      }
    }

    variables := map[string]interface{} {
      "q": buildQuery(entityQuery),
    }

    err := c.Query(context.Background(), &gql, variables)
    if err != nil {
      return nil, err
    }

    entities = append(entities, gql.Actor.EntitySearch.Results.Entities...)
    nextCursor = gql.Actor.EntitySearch.Results.NextCursor
    done = (nextCursor == "")
  }

  return entities, nil
}
