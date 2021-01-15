package boltup

import "github.com/boltdb/bolt"

// LimitedTx wraps the standard bolt tx but does not allow commit or rollback behavior
type LimitedTx struct {
	tx *bolt.Tx
}

// Bucket retrieves a bucket by name
func (l *LimitedTx) Bucket(name []byte) *bolt.Bucket {
	return l.tx.Bucket(name)
}

// CreateBucket creates a new bucket
func (l *LimitedTx) CreateBucket(name []byte) (*bolt.Bucket, error) {
	return l.tx.CreateBucket(name)
}

// CreateBucketIfNotExists creates a new bucket if it doesn't already exist
func (l *LimitedTx) CreateBucketIfNotExists(name []byte) (*bolt.Bucket, error) {
	return l.tx.CreateBucketIfNotExists(name)
}

// DeleteBucket deletes a bucket
func (l *LimitedTx) DeleteBucket(name []byte) error {
	return l.tx.DeleteBucket(name)
}

// Cursor creates a cursor associated with the root bucket
func (l *LimitedTx) Cursor() *bolt.Cursor {
	return l.tx.Cursor()
}

// ForEach executes a function for each bucket in the root
func (l *LimitedTx) ForEach(fn func(name []byte, b *bolt.Bucket) error) error {
	return l.tx.ForEach(fn)
}
