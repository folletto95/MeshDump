module meshdump

go 1.20

require (
    github.com/eclipse/paho.mqtt.golang v1.3.5
)

// The project previously used a local stub implementation of the MQTT client
// to allow building without external network access. The real dependency is
// now referenced so the application can connect to a broker.
