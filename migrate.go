package boltup

import (
	"encoding/binary"
	"fmt"

	"github.com/boltdb/bolt"
)

var (
	bucket = []byte("migrations")
	key    = []byte("version")
)

type Migration func(tx *bolt.Tx) error

func Up(db *bolt.DB, migrations ...Migration) error {
	tx, err := db.Begin(true)
	if err != nil {
		return errorf("failed to create transaction: %w", err)
	}
	defer tx.Rollback()

	libVersion := len(migrations)
	dbVersion, err := getVersion(tx)
	if err != nil {
		return errorf("failed to get db version: %w", err)
	}

	if dbVersion > libVersion {
		return errorf("db version %d is greater than library version %d", dbVersion, libVersion)
	}

	if dbVersion == libVersion {
		return nil
	}

	for i := dbVersion; i < libVersion; i++ {
		if err := migrations[i](tx); err != nil {
			return errorf("failed to migrate from version %d to %d: %w", i, i+1, err)
		}
	}

	if err := setVersion(tx, libVersion); err != nil {
		return errorf("failed to update migration version to %d: %w", libVersion, err)
	}

	if err := tx.Commit(); err != nil {
		return errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func getVersion(tx *bolt.Tx) (int, error) {
	b, err := tx.CreateBucketIfNotExists(bucket)
	if err != nil {
		return 0, err
	}
	v := b.Get(key)
	if v == nil {
		return 0, nil
	}
	return btoi(v), nil
}

func setVersion(tx *bolt.Tx, v int) error {
	b := tx.Bucket(bucket)
	return b.Put(key, itob(v))

}

func itob(v int) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	return buf
}

func btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}

func errorf(format string, a ...interface{}) error {
	return fmt.Errorf("boltup: "+format, a...)
}