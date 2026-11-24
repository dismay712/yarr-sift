---
title: Docker
description: 使用持久卷在 Docker 容器中运行 yarr。
weight: 3
---

[GitHub Container Registry 上的 Yarr Sift 镜像](https://github.com/dismay712/yarr-sift/pkgs/container/yarr-sift)支持 Linux amd64 和 arm64。使用 `latest` 获取最新稳定版，也可以用 `v0.1.0` 等版本标签固定部署版本。

从镜像仓库拉取镜像：

{{< code "docker-pull.sh" >}}

运行容器：

{{< code "docker-run.sh" >}}

数据库存储在数据卷 `yarr-data` 中。

## Docker Compose 与 PostgreSQL

要结合 PostgreSQL 运行 yarr，请使用 `docker-compose.yml` 文件：

{{< code "docker-compose.yml" >}}

启动服务：

{{< code "docker-compose-up.sh" >}}
