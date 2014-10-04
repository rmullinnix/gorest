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
//	"github.com/rmullinnix/logger"
	"reflect"
	"strings"
)

var decorators map[string]*Decorator

type entity struct {
	class		string
	title		string
	key		string
	typ		string
	href		string
	links		map[int]link
	actions		map[int]action
}

type link struct {
	title		string
	rel		string
	href		string
	typ		string
}

// leaving fields off for now
type action struct {
	name		string
	method		string
	href		string
	class		string
	title		string
	typ		string
}

type Entity bool
type Link bool
type Action bool
type Curie bool
type Query bool
type Error bool
type Template bool
type Data bool

var entityInitialized	bool
var entities 		map[string]entity

//Signiture of functions to be used as Decorators
type Decorator struct {
	Decorate func(interface{}, map[string]entity)(interface{})
}

//Registers an Hypermedia Decorator for the specified mime type
func RegisterHypermediaDecorator(mime string, dec *Decorator) {
	if decorators == nil {
		decorators = make(map[string]*Decorator, 0)
	}
	if _, found := decorators[mime]; !found {
		decorators[mime] = dec
	}
}

//Returns the registred decorator for the specified mime type
func GetHypermediaDecorator(mime string) (dec *Decorator) {
	if decorators == nil {
		decorators = make(map[string]*Decorator, 0)
	}
	dec, _ = decorators[mime]
	return
}

func RegisterEntity(i_ent interface{}) {

	if !entityInitialized {
		entities = make(map[string]entity)
		entityInitialized = true
	}

	t := reflect.TypeOf(i_ent)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	} else {
		panic(ERROR_INVALID_INTERFACE)
	}

	if t.Kind() == reflect.Struct {
		if field, found := t.FieldByName("Entity"); found {
			temp := strings.Join(strings.Fields(string(field.Tag)), " ")
			ent := prepEntityData(reflect.StructTag(temp))
			ent.links = make(map[int]link)
			ent.actions = make(map[int]action)

			linkcnt := 0
			actioncnt := 0
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				ftmp := strings.Join(strings.Fields(string(f.Tag)), " ")
				if f.Type.Name() == "Link" {
					lnk := prepLinkData(f.Name, reflect.StructTag(ftmp))
					ent.links[linkcnt] = lnk
					linkcnt++
				} else if f.Type.Name() == "Action" {
					act := prepActionData(f.Name, reflect.StructTag(ftmp))
					ent.actions[actioncnt] = act
					actioncnt++
				}
			}
			entities[ent.class] = ent	
		}
	}
}

func prepEntityData(tags reflect.StructTag) entity {
	ent := new(entity)

	var tag		string

	if tag = tags.Get("class"); tag != "" {
		ent.class = tag
	}

	if tag = tags.Get("title"); tag != "" {
		ent.title = tag
	}

	if tag = tags.Get("key"); tag != "" {
		ent.key = tag
	}

	if tag = tags.Get("href"); tag != "" {
		ent.href = tag
	}

	if tag = tags.Get("type"); tag != "" {
		ent.typ = tag
	}

	return *ent
}

func prepLinkData(rel string, tags reflect.StructTag) link {
	lnk := new(link)

	var tag		string

	lnk.rel = rel

	if tag = tags.Get("href"); tag != "" {
		lnk.href = tag
	}

	if tag = tags.Get("title"); tag != "" {
		lnk.title = tag
	}

	if tag = tags.Get("type"); tag != "" {
		lnk.typ = tag
	}

	return *lnk
}

func prepActionData(name string, tags reflect.StructTag) action {
	act := new(action)

	var tag		string

	act.name = name

	if tag = tags.Get("method"); tag != "" {
		act.method = tag
	}

	if tag = tags.Get("href"); tag != "" {
		act.href = tag
	}

	if tag = tags.Get("class"); tag != "" {
		act.class = tag
	}

	if tag = tags.Get("title"); tag != "" {
		act.title = tag
	}

	if tag = tags.Get("type"); tag != "" {
		act.typ = tag
	}

	return *act
}
