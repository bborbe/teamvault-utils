# Teamvault Utils

[![Go Reference](https://pkg.go.dev/badge/github.com/bborbe/teamvault-utils/v4.svg)](https://pkg.go.dev/github.com/bborbe/teamvault-utils/v4)
[![CI](https://github.com/bborbe/teamvault-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/bborbe/teamvault-utils/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bborbe/teamvault-utils/v4)](https://goreportcard.com/report/github.com/bborbe/teamvault-utils/v4)

A Go library and CLI tools for interacting with TeamVault secret management system. Provides type-safe access to passwords, usernames, URLs, and files stored in TeamVault, with support for template parsing and configuration generation.

## Features

- **Type-Safe API**: Strongly typed interfaces for accessing TeamVault secrets
- **Multiple Connectors**: Remote, cache, disk fallback, and dummy connectors
- **Template Parsing**: Parse configuration templates with TeamVault placeholders
- **Config Generation**: Generate configuration files from templates
- **CLI Tools**: Command-line utilities for quick secret access
- **Dependency Injection**: Clean architecture with testable components

---

* [Setup (macOS, recommended)](#setup-macos-recommended)
* [Installation](#installation)
* [Quick Start](#quick-start)
* [Library Usage](#library-usage)
* [API Documentation](#api-documentation)
* [CLI Tools](#cli-tools)
* [Development](#development)
* [Testing](#testing)
* [License](#license)

---

## Setup (macOS, recommended)

On macOS, store your TeamVault password in the login Keychain so it never needs to appear in a plaintext config file:

1. Create a config file with **only** `url` and `user` — leave out `pass`:

   ```json
   {
       "url": "https://teamvault.example.com",
       "user": "my-user"
   }
   ```

2. Run `teamvault-login` once to verify your credentials and store the password in the Keychain:

   ```bash
   teamvault-login --teamvault-config ~/.teamvault.json
   ```

   The command prompts for your TeamVault password (hidden), verifies it against the API, and writes it to the macOS login Keychain on success.

**Multi-vault setup:** repeat for each config file:

```bash
teamvault-login --teamvault-config ~/.teamvault.json
teamvault-login --teamvault-config ~/.teamvault-sm.json
```

**Removing a stored password:**

```bash
security delete-generic-password -s teamvault-utils -a https://teamvault.example.com
```

> **Note:** Putting `pass` directly in the config file still works (the legacy path), but the password is stored in plaintext on disk. The Keychain path is strongly preferred on macOS.

> **Non-macOS:** `teamvault-login` verifies credentials but does not persist them. Users on Linux/Windows should continue to supply the password via flag, environment variable, or config file for now.

---

## Installation

```bash
go get github.com/bborbe/teamvault-utils/v4
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "net/http"

    teamvault "github.com/bborbe/teamvault-utils/v4"
    libtime "github.com/bborbe/time"
)

func main() {
    ctx := context.Background()

    // Create a connector
    connector := teamvault.NewRemoteConnector(
        http.DefaultClient,
        teamvault.Url("https://teamvault.example.com"),
        teamvault.User("my-user"),
        teamvault.Password("my-pass"),
        libtime.NewCurrentDateTime(),
    )

    // Retrieve a password
    password, err := connector.Password(ctx, teamvault.Key("vLVLbm"))
    if err != nil {
        panic(err)
    }

    fmt.Printf("Password: %s\n", password)
}
```

## Library Usage

### Using the Connector Interface

The `Connector` interface provides access to TeamVault secrets:

```go
import (
    "context"
    "net/http"

    teamvault "github.com/bborbe/teamvault-utils/v4"
    libtime "github.com/bborbe/time"
)

func example() {
    ctx := context.Background()

    connector := teamvault.NewRemoteConnector(
        http.DefaultClient,
        teamvault.Url("https://teamvault.example.com"),
        teamvault.User("my-user"),
        teamvault.Password("my-pass"),
        libtime.NewCurrentDateTime(),
    )

    // Get password
    password, err := connector.Password(ctx, teamvault.Key("abc123"))
    if err != nil {
        // handle error
    }

    // Get username
    user, err := connector.User(ctx, teamvault.Key("abc123"))
    if err != nil {
        // handle error
    }

    // Get URL
    url, err := connector.Url(ctx, teamvault.Key("abc123"))
    if err != nil {
        // handle error
    }

    // Get file
    file, err := connector.File(ctx, teamvault.Key("abc123"))
    if err != nil {
        // handle error
    }

    // Search for secrets
    keys, err := connector.Search(ctx, "database")
    if err != nil {
        // handle error
    }
}
```

### Template Parsing with ConfigParser

Parse configuration templates containing TeamVault placeholders:

```go
import (
    "context"

    teamvault "github.com/bborbe/teamvault-utils/v4"
)

func parseConfig(connector teamvault.Connector) {
    ctx := context.Background()

    parser := teamvault.NewConfigParser(connector)

    template := []byte(`
database:
  username: {{ "vLVLbm" | teamvaultUser }}
  password: {{ "vLVLbm" | teamvaultPassword }}
  url: {{ "vLVLbm" | teamvaultUrl }}
`)

    result, err := parser.Parse(ctx, template)
    if err != nil {
        // handle error
    }

    // result now contains resolved values
}
```

### Using Different Connector Types

**Cache Connector** (for performance):

```go
connector := teamvault.NewCacheConnector(
    teamvault.NewRemoteConnector(
        http.DefaultClient,
        teamvault.Url("https://teamvault.example.com"),
        teamvault.User("my-user"),
        teamvault.Password("my-pass"),
        libtime.NewCurrentDateTime(),
    ),
)
```

**Disk Fallback Connector** (for reliability):

```go
connector := teamvault.NewDiskFallbackConnector(
    teamvault.NewRemoteConnector(
        http.DefaultClient,
        teamvault.Url("https://teamvault.example.com"),
        teamvault.User("my-user"),
        teamvault.Password("my-pass"),
        libtime.NewCurrentDateTime(),
    ),
)
```

**Dummy Connector** (for testing):

```go
connector := teamvault.NewDummyConnector()
```

### Creating Connectors with Factory

Use the factory package for simplified connector creation:

```go
import (
    "context"
    "net/http"

    teamvault "github.com/bborbe/teamvault-utils/v4"
    "github.com/bborbe/teamvault-utils/v4/factory"
    libtime "github.com/bborbe/time"
)

func createConnector() (teamvault.Connector, error) {
    ctx := context.Background()

    httpClient, err := factory.CreateHttpClient(ctx)
    if err != nil {
        return nil, err
    }

    connector, err := factory.CreateConnectorWithConfig(
        ctx,
        httpClient,
        teamvault.TeamvaultConfigPath("~/.teamvault.json"),
        teamvault.Url(""),
        teamvault.User(""),
        teamvault.Password(""),
        teamvault.Staging(false),
        true, // enable cache
        libtime.NewCurrentDateTime(),
    )
    if err != nil {
        return nil, err
    }

    return connector, nil
}
```

## API Documentation

For complete API documentation, visit [pkg.go.dev](https://pkg.go.dev/github.com/bborbe/teamvault-utils/v4).

---

## CLI Tools

The library includes several command-line tools for quick secret access.

### Common flags

All `teamvault-*` CLI tools accept these flags:

```bash
--teamvault-timeout=5s   HTTP request timeout for TeamVault API calls (env: TEAMVAULT_TIMEOUT; default: 5s)
--cache                  Enable disk-fallback cache (env: CACHE)
```

**Cache behavior:** Cache is enabled if EITHER the `--cache` / `CACHE` env var is `true` OR the config file's `cacheEnabled: true` is set. There is no way to force-disable via CLI when the config opts in; edit the config file to disable.

### Teamvault Login

Verify TeamVault credentials and store the password in the macOS Keychain. Recommended first step on macOS — see [Setup (macOS, recommended)](#setup-macos-recommended) for the full flow.

Install:

```bash
go get github.com/bborbe/teamvault-utils/v4/cmd/teamvault-login
```

Run:

```bash
teamvault-login --teamvault-config ~/.teamvault.json
```

The command prompts for your TeamVault password (hidden input), verifies it against the API, and on success stores it in the macOS login Keychain. On non-macOS platforms it verifies only — no Keychain write.

### Teamvault Get Password

Install:

```bash
go get github.com/bborbe/teamvault-utils/v4/cmd/teamvault-password
```

Run:

```bash
teamvault-password \
  --teamvault-config ~/.teamvault.json \
  --teamvault-key vLVLbm
```

### Teamvault Get Username

Install:

```bash
go get github.com/bborbe/teamvault-utils/v4/cmd/teamvault-username
```

Run:

```bash
teamvault-username \
  --teamvault-config ~/.teamvault.json \
  --teamvault-key vLVLbm
```

### Teamvault Get URL

Install:

```bash
go get github.com/bborbe/teamvault-utils/v4/cmd/teamvault-url
```

Run:

```bash
teamvault-url \
  --teamvault-config ~/.teamvault.json \
  --teamvault-key vLVLbm
```

### Parse Config with Teamvault Secrets

Install:

```bash
go get github.com/bborbe/teamvault-utils/v4/cmd/teamvault-config-parser
```

Sample config template:

```bash
foo=bar
username={{ "vLVLbm" | teamvaultUser }}
password={{ "vLVLbm" | teamvaultPassword }}
url={{ "vLVLbm" | teamvaultUrl }}
```

Run:

```bash
cat my.config | teamvault-config-parser \
  --teamvault-config ~/.teamvault.json \
  --logtostderr \
  -v=2
```

### Generate Config Directory from Templates

Install:

```bash
go get github.com/bborbe/teamvault-utils/v4/cmd/teamvault-config-dir-generator
```

TeamVault config file (~/.teamvault.json):

```json
{
    "url": "https://teamvault.example.com",
    "user": "my-user",
    "pass": "my-pass",
    "cacheEnabled": true,
    "timeout": "30s"
}
```

`cacheEnabled: true` enables disk-fallback caching. `timeout` sets the HTTP request timeout (default: 5 seconds when absent).

Run:

```bash
teamvault-config-dir-generator \
  --teamvault-config ~/.teamvault.json \
  --source-dir templates \
  --target-dir results \
  --logtostderr \
  -v=2
```

---

## Development

### Running Tests

```bash
make test
```

### Code Generation (Mocks)

```bash
make generate
```

### Full Development Workflow

```bash
make precommit  # Format, test, lint, and check
```

---

## Full Example

Here's a complete, runnable example demonstrating real-world usage patterns combining multiple features:

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"

    teamvault "github.com/bborbe/teamvault-utils/v4"
    libtime "github.com/bborbe/time"
)

func main() {
    ctx := context.Background()

    // Create a cached connector for better performance
    // The cache connector wraps the remote connector and caches responses
    connector := teamvault.NewCacheConnector(
        teamvault.NewRemoteConnector(
            http.DefaultClient,
            teamvault.Url("https://teamvault.example.com"),
            teamvault.User("my-user"),
            teamvault.Password("my-pass"),
            libtime.NewCurrentDateTime(),
        ),
    )

    // Example 1: Retrieve individual secrets
    password, err := connector.Password(ctx, teamvault.Key("vLVLbm"))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting password: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Retrieved password (length: %d)\n", len(password.String()))

    user, err := connector.User(ctx, teamvault.Key("vLVLbm"))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting user: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Retrieved user: %s\n", user.String())

    // Example 2: Parse a configuration template
    parser := teamvault.NewConfigParser(connector)

    configTemplate := []byte(`
# Database Configuration
database:
  host: {{ "vLVLbm" | teamvaultUrl }}
  username: {{ "vLVLbm" | teamvaultUser }}
  password: {{ "vLVLbm" | teamvaultPassword }}

# Application Settings
app:
  environment: {{ "production" | env }}
  debug: false
`)

    result, err := parser.Parse(ctx, configTemplate)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("\nGenerated Configuration:")
    fmt.Println(string(result))

    // Example 3: Generate configuration files from templates
    generator := teamvault.NewConfigGenerator(parser)

    err = generator.Generate(
        ctx,
        teamvault.SourceDirectory("./templates"),
        teamvault.TargetDirectory("./config"),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error generating configs: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("\nConfiguration files generated successfully!")
}
```

This example demonstrates:
- Creating a cached connector for performance optimization
- Retrieving individual secrets (password, user)
- Parsing configuration templates with TeamVault placeholders
- Generating multiple configuration files from a template directory

---

## Testing

Testing code that uses this library is straightforward using the mock connector or dummy connector:

```go
import (
    "context"
    "testing"

    teamvault "github.com/bborbe/teamvault-utils/v4"
    "github.com/bborbe/teamvault-utils/v4/mocks"
)

func TestYourCode(t *testing.T) {
    ctx := context.Background()

    // Use mock connector for testing
    mockConnector := &mocks.Connector{}
    mockConnector.PasswordReturns(teamvault.Password("test-password"), nil)

    // Test your code with the mock
    result, err := mockConnector.Password(ctx, teamvault.Key("test-key"))
    if err != nil {
        t.Fatal(err)
    }

    if result != "test-password" {
        t.Errorf("expected test-password, got %s", result)
    }
}

func TestWithDummyConnector(t *testing.T) {
    // Or use dummy connector for simple tests
    connector := teamvault.NewDummyConnector()

    // Test your code with dummy connector
}
```

---

## License

This project is licensed under the BSD-style license. See the [LICENSE](LICENSE) file for details.
