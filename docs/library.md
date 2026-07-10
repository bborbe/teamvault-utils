# Go Library Usage

`teamvault-cli` is also a Go library. Import it:

```bash
go get github.com/Seibert-Data/teamvault-cli/v5
```

The library package lives at `github.com/Seibert-Data/teamvault-cli/v5/pkg` (package `teamvault`). API reference: [pkg.go.dev](https://pkg.go.dev/github.com/Seibert-Data/teamvault-cli/v5/pkg).

## The `Connector` interface

`Connector` is the core interface for reading secrets:

```go
import (
    "context"
    "net/http"

    teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
    libtime "github.com/bborbe/time"
)

ctx := context.Background()

conn := teamvault.NewRemoteConnector(
    http.DefaultClient,
    teamvault.Url("https://teamvault.example.com"),
    teamvault.User("my-user"),
    teamvault.Password("my-pass"),
    libtime.NewCurrentDateTime(),
)

password, err := conn.Password(ctx, teamvault.Key("abc123"))
user, err := conn.User(ctx, teamvault.Key("abc123"))
url, err := conn.Url(ctx, teamvault.Key("abc123"))
file, err := conn.File(ctx, teamvault.Key("abc123"))
keys, err := conn.Search(ctx, "database")
```

## Connector variants

Wrap `NewRemoteConnector` to add behavior:

```go
teamvault.NewCacheConnector(remote)         // in-memory cache
teamvault.NewDiskFallbackConnector(remote)  // serve last-known value from disk when unreachable
teamvault.NewDummyConnector()               // fixtures, for tests
```

## Factory (recommended wiring)

`pkg/factory` builds a connector from a config file + flags/env, resolving the config, keychain, cache, and timeout for you — this is what the CLI uses:

```go
import (
    "github.com/Seibert-Data/teamvault-cli/v5/pkg/factory"
    teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
    libtime "github.com/bborbe/time"
)

httpClient, err := factory.CreateHttpClient(ctx)
conn, err := factory.CreateConnectorWithConfig(
    ctx, httpClient,
    teamvault.TeamvaultConfigPath("~/.teamvault.json"),
    teamvault.Url(""), teamvault.User(""), teamvault.Password(""),
    teamvault.Staging(false),
    true, // enable cache
    libtime.NewCurrentDateTime(),
)
```

## Template rendering

`ConfigParser` resolves `teamvaultUser`/`teamvaultPassword`/`teamvaultUrl` placeholders in a template; `ConfigGenerator` does it across a directory tree:

```go
parser := teamvault.NewConfigParser(conn)
out, err := parser.Parse(ctx, []byte(`password={{ "abc123" | teamvaultPassword }}`))

gen := teamvault.NewConfigGenerator(parser)
err = gen.Generate(ctx, teamvault.SourceDirectory("./templates"), teamvault.TargetDirectory("./config"))
```

## Testing against the library

Use the Counterfeiter mock or the dummy connector:

```go
import (
    teamvault "github.com/Seibert-Data/teamvault-cli/v5/pkg"
    "github.com/Seibert-Data/teamvault-cli/v5/pkg/mocks"
)

conn := &mocks.Connector{}
conn.PasswordReturns(teamvault.Password("test-password"), nil)
// ... inject conn into your code

// or, for no assertions:
conn := teamvault.NewDummyConnector()
```
