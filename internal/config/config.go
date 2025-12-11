package config

type JailConfig struct {
	Hostname          string
	Cwd               string
	TimeLimit         uint32
	CGroupPidsMax     uint64
	CGroupMemMax      uint64
	CGroupCpuMsPerSec uint32
	TmpfsSize         uint64

	SubmissionPath string
}

type Config struct {
	NsjailPath         string
	NsjailCfgPath      string
	WrapperPyPath      string
	HostSubmissionPath string

	Devices []string
	Jail    *JailConfig
}

func New() *Config {
	return &Config{
		NsjailPath:         "/nsjail",
		NsjailCfgPath:      "/nsjail.cfg",
		WrapperPyPath:      "/wrapper.py",
		HostSubmissionPath: "/tmp/submission",
		Jail: &JailConfig{
			SubmissionPath:    "/submission",
			Hostname:          "jail",
			Cwd:               "/",
			TimeLimit:         5 * 60 * 1000,     // 5 minutes
			CGroupPidsMax:     10,                // 10 processes
			CGroupMemMax:      100 * 1024 * 1024, // 100 MB
			CGroupCpuMsPerSec: 200,               // 20% CPU
			TmpfsSize:         100 * 1024 * 1024, // 100 MB
		},
		Devices: []string{"null", "zero", "random", "urandom"}, // devices might be reqd for python
	}
}
