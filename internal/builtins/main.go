// Copyright © 2017 Circonus, Inc. <support@circonus.com>
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// Package builtins marshals internal (non-plugin) metric collectors (e.g. procfs, wmi, etc.)
package builtins

import (
	"sync"
	"time"

	"github.com/circonus-labs/circonus-agent/internal/builtins/collector"
	cgm "github.com/circonus-labs/circonus-gometrics"
	appstats "github.com/maier/go-appstats"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// New creates a new builtins manager
func New() (*Builtins, error) {
	b := Builtins{
		collectors: make(map[string]collector.Collector),
		logger:     log.With().Str("pkg", "builtins").Logger(),
	}

	b.logger.Info().Msg("configuring builtins")

	err := b.configure()
	if err != nil {
		return nil, errors.Wrap(err, "configuring builtins")
	}

	return &b, nil
}

// Run triggers internal collectors to gather metrics
func (b *Builtins) Run(id string) error {
	b.Lock()

	if len(b.collectors) == 0 {
		b.Unlock()
		return nil // nothing to do
	}

	if b.running {
		b.logger.Warn().Msg("already in progress")
		b.Unlock()
		return nil
	}

	b.running = true
	b.Unlock()

	start := time.Now()
	appstats.MapSet("builtins", "last_start", start)

	var wg sync.WaitGroup

	if id == "" {
		wg.Add(len(b.collectors))
		for id, c := range b.collectors {
			b.logger.Debug().Str("builtin", id).Msg("collecting")
			go func(id string, c collector.Collector) {
				err := c.Collect()
				if err != nil {
					b.logger.Error().Err(err).Msg(id)
				}
				wg.Done()
			}(id, c)
		}
	} else {
		c, ok := b.collectors[id]
		if ok {
			wg.Add(1)
			b.logger.Debug().Str("builtin", id).Msg("collecting")
			go func(id string, c collector.Collector) {
				err := c.Collect()
				if err != nil {
					b.logger.Error().Err(err).Msg(id)
				}
				wg.Done()
			}(id, c)
		} else {
			b.logger.Warn().Str("id", id).Msg("unknown builtin")
		}
	}

	wg.Wait()

	b.logger.Debug().Msg("all builtins done")

	appstats.MapSet("builtins", "last_end", time.Now())
	appstats.MapSet("builtins", "last_duration", time.Since(start))

	b.Lock()
	b.running = false
	b.Unlock()

	return nil
}

// IsBuiltin determines if an id is a builtin or not
func (b *Builtins) IsBuiltin(id string) bool {
	if id == "" {
		return false
	}

	b.Lock()
	defer b.Unlock()

	if len(b.collectors) == 0 {
		return false
	}

	_, ok := b.collectors[id]

	return ok
}

// Flush returns current metrics for all collectors
func (b *Builtins) Flush(id string) *cgm.Metrics {
	b.Lock()
	defer b.Unlock()

	appstats.MapSet("builtins", "last_flush", time.Now())

	metrics := cgm.Metrics{}

	if len(b.collectors) == 0 {
		return &metrics // nothing to do
	}

	for _, c := range b.collectors {
		for name, val := range c.Flush() {
			metrics[name] = val
		}
	}

	return &metrics
}
