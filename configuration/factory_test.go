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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/go-mod-configuration/v2/pkg/types"
)

var config = types.ServiceConfig{
	Host:     "localhost",
	Port:     8500,
	BasePath: "config",
}

func TestNewClientConsul(t *testing.T) {

	config.Type = "consul"

	client, err := NewConfigurationClient(config)
	if assert.Nil(t, err, "New Configuration client failed: ", err) == false {
		t.Fatal()
	}

	assert.False(t, client.IsAlive(), "Consul service not expected be running")
}

func TestNewClientBogusType(t *testing.T) {

	config.Type = "bogus"

	_, err := NewConfigurationClient(config)
	if assert.NotNil(t, err, "Expected configuration type type error") == false {
		t.Fatal()
	}
}
