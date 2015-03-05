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


package gorest

import "github.com/rmullinnix/logger"

var authorizers map[string]Authorizer

//Signiture of functions to be used as Authorizers
//  token, scheme, scopes, method, ResponseBuilder
type Authorizer func(string, string, []string, string, *ResponseBuilder)(bool)

//Registers an Authorizer for the specified realm.
func RegisterAuthorizer(scheme string, auth Authorizer){
	if authorizers == nil{
		authorizers = make(map[string]Authorizer,0)
	}
	
	if _,found := authorizers[scheme]; !found{
		authorizers[scheme] = auth
	}
}

//Returns the registred Authorizer for the specified realm.
func GetAuthorizer(realm string)(a Authorizer){
	if authorizers ==nil{
		authorizers = make(map[string]Authorizer,0)
	}
	a,_ = authorizers[realm]
	return 
}

//This is the default and exmaple authorizer that is used to authorize requests to endpints with no security realms.
//It always allows access and returns nil for SessionData.
func DefaultAuthorizer(token string, scheme string, scopes []string, method string, rb *ResponseBuilder)(bool) {
	logger.Info.Println(token, scheme, method)
	return true
}

