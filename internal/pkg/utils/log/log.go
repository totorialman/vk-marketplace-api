package log

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func GetFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	if n == 0 {
		return "unknown"  
	}

	frames := runtime.CallersFrames(pc[:n])
	frame, more := frames.Next()
	if !more {
		return "unknown"  
	}

	values := strings.Split(frame.Function, "/")
	if len(values) == 0 {
		return "unknown"
	}

	funcName := values[len(values)-1]
	if dotIndex := strings.LastIndex(funcName, "."); dotIndex != -1 {
		funcName = funcName[dotIndex+1:] 
	}

	return funcName
}


func LogHandlerInfo(logger *slog.Logger, msg string, statusCode int) {
	logger = logger.With(slog.String("status", strconv.Itoa(statusCode)))
	logger.Info(msg)
}

func LogHandlerError(logger *slog.Logger, err error, statusCode int) {
	logger = logger.With(slog.String("status", strconv.Itoa(statusCode)))

	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr != nil {
		logger.Error(unwrappedErr.Error())
	} else {
		logger.Error(err.Error())
	}
}

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value("logger").(*slog.Logger); ok {
		return logger
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Error("Couldnt get logger from context")

	return logger
}