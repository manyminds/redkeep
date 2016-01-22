package redkeep

import (
	"errors"
	"log"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const requeryDuration = 1 * time.Second

//TailAgent the worker that tails the database
type TailAgent struct {
	config  Configuration
	session *mgo.Session
	tracker Tracker
}

func (t TailAgent) analyzeResult(dataset map[string]interface{}) {
	namespace, ok := dataset["ns"].(string)
	if namespace == "" || !ok {
		return
	}

	p := strings.Index(namespace, ".")
	if p == -1 {
		return
	}

	watches := t.config.Watches
	triggerDB := namespace[:p]
	triggerCollection := namespace[p+1:]
	operationType := dataset["op"]

	if command, ok := dataset["o"].(map[string]interface{}); ok {
		triggerID, _ := command["_id"].(bson.ObjectId)
		triggerRef := mgo.DBRef{
			Database:   triggerDB,
			Id:         triggerID,
			Collection: triggerCollection,
		}

		for _, w := range watches {
			switch operationType {
			case "i":
				if w.TargetCollection == namespace {
					t.tracker.HandleInsert(w, command, triggerRef)
				}
			case "u":
				if w.TrackCollection == namespace {
					if selector, ok := dataset["o2"].(map[string]interface{}); ok {
						t.tracker.HandleUpdate(w, command, selector)
					}
				}
			case "d":
				if w.TrackCollection == namespace {
					if selector, ok := dataset["o2"].(map[string]interface{}); ok {
						t.tracker.HandleRemove(w, command, selector)
					}
				}
			case "c":
				//system commands. We do not care.
			default:
				log.Printf("unsupported operation %s.\n", operationType)
				return
			}
			if w.TrackCollection == namespace {
				// updating stuff
			}
		}
	}
}

//getReference tries to create a reference from target
//returns true if valid, false otherwise
func getReference(target interface{}, originalDatabase string) (mgo.DBRef, bool) {
	id, okID := GetValue("$id", target).(bson.ObjectId)
	col, okRef := GetValue("$ref", target).(string)

	//database in references is an optional value
	db, okDb := GetValue("$db", target).(string)

	if !okDb {
		db = originalDatabase
	}

	return mgo.DBRef{Collection: col, Id: id, Database: db}, okID && okRef
}

//Tail will start an inifite look that tails the oplog
//as long as the channel does not get any input
//forceRescan (Default false) will update anything from the lowest oplog timestamp
//again. Can cause many redundant writes depending on your oplog size.
func (t TailAgent) Tail(quit chan bool, forceRescan bool) error {
	session := t.session.Copy()
	defer session.Close()

	oplogCollection := session.DB("local").C("oplog.rs")

	startTime := time.Now().Unix()
	if forceRescan {
		startTime = 0
	}

	query := oplogCollection.Find(bson.M{"ts": bson.M{"$gt": bson.MongoTimestamp(startTime)}})
	iter := query.LogReplay().Sort("$natural").Tail(requeryDuration)

	var lastTimestamp bson.MongoTimestamp
	for {
		select {
		case <-quit:
			log.Println("Agent stopped.")
			return nil
		default:
		}

		var result map[string]interface{}

		for iter.Next(&result) {
			lastTimestamp = result["ts"].(bson.MongoTimestamp)
			t.analyzeResult(result)
		}

		if iter.Err() != nil {
			return iter.Close()
		}

		if iter.Timeout() {
			continue
		}

		query := oplogCollection.Find(bson.M{"ts": bson.M{"$gt": lastTimestamp}})
		iter = query.LogReplay().Sort("$natural").Tail(requeryDuration)
	}

	iter.Close()

	return errors.New("Tailable cursor ended unexpectedly")
}

func (t *TailAgent) connect() error {
	log.Println("Connecting to", t.config.Mongo.ConnectionURI)
	session, err := mgo.Dial(t.config.Mongo.ConnectionURI)

	if err != nil {
		return err
	}

	session.SetMode(mgo.Strong, true)
	t.session = session
	t.tracker = NewChangeTracker(t.session)

	return nil
}

//NewTailAgent will generate a new tail agent
func NewTailAgent(c Configuration) (*TailAgent, error) {
	agent := &TailAgent{config: c}
	err := agent.connect()
	return agent, err
}
