### About

This experimental branch is trying to use podman-machine code for VM management in crc.
I've successfully used it on f40 (silverblue) with microshift 4.16.0 and usermode networking


### One-time setup
First, install `podman` and `virtiofsd`, and make sure that `podman machine start` works.
I needed to link `/usr/libexec/virtiofsd` to `~/.local/bin/virtiofsd` before this succeeds.
Once the basic install works, you can `podman machine stop; podman machine rm; podman machine reset`
to get rid of everything.

Now you can build `crc` from this branch, and set it up:
```
./out/linux-amd64/crc config set preset microshift
./out/linux-amd64/crc config set network-mode user
./out/linux-amd64/crc config set skip-check-vsock true
./out/linux-amd64/crc setup
```

### Testing

Before proceeding, make sure `crc daemon` is not running: `kill $(pgrep -f 'crc daemon')`.

The first attempt at running `crc start` will fail with:
```
Error creating machine: Error in driver during machine creation: cannot get file information for /var/home/teuf/.crc/machines/crc/crc.qcow2: stat /var/home/teuf/.crc/machines/crc/crc.qcow2: no such file or directory
```

2 files need to be created by hand at this step:
```
ln ~/.crc/machines/crc/crc.img ~/.crc/machines/crc/crc.qcow2
cp ~/.crc/cache/crc_microshift_libvirt_4.16.0_amd64/id_ecdsa_crc ~/.local/share/containers/podman/machine/machine
```

Running again `crc start --log-level debug` should go further:
```
DEBU gvproxy command-line: /usr/libexec/podman/gvproxy -debug -mtu 1500 -ssh-port 35329 -listen-qemu unix:///run/user/1000/podman/crc-gvproxy.sock -forward-sock /run/user/1000/podman/crc-api.sock -forward-dest /run/user/1000/podman/podman.sock -forward-user core -forward-identity /var/home/teuf/.local/share/containers/podman/machine/machine -pid-file /run/user/1000/podman/gvproxy.pid -log-file /run/user/1000/podman/gvproxy.log
DEBU socket length for /var/home/teuf/.config/containers/podman/machine/qemu/crc.ign is 61
DEBU socket length for /run/user/1000/podman/crc.sock is 30
DEBU socket length for /run/user/1000/podman/crc-gvproxy.sock is 38
DEBU socket length for /run/user/1000/podman/crc.sock is 30
DEBU socket length for /run/user/1000/podman/crc-gvproxy.sock is 38
DEBU checking that "gvproxy" socket is ready
WARN qemu cmd: [/usr/bin/qemu-system-x86_64 -accel kvm -cpu host -M memory-backend=mem -drive if=virtio,file=/var/home/teuf/.crc/machines/crc/crc.qcow2 -object memory-backend-memfd,id=mem,size=4096M,share=on -m 4096 -smp 90 -fw_cfg name=opt/com.coreos/config,file=/var/home/teuf/.config/containers/podman/machine/qemu/crc.ign -qmp unix:/run/user/1000/podman/qmp_crc.sock,server=on,wait=off -netdev stream,id=vlan,server=off,addr.type=unix,addr.path=/run/user/1000/podman/crc-gvproxy.sock -device virtio-net-pci,netdev=vlan,mac=5a:94:ef:e4:0c:ee -device virtio-serial -chardev socket,path=/run/user/1000/podman/crc.sock,server=on,wait=off,id=acrc_ready -device virtserialport,chardev=acrc_ready,name=org.fedoraproject.port.0 -pidfile /run/user/1000/podman/crc_vm.pid]

DEBU Started qemu pid 387099
```

At this point, you need to ssh into the VM to manually send a "VM Ready" notification. The ssh port can be found in the logs:
```
DEBU gvproxy command-line: /usr/libexec/podman/gvproxy -debug -mtu 1500 -ssh-port 35329 ...
```

Then you can run:
```
ssh  -i ~/.crc/cache/crc_microshift_libvirt_4.16.0_amd64/id_ecdsa_crc core@localhost -p 35329
sudo bash -c 'echo Ready >/dev/vport1p1'
```

Once this is done, crc startup will proceed, these warnings are expected:
```
DEBU retry loop: attempt 13
DEBU Running SSH command: host -R 3 foo.apps.crc.testing
DEBU SSH command results: err: Process exited with status 1, output: Host foo.apps.crc.testing not found: 3(NXDOMAIN)
DEBU error: Temporary error: ssh command error:
command : host -R 3 foo.apps.crc.testing
```


Once the messages below are printed, the cluster has started successfully.

```
INFO Adding microshift context to kubeconfig...
Started the MicroShift cluster.

Use the 'oc' command line interface:
  $ eval $(crc oc-env)
  $ oc COMMAND
```


You can run `eval $(crc oc-env)` and `oc get pods -A` against this cluster.











