#!/usr/local/bin/bash

# Grab application version
VERSION=$(_release/bin/godrive-osx-x64 version | awk 'NR==1 {print $2}')

declare -a filenames
filenames=(
    "godrive-osx-x64"
    "godrive-osx-386"
    "godrive-osx-arm"
    "godrive-linux-x64"
    "godrive-linux-386"
    "godrive-linux-rpi"
    "godrive-linux-arm64"
    "godrive-linux-arm"
    "godrive-linux-mips64"
    "godrive-linux-mips64le"
    "godrive-linux-ppc64"
    "godrive-linux-ppc64le"
    "godrive-windows-386.exe"
    "godrive-windows-x64.exe"
    "godrive-dragonfly-x64"
    "godrive-freebsd-x64"
    "godrive-freebsd-386"
    "godrive-freebsd-arm"
    "godrive-netbsd-x64"
    "godrive-netbsd-386"
    "godrive-netbsd-arm"
    "godrive-openbsd-x64"
    "godrive-openbsd-386"
    "godrive-openbsd-arm"
    "godrive-solaris-x64"
    "godrive-plan9-x64"
    "godrive-plan9-386"
)

# Note: associative array requires bash 4+
declare -A descriptions
descriptions=(
    ["godrive-osx-x64"]="OS X 64-bit"
    ["godrive-osx-386"]="OS X 32-bit"
    ["godrive-osx-arm"]="OS X arm"
    ["godrive-linux-x64"]="Linux 64-bit"
    ["godrive-linux-386"]="Linux 32-bit"
    ["godrive-linux-rpi"]="Linux Raspberry Pi"
    ["godrive-linux-arm64"]="Linux arm 64-bit"
    ["godrive-linux-arm"]="Linux arm 32-bit"
    ["godrive-linux-mips64"]="Linux mips 64-bit"
    ["godrive-linux-mips64le"]="Linux mips 64-bit le"
    ["godrive-linux-ppc64"]="Linux PPC 64-bit"
    ["godrive-linux-ppc64le"]="Linux PPC 64-bit le"
    ["godrive-windows-386.exe"]="Window 32-bit"
    ["godrive-windows-x64.exe"]="Windows 64-bit"
    ["godrive-dragonfly-x64"]="DragonFly BSD 64-bit"
    ["godrive-freebsd-x64"]="FreeBSD 64-bit"
    ["godrive-freebsd-386"]="FreeBSD 32-bit"
    ["godrive-freebsd-arm"]="FreeBSD arm"
    ["godrive-netbsd-x64"]="NetBSD 64-bit"
    ["godrive-netbsd-386"]="NetBSD 32-bit"
    ["godrive-netbsd-arm"]="NetBSD arm"
    ["godrive-openbsd-x64"]="OpenBSD 64-bit"
    ["godrive-openbsd-386"]="OpenBSD 32-bit"
    ["godrive-openbsd-arm"]="OpenBSD arm"
    ["godrive-solaris-x64"]="Solaris 64-bit"
    ["godrive-plan9-x64"]="Plan9 64-bit"
    ["godrive-plan9-386"]="Plan9 32-bit"
)

# Markdown helpers
HEADER='### Downloads
| Filename               | Version | Description        | Shasum                                   |
|:-----------------------|:--------|:-------------------|:-----------------------------------------|'

ROW_TEMPLATE="| [{{name}}]({{url}}) | $VERSION | {{description}} | {{sha}} |"


# Print header
echo "$HEADER"

for name in ${filenames[@]}; do
    bin_path="_release/bin/$name"

    # Upload file
    url=$(godrive upload --share $bin_path | awk '/https/ {print $7}')

    # Shasum
    sha="$(shasum -b $bin_path | awk '{print $1}')"

    # Filename
    name="$(basename $bin_path)"

    # Render markdown row
    row=${ROW_TEMPLATE//"{{name}}"/$name}
    row=${row//"{{url}}"/$url}
    row=${row//"{{description}}"/${descriptions[$name]}}
    row=${row//"{{sha}}"/$sha}

    # Print row
    echo "$row"
done
