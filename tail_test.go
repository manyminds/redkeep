package redkeep_test

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	. "github.com/manyminds/redkeep"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tail", func() {
	/*
	 *  var (
	 *    running chan bool
	 *  )
	 *
	 *  BeforeSuite(func() {
	 *    file, err := ioutil.ReadFile(testConfiguration)
	 *    Expect(err).ToNot(HaveOccurred())
	 *    config, err := NewConfiguration(file)
	 *    Expect(err).ToNot(HaveOccurred())
	 *    agent, err := NewTailAgent(*config)
	 *    if err != nil {
	 *      log.Fatal(err)
	 *    }
	 *    running = make(chan bool)
	 *    go agent.Tail(running, false)
	 *  })
	 *
	 *  AfterSuite(func() {
	 *    running <- false
	 *  })
	 *
	 */
	Context("Test HasKey", func() {
		It("validate functionality", func() {
			toTest := map[string]interface{}{
				"fish": map[string]interface{}{
					"dog": "cat",
				},
				"tree": nil,
				"yellow": map[string]interface{}{
					"red": nil,
				},
			}
			actual := HasKey("blub", toTest)
			Expect(actual).To(Equal(false))

			actual = HasKey("fish", toTest)
			Expect(actual).To(Equal(true))

			actual = HasKey("fish.dog", toTest)
			Expect(actual).To(Equal(true))

			actual = HasKey("yellow", toTest)
			Expect(actual).To(Equal(true))

			actual = HasKey("yellow.red", toTest)
			Expect(actual).To(Equal(true))

			actual = HasKey("tree", toTest)
			Expect(actual).To(Equal(true))
		})
	})

	Context("test GetValue", func() {
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
})
