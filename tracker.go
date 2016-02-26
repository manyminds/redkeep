package redkeep

import (
	"log"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//Tracker handles changes in the oplog
//it is a combination of CRUD Tracker
//Remove/Update/Create/Delete
//
type Tracker interface {
	RemoveTracker
	UpdateTracker
	InsertTracker
}

//RemoveTracker can handle removes
type RemoveTracker interface {
	HandleRemove(
		w Watch,
		command map[string]interface{},
		selector map[string]interface{},
	)
}

//UpdateTracker can handle updates
type UpdateTracker interface {
	HandleUpdate(
		w Watch,
		command map[string]interface{},
		selector map[string]interface{},
	)
}

//InsertTracker can handle inserts
type InsertTracker interface {
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
	if updateQuery == nil {
		return
	}

	selectQuery := bson.M{w.TriggerReference + ".$id": refID}
	_, err := collection.UpdateAll(selectQuery, updateQuery)
	if err != nil {
		log.Println("Query could not be executed successfully.")
	}
}

func (c changeTracker) HandleRemove(w Watch, command map[string]interface{}, selector map[string]interface{}) {
	log.Println("-- Remove is not yet implemented")
}

func (c changeTracker) HandleInsert(w Watch, command map[string]interface{}, originRef mgo.DBRef) {
	reference := GetValue(w.TriggerReference, command)
	if reference == nil {
		reference = GetValue("$set."+w.TriggerReference, command)
	}

	if reference == nil {
		return
	}

	ref, ok := getReference(reference, originRef.Database)
	if !ok {
		return
	}

	session := c.session.Copy()
	defer session.Close()

	user := map[string]interface{}{}

	collection := session.DB(ref.Database).C(ref.Collection)
	err := collection.FindId(ref.Id).One(&user)

	if err != nil {
		log.Println("User not found for update")
		return
	}

	query := BuildInsertQuery(w, user)
	if query == nil {
		log.Println("Empty query, need an update")
		return
	}

	collection = session.DB(originRef.Database).C(originRef.Collection)
	err = collection.Update(bson.M{"_id": originRef.Id.(bson.ObjectId)}, query)
	if err != nil {
		log.Println("Query could not be executed successfully." + err.Error())
		return
	}
}

//NewChangeTracker is the default tracker implementation of redkeep
func NewChangeTracker(session *mgo.Session) Tracker {
	return &changeTracker{session: session}
}
