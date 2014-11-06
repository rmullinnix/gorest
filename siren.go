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
	"strings"
	"reflect"
	"regexp"
	"strconv"
	"unicode"
)

type Siren struct {
	Class		string		`json:"class,omitempty"`
	Title		string		`json:"title,omitempty"`
	Properties	interface{}	`json:"properties"`
	Entities	[]SEntity	`json:"entities,omitempty"`
	Actions		[]SAction	`json:"actions,omitempty"`
	Links		[]SLink		`json:"links"`
}

type SEntity struct {
	Class		string		`json:"class,omitempty"`
	Rel		string		`json:"rel"`
	Properties	interface{}	`json:"properties"`
	Links		[]SLink		`json:"links,omitempty"`
}

type SAction struct {
	Name		string		`json:"name"`
	Class		string		`json:"class,omitempty"`
	Method		string		`json:"method,omitempty"`
	Href		string		`json:"href"`
	Title		string		`json:"title,omitempty"`
	Type		string		`json:"type,omitempty"`
}

type Field struct {
	Name		string		`json:"name"`
	Type		string		`json:"type"`
	Value		string		`json:"value,omitempty"`
	Title		string		`json:"title,omitempty"`
}

type SLink struct {
	Class		string		`json:"class,omitempty"`
	Title		string		`json:"title,omitempty"`
	Rel		string		`json:"rel"`
	Href		string		`json:"href"`
	Type		string		`json:"type,omitempty"`
}

var srvr_prefix		string
// This takes the data destined for the http response body and adds hypermedia content
// to the message prior to marshaling the data and returning it to the client
// The SirenDecorator loosely follows the siren specification
// mime type: applcation/vnd.siren+json 
func sirenDecorator(prefix string, response interface{}, entmap map[string]entity) (interface{}) {
	var hm_resp 	Siren

	srvr_prefix = prefix

	v := reflect.ValueOf(response)
	switch v.Kind() {
		case reflect.Struct:
			// Properties - not sub-entity items
			// Any sub-entities (struct or array), placed in Entities
			props, ents := stripSubentities(v, entmap)
			hm_resp.Properties = props
			hm_resp.Entities = ents
			hm_resp.Class = reflect.TypeOf(response).Name()
			hm_resp.Actions = sirenActions(entmap[hm_resp.Class], props)
			hm_resp.Links = sirenLinks(entmap[hm_resp.Class], props)
		case reflect.Slice, reflect.Array, reflect.Map:
			hm_resp.Entities, hm_resp.Class = getEntityList(v, entmap)
		default:
			hm_resp.Properties = response
			hm_resp.Class = reflect.TypeOf(response).Name()
	}

	return hm_resp
}

func sirenLinks(ent entity, props map[string]interface{}) []SLink {
	lnklist := make([]SLink, len(ent.links))
	i := 0
	
	for _, e_lnk := range ent.links {
		lnk := SLink{"", "", e_lnk.rel, e_lnk.href, ""}
		lnk.Href = updatePath(lnk.Href, props)
		lnklist[i] = lnk
		i++
	}
	return lnklist
}

func sirenActions(ent entity, props map[string]interface{}) []SAction {
	actlist := make([]SAction, len(ent.actions))
	i := 0
	
	for _, e_act := range ent.actions {
		act := SAction{e_act.name, e_act.class, e_act.method, e_act.href, "", ""}
		act.Href = updatePath(act.Href, props)
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
		switch vItem.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				if vItem.Len() > 0 {
					item := vItem.Index(0)
					if _, ok := entmap[item.Type().Name()]; ok {
						tmp, _ := getEntityList(vItem, entmap)
						ents = append(ents, tmp...)
					} else {
						out[typ.Field(i).Name] = vItem.Interface()
					}
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

			lnk := SLink{"", "", subent.links[j].rel, subent.links[j].href, ""}

			typ := reflect.TypeOf(item.Properties)
			val := reflect.ValueOf(item.Properties)
			props := make(map[string]interface{}, typ.NumField())
			for i := 0; i < typ.NumField(); i++ {
				valf := val.Field(i)
				props[typ.Field(i).Name] = valf.Interface()
			}

			lnk.Href = updatePath(lnk.Href, props)

			item.Links = append(item.Links, lnk)
		}
	}
	return item
}

func updatePath(path string, props map[string]interface{}) string {

	reg := regexp.MustCompile("{[^}]+}")
	parts := reg.FindAllString(path, -1)

	for _, str1 := range parts {
		if strings.HasPrefix(str1, "{") && strings.HasSuffix(str1, "}") {
			
			str2 := str1[1:len(str1) - 1]
			if pos := strings.IndexAny(str2, "+-"); pos > -1  {
				if item, found := props[str2[:pos]]; found {
					value := reflect.ValueOf(item).Int()
					val2, _ := strconv.Atoi(str2[pos+1:])
					if strings.Contains(str2, "+")  {
						value = value + int64(val2)
					} else {
						value = value - int64(val2)
					}
					path = strings.Replace(path, str1, strconv.FormatInt(value, 10), 1)
				}
			} else if item, found := props[str2]; found {
				path = strings.Replace(path, str1, getValueString(item), 1)
			}
		}
	}
	path = srvr_prefix + path

	return path
}

func getValueString(item interface{}) string {
	value := "<invalid>"
	vItem := reflect.ValueOf(item)
	switch vItem.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = strconv.FormatInt(vItem.Int(), 10)
		case reflect.Bool:
			value = strconv.FormatBool(vItem.Bool())
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			value = strconv.FormatUint(vItem.Uint(), 10)
		case reflect.Float32:
			value = strconv.FormatFloat(vItem.Float(), 'e', -1, 32)
		case reflect.String:
			value = strconv.FormatFloat(vItem.Float(), 'e', -1, 64)
	}
	return value
}

// creates a new Siren Decorator 
func NewSirenDecorator() *Decorator {
	dec := Decorator{sirenDecorator}
	return &dec
}
