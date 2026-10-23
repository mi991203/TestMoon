package platform

import "net"

// localGRPCEndpoint connects the HTTP gateway to its own gRPC listener. A
// wildcard is a bind address, not a remote endpoint (and may hit a proxy).
func localGRPCEndpoint(address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}
	switch host {
	case "", "0.0.0.0":
		return net.JoinHostPort("127.0.0.1", port)
	case "::":
		return net.JoinHostPort("::1", port)
	default:
		return address
	}
}
