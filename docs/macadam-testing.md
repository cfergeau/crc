## Prerequisites

- have podman 5 installed
- needs a newer gvproxy with https://github.com/containers/gvisor-tap-vsock/pull/429
  there's a preflight check detecting this
- download one of these bundles:
    - OpenShift Bundles
        - https://storage.googleapis.com/crc-bundle/crc_hyperv_4.17.9_amd64_990.crcbundle
        - https://storage.googleapis.com/crc-bundle/crc_libvirt_4.17.9_amd64_990.crcbundle
        - https://storage.googleapis.com/crc-bundle/crc_vfkit_4.17.9_amd64_990.crcbundle
        - https://cdk-builds.usersys.redhat.com/builds/crc/unreleased/openshift/pr-990-4.17.7/crc_vfkit_4.17.7_arm64.crcbundle
    - Microshift Bundles
        - https://storage.googleapis.com/crc-bundle/crc_microshift_hyperv_4.17.8_amd64_990.crcbundle
        - https://storage.googleapis.com/crc-bundle/crc_microshift_libvirt_4.17.8_amd64_990.crcbundle
        - https://storage.googleapis.com/crc-bundle/crc_microshift_vfkit_4.17.8_amd64_990.crcbundle
        - https://cdk-builds.usersys.redhat.com/builds/crc/unreleased/microshift/pr-990-4.17.8/crc_microshift_vfkit_4.17.8_arm64.crcbundle
    - see https://redhat-internal.slack.com/archives/CGG67T45P/p1734080339018939 for the discussion related to these bundles
- build crc from https://github.com/cfergeau/crc/tree/macadam
- then you can run `crc delete` `crc setup --bundle …` and `crc start --bundle …`


## Troubleshooting

After creating a VM with macadam, to manually remove a machine if crc delete is
not working:
- `podman machine remove crc`
- `rm -rf ~/.crc/machines/crc`

Issues can be reported/tracked in https://github.com/cfergeau/macadam/issues


## Notable differences with regular crc

- single machine driver, no more differences between linux/macos/windows code.
  This should make it possible to simplify further the machine driver
  instantiation, API, ...
- networking code goes through gvproxy, not the daemon. The daemon is still used
  for crc's REST API
- ssh key updates are a bit messy as this was not something already exposed in
  podman machine
- this has been mainly tested on linux, macos and windows may or may not work
- I've focused on microshift bundles as they are quicker to start but there
  should be no big difference with openshift bundles
- the bundle needs to have a "VM ready" systemd unit indicating to podman machine
  that it finished booting
- libvirt is not used on linux, qemu is run directly
- this code depends on modified versions of:
    - [podman](https://github.com/cfergeau/podman/tree/crc),
    - [machine](https://github.com/cfergeau/machine/tree/ssh) and
    - [gvisor-tap-vsock](https://github.com/cfergeau/gvisor-tap-vsock/tree/i425)
- the macadam machine driver code is in https://github.com/cfergeau/macadam/tree/machinedriver/pkg/machinedriver
