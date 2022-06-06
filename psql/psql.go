// Package psql implements an event sink backed by a PostgreSQL database.
package psql

import (
	"database/sql"
	_ "github.com/lib/pq"
)

const (
	driverName = "postgres"
)

// EventSink is an indexer backend providing the tx/block index services.  This
// implementation stores records in a PostgreSQL database using the schema
// defined in state/indexer/sink/psql/schema.sql.
type EventSink struct {
	store   *sql.DB
	chainID string
}

// NewEventSink constructs an event sink associated with the PostgreSQL
// database specified by connStr. Events written to the sink are attributed to
// the specified chainID.
func NewEventSink(connStr, chainID string) (*EventSink, error) {
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, err
	} else if err := db.Ping(); err != nil {
		return nil, err
	}

	return &EventSink{
		store:   db,
		chainID: chainID,
	}, nil
}

// DB returns the underlying Postgres connection used by the sink.
// This is exported to support testing.
func (es *EventSink) DB() *sql.DB { return es.store }

// Stop closes the underlying PostgreSQL database.
func (es *EventSink) Stop() error { return es.store.Close() }
