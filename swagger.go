package gorest

import (
	"strings"
	"reflect"
	"regexp"
	"strconv"
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
	Properties	map[string]interface{} 	`json:"properties"`
	SubTypes	[]string		`json:"subTypes,omitempty"`
	Discriminator	string			`json:"discriminator,omitempty"`
}

type Property struct {
	Type		string			`json:"type"`
	Format		string			`json:"format,omitempty"`
	Description	string			`json:"description,omitempty"`
}

type PropertyArray struct {
	Type		string			`json:"type"`
	Format		string			`json:"format,omitempty"`
	Description	string			`json:"description,omitempty"`
	Items		Property		`json:"items"`
}

type Authorization struct {
	Scope		string			`json:"scope"`
	Description	string			`json:"description,omitempty"`
}

var spec		*SwaggerAPI12

func newSpec(basePath string) *SwaggerAPI12 {
	spec = new(SwaggerAPI12)

	spec.SwaggerVersion = "1.2"
	spec.APIVersion	= ""
	spec.BasePath = basePath
	spec.ResourcePath = ""
	spec.APIs = make([]API, len(_manager().endpoints))
	spec.Produces = make([]string, 0)
	spec.Consumes = make([]string, len(_manager().serviceTypes))
	spec.Authorizations = make(map[string]Authorization, 0)
	spec.Models = make(map[string]Model, 0)

	return spec
}

func _spec() *SwaggerAPI12 {
	return spec
}

func buildSwaggerDoc(basePath string) SwaggerAPI12 {
	spec = newSpec(basePath)

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

		api.Path = cleanPath(ep.signiture)
		//api.Description = ep.description

		var op		Operation

		api.Operations = make([]Operation, 1)

		if field, found := svcInt.FieldByName(ep.name); found {
			temp := strings.Join(strings.Fields(string(field.Tag)), " ")
			tags := reflect.StructTag(temp)
			if tag := tags.Get("sw.summary"); tag != "" {
				op.Summary = tag
			}

			if tag := tags.Get("sw.notes"); tag != "" {
				op.Notes = tag
			}

			if tag := tags.Get("sw.nickname"); tag != "" {
				op.Nickname = tag
			} else {
				op.Nickname = ep.name
			}

			op.Responses = populateResponses(tags)
		}

		op.Method = ep.requestMethod
		op.Type = ep.outputType
		op.Parameters = make([]Parameter, len(ep.params) + len(ep.queryParams))
		op.Authorizations = make([]Authorization, 0)
		for j := 0; j < len(ep.params); j++ {
			var par		Parameter

			par.ParamType = "path"
			par.Name = ep.params[j].name
			par.Type = ep.params[j].typeName
			par.Description = ""
			par.Required = true
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

		if ep.postdataType != "" {
			var par		Parameter

			par.ParamType = "body"
			par.Name = ep.postdataType
			par.Type = ep.postdataType
			par.Description = ""
			par.Required = true
			par.AllowMultiple = false

			op.Parameters = append(op.Parameters, par)
		}

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

	return *spec
}

func cleanPath(inPath string) string {
	sig := strings.Split(inPath, "?")
	parts := strings.Split(sig[0], "{")

	path := parts[0]
	for i := 1; i < len(parts); i++ {
		pathVar := strings.Split(parts[i], ":")
		remPath := strings.Split(pathVar[1], "}")
		path = path + "{" + pathVar[0] + "}" + remPath[1]
	}

	return path
}

func populateResponses(tags reflect.StructTag) []ResponseMessage {
	var responses	[]ResponseMessage
	var tag		string

	responses = make([]ResponseMessage, 0)
	if tag = tags.Get("sw.response"); tag != "" {
		reg := regexp.MustCompile("{[^}]+}")
		parts := reg.FindAllString(tag, -1)
		for i := 0; i < len(parts); i++ {
			var resp	ResponseMessage

			cd_msg := strings.Split(parts[i], ":")
			resp.Code, _ = strconv.Atoi(strings.TrimPrefix(cd_msg[0], "{"))
			resp.Message = strings.TrimSuffix(cd_msg[1], "}")

			responses = append(responses, resp)
		}
	}
	return responses
}

func populateModel(t reflect.Type) Model {
	var model	Model

	model.ID = t.Name()
	model.Description = ""
	model.Required = make([]string, 0)
	model.Properties = make(map[string]interface{})

	for k := 0; k < t.NumField(); k++ {
		sMem := t.Field(k)
		switch sMem.Type.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				prop, required := populatePropertyArray(sMem)
				model.Properties[sMem.Name] = prop
				if required {
					model.Required = append(model.Required, sMem.Name)
				}
			default:
				prop, required := populateProperty(sMem)
				model.Properties[sMem.Name] = prop
				if required {
					model.Required = append(model.Required, sMem.Name)
				}
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

func populatePropertyArray(sf reflect.StructField) (PropertyArray, bool) {
	var prop	PropertyArray

	stmp := strings.Join(strings.Fields(string(sf.Tag)), " ")
	tags := reflect.StructTag(stmp)
	prop.Type = "array"

	// remove the package if present
	et := sf.Type.Elem()
	parts := strings.Split(et.String(), ".")
	if len(parts) > 1 {
		prop.Items.Type = parts[1]
	} else {
		prop.Items.Type = parts[0]
	}

	if et.Kind() == reflect.Struct {
		if _, ok := spec.Models[et.Name()]; !ok {
			model := populateModel(et)
			_spec().Models[model.ID] = model
		}
	}

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
