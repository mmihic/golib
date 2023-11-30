package dockerx

import (
	"bufio"
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// A Logger logs output from the container.
type Logger func(msg string)

// ZapLogger returns an adapter that passes container logs to a zap logger.
func ZapLogger(log *zap.Logger) Logger {
	return func(msg string) { log.Info(msg) }
}

// StreamContainerLogs streams container logs in a background goroutine.
func StreamContainerLogs(ctx context.Context,
	cli client.APIClient, containerID string, log Logger,
	opts types.ContainerLogsOptions) (io.Closer, error) {
	r, err := cli.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		return nil, err
	}

	go func() {
		s := bufio.NewScanner(r)
		for s.Scan() {
			log(s.Text())
		}
	}()

	return r, nil
}
