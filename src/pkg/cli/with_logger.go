package cli

import "go.uber.org/zap"

// WithLogger is a mix-in for commands that use a logger
type WithLogger struct {
	Log *zap.Logger `kong:"-"`
}

// AfterApply injects the log.
func (cmd *WithLogger) AfterApply(log *zap.Logger) error {
	cmd.Log = log
	return nil
}
