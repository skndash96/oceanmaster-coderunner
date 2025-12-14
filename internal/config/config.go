package config

import (
	"os"
)

type JailConfig struct {
}

type Config struct {
	IsProd bool

	NsjailPath    string
	NsjailCfgPath string

	WrapperPyPath      string
	HostSubmissionPath string

	JailHostname          string
	JailCwd               string
	JailSubmissionPath    string
	JailCGroupPidsMax     uint64
	JailCGroupMemMax      uint64
	JailCGroupCpuMsPerSec uint32
	JailTmpfsSize         uint64

	JailWallTimeoutMS      uint32
	JailHandshakeTimeoutMS uint32
	JailTickTimeoutMS      uint32
}

func New() *Config {
	isProd := os.Getenv("PROD") == "true"

	return &Config{
		IsProd: isProd,

		NsjailPath:    "/app/nsjail",
		NsjailCfgPath: "/app/nsjail.cfg",

		WrapperPyPath:      "/wrapper.py",
		HostSubmissionPath: "/tmp/submission",

		JailSubmissionPath:    "/submission",
		JailHostname:          "jail",
		JailCwd:               "/",
		JailCGroupPidsMax:     20,                // 20 processes
		JailCGroupMemMax:      100 * 1024 * 1024, // 100 MB
		JailCGroupCpuMsPerSec: 200,               // 20% CPU
		JailTmpfsSize:         100 * 1024 * 1024, // 100 MB

		// Wall >= Setup + 1000 * Tick
		JailWallTimeoutMS:      2 * 60 * 1000, // 2 minutes
		JailHandshakeTimeoutMS: 10 * 1000,     // 10 seconds
		JailTickTimeoutMS:      500,           // 500 milliseconds
	}
}
