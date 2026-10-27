#!/usr/bin/env bash
# Build a graphical Debian package from already-built AuronQ Linux binaries.
# Does not touch consensus, mainnet data, wallets or published releases.
set -euo pipefail

version="1.7.14"
revision="1"
package="auronq-desktop"
arch="amd64"
root="${1:?pass packaging root directory}"
out="${2:?pass output .deb path}"
desktop_bin="${3:?pass AuronQ-Desktop binary}"
cli_bin="${4:?pass auronq CLI binary}"

mkdir -p "$root/DEBIAN" "$root/opt/auronq" \
  "$root/usr/share/applications" \
  "$root/usr/share/icons/hicolor/scalable/apps" \
  "$root/usr/share/doc/auronq"
install -m 0755 "$desktop_bin" "$root/opt/auronq/AuronQ-Desktop"
install -m 0755 "$cli_bin" "$root/opt/auronq/auronq"
install -m 0644 network.json bootstrap.json "$root/opt/auronq/"
install -m 0644 packaging/linux/auronq.desktop "$root/usr/share/applications/auronq.desktop"
install -m 0644 assets/brand/auronq-icon.svg "$root/usr/share/icons/hicolor/scalable/apps/auronq.svg"
install -m 0644 README-LINUX.md "$root/usr/share/doc/auronq/README-LINUX.md"
install -m 0644 LICENSE "$root/usr/share/doc/auronq/LICENSE"

cat > "$root/DEBIAN/control" <<EOF
Package: $package
Version: $version-$revision
Section: net
Priority: optional
Architecture: $arch
Maintainer: AuronQ contributors <noreply@github.com>
Depends: xdg-utils
Description: AuronQ graphical desktop and locally validating Mainnet node
 Launch from the application menu. The interface opens in the default browser.
 Blockchain and wallet data stay under the user's private configuration folder.
EOF

mkdir -p "$(dirname "$out")"
dpkg-deb --build --root-owner-group "$root" "$out"
