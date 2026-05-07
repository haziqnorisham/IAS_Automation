package http

import (
	"encoding/json"
	ias_pg "ias/automation/db/pg"
	"log/slog"
	"net/http"
	"strconv"
)

// Non Http Utility functions related to HC schema and handlers
func SetupHcSchema() error {
	slog.Info("Creating HC Schema", "process", "hc_handler_main")
	ias_db := ias_pg.NewPostgresStorage(nil)
	defer ias_db.DB.Close()
	err := ias_db.CreateHcSchemaIfNotExists()
	if err != nil {
		slog.Error("Failed to create HC schema", "error", err)
		return err
	}
	slog.Info("HC schema created successfully", "process", "hc_handler_main")
	return nil
}

// HTTP Handlers related to HC schema
func GetAllDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	slog.Info("Retrieving all devices from HC schema", "process", "hc_handler_main")
	ias_db := ias_pg.NewPostgresStorage(nil)
	defer ias_db.DB.Close()
	devices, err := ias_db.GetAllDevices()
	if err != nil {
		slog.Error("Failed to retrieve devices", "error", err)
		return
	}
	slog.Info("Devices retrieved successfully", "process", "hc_handler_main")
	jsonData, err := json.Marshal(devices)
	if err != nil {
		slog.Error("Failed to marshal devices to JSON", "error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// GetRawIngest handles POST /api/get_raw_ingest?limit=50
// Returns all records from the hc_raw_ingest table, ordered by message_id DESC.
func GetRawIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	slog.Info("Querying raw ingest records", "limit", limit, "process", "hc_handler_main")

	ias_db := ias_pg.NewPostgresStorage(nil)
	defer ias_db.DB.Close()

	records, err := ias_db.QueryRawIngest(limit)
	if err != nil {
		slog.Error("Failed to query raw ingest", "error", err, "process", "hc_handler_main")
		http.Error(w, `{"error":"failed to query raw ingest"}`, http.StatusInternalServerError)
		return
	}

	slog.Info("Raw ingest records retrieved", "count", len(records), "process", "hc_handler_main")

	// Return empty array instead of null when no records
	if records == nil {
		records = []ias_pg.HcRawIngest{}
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		slog.Error("Failed to marshal raw ingest records", "error", err, "process", "hc_handler_main")
		http.Error(w, `{"error":"failed to marshal response"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
