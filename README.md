# MeshDump

MeshDump aims to collect telemetry from Meshtastic nodes, store the data and
present interactive graphs via a web interface. Nodes will be reachable by IP
address and selectable in the web UI together with the desired data types.

The software is written in **Go** so it can be compiled into a single
self-contained Windows binary while development and builds occur on Linux.

## Building

Run `./build.sh` on a Linux machine with Docker installed. This script uses a
Go Docker image to compile a self-contained Windows binary named
`MeshDump.exe`.
