package sync

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

type Match struct {
  ExtEntityKey      string
  Operator          string
  EntityKey         string
}

type Mapping map[string]string

type MappingConfig struct {
  ExtEntityQuery    map[string]interface{}
  EntityQuery       EntityQuery
  Match             Match
  Mapping           Mapping
}

type Mappings []MappingConfig
