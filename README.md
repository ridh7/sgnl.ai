# SGNL Adapter Template

The SGNL Adapter Template is the starting point for creating a new SGNL Adapter.

## Prerequisites

- A basic understanding of the Golang programming language.
- An understanding of the [gRPC](https://grpc.io/) framework and [protocol buffers](https://protobuf.dev/).
- Golang tools installed on your development machine, with an IDE suited for Golang development.

## Background Information

### Terminology

**Adapter** - A simple gRPC server that queries an external API and parses the response into a format suitable for the SGNL ingestion service. More information on adapters can be found in the [What is an adapter?](#what-is-an-adapter?) section.

**SGNL ingestion service** - One of SGNL's core microservices which is responsible for ingesting external data into SGNL's graph database.

**System of Record (SoR)** - An external system that provides data to be ingested into SGNL, typically via an API. For example, a CRM tool like Salesforce or HubSpot.

**Entity** - A SGNL term to represent a type of data within an SoR. An entity typically maps to a single API resource. For example, the IdentityNow Account entity is retrieved from the `/accounts` endpoint.

### What is an adapter?

An adapter is a gRPC server which has two main responsibilities:

1. Making requests to a System of Record (SoR) to retrieve data.
2. Transforming the response from the SoR into a format that can be consumed by the SGNL ingestion service.

An adapter is **stateless**. It simply acts as a proxy to send requests to SoRs and parse the responses. An adapter uses SGNL's [adapter-framework](https://github.com/SGNL-ai/adapter-framework) under the hood.

Requests to the adapter invoke the `GetPage` method.

### Adapter Authentication

An adapter authenticates incoming gRPC requests via the `token` metadata key. The value of this key must match one of the tokens in the `ADAPTER_TOKENS` file that you define. The `ADAPTER_TOKENS` file is a JSON array of strings, where each string is a token. Only one token is required, but multiple tokens can be defined in the event that a token needs to be rotated.

For example, an `ADAPTER_TOKENS` file may look like:

```
["<token1>", "<token2>", ...]
```

While an adapter does not validate the tokens defined, we recommend generating tokens with a length of at least 64 random bytes using a cryptographically secure pseudo random number generator (CSPRNG). For example, `openssl rand 64 | openssl enc -base64 -A`.

Once this file is created, set the `AUTH_TOKENS_PATH` environment variable to the path of the `ADAPTER_TOKENS` file. More information on starting an adapter is discussed below in the [Getting Started](#1-getting-started) section.

## Writing an Adapter

### 1. Getting Started

1. Clone this repository.

1. Update the names of `github.com/sgnl-ai/adapter-template/*` Golang packages in all files to match your new repository's name (e.g. `github.com/your-org/your-repo`):

   ```
   sed -e 's,^module github\.com/sgnl-ai/adapter-template,github.com/your-org/your-repo,' -i go.mod
   ```

   ```
   find pkg/ -type f -name '*.go' | xargs -n 1 sed -n -e 's,github\.com/sgnl-ai/adapter-template,github.com/your-org/your-repo,p' -i
   ```

1. Modify the adapter implementation in package `pkg/adapter` to query your datasource. All the code that must be modified is identified with `SCAFFOLDING` comments. More implementation details are discussed in the [Understanding this Template](#3-understanding-this-template) section. For these steps, the code can be left as-is just to get the adapter running.

1. Create a JSON file (for example, `authTokens.json`) that will contain the tokens used to authenticate requests to the adapter. The tokens must be stored in the following format and note down the path of the file. 

   ```
   ["<token1>", "<token2>", ...]
   ```

1. If you don't need to build a Docker image, you can directly run the adapter. Set the `AUTH_TOKENS_PATH` environment variable to the path of the tokens file created in the previous step. Then run `go run cmd/adapter/main.go`. Otherwise, proceed to the next step.

1. Build the Docker image with the `adapter` command.
   ```
   docker build -t adapter:latest .
   ```
   **WARNING:** The image will contain the `ADAPTER_TOKENS` secrets file. **Do not push** this image to a public registry.
1. Run the adapter as a Docker container.
   ```
   docker run --rm -it -e AUTH_TOKENS_PATH=/path/to/file adapter:latest
   ```

### 2. Research the System of Record

This can be done simultaneously with code development, however it's important to understand the SoR before writing any code.

For each of the entities (i.e. API resources) that must be retrieved from the SoR, take note of the following:

#### Entity Endpoints

The endpoints to query the entity. For example, for a Jira User entity

```bash
https://your-domain.atlassian.net/rest/api/3/users/search # full URL
/rest/api/3/users/search # endpoint
```

Different entities will have different endpoints, and the format may not necessarily be consistent across entities.

#### Response Schemas

The response schemas for each entity. For example, an entity response may look like

```jsonc
{
  "accountId": "5b10a2844c20165700ede21g", // String
  "accountType": "atlassian", // String
  "displayName": "Admin", // String
  "emailAddress": "test@gmail.com", // String
  "active": true, // Bool
  "lastUpdated": "2021-08-06T18:00:00.000Z" // Date
}
```

Each of these JSON fields has a respective type. For example, `accountId` is a string, `active` is a boolean, etc. These must be noted because an adapter needs to know how to parse the response (and consequently the type of each field).

The format of any `datetime` types can also be noted, e.g. RFC3339, as a parsing optimization for an adapter. For example, you can specify these options:

https://github.com/SGNL-ai/adapter-template/blob/7fdf875997030e428911d1a3800ca1072906afc8/pkg/adapter/adapter.go#L101-L113

An adapter supports the following types: https://github.com/SGNL-ai/adapter-framework/blob/f6ad1c42cd34e37be8d4ba800309b5fb858040e1/api/adapter/v1/adapter.proto#L136-L157.

#### Authentication

The required authentication method for connecting to the SoR API. The following types are currently supported by SGNL:

- Basic Auth
- Bearer Token
- OAuth2 (Client Credentials Flow)

Basic Auth credentials and Bearer tokens are passed directly to an adapter in a `GetPage` request.

OAuth flows are performed by SGNL, which then passes the access token to the adapter for use in requests to the SoR.

#### Authorization

Ensure that the credentials being passed to an adapter have proper authorization to access the entities that need to be retrieved. For example, this may require setting the `scope` of an OAuth2 token.

#### API Restrictions

The request restrictions for each entity. For example,

- **Page size limits.** For paginated APIs, this is the maximum number of results that can be returned in a single request.
- **Filters.** Responses can be filtered to return a subset of objects or fields. These are features of the SoR API which can be leveraged by an adapter, if needed.
- **Results Ordered.** Are the results of the response ordered by some field? If so, take note of the field. Ordered results provides an optimization, but is not required.

**WARNING:**

Do not assume the results are ordered unless the API explicitly states that they are. An incorrect assumption will cause data to be synced into SGNL incorrectly.

A gRPC request to an adapter contains the above information. An adapter uses this information to construct an appropriate request to the SoR.

### 3. Understanding this Template

A simplified flow chart of an incoming gRPC request to an adapter is shown below:

![Adapter Flow](docs/assets/adapter_flow.png)

1. A gRPC request which follows the [adapter Protobuf schema](https://github.com/SGNL-ai/adapter-framework/blob/f2cafb0d963b54c350350967906ce59776d720a1/api/adapter/v1/adapter.proto) is sent by the ingestion service to the adapter. For testing, you can use Postman to send a gRPC request instead. An example request can be found in the [Local Testing](#4-local-testing) section.

2. The gRPC request is validated by `config.go` and `validation.go` and sent to `adapter.go`.

`config.go`

Here, you can specify additional configuration options for the adapter. For example, the API version to use, any HTTP headers to be sent to the SoR API, etc.
https://github.com/SGNL-ai/adapter-template/blob/594dbc2b4eace82fb26f54204d4bdaf1c5bc2e7a/pkg/adapter/config.go#L24-L44

`validation.go`

Here, you can specify additional validation rules for the gRPC request. For example, the maximum page size, the protocol, the authorization format, etc. `validation.go` also calls the `Validate` method in `config.go`, so any rules specified in `config.go` will also be applied.

https://github.com/SGNL-ai/adapter-template/blob/594dbc2b4eace82fb26f54204d4bdaf1c5bc2e7a/pkg/adapter/validation.go#L33-L103

3. The gRPC request is further parsed in `adapter.go`, where it is converted into a [`Request` struct](https://github.com/SGNL-ai/adapter-template/blob/594dbc2b4eace82fb26f54204d4bdaf1c5bc2e7a/pkg/adapter/client.go#L36-L61). The `Request` struct contains all the information needed to construct a request to the SoR. Additionally, `adapter.go` is responsible for:

- Constructing the request to the SoR, including any parsing of page cursors and request parameters.
- Converting the response from the SoR into `framework.Objects`, which is the format expected by the ingestion service. Any options for parsing (e.g. the format of date fields) should be specified here as well.

In general, this file should be kept lean. It should serve as a top level caller to other functions such as validation or making the request to the SoR.

4. The prepared `Request` struct is received by `datasource.go` and it uses this information to send an HTTP request to the SoR.

`datasource.go`

This is where the bulk of the code to actually make the HTTP request, parse the response, and handle pagination should be written.

https://github.com/SGNL-ai/adapter-template/blob/594dbc2b4eace82fb26f54204d4bdaf1c5bc2e7a/pkg/adapter/datasource.go#L84-L155

5. The SoR response is received by `datasource.go` and parsed.

6. The parsed SoR response is sent to `adapter.go` where it is converted into `framework.Objects`.

7. The `framework.Objects` are returned to the ingestion service, which then ingests the data into SGNL.

The majority of the required code changes are identified with `SCAFFOLDING` comments throughout this template. Most of the code in steps 6 and 7 should work out of the box, with the majority of the development being spent in steps 2, 3, and 4.

### 4. Local Testing

As specified in the [Getting Started](#1-getting-started) section, you can run the adapter locally either through Docker or directly with `go run`.

```go
go run cmd/adapter/main.go
```

By default, the adapter will listen on port 8080.

Using Postman, you can send a gRPC request to the adapter.

1. Define the [`GetPage` Protobuf definition](https://github.com/SGNL-ai/adapter-framework/blob/f2cafb0d963b54c350350967906ce59776d720a1/api/adapter/v1/adapter.proto).

![Define the `GetPage` Protobuf definition](/docs/assets/postman_proto_definition.png)

2. In the sidebar, click on **Collections** and create a new collection with the type set to **gRPC**.

3. Within this new collection, create a new gRPC request. Enter the URL of the adapter (e.g. `http://localhost:8080`) and select the `GetPage` method, which should be available in the dropdown if step 1 was completed successfully.

![Create a new gRPC request](/docs/assets/postman_new_grpc_request.png)

4. In the **Metadata** tab, add a `token` key and set the value to one of the tokens in the `ADAPTER_TOKENS` file.

5. In the **Message** tab, enter the `GetPage` request. It must follow the schema defined in step 1.

An example gRPC request:

```jsonc
{
  "cursor": "",
  "datasource": {
    "type": "AdapterType-1.0.0", // The type here should match the adapter type defined in `cmd/adapter/main.go`.
    "address": "{{address}}}",
    "auth": {
      "http_authorization": "Bearer {{token}}"
    },
    "config": "{{b64_encoded_string}}"
  },
  "entity": {
    "attributes": [
      {
        "external_id": "id",
        "type": "ATTRIBUTE_TYPE_STRING",
        "id": "id"
      }
    ],
    "external_id": "users",
    "id": "User",
    "ordered": false
  },
  "page_size": "100"
}
```

The `config` should be a base64 encoded string of the `Config` struct defined in `config.go`. For example, if the `Config` struct is

```go
type Config struct {
  APIVersion string `json:"apiVersion,omitempty"`
}
```

then the `config` field should be

```json
{
  "apiVersion": "v1"
}
```

which is base64 encoded to `eyJhcGlWZXJzaW9uIjoidjEifQ==`.

### Adapter Implementation Best Practice

- Keep the adapter implementation as lean as possible.
  - A logger is not needed as any errors returned by the adapter will be logged by the ingestion service.
  - Limit package usage to the standard library as that should be sufficient for most use cases.
- All errors should be handled with an appropriate error defined in the `adapter-framework`. Framework error messages should be a complete sentence starting with a capital letter and ending with a period. The complete list of `adapter-framework` errors can be found here: https://github.com/SGNL-ai/adapter-framework/blob/f2cafb0d963b54c350350967906ce59776d720a1/api/adapter/v1/adapter.proto#L249-L288.
