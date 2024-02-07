# traefik-api-key-auth

A Traefik plugin (middleware) that protects your ingress routes by checking against a list of valid api keys.

[![Build Status](https://github.com/JimCronqvist/traefik-api-key-auth/workflows/Main/badge.svg?branch=master)](https://github.com/jimcronqvist/traefik-api-key-auth/actions)

## Usage

For a plugin to be active for a given Traefik instance, it must be declared in the static configuration.

Plugins are parsed and loaded exclusively during startup, which allows Traefik to check the integrity of the code and catch errors early on.
If an error occurs during loading, the plugin is disabled.

For security reasons, it is not possible to start a new plugin or modify an existing one while Traefik is running.

Once loaded, middleware plugins behave exactly like statically compiled middlewares.
Their instantiation and behavior are driven by the dynamic configuration.

### Configuration

For each plugin, the Traefik static configuration must define the module name (as is usual for Go packages).

The following declaration (given here in YAML) defines a plugin:

```yaml
# Static configuration - File (yaml) / Helm Chart values

experimental:
  plugins:
    traefik-api-key-auth:
      moduleName: github.com/jimcronqvist/traefik-api-key-auth
      version: 0.1.0
```

 
```
# Static configuration - CLI

--experimental.plugins.sablier.modulename=github.com/acouvreur/sablier
--experimental.plugins.sablier.version=v1.7.0-beta.2
```



### Kubernetes CRD - IngressRoute Middleware

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: application-api-key-middleware
spec:
  plugin:
    traefik-api-key-auth:
      headerName: x-api-key
      keys:
        - some-api-key
      #removeHeaderOnSuccess: true
      #bearerToken: true
```

### Local Mode

Traefik also offers a developer mode that can be used for temporary testing of plugins not hosted on GitHub.
To use a plugin in local mode, the Traefik static configuration must define the module name (as is usual for Go packages) and a path to a [Go workspace](https://golang.org/doc/gopath_code.html#Workspaces), which can be the local GOPATH or any directory.

The plugins must be placed in `./plugins-local` directory,
which should be in the working directory of the process running the Traefik binary.
The source code of the plugin should be organized as follows:

```
./plugins-local/
    └── src
        └── github.com
            └── traefik
                └── plugindemo
                    ├── demo.go
                    ├── demo_test.go
                    ├── go.mod
                    ├── LICENSE
                    ├── Makefile
                    └── readme.md
```

```yaml
# Static configuration

experimental:
  localPlugins:
    example:
      moduleName: github.com/traefik/plugindemo
```

(In the above example, the `plugindemo` plugin will be loaded from the path `./plugins-local/src/github.com/traefik/plugindemo`.)

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
        - my-plugin

  services:
   service-foo:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:5000
  
  middlewares:
    my-plugin:
      plugin:
        example:
          headers:
            Foo: Bar
```
