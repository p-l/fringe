package system

import "net"

// AllLocalIPAddresses returns all available interfaces' IP.
func AllLocalIPAddresses() []net.IP {
	var ips []net.IP

	ifs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}

	for _, address := range ifs {
		// check the address type and if it is not a loopback the display it
		if network, ok := address.(*net.IPNet); ok {
			ips = append(ips, network.IP)
		}
	}

	// return localhost if nothing better is found
	return ips
}

// FirstLocalIP Returns first none loopback IP available. Returns net.IPv6loopback if nothing is found.
func FirstLocalIP(ips []net.IP) net.IP {
	for _, ip := range ips {
		if !ip.IsLoopback() {
			return ip
		}
	}

	return net.IPv6loopback
}
