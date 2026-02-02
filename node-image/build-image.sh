#!/usr/bin/env sh
set -exu

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./rootfs/bin/mcprobe ../mcprobe

sudo alpine-make-vm-image \
        --arch x86_64 \
        --boot-mode BIOS \
        --fs-skel-dir ./rootfs \
        --fs-skel-chown root:root \
        --partition \
        --image-format qcow2 \
        --image-size 5G \
        --initfs-features "kms scsi virtio nvme" \
        --kernel-flavor "virt" \
        --keys-dir /dev/null \
        --packages "$(cat packages)" \
        --repositories-file repositories \
        --rootfs ext4 \
        --script-chroot \
        --serial-console \
        mcpulse-node.qcow2 -- ./configure.sh
