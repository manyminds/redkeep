package redkeep

//
//
//
//
//
//
type Mongo struct {
	ListenDatabase string
	ConnectionURL  string
}

type Watch struct {
}

type Configuration struct {
	Mongo   Mongo   `json:"mongo"`
	Watches []Watch `json:"watches"`
}
