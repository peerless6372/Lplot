package base

import (
	"net/http"
	_ "net/http/pprof"
)

type PprofConfig struct {
	Enable bool `yaml:"enable"`
}

func RegisterProf() {
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			panic("pprof server start error: " + err.Error())
		}
	}()
}
