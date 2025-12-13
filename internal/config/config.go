package config

import (
	"os"
	"time"
)

type JailConfig struct {
}

type Config struct {
	IsProd             bool

	NsjailPath         string
	NsjailCfgPath      string

	WrapperPyPath      string
	HostSubmissionPath string

	JailHostname          string
	JailCwd               string
	JailSubmissionPath string
	JailWallTimeLimit     uint32
	JailTickTimeLimit     time.Duration
	JailCGroupPidsMax     uint64
	JailCGroupMemMax      uint64
	JailCGroupCpuMsPerSec uint32
	JailTmpfsSize         uint64
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
		JailHostname:          "jail",
		JailCwd:               "/",
		JailWallTimeLimit:     2 * 60 * 1000,          // 2 minutes
		JailTickTimeLimit:     500 * time.Millisecond, // 500 ms
		JailCGroupPidsMax:     10,                     // 10 processes
		JailCGroupMemMax:      100 * 1024 * 1024,      // 100 MB
		JailCGroupCpuMsPerSec: 200,                    // 20% CPU
		JailTmpfsSize:         100 * 1024 * 1024,      // 100 MB
	}
}
