---
status: completed
spec: [007-create-update-write-commands]
summary: Added Writer interface with Create/Update/GeneratePassword to teamvault library
execution_id: sm-teamvault-cli-exec-012-spec-007-library-write-layer
dark-factory-version: dev
created: "2026-07-14T14:00:00Z"
queued: "2026-07-14T13:56:02Z"
started: "2026-07-14T13:56:29Z"
completed: "2026-07-14T14:04:55Z"
branch: dark-factory/create-update-write-commands
---

<summary>
- Adds a write capability to the TeamVault library so it can create new secrets and update existing ones, alongside the existing read-only connector.
- Creating a secret posts its metadata and value to the server and returns the new secret's key and API URL; the content type (password vs file) is chosen by the caller.
- Updating a secret patches only the fields the caller supplies and can create a new revision when a value is included; metadata-only updates are allowed.
- Adds a helper that asks the server to generate a strong password, so a secret can be created with a server-generated value.
- Reuses the exact Basic-auth header, HTTP client, timeout, and non-2xx / authentication error messaging the read path already uses — including the `login` hint on 401/403.
- The existing five-method read interface is left completely untouched, so this stays a minor (feature) release with no breaking change for external users of the library.
- Ships with hermetic HTTP tests that assert the request method, auth header, and JSON body shape for create, update, and generate-password.
</summary>

<objective>
Add a write layer to the `teamvault` library (package `teamvault`, in `pkg/`) that can `POST` a new secret, `PATCH` an existing secret, and call the generate-password endpoint — reusing the read path's Basic-auth header, HTTP client, timeout, and error/authentication messaging. Expose it through a NEW `Writer` interface and a `Create*` factory function WITHOUT adding any method to the existing exported `Connector` interface (so the release stays a minor bump). This prompt builds only the wire layer plus its hermetic HTTP tests; no CLI command exists yet.
</objective>

<context>
Read `CLAUDE.md` for project conventions (bborbe error wrapping, no `context.Background()` in business logic, `libtime` types, Ginkgo/Gomega + Counterfeiter, binary at module root, library in `pkg/` as `package teamvault`).

Read before changing:
- `pkg/connector.go` — the existing exported `Connector` interface. It MUST stay exactly these five methods and nothing else: `Password`, `User`, `Url`, `File`, `Search`. Do NOT add a method here.
- `pkg/remote-connector.go` — the read connector. Study and REUSE its patterns:
  - `NewRemoteConnector(httpClient *http.Client, url Url, user User, pass Password, currentDateTime time.CurrentDateTime) Connector` — constructor shape to mirror.
  - `remoteConnector.createHeader()` builds the Basic-auth header:
    ```go
    httpHeader.Add("Authorization", fmt.Sprintf("Basic %s",
        base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", r.user.String(), r.pass.String())))))
    httpHeader.Add("Content-Type", "application/json")
    ```
  - `remoteConnector.call(...)` shows the non-2xx handling to replicate for writes, in particular the 401/403 branch whose exact wording drives the login retry loop (`isAuthError` in `pkg/cli/login.go` matches on the literal `status: %d`):
    ```go
    if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
        return errors.Errorf(ctx,
            "request to %s failed with status: %d (authentication failed) — run `teamvault-cli login` to (re)store your TeamVault password in the Keychain",
            url, resp.StatusCode)
    }
    return errors.Errorf(ctx, "request to %s failed with status: %d", url, resp.StatusCode)
    ```
    Keep the `status: %d` substring (with colon) verbatim so the login retry loop keeps working.
  - The read path builds URLs as `fmt.Sprintf("%s/api/secrets/%s/", r.url.String(), key.String())` and normalizes the base URL with `.Normalize()` (trims trailing slash). Do the same for write URLs.
- `pkg/key.go` — `type Key string`, `Key.String()`, `Key.Validate(ctx)`.
- `pkg/api-url.go` — `type ApiUrl string`, `ApiUrl.String()`, `ApiUrl.Key() (Key, error)` (extracts the key from a secret api_url path). The create response's `api_url` yields the new key via `ApiUrl.Key()`.
- `pkg/file.go` — `type File string`; `File.Content()` base64-DECODES. For writes you base64-ENCODE raw file bytes with `base64.StdEncoding.EncodeToString(bytes)`.
- `pkg/factory/factory.go` — factory package (`package factory`). `CreateRemoteConnector(...)` is the pattern for a `Create*` composition function; add the writer factory alongside it. Note the factory has ZERO business logic — pure composition.
- `pkg/remote-connector_test.go` — the hermetic HTTP test pattern. Tests use `github.com/onsi/gomega/ghttp` (`ghttp.NewServer()`), `libhttp.CreateDefaultHttpClient()` from `github.com/bborbe/http`, and `libtime.NewCurrentDateTime()`. Mirror this exact style for the writer tests, using `ghttp.VerifyRequest`, `ghttp.VerifyBasicAuth`, `ghttp.VerifyJSONRepresenting` / `ghttp.VerifyContentType`, and `ghttp.RespondWithJSONEncoded`.

Read these coding guides (in-container paths):
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-patterns.md` — public interface + private struct + `New*` constructor, counterfeiter annotations, `errors.Wrapf(ctx, err, "...")`.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-factory-pattern.md` — zero-logic factories, `Create*` prefix, return interfaces.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-error-wrapping-guide.md` — `bborbe/errors` API, never `fmt.Errorf`, never `context.Background()` in `pkg/`.
- `/home/node/.claude/plugins/marketplaces/coding/docs/go-testing-guide.md` — Ginkgo/Gomega suite files, coverage ≥80%, external test package (`package teamvault_test`).

TeamVault write-API wire shapes (confirmed authoritative — from the Django source `teamvault/apps/secrets/api/`; the spec Constraints carry these as the contract):
- Create: `POST <base>/api/secrets/` with body `{content_type, name, username?, url?, description?, secret_data}`.
  - `content_type` is `"password"` or `"file"` (no `cc` — out of scope).
  - password secret: `secret_data = {"password": "<value>"}`.
  - file secret: `secret_data = {"file_content": "<base64 of raw bytes>"}`.
  - Success response is JSON containing `api_url` (e.g. `"http://host/api/secrets/<key>/"`); derive the new `Key` with `ApiUrl.Key()`.
- Update: `PATCH <base>/api/secrets/<key>/` with ANY subset of `name/username/url/description` and/or a new `secret_data`. Only include the fields the caller passed. Omit `secret_data` entirely for a metadata-only update.
- Generate password: `POST <base>/api/generate_password/` (empty JSON body `{}` is fine); success response is JSON with a `password` field. This is server-stateless (returns a value, stores nothing).
- Auth: HTTP Basic, identical to the read path.
</context>

<requirements>

1. **Create `pkg/writer.go`** (`package teamvault`) defining the write contract WITHOUT touching `pkg/connector.go`.

   a. Define value/param types for a create request. Use a struct with typed fields so the caller (the CLI) does not hand-build JSON:
      ```go
      // ContentType is the TeamVault secret content type. Only "password" and
      // "file" are supported; "cc" is deliberately out of scope.
      type ContentType string

      const (
          ContentTypePassword ContentType = "password"
          ContentTypeFile     ContentType = "file"
      )

      // CreateSecret describes a new secret to create. Exactly one of Password
      // or FileContent carries the value; ContentType selects which. Metadata
      // fields (Username/Url/Description) are optional and omitted from the
      // request body when empty.
      type CreateSecret struct {
          ContentType ContentType
          Name        string
          Username    string
          Url         string
          Description  string
          Password    Password // used when ContentType == ContentTypePassword
          FileContent []byte   // raw bytes; base64-encoded into secret_data when ContentType == ContentTypeFile
      }

      // UpdateSecret describes a partial update. Only non-nil pointer fields are
      // sent, so a metadata-only update omits secret_data entirely, and each
      // absent field is left untouched server-side. A non-nil Password or
      // FileContent creates a new revision.
      type UpdateSecret struct {
          Name        *string
          Username    *string
          Url         *string
          Description  *string
          Password    *Password // non-nil => new password revision
          FileContent []byte    // non-nil (may be empty) => new file revision
      }
      ```
      Use pointer fields on `UpdateSecret` so "field not provided" (nil) is distinguishable from "field set to empty string" — this is what lets a metadata-only PATCH omit `secret_data`. Add GoDoc to every exported item.

   b. Define the `Writer` interface with a counterfeiter directive (mirror the `Connector` directive form in `pkg/connector.go`):
      ```go
      //counterfeiter:generate -o mocks/writer.go --fake-name Writer . Writer

      // Writer creates and updates TeamVault secrets. It is intentionally
      // separate from Connector so the read interface stays unchanged (a new
      // method on the exported Connector would be a breaking, major-bump change).
      type Writer interface {
          // Create posts a new secret and returns its key and api_url.
          Create(ctx context.Context, secret CreateSecret) (Key, ApiUrl, error)
          // Update patches an existing secret named by key. Only the fields set
          // in UpdateSecret are sent.
          Update(ctx context.Context, key Key, secret UpdateSecret) (Key, ApiUrl, error)
          // GeneratePassword asks the server to generate a strong password.
          GeneratePassword(ctx context.Context) (Password, error)
      }
      ```
      `Create`/`Update` return `(Key, ApiUrl, error)` because the CLI's `--json` output needs both key and api_url (see prompt 2). For `Update`, return the key that was patched and the api_url from the server response.

2. **Create `pkg/remote-writer.go`** (`package teamvault`) with the private implementation and its `New*` constructor:
   ```go
   // NewRemoteWriter creates a Writer that issues POST/PATCH calls to a remote
   // TeamVault instance, reusing HTTP Basic auth identical to the read path.
   func NewRemoteWriter(
       httpClient *http.Client,
       url Url,
       user User,
       pass Password,
       currentDateTime time.CurrentDateTime,
   ) Writer {
       return &remoteWriter{
           httpClient:      httpClient,
           url:             url.Normalize(),
           user:            user,
           pass:            pass,
           currentDateTime: currentDateTime,
       }
   }
   ```
   Mirror `remoteConnector`'s fields and `createHeader()` exactly (same Basic-auth + `Content-Type: application/json` header). Do NOT export a second copy of the header helper — implement a private `createHeader()` on `remoteWriter` identical in body to the connector's, or factor a shared unexported helper; either is fine, but the produced header MUST be byte-identical to the read path's.

   a. Implement a private `call` helper on `remoteWriter` that sends a method + JSON body and decodes the response, reusing the read path's non-2xx handling VERBATIM (including the 401/403 branch wording with the literal `status: %d` and the `run \`teamvault-cli login\`` hint). Signature suggestion:
      ```go
      func (w *remoteWriter) call(ctx context.Context, method, url string, body any, response any) error
      ```
      - Marshal `body` with `encoding/json` (skip the body for a nil `body`). Build the request with `http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))`.
      - Add the Basic-auth + `Content-Type: application/json` headers.
      - Execute via `w.httpClient.Do(req)`; wrap transport errors with `errors.Wrapf(ctx, err, "execute request failed")`.
      - `defer resp.Body.Close()`.
      - On `resp.StatusCode/100 != 2`, return the SAME two error branches as the read path (401/403 hint branch, else generic `status: %d`). Read the response body into the error message when helpful, but NEVER log or return any secret value.
      - On 2xx, if `response != nil`, `json.NewDecoder(resp.Body).Decode(response)`.
      - Add a `#nosec G704` comment on `w.httpClient.Do(req)` mirroring the read path's existing comment (URLs are built from the configured base URL + API paths, not raw user input).
      - Do NOT log secret values. `glog.V(4)` may log the URL/method (as the read path does), but never the request body.

   b. Implement `Create`: build the `secret_data` sub-object from `ContentType`:
      - `ContentTypePassword` → `{"password": string(secret.Password)}`.
      - `ContentTypeFile` → `{"file_content": base64.StdEncoding.EncodeToString(secret.FileContent)}`.
      Build the top-level body `{content_type, name, secret_data}` plus `username/url/description` only when their string is non-empty (use `omitempty` JSON tags or conditional map construction). POST to `fmt.Sprintf("%s/api/secrets/", w.url.String())`. Decode the response into a struct with `ApiUrl ApiUrl \`json:"api_url"\``, then `key, err := apiUrl.Key()`; return `(key, apiUrl, nil)`.

   c. Implement `Update`: build a body containing ONLY the fields whose pointer is non-nil (`name/username/url/description`), and include `secret_data` ONLY when `secret.Password != nil` (`{"password": ...}`) or `secret.FileContent != nil` (`{"file_content": base64...}`). When no value field is set, the body MUST NOT contain a `secret_data` key at all (metadata-only update). PATCH to `fmt.Sprintf("%s/api/secrets/%s/", w.url.String(), key.String())`. Decode `api_url` from the response; return `(key, apiUrl, nil)` (key is the input key).

   d. Implement `GeneratePassword`: POST `{}` to `fmt.Sprintf("%s/api/generate_password/", w.url.String())`, decode a struct with `Password Password \`json:"password"\``, return it.

3. **Add a factory function** in `pkg/factory/factory.go` (`package factory`), a pure-composition `Create*` mirroring `CreateRemoteConnector`:
   ```go
   // CreateRemoteWriter creates a Writer that communicates directly with a remote TeamVault API.
   func CreateRemoteWriter(
       httpClient *http.Client,
       apiURL teamvault.Url,
       apiUser teamvault.User,
       apiPassword teamvault.Password,
       currentDateTime libtime.CurrentDateTime,
   ) teamvault.Writer {
       return teamvault.NewRemoteWriter(httpClient, apiURL, apiUser, apiPassword, currentDateTime)
   }
   ```
   Do NOT route the writer through staging/cache decorators — writes always hit the live server (spec Non-goal: no write-path caching or disk-fallback). Zero business logic in the factory.

4. **Regenerate mocks.** Run `make generate` so `pkg/mocks/writer.go` (fake `Writer`) is produced from the counterfeiter directive. Verify `pkg/mocks/writer.go` exists after.

5. **Tests** — add `pkg/remote-writer_test.go` (external package `teamvault_test`, Ginkgo/Gomega, `ghttp`), mirroring `pkg/remote-connector_test.go`'s setup. Cover ALL of:

   a. **Create password secret** — drive `writer.Create(ctx, CreateSecret{ContentType: ContentTypePassword, Name: "n", Password: "p"})` against a `ghttp.Server`. Assert with `ghttp.CombineHandlers(...)`:
      - `ghttp.VerifyRequest(http.MethodPost, "/api/secrets/")`
      - `ghttp.VerifyBasicAuth(user, pass)` — proves the Basic-auth header is present and correct.
      - the request body JSON has `content_type == "password"`, `name == "n"`, `secret_data.password == "p"`, and NO `file_content`. (Read the body and unmarshal, or use `ghttp.VerifyJSONRepresenting(map[string]any{...})` with the exact expected object.)
      - respond with `ghttp.RespondWithJSONEncoded(200, map[string]any{"api_url": "http://<server>/api/secrets/AbC123/"})` and assert the returned `Key == "AbC123"` and the returned `ApiUrl` matches.

   b. **Create file secret** — `CreateSecret{ContentType: ContentTypeFile, Name: "n", FileContent: []byte("hello")}`; assert the body has `content_type == "file"` and `secret_data.file_content == base64.StdEncoding.EncodeToString([]byte("hello"))` and NO `password`.

   c. **Update metadata-only** — `writer.Update(ctx, "AbC123", UpdateSecret{Description: ptr("d")})` (use a small `ptr` helper or `libtime`/existing pointer util — check `pkg/` for an existing `Ptr`/pointer helper first via `grep -rn "func Ptr" pkg/`; if none, define a tiny generic `func ptr[T any](v T) *T { return &v }` in the test file). Assert: method `PATCH`, path `/api/secrets/AbC123/`, body contains `description == "d"` and does NOT contain a `secret_data` key. This is the spec AC "update sends no secret_data when no value flag is passed" — assert the absence explicitly.

   d. **Update with new password** — `UpdateSecret{Password: ptr(Password("newpw"))}`; assert PATCH body contains `secret_data.password == "newpw"`.

   e. **GeneratePassword** — POST `/api/generate_password/`, respond `{"password":"gen3rat3d"}`, assert returned `Password == "gen3rat3d"`.

   f. **401 → auth-hint error** — respond `401`; assert the returned error message contains the literal substring `status: 401` AND `teamvault-cli login` (so `isAuthError` and the user hint both keep working).

   g. **Non-2xx server rejection (e.g. 400)** — respond `400`; assert a non-nil error containing `status: 400`. (Covers the "server rejects payload / immutable content_type" failure mode surfacing as a non-zero result.)

   h. **Transport failure** — point the writer at a closed/unreachable server (e.g. call `server.Close()` before the request, or use an unroutable URL) and assert `Create` returns a non-nil error and no panic.

   Aim for ≥80% statement coverage of `pkg/remote-writer.go`. Verify with:
   `go test -coverprofile=/tmp/cover.out ./pkg/... && go tool cover -func=/tmp/cover.out | grep remote-writer`.

</requirements>

<constraints>
- The existing exported `Connector` interface in `pkg/connector.go` MUST remain exactly `Password`, `User`, `Url`, `File`, `Search` and nothing else — this preserves the minor (feature) release. Do NOT add a write method to `Connector`. Write is exposed via the NEW `Writer` interface only.
- The read API in `docs/library.md` must still compile unchanged (do not alter existing exported read signatures/types).
- Reuse the read path's Basic-auth header construction and non-2xx error mapping. The 401/403 error message MUST keep the literal `status: %d` substring and the `run \`teamvault-cli login\`` hint verbatim — `isAuthError` in `pkg/cli/login.go` depends on `status: 401`/`status: 403`.
- No secret value (password, file bytes) is ever written to a `glog`/`slog` line or returned in an error string. Log only URL/method at `glog.V(4)`, never the body.
- Request bodies are built with `encoding/json` (values encoded, never string-interpolated into the path or body).
- Writes always hit the live server — do NOT wrap the writer in cache or disk-fallback decorators.
- Errors wrapped via `github.com/bborbe/errors` (`errors.Wrapf`/`errors.Errorf`); never `fmt.Errorf`; never `context.Background()` in `pkg/` business logic; never `errors.Wrapf(ctx, nil, ...)`.
- Time via `github.com/bborbe/time` types (`time.CurrentDateTime`); no direct `time.Now()`.
- All exported items keep GoDoc comments per `docs/dod.md`.
- Tests use Ginkgo/Gomega with Counterfeiter mocks under `pkg/mocks/`; external test package `teamvault_test`.
- Include the BSD license header block (copy from any existing `pkg/*.go` file) at the top of every new `.go` file.
- Do NOT add a `create`/`update` CLI command in this prompt — that is prompt 2. Do NOT touch fakevault, scenarios, or docs — that is prompt 3.
- Do NOT commit — dark-factory handles git.
- Existing tests must still pass.
</constraints>

<verification>
- `make test` passes (fast loop during development).
- `make precommit` exits 0 (final validation).
- `grep -A8 'type Connector interface' pkg/connector.go` still lists exactly `Password`, `User`, `Url`, `File`, `Search` and nothing else (Connector unchanged).
- `grep -n 'type Writer interface' pkg/writer.go` returns ≥1.
- `test -f pkg/mocks/writer.go` succeeds (mock regenerated).
- `grep -rn 'CreateRemoteWriter' pkg/factory/factory.go` returns ≥1.
- `grep -rn 'status: %d' pkg/remote-writer.go` returns ≥1 and `grep -rn 'teamvault-cli login' pkg/remote-writer.go` returns ≥1 (auth error wording reused).
- `grep -rn 'fmt.Errorf' pkg/writer.go pkg/remote-writer.go` returns 0 matches.
- `go test ./pkg/...` passes, including the new create/update/generate/401/400/transport tests.
- `go test -coverprofile=/tmp/cover.out ./pkg/... && go tool cover -func=/tmp/cover.out | grep remote-writer` shows ≥80% for the writer file.
</verification>
