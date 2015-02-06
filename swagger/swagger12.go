package swagger

import (
	"github.com/rmullinnix/gorest"
	"github.com/rmullinnix/logger"
	"strings"
	"reflect"
	"regexp"
	"strconv"
)

// Swagger 1.2 Specification Structures
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

var spec12		*SwaggerAPI12

func newSpec12(basePath string, numSvcTypes int, numEndPoints int) *SwaggerAPI12 {
	spec12 = new(SwaggerAPI12)

	spec12.SwaggerVersion = "1.2"
	spec12.APIVersion	= ""
	spec12.BasePath = basePath
	spec12.ResourcePath = ""
	spec12.APIs = make([]API, numEndPoints)
	spec12.Produces = make([]string, 0)
	spec12.Consumes = make([]string, numSvcTypes)
	spec12.Authorizations = make(map[string]Authorization, 0)
	spec12.Models = make(map[string]Model, 0)

	return spec12
}

func _spec12() *SwaggerAPI12 {
	return spec12
}

func swaggerDocumentor12(basePath string, svcTypes map[string]gorest.ServiceMetaData, endPoints map[string]gorest.EndPointStruct) interface{} {
	spec12 = newSpec12(basePath, len(svcTypes), len(endPoints))

	x := 0
	var svcInt 	reflect.Type 
	for _, st := range svcTypes {
		spec12.Produces = append(spec12.Produces, st.ProducesMime...)
		spec12.Consumes[x] = st.ConsumesMime
	
        	svcInt = reflect.TypeOf(st.Template)

	        if svcInt.Kind() == reflect.Ptr {
	                svcInt = svcInt.Elem()
       		}

		if field, found := svcInt.FieldByName("RestService"); found {
			temp := strings.Join(strings.Fields(string(field.Tag)), " ")
			tags := reflect.StructTag(temp)
			if tag := tags.Get("sw.apiVersion"); tag != "" {
				spec12.APIVersion = tag
			}
		}
	}

	// skip authorizations for now

	x = 0
	for _, ep := range endPoints {
		var api		API

		api.Path = cleanPath(ep.Signiture)
		//api.Description = ep.description

		var op		Operation

		api.Operations = make([]Operation, 1)

		if field, found := svcInt.FieldByName(ep.Name); found {
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
				op.Nickname = ep.Name
			}

			op.Responses = populateResponses(tags)
		}

		op.Method = ep.RequestMethod
		if strings.Index(ep.OutputType, ".") > 0 {
			op.Type = ep.OutputType[strings.Index(ep.OutputType, ".")+1:]
		} else {
			op.Type = ep.OutputType
		}
		op.Parameters = make([]Parameter, len(ep.Params) + len(ep.QueryParams))
		op.Authorizations = make([]Authorization, 0)
		pnum := 0
		for j := 0; j < len(ep.Params); j++ {
			var par		Parameter

			par.ParamType = "path"
			par.Name = ep.Params[j].Name
			par.Type = ep.Params[j].TypeName
			par.Description = ""
			par.Required = true
			par.AllowMultiple = false

			op.Parameters[pnum] = par
			pnum++
		}

		for j := 0; j < len(ep.QueryParams); j++ {
			var par		Parameter

			par.ParamType = "query"
			par.Name = ep.QueryParams[j].Name
			par.Type = ep.QueryParams[j].TypeName
			par.Description = ""
			par.Required = false
			par.AllowMultiple = false

			op.Parameters[pnum] = par
			pnum++
		}

		if ep.PostdataType != "" {
			var par		Parameter

			par.ParamType = "body"
			par.Name = ep.PostdataType
			par.Type = ep.PostdataType
			par.Description = ""
			par.Required = true
			par.AllowMultiple = false

			op.Parameters = append(op.Parameters, par)
		}

		api.Operations[0] = op
		spec12.APIs[x] = api
		x++

		methType := svcInt.Method(ep.MethodNumberInParent).Type
		// skip the fuction class pointer
		for i := 1; i < methType.NumIn(); i++ {
			inType := methType.In(i)
			if inType.Kind() == reflect.Struct {
				if _, ok := spec12.Models[inType.Name()]; ok {
					continue  // model already exists
				}

				model := populateModel(inType)

				spec12.Models[model.ID] = model
			}
		}

		for i := 0; i < methType.NumOut(); i++ {
			outType := methType.Out(i)
			if outType.Kind() == reflect.Struct {
				if _, ok := spec12.Models[outType.Name()]; ok {
					continue  // model already exists
				}

				model := populateModel(outType)

				spec12.Models[model.ID] = model
			}
		}
	}	

	return *spec12
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
			case reflect.Ptr:
				logger.Error.Println("Ptr ", sMem.Name)
				prop, required := populatePropertyPtr(sMem)
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

	if sf.Type.Kind() == reflect.Struct {
		parts := strings.Split(sf.Type.String(), ".")
		if len(parts) > 1 {
			prop.Type = parts[1]
		} else {
			prop.Type = parts[0]
		}

		if _, ok := spec12.Models[sf.Type.Name()]; !ok {
			model := populateModel(sf.Type)
			_spec12().Models[model.ID] = model
		}
	} else {
		prop.Type, prop.Format = primitiveFormat(sf.Type.String())
	}

	var tag         string

        if tag = tags.Get("sw.format"); tag != "" {
                prop.Format = tag
        } else {
		if prop.Format == "" {
			prop.Format = prop.Type
		}
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
		prop.Items.Type, _ = primitiveFormat(parts[1])
	} else {
		prop.Items.Type, _ = primitiveFormat(parts[0])
	}

	if et.Kind() == reflect.Struct {
		if _, ok := spec12.Models[et.Name()]; !ok {
			var placeHolder		Model
			_spec12().Models[et.Name()] = placeHolder
			model := populateModel(et)
			_spec12().Models[model.ID] = model
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

func populatePropertyPtr(sf reflect.StructField) (Property, bool) {
	var prop	Property

	stmp := strings.Join(strings.Fields(string(sf.Tag)), " ")
	tags := reflect.StructTag(stmp)

	// remove the package if present
	et := sf.Type.Elem()
	parts := strings.Split(et.String(), ".")
	if len(parts) > 1 {
		prop.Type = parts[1]
	} else {
		prop.Type = parts[0]
	}

	if et.Kind() == reflect.Struct {
		if _, ok := spec12.Models[et.Name()]; !ok {
			var placeHolder		Model
			_spec12().Models[et.Name()] = placeHolder
			model := populateModel(et)
			_spec12().Models[model.ID] = model
		}
	}

	prop.Format = prop.Type

        if tag := tags.Get("sw.description"); tag != "" {
                prop.Description = tag
        }

	required := false
        if tag := tags.Get("sw.required"); tag != "" {
		if tag == "true" {
                	required = true
		}
        }

	return prop, required
}
