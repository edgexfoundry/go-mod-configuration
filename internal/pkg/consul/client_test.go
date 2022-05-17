//
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package consul

import (
	"fmt"
	"log"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-configuration/v2/pkg/types"
)

const (
	serviceName    = "consulUnitTest"
	consulBasePath = "edgex/core/1.0/"
)

// change values to localhost and 8500 if you need to run tests against real Consul service running locally
var (
	testHost = ""
	port     = 0
)

type LoggingInfo struct {
	EnableRemote bool
	File         string
}

type MyConfig struct {
	Logging  LoggingInfo
	Port     int
	Host     string
	LogLevel string
}

var mockConsul *MockConsul

func TestMain(m *testing.M) {

	var testMockServer *httptest.Server
	if testHost == "" || port != 8500 {
		mockConsul = NewMockConsul()
		testMockServer = mockConsul.Start()

		URL, _ := url.Parse(testMockServer.URL)
		testHost = URL.Hostname()
		port, _ = strconv.Atoi(URL.Port())
	}

	exitCode := m.Run()
	if testMockServer != nil {
		defer testMockServer.Close()
	}
	os.Exit(exitCode)
}

func TestIsAlive(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)
	if !client.IsAlive() {
		t.Fatal("Consul not running")
	}
}

func TestHasConfigurationFalse(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the configuration doesn't already exists
	reset(t, client)

	// Don't push anything in yet so configuration will not exists

	actual, err := client.HasConfiguration()
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	assert.False(t, actual)
}

func TestHasConfigurationTrue(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the configuration doesn't already exists
	reset(t, client)

	// Now push a value so the configuration will exist
	_ = client.PutConfigurationValue("Dummy", []byte("Value"))

	actual, err := client.HasConfiguration()
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	assert.True(t, actual)
}

func TestHasConfigurationPartialServiceKey(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the configuration doesn't already exists
	reset(t, client)

	base := client.configBasePath
	if strings.LastIndex(base, "/") == len(base)-1 {
		base = base[:len(base)-1]
	}
	// Add a key with similar base path
	keyPair := api.KVPair{
		Key:   base + "-test/some-key",
		Value: []byte("Nothing"),
	}
	_, err := client.consulClient.KV().Put(&keyPair, nil)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	actual, err := client.HasConfiguration()
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	assert.False(t, actual)
}

func TestHasConfigurationError(t *testing.T) {
	goodPort := port
	port = 1234 // change the Consul port to bad port
	defer func() {
		port = goodPort
	}()

	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	_, err := client.HasConfiguration()
	assert.Error(t, err, "expected error checking configuration existence")

	assert.Contains(t, err.Error(), "checking configuration existence")
}

func TestHasSubConfigurationFalse(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the configuration doesn't already exists
	reset(t, client)

	// Now push a value so some configuration will exist
	_ = client.PutConfigurationValue("Dummy", []byte("Value"))

	actual, err := client.HasSubConfiguration("subDummy")
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	assert.False(t, actual)
}

func TestHasSubConfigurationTrue(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the configuration doesn't already exists
	reset(t, client)

	// Now push a value so some configuration will exist
	_ = client.PutConfigurationValue("Dummy", []byte("Value"))

	actual, err := client.HasSubConfiguration("Dummy")
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	assert.True(t, actual)
}

func TestHasSubConfigurationError(t *testing.T) {
	goodPort := port
	port = 1234 // change the Consul port to bad port
	defer func() {
		port = goodPort
	}()

	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	_, err := client.HasSubConfiguration("dummy")
	assert.Error(t, err, "expected error checking configuration existence")

	assert.Contains(t, err.Error(), "checking sub configuration existence")
}

func TestConfigurationValueExists(t *testing.T) {
	key := "Foo"
	value := []byte("bar")
	uniqueServiceName := getUniqueServiceName()
	fullKey := consulBasePath + uniqueServiceName + "/" + key

	client := makeConsulClient(t, uniqueServiceName, "", nil)
	expected := false

	// Make sure the configuration doesn't already exists
	reset(t, client)

	actual, err := client.ConfigurationValueExists(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.False(t, actual) {
		t.Fatal()
	}

	keyPair := api.KVPair{
		Key:   fullKey,
		Value: value,
	}

	_, err = client.consulClient.KV().Put(&keyPair, nil)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	expected = true
	actual, err = client.ConfigurationValueExists(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expected, actual) {
		t.Fatal()
	}
}

func TestGetConfigurationValue(t *testing.T) {
	key := "Foo"
	expected := []byte("bar")
	uniqueServiceName := getUniqueServiceName()
	fullKey := consulBasePath + uniqueServiceName + "/" + key
	client := makeConsulClient(t, uniqueServiceName, "", nil)

	// Make sure the target key/value exists
	keyPair := api.KVPair{
		Key:   fullKey,
		Value: expected,
	}

	_, err := client.consulClient.KV().Put(&keyPair, nil)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	actual, err := client.GetConfigurationValue(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expected, actual) {
		t.Fatal()
	}
}

func TestPutConfigurationValue(t *testing.T) {
	key := "Foo"
	expected := []byte("bar")
	uniqueServiceName := getUniqueServiceName()
	expectedFullKey := consulBasePath + uniqueServiceName + "/" + key

	client := makeConsulClient(t, uniqueServiceName, "", nil)

	// Make sure the configuration doesn't already exists
	reset(t, client)

	_, _ = client.consulClient.KV().Delete(expectedFullKey, nil)

	err := client.PutConfigurationValue(key, expected)
	assert.NoError(t, err)

	keyValue, _, err := client.consulClient.KV().Get(expectedFullKey, nil)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotNil(t, keyValue, "%s value not found", expectedFullKey) {
		t.Fatal()
	}

	actual := keyValue.Value

	assert.Equal(t, expected, actual)

}

func TestGetConfiguration(t *testing.T) {
	expected := MyConfig{
		Logging: LoggingInfo{
			EnableRemote: true,
			File:         "NONE",
		},
		Port:     8000,
		Host:     "localhost",
		LogLevel: "debug",
	}

	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	_ = client.PutConfigurationValue("Logging/EnableRemote", []byte(strconv.FormatBool(expected.Logging.EnableRemote)))
	_ = client.PutConfigurationValue("Logging/File", []byte(expected.Logging.File))
	_ = client.PutConfigurationValue("Port", []byte(strconv.Itoa(expected.Port)))
	_ = client.PutConfigurationValue("Host", []byte(expected.Host))
	_ = client.PutConfigurationValue("LogLevel", []byte(expected.LogLevel))

	result, err := client.GetConfiguration(&MyConfig{})

	if !assert.NoError(t, err) {
		t.Fatal()
	}

	configuration := result.(*MyConfig)

	if !assert.NotNil(t, configuration) {
		t.Fatal()
	}

	assert.Equal(t, expected.Logging.EnableRemote, configuration.Logging.EnableRemote, "Logging.EnableRemote not as expected")
	assert.Equal(t, expected.Logging.File, configuration.Logging.File, "Logging.File not as expected")
	assert.Equal(t, expected.Port, configuration.Port, "Port not as expected")
	assert.Equal(t, expected.Host, configuration.Host, "Host not as expected")
	assert.Equal(t, expected.LogLevel, configuration.LogLevel, "LogLevel not as expected")
}

func TestPutConfiguration(t *testing.T) {
	expected := MyConfig{
		Logging: LoggingInfo{
			EnableRemote: true,
			File:         "NONE",
		},
		Port:     8000,
		Host:     "localhost",
		LogLevel: "debug",
	}

	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the tree of values doesn't exist.
	_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)

	defer func() {
		// Clean up
		_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	err := client.PutConfiguration(expected, true)
	if !assert.NoErrorf(t, err, "unable to put configuration: %v", err) {
		t.Fatal()
	}

	actual, err := client.HasConfiguration()
	require.NoError(t, err)
	if !assert.True(t, actual, "Failed to put configuration") {
		t.Fail()
	}

	assert.True(t, configValueSet("Logging/EnableRemote", client))
	assert.True(t, configValueSet("Logging/File", client))
	assert.True(t, configValueSet("Port", client))
	assert.True(t, configValueSet("Host", client))
	assert.True(t, configValueSet("LogLevel", client))
}

func configValueSet(key string, client *consulClient) bool {
	exists, _ := client.ConfigurationValueExists(key)
	return exists
}

func TestPutConfigurationTomlNoPreviousValues(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the tree of values doesn't exist.
	_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)

	defer func() {
		// Clean up
		_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	configMap := createKeyValueMap()
	configuration, err := toml.TreeFromMap(configMap)
	if err != nil {
		log.Fatalf("unable to create TOML Tree from map: %v", err)
	}
	err = client.PutConfigurationToml(configuration, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	keyValues := convertInterfaceToConsulPairs("", configMap)
	for _, keyValue := range keyValues {
		expected := keyValue.Value
		value, err := client.GetConfigurationValue(keyValue.Key)
		if !assert.NoError(t, err) {
			t.Fatal()
		}
		actual := string(value)
		if !assert.Equal(t, expected, actual, "Values for %s are not equal", keyValue.Key) {
			t.Fatal()
		}
	}
}

func TestPutConfigurationTomlWithoutOverWrite(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the tree of values doesn't exist.
	_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)

	defer func() {
		// Clean up
		_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	configMap := createKeyValueMap()

	configuration, _ := toml.TreeFromMap(configMap)
	err := client.PutConfigurationToml(configuration, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	//Update map with new value and try to overwrite it
	configMap["int"] = 2
	configMap["int64"] = 164
	configMap["float64"] = 2.4
	configMap["string"] = "bye"
	configMap["bool"] = false

	// Try to put new values with overwrite = false
	configuration, _ = toml.TreeFromMap(configMap)
	err = client.PutConfigurationToml(configuration, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	keyValues := convertInterfaceToConsulPairs("", configMap)
	for _, keyValue := range keyValues {
		expected := keyValue.Value
		value, err := client.GetConfigurationValue(keyValue.Key)
		if !assert.NoError(t, err) {
			t.Fatal()
		}
		actual := string(value)
		if !assert.NotEqual(t, expected, actual, "Values for %s are equal, expected not equal", keyValue.Key) {
			t.Fatal()
		}
	}
}

func TestPutConfigurationTomlOverWrite(t *testing.T) {
	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the tree of values doesn't exist.
	_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)
	// Clean up after unit test
	defer func() {
		_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	configMap := createKeyValueMap()

	configuration, _ := toml.TreeFromMap(configMap)
	err := client.PutConfigurationToml(configuration, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	//Update map with new value and try to overwrite it
	configMap["int"] = 2
	configMap["float64"] = 2.4
	configMap["string"] = "bye"
	configMap["bool"] = false

	// Try to put new values with overwrite = True
	configuration, _ = toml.TreeFromMap(configMap)
	err = client.PutConfigurationToml(configuration, true)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	keyValues := convertInterfaceToConsulPairs("", configMap)
	for _, keyValue := range keyValues {
		expected := keyValue.Value
		value, err := client.GetConfigurationValue(keyValue.Key)
		if !assert.NoError(t, err) {
			t.Fatal()
		}
		actual := string(value)
		if !assert.Equal(t, expected, actual, "Values for %s are not equal", keyValue.Key) {
			t.Fatal()
		}
	}
}

func TestWatchForChanges(t *testing.T) {
	expectedConfig := MyConfig{
		Logging: LoggingInfo{
			EnableRemote: true,
			File:         "NONE",
		},
		Port:     8000,
		Host:     "localhost",
		LogLevel: "debug",
	}

	expectedChange := "random"

	client := makeConsulClient(t, getUniqueServiceName(), "", nil)

	// Make sure the tree of values doesn't exist.
	_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)
	// Clean up after unit test
	defer func() {
		_, _ = client.consulClient.KV().DeleteTree(consulBasePath, nil)
	}()

	_ = client.PutConfigurationValue("Logging/EnableRemote", []byte(strconv.FormatBool(expectedConfig.Logging.EnableRemote)))
	_ = client.PutConfigurationValue("Logging/File", []byte(expectedConfig.Logging.File))
	_ = client.PutConfigurationValue("Port", []byte(strconv.Itoa(expectedConfig.Port)))
	_ = client.PutConfigurationValue("Host", []byte(expectedConfig.Host))
	_ = client.PutConfigurationValue("LogLevel", []byte(expectedConfig.LogLevel))

	loggingUpdateChannel := make(chan interface{})
	errorChannel := make(chan error)

	client.WatchForChanges(loggingUpdateChannel, errorChannel, &LoggingInfo{}, "Logging")

	loggingPass := 1

	for {
		select {
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting on Logging configuration loggingChanges")

		case loggingChanges := <-loggingUpdateChannel:
			assert.NotNil(t, loggingChanges)
			logInfo := loggingChanges.(*LoggingInfo)

			// first pass is for Consul Decoder always sending data once watch has been setup. It hasn't actually changed
			if loggingPass == 1 {
				if !assert.Equal(t, logInfo.File, expectedConfig.Logging.File) {
					t.Fatal()
				}

				// Make a change to logging
				_ = client.PutConfigurationValue("Logging/File", []byte(expectedChange))

				loggingPass--
				continue
			}

			// Now the data should have changed
			assert.Equal(t, logInfo.File, expectedChange)
			return

		case waitError := <-errorChannel:
			t.Fatalf("received WatchForChanges error for Logging: %v", waitError)
		}
	}
}

func TestAccessToken(t *testing.T) {
	uniqueServiceName := getUniqueServiceName()
	client := makeConsulClient(t, uniqueServiceName, "", nil)
	expectedErrMsg := "Unexpected response code: 403"
	valueName := "testAccess"
	// Test if have access to endpoint w/o access token set

	_, err := client.GetConfigurationValue(valueName)
	require.NoError(t, err)

	expectedToken := "MyAccessToken"
	mockConsul.SetExpectedAccessToken(expectedToken)
	defer mockConsul.ClearExpectedAccessToken()

	// Now verify get error w/o providing the expected access token
	_, err = client.GetConfigurationValue(valueName)
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedErrMsg)
}

func makeConsulClient(t *testing.T, serviceName string, accessToken string, tokenCallback types.GetAccessTokenCallback) *consulClient {
	config := types.ServiceConfig{
		Host:           testHost,
		Port:           port,
		BasePath:       "edgex/core/1.0/" + serviceName,
		AccessToken:    accessToken,
		GetAccessToken: tokenCallback,
	}

	client, err := NewConsulClient(config)
	if assert.NoError(t, err) == false {
		t.Fatal()
	}

	return client
}

func createKeyValueMap() map[string]interface{} {
	configMap := make(map[string]interface{})

	configMap["int"] = 1
	configMap["int64"] = int64(64)
	configMap["float64"] = float64(1.4)
	configMap["string"] = "hello"
	configMap["bool"] = true

	return configMap
}

func reset(t *testing.T, client *consulClient) {
	// Make sure the configuration doesn't already exists
	if mockConsul != nil {
		mockConsul.Reset()
	} else {
		key := client.configBasePath
		if strings.LastIndex(key, "/") == len(key)-1 {
			key = key[:len(key)-1]
		}

		_, err := client.consulClient.KV().Delete(key, nil)
		if !assert.NoError(t, err) {
			t.Fatal()
		}
	}
}

func getUniqueServiceName() string {
	return serviceName + strconv.Itoa(time.Now().Nanosecond())
}

func TestRenewAccessToken(t *testing.T) {
	goodToken := "bfb78dc5-c6a3-33d9-88b5-e3a4b63dda77" // nolint: gosec
	badToken := "badToken-c6a3-33d9-88b5-e3a4b63dda77"  // nolint: gosec
	serviceName := "RenewAccessToken-Test"

	getAccessToken := func() (string, error) {
		fmt.Println("RenewAccessToken called")
		return goodToken, nil
	}

	createClient := func(resetConfig bool) *consulClient {
		client := makeConsulClient(t, serviceName, badToken, getAccessToken)
		if resetConfig {
			reset(t, client)
		}
		mockConsul.SetExpectedAccessToken(goodToken)
		return client
	}

	myConfig := MyConfig{
		Logging: LoggingInfo{
			EnableRemote: false,
			File:         "",
		},
		Port:     678,
		Host:     "Local",
		LogLevel: "Infinity",
	}

	putTestConfig := func() {
		// Put the configuration so we can test the GET
		client := makeConsulClient(t, serviceName, "", nil)
		mockConsul.SetExpectedAccessToken("")
		err := client.PutConfiguration(&myConfig, true)
		require.NoError(t, err)
	}

	t.Run("PutConfigurationValue", func(t *testing.T) {
		client := createClient(true)

		err := client.PutConfigurationValue("Host", []byte("Hello"))
		require.NoError(t, err)
	})

	t.Run("PutConfiguration", func(t *testing.T) {
		client := createClient(true)

		err := client.PutConfiguration(&myConfig, true)
		require.NoError(t, err)
	})

	t.Run("ConfigurationValueExists", func(t *testing.T) {
		client := createClient(true)

		_, err := client.ConfigurationValueExists("Host")
		require.NoError(t, err)
	})

	t.Run("HasConfiguration", func(t *testing.T) {
		client := createClient(true)

		_, err := client.HasConfiguration()
		require.NoError(t, err)
	})

	t.Run("GetConfiguration", func(t *testing.T) {
		putTestConfig()
		client := createClient(false)

		_, err := client.GetConfiguration(&myConfig)
		require.NoError(t, err)
	})

	t.Run("GetConfigurationValue", func(t *testing.T) {
		putTestConfig()
		client := createClient(false)

		_, err := client.GetConfigurationValue("Host")
		require.NoError(t, err)
	})

	t.Run("HasSubConfiguration", func(t *testing.T) {
		client := createClient(true)

		_, err := client.HasSubConfiguration("Logging")
		require.NoError(t, err)
	})

	t.Run("WatchForChanges", func(t *testing.T) {
		putTestConfig()
		client := createClient(false)

		receivedUpdate := false

		updates := make(chan interface{})
		errs := make(chan error)
		client.WatchForChanges(updates, errs, &myConfig, "Logging")

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for {
				select {
				case <-time.Tick(5 * time.Second):
					wg.Done()
					return

				case ex := <-errs:
					require.NoError(t, ex)

				case raw, ok := <-updates:
					if !ok {
						return
					}
					require.NotNil(t, raw)
					wg.Done()
					receivedUpdate = true
					fmt.Println("WatchForChanges update received")
					return
				}
			}
		}()

		// simulate change in writable section.
		err := client.PutConfigurationValue("LogLevel", []byte("DEBUG"))
		require.NoError(t, err)

		wg.Wait()

		assert.True(t, receivedUpdate)
	})

	t.Run("StopWatching", func(t *testing.T) {
		putTestConfig()
		client := createClient(false)

		allStopped := false
		updates := make(chan interface{})
		errs := make(chan error)
		client.WatchForChanges(updates, errs, &myConfig, "Host")
		client.WatchForChanges(updates, errs, &myConfig, "LogLevel")

		go func() {
			client.StopWatching()
			allStopped = true
		}()

		<-time.Tick(2 * time.Second)
		assert.True(t, allStopped)
	})
}
