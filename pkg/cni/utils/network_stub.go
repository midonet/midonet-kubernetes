// Copyright 2018 Tigera Inc
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

// +build !linux

package utils

import (
	"net"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/sirupsen/logrus"
)

// DoNetworking performs the networking for the given config and IPAM result
func DoNetworking(destNetworks []*net.IPNet, ips []*current.IPConfig, contNetNS, contVethName, hostVethName string, ipForward bool, logger *logrus.Entry) (contVethMAC string, err error) {
	logrus.Fatal("Stub implementation used")
	return "", nil
}
