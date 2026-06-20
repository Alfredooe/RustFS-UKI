# alfpine

**Alpine Linux, running entirely from initramfs as a Unified Kernel Image (UKI).**

This is a stripped-down, minimal Alpine system that boots directly from a single UKI file running entirely from initramfs.

## How it works

1. `alpine-make-rootfs` builds an Alpine rootfs from a declarative set of files
2. Everything (minus `/boot`) is packed into a gzipped cpio initramfs
3. `ukify` combines kernel + initramfs + cmdline into a **Unified Kernel Image** — a single `.efi` file
4. UEFI firmware boots it directly. No bootloader, no root filesystem, no disk.

The entire OS runs in RAM. Changes are lost on reboot unless you add persistent mounts in `/etc/fstab`.

## Prerequisites

- **Docker**
- **QEMU**
- **OVMF** edk2-ovmf`

## Quick start

```sh
# Build a UKI image
./alfpine build

# Test it in QEMU
./alfpine qemu

# Write it to a USB stick
./alfpine flash /dev/sdb
```

## Customizing

### Change the root password

Generate a new hash:
```sh
openssl passwd -6
```

Replace the hash in `setup.sh` in the `chpasswd` block.

### Add your SSH key

Copy your public key into the root skeleton before building:
```sh
mkdir -p root/root/.ssh
cp ~/.ssh/id_ed25519.pub root/root/.ssh/authorized_keys
```

### Add packages

Edit `packages` and add any Alpine package name. Rebuild.

### Add files

Drop files into `root/` — they're copied verbatim into the image. For example:
- `root/etc/network/interfaces` → `/etc/network/interfaces` in the image
- `root/usr/local/bin/myscript` → `/usr/local/bin/myscript` in the image

### Run commands at boot

Add a script to `root/etc/local.d/` — any executable `.start` file there runs at boot.

## structure

```
alfpine                  # Wrapper script (build/qemu/flash)
build-inner.sh           # Build script (runs inside Alpine container)
setup.sh                 # Chroot setup (services, users, SSH keys)
packages                 # List of apk packages to install
root/                    # Root filesystem skeleton
  etc/
    inittab              # Busybox init → OpenRC
    hostname             # System hostname
    fstab                # Filesystem mounts (add persistence here)
    network/interfaces   # Network config (DHCP on eth0)
    conf.d/sshd          # Disable non-Ed25519 SSH host keys
    motd                 # Login banner
```

