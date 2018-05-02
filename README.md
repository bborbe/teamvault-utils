# Teamvault Utils

## Generate config directory with Teamvault secrets

Install:

```
go get github.com/bborbe/teamvault-utils/cmd/teamvault-config-dir-generator
```

Config:

```
{
    "url": "https://teamvault.example.com",
    "user": "my-user",
    "pass": "my-pass"
}
```

Run:

```
teamvault-config-dir-generator \
-teamvault-config="~/.teamvault.json" \
-source-dir=templates \
-target-dir=results \
-logtostderr \
-v=2
```

## Parse variable Teamvault secrets

Install:

```
go get github.com/bborbe/teamvault-utils/cmd/teamvault-config-parser
```

Sample config:

```
foo=bar
username={{ "vLVLbm" | teamvaultUser }}
password={{ "vLVLbm" | teamvaultPassword }}
bar=foo 
```

Run:

```
cat my.config | teamvault-config-parser
-teamvault-config="~/.teamvault.json" \
-logtostderr \
-v=2
```

## Teamvault Get Username

Install:

```
go get github.com/bborbe/teamvault-utils/cmd/teamvault-username
```

Run:

```
teamvault-username \
--teamvault-config ~/.teamvault-sm.json \
--teamvault-key vLVLbm
```

## Teamvault Get Password

Install:

```
go get github.com/bborbe/teamvault-utils/cmd/teamvault-password
```

Run:

```
teamvault-password \
--teamvault-config ~/.teamvault-sm.json \
--teamvault-key vLVLbm
```

## Teamvault Get Url

Install:

```
go get github.com/bborbe/teamvault-utils/cmd/teamvault-url
```

Run:

```
teamvault-url \
--teamvault-config ~/.teamvault-sm.json \
--teamvault-key vLVLbm
```
