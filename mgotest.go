// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mgotest

import (
	"crypto/rand"
	"fmt"
	"os"

	"gopkg.in/errgo.v1"
	"gopkg.in/mgo.v2"
)

var ErrDisabled = errgo.New("MongoDB testing is disabled")

type Database struct {
	*mgo.Database
}

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
	if os.Getenv("MGOTESTDISABLE") != "" {
		return nil, ErrDisabled
	}
	connStr := os.Getenv("MGOCONNECTIONSTRING")
	if connStr == "" {
		connStr = "localhost"
	}
	session, err := mgo.Dial(connStr)
	if err != nil {
		return nil, errgo.Notef(err, "cannot dial MongoDB")
	}
	return &Database{
		Database: session.DB(randomDatabaseName()),
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
	return nil
}

func randomDatabaseName() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("cannot read random bytes: %v", err))
	}
	return fmt.Sprintf("go_test_%x", buf)
}
