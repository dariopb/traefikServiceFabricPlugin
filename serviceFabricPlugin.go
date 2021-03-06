package traefikServiceFabricPlugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/traefik/genconf/dynamic"
	"github.com/traefik/genconf/dynamic/tls"

	sf "github.com/jjcollinge/servicefabric"
)

const (
	traefikServiceFabricExtensionKey = "Traefik"

	kindStateful  = "Stateful"
	kindStateless = "Stateless"
)

// Config the plugin configuration.
type Config struct {
	PollInterval         string `json:"pollInterval,omitempty"`
	ClusterManagementURL string `json:"clusterManagementURL,omitempty"`

	Certificate    string `json:"certificate,omitempty"`
	CertificateKey string `json:"certificateKey,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		PollInterval: "5s",
	}
}

// Provider a simple provider plugin.
type Provider struct {
	name         string
	pollInterval time.Duration

	clusterManagementURL string
	apiVersion           string
	tlsConfig            *ClientTLS
	sfClient             sfClient

	cancel func()
}

// New creates a new Provider plugin.
func New(ctx context.Context, config *Config, name string) (*Provider, error) {
	pi, err := time.ParseDuration(config.PollInterval)
	if err != nil {
		return nil, err
	}

	p := &Provider{
		name:                 name,
		apiVersion:           sf.DefaultAPIVersion,
		pollInterval:         pi,
		clusterManagementURL: config.ClusterManagementURL,
	}

	if config.CertificateKey != "" && config.Certificate != "" {
		p.tlsConfig = &ClientTLS{
			Cert:               config.Certificate,
			Key:                config.CertificateKey,
			InsecureSkipVerify: true,
		}
	}

	return p, nil
}

// Init the provider.
func (p *Provider) Init() error {
	var err error
	if p.pollInterval <= 0 {
		return fmt.Errorf("poll interval must be greater than 0")
	}

	log.Printf("Initializing: %s, version: %s", p.clusterManagementURL, p.apiVersion)

	tlsConfig, err := p.tlsConfig.CreateTLSConfig()
	if err != nil {
		return err
	}

	sfClient, err := sf.NewClient(http.DefaultClient, p.clusterManagementURL, p.apiVersion, tlsConfig)
	if err != nil {
		return err
	}
	p.sfClient = sfClient

	return nil
}

// Provide creates and send dynamic configuration.
func (p *Provider) Provide(cfgChan chan<- json.Marshaler) error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Print(err)
			}
		}()

		p.loadConfiguration(ctx, cfgChan)
	}()

	return nil
}

// Stop to stop the provider and the related go routines.
func (p *Provider) Stop() error {
	p.cancel()
	return nil
}

func (p *Provider) loadConfiguration(ctx context.Context, cfgChan chan<- json.Marshaler) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e, err := p.fetchState()
			if err != nil {
				log.Print(err)
				continue
			}

			configuration := p.generateConfiguration(e)

			cfgChan <- &dynamic.JSONPayload{Configuration: configuration}

		case <-ctx.Done():
			return
		}
	}
}

// Normalize Replace all special chars with `-`.
func normalize(name string) string {
	fargs := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// get function
	return strings.Join(strings.FieldsFunc(name, fargs), "-")
}

func (p *Provider) fetchState() ([]ServiceItemExtended, error) {
	sfClient := p.sfClient
	apps, err := sfClient.GetApplications()
	if err != nil {
		return make([]ServiceItemExtended, 0), nil
	}

	var results []ServiceItemExtended
	for _, app := range apps.Items {
		services, err := sfClient.GetServices(app.ID)
		if err != nil {
			return nil, err
		}

		for _, service := range services.Items {
			item := ServiceItemExtended{
				ServiceItem: service,
				Application: app,
			}

			if labels, err := getLabels(sfClient, &service, &app); err != nil {
				log.Print(err)
			} else {
				item.Labels = labels
			}

			if partitions, err := sfClient.GetPartitions(app.ID, service.ID); err != nil {
				log.Print(err)
			} else {
				for _, partition := range partitions.Items {
					partitionExt := PartitionItemExtended{PartitionItem: partition}

					switch {
					case isStateful(item):
						partitionExt.Replicas = getValidReplicas(sfClient, app, service, partition)
					case isStateless(item):
						partitionExt.Instances = getValidInstances(sfClient, app, service, partition)
					default:
						log.Printf("Unsupported service kind %s in service %s", partition.ServiceKind, service.Name)
						continue
					}

					item.Partitions = append(item.Partitions, partitionExt)
				}
			}

			results = append(results, item)
		}
	}

	return results, nil
}

func getValidReplicas(sfClient sfClient, app sf.ApplicationItem, service sf.ServiceItem, partition sf.PartitionItem) []sf.ReplicaItem {
	var validReplicas []sf.ReplicaItem

	if replicas, err := sfClient.GetReplicas(app.ID, service.ID, partition.PartitionInformation.ID); err != nil {
		log.Print(err)
	} else {
		for _, instance := range replicas.Items {
			if isHealthy(instance.ReplicaItemBase) && hasHTTPEndpoint(instance.ReplicaItemBase) {
				validReplicas = append(validReplicas, instance)
			}
		}
	}
	return validReplicas
}

func getValidInstances(sfClient sfClient, app sf.ApplicationItem, service sf.ServiceItem, partition sf.PartitionItem) []sf.InstanceItem {
	var validInstances []sf.InstanceItem

	if instances, err := sfClient.GetInstances(app.ID, service.ID, partition.PartitionInformation.ID); err != nil {
		log.Print(err)
	} else {
		for _, instance := range instances.Items {
			if isHealthy(instance.ReplicaItemBase) && hasHTTPEndpoint(instance.ReplicaItemBase) {
				validInstances = append(validInstances, instance)
			}
		}
	}
	return validInstances
}

func isPrimary(instanceData *sf.ReplicaItemBase) bool {
	return instanceData.ReplicaRole == "Primary"
}

func isHealthy(instanceData *sf.ReplicaItemBase) bool {
	return instanceData != nil && (instanceData.ReplicaStatus == "Ready" && instanceData.HealthState != "Error")
}

func hasHTTPEndpoint(instanceData *sf.ReplicaItemBase) bool {
	_, err := getReplicaDefaultEndpoint(instanceData)
	return err == nil
}

func getReplicaDefaultEndpoint(replicaData *sf.ReplicaItemBase) (string, error) {
	endpoints, err := decodeEndpointData(replicaData.Address)
	if err != nil {
		return "", err
	}

	var defaultHTTPEndpoint string
	for _, v := range endpoints {
		if strings.Contains(v, "http") {
			defaultHTTPEndpoint = v
			break
		}
	}

	if len(defaultHTTPEndpoint) == 0 {
		return "", errors.New("no default endpoint found")
	}
	return defaultHTTPEndpoint, nil
}

func decodeEndpointData(endpointData string) (map[string]string, error) {
	var endpointsMap map[string]map[string]string

	if endpointData == "" {
		return nil, errors.New("endpoint data is empty")
	}

	err := json.Unmarshal([]byte(endpointData), &endpointsMap)
	if err != nil {
		return nil, err
	}

	endpoints, endpointsExist := endpointsMap["Endpoints"]
	if !endpointsExist {
		return nil, errors.New("endpoint doesn't exist in endpoint data")
	}

	return endpoints, nil
}

func isStateful(service ServiceItemExtended) bool {
	return service.ServiceKind == kindStateful
}

func isStateless(service ServiceItemExtended) bool {
	return service.ServiceKind == kindStateless
}

// Return a set of labels from the Extension and Property manager
// Allow Extension labels to disable importing labels from the property manager.
func getLabels(sfClient sfClient, service *sf.ServiceItem, app *sf.ApplicationItem) (map[string]string, error) {
	labels, err := sfClient.GetServiceExtensionMap(service, app, traefikServiceFabricExtensionKey)
	if err != nil {
		log.Printf("Error retrieving serviceExtensionMap: %v", err)
		return nil, err
	}

	//if label.GetBoolValue(labels, traefikSFEnableLabelOverrides, traefikSFEnableLabelOverridesDefault) {
	if exists, properties, err := sfClient.GetProperties(service.ID); err == nil && exists {
		for key, value := range properties {
			labels[key] = value
		}
	}
	//}
	return labels, nil
}

func (p *Provider) generateConfiguration(e []ServiceItemExtended) *dynamic.Configuration {
	configuration := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  make(map[string]*dynamic.TCPRouter),
			Services: make(map[string]*dynamic.TCPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
	}

	configuration.HTTP.Middlewares["sf-stripprefixregex_stateless"] = &dynamic.Middleware{
		StripPrefixRegex: &dynamic.StripPrefixRegex{
			Regex: []string{"^/[^/]*/[^/]*/*"},
		},
	}
	configuration.HTTP.Middlewares["sf-stripprefixregex_statefull"] = &dynamic.Middleware{
		StripPrefixRegex: &dynamic.StripPrefixRegex{
			Regex: []string{"^/[^/]*/[^/]*/[^/]*/*"},
		},
	}

	for _, i := range e {
		name := strings.ReplaceAll(i.Name, "/", "-")
		name = normalize(name)
		//rule := i.Labels["traefik.router.rule.pinger"]

		if i.ServiceKind == kindStateless {
			rule := fmt.Sprintf("PathPrefix(`/%s`)", i.ID)
			configuration.HTTP.Routers[name] = &dynamic.Router{
				EntryPoints: []string{"web"},
				Service:     name,
				Rule:        rule,
				Middlewares: []string{"sf-stripprefixregex_stateless"},
			}
		} else if i.ServiceKind == kindStateful {
			for _, p := range i.Partitions {
				partitionID := p.PartitionInformation.ID
				name = fmt.Sprintf("%s-%s", name, partitionID)
				rule := fmt.Sprintf("PathPrefix(`/%s/%s`)", i.ID, partitionID)
				configuration.HTTP.Routers[name] = &dynamic.Router{
					EntryPoints: []string{"web"},
					Service:     name,
					Rule:        rule,
					Middlewares: []string{"sf-stripprefixregex_stateful"},
				}
			}
		}
		for _, p := range i.Partitions {
			if p.ServiceKind == kindStateless {
				lbServers := make([]dynamic.Server, 0)

				configuration.HTTP.Services[name] = &dynamic.Service{
					LoadBalancer: &dynamic.ServersLoadBalancer{
						PassHostHeader: boolPtr(true),
					},
				}

				for _, instance := range p.Instances {
					url, err := getReplicaDefaultEndpoint(instance.ReplicaItemBase)
					if err == nil && url != "" {
						lbServers = append(lbServers, dynamic.Server{
							URL: url,
						})
					}
				}

				configuration.HTTP.Services[name].LoadBalancer.Servers = lbServers
			} else if p.ServiceKind == kindStateful {
				partitionID := p.PartitionInformation.ID
				name = fmt.Sprintf("%s/%s", name, partitionID)
				lbServers := make([]dynamic.Server, 0)

				configuration.HTTP.Services[name] = &dynamic.Service{
					LoadBalancer: &dynamic.ServersLoadBalancer{
						PassHostHeader: boolPtr(true),
					},
				}

				for _, replica := range p.Replicas {
					if isPrimary(replica.ReplicaItemBase) && isHealthy(replica.ReplicaItemBase) {
						url, err := getReplicaDefaultEndpoint(replica.ReplicaItemBase)
						if err == nil && url != "" {
							lbServers = append(lbServers, dynamic.Server{
								URL: url,
							})
						}
					}
				}

				configuration.HTTP.Services[name].LoadBalancer.Servers = lbServers
			}
		}
	}

	return configuration
}

func boolPtr(v bool) *bool {
	return &v
}
