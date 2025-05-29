package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

func New() (*zap.Logger, error) {
	isDebug := os.Getenv("DEBUG") != ""

	logConfig := zap.NewProductionConfig()
	if isDebug {
		logConfig = zap.NewDevelopmentConfig()
	}

	l, err := logConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return l, nil
}
