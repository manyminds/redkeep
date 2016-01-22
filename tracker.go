package redkeep

import (
	"log"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Tracker handles changes in the oplog
type Tracker interface {
	HandleUpdate(
		w Watch,
		command map[string]interface{},
		selector map[string]interface{},
	)

	HandleRemove(
		w Watch,
		command map[string]interface{},
		selector map[string]interface{},
	)

	HandleInsert(
		w Watch,
		command map[string]interface{},
		originRef mgo.DBRef,
	)
}

type changeTracker struct {
	session *mgo.Session
}

func (c changeTracker) HandleUpdate(w Watch, command map[string]interface{}, selector map[string]interface{}) {
	if _, ok := command["$set"]; !ok {
		log.Println("Only set is implemented at the moment")
		return
	}
	session := c.session.Copy()
	defer session.Close()
	p := strings.Index(w.TargetCollection, ".")
	targetDB := w.TargetCollection[:p]
	targetCollection := w.TargetCollection[p+1:]
	collection := session.DB(targetDB).C(targetCollection)

	refID, ok := selector["_id"]
	if !ok {
		log.Println("No id found.")
		return
	}

	updateQuery := BuildUpdateQuery(w, command)
	_, err := collection.UpdateAll(bson.M{w.TriggerReference + ".$id": refID}, updateQuery)
	if err != nil {
		log.Printf("Could not update: %s using query %#v\n", err.Error(), updateQuery)
		log.Printf("Query: %#v\n", command)
	}
}

func (c changeTracker) HandleRemove(w Watch, command map[string]interface{}, selector map[string]interface{}) {
	log.Println("-- Remove is not yet implemented")
}

func (c changeTracker) HandleInsert(w Watch, command map[string]interface{}, originRef mgo.DBRef) {
	reference := GetValue(w.TriggerReference, command)

	if reference == nil {
		return
	}

	ref, ok := getReference(reference, originRef.Database)

	if ok {
		session := c.session.Copy()
		defer session.Close()

		user := map[string]interface{}{}

		collection := session.DB(ref.Database).C(ref.Collection)
		err := collection.FindId(ref.Id).One(&user)

		if err != nil {
			return
		}

		normalizingFields := bson.M{}
		for i, s := range w.TrackFields {
			normalizingFields[w.TargetNormalizedField+"."+s] = GetValue(w.TrackFields[i], user)
		}

		collection = session.DB(originRef.Database).C(originRef.Collection)
		collection.Update(bson.M{"_id": originRef.Id}, bson.M{"$set": normalizingFields})
	}

}

//NewChangeTracker is the default tracker implementation of redkeep
func NewChangeTracker(session *mgo.Session) Tracker {
	return &changeTracker{session: session}
}
