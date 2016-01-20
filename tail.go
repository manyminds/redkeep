package redkeep

import (
	"errors"
	"log"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const requeryDuration = 2 * time.Second

//TailAgent the worker that tails the database
type TailAgent struct {
	config  Configuration
	session *mgo.Session
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
		for _, w := range watches {
			switch operationType {
			case "i":
				//insert only
				if w.TargetCollection == namespace {
					reference := GetValue(w.TriggerReference, command)

					ref, ok := getReference(reference, triggerDB)

					if ok {
						session := t.session.Copy()

						user := map[string]interface{}{}

						collection := session.DB(ref.Database).C(ref.Collection)
						collection.FindId(ref.Id).One(&user)

						username := GetValue("username", user)

						collection = session.DB(triggerDB).C(triggerCollection)
						collection.Update(bson.M{"_id": triggerID}, bson.M{"$set": bson.M{w.TargetNormalizedField: username}})
					}
				}
			case "u":
			case "d":
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

//GetValue works like this:
//from must be a selector like user.comment.author
//GetValue then looks recursively for that element
//therefore all of the following return values are possible
//map[string]interface{}
//nil
//string
//or basic mongodb types
func GetValue(from string, ds interface{}) interface{} {
	data, ok := ds.(map[string]interface{})
	if !ok {
		return nil
	}

	if index := strings.Index(from, "."); index != -1 {
		return GetValue(from[index+1:], data[from[:index]])
	}

	return data[from]
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

	startTime := bson.MongoTimestamp(time.Now().Unix())
	if forceRescan {
		startTime = bson.MongoTimestamp(0)
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

	return errors.New("Tailable has no more results")
}

func (t *TailAgent) connect() error {
	session, err := mgo.Dial(t.config.Mongo.ConnectionURI)

	if err != nil {
		log.Println(err)
		return err
	}

	session.SetMode(mgo.Strong, true)
	t.session = session

	return nil
}

//NewTailAgent will generate a new tail agent
func NewTailAgent(c Configuration) (*TailAgent, error) {
	agent := &TailAgent{config: c}
	err := agent.connect()
	return agent, err
}
