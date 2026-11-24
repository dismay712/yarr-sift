docker run -d --name yarr \
  -p 7070:7070 \
  -v yarr-data:/data \
  ghcr.io/dismay712/yarr-sift:latest -db /data/yarr.db -addr 0.0.0.0:7070
