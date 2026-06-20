# alfpine

Alpine Linux as a single-file UKI — no disk, no bootloader, no root partition.
Everything in RAM. Reboot = factory reset.

Bundles RustFS, 

## What you get

- **Console:** live RustFS server logs read only no shell
- **SSH:** key-only auth on port 2222
- **S3 API:** `http://<host>:9000`
- **RustFS Console:** `http://<host>:9001` (login: `rustfsadmin` / `rustfsadmin`)

## Usage

```sh
./alfpine build          # Build the UKI
./alfpine run            # Build + test in QEMU (ports 2222, 9000, 9001 forwarded)
./alfpine qemu           # Run latest image in QEMU
./alfpine flash /dev/sdb # Write to USB stick
```

Requires Docker to build

## Customize

- **SSH key:** `root/root/.ssh/authorized_keys` — baked into the image
- **Packages:** edit `packages`
- **Root files:** drop into `root/` (copied verbatim into the image)
- **Services:** init scripts in `root/etc/init.d/`, enable in `setup.sh`
- **RustFS creds:** defaults `rustfsadmin` / `rustfsadmin`, override via `RUSTFS_ROOT_USER` / `RUSTFS_ROOT_PASSWORD` env in `root/etc/init.d/rustfs`

## Architecture

```
UKI (.efi)
├── kernel (linux-lts)
├── initramfs
│   ├── Alpine base + packages (~150 MB)
│   └── RustFS binary (~285 MB, static musl)
└── cmdline: rdinit=/sbin/init
```

Everything loads into RAM at boot. No persistent state — mount a data disk at `/data`
for RustFS storage. Console tails RustFS logs; SSH is the only way in.
