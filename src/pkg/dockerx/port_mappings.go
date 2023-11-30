package dockerx

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// MakeSystemAssignedPortMap creates a port map for the given list
// of nat Ports, using system assigned ports for the host side
// of the port binding.
func MakeSystemAssignedPortMap(ports ...nat.Port) nat.PortMap {
	pm := make(nat.PortMap, len(ports))
	for _, p := range ports {
		pm[p] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: "0", // Use a system-assigned port on the host
			},
		}
	}
	return pm
}

// MustMakeNATPort creates a new nat.Port, panicking rather than
// returning an error if the port is malformed. Useful for constants
// and tests.
func MustMakeNATPort(protocol, portRange string) nat.Port {
	port, err := nat.NewPort(protocol, portRange)
	if err != nil {
		panic(err)
	}

	return port
}

// ContainerPortToHostPort finds the host port to which the given
// container port has been bound. Useful when starting
// test containers and allowing them to choose their
// own host port.
func ContainerPortToHostPort(
	ctx context.Context,
	cli client.APIClient,
	containerID string,
	containerPort nat.Port,
) (int, error) {
	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return 0, fmt.Errorf("unable to inspect container %s: %w", containerID, err)
	}

	if len(containerJSON.NetworkSettings.Ports) == 0 {
		return 0, fmt.Errorf("no exposed ports for %s", containerID)
	}

	hostPortBindings, ok := containerJSON.NetworkSettings.Ports[containerPort]
	if !ok {
		return 0, fmt.Errorf("unable to find host binding for port %s in %s", containerPort, containerID)
	}

	if len(hostPortBindings) == 0 {
		return 0, fmt.Errorf("no host bindings for port %s in %s", containerPort, containerID)
	}

	n, err := strconv.Atoi(hostPortBindings[0].HostPort)
	if err != nil {
		return 0, fmt.Errorf("invalid host port %s: %w", hostPortBindings[0].HostPort, err)
	}

	return n, nil
}
