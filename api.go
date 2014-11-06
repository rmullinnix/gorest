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

import (
	"compress/gzip"
	"github.com/rmullinnix/logger"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	XSXRF_COOKIE_NAME = "X-Xsrf-Cookie"
	XSXRF_PARAM_NAME  = "xsrft"
)

//Used to declare a new service. 
//See code example below:
//
//
//	type HelloService struct {
//	    gorest.RestService `root:"/tutorial/"`
//	    helloWorld  gorest.EndPoint `method:"GET" path:"/hello-world/" output:"string"`
//	    sayHello    gorest.EndPoint `method:"GET" path:"/hello/{name:string}" output:"string"`
//	}
//
type RestService struct {
	Context *Context
	rb	*ResponseBuilder
}

//Used to declare and EndPoint, wich represents a single point of entry to gorest applications, via a URL.
//See code example below:
//
//	type HelloService struct {
//	    gorest.RestService `root:"/tutorial/"`
// 	   helloWorld  gorest.EndPoint `method:"GET" path:"/hello-world/" output:"string"`
// 	   sayHello    gorest.EndPoint `method:"GET" path:"/hello/{name:string}" output:"string"`
//	}
//
type EndPoint bool

//Returns the ResponseBuilder associated with the current Context and Request. 
//This can be called multiple times within a service method, the same instance will be returned.
func (serv RestService) RB() *ResponseBuilder {
	return serv.ResponseBuilder()
}

//Returns the ResponseBuilder associated with the current Context and Request. 
//This can be called multiple times within a service method, the same instance will be returned.
func (serv RestService) ResponseBuilder() *ResponseBuilder {
	if serv.rb == nil {
		serv.rb = &ResponseBuilder{serv.Context}
//		serv.rb.ctx = serv.Context
	}
	return serv.rb
}

//Get the SessionData associated with the current request, as stored in the Context.
func (serv RestService) Session() SessionData {
	return serv.Context.sessData
}

//Get the SessionData associated with the current request, as stored in the Context.
func (this *ResponseBuilder) Session() SessionData {
	return this.ctx.sessData
}

//Get the value of the item referenced by key from the Session Data
func (sess SessionData) Get(key string) (interface{}, bool) {
	value, found := sess.relSessionData[key]
	return value, found
}

//Get the value of the item referenced by key as a string from the Session Data
func (sess SessionData) GetString(key string) (string, bool) {
	svalue := ""
	ivalue, found := sess.relSessionData[key]
	if found {
		svalue = ivalue.(string)
	}

	return svalue, found
}

// Set the value of the item referenced by key in the Session Data
func (sess SessionData) Set(key string, value interface{}) {
	sess.relSessionData[key] = value
}

//Returns a *http.Request associated with this Context
func (serv RestService) Request() *http.Request {
	return serv.rb.ctx.request
}

//Facilitates the construction of the response to be sent to the client.
type ResponseBuilder struct {
	ctx *Context

}

//Returns the Authorization token associated with the current request and hence session.
//This token is passed via the Authorization Header - supports Authorization: Bearer <token>
//The authorizer determines how the token is applied to securing requests
func (this *ResponseBuilder) SessionToken() string {
	return this.ctx.xsrftoken
}

//Sets the "xsrftoken" token associated with the current request and hence session, only valid for the sepcified root path and period.
//This creates a cookie and sets an http header with the name "X-Xsrf-Cookie"
func (this *ResponseBuilder) SetSessionToken(token string, path string, expires time.Time) {
	this.ctx.xsrftoken = token
	this.SetHeader(XSXRF_COOKIE_NAME, token)
	http.SetCookie(this.ctx.writer, &http.Cookie{Name: XSXRF_COOKIE_NAME, Value: token, Path: path, Expires: expires})
}

//This cleares the "xsrftoken" token associated with the current request and hence session. 
//Calling this will unlink the current session, making it un-addressable/invalid. Therefore if maintaining a session store
//you may want to evict/destroy the session there.
func (this *ResponseBuilder) RemoveSessionToken(path string) {
	this.SetSessionToken("", path, time.Unix(0, 0).UTC())
}
func (this *ResponseBuilder) writer() http.ResponseWriter {
	return this.ctx.writer
}

//Set the http code to be sent with the response, to the client.
func (this *ResponseBuilder) SetResponseCode(code int) *ResponseBuilder {
	this.ctx.responseCode = code
	logger.SetResponseCode(code)
	return this
}

//Set the http message to be sent with the response, to the client.
func (this *ResponseBuilder) SetResponseMsg(message string) *ResponseBuilder {
	this.ctx.responseMsg = message
	return this
}

//Set the content type of the http entity that is to be sent to the client.
func (this *ResponseBuilder) SetContentType(mime string) *ResponseBuilder {
	this.ctx.responseMimeSet = true
	this.writer().Header().Set("Content-Type", mime)
	return this
}

//This indicates whether the data returned by the endpoint service method should be ignored or appendend to the data
//already writen to the response via ResponseBuilder. A vlaue of "true" will discard, while a value of "false"" will append.
func (this *ResponseBuilder) Overide(overide bool) {
	this.ctx.overide = overide
}

type SessionData struct {
	relSessionData		map[string]interface{}
}

type Context struct {
	writer         http.ResponseWriter
	request        *http.Request
	xsrftoken      string
	sessData       SessionData

	// Response
	respPacket		io.ReadCloser

	//Response flags
	overide            bool
	responseCode       int
	responseMsg	   string
	responseMimeSet    bool
	responseMimeType   string
	dataHasBeenWritten bool
	encodeGzip	   bool
}

//This will write to the response and then call Overide(true), even if it had been set to "false" in a previous call.
func (this *ResponseBuilder) WriteAndOveride(data []byte) *ResponseBuilder {
	this.ctx.overide = true
	return this.Write(data)
}

//This will write to the response and then call Overide(false), even if it had been set to "true" in a previous call.
func (this *ResponseBuilder) WriteAndContinue(data []byte) *ResponseBuilder {
	this.ctx.overide = false
	return this.Write(data)
}

//This will write to the response and then call Overide(false), even if it had been set to "true" in a previous call.
func (this *ResponseBuilder) WritePacket() *ResponseBuilder {
	if !this.ctx.dataHasBeenWritten {
		if this.ctx.responseCode == 0 {
			this.SetResponseCode(getDefaultResponseCode(this.ctx.request.Method))
		}

		if !this.ctx.responseMimeSet {
			this.writer().Header().Set("Content-Type", this.ctx.responseMimeType)
		}

		if this.ctx.respPacket == nil {
			this.writer().WriteHeader(this.ctx.responseCode)
			this.writer().Write([]byte(this.ctx.responseMsg))
		} else if this.ctx.encodeGzip && strings.Contains(this.ctx.request.Header.Get("Accept-Encoding"), "gzip") {
			this.writer().Header().Set("Content-Encoding", "gzip")
			this.writer().WriteHeader(this.ctx.responseCode)
			gzipWriter := gzip.NewWriter(this.writer())
			defer gzipWriter.Close()
			io.Copy(gzipWriter, this.ctx.respPacket)
		} else {
			this.writer().WriteHeader(this.ctx.responseCode)
			io.Copy(this.writer(), this.ctx.respPacket)
		}
		this.ctx.dataHasBeenWritten = true
	}

	return this
}

//This will just write to the response without affecting the change done by a call to Overide().
func (this *ResponseBuilder) Write(data []byte) *ResponseBuilder {
	if this.ctx.responseCode == 0 {
		this.SetResponseCode(getDefaultResponseCode(this.ctx.request.Method))
	}
	if !this.ctx.dataHasBeenWritten {
		//TODO: Check for content type set.......
		this.writer().WriteHeader(this.ctx.responseCode)
	}

	this.writer().Write(data)
	this.ctx.dataHasBeenWritten = true
	return this
}

func (this *ResponseBuilder) LongPoll(delay int, producer func(interface{}) interface{}) *ResponseBuilder {

	return this
}

func (this *ResponseBuilder) SetHeader(key string, value string) *ResponseBuilder {
	this.writer().Header().Set(key, value)
	return this
}

//Used to add gereric/custom headers to the response.
//Good use for proxying and cross origins/site stuff.
//Example usage:
//
//
//	rb := serv.ResponseBuilder()
//	rb.AddHeader("Access-Control-Allow-Origin","http://127.0.0.1:8888")
//	rb.AddHeader("Access-Control-Allow-Headers","X-HTTP-Method-Override")
//	rb.AddHeader("Access-Control-Allow-Headers","X-Xsrf-Cookie")
//	rb.AddHeader("Access-Control-Expose-Headers","X-Xsrf-Cookie")
//		
func (this *ResponseBuilder) AddHeader(key string, value string) *ResponseBuilder {
	this.writer().Header().Add(key, value)
	return this
}
func (this *ResponseBuilder) DelHeader(key string) *ResponseBuilder {
	this.writer().Header().Del(key)
	return this
}
