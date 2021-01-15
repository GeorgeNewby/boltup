package boltup

import "github.com/boltdb/bolt"

type LimitedTx struct {
	tx *bolt.Tx
}

func (l *LimitedTx) Bucket(name []byte) *bolt.Bucket {
	return l.tx.Bucket(name)
}

func (l *LimitedTx) CreateBucket(name []byte) (*bolt.Bucket, error) {
	return l.tx.CreateBucket(name)
}

func (l *LimitedTx) CreateBucketIfNotExists(name []byte) (*bolt.Bucket, error) {
	return l.tx.CreateBucketIfNotExists(name)
}

func (l *LimitedTx) DeleteBucket(name []byte) error {
	return l.tx.DeleteBucket(name)
}

func (l *LimitedTx) Cursor() *bolt.Cursor {
	return l.tx.Cursor()
}

func (l *LimitedTx) ForEach(fn func(name []byte, b *bolt.Bucket) error) error {
	return l.tx.ForEach(fn)
}
