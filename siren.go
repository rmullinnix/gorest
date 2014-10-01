//Copyright 2014  (rmullinnix@yahoo.com). All rights reserved.
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


package gorest

import "reflect"

type Siren struct {
	Class		string
	Properties	interface{}
	Entities	[]Entity
	Actions		[]Action
	Links		[]Link
}

type Entity struct {
	Class		string
	Rel		string
	Properties	interface{}
	Entities	[]Link
	Links		[]Link
}

type Action struct {
	Class		string
	Href		string
	Method		string
	Fields		[]Field
}

type Field struct {
	Name		string
	Type		string
}

type Link struct {
	Rel		string
	Href		string
}

// This takes the data destined for the http response body and adds hypermedia content
// to the message prior to marshaling the data and returning it to the client
// The SirenDecorator loosely follows the siren specification
// mime type: applcation/vnd.siren+json 
func sirenDecorator(response interface{}, resource string, method string) (interface{}) {
	var hm_resp 	Siren

	v := reflect.ValueOf(response)
	switch v.Kind() {
		case reflect.Struct:
			hm_resp.Properties = response
		case reflect.Slice, reflect.Array, reflect.Map:
			hm_resp.Entities, hm_resp.Class = getEntityList(v)
		default:
			hm_resp.Properties = response
	}

	hm_resp.Class = reflect.TypeOf(response).Name()

	return hm_resp
}

func getEntityList(val reflect.Value) ([]Entity, string) {
	entList := make([]Entity, val.Len())
	var className string

	for i := 0; i < val.Len(); i++ {
		var item	Entity

		vItem := val.Index(i)

		item.Class = vItem.Type().Name() + " list-item"
		item.Rel = vItem.Type().Name()
		item.Properties = vItem.Interface()

		entList[i] = item

		if i == 0 {
			className = vItem.Type().Name() + " list"
		}
	}

	return entList, className
}

// creates a new Siren Decorator 
func NewSirenDecorator() *Decorator {
	dec := Decorator{sirenDecorator}
	return &dec
}
