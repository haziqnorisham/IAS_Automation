package mqtt

import (
	"log/slog"
	"strings"

	ias_pg "ias/automation/db/pg"
)

// extractDeviceID attempts to extract a device identifier from the MQTT topic.
// Assumes topic pattern: sensors/{device_id}/... or {prefix}/{device_id}/...
// Returns nil if no device ID can be inferred.
func extractDeviceID(topic string) *string {
	parts := strings.Split(strings.Trim(topic, "/"), "/")
	// Try the second segment (index 1) as the device ID, e.g. sensors/{id}/...
	if len(parts) >= 2 && parts[1] != "" {
		return &parts[1]
	}
	return nil
}

// HcDbHandler creates a MessageHandler that stores every incoming MQTT payload
// into the hc_raw_ingest PostgreSQL table with ingest_method="mqtt" and status="unprocessed".
func HcDbHandler() MessageHandler {
	return func(topic string, payload []byte) {
		deviceID := extractDeviceID(topic)

		db := ias_pg.NewPostgresStorage(nil)
		defer db.DB.Close()

		if err := db.InsertRawIngest(topic, payload, deviceID, "mqtt", ias_pg.IngestStatusUnprocessed); err != nil {
			slog.Error("Failed to store raw ingest",
				"topic", topic,
				"device_id", deviceID,
				"error", err,
				"process", "mqtt_hc_handler",
			)
			return
		}

		slog.Info("Raw ingest stored",
			"topic", topic,
			"device_id", deviceID,
			"payload_length", len(payload),
			"process", "mqtt_hc_handler",
		)
	}
}
