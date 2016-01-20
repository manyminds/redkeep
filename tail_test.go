package redkeep_test

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	. "github.com/manyminds/redkeep"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tail", func() {
	Context("Test getValue", func() {
		It("will find the first value", func() {
			testReference := mgo.DBRef{
				Collection: "diff",
				Database:   "this",
				Id:         bson.ObjectIdHex("569e787d14b9802c9628b300"),
			}
			toTest := map[string]interface{}{
				"fish": map[string]interface{}{
					"dog": "cat",
				},
				"author": testReference,
				"tree":   nil,
			}

			actual := GetValue("fish.dog", toTest)
			Expect(actual).To(Equal("cat"))

			actual = GetValue("invalid", toTest)
			Expect(actual).To(BeNil())

			actual = GetValue("tree", toTest)
			Expect(actual).To(BeNil())

			actual = GetValue("author", toTest)
			Expect(actual).To(Equal(testReference))

			actual = GetValue("author.fisch.baum", toTest)
			Expect(actual).To(BeNil())

			actual = GetValue("fish", toTest)
			Expect(actual).To(Equal(map[string]interface{}{"dog": "cat"}))
		})
	})

	Context("Test basic connectivity", func() {
		PIt("should connect to master", func() {
			config := Configuration{
				Mongo: Mongo{
					ConnectionURI: "localhost:30002,localhost:30001,localhost:30000",
				},
			}
			running := make(chan bool)
			agent, err := NewTailAgent(config)
			Expect(err).ToNot(HaveOccurred())
			go agent.Tail(running)

			time.Sleep(1 * time.Second)
			running <- false
		})
	})
})
