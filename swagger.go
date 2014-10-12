package gorest

import (
	"strings"
	"reflect"
)

type SwaggerAPI12 struct {
	SwaggerVersion	string 			`json:"swaggerVersion"`
	APIVersion	string			`json:"apiVersion"`
	BasePath	string			`json:"basePath"`
	ResourcePath	string			`json:"resourcePath"`
	APIs		[]API			`json:"apis"`
	Models		map[string]Model	`json:"models"`
	Produces	[]string		`json:"produces"`
	Consumes	[]string		`json:"consumes"`
	Authorizations	map[string]Authorization `json:"authorizations"`
}


type API struct {
	Path		string			`json:"path"`
	Description	string			`json:"description,omitempty"`
	Operations	[]Operation		`json:"operations"`
}

type Operation struct {
	Method		string			`json:"method"`
	Type		string			`json:"type"`
	Summary		string			`json:"summary,omitempty"`
	Notes		string			`json:"notes,omitempty"`
	Nickname	string			`json:"nickname"`
	Authorizations	[]Authorization		`json:"authorizations"`
	Parameters	[]Parameter		`json:"parameters"`
	Responses	[]ResponseMessage	`json:"responseMessages"`
	Produces	[]string		`json:"produces,omitempty"`
	Consumes	[]string		`json:"consumes,omitempty"`
	Depracated	string			`json:"depracated,omitempty"`
}

type Parameter struct {
	ParamType	string			`json:"paramType"`
	Name		string			`json:"name"`
	Type		string			`json:"type"`
	Description	string			`json:"description,omitempty"`
	Required	bool			`json:"required,omitempty"`
	AllowMultiple	bool			`json:"allowMultiple,omitempty"`
}

type ResponseMessage struct {
	Code		int			`json:"code"`
	Message		string			`json:"message"`
	ResponseModel	string			`json:"responseModel,omitempty"`
}

type Model struct {
	ID		string			`json:"id"`
	Description	string			`json:"description,omitempty"`
	Required	[]string		`json:"required,omitempty"`
	Properties	map[string]Property 	`json:"properties"`
	SubTypes	[]string		`json:"subTypes,omitempty"`
	Discriminator	string			`json:"discriminator,omitempty"`
}

type Property struct {
	Type		string			`json:"type"`
	Format		string			`json:"format,omitempty"`
	Description	string			`json:"description,omitempty"`
}

type Authorization struct {
	Scope		string			`json:"scope"`
	Description	string			`json:"description,omitempty"`
}

func buildSwaggerDoc(basePath string) SwaggerAPI12 {
	var spec		SwaggerAPI12

	spec.SwaggerVersion = "1.2"
	spec.APIVersion	= ""
	spec.BasePath = basePath
	spec.ResourcePath = ""
	spec.APIs = make([]API, len(_manager().endpoints))
	spec.Produces = make([]string, 0)
	spec.Consumes = make([]string, len(_manager().serviceTypes))
	spec.Authorizations = make(map[string]Authorization, 0)
	spec.Models = make(map[string]Model, 0)

	x := 0
	var svcInt 	reflect.Type 
	for _, st := range _manager().serviceTypes {
		spec.Produces = append(spec.Produces, st.producesMime...)
		spec.Consumes[x] = st.consumesMime
	
        	svcInt = reflect.TypeOf(st.template)

	        if svcInt.Kind() == reflect.Ptr {
	                svcInt = svcInt.Elem()
       		}

		if field, found := svcInt.FieldByName("RestService"); found {
			temp := strings.Join(strings.Fields(string(field.Tag)), " ")
			tags := reflect.StructTag(temp)
			if tag := tags.Get("sw.apiVersion"); tag != "" {
				spec.APIVersion = tag
			}
		}
	}

	// skip authorizations for now

	x = 0
	for _, ep := range _manager().endpoints {
		var api		API

		sig := strings.Split(ep.signiture, "?")
		api.Path = sig[0]
		//api.Description = ep.description

		var op		Operation

		api.Operations = make([]Operation, 1)
		op.Method = ep.requestMethod
		op.Type = ep.outputType
		op.Nickname = ep.name
		op.Parameters = make([]Parameter, len(ep.params) + len(ep.queryParams))
		op.Authorizations = make([]Authorization, 0)
		for j := 0; j < len(ep.params); j++ {
			var par		Parameter

			par.ParamType = "path"
			par.Name = ep.params[j].name
			par.Type = ep.params[j].typeName
			par.Description = ""
			par.Required = false
			par.AllowMultiple = false

			op.Parameters[j] = par
		}

		for j := len(ep.params); j < len(ep.params) + len(ep.queryParams); j++ {
			var par		Parameter

			par.ParamType = "query"
			par.Name = ep.queryParams[j].name
			par.Type = ep.queryParams[j].typeName
			par.Description = ""
			par.Required = false
			par.AllowMultiple = false

			op.Parameters[j] = par
		}

		var resp	ResponseMessage

		op.Responses = make([]ResponseMessage, 1)
		resp.Code = 200
		resp.Message = "OK"
		op.Responses[0] = resp

		api.Operations[0] = op
		spec.APIs[x] = api
		x++

		methType := svcInt.Method(ep.methodNumberInParent).Type
		// skip the fuction class pointer
		for i := 1; i < methType.NumIn(); i++ {
			inType := methType.In(i)
			if inType.Kind() == reflect.Struct {
				if _, ok := spec.Models[inType.Name()]; ok {
					continue  // model already exists
				}

				model := populateModel(inType)

				spec.Models[model.ID] = model
			}
		}

		for i := 0; i < methType.NumOut(); i++ {
			outType := methType.Out(i)
			if outType.Kind() == reflect.Struct {
				if _, ok := spec.Models[outType.Name()]; ok {
					continue  // model already exists
				}

				model := populateModel(outType)

				spec.Models[model.ID] = model
			}
		}
	}	

	return spec
}

func populateModel(t reflect.Type) Model {
	var model	Model

	model.ID = t.Name()
	model.Description = ""
	model.Required = make([]string, 0)
	model.Properties = make(map[string]Property)

	for k := 0; k < t.NumField(); k++ {
		sMem := t.Field(k)
		prop, required := populateProperty(sMem)
		model.Properties[sMem.Name] = prop

		if required {
			model.Required = append(model.Required, sMem.Name)
		}
	}

	return model
}

func populateProperty(sf reflect.StructField) (Property, bool) {
	var prop	Property

	stmp := strings.Join(strings.Fields(string(sf.Tag)), " ")
	tags := reflect.StructTag(stmp)
	prop.Type = sf.Type.String()
	prop.Format = sf.Type.String()

        var tag         string

        if tag = tags.Get("sw.format"); tag != "" {
                prop.Format = tag
        }

        if tag = tags.Get("sw.description"); tag != "" {
                prop.Description = tag
        }

	required := false
        if tag = tags.Get("sw.required"); tag != "" {
		if tag == "true" {
                	required = true
		}
        }

	return prop, required
}
