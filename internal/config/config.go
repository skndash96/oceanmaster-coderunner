package config

import (
	"os"
	"time"
)

type JailConfig struct {
	Hostname          string
	Cwd               string
	WallTimeLimit     uint32
	TickTimeLimit     time.Duration
	CGroupPidsMax     uint64
	CGroupMemMax      uint64
	CGroupCpuMsPerSec uint32
	TmpfsSize         uint64
}

type Config struct {
	NsjailPath         string
	NsjailCfgPath      string
	WrapperPyPath      string
	HostSubmissionPath string
	JailSubmissionPath string
	IsProd             bool

	Jail *JailConfig
}

func New() *Config {
	isProd := os.Getenv("PROD") == "true"

	return &Config{
		NsjailPath:         "/app/nsjail",
		NsjailCfgPath:      "/app/nsjail.cfg",
		WrapperPyPath:      "/wrapper.py",
		HostSubmissionPath: "/tmp/submission",
		JailSubmissionPath: "/submission",
		IsProd:             isProd,
		Jail: &JailConfig{
			Hostname:          "jail",
			Cwd:               "/",
			WallTimeLimit:     2 * 60 * 1000,                         // 2 minutes
			TickTimeLimit:     time.Duration(500 * time.Millisecond), // 500 ms
			CGroupPidsMax:     10,                                    // 10 processes
			CGroupMemMax:      100 * 1024 * 1024,                     // 100 MB
			CGroupCpuMsPerSec: 200,                                   // 20% CPU
			TmpfsSize:         100 * 1024 * 1024,                     // 100 MB
		},
	}
}
