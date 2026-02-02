#!/usr/bin/env bash
set -exu

cat > user-data <<EOF
#cloud-config
write_files:
  - path: /etc/mcpulse/gateway.env
    owner: root:root
    permissions: '0600'
    content: |
      export GATEWAY_URL="http://10.0.2.2:8989"
      export SERVER_TOKEN="local-dev-token"
  - path: /srv/mcpulse/server/.mcpulse.env
    owner: mcpulse-server-runner:mcpulse-server-runner
    permissions: '0600'
    content: |
      export SERVER_JAR="server.jar"
      export SERVER_ARGS=""
      export JAVA_VERSION=""
      export JVM_FLAGS="java -Xms10G -Xmx10G -XX:+UseG1GC -XX:+ParallelRefProcEnabled -XX:MaxGCPauseMillis=200 -XX:+UnlockExperimentalVMOptions -XX:+DisableExplicitGC -XX:+AlwaysPreTouch -XX:G1NewSizePercent=30 -XX:G1MaxNewSizePercent=40 -XX:G1HeapRegionSize=8M -XX:G1ReservePercent=20 -XX:G1HeapWastePercent=5 -XX:G1MixedGCCountTarget=4 -XX:InitiatingHeapOccupancyPercent=15 -XX:G1MixedGCLiveThresholdPercent=90 -XX:G1RSetUpdatingPauseTimePercent=5 -XX:SurvivorRatio=32 -XX:+PerfDisableSharedMem -XX:MaxTenuringThreshold=1 -Dusing.aikars.flags=https://mcflags.emc.gs -Daikars.new.flags=true -jar paper.jar --nogui"
EOF
touch meta-data

genisoimage -output cidata.iso -volid cidata -joliet -rock user-data meta-data

sudo qemu-system-x86_64 \
        -enable-kvm \
        -m 2G \
        -smp 2 \
        -drive file=mcpulse-node.qcow2,format=qcow2,if=virtio \
        -drive file=cidata.iso,driver=raw,if=virtio \
        -netdev user,id=net0 \
        -device virtio-net-pci,netdev=net0 \
        -nographic
