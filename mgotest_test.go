// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mgotest_test

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/juju/mgo/v2"
	"github.com/juju/mgo/v2/bson"
	"gopkg.in/errgo.v1"

	"github.com/juju/mgotest"
)

type testDoc struct {
	ID string `bson:"_id"`
	X  int
}

func TestNew(t *testing.T) {
	mgotest.ResetGlobalState()
	c := qt.New(t)
	defer c.Done()
	db, err := mgotest.New()
	c.Assert(err, qt.Equals, nil)

	coll := db.C("collection")
	// Check that we can actually use it.
	err = coll.Insert(testDoc{ID: "foo", X: 99})
	c.Assert(err, qt.Equals, nil)

	var doc testDoc
	err = coll.Find(bson.M{"_id": "foo"}).One(&doc)
	c.Assert(err, qt.Equals, nil)
	c.Assert(doc, qt.Equals, testDoc{"foo", 99})

	// Create another one and check that it doesn't interfere.
	db1, err := mgotest.New()
	c.Assert(err, qt.Equals, nil)

	coll1 := db1.C("collection")
	err = coll1.Find(bson.M{"_id": "foo"}).One(&doc)
	c.Assert(err, qt.Equals, mgo.ErrNotFound)

	err = db1.Close()
	c.Assert(err, qt.Equals, nil)

	err = db.Close()
	c.Assert(err, qt.Equals, nil)

	// Connect again and check that the database has been deleted.
	db2, err := mgotest.New()
	c.Assert(err, qt.Equals, nil)
	defer db2.Close()

	// Create a *mgo.Database instance pointing at the original
	// database name.
	db3 := db2.Session.DB(db.Name)
	coll3 := db3.C("collection")
	err = coll3.Find(bson.M{"_id": "foo"}).One(&doc)
	c.Assert(err, qt.Equals, mgo.ErrNotFound)
}

func TestNewDisabled(t *testing.T) {
	mgotest.ResetGlobalState()
	c := qt.New(t)
	defer c.Done()
	c.Setenv("MGOTESTDISABLE", "1")
	db, err := mgotest.New()
	c.Assert(err, qt.ErrorMatches, `MongoDB testing is disabled`)
	c.Assert(errgo.Cause(err), qt.Equals, mgotest.ErrDisabled)
	c.Assert(db, qt.IsNil)

	// ensure that a second attempt returns the same cause.
	db, err = mgotest.New()
	c.Assert(err, qt.ErrorMatches, `MongoDB testing is disabled`)
	c.Assert(errgo.Cause(err), qt.Equals, mgotest.ErrDisabled)
	c.Assert(db, qt.IsNil)
}

func TestNewExclusive(t *testing.T) {
	mgotest.ResetGlobalState()
	c := qt.New(t)
	defer c.Done()
	db, err := mgotest.New()
	c.Assert(err, qt.Equals, nil)

	// Check that NewExclusive gives us a different session.
	db1, err := mgotest.NewExclusive()
	c.Assert(err, qt.Equals, nil)
	c.Assert(db1.Session, qt.Not(qt.Equals), db.Session)

	// Sanity check that New does give us the same session.
	db2, err := mgotest.New()
	c.Assert(err, qt.Equals, nil)
	c.Assert(db2.Session, qt.Equals, db.Session)
}

func TestNewErrorCausesImmediateReturnLater(t *testing.T) {
	mgotest.ResetGlobalState()
	c := qt.New(t)
	defer c.Done()

	// Set connection string to an invalid address.
	c.Setenv("MGOCONNECTIONSTRING", "0.1.2.3")
	c.Patch(mgotest.DialTimeout, 50*time.Millisecond)

	db, err := mgotest.New()
	c.Assert(err, qt.ErrorMatches, `cannot dial MongoDB: .*`)
	c.Assert(db, qt.IsNil)

	// Set connection string to a valid address, but it should
	// be ignored now because it failed once.
	c.Setenv("MGOCONNECTIONSTRING", "")

	db, err = mgotest.New()
	c.Assert(err, qt.ErrorMatches, `dial failed previously: cannot dial MongoDB: .*`)
}
