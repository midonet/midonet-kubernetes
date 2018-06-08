// Copyright 2015 Tigera Inc
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

package utils

import (
	"fmt"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// CleanUpNamespace deletes the devices in the network namespace.
func CleanUpNamespace(args *skel.CmdArgs, logger *logrus.Entry) error {
	// Only try to delete the device if a namespace was passed in.
	if args.Netns != "" {
		logger.WithFields(logrus.Fields{
			"netns": args.Netns,
			"iface": args.IfName,
		}).Debug("Checking namespace & device exist.")
		devErr := ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
			_, err := netlink.LinkByName(args.IfName)
			return err
		})

		if devErr == nil {
			fmt.Fprintf(os.Stderr, "MidoNet CNI deleting device in netns %s\n", args.Netns)
			err := ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
				_, err := ip.DelLinkByNameAddr(args.IfName)
				return err
			})

			if err != nil {
				return err
			}
		} else {
			logger.WithField("ifName", args.IfName).Info("veth does not exist, no need to clean up.")
		}
	}

	return nil
}
