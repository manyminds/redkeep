package redkeep

import (
	"encoding/json"
	"errors"

	validator "gopkg.in/go-playground/validator.v8"
)

//Configuration for red keep
type Configuration struct {
	Mongo   Mongo   `json:"mongo" validate:"required"`
	Watches []Watch `json:"watches" validate:"required,gt=0,dive"`
}

//Mongo is a config struct that changes the way the client
//connects.
//ConnectionURI should be a string identifying your cluster
//if you have to slaves and one master it would be like
//slave-01:27018,slave-02:27018,master:27018
//where slave-01 is either a hostname or an ip
type Mongo struct {
	ConnectionURI string `json:"connectionURI" validate:"required,gt=0"`
}

//Watch defines one watch that redkeep will do for you
type Watch struct {
	TrackCollection       string            `json:"trackCollection" validate:"required,gt=0"`
	TrackFields           []string          `json:"trackFields" validate:"required,min=1,dive,min=1"`
	TargetCollection      string            `json:"targetCollection" validate:"required,min=1"`
	TargetNormalizedField string            `json:"targetNormalizedField" validate:"required,min=1"`
	TriggerReference      string            `json:"triggerReference" validate:"required,min=1"`
	BehaviourSettings     BehaviourSettings `json:"behaviourSettings"`
}

//BehaviourSettings can define how one specific
//watch handles special cases
type BehaviourSettings struct {
	CascadeDelete bool `json:"cascadeDelete"`
}

//NewConfiguration loads a configuration from data
//if it is not valid json, it will return an error
func NewConfiguration(configData []byte) (*Configuration, error) {
	var config Configuration
	err := json.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	validate := validator.New(&validator.Config{TagName: "validate"})
	err = validate.Struct(config)

	if err != nil {
		return nil, getValidationError(err.(validator.ValidationErrors))
	}

	return &config, err
}

func getValidationError(allErrors validator.ValidationErrors) error {
	for _, e := range allErrors {
		switch e.Field {
		case "Watches":
			return errors.New("Please add atleast one entry in watches")
		case "ConnectionURI":
			return errors.New("Mongo configuration must be defined")
		case "TargetCollection":
			return errors.New("TargetCollection must not be empty")
		case "TriggerReference":
			return errors.New("TriggerReference must not be empty")
		case "TrackCollection":
			return errors.New("TrackCollection must not be empty")
		case "TrackFields":
			fallthrough
		case "TrackFields[0]":
			return errors.New("TrackFields must exactly have one non-empty field, more are currently not supported")
		case "TargetNormalizedField":
			return errors.New("TargetNormalizedField must not be empty")
		default:
			return allErrors
		}
	}

	return errors.New("Something went wrong")
}
