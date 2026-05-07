package mqtt

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	client    mqtt.Client
	isRunning bool
	mu        sync.Mutex
	done      chan struct{}
)

// MessageHandler defines the callback signature for processing incoming MQTT messages.
type MessageHandler func(topic string, payload []byte)

// defaultHandler logs incoming messages at info level.
func defaultHandler(topic string, payload []byte) {
	slog.Info("MQTT message received",
		"topic", topic,
		"payload_length", len(payload),
		"payload", string(payload),
		"process", "mqtt_client",
	)
}

// ConnectAndSubscribe connects to the MQTT broker and subscribes to the configured topic.
// It launches a background goroutine to monitor the connection health.
func ConnectAndSubscribe(handler ...MessageHandler) error {
	mu.Lock()
	defer mu.Unlock()

	if isRunning {
		return fmt.Errorf("MQTT client is already running")
	}

	broker := os.Getenv("MQTT_BROKER_URL")
	if broker == "" {
		return fmt.Errorf("MQTT_BROKER_URL environment variable is not set")
	}

	clientID := os.Getenv("MQTT_CLIENT_ID")
	if clientID == "" {
		clientID = "ias_automation_client"
	}

	topic := os.Getenv("MQTT_TOPIC")
	if topic == "" {
		return fmt.Errorf("MQTT_TOPIC environment variable is not set")
	}

	qos := 0
	if q := os.Getenv("MQTT_QOS"); q != "" {
		fmt.Sscanf(q, "%d", &qos)
	}

	// Build a composite handler that chains all provided handlers plus the default logger.
	// The default handler (logging) always runs; any additional handlers are chained after it.
	var msgHandler MessageHandler
	if len(handler) == 0 {
		msgHandler = defaultHandler
	} else {
		msgHandler = func(topic string, payload []byte) {
			// Always log the incoming message
			defaultHandler(topic, payload)
			// Run all user-supplied handlers
			for _, h := range handler {
				if h != nil {
					h(topic, payload)
				}
			}
		}
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		slog.Error("MQTT connection lost", "error", err, "process", "mqtt_client")
	})
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		slog.Info("MQTT connected, subscribing to topic",
			"topic", topic,
			"qos", qos,
			"process", "mqtt_client",
		)
		if token := c.Subscribe(topic, byte(qos), func(c mqtt.Client, msg mqtt.Message) {
			msgHandler(msg.Topic(), msg.Payload())
		}); token.Wait() && token.Error() != nil {
			slog.Error("Failed to subscribe to MQTT topic",
				"topic", topic,
				"error", token.Error(),
				"process", "mqtt_client",
			)
		} else {
			slog.Info("Subscribed to MQTT topic", "topic", topic, "process", "mqtt_client")
		}
	})

	client = mqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker at %s: %w", broker, token.Error())
	}

	isRunning = true
	done = make(chan struct{})
	slog.Info("MQTT client started",
		"broker", broker,
		"client_id", clientID,
		"topic", topic,
		"process", "mqtt_client",
	)
	return nil
}

// StopClient disconnects from the MQTT broker gracefully.
func StopClient() {
	mu.Lock()
	defer mu.Unlock()

	if !isRunning {
		return
	}

	slog.Info("Stopping MQTT client", "process", "mqtt_client")
	client.Disconnect(250) // 250ms quiesce timeout
	isRunning = false
	close(done)
	slog.Info("MQTT client stopped", "process", "mqtt_client")
}

// IsRunning returns whether the MQTT client is currently connected.
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return isRunning
}
