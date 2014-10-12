gorest
======

* copy and modification of gorest libary from code.google.com/p/gorest
* synched some things from https://github.com/fromkeith/gorest (not sure if it will get in synch enough to go to merge)

### Changes
* xsrftoken - altered the use of the token to support Authorization: Bearer <token> through the http header
* logging - put in a different log framework - adding logging of request/response with elapsed time (for integration with logstash)
* hypermedia - currently adding hypermedia decorators (applicatin/vnd.siren+json, application/hal+json in dev)
* swagger - generate swagger 1.2 doc during runtime (skeleton working, need more work on models section)

### Other things connected to the framework
* using Consul for service registry and k/v store
* using HAPorxy with Consul-HAProxy for dynamic service registration / load balancing / friendly URL addressing
* have hooked up logstash to logfiles produced from framework, dump in elastic search, and front-end with Kibana

Example with Siren 
```
// test_Dec.go
package main

import (
	"github.com/rmullinnix/gorest"
	"github.com/rmullinnix/logger"
	"net/http"
	"strconv"
)

type State struct {
	Name		string
	Value		string
}

//Service Definition
type ReferenceService struct {
	gorest.RestService `root:"/example/dectest" consumes:"application/json" produces:"application/vnd.siren+json"`
	getLookup    gorest.EndPoint `method:"GET"  path:"/lookup?{name:string}" output:"State"`
	getString    gorest.EndPoint `method:"GET"  path:"/string?{name:string}" output:"string"`
	getArray     gorest.EndPoint `method:"GET"  path:"/array?{name:string}" output:"[]State"`
}

func main() {
	listen := ":" + strconv.Itoa(8713)

	logger.Init("info")

	gorest.RegisterService(new(ReferenceService))

	http.Handle("/", gorest.Handle())
	http.ListenAndServe(listen, nil)
}

func (serv ReferenceService) GetLookup(name string) State {
	var states	State

	states.Name = name
	states.Value = "Decorator Test"

	return states
}

func (serv ReferenceService) GetString(name string) string {
	return name
}

func (serv ReferenceService) GetArray(name string) []State {
	list := make([]State, 4)
	list[0].Name = "Missouri"
	list[0].Value = "MO"
	list[1].Name = "Kansas"
	list[1].Value = "KS"
	list[2].Name = "Iowa"
	list[2].Value = "IA"
	list[3].Name = "Texas"
	list[3].Value = "TX"

	return list
}
```
