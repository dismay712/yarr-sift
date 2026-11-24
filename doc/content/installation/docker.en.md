---
title: Docker
description: Run yarr in a Docker container with a persistent volume.
weight: 3
---

The [Yarr Sift image on GitHub Container Registry](https://github.com/dismay712/yarr-sift/pkgs/container/yarr-sift) supports Linux amd64 and arm64. Use `latest` for the latest stable release or a version tag such as `v0.1.0` to pin your deployment.

Pull the image from the registry:

{{< code "docker-pull.sh" >}}

Run the container:

{{< code "docker-run.sh" >}}

The database is stored in the volume `yarr-data`.

## Docker Compose with PostgreSQL

To run yarr with PostgreSQL, use a `docker-compose.yml` file:

{{< code "docker-compose.yml" >}}

Start the services:

{{< code "docker-compose-up.sh" >}}
