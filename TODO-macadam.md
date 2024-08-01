### podman-machine

- podman-machine does not work on RHEL (there's a Jira issue about it)
    - might need to discuss libvirt VS direct QEMU use on linux
- I tried to reuse the code in pkg/machine/shim/ as much as possible, the goal
  would be to make it more flexible to accommodate crc/... use cases
- changes to podman-machine:
    - configurable way to fetch the disk image so that I can tell podman-machine to use
      a crc bundle instead of fetching an image from quay.io (there's already a
      todo in the code for this)
    - more configurable env.MachineDirs?
    - will it be an issue to not use ignition? (I think not)
    - make configuration/forwarding of the podman connection socket optional? (though crc also does it)
    - logging/printing to stdout would ideally be optional

### crc

- SSH code needs changes to work well with podman-machine
    - need to tell podman-machine where the bundle SSH key is
    - need to teach crc to use a dynamic SSH port
    - need to understand if podman-machine can update the SSH key it uses to
      connect to the VM as it is one of the first things CRC does
- network stack needs to be removed from daemon: podman-machine uses gvproxy for this.
- podman-machine does not start gvproxy with a http api endpoint on the host,
  only in the guest. This is different from what crc does.
  Need to decide on which side to address this
- Need an equivalent of https://github.com/containers/podman/blob/main/pkg/machine/ignition/ready.go
  in crc/snc
- gvproxy is not privileged enough to open port 443 or 80
- implement the other machine driver methods (stop/delete/...)
