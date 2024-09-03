# Traefik OpenFeature plugin

The plugin take header from request and evaluate the configuration feature flag against an Open Feature compliance backend. Then, forward the request to downstream service with enrich header

## Usage

### Configuration

For each plugin, the Traefik static configuration must define the module name (as is usual for Go packages).

The following declaration (given here in YAML) defines a plugin:

```yaml
# Static configuration

experimental:
  plugins:
    openfeatures:
      moduleName: github.com/dzungtr/traefik-openfeatures
      version: v0.0.1
```

Here is an example of a file provider dynamic configuration (given here in YAML), where the interesting part is the http.middlewares section:

```yaml
# Dynamic configuration

http:
  routers:
    my-router:
      rule: host(`demo.localhost`)
      service: service-foo
      entryPoints:
        - web
      middlewares:
        - check-featureflag

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    check-featureflag:
      plugin:
        openfeatures:
          Endpoint: http://host.internal.docker:9000
          Provider: flipt
          ContextHeaderKeys:
            - organization
            - user
          Service: my-service
          UserHeader: user
          FeatureHeaderPrefix: openfeature_
          Flags:
          # get value for flag
            api_version: string
            enable_check: bool
          # the request to to downstream service will have headers
          # openfeature_api_version: "some-version"
          # openfeature_enable_check: "false" 
```

### Reference

Key | Type | Description | Example
---|---|---|---
`Endpoint` | `string` | OpenFeature Backend url | http://host.internal.docker:9000
`Authorization` | `string` | authorization header to use with feature backend | `Bearer <token>`
`Provider` | `string` | OpenFeature provider | `flipt`, `noop`
`ContextHeaderKeys` | `[]string` | The header taken for context evaluation | `organization`, `user`
`Service` | `string` | Name of service | `my-service>`
`UserHeader` | `string` | The name of header used to extract user identifier | `x-user-id`
`FeatureHeaderPrefix` | `string` | The header's prefix for flag attatching to request to downstream service  | `openfeature_`
`Flags` | `{ <flag>: <type> }` | list of flag to evaluate in open feature backend | `api_version: string\|int\|float\|bool\|object`
