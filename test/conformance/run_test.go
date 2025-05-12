//go:build conformance
// +build conformance

/*
Copyright 2021 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/
package conformance

import (
	"flag"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
}

func TestMain(m *testing.M) {
	flag.Parse()
	if len(flag.Args()) > 0 {
		opts.Paths = flag.Args()
	} else {
		opts.Paths = []string{
			"../../conformance/features/",
		}
	}

	status := godog.TestSuite{
		Name:                "CloudEvents",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	CloudEventsFeatureContext(ctx)
	HTTPFeatureContext(ctx)
	KafkaFeatureContext(ctx)
}
