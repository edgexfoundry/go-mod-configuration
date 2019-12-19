//
// Copyright (c) 2019 Intel Corporation
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

package configuration

import (
	"github.com/pelletier/go-toml"
)

type Client interface {
	// Checks to see if the Configuration service contains the service's configuration.
	HasConfiguration() (bool, error)

	// Puts a full toml configuration into the Configuration service
	PutConfigurationToml(configuration *toml.Tree, overwrite bool) error

	// Puts a full configuration struct into the Configuration service
	PutConfiguration(configStruct interface{}, overwrite bool) error

	// Gets the full configuration from Consul into the target configuration struct.
	// Passed in struct is only a reference for Configuration service. Empty struct is fine
	// Returns the configuration in the target struct as interface{}, which caller must cast
	GetConfiguration(configStruct interface{}) (interface{}, error)

	// Sets up a Consul watch for the target key and send back updates on the update channel.
	// Passed in struct is only a reference for Configuration service, empty struct is ok
	// Sends the configuration in the target struct as interface{} on updateChannel, which caller must cast
	WatchForChanges(updateChannel chan<- interface{}, errorChannel chan<- error, configuration interface{}, waitKey string)

	// Simply checks if Configuration service is up and running at the configured URL
	IsAlive() bool

	// Checks if a configuration value exists in the Configuration service
	ConfigurationValueExists(name string) (bool, error)

	// Gets a specific configuration value from the Configuration service
	GetConfigurationValue(name string) ([]byte, error)

	// Puts a specific configuration value into the Configuration service
	PutConfigurationValue(name string, value []byte) error
}
