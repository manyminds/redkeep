package redkeep_test

import (
	"io/ioutil"
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
			file, err := ioutil.ReadFile("./example-configuration.json")
			Expect(err).ToNot(HaveOccurred())
			config, err := NewConfiguration(file)
			Expect(err).ToNot(HaveOccurred())
			running := make(chan bool)
			agent, err := NewTailAgent(*config)
			Expect(err).ToNot(HaveOccurred())
			go agent.Tail(running, true)

			time.Sleep(3000 * time.Second)
			running <- false
		})
	})
})
