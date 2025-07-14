# MeshDump


MeshDump collects telemetry from Meshtastic nodes and exposes the data through
a small web interface. Nodes can be polled over HTTP or data can be ingested
directly from an MQTT broker. The telemetry history is kept in memory and can
optionally be persisted to a file. From the browser you can choose which node
to inspect and view line charts of the available data types.


The software is written in **Go** so it can be compiled into a single
self-contained Windows binary while development and builds occur on Linux.

Set the environment variable `NODES` to a comma separated list of node IP
addresses before running the binary so the application knows which nodes to
poll.
To subscribe to an MQTT broker instead, set `MQTT_BROKER` to the broker URL and
optionally `MQTT_TOPIC` (defaults to `telemetry/#`).

If `DATA_FILE` is specified, telemetry is also saved to that path and reloaded
on startup so historical data is preserved across restarts.

## Building

Run `./build.sh` on a Linux machine with Docker installed. The script compiles
a self-contained Windows binary named `MeshDump.exe` using a Go Docker image.
