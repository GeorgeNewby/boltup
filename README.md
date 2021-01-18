# boltup

Boltup is a simple one way migration tool for boltdb inspired by the simplicity of https://github.com/BurntSushi/migration.
It performs a series on migrations on a bolt database as one transaction.
If a migration fails, the transaction is rolled back and nothing is changed.
Once a migration has been applied, it is perminant and should never be modified.
Internally, boltup creates a migrations bucket with a version field to keep track of the amount of migrations applied.

## Usage

A migration updates the database by one version.
As all migrations are ran in one transaction, a migration shouldn't be able to commit any changes early.
Therefore a LimitedTx is used which wraps the regular bolt transaction but does not allow commit or rollback functionality.

```go
type Migration func(tx *LimitedTx) error
```

Up takes in a bolt db instance and runs all of the supplied migrations in order.
If any migration fails an error is returned and everything is rolled back.

```go
func Up(db *bolt.DB, migrations ...Migration) error
```

## Example

Adding a new field with a default value

```go
package main

import (
    "encoding/json"
    "log"

    "github.com/GeorgeNewby/boltup"
    "github.com/boltdb/bolt"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Group string `json:"group"` // new field
}

func main() {
    db, err := bolt.Open("users.db", 0600, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    err = boltup.Up(db, createUserBucket, addGroupField)
    if err != nil {
        log.Fatal(err)
    }
}

func createUserBucket(tx *boltup.LimitedTx) error {
    _, err := tx.CreateBucket([]byte("users"))
    return err
}

func addGroupField(tx *boltup.LimitedTx) error {
    b := tx.Bucket([]byte("users"))

    b.ForEach(func(k, v []byte) error {
        var user User
        if err := json.Unmarshal(v, &user); err != nil {
            return err
        }

        user.Group = "A"

        data, err := json.Marshal(user)
        if err != nil {
            return err
        }

        if err := b.Put(k, data); err != nil {
            return err
        }
        return nil
    })
    return nil
}
```

