// Copyright (c) 2017 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Minimum log level to emit.
	LogLevel string `default:"info" split_words:"true"`

	// Which controllers to run.
	EnabledControllers string `default:"node,pod,service,endpoints,pusher,hostresolver" split_words:"true"`

	// Path to a kubeconfig file to use for accessing the k8s API.
	Kubeconfig string `default:"" split_words:"false"`

	// MidoNet API URL and credential
	MidoNetAPI      string `envconfig:"midonet_api" default:"https://localhost:8181/midonet-api"`
	MidoNetUserName string `envconfig:"midonet_username" default:"admin"`
	MidoNetPassword string `envconfig:"midonet_password" default:""`
	MidoNetProject  string `envconfig:"midonet_project" default:""`

	// MidoNet tenantId to group resources maintained by our controllers
	Tenant string `default:"midonetkube"`
}

// Parse parses envconfig and stores in Config struct
func (c *Config) Parse() error {
	return envconfig.Process("midonetkube", c)
}
