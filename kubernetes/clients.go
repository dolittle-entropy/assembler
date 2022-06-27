package kubernetes

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

// CreateClients creates Kubernetes clients using the config from LoadConfig
func CreateClients() (dynamic.Interface, *discovery.DiscoveryClient, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return dynamicClient, discoveryClient, nil
}
