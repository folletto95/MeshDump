# MeshDump

MeshDump collects telemetry from Meshtastic nodes and exposes the data through
a small web interface. Nodes are polled at regular intervals and their
telemetry history is kept in memory. From the browser you can choose which node
to inspect and view line charts of the available data types.

The software is written in **Go** so it can be compiled into a single
self-contained Windows binary while development and builds occur on Linux.

Set the environment variable `NODES` to a comma separated list of node IP
addresses before running the binary so the application knows which nodes to
poll.

## Building

Run `./build.sh` on a Linux machine with Docker installed. The script compiles
a self-contained Windows binary named `MeshDump.exe` using a Go Docker image.
