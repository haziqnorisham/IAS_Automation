package pg

import "time"

type Hc_Device_Metadata struct {
	DeviceId    string  `db:"id"`
	DeviceEui   string  `db:"device_eui"`
	DeviceName  string  `db:"device_name"`
	Latitude    float64 `db:"latitude"`
	Longitude   float64 `db:"longitude"`
	CreatedDate string  `db:"created_date"`
}

// Ingest status constants.
const (
	IngestStatusUnprocessed = "unprocessed"
	IngestStatusProcessed   = "processed"
)

// HcRawIngest represents a raw ingest message stored in PostgreSQL.
type HcRawIngest struct {
	MessageID    int64     `db:"message_id"`
	Topic        string    `db:"topic"`
	Payload      string    `db:"payload"`
	DeviceID     *string   `db:"device_id"`
	IngestMethod string    `db:"ingest_method"`
	Status       string    `db:"status"`
	ReceivedAt   time.Time `db:"received_at"`
}

func (p *PostgresStorage) CreateHcSchemaIfNotExists() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS hc_device_metadata (
			id SERIAL PRIMARY KEY,
			device_eui VARCHAR(50) UNIQUE NOT NULL,
			device_name VARCHAR(100) NOT NULL,
			latitude DECIMAL(10, 8),
			longitude DECIMAL(11, 8),
			created_date DATE DEFAULT CURRENT_DATE
		);`,
		`CREATE TABLE IF NOT EXISTS hc_raw_ingest (
			message_id BIGSERIAL PRIMARY KEY,
			topic VARCHAR(255) NOT NULL,
			payload TEXT NOT NULL,
			device_id VARCHAR(50),
			ingest_method VARCHAR(20) NOT NULL DEFAULT 'mqtt',
			status VARCHAR(20) NOT NULL DEFAULT 'unprocessed',
			received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}

	for _, q := range queries {
		if _, err := p.DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// InsertRawIngest stores a raw ingest message into the hc_raw_ingest table.
func (p *PostgresStorage) InsertRawIngest(topic string, payload []byte, deviceID *string, ingestMethod string, status string) error {
	query := `INSERT INTO hc_raw_ingest (topic, payload, device_id, ingest_method, status) VALUES ($1, $2, $3, $4, $5);`
	_, err := p.DB.Exec(query, topic, string(payload), deviceID, ingestMethod, status)
	return err
}

// QueryRawIngest retrieves raw ingest records with an optional limit (default 100).
// Results are ordered by message_id descending so higher IDs (newer messages) appear first.
func (p *PostgresStorage) QueryRawIngest(limit int) ([]HcRawIngest, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT message_id, topic, payload, device_id, ingest_method, status, received_at FROM hc_raw_ingest ORDER BY message_id DESC LIMIT $1;`
	rows, err := p.DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []HcRawIngest
	for rows.Next() {
		var r HcRawIngest
		if err := rows.Scan(&r.MessageID, &r.Topic, &r.Payload, &r.DeviceID, &r.IngestMethod, &r.Status, &r.ReceivedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

func (p *PostgresStorage) GetAllDevices() ([]Hc_Device_Metadata, error) {
	query := `SELECT id, device_eui, device_name, latitude, longitude, created_date FROM hc_device_metadata;`
	rows, err := p.DB.Query(query)
	if err != nil {
		defer rows.Close()
		return nil, err
	}
	defer rows.Close()

	var devices []Hc_Device_Metadata

	for rows.Next() {
		var hc_device Hc_Device_Metadata
		err := rows.Scan(
			&hc_device.DeviceId,
			&hc_device.DeviceEui,
			&hc_device.DeviceName,
			&hc_device.Latitude,
			&hc_device.Longitude,
			&hc_device.CreatedDate,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, hc_device)
	}

	return devices, nil
}
