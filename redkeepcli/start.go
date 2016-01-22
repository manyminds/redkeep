package main

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/manyminds/redkeep"
)

func main() {
	file, err := ioutil.ReadFile("example-configuration.json")
	if err != nil {
		log.Fatal(err)
	}
	config, err := redkeep.NewConfiguration(file)
	if err != nil {
		log.Fatal(err)
	}
	running := make(chan bool)
	agent, err := redkeep.NewTailAgent(*config)
	if err != nil {
		log.Fatal(err)
	}
	go agent.Tail(running, false)

	time.Sleep(3000 * time.Second)
	running <- false
}
