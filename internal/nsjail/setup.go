package nsjail

import (
	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/nsjail/proto_nsjail"
)

func Setup(cfg *config.Config) *proto_nsjail.NsJailConfig {
	// why doesn't this work without --privileged?
	// err := MountDev(cfg)
	// if err != nil {
	// 	panic(err)
	// }

	msg, err := GetNsjailConfig(cfg)
	if err != nil {
		panic(err)
	}

	if err = Write(cfg, msg); err != nil {
		panic(err)
	}

	return msg
}
