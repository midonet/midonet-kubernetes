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
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/midonet/midonet-kubernetes/pkg/config"
)

// REVISIT: maybe separate to ClientConfig and ConverterConfig

type Config struct {
	// Client
	api      string
	username string
	password string
	project  string

	// Converter
	ClusterRouter uuid.UUID
	Tenant        string
}

func NewConfigFromEnvConfig(config *config.Config) *Config {
	router, err := uuid.Parse(config.ClusterRouter)
	if err != nil {
		log.WithError(err).Fatal("Failed to parse cluster router")
	}
	return &Config{
		api:      config.MidoNetAPI,
		username: config.MidoNetUserName,
		password: config.MidoNetPassword,
		project:  config.MidoNetProject,

		ClusterRouter: router,
		Tenant:        config.Tenant,
	}
}
