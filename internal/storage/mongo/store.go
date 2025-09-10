// ... existing code ...
package mongostore

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/moseye/docinator/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Store wraps a MongoDB client and collection for document persistence.
type Store struct {
	enabled bool
	client  *mongo.Client
	coll    *mongo.Collection
}

// NewFromEnv initializes the store from env:
// - MONGODB_URI (required to enable; if empty, store is disabled)
// - MONGODB_DB (default: "docinator")
// - MONGODB_COLLECTION (default: "packages")
// Logging approach: use slog.Debug for start/success paths and slog.Error on errors,
// include operation label and duration for observability.
func NewFromEnv(ctx context.Context) (*Store, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		// Return a disabled store to allow graceful no-DB operation.
		slog.Debug("mongo: store disabled; no MONGODB_URI", "operation", "mongo_connect")
		return &Store{enabled: false}, nil
	}

	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "docinator"
	}
	collName := os.Getenv("MONGODB_COLLECTION")
	if collName == "" {
		collName = "packages"
	}

	// Debug: attempting connection and ping; measure duration for connect flow.
	start := time.Now()
	slog.Debug("mongo: connecting", "operation", "mongo_connect", "db", dbName, "collection", collName)

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		slog.Error("mongo: connect failed", "operation", "mongo_connect", "error", err)
		return nil, err
	}

	// Verify connectivity
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		slog.Error("mongo: ping failed", "operation", "mongo_connect", "error", err)
		_ = client.Disconnect(ctx)
		return nil, err
	}

	coll := client.Database(dbName).Collection(collName)
	slog.Debug("mongo: connected", "operation", "mongo_connect", "db", dbName, "collection", collName, "duration", time.Since(start))
	return &Store{
		enabled: true,
		client:  client,
		coll:    coll,
	}, nil
}

// Enabled reports whether the store is active.
func (s *Store) Enabled() bool {
	return s != nil && s.enabled
}

// Close disconnects the MongoDB client if enabled.
// Logging approach: always log intent and outcome with timing.
func (s *Store) Close(ctx context.Context) error {
	if !s.Enabled() {
		slog.Debug("mongo: close skipped; store disabled", "operation", "mongo_close")
		return nil
	}
	start := time.Now()
	slog.Debug("mongo: disconnecting", "operation", "mongo_close")
	if err := s.client.Disconnect(ctx); err != nil {
		slog.Error("mongo: disconnect failed", "operation", "mongo_close", "error", err)
		return err
	}
	slog.Debug("mongo: disconnected", "operation", "mongo_close", "duration", time.Since(start))
	return nil
}

// GetByID returns a stored document by its import path (_id) or nil if not found.
// Logging approach: log start, cache-like hit/miss semantics, errors, and timing.
func (s *Store) GetByID(ctx context.Context, id string) (*models.Document, error) {
	if !s.Enabled() {
		slog.Debug("mongo: get_by_id skipped; store disabled", "operation", "mongo_get_by_id", "id", id)
		return nil, errors.New("store disabled")
	}
	start := time.Now()
	slog.Debug("mongo: get_by_id", "operation", "mongo_get_by_id", "id", id)

	var doc models.Document
	err := s.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			slog.Debug("mongo: get_by_id miss", "operation", "mongo_get_by_id", "id", id, "duration", time.Since(start))
			return nil, nil
		}
		slog.Error("mongo: get_by_id failed", "operation", "mongo_get_by_id", "id", id, "error", err, "duration", time.Since(start))
		return nil, err
	}
	slog.Debug("mongo: get_by_id hit", "operation", "mongo_get_by_id", "id", id, "duration", time.Since(start))
	return &doc, nil
}

// Upsert replaces the document by _id or inserts it if missing.
// Logging approach: log start, success (with doc ID), errors, and timing.
func (s *Store) Upsert(ctx context.Context, doc *models.Document) error {
	if !s.Enabled() {
		slog.Debug("mongo: upsert skipped; store disabled", "operation", "mongo_upsert")
		return errors.New("store disabled")
	}
	if doc == nil || doc.ID == "" {
		slog.Error("mongo: upsert invalid document", "operation", "mongo_upsert")
		return errors.New("invalid document or missing ID")
	}

	filter := bson.M{"_id": doc.ID}

	// Pass the v2 options builder directly (implements options.Lister)
	start := time.Now()
	slog.Debug("mongo: upsert starting", "operation", "mongo_upsert", "id", doc.ID)
	_, err := s.coll.ReplaceOne(ctx, filter, doc, options.Replace().SetUpsert(true))
	if err != nil {
		slog.Error("mongo: upsert failed", "operation", "mongo_upsert", "id", doc.ID, "error", err, "duration", time.Since(start))
		return err
	}
	slog.Debug("mongo: upsert success", "operation", "mongo_upsert", "id", doc.ID, "duration", time.Since(start))
	return nil
}
