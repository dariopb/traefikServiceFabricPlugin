# Traefik service-fabric-plugin

This provider plugin allows Traefik to query the Service Fabric management API to discover what services are currently running on a Service Fabric cluster. The provider then maps routing rules to these service instances. The provider will take into account the Health and Status of each of the services to ensure requests are only routed to healthy service instances.
* **This is a community/unofficial implementation in development**

## Installation
The plugin needs to be configured in the Traefik static configuration before it can be used.

## Configuration
The plugin currently supports the following configuration settings:
* **pollInterval:**          The interval for polling the management endpoint for changes, in seconds.
* **clusterManagementURL:**  The URL for the Service Fabric Management API endpoint (e.g. `http://dariotraefik1.southcentralus.cloudapp.azure.com:19080/`)
* **certificate:**           The path to a certificate file or the PEM certificate content. If not provided, HTTP will be used.
* **certificateKey:**        The path to a private key file or the key content. If not provided, HTTP will be used.

## Example configuration

```yaml
entryPoints:
  web:
    address: :9999
    
api:
  dashboard: true

log:
  level: DEBUG

pilot:
  token: xxxxx

experimental:
  traefikServiceFabricPlugin:
    moduleName: github.com/dariopb/traefikServiceFabricPlugin
    version: v0.1.0

providers:
  plugin:
    traefikServiceFabricPlugin:
      pollInterval: 4s
      clusterManagementURL: http://dariotraefik1.southcentralus.cloudapp.azure.com:19080/
      #certificate : ./cert.pem
      #certificateKey: ./cert.key
```

## License
This software is released under the MIT License