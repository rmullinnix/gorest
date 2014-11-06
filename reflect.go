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
				_manager().swaggerEP = tags.Get("root") + tag
			}
			
			meta := prepServiceMetaData(root, tags, h, t.Name())
			tFullName := _manager().addType(t.PkgPath()+"/"+t.Name(), meta)
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if f.Name != "RestService" {
					mapFieldsToMethods(t, f, tFullName, meta)
				}
			}
		}
		return
	}

	panic(ERROR_INVALID_INTERFACE)
}

func mapFieldsToMethods(t reflect.Type, f reflect.StructField, typeFullName string, serviceRoot ServiceMetaData) {

	if f.Name != "RestService" && f.Type.Name() == "EndPoint" { //TODO: Proper type checking, not by name
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
				logger.Error.Panicln("Method name not found. " + panicMethNotFound(methFound, ep, t, f, methodName))
			}
			if !isLegalForRequestType(method.Type, ep) {
				logger.Error.Panicln("Parameter list not matching. " + panicMethNotFound(methFound, ep, t, f, methodName))
			}
		}
		ep.methodNumberInParent = methodNumberInParent
		_manager().addEndPoint(ep)
		logger.Info.Println("Registerd service:", t.Name(), " endpoint:", ep.RequestMethod, ep.Signiture)
	}
}

func isLegalForRequestType(methType reflect.Type, ep EndPointStruct) (cool bool) {
	cool = true

	numInputIgnore := 0
	numOut := 0

	switch ep.RequestMethod {
	case POST, PUT:
		{
			numInputIgnore = 2 //The first param is the struct, the second the posted object
			numOut = 0
		}
	case GET:
		{
			numInputIgnore = 1 //The first param is the default service struct
			numOut = 1
		}
	case DELETE, HEAD, OPTIONS:
		{
			numInputIgnore = 1 //The first param is the default service struct
			numOut = 0
		}

	}

	if (methType.NumIn() - numInputIgnore) != (ep.paramLen + len(ep.QueryParams)) {
		cool = false
	} else if methType.NumOut() != numOut {
		cool = false
	} else {
		//Check the first parameter type for POST and PUT
		if numInputIgnore == 2 {
			methVal := methType.In(1)
			if ep.postdataTypeIsArray {
				if methVal.Kind() == reflect.Slice {
					methVal = methVal.Elem()
				} else {
					cool = false
					return
				}
			}
			if ep.postdataTypeIsMap {
				if methVal.Kind() == reflect.Map {
					methVal = methVal.Elem()
				} else {
					cool = false
					return
				}
			}

			if !typeNamesEqual(methVal, ep.PostdataType) {
				cool = false
				return
			}
		}
		//Check the rest of input path param types
		i := numInputIgnore
		if ep.isVariableLength {
			if methType.NumIn() != numInputIgnore+1+len(ep.QueryParams) {
				cool = false
			}
			cool = false
			if methType.In(i).Kind() == reflect.Slice { //Variable args Slice
				if typeNamesEqual(methType.In(i).Elem(), ep.Params[0].TypeName) { //Check the correct type for the Slice
					cool = true
				}
			}

		} else {
			for ; i < methType.NumIn() && (i-numInputIgnore < ep.paramLen); i++ {
				if !typeNamesEqual(methType.In(i), ep.Params[i-numInputIgnore].TypeName) {
					cool = false
					break
				}
			}
		}

		//Check the input Query param types
		for j := 0; i < methType.NumIn() && (j < len(ep.QueryParams)); i++ {
			if !typeNamesEqual(methType.In(i), ep.QueryParams[j].TypeName) {
				cool = false
				break
			}
			j++
		}
		//Check output param type.
		if numOut == 1 {
			methVal := methType.Out(0)
			if ep.outputTypeIsArray {
				if methVal.Kind() == reflect.Slice {
					methVal = methVal.Elem() //Only convert if it is mentioned as a slice in the tags, otherwise allow for failure panic
				} else {
					cool = false
					return
				}
			}
			if ep.outputTypeIsMap {
				if methVal.Kind() == reflect.Map {
					methVal = methVal.Elem()
				} else {
					cool = false
					return
				}
			}

			if !typeNamesEqual(methVal, ep.OutputType) {
				cool = false
			}
		}
	}

	return
}

func typeNamesEqual(methVal reflect.Type, name2 string) bool {
	if strings.Index(name2, ".") == -1 {
		return methVal.Name() == name2
	}
	fullName := strings.Replace(methVal.PkgPath(), "/", ".", -1) + "." + methVal.Name()
	return fullName == name2
}

func panicMethNotFound(methFound bool, ep EndPointStruct, t reflect.Type, f reflect.StructField, methodName string) string {

	var str string
	isArr := ""
	postIsArr := ""
	if ep.outputTypeIsArray {
		isArr = "[]"
	}
	if ep.outputTypeIsMap {
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

func prepareServe(context *Context, ep EndPointStruct, args map[string]string, queryArgs map[string]string) (*ResponseBuilder) {
	servMeta := _manager().getType(ep.parentTypeName)

	t := reflect.TypeOf(servMeta.template).Elem() //Get the type first, and it's pointer so Elem(), we created service with new (why??)
	servVal := reflect.New(t).Elem() //Key to creating new instance of service, from the type above

	//Set the Context; the user can get the context from her services function param
	servVal.FieldByName("RestService").FieldByName("Context").Set(reflect.ValueOf(context))
	rs := servVal.FieldByName("RestService").Interface().(RestService)
	rb := rs.ResponseBuilder()

	//Check Authorization

	if servMeta.realm != "" {
		if !GetAuthorizer(servMeta.realm)(context.xsrftoken, servMeta.realm, context.request.Method, rb) {
			return rb
		}
	}

	arrArgs := make([]reflect.Value, 0)

	targetMethod := servVal.Type().Method(ep.methodNumberInParent)
	mime := servMeta.ConsumesMime
	if ep.overrideConsumesMime != "" {
		mime = ep.overrideConsumesMime
	}
	//For POST and PUT, make and add the first "postdata" argument to the argument list
	if ep.RequestMethod == POST || ep.RequestMethod == PUT {

		//Get postdata here
		//TODO: Also check if this is a multipart post and handle as required.
		buf := new(bytes.Buffer)
		io.Copy(buf, context.request.Body)
		body := buf.String()

		//println("This is the body of the post:",body)

		if v, valid := makeArg(body, targetMethod.Type.In(1), mime); valid {
			arrArgs = append(arrArgs, v)
		} else {
			rb.SetResponseCode(http.StatusBadRequest)
			rb.SetResponseMsg("Error unmarshalling data using " + mime)
			return rb
		}
	}

	if len(args) == ep.paramLen || (ep.isVariableLength && ep.paramLen == 1) {
		startIndex := 1
		if ep.RequestMethod == POST || ep.RequestMethod == PUT {
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
					return rb
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
					return rb
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
				return rb
			}

			startIndex++
		}

		//Now call the actual method with the data
		var ret []reflect.Value
		if ep.isVariableLength {
			ret = servVal.Method(ep.methodNumberInParent).CallSlice(arrArgs)
		} else {
			ret = servVal.Method(ep.methodNumberInParent).Call(arrArgs)
		}

		if len(ret) == 1 { //This is when we have just called a GET
			var mimeType	string

			accept := context.request.Header.Get("Accept")

			if mimeType = ep.overrideProducesMime; mimeType == "" {
				mimeType = servMeta.ProducesMime[0]
				if len(accept) > 0 {
				 	for i := 0; i < len(servMeta.ProducesMime); i++ {
						if strings.Contains(accept, servMeta.ProducesMime[i]) {
							mimeType = servMeta.ProducesMime[i]
							break
						}
					}
				}
			}

			// check for hypermedia decorator
			dec := GetHypermediaDecorator(mimeType)
			hidec := ret[0].Interface()
			if dec != nil {
				prefix := "http://" + context.request.Host
				hidec = dec.Decorate(prefix, hidec, entities)
			}

			rb.ctx.responseMimeType = mimeType
			//At this stage we should be ready to write the response to client
			if bytarr, err := interfaceToBytes(hidec, mimeType); err == nil {
				rb.ctx.respPacket = bytarr
				rb.SetResponseCode(http.StatusOK)
				return rb
			} else {
				//This is an internal error with the registered marshaller not being able to marshal internal structs
				rb.SetResponseCode(http.StatusInternalServerError)
				rb.SetResponseMsg("Internal server error. Could not Marshal/UnMarshal data: " + err.Error())
				return rb
			}
		} else {
			rb.SetResponseCode(http.StatusOK)
			return rb
		}
	}

	//Just in case the whole civilization crashes and it falls thru to here. This shall never happen though... well tested
	logger.Error.Panicln("There was a problem with request handing. Probably a bug, please report.") //Add client data, and send support alert
	rb.SetResponseCode(http.StatusInternalServerError)
	rb.SetResponseMsg("GoRest: Internal server error.")
	return rb 
}

func makeArg(data string, template reflect.Type, mime string) (reflect.Value, bool) {
	i := reflect.New(template).Interface()

	if data == "" {
		return reflect.ValueOf(i).Elem(), true
	}

	buf := bytes.NewBufferString(data)
	err := bytesToInterface(buf, i, mime)

	if err != nil {
		logger.Error.Println("Error Unmarshalling data using " + mime + ". Incompatable data format in entity. (" + err.Error() + ")")
		return reflect.ValueOf(nil), false
	}
	
	return reflect.ValueOf(i).Elem(), true
}
