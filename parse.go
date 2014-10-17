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
	parameter param
	data      string
}
type param struct {
	positionInPath int
	name           string
	typeName       string
}

var aLLOWED_PAR_TYPES = []string{"string", "int", "int32", "int64", "bool", "float32", "float64"}

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
func prepServiceMetaData(root string, tags reflect.StructTag, i interface{}, name string) serviceMetaData {
	md := new(serviceMetaData)

	var tag		string

	if tag = tags.Get("root"); tag != "" {
		md.root = tag
	}
	if root != "" {
		md.root = root + md.root
	}
	logger.Info.Println("All EndPoints for service [", name, "] , registered under root path: ", md.root)

	if tag = tags.Get("consumes"); tag == "" {
		tag = Application_Json // Default
	}
	md.consumesMime = tag

	if !addMimeType(tag) {
		logger.Error.Fatalf(errorString_MarshalMimeType, tag)
	}

	md.producesMime = make([]string, 0)
	if tag = tags.Get("produces"); tag == "" {
		tag = Application_Json // Default
		md.producesMime = append(md.producesMime, tag)
	} else {
		prods := strings.Split(tag, ",")
		md.producesMime = append(md.producesMime, prods...)
	}

	for i := 0; i < len(md.producesMime); i++ {
		mimeType := md.producesMime[i]
		if !addMimeType(mimeType) {
			logger.Error.Fatalf(errorString_MarshalMimeType, tag)
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

	md.template = i
	return *md
}

func makeEndPointStruct(tags reflect.StructTag, serviceRoot string) endPointStruct {

	methodMap := map[string]string {
		"GET":		GET,
		"PATCH":	PATCH,
		"POST":		POST,
		"PUT":		PUT,
		"DELETE":	DELETE,
		"HEAD":		HEAD,
		"OPTIONS":	OPTIONS,
	}

	ms := new(endPointStruct)

	if tag := tags.Get("method"); tag != "" {
		ok := false
		if ms.requestMethod, ok = methodMap[tag]; !ok {
			logger.Error.Fatalf(errorString_UnknownMethod, tag)
		}

		if tag := tags.Get("path"); tag != "" {
			serviceRoot = strings.TrimRight(serviceRoot, "/")
			ms.signiture = serviceRoot + "/" + strings.Trim(tag, "/")
		} else {
			logger.Error.Fatalln(errorString_EndpointDecl)
		}

		if tag := tags.Get("output"); tag != "" {
			ms.outputType = tag
			if strings.HasPrefix(tag, "[]") { //Check for slice/array/list types.
				ms.outputTypeIsArray = true
				ms.outputType = ms.outputType[2:]
			}
			if strings.HasPrefix(tag, "map[") { //Check for map[string]. We only handle string keyed maps!!!

				if ms.outputType[4:10] == "string" {
					ms.outputTypeIsMap = true
					ms.outputType = ms.outputType[11:]
				} else {
					logger.Error.Fatalf(errorString_StringMap, "output", ms.signiture)
				}

			}
		}

		if tag := tags.Get("postdata"); tag != "" {
			ms.postdataType = tag
			if strings.HasPrefix(tag, "[]") { //Check for slice/array/list types.
				ms.postdataTypeIsArray = true
				ms.postdataType = ms.postdataType[2:]
			}
			if strings.HasPrefix(tag, "map[") { //Check for map[string]. We only handle string keyed maps!!!

				if ms.postdataType[4:10] == "string" {
					ms.postdataTypeIsMap = true
					ms.postdataType = ms.postdataType[11:]
				} else {
					logger.Error.Fatalf(errorString_StringMap, "postdata", ms.signiture)
				}

			}
		}

		if tag := tags.Get("role"); tag != "" {
			ms.role = tag
		}

		if tag := tags.Get("consumes"); tag != "" {
			ms.overrideConsumesMime = tag
			if !addMimeType(tag) {
				logger.Error.Panicf(errorString_MarshalMimeType, tag)
			}
		}

		if tag := tags.Get("produces"); tag != "" {
			ms.overrideProducesMime = tag
			if !addMimeType(tag) {
				logger.Error.Panicf(errorString_MarshalMimeType, tag)
			}
		}

		if tag := tags.Get("gzip"); tag != "" {
			b, err := strconv.ParseBool(tag)
			if err != nil {
				logger.Warning.Printf(errorString_Gzip, ms.name)
				ms.allowGzip = 2
			} else if b {
				ms.allowGzip = 1
			} else {
				ms.allowGzip = 0
			}
		} else {
			ms.allowGzip = 2
		}

		parseParams(ms)
		return *ms
	}

	logger.Error.Fatalln(errorString_EndpointDecl)
	return *ms //Should not get here
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

	if GetHypermediaDecorator(mimeType) == nil {
		if strings.Contains(mimeType, "siren") {
			RegisterHypermediaDecorator(mimeType, NewSirenDecorator())
		} else if strings.Contains(mimeType, "hal+") {
			RegisterHypermediaDecorator(mimeType, NewHalDecorator())
		}
	}

	return true
}

func parseParams(e *endPointStruct) {
	e.signiture = strings.Trim(e.signiture, "/")
	e.params = make([]param, 0)
	e.queryParams = make([]param, 0)
	e.nonParamPathPart = make(map[int]string, 0)

	pathPart := e.signiture
	queryPart := ""

	if i := strings.Index(e.signiture, "?"); i != -1 {

		pathPart = e.signiture[:i]
		//e.root = pathPart
		pathPart = strings.TrimRight(pathPart, "/")
		queryPart = e.signiture[i+1:]

		//Extract Query Parameters

		for pos, str1 := range strings.Split(queryPart, "&") {
			if strings.HasPrefix(str1, "{") && strings.HasSuffix(str1, "}") {
				parName, typeName := getVarTypePair(str1, e.signiture)

				for _, par := range e.queryParams {
					if par.name == parName {
						logger.Error.Fatalln("Duplicate Query Parameter name(" + parName + ") in REST path: " + e.signiture)
					}
				}
				//e.queryParams[len(e.queryParams)] = param{pos, parName, typeName}
				e.queryParams = append(e.queryParams, param{pos, parName, typeName})
			} else {
				logger.Error.Fatalln("Please check that your Query Parameters are configured correctly for endpoint: " + e.signiture)
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

			parName, typeName := getVarTypePair(str1, e.signiture)

			if parName == "..." {
				e.isVariableLength = true
				parName, typeName := getVarTypePair(str1, e.signiture)
				e.params = append(e.params, param{pos, parName, typeName})
				e.paramLen++
				break
			}
			for _, par := range e.params {
				if par.name == parName {
					logger.Error.Fatalln("Duplicate Path Parameter name(" + parName + ") in REST path: " + e.signiture)
				}
			}

			e.params = append(e.params, param{pos, parName, typeName})
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
		if ep.root == e.root && ep.signitureLen == e.signitureLen && reflect.DeepEqual(ep.nonParamPathPart, e.nonParamPathPart) && ep.requestMethod == e.requestMethod {
			logger.Error.Fatalln("Can not register two endpoints with same request-method(" + ep.requestMethod + ") and same signature: " + e.signiture + " VS " + ep.signiture)
		}
		if ep.requestMethod == e.requestMethod && pathPart == key {
			logger.Error.Fatalln("Endpoint already registered: " + pathPart)
		}
		if e.isVariableLength && (strings.Index(ep.root+"/", e.root+"/") == 0 || strings.Index(e.root+"/", ep.root+"/") == 0) && ep.requestMethod == e.requestMethod {
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

func getEndPointByUrl(method string, url string) (endPointStruct, map[string]string, map[string]string, string, bool) {
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

	epRet := new(endPointStruct)
	pathArgs := make(map[string]string, 0)
	queryArgs := make(map[string]string, 0)

	return *epRet, pathArgs, queryArgs, "", false
}

func matchEndPoint(method string, pathPart string) (ep *endPointStruct, found bool) {
	// try exact match first
	pathMatch := method + ":" + pathPart
	if mEp, match := _manager().endpoints[pathMatch]; match {
		ep = &mEp
		return ep, true
	}

	totalParts := strings.Count(pathPart, "/") + 1
	for _, loopEp := range _manager().endpoints {

		if (strings.Index(pathPart+"/", loopEp.root+"/") == 0) && loopEp.requestMethod == method {
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

func parseVarPathArgs(ep *endPointStruct, pathPart string) map[string]string {
	pathArgs := make(map[string]string, 0)
	varsPart := strings.Trim(pathPart[len(ep.root):], "/")

	for upos, str1 := range strings.Split(varsPart, "/") {
		pathArgs[string(upos)] = strings.Trim(str1, " ")
	}
	return pathArgs
}

func parsePathArgs(ep *endPointStruct, pathPart string) map[string]string {
	pathArgs := make(map[string]string, 0)
	//Extract Path Arguments
	for _, par := range ep.params {
		for upos, str1 := range strings.Split(pathPart, "/") {

			if par.positionInPath == upos {
				pathArgs[par.name] = strings.Trim(str1, " ")
				break
			}
		}
	}
	return pathArgs
}

func parseQueryArgs(ep *endPointStruct, queryPart string) (map[string]string, string) {
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
				for _, par := range ep.queryParams {
					if par.name == pName {
						queryArgs[pName] = strings.Trim(dataString, " ")
						break
					}
				}
			}

		}
	}

	return queryArgs, xsrft
}
