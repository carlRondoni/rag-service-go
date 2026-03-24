package service_container

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func InitLogs() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logger := zerolog.New(os.Stdout).
		With().
		Str("service", "llm-agent-go").
		Timestamp().
		Logger()

	return logger
}
