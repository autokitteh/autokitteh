package kittehs

import (
	"net"
)

// BindingAddress returns an unspecified address with the given port.
// This is needed in order to expose AutoKitteh ports in Docker containers,
// because using "localhost" or "127.0.0.1" will bind the port to the loopback
// interface, making it inaccessible from outside the container.
func BindingAddress(port string) string {
	return net.JoinHostPort("0.0.0.0", port)
}

// DisplayAddress returns a human-readable address for the given binding address.
// If the host is unspecified, it will be replaced with "localhost". Motivation:
// https://www.oligo.security/blog/0-0-0-0-day-exploiting-localhost-apis-from-the-browser
func DisplayAddress(bindingAddress string) string {
	host, port, err := net.SplitHostPort(bindingAddress)
	if err != nil {
		return bindingAddress
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		return net.JoinHostPort("localhost", port)
	}

	return bindingAddress
}
