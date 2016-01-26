package redkeep_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"text/template"

	. "github.com/manyminds/redkeep"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var testConfigurationTemplate = `{
  "mongo": { 
    "connectionURI": "localhost:30000,localhost:30001,localhost:30002"
  }, 
  "watches": [ 
    {
      "trackCollection": "{{.Database}}.user",
      "trackFields": ["username", "gender"], 
      "targetCollection": "{{.Database}}.comment",
      "targetNormalizedField": "meta",
      "triggerReference": "user",
      "behaviourSettings": {
        "cascadeDelete": false
      }
    },
    {
      "trackCollection": "{{.Database}}.user",
      "trackFields": ["name", "username"], 
      "targetCollection": "{{.Database}}.answer",
      "targetNormalizedField": "meta",
      "triggerReference": "user",
      "behaviourSettings": {
        "cascadeDelete": false
      }
    }
  ]
}`

var _ = Describe("Tail", func() {
	var (
		running  chan bool
		database string
	)

	BeforeSuite(func() {
		randomDB := func() string {
			return fmt.Sprintf("redkeep_tests_%d", time.Now().UnixNano())
		}()
		tmp, err := template.New("config").Parse(testConfigurationTemplate)
		Expect(err).ToNot(HaveOccurred())
		var data []byte
		buffer := bytes.NewBuffer(data)
		err = tmp.Execute(buffer, struct{ Database string }{Database: randomDB})
		Expect(err).ToNot(HaveOccurred())
		data, err = ioutil.ReadAll(buffer)
		Expect(err).ToNot(HaveOccurred())
		config, err := NewConfiguration(data)
		Expect(err).ToNot(HaveOccurred())
		database = randomDB
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
		type comment struct {
			Text string
			User mgo.DBRef
			Meta map[string]interface{}
		}

		type answer struct {
			AnswerText string `bson:"answerText"`
			User       mgo.DBRef
			Meta       map[string]interface{}
		}

		var (
			db         *mgo.Session
			userOneRef mgo.DBRef
		)

		BeforeEach(func() {
			var err error
			db, err = mgo.Dial("localhost:30000,localhost:30001,localhost:30002")
			Expect(err).ToNot(HaveOccurred())
			userID := bson.ObjectIdHex("56a65494b204ccd1edc0b055")
			userOneRef = mgo.DBRef{
				Database:   database,
				Id:         userID,
				Collection: "user",
			}
		})

		It("Should update infos on insert correctly", func() {
			db.DB(database).C("user").Insert(
				bson.M{
					"_id":      userOneRef.Id,
					"username": "naan",
					"gender":   "male",
					"name": bson.M{
						"firstName": "Naan",
						"lastName":  "Waana",
					},
				},
			)

			db.DB(database).C("comment").Insert(
				bson.M{
					"text": "this is my first comment",
					"user": userOneRef,
				},
			)

			actual := comment{}
			time.Sleep(10 * time.Millisecond)
			db.Copy().DB(database).C("comment").Find(bson.M{}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("naan"))
			Expect(actual.Meta["gender"]).To(Equal("male"))
		})

		It("will also work with answers and different mapping", func() {
			db.DB(database).C("answer").Insert(&answer{AnswerText: "this is my answer", User: userOneRef})

			actual := answer{}
			time.Sleep(10 * time.Millisecond)
			db.Copy().DB(database).C("answer").Find(bson.M{"answerText": "this is my answer"}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("naan"))
			Expect(actual.Meta["name"]).To(Equal(map[string]interface{}{
				"firstName": "Naan",
				"lastName":  "Waana",
			}))
		})

		It("will then update usernames everywhere", func() {
			_, err := db.DB(database).C("user").UpdateAll(
				bson.M{"username": "naan"},
				bson.M{
					"$set": bson.M{
						"username": "anonym",
						"name": bson.M{
							"firstName": "Not",
							"lastName":  "Known",
						},
					},
				},
			)

			Expect(err).ToNot(HaveOccurred())
			time.Sleep(10 * time.Millisecond)
			actual := answer{}
			db.Copy().DB(database).C("answer").Find(bson.M{"answerText": "this is my answer"}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("anonym"))
			Expect(actual.Meta["name"]).To(HaveKey("firstName"))
			Expect(actual.Meta["name"]).To(HaveKey("lastName"))
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
