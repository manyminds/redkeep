package redkeep_test

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	. "github.com/manyminds/redkeep"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tail", func() {
	var (
		running chan bool
	)

	BeforeSuite(func() {
		file, err := ioutil.ReadFile(testConfiguration)
		Expect(err).ToNot(HaveOccurred())
		config, err := NewConfiguration(file)
		Expect(err).ToNot(HaveOccurred())
		agent, err := NewTailAgent(*config)
		if err != nil {
			log.Fatal(err)
		}
		running = make(chan bool)
		go agent.Tail(running, false)
	})

	AfterSuite(func() {
		running <- false
	})

	Context("Database testcases", func() {
		var (
			db *mgo.Session
		)

		BeforeEach(func() {
			var err error
			db, err = mgo.Dial("localhost:30000,localhost:30001,localhost:30002")
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should update infos on insert correctly", func() {

			userID := bson.ObjectIdHex("56a65494b204ccd1edc0b055")
			userRef := mgo.DBRef{
				Database:   "testing",
				Id:         userID,
				Collection: "user",
			}

			db.DB("testing").C("user").Insert(
				bson.M{
					"_id":      userID,
					"username": "naan",
					"gender":   "male",
					"name": bson.M{
						"firstName": "Naan",
						"lastName":  "Waana",
					},
				},
			)

			db.DB("testing").C("comment").Insert(
				bson.M{
					"text": "this is my first comment",
					"user": userRef,
				},
			)

			type comment struct {
				Text string
				User mgo.DBRef
				Meta map[string]interface{}
			}

			actual := comment{}
			time.Sleep(10 * time.Millisecond)
			db.Copy().DB("testing").C("comment").Find(bson.M{}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("naan"))
			Expect(actual.Meta["gender"]).To(Equal("male"))
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
