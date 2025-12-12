package config

import "os"

type JailConfig struct {
	Hostname          string
	Cwd               string
	TimeLimit         uint32
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
			TimeLimit:         5 * 60 * 1000,     // 5 minutes
			CGroupPidsMax:     10,                // 10 processes
			CGroupMemMax:      100 * 1024 * 1024, // 100 MB
			CGroupCpuMsPerSec: 200,               // 20% CPU
			TmpfsSize:         100 * 1024 * 1024, // 100 MB
		},
	}
}
