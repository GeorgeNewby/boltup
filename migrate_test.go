package boltup

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
)

func TestUp_NoMigrations(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	whenUpIsCalled(t, db)
	verifyNoMigrations(t, db)
}

func TestUp_SingleMigration(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	createTestBucket := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("test"))
		return err
	}

	whenUpIsCalled(t, db, createTestBucket)
	verifyBucketExists(t, db, []byte("test"))
	verifyMigrationVersion(t, db, 1)
}

func TestUp_MultipleMigrations(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	createTestBucket := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("test"))
		return err
	}

	updateValueToOne := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("test"))
		return b.Put([]byte("value"), []byte("1"))
	}

	updateValueToTwo := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("test"))
		return b.Put([]byte("value"), []byte("2"))
	}

	whenUpIsCalled(t, db, createTestBucket, updateValueToOne, updateValueToTwo)
	verifyBucketKeyContains(t, db, []byte("test"), []byte("value"), []byte("2"))
	verifyMigrationVersion(t, db, 3)
}

func TestUp_FailedMigration(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	createTestBucket := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("test"))
		return err
	}

	failedMigration := func(tx *bolt.Tx) error {
		return errors.New("Something went wrong")
	}

	verifyMigrationFails(t, db, createTestBucket, failedMigration)
	verifyBucketDoesNotExist(t, db, []byte("test"))
	verifyNoMigrations(t, db)
}

func TestUp_ExistingMigrations(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	createTestBucket := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("test"))
		return err
	}

	updateValueToOne := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("test"))
		return b.Put([]byte("value"), []byte("1"))
	}

	givenExistingMigrations(t, db, createTestBucket)
	whenUpIsCalled(t, db, createTestBucket, updateValueToOne)
	verifyBucketKeyContains(t, db, []byte("test"), []byte("value"), []byte("1"))
	verifyMigrationVersion(t, db, 2)
}

func TestUp_ExistingMigrations_FailedMigration(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	createTestBucket := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("test"))
		return err
	}

	failedMigration := func(tx *bolt.Tx) error {
		return errors.New("Something went wrong")
	}

	givenExistingMigrations(t, db, createTestBucket)
	verifyMigrationFails(t, db, createTestBucket, failedMigration)
	verifyBucketExists(t, db, []byte("test"))
	verifyMigrationVersion(t, db, 1)
}

func TestUp_ExistingMigrations_BadVersion(t *testing.T) {
	db, cleanup := newDB(t)
	defer cleanup()

	createTestBucket := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("test"))
		return err
	}

	updateValueToOne := func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("test"))
		return b.Put([]byte("value"), []byte("1"))
	}

	badMigration := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("bad"))
		return err
	}

	givenExistingMigrations(t, db, createTestBucket, updateValueToOne)
	verifyMigrationFails(t, db, badMigration)
	verifyBucketDoesNotExist(t, db, []byte("bad"))
	verifyMigrationVersion(t, db, 2)
}

// Test Assertions

func givenExistingMigrations(t *testing.T, db *bolt.DB, migrations ...Migration) {
	if err := Up(db, migrations...); err != nil {
		t.Fatalf("failed to create inital migrations: %v", err)
	}
}

func whenUpIsCalled(t *testing.T, db *bolt.DB, migrations ...Migration) {
	if err := Up(db, migrations...); err != nil {
		t.Fatalf("migration error: %v", err)
	}
}

func verifyMigrationFails(t *testing.T, db *bolt.DB, migrations ...Migration) {
	if err := Up(db, migrations...); err == nil {
		t.Fatal("expected migration to fail")
	}
}

func verifyBucketExists(t *testing.T, db *bolt.DB, bucket []byte) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			t.Errorf("%s bucket was not created", bucket)
		}
		return nil
	})
}

func verifyBucketDoesNotExist(t *testing.T, db *bolt.DB, bucket []byte) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			t.Errorf("%s bucket should not exist", bucket)
		}
		return nil
	})
}

func verifyBucketKeyContains(t *testing.T, db *bolt.DB, bucket []byte, key []byte, value []byte) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			t.Errorf("%s bucket does not exist", bucket)
		}
		v := b.Get(key)
		if !bytes.Equal(value, v) {
			t.Errorf("expected %v, got %v", value, v)
		}
		return nil
	})
}

func verifyMigrationVersion(t *testing.T, db *bolt.DB, version int) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("migrations"))
		if b == nil {
			t.Error("migrations bucket does not exist")
		}
		v := b.Get([]byte("version"))
		if !bytes.Equal(itob(version), v) {
			t.Errorf("migration version: expected %d, got %d", version, btoi(v))
		}
		return nil
	})
}

func verifyNoMigrations(t *testing.T, db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("migrations"))
		if b != nil {
			t.Errorf("migrations bucket should not exist")
		}
		return nil
	})
}

// Test Helpers

func newDB(t *testing.T) (db *bolt.DB, cleanup func()) {
	path, err := tempPath()
	if err != nil {
		t.Fatal(err)
	}
	db, err = bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	cleanup = func() {
		db.Close()
		os.Remove(path)
	}
	return
}

func tempPath() (string, error) {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}
