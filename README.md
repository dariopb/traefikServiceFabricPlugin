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
    version: v0.2.1

providers:
  plugin:
    traefikServiceFabricPlugin:
      pollInterval: 4s
      clusterManagementURL: http://dariotraefik1.southcentralus.cloudapp.azure.com:19080/
      #certificate : ./cert.pem
      #certificateKey: ./cert.key
      #insecureSkipVerify: true
```

## Sample deployments

For a complete way to deploy and use the proxy in a Service Fabric cluster, please look here: https://github.com/dariopb/sf-reverseproxies-templates (those are **samples** only and should be modified according to your needs)

### ServiceManifest file

This is a sample SF enabled service showing the currently supported labels. If the sf name is fabric:/pinger/PingerService, the endpoint [endpointName] will be expose at that prefix: '/pinger/PingerService/'

```xml
  ...
  <ServiceTypes>
    <StatelessServiceType ServiceTypeName="PingerServiceType" UseImplicitHost="true">
      <Extensions>
        <Extension Name="traefik">
        <Labels xmlns="http://schemas.microsoft.com/2015/03/fabact-no-schema">
          <Label Key="traefik.http.defaultEP">true</Label>
          <Label Key="traefik.http.defaultEP.loadbalancer.passhostheader">true</Label>
          <Label Key="traefik.http.defaultEP.loadbalancer.healthcheck.path">/</Label>
          <Label Key="traefik.http.defaultEP.loadbalancer.healthcheck.interval">10s</Label>
          <Label Key="traefik.http.defaultEP.loadbalancer.healthcheck.scheme">http</Label>
        </Labels>
        </Extension>
      </Extensions>
    </StatelessServiceType>
  </ServiceTypes>
  ...
```

---

The only required label to expose a service via the reverse proxy is the **traefik.http.[endpointName]** one. Setting only this label will expose the service on a well known path and handle the basic scenarios.

```
http(s)://<Cluster FQDN | internal IP>:Port/ApplicationInstanceName/ServiceInstanceName/{PartitionGuid}/<Suffix path>
```

If you need to change the routes or add middleware then you can add different labels configuring them.


## Supported Labels (since 0.2.x)

*Http enable section*

* **traefik.http.[endpointName]**    Enables exposing an http service via the reverse proxy.

Rule section

* **traefik.http.[endpointName].rule**    Traefik rule to apply [PathPrefix(`/dario`))]. This rule is added on top of the default path generation. If this is set, you **have** to define a middleware to remove the prefix for the service to receive the stripped path.

*Loadbalancer section*

* **traefik.http.[endpointName].loadbalancer.passhostheader**          passhostheaders ['true'/'false']
* **traefik.http.[endpointName].loadbalancer.healthcheck.path**        Healthcheck endpoint path ['/healtz']
* **traefik.http.[endpointName].loadbalancer.healthcheck.interval**    Healthcheck interval ['10s']
* **traefik.http.[endpointName].loadbalancer.healthcheck.scheme**      Healthcheck scheme ['http']

*Middleware section*

* **traefik.http.[endpointName].middleware.stripprefix**    prefix to strip ['/dario']

## License

This software is released under the MIT License
