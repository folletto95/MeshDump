# MeshDump


MeshDump collects telemetry from Meshtastic nodes and exposes the data through
a small web interface. Data is typically ingested from an MQTT broker and the
telemetry history is kept in memory. It can optionally be persisted to a SQLite
database file.

The program uses the Eclipse Paho MQTT client library to connect to a broker
and subscribe to telemetry topics. Incoming messages can be JSON encoded
`Telemetry` structs or binary protobuf `MapReport` messages published by
Meshtastic nodes. Map reports are used to populate node metadata such as the
firmware version.

From the browser you can choose which node to inspect and view line charts of
the available data types.


The software is written in **Go** so it can be compiled into a single
self-contained Windows binary while development and builds occur on Linux.

Set `MQTT_BROKER` to the broker URL and, if required, `MQTT_USERNAME` and
`MQTT_PASSWORD` for authentication. `MQTT_TOPIC` defaults to `#`.
Nodes appear in the interface as soon as they publish telemetry, so you do not
need to list them ahead of time.

If `DATA_FILE` is specified, telemetry and node metadata are stored in a small
SQLite database at that path (for example `telemetry.db`). The file is created
automatically and reloaded on startup so historical data is preserved across
restarts.
Node metadata now includes the firmware version when available.

Set `DEBUG=1` to print additional information, including the list of nodes and
their names, to the terminal.



MeshDump automatically loads environment variables from a `.env` file. It first
looks in the current working directory and then in the directory containing the
executable. This lets you keep the configuration next to the binary when
running it outside of the source tree.


## Parsing logs

The helper `parse_log.go` extracts timestamp, node name and value from MeshDump log lines. Pipe a log through the program to obtain a tab-separated list of values:

```bash
grep "store: add" log.txt | go run parse_log.go
```

Both MQTT and store messages are matched by the regular expression used in the program.

## Building

Run `./build.sh` on a Linux machine with Docker installed. The script compiles
a self-contained Windows binary named `MeshDump.exe` using a Go Docker image.

## License

This project is licensed under the [MIT License](LICENSE).
