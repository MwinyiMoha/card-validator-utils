package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggerConf(t *testing.T) {
	cfg := NewLoggerConfig()
	assert.Equal(t, "timestamp", cfg.Config.EncoderConfig.TimeKey)
}

func TestBuildLoggerSuccess(t *testing.T) {
	loggerCfg := NewLoggerConfig()
	logger, err := loggerCfg.BuildLogger()

	assert.NoError(t, err)
	assert.NotNil(t, logger)

	defer logger.Sync()
}

func TestBuildLoggerFailure(t *testing.T) {
	loggerCfg := NewLoggerConfig()
	loggerCfg.Config.OutputPaths = []string{"/invalid/path/to/logs"}

	logger, err := loggerCfg.BuildLogger()

	assert.Error(t, err)
	assert.Nil(t, logger)
}
