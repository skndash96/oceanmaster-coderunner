package nsjail

//go:generate protoc -I../../nsjail --go_out ./ --go_opt Mconfig.proto=/proto_nsjail config.proto

import (
	"fmt"
	"os"

	"github.com/delta/code-runner/internal/cgroup"
	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/nsjail/proto_nsjail"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func WriteConfig(c *config.Config, cg cgroup.Cgroup) (*proto_nsjail.NsJailConfig, error) {
	msg := &proto_nsjail.NsJailConfig{}

	msg.Mode = proto_nsjail.Mode_ONCE.Enum()
	msg.TimeLimit = proto.Uint32(c.JailWallTimeoutMS)
	msg.RlimitAsType = proto_nsjail.RLimit_HARD.Enum()
	msg.RlimitCpuType = proto_nsjail.RLimit_HARD.Enum()
	msg.RlimitFsizeType = proto_nsjail.RLimit_HARD.Enum()
	msg.RlimitNofileType = proto_nsjail.RLimit_HARD.Enum()

	msg.CgroupPidsMax = proto.Uint64(c.JailCGroupPidsMax)
	msg.CgroupMemMax = proto.Uint64(c.JailCGroupMemMax)
	msg.CgroupCpuMsPerSec = proto.Uint32(c.JailCGroupCpuMsPerSec)

	// Set the respective v1 or v2 cgroup configuration
	cg.SetConfig(msg)

	msg.Envar = []string{
		fmt.Sprintf("PYTHONPATH=%s", c.JailSubmissionPath),
	}

	if c.IsProd {
		msg.LogLevel = proto_nsjail.LogLevel_ERROR.Enum()
	} else {
		msg.LogLevel = proto_nsjail.LogLevel_DEBUG.Enum()
	}

	msg.Hostname = proto.String(c.JailHostname)
	msg.Cwd = proto.String(c.JailCwd)

	msg.ExecBin = &proto_nsjail.Exe{
		Path: proto.String("/bin/sh"),
		Arg: []string{
			"-c",
			"echo testerr >&2 && exec /usr/local/bin/python3 wrapper.py",
		},
	}

	msg.Mount = []*proto_nsjail.MountPt{{
		Src:    proto.String("/srv"),
		Dst:    proto.String("/"),
		IsBind: proto.Bool(true),
		Nodev:  proto.Bool(true),
		Nosuid: proto.Bool(true),
	}, {
		Dst:     proto.String("/tmp"),
		Fstype:  proto.String("tmpfs"),
		Rw:      proto.Bool(true),
		Options: proto.String(fmt.Sprintf("size=%d", c.JailTmpfsSize)),
		Nodev:   proto.Bool(true),
		Nosuid:  proto.Bool(true),
	},
	}

	_, err := os.Stat("/srv/proc")
	if err == nil {
		msg.Mount = append(msg.Mount, &proto_nsjail.MountPt{
			Dst:    proto.String("/proc"),
			Fstype: proto.String("proc"),
			Nodev:  proto.Bool(true),
			Nosuid: proto.Bool(true),
			Noexec: proto.Bool(true),
		})
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("check /srv/proc: %w", err)
	}

	err = write(c.NsjailCfgPath, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func write(path string, msg *proto_nsjail.NsJailConfig) error {
	content, err := prototext.Marshal(msg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		return err
	}
	return nil
}
