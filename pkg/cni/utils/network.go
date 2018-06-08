// +build linux

package utils

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// DoNetworking performs the networking for the given config and IPAM result
func DoNetworking(destNetworks []*net.IPNet, ips []*current.IPConfig, contNetNS, contVethName, hostVethName string, ipForward bool, logger *logrus.Entry) (contVethMAC string, err error) {
	var hasIPv4, hasIPv6 bool
	MTU := 1500 // XXX

	logger.Infof("Setting the host side veth name to %s", hostVethName)

	// Clean up if hostVeth exists.
	if oldHostVeth, err := netlink.LinkByName(hostVethName); err == nil {
		if err = netlink.LinkDel(oldHostVeth); err != nil {
			return "", fmt.Errorf("failed to delete old hostVeth %v: %v", hostVethName, err)
		}
		logger.Infof("Cleaning old hostVeth: %v", hostVethName)
	}

	err = ns.WithNetNSPath(contNetNS, func(hostNS ns.NetNS) error {
		veth := &netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{
				Name:  contVethName,
				Flags: net.FlagUp,
				MTU:   MTU,
			},
			PeerName: hostVethName,
		}

		if err := netlink.LinkAdd(veth); err != nil {
			logger.Errorf("Error adding veth %+v: %s", veth, err)
			return err
		}

		hostVeth, err := netlink.LinkByName(hostVethName)
		if err != nil {
			err = fmt.Errorf("failed to lookup %q: %v", hostVethName, err)
			return err
		}

		if mac, err := net.ParseMAC("EE:EE:EE:EE:EE:EE"); err != nil {
			logger.Infof("failed to parse MAC Address: %v. Using kernel generated MAC.", err)
		} else {
			// Set the MAC address on the host side interface so the kernel does not
			// have to generate a persistent address which fails some times.
			if err = netlink.LinkSetHardwareAddr(hostVeth, mac); err != nil {
				logger.Warnf("failed to Set MAC of %q: %v. Using kernel generated MAC.", hostVethName, err)
			}
		}

		// Explicitly set the veth to UP state, because netlink doesn't always do that on all the platforms with net.FlagUp.
		// veth won't get a link local address unless it's set to UP state.
		if err = netlink.LinkSetUp(hostVeth); err != nil {
			return fmt.Errorf("failed to set %q up: %v", hostVethName, err)
		}

		contVeth, err := netlink.LinkByName(contVethName)
		if err != nil {
			err = fmt.Errorf("failed to lookup %q: %v", contVethName, err)
			return err
		}

		// Fetch the MAC from the container Veth. This is needed by Calico.
		contVethMAC = contVeth.Attrs().HardwareAddr.String()
		logger.WithField("MAC", contVethMAC).Debug("Found MAC for container veth")

		// At this point, the virtual ethernet pair has been created, and both ends have the right names.
		// Both ends of the veth are still in the container's network namespace.

		for _, addr := range ips {

			// Before returning, create the routes inside the namespace, first for IPv4 then IPv6.
			if addr.Version == "4" {
				if err = netlink.AddrAdd(contVeth, &netlink.Addr{IPNet: &addr.Address}); err != nil {
					return fmt.Errorf("failed to add IP addr to %q: %v", contVethName, err)
				}

				for _, dest := range destNetworks {
					gw := addr.Gateway
					if err = ip.AddRoute(dest, gw, contVeth); err != nil {
						return fmt.Errorf("failed to add the default route inside the container: %v", err)
					}
				}

				// Set hasIPv4 to true so sysctls for IPv4 can be programmed when the host side of
				// the veth finishes moving to the host namespace.
				hasIPv4 = true
			}
		}

		if err = ConfigureIPForwarding(hasIPv4, hasIPv6, ipForward); err != nil {
			return fmt.Errorf("error configuring sysctls for the container netns, error: %s", err)
		}

		// Now that the everything has been successfully set up in the container, move the "host" end of the
		// veth into the host namespace.
		if err = netlink.LinkSetNsFd(hostVeth, int(hostNS.Fd())); err != nil {
			return fmt.Errorf("failed to move veth to host netns: %v", err)
		}

		return nil
	})

	if err != nil {
		logger.Errorf("Error creating veth: %s", err)
		return "", err
	}

	// Moving a veth between namespaces always leaves it in the "DOWN" state. Set it back to "UP" now that we're
	// back in the host namespace.
	hostVeth, err := netlink.LinkByName(hostVethName)
	if err != nil {
		return "", fmt.Errorf("failed to lookup %q: %v", hostVethName, err)
	}

	if err = netlink.LinkSetUp(hostVeth); err != nil {
		return "", fmt.Errorf("failed to set %q up: %v", hostVethName, err)
	}

	return contVethMAC, err
}

func ConfigureIPForwarding(hasIPv4, hasIPv6, enable bool) error {
	var err error

	value := "0"
	if enable {
		value = "1"
	}
	if hasIPv4 {
		if err = writeProcSys("/proc/sys/net/ipv4/ip_forward", value); err != nil {
			return err
		}
	}

	if hasIPv6 {
		if err = writeProcSys("/proc/sys/net/ipv6/conf/all/forwarding", value); err != nil {
			return err
		}
	}

	return nil
}

// writeProcSys takes the sysctl path and a string value to set i.e. "0" or "1" and sets the sysctl.
func writeProcSys(path, value string) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	n, err := f.Write([]byte(value))
	if err == nil && n < len(value) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
