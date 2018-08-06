// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mgotest

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/errgo.v1"
	"gopkg.in/mgo.v2"
)

var ErrDisabled = errgo.New("MongoDB testing is disabled")

var dialTimeout = time.Second

type Database struct {
	*mgo.Database
	exclusive bool
}

var (
	sessionMu sync.Mutex
	session   *mgo.Session
	dialError error
)

// New connects to a MongoDB instance and returns a database
// instance that uses it. The database name is randomly chosen; all
// collections in the database will be removed when Close is called.
//
// The environment variable MGOCONNECTIONSTRING will
// be consulted to determine the connection string to use
// (see https://docs.mongodb.com/manual/reference/connection-string/
// or https://godoc.org/gopkg.in/mgo.v2#Dial for details of the format).
//
// If MGOCONNECTIONSTRING is empty, "localhost" will be used.
//
// If the environment variable MGOTESTDISABLE is non-empty,
// ErrDisabled will be returned.
func New() (*Database, error) {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	if session != nil {
		return &Database{
			Database: session.DB(randomDatabaseName()),
		}, nil
	}
	if dialError != nil {
		return nil, errgo.Notef(dialError, "dial failed previously")
	}
	db, err := NewExclusive()
	if err != nil {
		dialError = err
		return nil, errgo.Mask(err)
	}
	db.exclusive = false
	session = db.Database.Session
	return db, nil
}

// NewExclusive is like New except that it always returns a session
// is freshly created, rather than using one that's shared with other
// tests.
func NewExclusive() (*Database, error) {
	if os.Getenv("MGOTESTDISABLE") != "" {
		return nil, ErrDisabled
	}
	connStr := os.Getenv("MGOCONNECTIONSTRING")
	if connStr == "" {
		connStr = "localhost"
	}
	session, err := mgo.DialWithTimeout(connStr, dialTimeout)
	if err != nil {
		return nil, errgo.Notef(err, "cannot dial MongoDB")
	}
	return &Database{
		Database:  session.DB(randomDatabaseName()),
		exclusive: true,
	}, nil
}

// Close drops the database and closes its associated session.
func (db *Database) Close() error {
	// TODO it would be nice if this could check that there
	// were no currently open connections to the underlying
	// session.
	err := db.DropDatabase()
	if err != nil {
		return errgo.Notef(err, "cannot drop test database")
	}
	if db.exclusive {
		db.Session.Close()
	}
	return nil
}

func randomDatabaseName() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("cannot read random bytes: %v", err))
	}
	return fmt.Sprintf("go_test_%x", buf)
}
