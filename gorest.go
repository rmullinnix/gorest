//Copyright 2011 Siyabonga Dlamini (siyabonga.dlamini@gmail.com). All rights reserved.
//
//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions
//are met:
//
//  1. Redistributions of source code must retain the above copyright
//     notice, this list of conditions and the following disclaimer.
//
//  2. Redistributions in binary form must reproduce the above copyright
//     notice, this list of conditions and the following disclaimer
//     in the documentation and/or other materials provided with the
//     distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE AUTHOR ``AS IS'' AND ANY EXPRESS OR
//IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
//OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
//IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
//PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS;
//OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
//WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR
//OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
//ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Notice: This code has been modified from its original source.
// Modifications are licensed as specified below.
//
// Copyright (c) 2014, fromkeith
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice, this
//   list of conditions and the following disclaimer in the documentation and/or
//   other materials provided with the distribution.
//
// * Neither the name of the fromkeith nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
// ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
///

package gorest

import (
	"encoding/json"
	"github.com/rmullinnix/logger"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	//"compress/gzip"
)

type GoRestService interface {
	ResponseBuilder() *ResponseBuilder
}

const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
)

type EndPointStruct struct {
	Name                 string
	RequestMethod        string
	Signiture            string
	muxRoot              string
	root                 string
	nonParamPathPart     map[int]string
	Params               []Param //path parameter name and position
	QueryParams          []Param
	signitureLen         int
	paramLen             int
	OutputType           string
	OutputTypeIsArray    bool
	outputTypeIsMap      bool
	PostdataType         string
	postdataTypeIsArray  bool
	postdataTypeIsMap    bool
	isVariableLength     bool
	parentTypeName       string
	MethodNumberInParent int
	role                 string
	overrideProducesMime string // overrides the produces mime type
	overrideConsumesMime string // overrides the produces mime type
	allowGzip 	     int // 0 false, 1 true, 2 unitialized
}

type restStatus struct {
	httpCode int
	reason   string //Especially for code in range 4XX to 5XX
	header	 string
}

func (err restStatus) String() string {
	return err.reason
}

type ServiceMetaData struct {
	Template     interface{}
	ConsumesMime string
	ProducesMime []string
	Root         string
	realm        string
	allowGzip    bool
}

var restManager *manager
var handlerInitialised bool

type manager struct {
	root		string
	serviceTypes 	map[string]ServiceMetaData
	endpoints    	map[string]EndPointStruct
	swaggerEP	string
}

func newManager() *manager {
	man := new(manager)
	man.serviceTypes = make(map[string]ServiceMetaData, 0)
	man.endpoints = make(map[string]EndPointStruct, 0)
	return man
}

//Registers a service on the rootpath.
//See example below:
//
//	package main
//	import (
// 	   "github.com/rmullinnix/gorest"
// 	   "github.com/rmullinnix/logger"
//	        "http"
//	)
//	func main() {
//	    logger.Init("info")
//	    gorest.RegisterService(new(HelloService)) //Register our service
//	    http.Handle("/",gorest.Handle())
// 	   http.ListenAndServe(":8787",nil)
//	}
//
//	//Service Definition
//	type HelloService struct {
//	    gorest.RestService `root:"/tutorial/"`
//	    helloWorld  gorest.EndPoint `method:"GET" path:"/hello-world/" output:"string"`
//	    sayHello    gorest.EndPoint `method:"GET" path:"/hello/{name:string}" output:"string"`
//	}
//	func(serv HelloService) HelloWorld() string{
// 	   return "Hello World"
//	}
//	func(serv HelloService) SayHello(name string) string{
//	    return "Hello " + name
//	}
func RegisterService(h interface{}) {
	RegisterServiceOnPath("", h)
}

//Registeres a service under the specified path.
//See example below:
//
//	package main
//	import (
//	    "github.com/rmullinnix/gorest"
//	        "http"
//	)
//	func main() {
//	    gorest.RegisterServiceOnPath("/rest/",new(HelloService)) //Register our service
//	    http.Handle("/",gorest.Handle())
//	    http.ListenAndServe(":8787",nil)
//	}
//
//	//Service Definition
//	type HelloService struct {
//	    gorest.RestService `root:"/tutorial/"`
//	    helloWorld  gorest.EndPoint `method:"GET" path:"/hello-world/" output:"string"`
//	    sayHello    gorest.EndPoint `method:"GET" path:"/hello/{name:string}" output:"string"`
//	}
//	func(serv HelloService) HelloWorld() string{
//	    return "Hello World"
//	}
//	func(serv HelloService) SayHello(name string) string{
//	    return "Hello " + name
//	}
func RegisterServiceOnPath(root string, h interface{}) {
	//We only initialise the handler management once we know gorest is being used to hanlde request as well, not just client.
	if !handlerInitialised {
		restManager = newManager()
		handlerInitialised = true
	}

	if root == "/" {
		root = ""
	}

	if root != "" {
		root = strings.Trim(root, "/")
		root = "/" + root
	}

	registerService(root, h)
}

//ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.
func (this manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url_, err := url.QueryUnescape(r.URL.RequestURI())

	message := "url: " + url_ + " method: " + r.Method + " "
	defer logger.Elapsed(time.Now(), message)

	if err != nil {
		logger.Warning.Println("Could not serve page: ", r.Method, r.URL.RequestURI(), "Error:", err)
		logger.SetResponseCode(400)
		w.WriteHeader(400)
		w.Write([]byte("Client sent bad request."))
		return
	}

	if r.Header.Get("Origin") != "" {
		if r.Method == OPTIONS {
			logger.Error.Println("CORS pre-flight OPTONS")
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
			return
		}
	}

	if url_ == _manager().swaggerEP {
		basePath := "http://" + r.Host + "/"
		doc := GetDocumentor("swagger")
		swagDoc := doc.Document(basePath, this.serviceTypes, this.endpoints)
		data, _ := json.Marshal(swagDoc)
		logger.SetResponseCode(http.StatusOK)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	} 

	authkey := r.Header.Get("Authorization")
	if len(authkey) > 0 {
		if strings.Contains(authkey, "Bearer") {
			w.Header().Set("Authorization", authkey)
		} else {
			authkey = ""
		}
	}

	if ep, args, queryArgs, _, found := getEndPointByUrl(r.Method, url_); found {
		ctx := new(Context)
		ctx.writer = w
		ctx.request = r
		ctx.xsrftoken = strings.TrimPrefix(authkey, "Bearer ")
		ctx.sessData.relSessionData = make(map[string]interface{})
		ctx.sessData.relSessionData["Host"] = r.Host

		rb := prepareServe(ctx, ep, args, queryArgs)

		rb.WritePacket()
	} else {
		logger.Warning.Println("Could not serve page, path not found: ", r.Method, url_)
		logger.SetResponseCode(http.StatusNotFound)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("The resource in the requested path could not be found."))
	}
}

func (man *manager) getType(name string) ServiceMetaData {

	return man.serviceTypes[name]
}
func (man *manager) addType(name string, i ServiceMetaData) string {
	for str, _ := range man.serviceTypes {
		if name == str {
			return str
		}
	}

	man.serviceTypes[name] = i
	return name
}
func (man *manager) addEndPoint(ep EndPointStruct) {
	man.endpoints[ep.RequestMethod+":"+ep.Signiture] = ep
}

//Registeres the function to be used for handling all requests directed to gorest.
func HandleFunc(w http.ResponseWriter, r *http.Request) {
	logger.Info.Println("Serving URL : ", r.Method, r.URL.RequestURI())
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error.Println("Internal Server Error: Could not serve page: ", r.Method, r.RequestURI)
			logger.Error.Println(rec)
			logger.Error.Printf("%s", debug.Stack())
			logger.SetResponseCode(http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	restManager.ServeHTTP(w, r)
}

//Runs the default "net/http" DefaultServeMux on the specified port.
//All requests are handled using gorest.HandleFunc()
func ServeStandAlone(port int) {
	http.HandleFunc("/", HandleFunc)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func _manager() *manager {
	return restManager
}

func Handle() manager {
	return *restManager
}

func getDefaultResponseCode(method string) int {
	switch method {
	case GET, PUT, DELETE:
		{
			return 200
		}
	case POST:
		{
			return 202
		}
	default:
		{
			return 200
		}
	}

	return 200
}
