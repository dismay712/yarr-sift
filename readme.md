# Yarr Sift

**Yarr Sift** extends [yarr](https://github.com/nkanaev/yarr) with per-feed rules for extracting text from the original web pages linked in RSS entries. For example, keep `article .entry-content` and exclude `.ad, .related`. The reader applies your rules when feeds refresh.

[原文提取使用说明](doc/content/usage/overview.zh.md#定时提取原文) · [Article extraction guide](doc/content/usage/overview.en.md#scheduled-article-extraction)

To run this fork from source, install Go 1.26 or later, Node.js with npm, Make, and a C compiler, then run `npm ci`, `make host`, and `./out/yarr`.

* Single binary, no dependencies
* Cross-platform - Windows / macOS / Linux
* Desktop app or Self-hosted
* SQLite or Postgres

![screenshot](etc/promo.png)

<div align="center">
  <a href="https://nkanaev.github.io/yarr/en/">upstream documentation</a> /
  <a href="https://github.com/nkanaev/yarr/releases">upstream releases</a>
    (<a href="https://github.com/nkanaev/yarr/releases.atom">feed</a>)
</div>
