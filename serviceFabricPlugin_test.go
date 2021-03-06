package traefikServiceFabricPlugin

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNew(t *testing.T) {
	config := CreateConfig()
	config.PollInterval = "1s"
	//config.ClusterManagementURL = "http://darioclus3.southcentralus.cloudapp.azure.com:19080/"
	config.ClusterManagementURL = "http://localhost:19080/"

	provider, err := New(context.Background(), config, "test")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		err = provider.Stop()
		if err != nil {
			t.Fatal(err)
		}
	})

	err = provider.Init()
	if err != nil {
		t.Fatal(err)
	}

	cfgChan := make(chan json.Marshaler)

	err = provider.Provide(cfgChan)
	if err != nil {
		t.Fatal(err)
	}

	data := <-cfgChan
	data = data
}
