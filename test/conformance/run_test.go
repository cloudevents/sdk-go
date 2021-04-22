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

var opt = godog.Options{
	Output: colors.Colored(os.Stdout),
}

func TestMain(m *testing.M) {
	flag.Parse()
	if len(flag.Args()) > 0 {
		opt.Paths = flag.Args()
	} else {
		opt.Paths = []string{
			"../../conformance/features/",
		}
	}

	opt.Format = "pretty"

	status := godog.RunWithOptions("CloudEvents", func(s *godog.Suite) {
		CloudEventsFeatureContext(s)
		HTTPFeatureContext(s)
		KafkaFeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
