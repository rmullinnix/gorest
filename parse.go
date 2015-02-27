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
	"github.com/rmullinnix/logger"
	"reflect"
	"strings"
	"strconv"
)

type argumentData struct {
	parameter Param
	data      string
}
type Param struct {
	positionInPath int
	Name           string
	TypeName       string
}

var aLLOWED_PAR_TYPES = []string{"string", "int", "int32", "int64", "bool", "float32", "float64", "[]string", "[]int"}

const (
	errorString_MarshalMimeType = "The Marshaller for mime-type:[%s], is not registered. Please register this type before registering your service."
	errorString_Realm = "The realm:[%s], is not registered. Please register this realm before registering your service."
	errorString_UnknownMethod = "Unknown method type:[%s] in endpoint declaration. Allowed types {GET,POST,PUT,DELETE,HEAD,OPTIONS}"
	errorString_EndpointDecl = "Endpoint declaration must have the tags 'method' and 'path' "
	errorString_StringMap = "Only string keyed maps e.g( map[string]... ) are allowed on the [%s] tag. Endpoint: %s"
	errorString_DuplicateQueryParam = "Duplicate Query Parameter name(%s) in REST path: %s"
	errorString_QueryParamConfig = "Please check that your Query Parameters are configured correctly for endpoint: %s"
	errorString_VariableLength = "Variable length endpoints can only have one parameter declaration: %s"
	errorString_RegisterSameMethod = "Can not register two endpoints with same request-method(%s) and same signature: %s VS %s"
	errorString_UniqueRoot = "Variable length endpoints can only be mounted on a unique root. Root already used: %s <> %s"
	errorString_Gzip = "Service has invalid gzip value. Defaulting to off settings! %s"
)

func prepServiceMetaData(root string, tags reflect.StructTag, i interface{}, name string) ServiceMetaData {
	md := new(ServiceMetaData)

	var tag		string

	if tag = tags.Get("root"); tag != "" {
		md.Root = tag
	}
	if root != "" {
		md.Root = root + md.Root
	}
	logger.Info.Println("All EndPoints for service [", name, "] , registered under root path: ", md.Root)

	md.ConsumesMime = make([]string, 0)
	if tag = tags.Get("consumes"); tag == "" {
		tag = Application_Json // Default
		md.ConsumesMime = append(md.ConsumesMime, tag)
	} else {
		cons := strings.Split(tag, ",")
		md.ConsumesMime = append(md.ConsumesMime, cons...)
	}

	for i := 0; i < len(md.ConsumesMime); i++ {
		mimeType := md.ConsumesMime[i]
		if !addMimeType(mimeType) {
			logger.Error.Fatalf(errorString_MarshalMimeType, mimeType)
		}
	}

	md.ProducesMime = make([]string, 0)
	if tag = tags.Get("produces"); tag == "" {
		tag = Application_Json // Default
		md.ProducesMime = append(md.ProducesMime, tag)
	} else {
		prods := strings.Split(tag, ",")
		md.ProducesMime = append(md.ProducesMime, prods...)
	}

	for i := 0; i < len(md.ProducesMime); i++ {
		mimeType := md.ProducesMime[i]
		if !addMimeType(mimeType) {
			logger.Error.Fatalf(errorString_MarshalMimeType, mimeType)
		}
	}

	if tag = tags.Get("realm"); tag != "" {
		md.realm = tag
		if GetAuthorizer(tag) == nil {
			logger.Error.Fatalf(errorString_Realm, tag)
		}
	}

	if tag := tags.Get("gzip"); tag != "" {
		b, err := strconv.ParseBool(tag)
		if err != nil {
			logger.Warning.Printf(errorString_Gzip, name)
			md.allowGzip = false
		} else {
			md.allowGzip = b
		}
	} else {
		md.allowGzip = false
	}

	md.Template = i
	return *md
}

func makeEndPointStruct(tags reflect.StructTag, serviceRoot string) EndPointStruct {

	methodMap := map[string]string {
		"GET":		GET,
		"PATCH":	PATCH,
		"POST":		POST,
		"PUT":		PUT,
		"DELETE":	DELETE,
		"HEAD":		HEAD,
		"OPTIONS":	OPTIONS,
	}

	ms := new(EndPointStruct)

	if tag := tags.Get("method"); tag != "" {
		ok := false
		if ms.RequestMethod, ok = methodMap[tag]; !ok {
			logger.Error.Fatalf(errorString_UnknownMethod, tag)
		}

		if tag := tags.Get("path"); tag != "" {
			serviceRoot = strings.TrimRight(serviceRoot, "/")
			ms.Signiture = serviceRoot + "/" + strings.Trim(tag, "/")
		} else {
			logger.Error.Fatalln(errorString_EndpointDecl)
		}

		if tag := tags.Get("output"); tag != "" {
			ms.OutputType = tag
			if strings.HasPrefix(tag, "[]") { //Check for slice/array/list types.
				ms.OutputTypeIsArray = true
				ms.OutputType = ms.OutputType[2:]
			}
			if strings.HasPrefix(tag, "map[") { //Check for map[string]. We only handle string keyed maps!!!

				if ms.OutputType[4:10] == "string" {
					ms.outputTypeIsMap = true
					ms.OutputType = ms.OutputType[11:]
				} else {
					logger.Error.Fatalf(errorString_StringMap, "output", ms.Signiture)
				}

			}
		}

		if tag := tags.Get("postdata"); tag != "" {
			ms.PostdataType = tag
			if strings.HasPrefix(tag, "[]") { //Check for slice/array/list types.
				ms.postdataTypeIsArray = true
				ms.PostdataType = ms.PostdataType[2:]
			}
			if strings.HasPrefix(tag, "map[") { //Check for map[string]. We only handle string keyed maps!!!

				if ms.PostdataType[4:10] == "string" {
					ms.postdataTypeIsMap = true
					ms.PostdataType = ms.PostdataType[11:]
				} else {
					logger.Error.Fatalf(errorString_StringMap, "postdata", ms.Signiture)
				}

			}
		}

		if tag := tags.Get("role"); tag != "" {
			ms.role = tag
		}

		ms.ConsumesMime = make([]string, 0)
		if tag = tags.Get("consumes"); tag == "" {
			tag = Application_Json // Default
			ms.ConsumesMime = append(ms.ConsumesMime, tag)
		} else {
			cons := strings.Split(tag, ",")
			ms.ConsumesMime = append(ms.ConsumesMime, cons...)
		}

		for i := 0; i < len(ms.ConsumesMime); i++ {
			mimeType := ms.ConsumesMime[i]
			if !addMimeType(mimeType) {
				logger.Error.Fatalf(errorString_MarshalMimeType, mimeType)
			}
		}

		ms.ProducesMime = make([]string, 0)
		if tag = tags.Get("produces"); tag == "" {
			tag = Application_Json // Default
			ms.ProducesMime = append(ms.ProducesMime, tag)
		} else {
			prods := strings.Split(tag, ",")
			ms.ProducesMime = append(ms.ProducesMime, prods...)
		}

		for i := 0; i < len(ms.ProducesMime); i++ {
			mimeType := ms.ProducesMime[i]
			if !addMimeType(mimeType) {
				logger.Error.Fatalf(errorString_MarshalMimeType, mimeType)
			}
		}

		if tag := tags.Get("gzip"); tag != "" {
			b, err := strconv.ParseBool(tag)
			if err != nil {
				logger.Warning.Printf(errorString_Gzip, ms.Name)
				ms.allowGzip = 2
			} else if b {
				ms.allowGzip = 1
			} else {
				ms.allowGzip = 0
			}
		} else {
			ms.allowGzip = 2
		}

		if tag := tags.Get("security"); tag != "" {
			ms.SecurityScheme = tag
		}

		parseParams(ms)
		return *ms
	}

	logger.Error.Fatalln(errorString_EndpointDecl)
	return *ms //Should not get here
}

func prepSecurityMetaData(tags reflect.StructTag) SecurityStruct {
	secDef := new(SecurityStruct)	
	secDef.Scope = make([]string, 0)

	if tag := tags.Get("mode"); tag != "" {
		secDef.Mode = tag
	}

	if tag := tags.Get("location"); tag != "" {
		secDef.Location = tag
	}

	if tag := tags.Get("name"); tag != "" {
		secDef.Name = tag
	}

	if tag := tags.Get("prefix"); tag != "" {
		secDef.Prefix = tag
	}

	if tag := tags.Get("flow"); tag != "" {
		secDef.Flow = tag
	}

	if tag := tags.Get("authURL"); tag != "" {
		secDef.AuthURL = tag
	}

	if tag := tags.Get("tokenURL"); tag != "" {
		secDef.TokenURL = tag
	}
	if tag := tags.Get("scope"); tag != "" {
		scope := strings.Split(tag, ",")
		secDef.Scope = append(secDef.Scope, scope...)
	}

	logger.Info.Println(secDef)
	return *secDef
}

func addMimeType(mimeType string) bool {
	if GetMarshallerByMime(mimeType) == nil {
		if strings.Contains(mimeType, "json") {
			RegisterMarshaller("json", NewJSONMarshaller())
		} else if strings.Contains(mimeType, "xml") {
			RegisterMarshaller("xml", NewXMLMarshaller())
		} else {
			return false
		}
	}

	return true
}

func parseParams(e *EndPointStruct) {
	e.Signiture = strings.Trim(e.Signiture, "/")
	e.Params = make([]Param, 0)
	e.QueryParams = make([]Param, 0)
	e.nonParamPathPart = make(map[int]string, 0)

	pathPart := e.Signiture
	queryPart := ""

	if i := strings.Index(e.Signiture, "?"); i != -1 {

		pathPart = e.Signiture[:i]
		//e.root = pathPart
		pathPart = strings.TrimRight(pathPart, "/")
		queryPart = e.Signiture[i+1:]

		//Extract Query Parameters

		for pos, str1 := range strings.Split(queryPart, "&") {
			if strings.HasPrefix(str1, "{") && strings.HasSuffix(str1, "}") {
				parName, typeName := getVarTypePair(str1, e.Signiture)

				for _, par := range e.QueryParams {
					if par.Name == parName {
						logger.Error.Fatalln("Duplicate Query Parameter name(" + parName + ") in REST path: " + e.Signiture)
					}
				}
				//e.QueryParams[len(e.QueryParams)] = Param{pos, parName, typeName}
				e.QueryParams = append(e.QueryParams, Param{pos, parName, typeName})
			} else {
				logger.Error.Fatalln("Please check that your Query Parameters are configured correctly for endpoint: " + e.Signiture)
			}
		}
	}

	if i := strings.Index(pathPart, "{"); i != -1 {
		e.root = pathPart[:i]
	} else {
		e.root = pathPart
	}

	//Extract Path Parameters
	for pos, str1 := range strings.Split(pathPart, "/") {
		e.signitureLen++

		if strings.HasPrefix(str1, "{") && strings.HasSuffix(str1, "}") { //This just ensures we re dealing with a varibale not normal path.

			parName, typeName := getVarTypePair(str1, e.Signiture)

			if parName == "..." {
				e.isVariableLength = true
				parName, typeName := getVarTypePair(str1, e.Signiture)
				e.Params = append(e.Params, Param{pos, parName, typeName})
				e.paramLen++
				break
			}
			for _, par := range e.Params {
				if par.Name == parName {
					logger.Error.Fatalln("Duplicate Path Parameter name(" + parName + ") in REST path: " + e.Signiture)
				}
			}

			e.Params = append(e.Params, Param{pos, parName, typeName})
			e.paramLen++
		} else {
			e.nonParamPathPart[pos] = str1

		}
	}

	e.root = strings.TrimRight(e.root, "/")

	if e.isVariableLength && e.paramLen > 1 {
		logger.Error.Fatalln("Variable length endpoints can only have one parameter declaration: " + pathPart)
	}

	for key, ep := range _manager().endpoints {
		if ep.root == e.root && ep.signitureLen == e.signitureLen && reflect.DeepEqual(ep.nonParamPathPart, e.nonParamPathPart) && ep.RequestMethod == e.RequestMethod {
			logger.Error.Fatalln("Can not register two endpoints with same request-method(" + ep.RequestMethod + ") and same signature: " + e.Signiture + " VS " + ep.Signiture)
		}
		if ep.RequestMethod == e.RequestMethod && pathPart == key {
			logger.Error.Fatalln("Endpoint already registered: " + pathPart)
		}
		if e.isVariableLength && (strings.Index(ep.root+"/", e.root+"/") == 0 || strings.Index(e.root+"/", ep.root+"/") == 0) && ep.RequestMethod == e.RequestMethod {
			logger.Error.Fatalln("Variable length endpoints can only be mounted on a unique root. Root already used: " + ep.root + " <> " + e.root)
		}
	}
}

func getVarTypePair(part string, sign string) (parName string, typeName string) {

	temp := strings.Trim(part, "{}")
	ind := 0
	if ind = strings.Index(temp, ":"); ind == -1 {
		logger.Error.Fatalln("Please ensure that parameter names(" + temp + ") have associated types in REST path: " + sign)
	}
	parName = temp[:ind]
	typeName = temp[ind+1:]

	if !isAllowedParamType(typeName) {
		logger.Error.Fatalln("Type " + typeName + " is not allowed for Path/Query-parameters in REST path: " + sign)
	}

	return
}

func isAllowedParamType(typeName string) bool {
	for _, s := range aLLOWED_PAR_TYPES {
		if s == strings.ToLower(typeName) {
			return true
		}
	}
	return false
}

func getEndPointByUrl(method string, url string) (EndPointStruct, map[string]string, map[string]string, string, bool) {
	pathPart := url
	queryPart := ""

	if i := strings.Index(url, "?"); i != -1 {
		pathPart = url[:i]
		queryPart = url[i+1:]
	}

	pathPart = strings.Trim(pathPart, "/")

	if ep, found := matchEndPoint(method, pathPart); found {

		var pathArgs	map[string]string

		if ep.isVariableLength {
			pathArgs = parseVarPathArgs(ep, pathPart)
		} else {
			pathArgs = parsePathArgs(ep, pathPart)
		}

		queryArgs, xsrft := parseQueryArgs(ep, queryPart)
	
		return *ep, pathArgs, queryArgs, xsrft, found
	}

	epRet := new(EndPointStruct)
	pathArgs := make(map[string]string, 0)
	queryArgs := make(map[string]string, 0)

	return *epRet, pathArgs, queryArgs, "", false
}

func matchEndPoint(method string, pathPart string) (ep *EndPointStruct, found bool) {
	// try exact match first
	pathMatch := method + ":" + pathPart
	if mEp, match := _manager().endpoints[pathMatch]; match {
		ep = &mEp
		return ep, true
	}

	totalParts := strings.Count(pathPart, "/") + 1
	for _, loopEp := range _manager().endpoints {

		if (strings.Index(pathPart+"/", loopEp.root+"/") == 0) && loopEp.RequestMethod == method {
			if loopEp.isVariableLength {
				ep = &loopEp
				return ep, true
			}

			if loopEp.signitureLen == totalParts {
				ep = &loopEp
				//We first make sure that the other parts of the path that are not parameters do actully match with the signature.
				//If not we exit. We do not have to carry on looking since we only allow one registration per root and length.
				for pos, name := range ep.nonParamPathPart {
					for upos, str1 := range strings.Split(pathPart, "/") {
						if upos == pos {
							if name != str1 {
								//Even though the beginning of the path matched, some other part didn't, keep looking.
								ep = nil
							}
							break
						}
					}

					if ep == nil {
						break
					}
				}

				if ep != nil {
					return ep, true
				}
			}
		}
	}
	return ep, false
}

func parseVarPathArgs(ep *EndPointStruct, pathPart string) map[string]string {
	pathArgs := make(map[string]string, 0)
	varsPart := strings.Trim(pathPart[len(ep.root):], "/")

	for upos, str1 := range strings.Split(varsPart, "/") {
		pathArgs[string(upos)] = strings.Trim(str1, " ")
	}
	return pathArgs
}

func parsePathArgs(ep *EndPointStruct, pathPart string) map[string]string {
	pathArgs := make(map[string]string, 0)
	//Extract Path Arguments
	for _, par := range ep.Params {
		for upos, str1 := range strings.Split(pathPart, "/") {

			if par.positionInPath == upos {
				pathArgs[par.Name] = strings.Trim(str1, " ")
				break
			}
		}
	}
	return pathArgs
}

func parseQueryArgs(ep *EndPointStruct, queryPart string) (map[string]string, string) {
	queryArgs := make(map[string]string, 0)

	xsrft := ""
	//Extract Query Arguments: These are optional in the query, so some or all of them might not be there.
	//Also, if they are there, they do not have to be in the same order they were sepcified in on the declaration signature.
	for _, str1 := range strings.Split(queryPart, "&") {
		if i := strings.Index(str1, "="); i != -1 {
			pName := str1[:i]
			dataString := str1[i+1:]
			if pName == XSXRF_PARAM_NAME {
				xsrft = strings.Trim(dataString, " ")
			} else {
//				for _, par := range ep.QueryParams {
//					if par.Name == pName {
						queryArgs[pName] = strings.Trim(dataString, " ")
//						break
//					}
//				}
			}

		}
	}

	return queryArgs, xsrft
}
