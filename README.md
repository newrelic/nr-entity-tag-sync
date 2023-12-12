[![Community Project header](https://github.com/newrelic/open-source-office/raw/master/examples/categories/images/Community_Project.png)](https://github.com/newrelic/open-source-office/blob/master/examples/categories/index.md#category-community-project)

# New Relic Entity Tag Sync


![GitHub forks](https://img.shields.io/github/forks/newrelic/nr-entity-tag-sync?style=social)
![GitHub stars](https://img.shields.io/github/stars/newrelic/nr-entity-tag-sync?style=social)
![GitHub watchers](https://img.shields.io/github/watchers/newrelic/nr-entity-tag-sync?style=social)

![GitHub all releases](https://img.shields.io/github/downloads/newrelic/nr-entity-tag-sync/total)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/newrelic/nr-entity-tag-sync)
![GitHub last commit](https://img.shields.io/github/last-commit/newrelic/nr-entity-tag-sync)
![GitHub Release Date](https://img.shields.io/github/release-date/newrelic/nr-entity-tag-sync)

![GitHub issues](https://img.shields.io/github/issues/newrelic/nr-entity-tag-sync)
![GitHub issues closed](https://img.shields.io/github/issues-closed/newrelic/nr-entity-tag-sync)
![GitHub pull requests](https://img.shields.io/github/issues-pr/newrelic/nr-entity-tag-sync)
![GitHub pull requests closed](https://img.shields.io/github/issues-pr-closed/newrelic/nr-entity-tag-sync)

## Overview

The New Relic Entity Tag Sync application is a tool used to map entity metadata
from entities in an external system of record to tags on New Relic entities and
to keep such tag values synchronized with the external entity metadata values
across time.

### Concepts

There are several key concepts to be aware of in order to understand how the
entity tag sync application works.

#### New Relic Entities

A New Relic entity is anything that reports data to New Relic or that
contains data that we have access to. More information on New Relic entities
is available [in our documentation](https://docs.newrelic.com/docs/new-relic-solutions/new-relic-one/core-concepts/what-entity-new-relic/).

#### External Entities

An external entity can be generally thought of as any object with metadata in
an external system of record. The specific definition of an external entity is
particular to the external system.

Within the entity tag sync application, external entities are modeled as a set
of key-value pairs and a unique ID. It is these key-value pairs that are used
during the synchronization process to populate and update tags on the New Relic
entities.

#### Providers

A provider is a module capable of retrieving entity metadata from an external
system of record and mapping that data to the internal representation of an
external entity as a set of key-value pairs.

The following providers are supported:

* ServiceNow CMDB

##### ServiceNow CMDB provider

The ServiceNow CMDB provider models CMDB configuration items (CIs) as external
entities, enabling fields on CIs to be mapped to tags on New Relic entities.

The ServiceNow CMDB provider leverages the
[ServiceNow ReST API](https://docs.servicenow.com/bundle/utah-api-reference/page/integrate/inbound-rest/concept/c_RESTAPI.html)
to retrieve CI data and supports both HTTP Basic authentication and OAuth 2.0
authentication. See the [ReST API Security](https://docs.servicenow.com/bundle/utah-api-reference/page/integrate/inbound-rest/concept/c_RESTAPI.html#d773849e666)
documentation for more details on using these authentication methods with the
ServiceNow ReST API. See the [ServiceNow CMDB provider parameters section](#servicenow-cmdb-provider-parameters)
for details on how to configure the ServiceNow CMDB provider to use these
authentication methods.

#### Mappings

Mappings drive the actual synchronization process. Each mapping tells the entity
tag sync application what external entities to select from the provider, what
entities to select from New Relic, how to match external entities with New Relic
entities, and how key-value pairs from the external entity map to tags on the
New Relic entity.

Mappings are specified in the [mappings](#mapping-parameters) section of
[the configuration file](#configuration). During the synchronization process,
the entity tag sync application processes each mapping in order. A mapping is
processed as follows.

1. Fetch external entities

   All external entities matching the [external entity query criteria](#external-entity-query-criteria)
   are retrieved via the provider. The keys of the [`mapping`](#mapping) node as
   well as the value of the `extEntityKey` in the [`match`](#match-strategy)
   node are passed to the provider indicating the key-value pairs to retrieve
   for each external entity. The [last update timestamp](#delta-synchronization)
   is also passed to the provider if one was retrieved.

2. Fetch New Relic entities

   All New Relic entities matching the [New Relic entity query critera](#new-relic-entity-query-criteria)
   are retrieved via the New Relic GraphQL API

3. Find matching entities

   For each selected New Relic entity, the [match strategy](#match-strategy) is
   applied to find a matching external entity. If no match is found, processing
   proceeds to the next New Relic entity or, if all candidate New Relic entities
   have been processed, to the next mapping. If a match is found, processing
   continues to the synchronize step.

4. Synchronize tags

   If a match is discovered, the synchronization process executes the following
   logic for each pair of external entity key, **E**, to New Relic tag name,
   **T** in the [`mapping`](#mapping) node.

   - If **E** *does not* exist in the external entity metadata and **T**
     *does not* exist in the tags of the matching New Relic entity, do nothing
     and continue to the next pair.
   - If **E** *does not* exist in the external entity metadata and **T** *does*
     exist in the tags of the matching New Relic entity, delete the tag **T**
     from the New Relic entity.
   - If **E** *does* exist in the external entity metadata and **T** *does not*
     exist in the tags of the matching New Relic entity, add a tag to thew New
     Relic entity with **E** for the tag name and the value of key **E** in the
     external entity metadata as the singular tag value.
   - If **E** *does* exist in the external entity metadata and **T** *does*
     exist in the tags of the matching New Relic entity, scan the values for the
     tag **T** in the matching New Relic entity.
     - If the value of key **E** in the external entity metadata *is* in the
       values of tag **T**, do nothing and continue to the next pair.
     - If the value of key **E** in the external entity metadata *is not* in the
       values of tag **T**, the values of tag **T** are **replaced** with the
       value of key **E** in the external entity metadata.

   **NOTE:** Case 2 and case 4, subcase 2 above are destructive. In both cases,
   existing tag values removed. In general it should probably be assumed that
   the tags being synchronized are managed by the entity tag sync application.
   and should be modified via other means.

### Audit Events

The entity tag sync application is capable of producing audit events at various
points during the synchronization process. This feature is disabled by default
but can be enabled by setting the `events.enabled`
[general configuration parameter](#general-parameters). When enabled, the entity
tag sync application will produce events with the event type specified in the
`events.eventType` [general configuration parameter](#general-parameters) or the
event type `EntityTagSync` by default. The following attributes and values are
captured for _every_ event. Additional attributes are [action](#event-actions)
dependent.

| Name | Type | Description |
| --- | --- | --- |
| `id` | string | A canonical RFC-4122 UUID string that uniquely identifies each synchronization cycle |
| `action` | string | A string identifying the [action](#event-actions) that this event describes |
| `error` | bool | Flag indicating if an error occurred during this transaction or not |
| `errorMessage` | string | If an error occurred, a message describing what happened |

#### Event Actions

The following `action`s are produced along with any additional attributes
captured by each action.

**sync_start**

This action is produced at the start of each sync cycle and does not carry any
additional attributes. Note that the `error` attribute for this action will
always be `false` and the `errorMessage` attribute  will always be empty.

**sync_end**

This action is produced at the end of each sync cycle and does not carry any
additional attributes. The `error` attribute for this action will be set to
`true` if _any_ error occurred, including the case where the sync cycle finishes
successfully but a specific mapping has update errors. The `errorMessage`
attribute will be set providing more details.

**mapping_complete**

This action is produced each time during a sync cycle that the entity tag sync
application finishes processing a [mapping](#mappings). The set of attributes
captured for this action fall into one of three cases.

1. If an error occurred while processing the mapping, The `error` attribute for
   this action will be set to `true`, the `errorMessage` attribute will be set
   providing more details, and no additional attributes will be present.
1. If no external entities were returned by the [provider](#providers), not due
   to an error, the `error` attribute for this action will be `false`, the
   `errorMessage` attribute will be empty, and the `extEntityCount` attribute
   will be set to `0`, indicating that no external entities were returned and
   the subsequent mapping process was skipped since there was nothing to do.
1. If the overall mapping process completed, even if there were errors updating
   _some_ entities, the `error` attribute for this action will be `false`, the
   `errorMessage` attribute will be empty and the following attributes will be
   set.

    * `extEntityCount` - the number of external entities returned by the
      [provider](#providers)
    * `totalEntityCount` - the total number of New Relic entities that matched
      the [New Relic entity criteria](#new-relic-entity-query-criteria)
    * `totalEntitiesScanned` - the total number of New Relic entities that were
      tested against the external entities using [the match strategy](#match-strategy).
      This might be different than the `totalEntityCount` if an error occurred
      before all entities could be processed.
    * `totalEntitiesMatched` - the total number of New Relic entities that
      matched an external entity according to [the match strategy](#match-strategy)
    * `totalEntitiesNoMatch` - the total number of New Relic entities that did
      not match an external entity according to [the match strategy](#match-strategy)
    * `totalEntitiesSkipped` - the total number of New Relic entities that
      matched an external entity according to [the match strategy](#match-strategy)
      but were up-to-date with the external entity and did not require updates
    * `totalEntitiesUpdated` - the total number of New Relic entities that
      matched an external entity according to [the match strategy](#match-strategy)
      and were updated successfully
    * `totalEntitiesWithErrors` - the total number of New Relic entities that
      matched an external entity according to [the match strategy](#match-strategy)
      but were not updated successfully due to errors

### Delta Synchronization

By default, the synchronization cycle is stateless. As a result, unless
implemented directly by the provider, the provider has no way to determine when
the last synchronization was run which will likely cause the provider to
retrieve all external entities during each synchronization cycle. This may be
expensive in terms of consumed resources, especially as the data set of external
entities and/or the number of key-value pairs being retrieved increase.

To address this issue, the entity tag sync application can be configured to
"maintain" the timestamp of the last synchronization and pass the timestamp to
the provider when retrieving external entities. To enable this feature, the
`provider.useLastUpdate` flag must be set to `true` and [audit events](#audit-events)
must be enabled.

When enabled, the entity tag sync application will query NRDB for the latest
occurence of the [audit event](#audit-events) with the event type specified in
the `events.eventName` configuration parameter for which the value of the
`action` attribute is set to `sync_end`. The resulting timestamp is passed to
the provider implementation as the third parameter of the
[`GetEntities`](https://github.com/newrelic/nr-entity-tag-sync/blob/main/internal/provider/provider.go#L17)
function.

Provider implementations are not required to support this feature but providers
that do support it must honor it when it is passed.

**NOTE:** This functionality is currently implemented in a fairly primitive way.
The timestamp of the last synchronization is determined by querying NRDB for the
latest timestamp of the most recent audit event. This is why
[audit events](#audit-events)must be enabled in order to leverage this feature.

## Installation

The New Relic Entity Tag Sync application can be run as a standalone application
or as an AWS Lambda function.

### Standalone Installation

1. Download [the latest release](https://github.com/newrelic/nr-entity-tag-sync/releases)
   for your platform
2. Extract the archive to a new directory
3. Make a copy of config.yml from [the sample configuration file](configs/config.sample.yml) and place it 
   inside 'configs' folder created at same folder location as the CMDB utility executable.
4. Set the appropriate environment variables. Environment variable 'NEW_RELIC_LICENSE_KEY' is mandatory.
5. Execute the application

### AWS Lambda Installation

TODO

## Usage

### Configuration

The Entity Tag Sync application is driven by a YAML configuration file. The
configuration file consists of a set of [general parameters](#general-parameters),
a set of [provider parameters](#provider-parameters), and an array of
[mappings](#mappings). [A sample configuration file](configs/config.sample.yml)
is provided that shows an example of all parameters.

#### General parameters

The following general configuration parameters are supported. Some parameters
can be specified as environment variables as indicated below. Parameters listed
below with dots in their names correspond to nested YAML structures. For example
`log.level` corresponds to the following YAML.

```yaml
log:
  level: warn
```

| Name | Environment Variable | Description | Required | Example | Default |
| --- | --- | --- | --- | --- | --- |
| `apiKey` | `NEW_RELIC_API_KEY` | A New Relic User API key | Y | `NRAK-123456` | |
| `licenseKey` | `NEW_RELIC_LICENSE_KEY` | A New Relic Ingest License Key used for the Event API | Y if events enabled | `123456NRAL` | |
| `region` | `NEW_RELIC_REGION` | The New Relic datacenter to access (`US` or `EU`) | N | `US` | `US` |
| `log.level` | | The application log level | N | `debug` | `warn` |
| `log.fileName` | | Log file name | N | `app.log` | Standard output |
| `events.enabled` | | Flag to enable [audit event](#audit-events) | N | `true` | `false` |
| `events.accountId` | `NEW_RELIC_ACCOUNT_ID` | New Relic account where [audit events](#audit-events) are posted | Y if events enabled | `12345` | |
| `events.eventName` | | Name of [audit event](#audit-events) type | N | `MyCustomTagSyncEvent` | `EntityTagSync` |

**NOTE:** The `licenseKey` parameter in the configuration file can *not* be used
for configuring the Go APM agent that is used to instrument the app. The Go APM
agent bootstraps before the configuration is read and therefore the
`NEW_RELIC_LICENSE_KEY` environment variable must be used for this purpose.

#### Provider parameters

The `provider` section of the configuration file is used to specify the
parameters for the external entity provider. This section contains common
parameters that are provider indepent and parameters that are specific to the
selected provider. The following common parameters are supported.

| Name | Description | Required | Example | Default |
| --- | --- | --- | --- | --- |
| `type` | The provider implementation to use | Y | `servicenow` | |
| `useLastUpdate` | Flag to enable [delta synchronization](#delta-synchronization) | N | `true` | `false` |

The following values for the `type` parameter are supported.

* [`servicenow`](#servicenow-cmdb-provider-parameters)

##### ServiceNow CMDB provider parameters

The ServiceNow CMDB provider supports the following configuration parmaeters.

| Name | Environment Variable | Description | Required | Example | Default |
| --- | --- | --- | --- | --- | --- |
| `apiUrl` | `NR_CMDB_SNOW_APIURL` | The ServiceNow ReST API URL | Y | https://my-service-now.service-now.com | |
| `authType` | `NR_CMDB_SNOW_AUTHTYPE` | The type of authentication to use to authenticate with the ServiceNow instance (`basic` or `oauth`) | Y | `basic` | `basic` |
| `apiUser` | `NR_CMDB_SNOW_APIUSER` | The ServiceNow username to use when using `basic` authentication | Y if `authType` is `basic` | `admin` | |
| `apiPassword` | `NR_CMDB_SNOW_APIPASSWORD` | The password to use for the specified ServiceNow username when using `basic` authentication | Y if `authType` is `basic` | `abcd123` | |
| `oauthTokenUrl` | `NR_CMDB_SNOW_OAUTHTOKENURL` | The token URL to use when using `oauth` authentication | N | `https://myco.apis.com/auth` | `${apiUrl}/oauth_token.do` |
| `oauthGrantType` | `NR_CMDB_SNOW_GRANTTYPE` | The grant type to use when using `oauth` authentication | N | `client_credentials` | `password` |
| `oauthClientId` | `NR_CMDB_SNOW_OAUTHCLIENTID` | The client ID to use when using `oauth` authentication | Y if `authType` is `oauth` | `12345` | |
| `oauthClientSecret` | `NR_CMDB_SNOW_OAUTHCLIENTKEY` | The client secret to use when using `oauth` authentication | Y if `authType` is `oauth` | `12345` | |
| `oauthClientScopes` | `NR_CMDB_SNOW_OAUTHCLIENTKEY` | The list of OAuth scopes to request when using `oauth` authentication | N | `read_profile` | |
| `pageSize` | `NR_CMDB_SNOW_PAGESIZE` | A New Relic User API key | N | `10` | `10000` |

#### Mapping parameters

The `mappings` section of the configuration file is used to specify one or more
mapping configurations. Each mapping configuration specifies a set of
configuration parameters that defines the set of criteria for selecting external
entities, the set of criteria for selecting New Relic entities, the criteria
used to match external entities to New Relic entities, and the mapping from
external entity key-values to New Relic entity tags.

##### External entity query criteria

The `extEntityQuery` section of a mapping configuration specifies the query
criteria for selecting a set of external entities. The configuration parameters
in this section are specific to the provider.

###### ServiceNow CMDB entity query criteria

The ServiceNow CMDB provider supports the following configuration parameters for
selecting the set of CIs that are candidates for matching against New Relic
entities.

| Name | Description | Required | Example | Default |
| --- | --- | --- | --- | --- |
| `type` | The ServiceNow CMDB configuration item type name | Y | `cmdb_ci_email_server` | |
| `query` | An [encoded query string](https://docs.servicenow.com/csh?topicname=c_EncodedQueryStrings&version=utah&pubname=utah-platform-user-interface) to use to filter the result using  the `sysparm_query` parameter | N | `sys_updated_on>javascript:gs.dateGenerate('{{ .lastUpdateDate }}','{{ .lastUpdateTime }}')^operational_status!=2` | |
| `extraQueryParms` | Additional Service Now [query params](https://docs.servicenow.com/bundle/vancouver-api-reference/page/integrate/inbound-rest/concept/c_TableAPI.html) other than `sysparm_query` parameter | N | `&sysparm_display_value=true&sysparm_exclude_reference_link=true` | |
| `serverTimezone` | A [location name](https://pkg.go.dev/time#LoadLocation) corresponding to a file in the IANA Time Zone database for the time zone of the local ServiceNow instance | N | `America/New_York` | |

The ServiceNow CMDB query is executed by querying the `table` API using a URL
like the following.

`https://my-service-now.service-now.com/my/api/now/table/CI_TYPE?sysparm_fields=sys_id,FIELD1,...,FIELDN&sysparm_limit=PAGESIZE&sysparm_offset=0`

The value of `CI_TYPE` is the value of the `type` configuration parameter. The
values of `FIELD1,...,FIELDN` are the keys of the `mapping` node as well as the
value of the `extEntityKey` in the `match` node. The value of `PAGESIZE` is the
value of the `pageSize` parameter of the `provider` node.

**Query parameter**

If a `query` is specified in the entity query criteria, the `sysparm_query`
query parameter will be added to the query portion of the URL. The `query`
parameter value will be _automatically_ URL encoded so it should not be
specified in URL encoded format. For example to filter records where the
`active` field is `true` and the `roles` field is `itil`, specify the string
`active=true^roles=itil` and not `active%3Dtrue%5Eroles%3Ditil`.

When using [delta synchronization](#delta-synchronization), the special
character sequences `${lastUpdateDate}` and `${lastUpdateTime}` will be replaced
with the date and time strings, respectively, for the date and time specified by
the last synchronization timestamp. The date and time strings will be in the
format required by the [dateGenerate](https://docs.servicenow.com/bundle/utah-api-reference/page/app-store/dev_portal/API_reference/glideSystemScoped/concept/c_GlideSystemScopedAPI.html#title_r_SGSYS-dateGenerate_S_S)
function and will be converted to strings using the timezone specified in the
`serverTimezone` parameter.

**Example**

Consider the following YAML.

```yaml
...
provider:
  type: servicenow
  pageSize: 50
mappings:
- extEntityQuery:
    type: cmdb_ci_email_server
    query: 'sys_updated_on>javascript:gs.dateGenerate('${lastUpdateDate}','${lastUpdateTime}')^operational_status!=2'
    serverTimezone: America/Los_Angeles
  entityQuery:
    ...
  match:
    extEntityKey: name
    ...
  mapping:
    sys_class_name: foo
    environment: bar
```

Given this YAML, the ServiceNow CMDB provider will access the URL
`https://my-service-now.service-now.com/my/api/now/table/cmdb_ci_email_server?sysparm_fields=sys_id,name,sys_class_name,environment&sysparm_limit=50&sysparm_offset=0&sysparm_query=sys_updated_on%3Ejavascript%3Ags.dateGenerate%28%272023-06-01%27%2C%2712%3A00%3A00%27%29%5Eoperational_status%21%3D2`
to retrieve the fields `sys_id`, `name`, `sys_class_name`, and `environment` for
the CI records of type `cmdb_ci_email_server` that were updated on or after June
1st, 2023 at 12:00:00 GMT-7 and do not have an `operational_status` of `2`.

##### New Relic entity query criteria

The `entityQuery` section of a mapping configuration specifies the query
criteria for selecting a set of entities in New Relic.

| Name | Description | Required | Example | Default |
| --- | --- | --- | --- | --- |
| `type` | The entity type (`APPLICATION`, `HOST`, etc) | N | `WORKLOAD` | |
| `domain` | The entity domain (`APM`, `BROWSER`, etc) | N | `INFRA` | |
| `name` | The entity name | N | `Billing Service` | |
| `accountId` | The New Relic account ID | N | 123456 | |
| `tags` | A set of tag key + values pairs | N | (see below) | |
| `query` | A raw `entitySearch` query | N | `type IN ('APPLICATION')` | |

The `tags` value is an array of key and values pairs as in the following
example.

```yaml
mappings:
- entityQuery:
  ...
  tags:
    foo:
    - bar
    - baz
  ...
```

Note that the value part of the pair is an _array_ since New Relic tags can have
multiple values per key.

The query is executed via the Nerdgraph GraphQL API using the `entitySearch`
type of the `actor` type as follows.

- If a value is specified for the `query` key, it will take precedence over the
  other values.
- Otherwise, the `entitySearch` query value will be built by `AND`'ing the other
  values together. The following examples show the query string that would be
  produced for a given `entityQuery` configuration.

**Example 1: YAML**

```yaml
mappings:
- entityQuery:
  type: 'HOST'
  domain: 'INFRA'
  name: 'myinstance'
```

**Example 1: Query**

`type IN ('HOST') AND domain IN ('INFRA') AND name LIKE 'myinstance'`

**Example 2: YAML**

```yaml
mappings:
- entityQuery:
  accountId: 12345
  tags:
    foo:
    - bar
    beep:
    - boop
```

**Example 2: Query**

`tags.accountId = 12345 AND tags.foo IN ('bar') AND tags.beep IN ('boop')`

**Example 3: YAML**

```yaml
mappings:
- entityQuery:
  type: 'HOST'
  domain: 'INFRA'
  query: "name LIKE 'abc123'"
```

**Example 3: Query**

`name LIKE 'abc123'`

##### Match strategy

The `match` section of a mapping configuration is a tuplet that specifies the
name of an external entity key, the name of a New Relic entity attribute/tag,
and an operator used to compare the values of the external entity key to
the New Relic entity attribute/tag.

| Name | Description | Required | Example | Default |
| --- | --- | --- | --- | --- |
| `extEntityKey` | The external entity key to use for comparison | Y | `environment` | |
| `operator` | The type of comparison to use | Y | `equal` | |
| `entityKey` | The New Relic entity attribute/tag to use for comparison  | Y | `equal` | |

The `extEntityKey` external entity key will be implicitly added to the list of
keys requested from the provider for each external entity , regardless of whether
the key is referenced in the [mapping](#mapping) section.

The New Relic entity attribute/tag may be any one of the following.

* `name` - the New Relic entity name
* `guid` - the New Relic entity GUID
* `accountId` - the New Relic account ID that the entity belongs to
* Any tag name

The following values for the `operator` attribute are supported.

| Value | Meaning |
| --- | --- |
| `equal` | external entity key value is case-sensitive equivalent to New Relic entity attribute/tag value |
| `equal-ignore-case` | external entity key value is case-insensitive equivalent to New Relic entity attribute/tag value |
| `contains` | New Relic entity attribute/tag value is a case-sensitive sub-string within external entity key value |
| `contains-ignore-case` | New Relic entity attribute/tag value is a case-insensitive sub-string within external entity key value |
| `inverse-contains-ignore-case` | external entity key value is a case-insensitive sub-string within New Relic entity attribute/tag value |

For example, consider the following YAML.

```yaml
match:
  extEntityKey: foo
  operator: equal-ignore-case
  entityKey: bar
```

Given this YAML, New Relic entities will be matched against external entities by
comparing the values of the tag `bar` on the New Relic entities to the values of
the key `foo` on the external entities, case insensitively.

##### Mapping

The `mapping` node of a mapping configuration specifies the mapping from
external entity keys to New Relic entity tags. The keys of the `mapping` node
represent the external entity keys while the values represent the New Relic
entity tag names.

For a given external entity to New Relic entity pair, the value of each of the
external entity keys specified by the keys in the `mapping` node are used as the
values of the New Relic entity tags specified by the values in the `mapping`
node.

For example, consider the following YAML.

```yaml
mapping:
  foo: bar
  beep: boop
```

Given this YAML, the values of the external entity key `foo` and `beep` of any
external entity that matches a New Relic entity will be set as the values of the
tags `bar` and `boop` on the matching New Relic entity.

#### Full example

This section provides an example configuration and set of entities followed by
a full walk-through of the synchronization process.

##### Sample configuration

```yaml
apiUrl: https://api.newrelic.com
log:
  level: warn
provider:
  type: servicenow
  apiUrl: https://my-service-now.service-now.com
  apiUser: admin
mappings:
- extEntityQuery:
    type: cmdb_ci_email_server
  entityQuery:
    type:
    - APPLICATION
    domain:
    - APM
    accountId: 1
  match:
    extEntityKey: name
    operator: equal-ignore-case
    entityKey: name
  mapping:
    sys_class_name: SNOW_CI_CLASS
    sys_id: SNOW_CMDB_CI
    environment: SNOW_ENVIRONMENT
    sys_domain.value: SNOW_SYS_DOMAIN
- extEntityQuery:
    type: cmdb_ci_app_server
  entityQuery:
    type:
    - HOST
    domain:
    - INFRA
  match:
    extEntityKey: name
    operator: contains
    entityKey: ciMatch
  mapping:
    sys_class_name: SNOW_CI_CLASS
    sys_id: SNOW_CMDB_CI
    environment: SNOW_ENVIRONMENT
    sys_domain.value: SNOW_SYS_DOMAIN
```

##### ServiceNow CIs

| Type | Name | sys_id | sys_class_name | sys_domain.value | environment |
| --- | --- | --- | --- | --- | --- |
| cmdb_ci_email_server | Microsoft Exchange | abcd123 | cmdb_ci_email_server | global | Production |
| cmdb_ci_app_server | WebSphere Application Server | efgh456 | cmdb_ci_app_server | local | Development |

##### New Relic Entities

| Type | Domain | Name | GUID | tags.ciMatch | tags.SNOW_CI_CLASS | tags.SNOW_CMDB_CI | tags.SNOW_SYS_DOMAIN | tags.SNOW_ENVIRONMENT |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| APPLICATION | APM | microsoft exchange | NR12345 | | cmdb_ci_email_server | abcd123 |  | Development |
| HOST | INFRA | WebSphere Application Server | NR45678 | WebSphere | cmdb_ci_app_server | efgh456 | global |  |

##### Walkthrough

1. The entity tag sync application starts up and reads in the configuration.
1. A new ServiceNow provider is created with the API base URL
  `https://my-service-now.service-now.com` and the API username `admin`. The API
   password will be read from the environment variable `NR_CMDB_SNOW_APIPASSWORD`.
   The page size defaults to 10000.
1. The application starts processing the first mapping configuration by
   inspecting the `extEntityQuery` node.
1. Using the `type` parameter specified in the `extEntityQuery` node as the
   table name to query and using the keys of the `mapping` node as well as the
   value of the `extEntityKey` of the `match` node as the CI fields to return,
   the ServiceNow CMDB provider will make an HTTP `GET` request for the
   following URL.

   `https://my-service-now.service-now.com/my/api/now/table/cmdb_ci_email_server?sysparm_fields=sys_id,name,sys_class_name,environment,sys_domain&sysparm_limit=10000&sysparm_offset=0`

   This API call will return the requested key-value pairs for all CIs of type
   `cmdb_ci_email_server`, including the CI listed above with `sys_id`
   `abcd123`.
1. The application next uses the `entityQuery` node to request to make the
   following GraphQL query against the New Relic GraphQL API.

   ```graphql
   {
     actor {
       entitySearch(query: "type IN ('APPLICATION') AND domain IN ('APM') AND tags.accountId` = 1") {
         count
         results {
           entities {
             guid
             name
             accountId
             domain
             type
             alertSeverity
             permalink
             reporting
             tags {
               key
               values
             }
           }
         }
       }
     }
   }
   ```

   This query will return all APM service entities in account ID 1, including
   the APM application listed above with GUID `NR12345`.
1. Next, the application will iterate over the entities returned from the
   GraphQL call and, per the `match` node values, will perform a
   case-insensitive equals check between each CI name and the current entity
   name. In this case, the application will find a match between the CI with
   `sys_id` `abcd123` and the New Relic entity with GUID `NR12345` because the
   values `Microsoft Exchange` and `microsoft exchange` are case-insensitively
   equivalent.
1. The application will update the tags on entity `NR12345` using the values of
   the key-value pairs from CI `abcd123` as follows.

   | CI Key | CI Value | Entity Tag | Entity `NR12345` before | Entity `NR12345` after|
   | --- | --- | --- | --- | --- |
   | sys_class_name | cmdb_ci_email_server | SNOW_CI_CLASS | cmdb_ci_email_server | cmdb_ci_email_server |
   | sys_id | abcd123 | SNOW_CMDB_CI | abcd123 | abcd123 |
   | sys_domain.value | global | SNOW_SYS_DOMAIN | | global |
   | environment | Production | SNOW_ENVIRONMENT | Development | Production |
1. The application starts processing the second mapping configuration by
   inspecting the `extEntityQuery` node.
1. Using the specified `type` parameter and the keys of the `mapping` node as
   well as the value of the `extEntityKey` of the `match` node specified by the
   configuration, the ServiceNow CMDB provider will make an HTTP `GET` request
   for the following URL.

   `https://my-service-now.service-now.com/my/api/now/table/cmdb_ci_app_server?sysparm_fields=sys_id,name,sys_class_name,environment,sys_domain&sysparm_limit=10000&sysparm_offset=0`

   This API call will return the requested key-value pairs for all CIs of type
   `cmdb_ci_app_server`, including the CI listed above with `sys_id`
   `efgh456`.
1. The application next uses the `entityQuery` node to request to make the
   following GraphQL query against the New Relic GraphQL API.

   ```graphql
   {
     actor {
       entitySearch(query: "type IN ('HOST') AND domain IN ('INFRA')") {
         count
         results {
           entities {
             guid
             name
             accountId
             domain
             type
             alertSeverity
             permalink
             reporting
             tags {
               key
               values
             }
           }
         }
       }
     }
   }
   ```
   This query will return all infrastructure host entities accessible by the
   given license key, including the infrastructure host listed above with GUID
   `NR45678`.
1. Next, the application will iterate over the entities returned from the
   GraphQL call and, per the `match` node values, will perform a
   case-insensitive "contains" check between each CI name and the _first_ value
   of the `ciMatch` tag of each entity. In this case, the application will find
   a match between the CI with `sys_id` `efgh456` and the New Relic entity with
   GUID `NR45678` because the value `WebSphere Application Server` includes
   the value `WebSphere`.
1. The application will update the tags on entity `NR45678` using the values of
   the key-value pairs from CI `efgh456` as follows.

   | CI Key | CI Value | Entity Tag | Entity `NR45678` before | Entity `NR45678` after|
   | --- | --- | --- | --- | --- |
   | sys_class_name | cmdb_ci_app_server | SNOW_CI_CLASS | cmdb_ci_email_server | cmdb_ci_email_server |
   | sys_id | efgh456 | SNOW_CMDB_CI | efgh456 | efgh456 |
   | sys_domain.value | local | SNOW_SYS_DOMAIN | global | local |
   | environment | Production | SNOW_ENVIRONMENT | | Development |

## Building

### Coding Conventions

#### Style Guidelines

While not strictly enforced, the basic preferred editor settings are set in the
[.editorconfig](./.editorconfig). Other than this, no style guidelines are
currently imposed.

#### Static Analysis

This project uses both [`go vet`](https://pkg.go.dev/cmd/vet) and
[`staticcheck`](https://staticcheck.io/) to perform static code analysis. These
checks are run via [`precommit`](https://pre-commit.com) on all commits. Though
this can be bypassed on local commit, both tasks are also run during
[the `validate` workflow](./.github/workflows/validate.yml) and must have no
errors in order to be merged.

#### Commit Messages

Commit messages must follow [the conventional commit format](https://www.conventionalcommits.org/en/v1.0.0/).
Again, while this can be bypassed on local commit, it is strictly enforced in
[the `validate` workflow](./.github/workflows/validate.yml).

The basic commit message structure is as follows.

```
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

In addition to providing consistency, the commit message is used by
[svu](https://github.com/caarlos0/svu) during
[the release workflow](./.github/workflows/release.yml). The presence and values
of certain elements within the commit message affect auto-versioning. For
example, the `feat` type will bump the minor version. Therefore, it is important
to use the guidelines below and carefully consider the content of the commit
message.

Please use one of the types below.

- `feat` (bumps minor version)
- `fix` (bumps patch version)
- `chore`
- `build`
- `docs`
- `test`

Please use one of the scopes below.

- `sync` - work related to the synchronization process
- `config` - work related to the config system
- `provider[-optional-name]` - work related to the provider framework or a
  specific provider
- `lambda` - work related to the lambda code and/or deployment
- `ci` - work related to continuous integration (GitHub workflow/actions)
- `release` - work related to creating a new release

Any type/scope can be followed by the `!` character to indicate a breaking
change. Additionally, any commit that has the text `BREAKING CHANGE:` in the
footer will indicate a breaking change.

### Local Development

For local development, simply use `go build` and `go run`. For example,

```bash
go build cmd/nr-entity-tag-sync/nr-entity-tag-sync.go
```

Or

```bash
go run cmd/nr-entity-tag-sync/nr-entity-tag-sync.go
```

If you prefer, you can also use [`goreleaser`](https://goreleaser.com/) with
the `--single-target` option to build the binary for the local `GOOS` and
`GOARCH` only.

```bash
goreleaser build --single-target
```

### Releases

Releases are built and packaged using [`goreleaser`](https://goreleaser.com/).
By default, a new release will be built automatically on any push to the `main`
branch. For more details, review the [`.goreleaser.yaml`](./.goreleaser.yaml)
and [the `goreleaser` documentation](https://goreleaser.com/intro/).

The [svu](https://github.com/caarlos0/svu) utility is used to generate the next
tag value [based on commit messages](https://github.com/caarlos0/svu#commit-messages-vs-what-they-do).

### GitHub Workflows

This project utilizes GitHub workflows to perform actions in response to
certain GitHub events.

| Workflow | Events | Description
| --- | --- | --- |
| [validate](./.github/workflows/validate.yml) | `push` | Runs [precommit](https://pre-commit.com) to perform static analysis and runs [commitlint](https://commitlint.js.org/#/) to validate the last commit message |
| [build](./.github/workflows/build.yml) | `pull_request` | Builds and tests code |
| [release](./.github/workflows/release.yml) | `push` to `main` branch | Generates a new tag using [svu](https://github.com/caarlos0/svu) and runs [`goreleaser`](https://goreleaser.com/) |
| [repolinter](./.github/workflows/repolinter.yml) | `pull_request` | Enforces repository content guidelines |

## Testing

TBD

# Support

New Relic has open-sourced this project. This project is provided AS-IS WITHOUT
WARRANTY OR DEDICATED SUPPORT. Issues and contributions should be reported to
the project here on GitHub.

We encourage you to bring your experiences and questions to the
[Explorers Hub](https://discuss.newrelic.com/) where our community members
collaborate on solutions and new ideas.

## Privacy

At New Relic we take your privacy and the security of your information
seriously, and are committed to protecting your information. We must emphasize
the importance of not sharing personal data in public forums, and ask all users
to scrub logs and diagnostic information for sensitive information, whether
personal, proprietary, or otherwise.

We define “Personal Data” as any information relating to an identified or
identifiable individual, including, for example, your name, phone number, post
code or zip code, Device ID, IP address, and email address.

For more information, review [New Relic’s General Data Privacy Notice](https://newrelic.com/termsandconditions/privacy).

## Contribute

We encourage your contributions to improve this project! Keep in mind that
when you submit your pull request, you'll need to sign the CLA via the
click-through using CLA-Assistant. You only have to sign the CLA one time per
project.

If you have any questions, or to execute our corporate CLA (which is required
if your contribution is on behalf of a company), drop us an email at
opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed
to the privacy and security of our customers and their data. We believe that
providing coordinated disclosure by security researchers and engaging with the
security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of
New Relic's products or websites, we welcome and greatly appreciate you
reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project
would not be what it is today.

## License

The [New Relic Integration for Conviva] is licensed under the
[Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
