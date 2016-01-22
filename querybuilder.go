package redkeep

import "gopkg.in/mgo.v2/bson"

//TODO this logic is buggy with nested updates
//on a $set.name.firstName it won't trigger
func buildUpdateQuery(w Watch, command map[string]interface{}) bson.M {
	normalizingFields := bson.M{}
	for _, field := range w.TrackFields {
		value := GetValue("$set."+field, command)
		if HasKey("$set."+field, command) {
			if value == nil {
				value = "null"
			}

			normalizingFields[w.TargetNormalizedField+"."+field] = value
		}
	}

	return bson.M{"$set": normalizingFields}
}
