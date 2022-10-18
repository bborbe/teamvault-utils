# Teamvault Utils

## Generate config directory with Teamvault secrets

Install:

```bash
go get github.com/bborbe/teamvault-utils/cmd/teamvault-config-dir-generator
```

Config:

```json
{
    "url": "https://teamvault.example.com",
    "user": "my-user",
    "pass": "my-pass"
}
```

Run:

```bash
teamvault-config-dir-generator \
-teamvault-config="~/.teamvault.json" \
-source-dir=templates \
-target-dir=results \
-logtostderr \
-v=2
```

## Parse variable Teamvault secrets

Install:

```bash
go get github.com/bborbe/teamvault-utils/cmd/teamvault-config-parser
```

Sample config:

```bash
foo=bar
username={{ "vLVLbm" | teamvaultUser }}
password={{ "vLVLbm" | teamvaultPassword }}
bar=foo 
```

Run:

```bash
cat my.config | teamvault-config-parser
-teamvault-config="~/.teamvault.json" \
-logtostderr \
-v=2
```

## Teamvault Get Username

Install:

```bash
go get github.com/bborbe/teamvault-utils/cmd/teamvault-username
```

Run:

```bash
teamvault-username \
--teamvault-config ~/.teamvault-sm.json \
--teamvault-key vLVLbm
```

## Teamvault Get Password

Install:

```bash
go get github.com/bborbe/teamvault-utils/cmd/teamvault-password
```

Run:

```bash
teamvault-password \
--teamvault-config ~/.teamvault-sm.json \
--teamvault-key vLVLbm
```

## Teamvault Get Url

Install:

```bash
go get github.com/bborbe/teamvault-utils/cmd/teamvault-url
```

Run:

```bash
teamvault-url \
--teamvault-config ~/.teamvault-sm.json \
--teamvault-key vLVLbm
```
