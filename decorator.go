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

var decorators map[string]*Decorator

//Signiture of functions to be used as Decorators
type Decorator struct {
	Decorate func(interface{}, string, string)(interface{})
}

//Registers an Hypermedia Decorator
func RegisterHypermediaDecorator(mime string, dec *Decorator) {
	if decorators == nil {
		decorators = make(map[string]*Decorator, 0)
	}
	if _, found := decorators[mime]; !found {
		decorators[mime] = dec
	}
}

//Returns the registred Authorizer for the specified realm.
func GetHypermediaDecorator(mime string) (dec *Decorator) {
	if decorators == nil {
		decorators = make(map[string]*Decorator, 0)
	}
	dec, _ = decorators[mime]
	return
}
