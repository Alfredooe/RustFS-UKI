# alfpine

Alpine Linux as a single-file UKI. Boots directly from UEFI — no disk, no bootloader, no root partition.

Everything runs in RAM. Reboot = factory reset.

## Usage

```sh
./alfpine build          # Build a UKI image
./alfpine qemu           # Test in QEMU
./alfpine flash /dev/sdb # Write to USB stick
```

Only dependency is **Docker**. Build and QEMU Testing run inside containers.

Login: `root` / `root`.

## Customizing

**Change password:** generate a hash with `openssl passwd -6`, replace it in `setup.sh`.

**Add packages:** edit `packages`, rebuild.

**Add files:** drop them into `root/` — they're copied verbatim. `root/etc/hostname` becomes `/etc/hostname`.

**Add SSH key:**
```sh
cp ~/.ssh/id_ed25519.pub root/root/.ssh/authorized_keys
```

**Persist data across reboots:** add a mount to `root/etc/fstab`.

## How it works

`alpine-make-rootfs` builds a rootfs from the declarative files in `root/` + `packages` + `setup.sh`. The result is packed into a cpio initramfs. `ukify` combines kernel + initramfs + cmdline into one `.efi` file. UEFI loads it, done.

Based on [frood](https://github.com/FiloSottile/mostly-harmless/tree/main/frood) by Filippo Valsorda. CC0.
