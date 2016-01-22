package redkeep

import (
	"strings"

	"gopkg.in/mgo.v2/bson"
)

func checkKey(hackstack []string, field string) bool {
	for _, b := range hackstack {
		if b == field {
			return true
		}

		if strings.HasPrefix(field, b+".") {
			return true
		}
	}

	return false
}

//BuildUpdateQuery generates the query
func BuildUpdateQuery(w Watch, command map[string]interface{}) bson.M {
	normalizingFields := bson.M{}
	for queryType, query := range command {
		if mappedQuery, ok := query.(map[string]interface{}); ok {
			for key, value := range mappedQuery {
				if checkKey(w.TrackFields, key) {
					normalizingFields[w.TargetNormalizedField+"."+key] = value
				}
			}
		}

		return bson.M{queryType: normalizingFields}
	}

	return bson.M{}
}
