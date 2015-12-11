package mgolock

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func mgoTestSession() *mgo.Session {
	mongoURI := fmt.Sprintf("mongodb://127.0.0.1/test_%s", randString(20))
	mgoSession, err := mgo.Dial(mongoURI)
	if err != nil {
		log.Fatalf("Error connecting to Mongo, cannot run the tests: %s", err)
	}

	return mgoSession
}

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestLock(t *testing.T) {
	ms := mgoTestSession()
	ms.EnsureSafe(&mgo.Safe{})
	lc := ms.DB("").C("lock")
	defer func() {
		ms.DB("").DropDatabase()
		ms.Close()
	}()

	// try to grab the first lock
	ok, err := Lock(lc, "test", 1*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := true, ok; want != got {
		t.Errorf("Lock(): want %t got %t", want, got)
	}

	// try to extend the lock for 20 minutes
	ok, err = Lock(lc, "test", 20*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := true, ok; want != got {
		t.Errorf("Lock(): want %t got %t", want, got)
	}

	// change the hostname (imitating another process have the lock)
	lc.Update(bson.M{"owner": getOwner()}, bson.M{"$set": bson.M{"owner": "test"}})

	// try to grab the lock now
	ok, err = Lock(lc, "test", 1*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := false, ok; want != got {
		t.Errorf("Lock(): want %t got %t", want, got)
	}

	// or try to extend the lock for 20 minutes
	ok, err = Lock(lc, "test", 20*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := false, ok; want != got {
		t.Errorf("Lock(): want %t got %t", want, got)
	}
}
