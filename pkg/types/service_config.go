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

package types

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ServiceConfig defines the information need to connect to the Configuration service and optionally register the service
// for discovery and health checks
type ServiceConfig struct {
	// The Protocol that should be used to connect to the Configuration service. HTTP is used if not set.
	Protocol string
	// Host is the hostname or IP address of the Configuration service
	Host string
	// Port is the HTTP port of the Configuration service
	Port int
	// Type is the implementation type of the Configuration service, i.e. consul
	Type string
	// BasePath is the base path with in the Configuration service where the your service's configuration is stored
	BasePath string
}

//
// A few helper functions for building URLs.
//

func (config ServiceConfig) GetUrl() string {
	return fmt.Sprintf("%s://%s:%v", config.GetProtocol(), config.Host, config.Port)
}

func (config *ServiceConfig) GetProtocol() string {
	if config.Protocol == "" {
		return "http"
	}

	return config.Protocol
}

func (config *ServiceConfig) PopulateFromUrl(providerUrl string) error {
	url, err := url.Parse(providerUrl)
	if err != nil {
		return fmt.Errorf("the format of Configuration Provider path from argument is wrong: %s", err.Error())
	}

	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return fmt.Errorf("the port format of Configuration Provider path from argument is wrong: %s", err.Error())
	}

	typeAndProtocol := strings.Split(url.Scheme, ".")
	if len(typeAndProtocol) != 2 {
		return fmt.Errorf("the Type and Protocol spec of Configuration Provider path from argument is wrong: %s", err.Error())
	}

	config.Protocol = typeAndProtocol[1]
	config.Host = url.Hostname()
	config.Port = port
	config.Type = typeAndProtocol[0]

	return nil
}
