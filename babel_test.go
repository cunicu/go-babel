// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package babel_test

import (
	"log/slog"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func SetupLogging() {
	programLevel := &slog.LevelVar{}
	programLevel.Set(slog.LevelDebug)

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	logger := slog.New(handler)

	slog.SetDefault(logger)
}

func TestMain(t *testing.T) {
	SetupLogging()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Proto suite")
}
