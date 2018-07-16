// Copyright 2018 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package mgotest_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/juju/mgotest"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type testDoc struct {
	ID string `bson:"_id"`
	X  int
}

func TestNew(t *testing.T) {
	c := qt.New(t)
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
