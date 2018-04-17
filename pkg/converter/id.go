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

package converter

import (
	"crypto/sha256"
	"net"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	kubernetesSpaceUUIDString    = "CAC60164-F74C-404A-AB39-3C1320124A17"
	midonetTenantSpaceUUIDString = "3978567E-91C4-465C-A0D1-67575F6B4C7F"
)

func idForString(spaceStr string, key string) uuid.UUID {
	space, err := uuid.Parse(spaceStr)
	if err != nil {
		log.WithError(err).Fatal("space")
	}
	return uuid.NewSHA1(space, []byte(key))
}

func IDForTenant(tenant string) uuid.UUID {
	return idForString(midonetTenantSpaceUUIDString, tenant)
}

func IDForKey(key string) uuid.UUID {
	return idForString(kubernetesSpaceUUIDString, key)
}

func SubID(id uuid.UUID, s string) uuid.UUID {
	return uuid.NewSHA1(id, []byte(s))
}

func MACForKey(key string) net.HardwareAddr {
	hash := sha256.Sum256([]byte(key))
	// AC-CA-BA  Midokura Co., Ltd.
	addr := [6]byte{0xac, 0xca, 0xba, hash[0], hash[1], hash[2]}
	return net.HardwareAddr(addr[:])
}
