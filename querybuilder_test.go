package redkeep_test

import (
	. "github.com/manyminds/redkeep"
	"gopkg.in/mgo.v2/bson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Querybuilder tests", func() {
	Context("Validate queries", func() {

		var (
			w Watch
		)

		BeforeEach(func() {
			w = Watch{
				TrackFields:           []string{"username", "name"},
				TargetNormalizedField: "norm",
			}
		})

		It("will generate simple updates correctly", func() {
			command := map[string]interface{}{
				"$set": map[string]interface{}{
					"username": "nino",
				},
			}

			expected := bson.M{"$set": bson.M{"norm.username": "nino"}}
			actual := BuildUpdateQuery(w, command)
			Expect(actual).To(Equal(expected))
		})

		It("will generate update only whitelisted fields", func() {
			command := map[string]interface{}{
				"$set": map[string]interface{}{
					"username":   "nino",
					"otherField": "A",
				},
			}

			expected := bson.M{"$set": bson.M{"norm.username": "nino"}}
			actual := BuildUpdateQuery(w, command)
			Expect(actual).To(Equal(expected))
		})

		It("will generate nested updates correctly", func() {
			command := map[string]interface{}{
				"$set": map[string]interface{}{
					"name": map[string]interface{}{
						"firstName": "nino",
					},
				},
			}

			expected := bson.M{"$set": bson.M{"norm.name": map[string]interface{}{"firstName": "nino"}}}
			actual := BuildUpdateQuery(w, command)
			Expect(actual).To(Equal(expected))
		})

		It("will generate nested big updates correctly", func() {
			command := map[string]interface{}{
				"$set": map[string]interface{}{
					"name.firstName": "nino",
				},
			}

			expected := bson.M{"$set": bson.M{"norm.name.firstName": "nino"}}
			actual := BuildUpdateQuery(w, command)
			Expect(actual).To(Equal(expected))
		})
	})
})
