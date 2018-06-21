// Copyright (C) 2018 Midokura SARL.
// All rights reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package midonet

import (
	"github.com/midonet/midonet-kubernetes/pkg/config"
)

// Config contains MidoNet API configuration.
type Config struct {
	api      string
	username string
	password string
	project  string
}

// NewConfigFromEnvConfig creates Config from envconfig instance.
func NewConfigFromEnvConfig(config *config.Config) *Config {
	return &Config{
		api:      config.MidoNetAPI,
		username: config.MidoNetUserName,
		password: config.MidoNetPassword,
		project:  config.MidoNetProject,
	}
}
