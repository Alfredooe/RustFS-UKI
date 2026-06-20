#!/bin/sh
set -e

__() { printf "\n\033[1;32m* %s [%s]\033[0m\n" "$1" "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"; }

ROOTFS_DEST=$(mktemp -d)

__ "Fetching alpine-make-rootfs"

wget https://raw.githubusercontent.com/alpinelinux/alpine-make-rootfs/v0.7.0/alpine-make-rootfs \
    && echo '91ceb95b020260832417b01e45ce02c3a250c4527835d1bdf486bf44f80287dc  alpine-make-rootfs' \
    | sha256sum -c || exit 1 && chmod +x alpine-make-rootfs

__ "Building rootfs"

mkdir -p "$ROOTFS_DEST/etc"
basename "$1" > "$ROOTFS_DEST/etc/alfpine-release"

# Stop mkinitfs from running during apk install.
mkdir -p "$ROOTFS_DEST/etc/mkinitfs"
echo "disable_trigger=yes" > "$ROOTFS_DEST/etc/mkinitfs/mkinitfs.conf"

export ALPINE_BRANCH=edge
export SCRIPT_CHROOT=yes
export FS_SKEL_DIR=/mnt/root
export FS_SKEL_CHOWN=root:root
PACKAGES="$(grep -v -e '^#' -e '^$' /mnt/packages)"
export PACKAGES
./alpine-make-rootfs "$ROOTFS_DEST" /mnt/setup.sh

__ "Downloading RustFS"
RUSTFS_VER="1.0.0-beta.8"
RUSTFS_URL="https://github.com/rustfs/rustfs/releases/download/${RUSTFS_VER}/rustfs-linux-x86_64-musl-v${RUSTFS_VER}.zip"
wget -q "$RUSTFS_URL" -O /tmp/rustfs.zip
apk add --no-cache unzip >/dev/null 2>&1
mkdir -p "$ROOTFS_DEST/usr/local/bin"
unzip -p /tmp/rustfs.zip > "$ROOTFS_DEST/usr/local/bin/rustfs"
chmod +x "$ROOTFS_DEST/usr/local/bin/rustfs"
rm /tmp/rustfs.zip

__ "Compiling dashboard"
apk add --no-cache go
cd /mnt/dashboard
CGO_ENABLED=0 go build -o "$ROOTFS_DEST/usr/local/bin/rustfs-dashboard" -ldflags="-s -w" .
cd /

__ "Building initramfs"

cd "$ROOTFS_DEST"
find . -path "./boot" -prune -o -print | cpio -o -H newc | gzip > "$ROOTFS_DEST/boot/initramfs-lts"

__ "Building UKI image"

apk add --no-cache systemd-efistub ukify

# Console setup: use ttyS0 for serial + tty1 for VGA
# rdinit=/sbin/init because busybox init is at /sbin/init (not /init)
CMDLINE="rdinit=/sbin/init console=tty1 console=ttyS0"

ukify build --output "$1" --cmdline "$CMDLINE" \
    --linux "$ROOTFS_DEST/boot/vmlinuz-lts" \
    --initrd "$ROOTFS_DEST/boot/initramfs-lts" \
    --os-release "@$ROOTFS_DEST/etc/alfpine-release"

__ "Created image!"

ls -lh "$1"
