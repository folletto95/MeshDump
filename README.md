# MeshDump


MeshDump collects telemetry from Meshtastic nodes and exposes the data through
a small web interface. Data is typically ingested from an MQTT broker and the
telemetry history is kept in memory. It can optionally be persisted to a file.
The program contains a very small built-in MQTT client so it can connect to a
broker without requiring external Go modules or internet access.
From the browser you can choose which node to inspect and view line charts of
the available data types.


The software is written in **Go** so it can be compiled into a single
self-contained Windows binary while development and builds occur on Linux.

Set `MQTT_BROKER` to the broker URL and, if required, `MQTT_USERNAME` and
`MQTT_PASSWORD` for authentication. `MQTT_TOPIC` defaults to `#`.
Nodes appear in the interface as soon as they publish telemetry, so you do not
need to list them ahead of time. HTTP polling of nodes is only for testing and
can be enabled by setting `NODES` to a comma separated list of IP addresses.

If `DATA_FILE` is specified, telemetry and node metadata are stored in a small
SQLite database at that path. The file is created automatically and reloaded on
startup so historical data is preserved across restarts.

Set `DEBUG=1` to print additional information, including the list of nodes and
their names, to the terminal.



MeshDump automatically loads environment variables from a `.env` file. It first
looks in the current working directory and then in the directory containing the
executable. This lets you keep the configuration next to the binary when
running it outside of the source tree.


## Building

Run `./build.sh` on a Linux machine with Docker installed. The script compiles
a self-contained Windows binary named `MeshDump.exe` using a Go Docker image.
