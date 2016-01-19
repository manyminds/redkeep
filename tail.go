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
	isConnected bool
	config      Configuration
	session     *mgo.Session
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

	triggerDB := namespace[:p]
	triggerCollection := namespace[p+1:]

	operationType := dataset["op"]

	if command, ok := dataset["o"].(map[string]interface{}); ok {
		triggerID, _ := command["_id"].(bson.ObjectId)

		switch operationType {
		case "i":
			if namespace == "live.comment" {
				author := GetValue("author", command)

				id, okA := GetValue("$id", author).(bson.ObjectId)
				col, okB := GetValue("$ref", author).(string)
				db, okC := GetValue("$db", author).(string)

				if okA && okB && okC {
					session := t.session.Copy()
					session.SetMode(mgo.Strong, true)

					user := map[string]interface{}{}

					collection := session.DB(db).C(col)
					collection.FindId(id).One(&user)

					username := GetValue("username", user)

					collection = session.DB(triggerDB).C(triggerCollection)
					collection.Update(bson.M{"_id": triggerID}, bson.M{"$set": bson.M{"redkeep.metadata.username": username}})
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

//Tail will start an inifite look that tails the oplog
//as long as the channel does not get any input
func (t TailAgent) Tail(quit chan bool) error {
	session := t.session.Copy()
	defer session.Close()

	oplogCollection := session.DB("local").C("oplog.rs")

	startTime := bson.MongoTimestamp(time.Now().Unix())

	iter := oplogCollection.Find(
		bson.M{"ts": bson.M{"$gt": bson.MongoTimestamp(startTime)}},
	).LogReplay().Sort("$natural").Tail(requeryDuration)

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
	if !t.isConnected {
		session, err := mgo.Dial(t.config.Mongo.ConnectionURL)

		if err != nil {
			log.Println(err)
			return err
		}

		session.SetMode(mgo.Monotonic, true)
		t.session = session
	}

	return nil
}

//NewTailAgent will generate a new tail agent
func NewTailAgent(c Configuration) (*TailAgent, error) {
	agent := &TailAgent{config: c}
	err := agent.connect()
	return agent, err
}
