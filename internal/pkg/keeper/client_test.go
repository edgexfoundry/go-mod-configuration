//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-configuration/v4/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceName = "coreKeeperUnitTest"
	dummyConfig = "dummy"
)

// change values to localhost and 59883 if you need to run tests against real Core Keeper service running locally
var (
	testHost = ""
	port     = 0
)

type LoggingInfo struct {
	EnableRemote bool
	File         string
}

type TestConfig struct {
	Logging  LoggingInfo
	Port     int
	Host     string
	LogLevel string
	Temp     float64
}

func makeCoreKeeperClient(serviceName string) *keeperClient {
	config := types.ServiceConfig{
		Host:         testHost,
		Port:         port,
		BasePath:     serviceName,
		AuthInjector: NewNullAuthenticationInjector(),
	}

	client := NewKeeperClient(config)
	return client
}

func getUniqueServiceName() string {
	return serviceName + strconv.Itoa(time.Now().Nanosecond())
}

func configValueExists(key string, client *keeperClient) bool {
	exists, _ := client.ConfigurationValueExists(key)
	return exists
}

var mockCoreKeeper *MockCoreKeeper

func reset(t *testing.T, client *keeperClient) {
	// Make sure the configuration not exists
	if mockCoreKeeper != nil {
		mockCoreKeeper.Reset()
	} else {
		// delete the key(s) created in each test if testing on real Keeper service
		key := client.configBasePath
		_, err := client.kvsClient.DeleteKeysByPrefix(context.Background(), key)
		if !assert.NoError(t, err) {
			t.Fatal()
		}
	}
}

func TestMain(m *testing.M) {
	var testMockServer *httptest.Server
	if testHost == "" || port != 59883 {
		mockCoreKeeper = NewMockCoreKeeper()
		testMockServer = mockCoreKeeper.Start()

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
	client := makeCoreKeeperClient(getUniqueServiceName())
	if !client.IsAlive() {
		t.Fatal("Core Keeper is not running")
	}
}

func TestHasConfigurationFalse(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	actual, err := client.HasConfiguration()
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	assert.False(t, actual)
}

func TestHasConfigurationTrue(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	err := client.PutConfiguration(dummyConfig, true)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	actual, err := client.HasConfiguration()
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	assert.True(t, actual)
}

func TestHasSubConfigurationFalse(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	actual, err := client.HasSubConfiguration(dummyConfig)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	assert.False(t, actual)
}

func TestHasSubConfigurationTrue(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	_ = client.PutConfigurationValue(dummyConfig, []byte(dummyConfig))

	actual, err := client.HasSubConfiguration(dummyConfig)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	assert.True(t, actual)
}

func createConfigMap() map[string]interface{} {
	configMap := make(map[string]interface{})

	configMap["int"] = 1
	configMap["int64"] = int64(64)
	configMap["float64"] = float64(1.4)
	configMap["string"] = "hello"
	configMap["bool"] = true
	configMap["nestedNode"] = map[string]interface{}{"field1": "value1", "field2": "value2"}

	return configMap
}

func TestPutConfigurationMapNoPreValues(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	configMap := createConfigMap()
	err := client.PutConfigurationMap(configMap, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
}

func TestPutConfigurationMapWithoutOverwrite(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	configMap := createConfigMap()
	err := client.PutConfigurationMap(configMap, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	// the initial config value before updating
	expected, err := client.GetConfigurationValue("nestedNode/field1")
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	// overwrite the configMap fields
	configMap["nestedNode"] = map[string]interface{}{"field1": "overwrite1", "field2": "overwrite2"}
	err = client.PutConfigurationMap(configMap, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	actual, err := client.GetConfigurationValue("nestedNode/field1")
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expected, actual, "Values for %s are not equal, expected equal", "nestedNode/field1") {
		t.Fatal()
	}
}

func TestPutConfigurationMapWithOverwrite(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	configMap := createConfigMap()
	err := client.PutConfigurationMap(configMap, false)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	// the initial config value before updating
	expected, err := client.GetConfigurationValue("nestedNode/field1")
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	// overwrite the configMap fields
	configMap["nestedNode"] = map[string]interface{}{"field1": "overwrite1", "field2": "overwrite2"}
	err = client.PutConfigurationMap(configMap, true)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	actual, err := client.GetConfigurationValue("nestedNode/field1")
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotEqual(t, expected, actual, "Values for %s are equal, expected not equal", "nestedNode/field1") {
		t.Fatal()
	}
}

func TestPutConfiguration(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	expected := TestConfig{
		Logging: LoggingInfo{
			EnableRemote: true,
			File:         "NONE",
		},
		Port:     8000,
		Host:     "localhost",
		LogLevel: "debug",
		Temp:     36.123456,
	}
	err := client.PutConfiguration(expected, true)
	if !assert.NoErrorf(t, err, "unable to put configuration: %v", err) {
		t.Fatal()
	}

	actual, err := client.HasConfiguration()
	require.NoError(t, err)
	if !assert.True(t, actual, "Failed to put configuration") {
		t.Fail()
	}

	assert.True(t, configValueExists("Logging/EnableRemote", client))
	assert.True(t, configValueExists("Logging/File", client))
	assert.True(t, configValueExists("Port", client))
	assert.True(t, configValueExists("Host", client))
	assert.True(t, configValueExists("LogLevel", client))
	assert.True(t, configValueExists("Temp", client))
}

func TestGetConfiguration(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	expected := TestConfig{
		Logging: LoggingInfo{
			EnableRemote: true,
			File:         "NONE",
		},
		Port:     8000,
		Host:     "localhost",
		LogLevel: "debug",
		Temp:     36.123456,
	}

	err := client.PutConfiguration(expected, true)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	result, err := client.GetConfiguration(&TestConfig{})
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	actual := (result).(*TestConfig)

	if !assert.NotNil(t, expected) {
		t.Fatal()
	}

	assert.Equal(t, expected.Logging.EnableRemote, actual.Logging.EnableRemote, "Logging.EnableRemote not as expected")
	assert.Equal(t, expected.Logging.File, actual.Logging.File, "Logging.File not as expected")
	assert.Equal(t, expected.Port, actual.Port, "Port not as expected")
	assert.Equal(t, expected.Host, actual.Host, "Host not as expected")
	assert.Equal(t, expected.LogLevel, actual.LogLevel, "LogLevel not as expected")
	assert.Equal(t, expected.Temp, actual.Temp, "Temp not as expected")
}

func TestConfigurationValueExists(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	key := "Foo"
	value := []byte("bar")

	// verify the config value not exists initially
	actual, err := client.ConfigurationValueExists(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.False(t, actual) {
		t.Fatal()
	}

	err = client.PutConfigurationValue(key, value)
	assert.NoError(t, err)

	actual, err = client.ConfigurationValueExists(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.True(t, actual) {
		t.Fatal()
	}
}

func TestGetConfigurationValue(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	key := "Foo"
	expected := []byte("bar")
	err := client.PutConfigurationValue(key, expected)
	assert.NoError(t, err)

	actual, err := client.GetConfigurationValue(key)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.Equal(t, expected, actual) {
		t.Fatal()
	}
}

func TestPutConfigurationValue(t *testing.T) {
	client := makeCoreKeeperClient(getUniqueServiceName())

	// delete the configuration created
	defer reset(t, client)

	key := "Foo"
	expected := []byte("bar")
	err := client.PutConfigurationValue(key, expected)
	assert.NoError(t, err)

	resp, err := client.kvsClient.ValuesByKey(context.Background(), client.fullPath(key))
	if !assert.NoError(t, err) {
		t.Fatal()
	}
	if !assert.NotNil(t, resp, "%s value not found", key) {
		t.Fatal()
	}

	actual := []byte(resp.Response[0].Value.(string))

	assert.Equal(t, expected, actual)
}
