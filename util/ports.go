package util

import (
	"fmt"
	"strconv"
	"strings"
)

// PortAndProtol adds the protocol tag '/tcp' to the bare port number
// if it's missing.
func PortAndProtocol(port string) string {
	if len(strings.Split(port, "/")) == 1 {
		port += "/tcp"
	}
	return port
}

// PortComponents splits the port element from the definition file
// into 3 components: an IP address (if any), the published port number,
// and the exposed port number. The protocol tag is added to the exposed
// port if it's missing (“8080” -> “8080/tcp”).
//
// Expected inputs:
//
//           published
//           published:exposed
//        IP:published:exposed
//
func PortComponents(port string) (ip, exposed, published string) {
	if port == "" {
		return "", "", ""
	}

	components := strings.Split(port, ":")

	if strings.Count(port, ":") == 2 {
		return components[0], components[1], PortAndProtocol(components[2])
	}
	if strings.Count(port, ":") == 1 {
		return "", components[0], PortAndProtocol(components[1])
	}
	return "", port, PortAndProtocol(port)
}

// MapPorts reassigns ports mappings (usually taken from the chain or service
// definition file) according to a list. MapPorts returns a map with keys
// - exposed (container) ports and values - published ports (ports accessible
// on the host network).
//
// Reassignment rules:
//   <n>      - map the published port <n> to the container port in `ports`
//              at this current position (slice-wise)
//   <n>:<m>  - map the published port <n> to container port <m> in `ports`
//   <n>-     - map the published port <n> to the exposed port in ports
//              at this current position (slice-wise), and use the consecutive
//              port numbers for the remaining ports in `ports`.
//
func MapPorts(ports, assignments []string) map[string]string {
	m := make(map[string]string)

	// First pass. Default mapping from the definition file.
	for _, entry := range ports {
		_, published, exposed := PortComponents(entry)

		m[exposed] = published
	}

	// Second pass. Reassign ports.
	var (
		autoincrement     bool
		autoincrementPort int
	)
	for i, entry := range ports {
		_, _, exposed := PortComponents(entry)
		if i < len(assignments) {
			switch {
			// Explicit mapping, e.g. “9001:9009”.
			case strings.Contains(assignments[i], ":"):
				_, published, exposed := PortComponents(assignments[i])
				m[exposed] = published

			// Autoincrement mapping, e.g. “9001-”.
			case strings.HasSuffix(assignments[i], "-"):
				rangeStart := strings.TrimRight(assignments[i], "-")

				var err error
				autoincrementPort, err = strconv.Atoi(rangeStart)
				if err != nil {
					continue
				}

				autoincrement = true

				m[exposed] = rangeStart

			// Simple reassignment.
			default:
				m[exposed] = assignments[i]
			}
		} else if autoincrement {
			autoincrementPort += 1

			m[exposed] = fmt.Sprintf("%d", autoincrementPort)
		}
	}

	return m
}
