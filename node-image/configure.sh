#!/bin/sh
set -exu

_step_counter=0
step() {
  set +x
  _step_counter=$(( _step_counter + 1 ))
  printf '\n\033[1;36m%d) %s\033[0m\n' $_step_counter "$@" >&2  # bold cyan
  set -x
}

uname -a

step 'Set up Cloud-Init'
# https://gitlab.alpinelinux.org/alpine/aports/-/blob/d0b38d72631b17730c110ee0abebf23bfc9c7017/community/cloud-init/README.Alpine

setup-devd udev || true
setup-cloud-init

mkdir -p /etc/cloud/cloud.cfg.d
cat > /etc/cloud/cloud.cfg.d/01_datasources.cfg <<-EOF
datasource_list: [ NoCloud, Hetzner, GCE, DigitalOcean, Ec2,  None ]
EOF
cat > /etc/cloud/cloud.cfg.d/99_minimal.cfg <<-EOF
users: []
disable_root: false

cloud_init_modules:
 - write-files

cloud_config_modules:
 - mounts
 - growpart
 - resizefs

cloud_final_modules: []
EOF

step 'Download Server JAR'
mkdir -p /srv/mcpulse/server
curl -o /srv/mcpulse/server/server.jar https://api.papermc.io/v2/projects/paper/versions/1.21.4/builds/88/downloads/paper-1.21.4-88.jar
echo 'eula=true' > /srv/mcpulse/server/eula.txt

step 'Create mcpulse-server-runner user'
adduser -D -h /srv/mcpulse/server mcpulse-server-runner
chown -R mcpulse-server-runner:mcpulse-server-runner \
  /srv/mcpulse/server \
  /srv/mcpulse/scripts/start-server.sh

step 'Ensure permissions'
chmod 0700 /bin/mcprobe \
           /srv/mcpulse/scripts/watchdog.sh \
           /srv/mcpulse/scripts/start-server.sh \
           /etc/init.d/mcpulse-watchdog \
           /etc/init.d/mcpulse-server

step 'Enable system services'
rc-update add acpid default
rc-update add ntpd default

step 'Enable mcpulse services'
rc-update add mcpulse-server default
rc-update add mcpulse-watchdog default

step 'Ensure services logging is on'
sed -Ei -e 's/^[# ](rc_logger)=.*/\1=YES/' /etc/rc.conf
