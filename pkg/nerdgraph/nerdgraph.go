package nerdgraph

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hasura/go-graphql-client"
	log "github.com/sirupsen/logrus"
)

type NerdgraphClient struct {
  ApiURL            string
  ApiKey            string
  Logger            *log.Logger
}

func NewNerdgraphClient(
  apiUrl            string,
  apiKey            string,
  log               *log.Logger,
) *NerdgraphClient {
  return &NerdgraphClient{apiUrl, apiKey, log}
}

type Tag struct {
  Key               string
  Values            []string
}

type EntityQuery struct {
  Type              []string
  Domain            []string
  Name              string
  AccountId         int
  Tags              []Tag
  Query             string
}

type EntityOutline struct {
  Guid              string
  Name              string
  AccountId         int
  Domain            string
  Type              string
  AlertSeverity     string
  Permalink         string
  Reporting         bool
  Tags              []Tag
}

type EntityGuid string

type TaggingTagInput struct {
  Key  string
  Values []string
}

func (c *NerdgraphClient) Query(
  context           context.Context,
  gql               interface{},
  variables         map[string]interface{},
) error {
  client := c.newClient()

  err := client.Query(context, gql, variables)
  if err != nil {
    return err
  }

  return nil
}

func (c *NerdgraphClient) Mutate(
  context context.Context,
  gql interface{},
  variables map[string]interface{},
) error {
  client := c.newClient()

  err := client.Mutate(context, gql, variables)
  if err != nil {
    return err
  }

  return nil
}

func (c *NerdgraphClient) newClient() *graphql.Client {
  url := fmt.Sprintf("%s/graphql", c.ApiURL)

  return graphql.NewClient(url, nil).WithRequestModifier(
    func(req *http.Request) {
      req.Method = "POST"
      req.Header.Add("Accept", "application/json")
      req.Header.Add("Content-Type", "application/json")
      req.Header.Add("Api-Key", c.ApiKey)
    },
  )
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
