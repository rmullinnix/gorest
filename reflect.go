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
	"bytes"
	"io"
	"github.com/rmullinnix/logger"
	"net/http"
	"reflect"
	"strings"
)

const (
	ERROR_INVALID_INTERFACE = "RegisterService(interface{}) takes a pointer to a struct that inherits from type RestService. Example usage: gorest.RegisterService(new(ServiceOne)) "
)

//Bootstrap functions below
//------------------------------------------------------------------------------------------

//Takes a value of a struct representing a service.
func registerService(root string, h interface{}) {

	if _, ok := h.(GoRestService); !ok {
		panic(ERROR_INVALID_INTERFACE)
	}

	t := reflect.TypeOf(h)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		panic(ERROR_INVALID_INTERFACE)
	}

	if t.Kind() == reflect.Struct {
		if field, found := t.FieldByName("RestService"); found {
			temp := strings.Join(strings.Fields(string(field.Tag)), " ")
			tags := reflect.StructTag(temp)
			_manager().root = tags.Get("root")
			if tag := tags.Get("swagger"); tag != "" {
				logger.Info.Println("[gen] Registered swagger endpoint: ", tags.Get("root") + tag)
				_manager().swaggerEP = tags.Get("root") + tag
			}
			
			meta := prepServiceMetaData(root, tags, h, t.Name())
			tFullName := _manager().addType(t.PkgPath()+"/"+t.Name(), meta)
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if f.Name != "RestService" {
					if f.Type.Name() == "EndPoint" {
						mapFieldsToMethods(t, f, tFullName, meta)
					} else if f.Type.Name() == "Security" {
						temp := strings.Join(strings.Fields(string(f.Tag)), " ")
						secDef := prepSecurityMetaData(reflect.StructTag(temp))
						_manager().addSecurityDefinition(f.Name, secDef)
					}
				}
			}
		}
		return
	}

	panic(ERROR_INVALID_INTERFACE)
}

func mapFieldsToMethods(t reflect.Type, f reflect.StructField, typeFullName string, serviceRoot ServiceMetaData) {

	temp := strings.Join(strings.Fields(string(f.Tag)), " ")
	ep := makeEndPointStruct(reflect.StructTag(temp), serviceRoot.Root)
	ep.parentTypeName = typeFullName
	ep.Name = f.Name
	// override the endpoint with our default value for gzip
	if ep.allowGzip == 2 {
		if !serviceRoot.allowGzip {
			ep.allowGzip = 0
		} else {
			ep.allowGzip = 1
		}
	}

	var method reflect.Method
	methodName := strings.ToUpper(f.Name[:1]) + f.Name[1:]

	methFound := false
	methodNumberInParent := 0
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if methodName == m.Name {
			method = m //As long as the name is the same, we know we have found the method, since go has no overloading
			methFound = true
			methodNumberInParent = i
			break
		}
	}

	{ //Panic Checks
		if !methFound {
			logger.Error.Panicln("[fatal] Method name not found. " + panicMethNotFound(methFound, ep, t, f, methodName))
		}
		if !isLegalForRequestType(method.Type, ep) {
			logger.Error.Panicln("[fatal] Parameter list not matching. " + panicMethNotFound(methFound, ep, t, f, methodName))
		}
	}

	ep.MethodNumberInParent = methodNumberInParent
	_manager().addEndPoint(ep)

	logger.Info.Println("[gen] Registerd service:", t.Name(), " endpoint:", ep.RequestMethod, ep.Signiture)
}

func isLegalForRequestType(methType reflect.Type, ep EndPointStruct) bool {
	startParam := 1

//	switch ep.RequestMethod {
//	case POST, PUT:
//		{
//			numInputIgnore = 2 //The first param is the struct, the second the posted object
//		}
//	case GET, DELETE, HEAD, OPTIONS:
//		{
//			numInputIgnore = 1 //The first param is the default service struct
//		}
//	}
	if len(ep.PostdataType) > 0 {
		startParam = 2
	}

	if (methType.NumIn() - startParam) != (ep.paramLen + len(ep.QueryParams)) {
		return false
	}

	//Check the first parameter type for POST and PUT
	if len(ep.PostdataType) > 0 {
		startParam++
		methVal := methType.In(1)
		if ep.postdataTypeIsArray {
			if methVal.Kind() == reflect.Slice {
				methVal = methVal.Elem()
			} else {
				return false
			}
		}
		if ep.postdataTypeIsMap {
			if methVal.Kind() == reflect.Map {
				methVal = methVal.Elem()
			} else {
				return false
			}
		}

		if !typeNamesEqual(methVal, ep.PostdataType) {
			return false
		}
	}
	//Check the rest of input path param types
	i := startParam
	if ep.isVariableLength {
		if methType.NumIn() != startParam+1+len(ep.QueryParams) {
			return false
		}

		if methType.In(i).Kind() == reflect.Slice { //Variable args Slice
			if !typeNamesEqual(methType.In(i).Elem(), ep.Params[0].TypeName) { //Check the correct type for the Slice
				return false
			}
		}
	} else {
		for ; i < methType.NumIn() && (i-startParam < ep.paramLen); i++ {
			if !typeNamesEqual(methType.In(i), ep.Params[i-startParam].TypeName) {
				return false
			}
		}
	}

	//Check the input Query param types
	for j := 0; i < methType.NumIn() && (j < len(ep.QueryParams)); i++ {
		if ep.QueryParams[j].TypeName[:2] == "[]" {
			if methType.In(i).Elem().String() != ep.QueryParams[j].TypeName[2:] {
				return false
			}
		} else if !typeNamesEqual(methType.In(i), ep.QueryParams[j].TypeName) {
			return false
		}
		j++
	}
	//Check output param type.
	if methType.NumOut() > 0 {
		methVal := methType.Out(0)

		if ep.OutputTypeIsArray {
			if methVal.Kind() == reflect.Slice {
				methVal = methVal.Elem() //Only convert if it is mentioned as a slice in the tags, otherwise allow for failure panic
			} else {
				return false
			}
		}
		if ep.OutputTypeIsMap {
			if methVal.Kind() == reflect.Map {
				methVal = methVal.Elem()
			} else {
				return false
			}
		}

		if !typeNamesEqual(methVal, ep.OutputType) {
			return false
		}
	}

	return true
}

func typeNamesEqual(methVal reflect.Type, name2 string) bool {
	if strings.Index(name2, ".") == -1 {
		return methVal.Name() == name2
	}
	//fullName := strings.Replace(methVal.PkgPath(), "/", ".", -1) + "." + methVal.Name()
	abbrevName := name2[strings.Index(name2, ".") + 1:]

	return abbrevName == methVal.Name()
}

func panicMethNotFound(methFound bool, ep EndPointStruct, t reflect.Type, f reflect.StructField, methodName string) string {

	var str string
	isArr := ""
	postIsArr := ""
	if ep.OutputTypeIsArray {
		isArr = "[]"
	}
	if ep.OutputTypeIsMap {
		isArr = "map[string]"
	}
	if ep.postdataTypeIsArray {
		postIsArr = "[]"
	}
	if ep.postdataTypeIsMap {
		postIsArr = "map[string]"
	}
	var suffix string = "(" + isArr + ep.OutputType + ")# with one(" + isArr + ep.OutputType + ") return parameter."
	if ep.RequestMethod == POST || ep.RequestMethod == PUT {
		str = "PostData " + postIsArr + ep.PostdataType
		if ep.paramLen > 0 {
			str += ", "
		}

	}
	if ep.RequestMethod == POST || ep.RequestMethod == PUT || ep.RequestMethod == DELETE {
		suffix = "# with no return parameters."
	}
	if ep.isVariableLength {
		str += "varArgs ..." + ep.Params[0].TypeName + ","
	} else {
		for i := 0; i < ep.paramLen; i++ {
			str += ep.Params[i].Name + " " + ep.Params[i].TypeName + ","
		}
	}

	for i := 0; i < len(ep.QueryParams); i++ {
		str += ep.QueryParams[i].Name + " " + ep.QueryParams[i].TypeName + ","
	}
	str = strings.TrimRight(str, ",")
	return "No matching Method found for EndPoint:[" + f.Name + "],type:[" + ep.RequestMethod + "] . Expecting: #func(serv " + t.Name() + ") " + methodName + "(" + str + ")" + suffix
}

//Runtime functions below:
//-----------------------------------------------------------------------------------------------------------------

func prepareServe(rb *ResponseBuilder, ep EndPointStruct, args map[string]string, queryArgs map[string]string) {
	servMeta := _manager().getType(ep.parentTypeName)

	t := reflect.TypeOf(servMeta.Template).Elem() //Get the type first, and it's pointer so Elem(), we created service with new (why??)
	servVal := reflect.New(t).Elem() //Key to creating new instance of service, from the type above

	//Set the Context; the user can get the context from her services function param
	servVal.FieldByName("RestService").FieldByName("Context").Set(reflect.ValueOf(rb.ctx))

	//Check Authorization

	if ep.SecurityScheme != nil {
		authorized := false
		for key, scopes := range ep.SecurityScheme {
			alteredScopes := make([]string, len(scopes))
			for i := range scopes {
				alteredScopes[i] = replaceScopeKey(scopes[i], args)
			}
			authorized = GetAuthorizer(key)(rb.ctx.xsrftoken, key, alteredScopes, rb.ctx.request.Method, rb)
			if authorized {
				break
			}
		}
		if !authorized {
			// authorizer should log failure reason
			rb.SetResponseCode(401)
			return
		}
	}

	arrArgs := make([]reflect.Value, 0)

	targetMethod := servVal.Type().Method(ep.MethodNumberInParent)

	contentType := rb.ctx.request.Header.Get("Content-Type")

	if contentType == "" {
		contentType = servMeta.ConsumesMime[0]
	}

	valid, mime := validMime(contentType, ep.ConsumesMime, servMeta.ConsumesMime)
	if !valid {
		if len(ep.PostdataType) > 0 {
			// error - can not accept request
			logger.Error.Println("[gen] service is not configured to accept Content-Type " + contentType)
			rb.SetResponseCode(http.StatusBadRequest)
			rb.SetResponseMsg("Service is not configured to accept Content-Type " + contentType)
			return
		}
	}

	//For POST and PUT, make and add the first "postdata" argument to the argument list
	if len(ep.PostdataType) > 0 {

		//Get postdata here
		//TODO: Also check if this is a multipart post and handle as required.
		buf := new(bytes.Buffer)
		io.Copy(buf, rb.ctx.request.Body)
		body := buf.String()

		//println("This is the body of the post:",body)

		if v, valid := makeArg(body, targetMethod.Type.In(1), mime); valid {
			arrArgs = append(arrArgs, v)
		} else {
			rb.SetResponseCode(http.StatusBadRequest)
			rb.SetResponseMsg("Error unmarshalling data using " + mime)
			return
		}
	}

	if len(args) == ep.paramLen || (ep.isVariableLength && ep.paramLen == 1) {
		startIndex := 1
		if len(ep.PostdataType) > 0 {
			startIndex = 2
		}

		if ep.isVariableLength {
			varSliceArgs := reflect.New(targetMethod.Type.In(startIndex)).Elem()
			for ij := 0; ij < len(args); ij++ {
				dat := args[string(ij)]

				if v, valid := makeArg(dat, targetMethod.Type.In(startIndex).Elem(), mime); valid {
					varSliceArgs = reflect.Append(varSliceArgs, v)
				} else {
					rb.SetResponseCode(http.StatusBadRequest)
					rb.SetResponseMsg("Error unmarshalling data using " + mime)
					return
				}
			}
			arrArgs = append(arrArgs, varSliceArgs)
		} else {
			//Now add the rest of the PATH arguments to the argument list and then call the method
			// GET and DELETE will only need these arguments, not the "postdata" one in their method calls
			for _, par := range ep.Params {
				dat := ""
				if str, found := args[par.Name]; found {
					dat = str
				}

				if v, valid := makeArg(dat, targetMethod.Type.In(startIndex), mime); valid {
					arrArgs = append(arrArgs, v)
				} else {
					rb.SetResponseCode(http.StatusBadRequest)
					rb.SetResponseMsg("Error unmarshalling data using " + mime)
					return
				}
				startIndex++
			}

		}

		//Query arguments are not compulsory on query, so the caller may ommit them, in which case we send a zero value f its type to the method.
		//Also they may be sent through in any order.
		for _, par := range ep.QueryParams {
			dat := ""
			if str, found := queryArgs[par.Name]; found {
				dat = str
			}

			if v, valid := makeArg(dat, targetMethod.Type.In(startIndex), mime); valid {
				arrArgs = append(arrArgs, v)
			} else {
				rb.SetResponseCode(http.StatusBadRequest)
				rb.SetResponseMsg("Error unmarshalling data using " + mime)
				return
			}

			startIndex++
		}

		//Now call the actual method with the data
		var ret []reflect.Value
		if ep.isVariableLength {
			ret = servVal.Method(ep.MethodNumberInParent).CallSlice(arrArgs)
		} else {
			ret = servVal.Method(ep.MethodNumberInParent).Call(arrArgs)
		}

		if len(ret) == 1 { //This is when we have just called a GET
			var mimeType	string

			accept := rb.ctx.request.Header.Get("Accept")
			valid := false

			if len(accept) > 0 {
				if valid, mimeType = validMime(accept, ep.ProducesMime, servMeta.ProducesMime); valid {
			//		rb.SetResponseCode(http.StatusBadRequest)
			//		rb.SetResponseMsg("Service does not support Accept mime type " + mime)
			//		return rb
				}
			}
			if !valid {
				if len(ep.ProducesMime) > 0 {
					mimeType = ep.ProducesMime[0]
				} else {
					mimeType = servMeta.ProducesMime[0]
				}
			}

			rb.SetContentType(mimeType)

			// check for hypermedia decorator
			dec := GetHypermedia()
			hidec := ret[0].Interface()
			if dec != nil {
				scope := make([]string, 0)
				prefix := "http://" + rb.ctx.request.Host
				item, found := rb.Session().Get("Scope")
				if found {
					iScope := item.([]interface{})
					scope = make([]string, len(iScope))
					for i := range iScope {
						scope[i] = iScope[i].(string)
					}
				}
				hidec = dec.Decorate(mimeType, prefix, hidec, scope)
			}

			rb.ctx.responseMimeType = mimeType
			//At this stage we should be ready to write the response to client
			if bytarr, err := interfaceToBytes(hidec, mimeType); err == nil {
				rb.ctx.respPacket = bytarr
				rb.AddHeader("Content-Type", mimeType)
				//rb.SetResponseCode(http.StatusOK)
				return
			} else {
				//This is an internal error with the registered marshaller not being able to marshal internal structs
				rb.SetResponseCode(http.StatusInternalServerError)
				rb.SetResponseMsg("Internal server error. Could not Marshal/UnMarshal data: " + err.Error())
				return
			}
		} else {
			//rb.SetResponseCode(http.StatusOK)
			return
		}
	}

	//Just in case the whole civilization crashes and it falls thru to here. This shall never happen though... well tested
	logger.Error.Panicln("[gen] There was a problem with request handing. Probably a bug, please report.") //Add client data, and send support alert
	rb.SetResponseCode(http.StatusInternalServerError)
	rb.SetResponseMsg("GoRest: Internal server error.")
	return
}

func makeArg(data string, template reflect.Type, mime string) (reflect.Value, bool) {

	kind := template.Kind()
	// convert array arg from string to array format before marshalling
	if kind == reflect.Slice || kind == reflect.Array {
		if template.Elem().Kind() == reflect.String {
			data = "[\"" + strings.Replace(data, ",", "\",\"", -1) + "\"]"
		} else {
			data = "[" + data + "]"
		}
	}

	i := reflect.New(template).Interface()

	if data == "" {
		return reflect.ValueOf(i).Elem(), true
	}

	buf := bytes.NewBufferString(data)
	err := bytesToInterface(buf, i, mime)

	if err != nil {
		logger.Error.Println("[gen] Error Unmarshalling data using " + mime + ". Incompatable data format in entity. (" + err.Error() + ")")
		return reflect.ValueOf(nil), false
	}
	
	return reflect.ValueOf(i).Elem(), true
}

func validMime(mimeType string, epMime []string, srvMime []string) (bool, string) {
	found := false
	ctype := ""

	parts := strings.Split(mimeType, ";")
	for j := range parts {
		// Endpoint mime type list overrides the one defined at the server
		// It is not a union
		if len(epMime) > 0 {
			for i := range epMime {
				if parts[j] == epMime[i] {
					found = true
					ctype = parts[j]
					break
				}
			}
		} else {
			for i := range srvMime {
				if parts[j] == srvMime[i] {
					found = true
					ctype = parts[j]
					break
				}
			}
		}
		if found {
			break
		}
	}

	return found, ctype
}

func replaceScopeKey(scope string, args map[string]string) string {
        out := scope
	value := ""
        if pos := strings.Index(scope, "{"); pos > -1 {
                key := scope[pos + 1 : strings.Index(scope, "}")]
		if key != ""  {
	                value = args[key]
		}
                out = scope[:pos] + "[" + value + "]"
        }

        return out
}
