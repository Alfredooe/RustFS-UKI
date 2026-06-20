#!/bin/sh
set -e

# --- sysinit ---
rc-update add devfs sysinit
rc-update add dmesg sysinit
rc-update add udev sysinit
rc-update add udev-trigger sysinit
rc-update add udev-settle sysinit

# --- boot ---
rc-update add hwclock boot
rc-update add modules boot
rc-update add sysctl boot
rc-update add hostname boot
rc-update add bootmisc boot
rc-update add syslog boot
rc-update add klogd boot
rc-update add networking boot
rc-update add seedrng boot
rc-update add alfpine-disks boot

# --- shutdown ---
rc-update add mount-ro shutdown
rc-update add killprocs shutdown

# --- default ---
rc-update add acpid default
rc-update add local default
rc-update add openntpd default
rc-update add sshd default
rc-update add node-exporter default
rc-update add rustfs default
rc-update add udev-postmount default

# RustFS data directory (ephemeral — mount persistent storage here for production)
mkdir -p /data

# --- Generate SSH host key at build time ---
# Each build gets its own Ed25519 host key.
ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key -N "" < /dev/null

# --- Set root password ---
# IMPORTANT: Change this! Generate with: openssl passwd -6
# Default password is "alpine" — CHANGE IT on first boot or override this file.
chpasswd -e <<'EOF'
root:$6$HDoXyxnETHAxrOQ7$MnIcSctsJdw19r4lZ7OmowJg4CzJ1XCYVXOqi/5.W3ZWR7j4oqpEM77LUMUBY1x8c.OaNa044eqn/H6GElctG.
EOF

# --- Set up authorized_keys for root ---
# If you placed a key at root/root/.ssh/authorized_keys, it's already copied.
# Otherwise, create an empty file so the directory exists.
mkdir -p /root/.ssh
chmod 700 /root/.ssh
touch /root/.ssh/authorized_keys
chmod 600 /root/.ssh/authorized_keys

# --- Enable passwordless sudo for wheel group (optional) ---
mkdir -p /etc/sudoers.d
echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel-nopasswd
