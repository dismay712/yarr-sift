---
title: Terminal
description: Run yarr as a server and access it from any browser.
weight: 2
---

## Download

Download the command line version from the [releases page](https://github.com/dismay712/yarr-sift/releases/latest). The file name does not contain `_gui`.

| OS | Architecture | Download |
|---|---|---|
| macOS | Apple Silicon | [yarr-sift_darwin_arm64.zip](https://github.com/dismay712/yarr-sift/releases/latest/download/yarr-sift_darwin_arm64.zip) |
| macOS | Intel | [yarr-sift_darwin_amd64.zip](https://github.com/dismay712/yarr-sift/releases/latest/download/yarr-sift_darwin_amd64.zip) |
| Windows | ARM64 | [yarr-sift_windows_arm64.zip](https://github.com/dismay712/yarr-sift/releases/latest/download/yarr-sift_windows_arm64.zip) |
| Windows | x86-64 | [yarr-sift_windows_amd64.zip](https://github.com/dismay712/yarr-sift/releases/latest/download/yarr-sift_windows_amd64.zip) |
| Linux | ARM64 | [yarr-sift_linux_arm64.zip](https://github.com/dismay712/yarr-sift/releases/latest/download/yarr-sift_linux_arm64.zip) |
| Linux | ARMv7 | [yarr-sift_linux_armv7.zip](https://github.com/dismay712/yarr-sift/releases/latest/download/yarr-sift_linux_armv7.zip) |
| Linux | x86-64 | [yarr-sift_linux_amd64.zip](https://github.com/dismay712/yarr-sift/releases/latest/download/yarr-sift_linux_amd64.zip) |

## Run

Start the server:

{{< code "command-line-run.sh" >}}

The server listens on `127.0.0.1:7070` by default. yarr stores the data automatically in the user config folder.

## Example

Run the server on all interfaces with password protection:

{{< code "command-line-auth.sh" >}}

Open `http://host:7070` in a browser and sign in with `alice` / `secret`.
