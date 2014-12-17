// test_Dec.go
package main

import (
	"github.com/rmullinnix/gorest"
	"github.com/rmullinnix/gorest/swagger"
	"github.com/rmullinnix/hypermedia"
	"github.com/rmullinnix/logger"
	"net/http"
	"strconv"
)

type States struct {
	Count	int
	State	[]State
}

type State struct {
	Name		string		`sw.description:"The full name of the state" sw.required:"true"`
	Value		string		`sw.description:"The two letter abbreviation of the state" sw.required:"true"`
}

//Service Definition
type ReferenceService struct {
	gorest.RestService `root:"/galaga/dectest" consumes:"application/json" 
		   	    produces:"application/json,application/vnd.siren+json,application/hal+json"
			    swagger:"/swagger" sw.apiVersion:"1.0"`
	getLookup    gorest.EndPoint `method:"GET"  path:"/lookup?{name:string}" output:"State"
			    sw.summary:"Retrieve state name and abbreviation"
			    sw.notes:"The input variable doesn't do anything in this example"
			    sw.nickname:"GetState"
			    sw.response:"{200:OK},{500:Internal Server Error}"`
	getTest      gorest.EndPoint `method:"GET"  path:"/string/{name:string}/{test:string}/multiple" output:"string"`
	getString    gorest.EndPoint `method:"GET"  path:"/string?{name:string}" output:"string"`
	getArray     gorest.EndPoint `method:"GET"  path:"/array?{name:string}" output:"States"`
}

type StatesHypermedia struct{
	hypermedia.Entity	`class:"States" href:"/galaga/dectest/states"`
	newstate	hypermedia.Action	`class:"State" method:"POST" href:"/galaga/dectest/state"`
	first		hypermedia.Link	`class:"State" href:"/contra/secuirty/states?page=first"`
	last		hypermedia.Link	`class:"State" href:"/contra/secuirty/states?page=last"`
	next		hypermedia.Link	`class:"State" href:"/contra/secuirty/states?page={page}+1"`
	prev		hypermedia.Link	`class:"State" href:"/contra/secuirty/states?page={page}-1"`
	Users		hypermedia.Link	`class:"State" href:"/contra/secuirty/roles"`
	Roles		hypermedia.Link	`class:"Role" href:"/contra/secuirty/users"`
}

type StateHypermedia struct {
	hypermedia.Entity	`class:"State" href:"/galaga/dectest/state/{Value}"`
	edit		hypermedia.Action	`class:"State" method:"PUT" href:"/galaga/dectest/state/{Value}"`
	delete		hypermedia.Action	`class:"State" method:"DELETE" href:"/galaga/dectest/state/{Value}"`
	self		hypermedia.Link	`class:"State" href:"/galaga/dectest/state/{Value}"`
	States		hypermedia.Link	`class:"State" href:"/galaga/dectest/states"`
}

func main() {
	listen := ":" + strconv.Itoa(8713)

	logger.Init("info")

	gorest.RegisterDocumentor("swagger", swagger.NewSwaggerDocumentor())
	hypermedia.NewHypermediaDecorator()
	dec := gorest.Decorator{hypermedia.Decorate}
	gorest.RegisterHypermedia(&dec)
	gorest.RegisterService(new(ReferenceService))
	hypermedia.RegisterEntity(new(StateHypermedia))
	hypermedia.RegisterEntity(new(StatesHypermedia))
	hypermedia.AddAccessRights("States", "public", "read")
	hypermedia.AddAccessRights("State", "public", "create,read,delete")

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

func (serv ReferenceService) GetTest(name string, test string) string {
	return name + test
}

func (serv ReferenceService) GetArray(name string) States {
	var states	States

	states.State = make([]State, 4)
	states.State[0].Name = "Missouri"
	states.State[0].Value = "MO"
	states.State[1].Name = "Kansas"
	states.State[1].Value = "KS"
	states.State[2].Name = "Iowa"
	states.State[2].Value = "IA"
	states.State[3].Name = "Texas"
	states.State[3].Value = "TX"

	states.Count = 4
	return states
}
