package store

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	bucketProducts = "products"
	bucketMeta     = "meta"
)

type ProductRecord struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	InStock   bool      `json:"in_stock"`
	SeenAt    time.Time `json:"seen_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	db *bolt.DB
}

func New(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range []string{bucketProducts, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("init buckets: %w", err)
	}

	log.Printf("[Store] Opened BoltDB at %s", dbPath)
	return &Store{db: db}, nil
}

// ProductStatus holds the looked-up state for a single product.
type ProductStatus struct {
	IsNew      bool
	WasOutOfStock bool
}

// GetProductStatus checks whether a product is new and whether it was previously
// out of stock — in a single read transaction instead of two.
func (s *Store) GetProductStatus(productID string) (ProductStatus, error) {
	var status ProductStatus
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketProducts))
		data := b.Get([]byte(productID))
		if data == nil {
			status.IsNew = true
			return nil
		}
		var rec ProductRecord
		if err := json.Unmarshal(data, &rec); err != nil {
			return err
		}
		status.WasOutOfStock = !rec.InStock
		return nil
	})
	return status, err
}

// SaveProduct persists a single product record.
func (s *Store) SaveProduct(id, name string, inStock bool) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketProducts))
		return s.putProduct(b, id, name, inStock)
	})
}

// BatchSaveProducts persists many product records in a single write transaction
// instead of N separate ones. This reduces fsync calls from N to 1.
func (s *Store) BatchSaveProducts(products []ProductInput) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketProducts))
		for _, p := range products {
			if err := s.putProduct(b, p.ID, p.Name, p.InStock); err != nil {
				log.Printf("[Store] Failed to save product %s: %v", p.ID, err)
				// Continue saving the rest rather than aborting the entire batch
			}
		}
		return nil
	})
}

// ProductInput is the minimal info needed to save a product.
type ProductInput struct {
	ID      string
	Name    string
	InStock bool
}

func (s *Store) putProduct(b *bolt.Bucket, id, name string, inStock bool) error {
	rec := ProductRecord{
		ID:        id,
		Name:      name,
		InStock:   inStock,
		UpdatedAt: time.Now(),
	}

	existing := b.Get([]byte(id))
	if existing != nil {
		var old ProductRecord
		if err := json.Unmarshal(existing, &old); err == nil {
			rec.SeenAt = old.SeenAt
		}
	} else {
		rec.SeenAt = time.Now()
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return b.Put([]byte(id), data)
}

func (s *Store) CountProducts() (int, error) {
	var count int
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketProducts))
		count = b.Stats().KeyN
		return nil
	})
	return count, err
}

func (s *Store) Close() error {
	return s.db.Close()
}
