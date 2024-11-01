package log

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import (
	slog "log"
	"os"
)

type ConsoleLog struct {
	LogHandler
}

func NewConsoleLog(date bool) (consolelog *ConsoleLog, err error) {
	log := slog.New(os.Stdout, "", 0)
	consolelog = &ConsoleLog{}
	consolelog.LogHandler = LogHandler{log: log, level: DEBUG, date: date}
	return consolelog, nil
}
