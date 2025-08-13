# Building the Debian Package under Kali

> Due to the age of golang in Debian stable, we only recommend building
> and running the Debian package on Kali Linux

podman run -it -v .:/src/BloodHound docker.io/kalilinux/kali-last-release
apt update && apt install devscripts sed jq wget dh-golang golang-any yarnpkg
cd /src/BloodHound && debuild -uc -b
(copy out deb packages from /src/)

## Testing

qemu-system-x86_64 -cdrom Downloads/kali-linux-2025.2-live-amd64.iso -m 8G --enable-kvm -usb -device usb-tablet \
  -drive file=fat:rw:(PATH_TO_DEB)
