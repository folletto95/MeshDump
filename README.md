# MeshDump


MeshDump collects telemetry from Meshtastic nodes and exposes the data through
a small web interface. Data is typically ingested from an MQTT broker and the
telemetry history is kept in memory. It can optionally be persisted to a file.
From the browser you can choose which node to inspect and view line charts of
the available data types.


The software is written in **Go** so it can be compiled into a single
self-contained Windows binary while development and builds occur on Linux.

Set `MQTT_BROKER` to the broker URL and, if required, `MQTT_USERNAME` and
`MQTT_PASSWORD` for authentication. `MQTT_TOPIC` defaults to `#`.
Node names are discovered automatically from incoming telemetry. HTTP polling
of nodes is still available for testing by setting `NODES` to a comma separated
list of IP addresses.

If `DATA_FILE` is specified, telemetry is also saved to that path and reloaded
on startup so historical data is preserved across restarts.

## Building

Run `./build.sh` on a Linux machine with Docker installed. The script compiles
a self-contained Windows binary named `MeshDump.exe` using a Go Docker image.
