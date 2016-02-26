package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/manyminds/redkeep"
)

func main() {
	configurationFilepath := flag.String("config", "configuration.json", "path to the configuration file")
	rescan := flag.Bool("rescan", false, "shall we start from the oplog beginnging?")
	flag.Parse()

	if configurationFilepath == nil {
		return
	}

	file, err := ioutil.ReadFile(*configurationFilepath)
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

	log.Println("Agent started.")
	agent.Tail(running, *rescan)
	running <- false
}
