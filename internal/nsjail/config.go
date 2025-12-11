package nsjail

//go:generate protoc -I../../nsjail --go_out ./ --go_opt Mconfig.proto=/proto_nsjail config.proto

import (
	"fmt"
	"os"

	"github.com/delta/code-runner/internal/config"
	"github.com/delta/code-runner/internal/nsjail/proto_nsjail"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func GetNsjailConfig(c *config.Config) (*proto_nsjail.NsJailConfig, error) {
	msg := &proto_nsjail.NsJailConfig{}

	msg.Mode = proto_nsjail.Mode_ONCE.Enum()
	msg.TimeLimit = proto.Uint32(c.Jail.TimeLimit)
	msg.RlimitAsType = proto_nsjail.RLimit_HARD.Enum()
	msg.RlimitCpuType = proto_nsjail.RLimit_HARD.Enum()
	msg.RlimitFsizeType = proto_nsjail.RLimit_HARD.Enum()
	msg.RlimitNofileType = proto_nsjail.RLimit_HARD.Enum()
	msg.CgroupPidsMax = proto.Uint64(c.Jail.CGroupPidsMax)
	msg.CgroupMemMax = proto.Uint64(c.Jail.CGroupMemMax)
	msg.CgroupCpuMsPerSec = proto.Uint32(c.Jail.CGroupCpuMsPerSec)

	msg.Hostname = proto.String(c.Jail.Hostname)
	msg.Cwd = proto.String(c.Jail.Cwd)

	msg.ExecBin = &proto_nsjail.Exe{
		Path: proto.String("python3"),
		Arg:  []string{"wrapper.py"},
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
		Options: proto.String(fmt.Sprintf("size=%d", c.Jail.TmpfsSize)),
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

	return msg, nil
}

func Write(c *config.Config, msg *proto_nsjail.NsJailConfig) error {
	content, err := prototext.Marshal(msg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(c.NsjailCfgPath, content, 0644); err != nil {
		return err
	}
	return nil
}
