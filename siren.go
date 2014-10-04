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

import (
	"fmt"
	"strings"
	"reflect"
	"unicode"
)

type Siren struct {
	Class		string
	Properties	interface{}
	Entities	[]SEntity
	Actions		[]SAction
	Links		[]SLink
}

type SEntity struct {
	Class		string
	Rel		string
	Properties	interface{}
	Links		[]SLink
}

type SAction struct {
	Name		string
	Class		string
	Href		string
	Method		string
}

type Field struct {
	Name		string
	Type		string
}

type SLink struct {
	Rel		string
	Href		string
}

// This takes the data destined for the http response body and adds hypermedia content
// to the message prior to marshaling the data and returning it to the client
// The SirenDecorator loosely follows the siren specification
// mime type: applcation/vnd.siren+json 
func sirenDecorator(response interface{}, entmap map[string]entity) (interface{}) {
	var hm_resp 	Siren

	v := reflect.ValueOf(response)
	switch v.Kind() {
		case reflect.Struct:
			// Properties - not sub-entity items
			// Any sub-entities (struct or array), placed in Entities
			hm_resp.Properties, hm_resp.Entities = stripSubentities(v, entmap)
			hm_resp.Class = reflect.TypeOf(response).Name()
			key := v.FieldByName(entmap[hm_resp.Class].key)
			hm_resp.Actions = sirenActions(entmap[hm_resp.Class], key.String())
			hm_resp.Links = sirenLinks(entmap[hm_resp.Class], key.String())
		case reflect.Slice, reflect.Array, reflect.Map:
			hm_resp.Entities, hm_resp.Class = getEntityList(v, entmap)
		default:
			hm_resp.Properties = response
			hm_resp.Class = reflect.TypeOf(response).Name()
	}

	return hm_resp
}

func sirenLinks(ent entity, keyval string) []SLink {
	lnklist := make([]SLink, len(ent.links))
	i := 0
	
	for _, e_lnk := range ent.links {
		lnk := SLink{e_lnk.rel, e_lnk.href}
		if strings.Contains(lnk.Href, "{key}") {
			lnk.Href = strings.Replace(lnk.Href, "{key}", keyval, 1)
		}
		lnklist[i] = lnk
		i++
	}
	return lnklist
}

func sirenActions(ent entity, keyval string) []SAction {
	actlist := make([]SAction, len(ent.actions))
	i := 0
	
	for _, e_act := range ent.actions {
		act := SAction{e_act.name, e_act.class, e_act.href, e_act.method}
		if strings.Contains(act.Href, "{key}") {
			act.Href = strings.Replace(act.Href, "{key}", keyval, 1)
		}
		actlist[i] = act
		i++
	}
	return actlist
}

func stripSubentities(in reflect.Value, entmap map[string]entity) (map[string]interface{}, []SEntity) {
	ents :=	[]SEntity{}
	out := make(map[string]interface{}, 30)

	typ := reflect.TypeOf(in.Interface())
	for i := 0; i < typ.NumField(); i++ {
		vItem := in.Field(i)
		fmt.Println(vItem)
		switch vItem.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				item := vItem.Index(0)
				if _, ok := entmap[item.Type().Name()]; ok {
					tmp, _ := getEntityList(vItem, entmap)
					ents = append(ents, tmp...)
				} else {
					out[typ.Field(i).Name] = vItem.Interface()
				}
			default:
				if _, ok := entmap[typ.Field(i).Name]; ok {
					item := getEntity(false, vItem, entmap)
					ents = append(ents, item)
				} else {
					out[typ.Field(i).Name] = vItem.Interface()
				}
		}
	}

	return out, ents
}

func getEntityList(val reflect.Value, entmap map[string]entity) ([]SEntity, string) {
	entList := []SEntity{}
	var className string

	for i := 0; i < val.Len(); i++ {
		vItem := val.Index(i)

		item := getEntity(true, vItem, entmap)
		item.Class = vItem.Type().Name() + " list-item"

		entList = append(entList, item)

		if i == 0 {
			className = vItem.Type().Name() + " list"
		}
	}

	return entList, className
}

func getEntity(sub bool, vItem reflect.Value, entmap map[string]entity) SEntity {
	var item	SEntity

	item.Class = vItem.Type().Name()
	item.Rel = vItem.Type().Name()
	item.Properties = vItem.Interface()

	if subent, ok := entmap[vItem.Type().Name()]; ok {
		for j:= 0; j < len(subent.links); j++ {
			if sub && strings.IndexFunc(subent.links[j].rel[:1], unicode.IsUpper) == 0 {
				continue
			}

			lnk := SLink{subent.links[j].rel, subent.links[j].href}
			if strings.Contains(lnk.Href, "{key}") {
				keyval := vItem.FieldByName(subent.key)
//				if !val.IsNil() {
					lnk.Href = strings.Replace(lnk.Href, "{key}", keyval.String(), 1)
//				}
			}

			item.Links = append(item.Links, lnk)
		}
	}
	return item
}

// creates a new Siren Decorator 
func NewSirenDecorator() *Decorator {
	dec := Decorator{sirenDecorator}
	return &dec
}
