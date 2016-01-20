package redkeep_test

import (
	"io/ioutil"
	"strings"

	. "github.com/manyminds/redkeep"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var mongoMissingConfig = `
{
  "watches": [ 
    {
      "trackCollection": "live.user",
      "trackFields": ["username"], 
      "targetCollection": "live.comment",
      "targetNormalizedField": "normalizedUser",
      "triggerReference": "user",
      "behaviourSettings": {
        "cascadeDelete": false
      }
    }
  ]
}`

var watchesMissingConfig = `
{
  "mongo": { 
    "connectionURI": "localhost:30000,localhost:30001,localhost:30002"
  }, 
  "watches": [ 
  ]
}
`

var emptyConfig = `
	{
	}
`

var templateForTestsConfig = `
{
  "mongo": { 
    "connectionURI": "localhost:30000,localhost:30001,localhost:30002"
  }, 
  "watches": [ 
    {
      "trackCollection": "xAx",
      "trackFields": ["xBx"], 
      "targetCollection": "xCx",
      "targetNormalizedField": "xDx",
      "triggerReference": "xEx"
    }
  ]
}`

var _ = Describe("Config Testsuite", func() {
	Context("it will load and validate a config file", func() {
		It("will error with an empty config", func() {
			_, err := NewConfiguration([]byte(emptyConfig))
			Expect(err).To(HaveOccurred())
		})

		It("will error with missing mongo key", func() {
			_, err := NewConfiguration([]byte(mongoMissingConfig))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Mongo configuration must be defined"))
		})

		It("will error with missing watches entries", func() {
			_, err := NewConfiguration([]byte(watchesMissingConfig))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Please add atleast one entry in watches"))
		})

		It("will error with correct data but empty trackCollection", func() {
			_, err := NewConfiguration([]byte(strings.Replace(templateForTestsConfig, "xAx", "", 1)))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("TrackCollection must not be empty"))
		})

		It("will error with correct data but empty targetCollection", func() {
			_, err := NewConfiguration([]byte(strings.Replace(templateForTestsConfig, "xCx", "", 1)))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("TargetCollection must not be empty"))
		})

		It("will error with correct data but empty targetNormalizedField", func() {
			_, err := NewConfiguration([]byte(strings.Replace(templateForTestsConfig, "xDx", "", 1)))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("TargetNormalizedField must not be empty"))
		})

		It("will error with correct data but empty triggerReference", func() {
			_, err := NewConfiguration([]byte(strings.Replace(templateForTestsConfig, "xEx", "", 1)))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("TriggerReference must not be empty"))
		})

		It("will error with correct data but empty trigger", func() {
			_, err := NewConfiguration([]byte(strings.Replace(templateForTestsConfig, "xBx", "", 1)))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("TrackFields must exactly have one non-empty field, more are currently not supported"))
		})

		It("will error with correct data but multiple trackFields", func() {
			_, err := NewConfiguration([]byte(strings.Replace(templateForTestsConfig, `xBx"`, `", "xXx"`, 1)))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("TrackFields must exactly have one non-empty field, more are currently not supported"))
		})

		It("will load correctly", func() {
			file, err := ioutil.ReadFile("./example-configuration.json")
			Expect(err).ToNot(HaveOccurred())

			_, err = NewConfiguration(file)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
