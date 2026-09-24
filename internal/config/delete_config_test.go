package config

import (
	"flag"
	"os"
	"testing"
	"time"
)

func configWithArgs(t *testing.T, args ...string) *Config {
	t.Helper()

	originalFlags := flag.CommandLine
	originalArgs := os.Args

	flag.CommandLine = flag.NewFlagSet(
		"shortener-test",
		flag.ContinueOnError,
	)
	os.Args = append([]string{"shortener-test"}, args...)

	t.Cleanup(func() {
		flag.CommandLine = originalFlags
		os.Args = originalArgs
	})

	return NewConfig()
}

func TestDeletionFlags(t *testing.T) {
	cfg := configWithArgs(
		t,
		"-delete-batch-size=25",
		"-delete-flush-interval=750ms",
		"-delete-workers=4",
	)

	if cfg.DeleteBatchSize != 25 ||
		cfg.DeleteFlushInterval != 750*time.Millisecond ||
		cfg.DeleteWorkers != 4 {
		t.Fatalf(
			"unexpected deletion settings: %+v",
			cfg,
		)
	}
}

func TestDeletionFlagsKeepServiceDefaults(t *testing.T) {
	cfg := configWithArgs(t)

	if cfg.DeleteBatchSize != 0 ||
		cfg.DeleteFlushInterval != 0 ||
		cfg.DeleteWorkers != 0 {
		t.Fatalf(
			"unset flags must keep service defaults: %+v",
			cfg,
		)
	}
}
