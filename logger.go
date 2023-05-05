// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel

import (
	"fmt"
	"log"
)

type Logger interface {
	Debug(v ...any)
	Debugf(format string, v ...any)
	Debugln(v ...any)

	Info(v ...any)
	Infof(format string, v ...any)
	Infoln(v ...any)

	Warn(v ...any)
	Warnf(format string, v ...any)
	Warnln(v ...any)

	Error(v ...any)
	Errorf(format string, v ...any)
	Errorln(v ...any)

	Fatal(v ...any)
	Fatalf(format string, v ...any)
	Fatalln(v ...any)

	Panic(v ...any)
	Panicf(format string, v ...any)
	Panicln(v ...any)
}

type LoggerFactory interface {
	New(name string) Logger
}

type DefaultLoggerFactory struct{}

func (f *DefaultLoggerFactory) New(name string) Logger {
	l := log.Default()

	l.SetPrefix(fmt.Sprintf("%s: ", name))

	return &DefaultLogger{
		Logger: l,
	}
}

type DefaultLogger struct {
	*log.Logger
}

func (l *DefaultLogger) Debug(v ...any) {
	l.Logger.Print(v...)
}

func (l *DefaultLogger) Debugf(format string, v ...any) {
	l.Logger.Printf(format, v...)
}

func (l *DefaultLogger) Debugln(v ...any) {
	l.Logger.Println(v...)
}

func (l *DefaultLogger) Info(v ...any) {
	l.Logger.Print(v...)
}

func (l *DefaultLogger) Infof(format string, v ...any) {
	l.Logger.Printf(format, v...)
}

func (l *DefaultLogger) Infoln(v ...any) {
	l.Logger.Println(v...)
}

func (l *DefaultLogger) Warn(v ...any) {
	l.Logger.Print(v...)
}

func (l *DefaultLogger) Warnf(format string, v ...any) {
	l.Logger.Printf(format, v...)
}

func (l *DefaultLogger) Warnln(v ...any) {
	l.Logger.Println(v...)
}

func (l *DefaultLogger) Error(v ...any) {
	l.Logger.Print(v...)
}

func (l *DefaultLogger) Errorf(format string, v ...any) {
	l.Logger.Printf(format, v...)
}

func (l *DefaultLogger) Errorln(v ...any) {
	l.Logger.Println(v...)
}
