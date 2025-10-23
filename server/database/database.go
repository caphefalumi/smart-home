package database

import (
	"context"
	"fmt"
	"log"

	"github.com/qiniu/qmgo"
)

// Database wraps the qmgo client and database
type Database struct {
	client *qmgo.Client
	db     *qmgo.Database
}

// Connect establishes a connection to MongoDB
func Connect(uri string) (*Database, error) {
	ctx := context.Background()

	client, err := qmgo.NewClient(ctx, &qmgo.Config{Uri: uri})
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %w", err)
	}

	// Test the connection
	if err := client.Ping(5); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database("smarthome")

	// Create indexes
	if err := createIndexes(db); err != nil {
		log.Printf("Warning: failed to create indexes: %v", err)
	}

	log.Println("✓ Connected to MongoDB")

	return &Database{
		client: client,
		db:     db,
	}, nil
}

// Close closes the database connection
func (d *Database) Close(ctx context.Context) error {
	if d.client != nil {
		return d.client.Close(ctx)
	}
	return nil
}

// GetCollection returns a collection
func (d *Database) GetCollection(name string) *qmgo.Collection {
	return d.db.Collection(name)
}

// createIndexes creates necessary database indexes
func createIndexes(db *qmgo.Database) error {
	// Skip index creation for now - can be added later if needed
	// MongoDB will still work without explicit indexes
	log.Println("✓ Database indexes initialization skipped")
	return nil
}
