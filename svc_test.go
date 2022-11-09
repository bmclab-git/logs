package main

import (
	"context"
	"testing"
	"time"

	"github.com/Lukiya/logs/model"
	"github.com/Lukiya/logs/svc"
)

func TestWriteLog(t *testing.T) {
	logService := new(svc.LogService)
	logService.WriteLogEntry(context.Background(), &model.WriteLogCommand{
		ClientID: "DL",
		LogEntry: &model.LogEntry{
			Level:        model.LogLevel_Debug,
			User:         "Lucas",
			TraceNo:      "xxxxx",
			Message:      "Errors are a language-agnostic part that helps to write code in such a way that no unexpected thing happens.",
			Error:        "BBB",
			CreatedOnUtc: time.Now().UTC().UnixMilli(),
			Payload:      `{"name":"test","score":3.98}`,
		},
	})

	time.Sleep(1 * time.Second)
}
