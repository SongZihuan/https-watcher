package config

import (
	"fmt"
	"github.com/SongZihuan/https-watcher/src/utils"
	"time"
)

type URLConfig struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Deadline string `yaml:"deadline"`
	Mark     string `yaml:"mark"`

	DeadlineDuration time.Duration `yaml:"-"`
}

type WatcherConfig struct {
	URLs []*URLConfig `yaml:"urls"`
}

func (w *WatcherConfig) setDefault() {
	for _, url := range w.URLs {
		if url.Name == "" {
			url.Name = url.URL
		}

		if url.Deadline == "" {
			url.Deadline = "15d"
		}
	}
	return
}

func (w *WatcherConfig) check() (err ConfigError) {
	if len(w.URLs) == 0 {
		return NewConfigError("not any urls")
	}

	for _, url := range w.URLs {
		if !utils.IsValidHTTPSURL(url.URL) {
			return NewConfigError(fmt.Sprintf("'%s' is not a valid https url", url))
		}

		url.DeadlineDuration = utils.ReadTimeDuration(url.Deadline)
		if url.DeadlineDuration <= 0 {
			return NewConfigError(fmt.Sprintf("'%s' is not a valid deadline", url.Deadline))
		}
	}

	return nil
}
