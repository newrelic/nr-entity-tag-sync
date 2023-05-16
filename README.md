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

#### Mappings

A mapping is a set of configuration parameters that defines the set of criteria
for selecting external entities, the set of criteria for selecting New Relic
entities, the criteria used to match external entities to New Relic entities,
and the mapping from external entity key-values to New Relic entity tags.

### How It Works

At a high level, the synchronization process works as follows.

- On startup, the entity tag sync application reads in a configuration. The
  configuration is composed of 3 main parts.
  1. General configuration parameters
  2. Provider configuration parameters
  3. A set of entity mappings
- For each entity mapping, the following steps are performed.
  1. All external entities matching the external entity criteria are retrieved
     via the provider. The keys of the `mapping` node as well as the value of
     the `extEntityKey` in the `match` node are passed to the provider
     indicating the key-value pairs to retrieve for each external entity.
  2. All New Relic entities matching the New Relic entity critera are retrieved
     via the New Relic GraphQL API
  3. For each New Relic entity, the match criteria is used to resolve the New
     Relic entity with a matching external entity
  4. If a match is discovered, the key-value pairs retrieved for the matching
     external entity are used to populate/update the tags of the matching New
     Relic entity

## Installation

The New Relic Entity Tag Sync application can be run as a standalone application
or as an AWS Lambda function.

### Standalone Installation

1. Download [the latest release](https://github.com/newrelic/nr-entity-tag-sync/releases)
   for your platform
2. Extract the archive to a new directory
3. Create a new configuration file from
   [the sample configuration file](configs/config.sample.yml)
4. Set the appropriate environment variables
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
| `apiUrl` | `NEW_RELIC_API_URL` | The New Relic GraphQL API URL | Y | https://api.newrelic.com | https://api.newrelic.com |
| `apiKey` | `NEW_RELIC_API_KEY` | A New Relic User API key | Y | `NRAK-123456` | |
| `log.level` | | The application log level | N | `debug` | `warn` |
| `log.fileName` | | Log file name | N | `app.log` | Standard output |

#### Provider parameters

The `provider` section of the configuration file is used to specify the
parameters for the external entity provider. The `type` parameter specifies the
provider implementation to use and is the only common parameter in the
`provider` section. The following values for the `type` parameter are supported.

* `servicenow` (default)

The remaining parameters in this section are particular to the provider
implementation.

##### ServiceNow CMDB provider parameters

The ServiceNow CMDB provider supports the following configuration parmaeters.

| Name | Environment Variable | Description | Required | Example | Default |
| --- | --- | --- | --- | --- | --- |
| `apiUrl` | `NR_CMDB_SNOW_APIURL` | The ServiceNow ReST API URL | Y | https://my-service-now.service-now.com | |
| `apiUser` | `NR_CMDB_SNOW_APIUSER` | A New Relic User API key | Y | `admin` | |
| `apiPassword` | `NR_CMDB_SNOW_APIPASSWORD` | A New Relic User API key | Y | `abcd123` | |
| `pageSize` | `NR_CMDB_SNOW_PAGESIZE` | A New Relic User API key | N | `10` | `10000` |

#### Mapping parameters

The `mappings` section of the configuration file is used to specify one or more
mapping configurations. Each mapping configuration specifies a set of
configuration parameters that defines the set of criteria for selecting external
entities, the set of criteria for selecting New Relic entities, the criteria
used to match external entities to New Relic entities, and the mapping from
external entity key-values to New Relic entity tags.

##### Provider entity query criteria

The `extEntityQuery` section of a mapping configuration specifies the query
criteria for selecting a set of external entities. The configuration parameters
in this section are specific to the provider.

###### ServiceNow CMDB provider entity query criteria

The ServiceNow CMDB provider supports the following configuration parameters for
selecting the set of CIs that are candidates for matching against New Relic
entities.

| Name | Description | Required | Example | Default |
| --- | --- | --- | --- | --- |
| `type` | The ServiceNow CMDB configuration item type name | Y | `cmdb_ci_email_server` | |

The ServiceNow CMDB query is executed by querying the `table` API using a URL
like the following.

`https://my-service-now.service-now.com/my/api/now/table/CI_TYPE?sysparm_fields=sys_id,FIELD1,...,FIELDN&sysparm_limit=PAGESIZE&sysparm_offset=0`

The value of `CI_TYPE` is the value of the `type` configuration parameter. The
values of `FIELD1,...,FIELDN` are the keys of the `mapping` node as well as the
value of the `extEntityKey` in the `match` node. The value of `PAGESIZE` is the
value of the `pageSize` parameter of the `provider` node.

For example, consider the following YAML.

```yaml
...
provider:
  type: servicenow
  pageSize: 50
mappings:
- extEntityQuery:
    type: cmdb_ci_email_server
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
`https://my-service-now.service-now.com/my/api/now/table/cmdb_ci_email_server?sysparm_fields=sys_id,name,sys_class_name,environment&sysparm_limit=50&sysparm_offset=0`
to retrieve the fields `sys_id`, `name`, `sys_class_name`, and `environment` for
each CI record of type `cmdb_ci_email_server`.

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

The specified external entity key will be added to the list of keys requested
from the provider for each external entity.

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

##### Tag mapping

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
