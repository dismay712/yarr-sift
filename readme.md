# Yarr Sift

**Yarr Sift** extends [yarr](https://github.com/nkanaev/yarr) with per-feed rules for extracting text from the original web pages linked in RSS entries. For example, keep `article .entry-content` and exclude `.ad, .related`. The reader applies your rules when feeds refresh.

[原文提取使用说明](doc/content/usage/overview.zh.md#定时提取原文) · [Article extraction guide](doc/content/usage/overview.en.md#scheduled-article-extraction)

Download a desktop or command-line build from [Yarr Sift Releases](https://github.com/dismay712/yarr-sift/releases/latest). For Docker on Linux amd64 or arm64, run `docker pull ghcr.io/dismay712/yarr-sift:latest`. To pin a version, replace `latest` with its release tag, such as `v0.1.0`.

To build from source, install Go 1.26 or later, Node.js with npm, Make, and a C compiler, then run `npm ci`, `make host`, and `./out/yarr`.

* Single binary, no dependencies
* Cross-platform - Windows / macOS / Linux
* Desktop app or Self-hosted
* SQLite or Postgres

![screenshot](etc/promo.png)

<div align="center">
  <a href="https://nkanaev.github.io/yarr/en/">upstream documentation</a> /
  <a href="https://github.com/dismay712/yarr-sift/releases">Yarr Sift releases</a>
    (<a href="https://github.com/dismay712/yarr-sift/releases.atom">feed</a>)
</div>
