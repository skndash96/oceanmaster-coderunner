# What's this dir?

So inorder to setup cgroup limitations (mem, disk, i/o), nsjail wants to create cgroup namespaces (user ns etc.)

Inside docker (even with --privileged and/or --userns=host), it cannot do that. It errors out `Couldn't initialize cgroup user namespace for pid=16` by nsjail (i guess)

So chatgpt said cgroup is just optional for my case. So be it.

## How to re-add this cgroup limitation
See patch of commit `7bd8f1922aec5bac73440250777ba0b81083ebb0` to next commit.

Basically, add these:
```go
	//main.go
	cg, err := cgroup.UnshareAndMount()
	if err != nil {
		return err
	}
	
	_, err := nsjail.WriteConfig(cfg, cg)
```

```go
	msg.CgroupPidsMax = proto.Uint64(c.JailCGroupPidsMax)
	msg.CgroupMemMax = proto.Uint64(c.JailCGroupMemMax)
	msg.CgroupCpuMsPerSec = proto.Uint32(c.JailCGroupCpuMsPerSec)

	// Set the respective v1 or v2 cgroup configuration
	cg.SetConfig(msg)
```
