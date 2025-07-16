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
If `MQTT_SERVER` is set to `internal` the program starts an embedded MQTT broker
listening on `MQTT_ADDRESS` (default `:1883`). Credentials default to
`meshdump`/`meshdump` when not provided.
Nodes appear in the interface as soon as they publish telemetry, so you do not
need to list them ahead of time.

If `DATA_FILE` is specified, telemetry and node metadata are stored in a small
SQLite database at that path (for example `telemetry.db`). The file is created
automatically and reloaded on startup so historical data is preserved across
restarts.
Node metadata now includes the firmware version when available.

Each node is identified by a unique `node_id`. The SQLite database keeps all
telemetry for a node grouped under this identifier. The `nodes` table uses
`node_id` as its primary key and the `telemetry` table references it so that all
measurements for the same node can be efficiently queried.

Set `DEBUG=1` to print additional information, including the list of nodes and
their names, to the terminal. Failed MQTT decode attempts are also logged with
the topic and a truncated payload.



MeshDump automatically loads environment variables from a `.env` file. It first
looks in the current working directory and then in the directory containing the
executable. This lets you keep the configuration next to the binary when
running it outside of the source tree.

Two optional variables control the Git identity used by `build.sh` when
committing build artifacts:

```
GIT_USER_NAME="MeshDump Builder"
GIT_USER_EMAIL=builder@example.com
```


## Parsing logs

The helper `parse_log.go` extracts timestamp, node name and value from MeshDump log lines. Pipe a log through the program to obtain a tab-separated list of values:

```bash
grep "store: add" log.txt | go run parse_log.go
```

Both MQTT and store messages are matched by the regular expression used in the program.

## Building

Run `./build.sh [os] [arch]` on a Linux machine with Docker installed and the
daemon running. The script compiles a binary for the specified target using a
Go Docker image. For example `./build.sh windows amd64` builds
`MeshDump.exe` for Windows. Passing `all` as the first argument builds both the
Linux and Windows binaries for the given architecture in one go, e.g.
`./build.sh all amd64`. Use `./build.sh rpi` to build images for Raspberry Pi

(armhf and arm64, excluding armv6). Omitting the architecture when using
`all` builds everything at once, including Raspberry Pi targets:
`./build.sh all` or `./build.sh all all`.

The script also commits the generated binaries to the Git repository

and pushes them if a remote is configured. If `user.name` or
`user.email` are not configured locally, the script sets them using the
`GIT_USER_NAME` and `GIT_USER_EMAIL` environment variables (configured
in `.env`). If those variables are absent, a default identity is used so
the commit succeeds.

## License

This project is licensed under the [MIT License](LICENSE).
