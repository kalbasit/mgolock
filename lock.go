package mgolock

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Lock acquires the lock identified by name in the collection col. It will
// return true if the lock was successfully acquired.
func Lock(col *mgo.Collection, name string, ttl time.Duration) (bool, error) {
	// first try to extend the lock.
	ok, err := extendLock(col, name, ttl)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	// now try to acquire the first lock if we couldn't extend it.
	return reserveLock(col, name, ttl)
}

func reserveLock(col *mgo.Collection, name string, ttl time.Duration) (bool, error) {
	var update, query bson.M
	{
		now := time.Now()
		update = bson.M{
			"$set": bson.M{
				"reserved": now.Add(ttl),
				"owner":    getOwner(),
			},
		}
		query = bson.M{
			"_id":      name,
			"reserved": bson.M{"$lt": now},
		}
	}

	ci, err := col.Upsert(query, update)
	if err != nil {
		if mgo.IsDup(err) {
			return false, nil
		}
		return false, err
	}
	if ci.Updated > 0 || ci.UpsertedId == name {
		return true, nil
	}

	return false, nil
}

func extendLock(col *mgo.Collection, name string, ttl time.Duration) (bool, error) {
	var update, query bson.M
	{
		now := time.Now()
		update = bson.M{
			"$set": bson.M{
				"reserved": now.Add(ttl),
			},
		}
		query = bson.M{
			"_id":   name,
			"owner": getOwner(),
		}
	}
	ci, err := col.Upsert(query, update)
	if err != nil {
		if mgo.IsDup(err) {
			return false, nil
		}
		return false, err
	}
	if ci.Updated > 0 {
		return true, nil
	}

	return false, nil
}

func getOwner() string {
	host, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", host, os.Getpid())
}
