// This program provides a sample application for using MongoDB with
// the mgo driver.
package main

import (
	"log"
	"sync"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	mgo "gopkg.in/mgo.v2"
	"labix.org/v2/mgo/bson"
)

var MgoSession *mgo.Session

func init() {
	if MgoSession == nil {
		var err error
		MgoSession, err = mgo.Dial("localhost")
		if err != nil {
			panic(err)
		}
	}
}

type Movies struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	title    string        `bson:"title"`
	year     string        `bson:"year"`
	director string        `bson:"director"`
	cast     string        `bson:"cast"`
	notes    string        `bson:"notes"`
}

func getPages() interface{} {
	raw, err := ioutil.ReadFile("./movies.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var m interface{}
	json.Unmarshal(raw, &m)

	return m

}

// main is the entry point for the application.
func main() {

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).

	MgoSession.SetMode(mgo.Monotonic, true)

	// Create a wait group to manage the goroutines.
	var waitGroup sync.WaitGroup

	// Perform 4 concurrent queries against the database.
	waitGroup.Add(4)
	for query := 0; query < 4; query++ {
		go RunQuery(query, &waitGroup, MgoSession)
	}

	// Wait for all the queries to complete.
	waitGroup.Wait()
	log.Println("All Queries Completed")
}

// RunQuery is a function that is launched as a goroutine to perform
// the MongoDB work.
func RunQuery(query int, waitGroup *sync.WaitGroup, MgoSession *mgo.Session) {
	// Decrement the wait group count so the program knows this
	// has been completed once the goroutine exits.
	defer waitGroup.Done()

	// Request a socket connection from the session to process our query.
	// Close the session when the goroutine exits and put the connection back
	// into the pool.
	sessionCopy := MgoSession.Clone()
	defer sessionCopy.Close()

	// Get a collection to execute the query against.
	collection := sessionCopy.DB("MovieDatabase").C("movies")

	log.Printf("RunQuery : %d : Executing\n", query)

	// Retrieve the inteface of movies.
	var m interface{}
	m = getPages()
	err := collection.Insert(m)
	if err != nil {
		log.Printf("RunQuery : ERROR : %s\n", err)
		return
	}

	log.Printf("RunQuery : %d : Count[%d]\n", query, m)
}
