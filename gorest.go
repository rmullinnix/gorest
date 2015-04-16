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
	"encoding/base64"
	"github.com/rmullinnix/logger"
	"net/http"
	"net/url"
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
	encSigniture	     string
	muxRoot              string
	root                 string
	nonParamPathPart     map[int]string
	Params               []Param //path parameter name and position
	QueryParams          []Param
	signitureLen         int
	paramLen             int
	OutputType           string
	OutputTypeIsArray    bool
	OutputTypeIsMap      bool
	PostdataType         string
	postdataTypeIsArray  bool
	postdataTypeIsMap    bool
	isVariableLength     bool
	parentTypeName       string
	MethodNumberInParent int
	role                 string
	ProducesMime 	     []string // overrides the produces mime type
	ConsumesMime 	     []string // overrides the consumes mime type
	allowGzip 	     int // 0 false, 1 true, 2 unitialized
	SecurityScheme	     map[string][]string // must match one of securityDef
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
	ConsumesMime []string
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
	securityDef     map[string]SecurityStruct
	pathDict	map[string]int
	pathDictIndex	int
	allowOrigin	string
	allowOriginSet	bool
	swaggerEP	string
}

type SecurityStruct struct {
	Mode		string // basic, api_key or oauth2
	Description	string
	Location	string // header or query
	Name		string // name of query param or header element
	Prefix		string // if auth header has a prefix (e.g., "Bearer ")
	Flow		string
	AuthURL		string
	TokenURL	string
	Scope		[]string
}

type PathSecurity struct {
	Path		string
	Method		string
	Scope		[]string
}

func newManager() *manager {
	man := new(manager)
	man.serviceTypes = make(map[string]ServiceMetaData, 0)
	man.endpoints = make(map[string]EndPointStruct, 0)
	man.securityDef = make(map[string]SecurityStruct, 0)
	man.allowOriginSet = false

	man.pathDict = make(map[string]int, 0)
	man.pathDict["bool"] = 1
	man.pathDict["int"] = 2
	man.pathDict["string"] = 3
	man.pathDict["[]int"] = 4
	man.pathDict["[]string"] = 5
	man.pathDictIndex = 6

	return man
}

func SetAllowOrigin(origin string) {
	_manager().allowOrigin = origin
	_manager().allowOriginSet = true
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
	rb := new(ResponseBuilder)
	rb.ctx = new(Context)

	rb.ctx.writer = w
	rb.ctx.request = r
	rb.ctx.sessData.relSessionData = make(map[string]interface{})
	rb.ctx.sessData.relSessionData["Host"] = r.Host
	rb.ctx.sessStart = time.Now().Local()
	if _manager().allowOriginSet {
		rb.ctx.sessData.relSessionData["Origin"] = _manager().allowOrigin
	}
	defer rb.PerfLog()

	url_, err := url.QueryUnescape(r.URL.RequestURI())

	if err != nil {
		logger.Warning.Println("[gen] Could not serve page: ", r.Method, r.URL.RequestURI(), "Error:", err)
		rb.SetResponseCode(400)
		rb.WriteAndOveride([]byte("Client sent bad request."))
		return
	}

	if r.Header.Get("Origin") != "" {
		if r.Method == OPTIONS {
			rb.AddHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			rb.AddHeader("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, Location")
			if _manager().allowOriginSet {
				rb.AddHeader("Access-Control-Allow-Origin", _manager().allowOrigin)
			}
			rb.SetResponseCode(http.StatusOK)
			rb.WriteAndOveride([]byte(""))
			return
		}
	}

	if url_ == _manager().swaggerEP {
		basePath :=  _manager().root
		doc := GetDocumentor("swagger")
		swagDoc := doc.Document(basePath, this.serviceTypes, this.endpoints, this.securityDef)
		data, _ := json.Marshal(swagDoc)
		rb.SetResponseCode(http.StatusOK)
		rb.AddHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		rb.AddHeader("Access-Control-Allow-Origin", "*")
		rb.WriteAndOveride(data)
		return
	} 

	if ep, args, queryArgs, _, found := getEndPointByUrl(r.Method, url_); found {
		rb.ctx.xsrftoken = getAuthKey(ep.SecurityScheme, queryArgs, r, w)

		prepareServe(rb, ep, args, queryArgs)

		rb.WritePacket()
	} else {
		logger.Warning.Println("[gen] Could not serve page, path not found: ", r.Method, url_)
		rb.SetResponseCode(http.StatusNotFound)
		rb.WriteAndOveride([]byte("The resource in the requested path could not be found."))
	}
}

func getAuthKey(schemes map[string][]string, queryArgs map[string]string, r *http.Request, w http.ResponseWriter) string {
	authKey := ""

	// why we would have multiple schemes against an endpoint - don't know
	for scheme, _ := range schemes {
		// three modes - basic, api_key, oauth2
		if def, found := _manager().securityDef[scheme]; found {
			if def.Mode == "basic" {
				authKey = r.Header.Get("Authorization")
				if len(authKey) > 0 {
					authKey = strings.TrimPrefix(authKey, "Basic ")
					payload, _ := base64.StdEncoding.DecodeString(authKey)
					authKey = string(payload)
				}
			} else {
				location := def.Location
				name := def.Name
				prefix := def.Prefix
				if def.Mode == "oauth2"  {
					location = "header"
					name = "Authorization"
					prefix = "Bearer "
				}

				if location == "header" {
					authKey = r.Header.Get(name)
					if len(authKey) > 0 {
						w.Header().Set(name, authKey)
						if strings.Contains(authKey, prefix) {
							authKey = strings.TrimPrefix(authKey, prefix)
						}
					}
				} else if location == "query" {
					if authKey, found = queryArgs[name]; found {
						if strings.Contains(authKey, prefix) {
							authKey = strings.TrimPrefix(authKey, prefix)
						}
					}
				}
			}
			if len(authKey) > 0 {
				break
			}
		}
	}

	return authKey
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
	man.endpoints[ep.encSigniture] = ep
}

func (man *manager) addSecurityDefinition(name string, secDef SecurityStruct) {
	man.securityDef[name] = secDef
}

//Registeres the function to be used for handling all requests directed to gorest.
func HandleFunc(w http.ResponseWriter, r *http.Request) {
	logger.Info.Println("[gen] Serving URL : ", r.Method, r.URL.RequestURI())
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error.Println("Internal Server Error: Could not serve page: ", r.Method, r.RequestURI)
			logger.Error.Println(rec)
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
			return 201
		}
	default:
		{
			return 200
		}
	}

	return 200
}

func GetPathSecurity() []PathSecurity {
	eps := _manager().endpoints
	output := make([]PathSecurity, 0)

	for key, _ := range eps {
		var item	PathSecurity

		item.Path = cleanPath(eps[key].Signiture)
		item.Method = eps[key].RequestMethod
		if len(eps[key].SecurityScheme) > 0 {
			item.Scope = make([]string, 0)
			for _, scope := range eps[key].SecurityScheme {
				item.Scope = append(item.Scope, scope...)
			}
		}
		output = append(output, item)
	}
	return output
}

func cleanPath(inPath string) string {
        sig := strings.Split(inPath, "?")
        parts := strings.Split(sig[0], "{")

        path := parts[0]
        for i := 1; i < len(parts); i++ {
                pathVar := strings.Split(parts[i], ":")
                remPath := strings.Split(pathVar[1], "}")
                path = path + "{" + pathVar[0] + "}" + remPath[1]
        }

        return path
}

