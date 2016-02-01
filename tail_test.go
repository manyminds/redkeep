package redkeep_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
		running      chan bool
		database     string
		answerString string
	)

	BeforeSuite(func() {
		rndDB := func() string {
			return fmt.Sprintf("redkeep_tests_%d", time.Now().UnixNano())
		}()

		database = rndDB

		running = make(chan bool)
		answerString = "this is my answer"

		var data []byte
		tmp, err := template.New("config").Parse(testConfigurationTemplate)
		Expect(err).ToNot(HaveOccurred())

		buffer := bytes.NewBuffer(data)
		err = tmp.Execute(buffer, struct{ Database string }{Database: database})
		Expect(err).ToNot(HaveOccurred())

		data, err = ioutil.ReadAll(buffer)
		Expect(err).ToNot(HaveOccurred())

		config, err := NewConfiguration(data)
		Expect(err).ToNot(HaveOccurred())

		agent, err := NewTailAgent(*config)
		Expect(err).ToNot(HaveOccurred())

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
			db           *mgo.Session
			userOneRef   mgo.DBRef
			userTwoRef   mgo.DBRef
			userThreeRef mgo.DBRef
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
			userIDTwo := bson.ObjectIdHex("56a87196ea3265eb5589c9c8")
			userTwoRef = mgo.DBRef{
				Database:   database,
				Id:         userIDTwo,
				Collection: "user",
			}
			userIDThree := bson.ObjectIdHex("56a87196ea3265eb5589c9dd")
			userThreeRef = mgo.DBRef{
				Database:   database,
				Id:         userIDThree,
				Collection: "user",
			}
		})

		It("Should update nothing with invalid fields", func() {
			userRef := mgo.DBRef{
				Database:   database,
				Id:         bson.NewObjectId(),
				Collection: "user",
			}
			db.DB(database).C("user").Insert(
				bson.M{
					"_id":      userRef.Id,
					"nickname": "captain america",
					"contact": bson.M{
						"firstName": "Steve",
						"lastName":  "Rogers",
					},
				},
			)

			db.DB(database).C("comment").Insert(
				bson.M{
					"text": "i am captain of a warmonger",
					"user": userRef.Id,
				},
			)

			c := comment{}
			time.Sleep(10 * time.Millisecond)
			db.Copy().DB(database).C("comment").Find(bson.M{"text": "i am captain of a warmonger"}).One(&c)

			Expect(c.Text).To(Equal("i am captain of a warmonger"))
			Expect(c.Meta).To(BeEmpty())
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
			db.Copy().DB(database).C("comment").Find(bson.M{"text": "this is my first comment"}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("naan"))
			Expect(actual.Meta["gender"]).To(Equal("male"))
		})

		It("Can handle user changes", func() {
			db.DB(database).C("user").Insert(
				bson.M{
					"_id":      userThreeRef.Id,
					"username": "songoku",
					"gender":   "male",
					"name": bson.M{
						"firstName": "Songoku",
						"lastName":  "Kakarotto",
					},
				},
			)

			_, err := db.DB(database).C("comment").UpdateAll(
				bson.M{
					"user": userOneRef,
				},
				bson.M{
					"$set": bson.M{
						"user": userThreeRef,
					},
				},
			)
			Expect(err).ToNot(HaveOccurred())

			actual := comment{}
			time.Sleep(10 * time.Millisecond)
			db.Copy().DB(database).C("comment").Find(bson.M{"text": "this is my first comment"}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("songoku"))
			Expect(actual.Meta["gender"]).To(Equal("male"))
			_, err = db.DB(database).C("comment").UpdateAll(
				bson.M{
					"user": userThreeRef,
				},
				bson.M{
					"$set": bson.M{
						"user": userOneRef,
					},
				},
			)
			Expect(err).ToNot(HaveOccurred())

		})

		It("will also work with answers and different mapping", func() {
			db.DB(database).C("answer").Insert(&answer{AnswerText: answerString, User: userOneRef})

			actual := answer{}
			time.Sleep(10 * time.Millisecond)
			db.Copy().DB(database).C("answer").Find(bson.M{"answerText": answerString}).One(&actual)

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
			db.Copy().DB(database).C("answer").Find(bson.M{"answerText": answerString}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("anonym"))
			Expect(actual.Meta["name"]).To(HaveKey("firstName"))
			Expect(actual.Meta["name"]).To(HaveKey("lastName"))
			Expect(actual.Meta["name"].(map[string]interface{})["firstName"]).To(Equal("Not"))
			Expect(actual.Meta["name"].(map[string]interface{})["lastName"]).To(Equal("Known"))
		})

		It("will work with upsert id on the user as well", func() {
			cl, err := db.DB(database).C("user").UpsertId(
				userOneRef.Id,
				bson.M{
					"$set": bson.M{
						"username": "ironman",
						"gender":   "male",
						"name": bson.M{
							"firstName": "Tony",
							"lastName":  "Stark",
						},
					},
				},
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(cl.Updated).To(Equal(1))

			time.Sleep(10 * time.Millisecond)
			actual := answer{}
			db.Copy().DB(database).C("answer").Find(bson.M{"answerText": answerString}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("ironman"))
			Expect(actual.Meta["name"]).To(HaveKey("firstName"))
			Expect(actual.Meta["name"]).To(HaveKey("lastName"))
			Expect(actual.Meta["name"].(map[string]interface{})["firstName"]).To(Equal("Tony"))
			Expect(actual.Meta["name"].(map[string]interface{})["lastName"]).To(Equal("Stark"))

			actualC := comment{}

			time.Sleep(10 * time.Millisecond)
			db.Copy().DB(database).C("comment").Find(bson.M{"text": "this is my first comment"}).One(&actualC)

			Expect(actualC.Meta["username"]).To(Equal("ironman"))
			Expect(actualC.Meta["gender"]).To(Equal("male"))
		})

		It("will work with upserts on the user as well", func() {
			cl, err := db.DB(database).C("user").Upsert(
				bson.M{
					"username": "ironman",
				},
				bson.M{
					"$set": bson.M{
						"username": "blackwidow",
						"gender":   "female",
						"name": bson.M{
							"firstName": "Natasha",
							"lastName":  "Romanoff",
						},
					},
				},
			)

			Expect(err).ToNot(HaveOccurred())
			Expect(cl.Updated).To(Equal(1))

			time.Sleep(10 * time.Millisecond)
			actual := answer{}
			db.Copy().DB(database).C("answer").Find(bson.M{"answerText": answerString}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("blackwidow"))
			Expect(actual.Meta["name"]).To(HaveKey("firstName"))
			Expect(actual.Meta["name"]).To(HaveKey("lastName"))
			Expect(actual.Meta["name"].(map[string]interface{})["firstName"]).To(Equal("Natasha"))
			Expect(actual.Meta["name"].(map[string]interface{})["lastName"]).To(Equal("Romanoff"))

			actualC := comment{}

			time.Sleep(10 * time.Millisecond)
			db.Copy().DB(database).C("comment").Find(bson.M{"text": "this is my first comment"}).One(&actualC)

			Expect(actualC.Meta["username"]).To(Equal("blackwidow"))
			Expect(actualC.Meta["gender"]).To(Equal("female"))
		})

		It("will also update multiple comments for multiple user updates", func() {
			err := db.DB(database).C("user").Insert(
				bson.M{
					"_id":      userTwoRef.Id,
					"username": "hawkeye",
					"gender":   "male",
					"name": bson.M{
						"firstName": "Clinton Francis",
						"lastName":  "Barton",
					},
				},
			)

			answerString = "I am hawkeye"

			db.DB(database).C("answer").Insert(&answer{AnswerText: answerString, User: userTwoRef})
			time.Sleep(10 * time.Millisecond)
			actual := answer{}
			db.Copy().DB(database).C("answer").Find(bson.M{"answerText": answerString}).One(&actual)

			Expect(actual.Meta["username"]).To(Equal("hawkeye"))
			Expect(actual.Meta["name"]).To(HaveKey("firstName"))
			Expect(actual.Meta["name"]).To(HaveKey("lastName"))
			Expect(actual.Meta["name"].(map[string]interface{})["firstName"]).To(Equal("Clinton Francis"))
			Expect(actual.Meta["name"].(map[string]interface{})["lastName"]).To(Equal("Barton"))
			Expect(err).ToNot(HaveOccurred())

			db.DB(database).C("comment").Insert(
				bson.M{
					"text": "hawkeye comment",
					"user": userTwoRef,
				},
			)

			actualComment := comment{}
			time.Sleep(10 * time.Millisecond)
			db.Copy().DB(database).C("comment").Find(bson.M{"text": "hawkeye comment"}).One(&actualComment)

			Expect(actualComment.Meta["username"]).To(Equal("hawkeye"))
			Expect(actualComment.Meta["gender"]).To(Equal("male"))

			cl, err := db.Copy().DB(database).C("user").UpdateAll(
				bson.M{},
				bson.M{
					"$set": bson.M{
						"gender": "confidential",
					},
				})

			Expect(err).ToNot(HaveOccurred())
			Expect(cl.Updated).To(BeNumerically(">", 1))
			time.Sleep(30 * time.Millisecond)

			iter := db.Copy().DB(database).C("comment").Find(bson.M{"meta": bson.M{"$exists": true}}).Iter()

			for iter.Next(&actualComment) {
				Expect(actualComment.Meta["gender"]).To(Equal("confidential"))
			}
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
				"$set": map[string]interface{}{
					"user": map[string]interface{}{
						"$id": "id",
					},
				},
			}

			actual := GetValue("fish.dog", toTest)
			Expect(actual).To(Equal("cat"))

			actual = GetValue("invalid", toTest)
			Expect(actual).To(BeNil())

			actual = GetValue("$set.user", toTest)
			Expect(map[string]interface{}{"$id": "id"}).To(Equal(actual))

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
