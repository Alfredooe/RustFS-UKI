# alfpine

Alpine Linux as a single-file UKI. No disk, no bootloader, no root partition. Everything in RAM.
Reboot = factory reset. Includes Docker.

## Usage

```sh
./alfpine run            # Build + test in QEMU
./alfpine build          # Build UKI
./alfpine flash /dev/sdb # Write to USB
```

Requires Docker. Login: `root` / `root`.

## Customize

- **Password:** replace hash in `setup.sh` (`openssl passwd -6`)
- **Packages:** edit `packages`
- **Files:** drop into `root/` (copied verbatim)
- **SSH:** `cp ~/.ssh/id_ed25519.pub root/root/.ssh/authorized_keys`
- **Persistence:** add mounts to `root/etc/fstab`

Based on [frood](https://github.com/FiloSottile/mostly-harmless/tree/main/frood). CC0.
