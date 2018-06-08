// +build !linux

package utils

import (
	"net"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/sirupsen/logrus"
)

func DoNetworking(destNetworks []*net.IPNet, ips []*current.IPConfig, contNetNS, contVethName, hostVethName string, ipForward bool, logger *logrus.Entry) (contVethMAC string, err error) {
	logrus.Fatal("Stub implementation used")
	return "", nil
}
