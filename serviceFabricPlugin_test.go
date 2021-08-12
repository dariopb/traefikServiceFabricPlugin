package traefikServiceFabricPlugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.Handle("/Applications/TestApplication/$/GetServices", mock("services.json"))
	mux.Handle("/Applications/TestApplication/$/GetServices/TestApplication/TestService/$/GetPartitions/", mock("partitions.json"))
	mux.Handle("/Applications/", mock("applications.json"))
	mux.Handle("/ApplicationTypes/TestApplicationType/$/GetServiceTypes", mock("extensions.json"))
	mux.Handle("/Names/TestApplication/TestService", mock("services.json"))
	mux.Handle("/Names/TestApplication/TestService/$/GetProperties", mock("properties.json"))

	config := CreateConfig()
	config.PollInterval = "1s"
	//config.ClusterManagementURL = "http://dariolinclus1.southcentralus.cloudapp.azure.com:19080"
	config.ClusterManagementURL = server.URL

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

	expectedJSON, err := os.ReadFile(filepath.FromSlash("./testdata/dynamic_configuration.json"))
	if err != nil {
		t.Fatal(err)
	}

	dataJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	expected := strings.ReplaceAll(string(expectedJSON), "\r", "")
	expected = strings.TrimSpace(expected)

	dataClean := strings.ReplaceAll(string(dataJSON), "\r", "")

	if expected != dataClean {
		t.Fatalf("got %s, want: %s", expected, dataClean)
	}
}

func mock(filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		file, err := os.Open(filepath.Join("testdata", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = file.Close() }()

		_, _ = io.Copy(rw, file)
	}
}
