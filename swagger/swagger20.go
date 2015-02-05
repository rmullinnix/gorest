package swagger

import (
	"github.com/rmullinnix/gorest"
	"strings"
	"reflect"
	"regexp"
)

// Swagger 2.0 Specifiction Structures
// This is the root document object for the API specification. It combines what previously was
// the Resource Listing and API Declaration (version 1.2 and earlier) together into one document.
type SwaggerAPI20 struct {
	SwaggerVersion	string			`json:"swagger"`
	Info		InfoObject		`json:"info"`
	Host		string			`json:"host"`
	BasePath	string			`json:"basePath"`
	Schemes		[]string		`json:"schemes,omitempty"`
	Consumes	[]string		`json:"consumes"`
	Produces	[]string		`json:"produces"`
	Paths		map[string]PathItem	`json:"paths"`
	Definitions	map[string]SchemaObject	`json:"definitions"`
	Parameters	map[string]ParameterObject	`json:"parameters,omitempty"`
	Responses	map[string]ResponseObject	`json:"responses,omitempty"`
	SecurityDefs	map[string]SecurityScheme	`json:"securityDefinitions,omitempty"`
	Security	*SecurityRequirement	`json:"security,omitempty"`
	Tags		[]Tag			`json:"tags,omitempty"`
	ExternalDocs	*ExtDocObject		`json:"externalDocs,omitempty"`
}

// The object provides metadata about the API. The metadata can be used by the clients if needed,
// and can be presented in the Swagger-UI for convenience.
type InfoObject struct {
	Title		string			`json:"title"`
	Description	string			`json:"description"`
	TermsOfService	string			`json:"termsOfService"`
	Contact		ContactObject		`json:"contact"`
	License		LicenseObject		`json:"license"`
	Version		string			`json:"version"`
}

// Contact information for the exposed API
type ContactObject struct {
	Name		string			`json:"name"`
	Url		string			`json:"url"`
	Email		string			`json:"email"`
}

// License information for the exposed API
type LicenseObject struct {
	Name		string			`json:"name"`
	Url		string			`json:"url"`
}

// Paths Object
// Holds the relative paths to the individual endpoints. The path is appended to the basePath
// in order to construct the full URL. The Paths may be empty, due to ACL constraints.
// Paths is a map[string]PathItem where the string is the /{path}
//
// Path Item - Describes the operations available on a single path. A Path Item may be empty, 
// due to ACL constraints. The path itself is still exposed to the documentation viewer but they
// will not know which operations and parameters are available.
//   todo -- more than likely, can use map[string]OperationObject with the key being the http method
type PathItem struct {
	Ref		string			`json:"$ref,omitempty"`
	Get		*OperationObject	`json:"get,omitempty"`
	Put		*OperationObject	`json:"put,omitempty"`
	Post		*OperationObject	`json:"post,omitempty"`
	Delete		*OperationObject	`json:"delete,omitempty"`
	Options		*OperationObject	`json:"options,omitempty"`
	Head		*OperationObject	`json:"head,omitempty"`
	Patch		*OperationObject	`json:"patch,omitempty"`
	Parameters	[]ParameterObject	`json:"parameters,omitempty"`
}

// Describes a single API operation on a path
type OperationObject struct {
	Tags		[]string		`json:"tags"`
	Summary		string			`json:"summary,omitempty"`
	Description	string			`json:"description,omitempty"`
	ExternalDocs	*ExtDocObject		`json:"externalDocs,omitempty"`
	OperationId	string			`json:"operationId"`
	Consumes	[]string		`json:"consumes,omitempty"`
	Produces	[]string		`json:"produces,omitempty"`
	Parameters	[]ParameterObject	`json:"parameters,omitempty"`
	Responses	map[string]ResponseObject	`json:"responses"`
	Schemes		[]string		`json:"schemes,omitempty"`
	Deprecated	bool			`json:"deprecated,omitempty"`
	Security	[]SecurityRequirement	`json:"security,omitempty"`
}

// Allows Referencing an external resource for extended documentation
type ExtDocObject struct {
	Description	string			`json:"description,omitempty"`
	Url		string			`json:"url,omitempty"`
}

// Describes a single operation parameter
// A unique parameter is defined by a combination of a name and location
// There are five possible parameter types:  Path, Query, Header, Body, and Form
type ParameterObject struct {
	Name		string			`json:"name"`
	In		string			`json:"in"`
	Description	string			`json:"description,omitempty"`
	Required	bool			`json:"required,omitempty"`
	Schema		*SchemaObject		`json:"schema,omitempty"`
	Type		string			`json:"type,omitempty"`
	Format		string			`json:"format,omitempty"`
	Items		*ItemsObject		`json:"items,omitempty"`
	CollectionFormat	string		`json:"collectionFormat,omitempty"`
	Default		interface{}		`json:"default,omitempty"`
	Maximum		float64			`json:"maximum,omitempty"`
	ExclusiveMax	bool			`json:"exclusiveMaximum,omitempty"`
	Minimum		float64			`json:"minimum,omitempty"`
	ExclusiveMin	bool			`json:"exclusiveMinimum,omitempty"`
	MaxLength	int64			`json:"maxLength,omitempty"`
	MinLength	int64			`json:"minLength,omitempty"`
	Pattern		string			`json:"pattern,omitempty"`
	MaxItems	int64			`json:"maxItems,omitempty"`
	MinItems	int64			`json:"minItems,omitempty"`
	UniqueItems	bool			`json:"uniqueItems,omitempty"`
	Enum		[]interface{}		`json:"enum,omitempty"`
	MultipleOf	float64			`json:"multipleOf,omitempty"`
}

// A limited subset of JSON-Schema's items object.  It is used by parameter definitions that
// are not located in "body"
type ItemsObject struct {
	Type		string			`json:"type"`
	Format		string			`json:"format"`
	Items		*ItemsObject		`json:"items,omitempty"`
	CollectionFormat	string		`json:"collectionFormat,omitempty"`
	Default		interface{}		`json:"default,omitempty"`
	Maximum		float64			`json:"maximum,omitempty"`
	ExclusiveMax	bool			`json:"exclusiveMaximum,omitempty"`
	Minimum		float64			`json:"minimum,omitempty"`
	ExclusiveMin	bool			`json:"exclusiveMinimum,omitempty"`
	MaxLength	int64			`json:"maxLength,omitempty"`
	MinLength	int64			`json:"minLength,omitempty"`
	Pattern		string			`json:"pattern,omitempty"`
	MaxItems	int64			`json:"maxItems,omitempty"`
	MinItems	int64			`json:"minItems,omitempty"`
	UniqueItems	bool			`json:"uniqueItems,omitempty"`
	Enum		[]interface{}		`json:"enum,omitempty"`
	MultipleOf	float64			`json:"multipleOf,omitempty"`
}

// Responses Definition Ojbect - implement as a map[string]ResponseObject
// A container for the expected responses of an operation. The container maps a HTTP 
// response code to the expected response. It is not expected from the documentation to 
// necessarily cover all possible HTTP response codes, since they may not be known in advance.
// However, it is expected from the documentation to cover a successful operation response
// and any known errors.

// Describes a single respone from an API Operation
type ResponseObject struct {
	Description	string			`json:"description"`
	Schema		*SchemaObject		`json:"schema,omitempty"`
	Headers		map[string]HeaderObject	`json:"headers,omitempty"`
	Examples	map[string]interface{}	`json:"examples,omitempty"`
}

// Header that can be sent as part of a response
type HeaderObject struct {
	Description	string			`json:"description"`
	Type		string			`json:"type"`
	Format		string			`json:"format"`
	Items		ItemsObject		`json:"items,omitempty"`
	CollectionFormat	string		`json:"collectionFormat,omitempty"`
	Default		interface{}		`json:"default,omitempty"`
	Maximum		float64			`json:"maximum,omitempty"`
	ExclusiveMax	bool			`json:"exclusiveMaximum,omitempty"`
	Minimum		float64			`json:"minimum,omitempty"`
	ExclusiveMin	bool			`json:"exclusiveMinimum"`
	MaxLength	int64			`json:"maxLength,omitempty"`
	MinLength	int64			`json:"minLength,omitempty"`
	Pattern		string			`json:"pattern,omitempty"`
	MaxItems	int64			`json:"maxItems,omitempty"`
	MinItems	int64			`json:"minItems,omitempty"`
	UniqueItems	bool			`json:"uniqueItems,omitempty"`
	Enum		[]interface{}		`json:"enum,omitempty"`
	MultipleOf	float64			`json:"multipleOf,omitempty"`
}

// A simple object to allow referencing other definitions in the specification. 
// It can be used to reference parameters and responses that are defined at the top
// level for reuse.
type ReferenceObject struct {
	Ref		string			`json:"$ref"`
}

// The Schema Object allows the definition of input and output data types. These types
// can be objects, but also primitives and arrays. This object is based on the JSON Schema
// Specification Draft 4 and uses a predefined subset of it. On top of this subset,
// there are extensions provided by this specification to allow for more complete documentation.
type SchemaObject struct {
	Ref		string			`json:"$ref,omitempty"`
	Title		string			`json:"title,omitempty"`
	Description	string			`json:"description,omitempty"`
	Type		string			`json:"type,omitempty"`
	Format		string			`json:"format,omitempty"`
	Required	[]string		`json:"required,omitempty"`
	Items		*SchemaObject		`json:"items,omitempty"`
	MaxItems	int64			`json:"maxItems,omitempty"`
	MinItems	int64			`json:"minItems,omitempty"`
	Properties	map[string]SchemaObject	`json:"properties,omitempty"`
	MaxProperties	int64			`json:"maxProperties,omitempty"`
	MinProperties	int64			`json:"minProperties,omitempty"`
	AllOf		*SchemaObject		`json:"allOf,omitempty"`
	Default		interface{}		`json:"default,omitempty"`
	Maximum		float64			`json:"maximum,omitempty"`
	ExclusiveMax	bool			`json:"exclusiveMaximum,omitempty"`
	Minimum		float64			`json:"minimum,omitempty"`
	ExclusiveMin	bool			`json:"exclusiveMinimum,omitempty"`
	MaxLength	int64			`json:"maxLength,omitempty"`
	MinLength	int64			`json:"minLength,omitempty"`
	Pattern		string			`json:"pattern,omitempty"`
	UniqueItems	bool			`json:"uniqueItems,omitempty"`
	Enum		[]interface{}		`json:"enum,omitempty"`
	MultipleOf	float64			`json:"multipleOf,omitempty"`
	Discriminator	string			`json:"discriminator,omitempty"`
	ReadOnly	bool			`json:"readOnly,omitempty"`
	Xml		*XMLObject		`json:"xml,omitempty"`
	ExternalDocs	*ExtDocObject		`json:"externalDocs,omitempty"`
	Example		interface{}		`json:"example,omitempty"`
}

// A metadata object that allows for more fine-tuned XML model definitions
type XMLObject struct {
	Name		string			`json:"name,omitempty"`
	Namespace	string			`json:"namespace,omitempty"`
	Prefix		string			`json:"prefix,omitempty"`
	Attribute	bool			`json:"attribute,omitempty"`
	Wrapped		bool			`json:"wrapped,omitempty"`
}

// Allows the definition of a security scheme that can be used by the operations.
// Supported schemes are basic authentication, an API key (either as a header or as a
// query parameter) and OAth2's common flows (implicit, password, application and 
// access code).
type SecurityScheme struct {
	Type		string			`json:"type,omitempty"`
	Description	string			`json:"description,omitempty"`
	Name		string			`json:"name,omitempty"`
	In		string			`json:"in,omitempty"`
	Flow		string			`json:"flow,omitempty"`
	AuthorizationUrL	string		`json:"authorizationUrl,omitempty"`
	TokenUrl	string			`json:"tokenUrl,omitempty"`
	Scopes		map[string]string	`json:"scopes,omitempty"`
}

type SecurityDefObject struct {
}

type SecurityRequirement struct {
}

// Allows adding meta data to a single tag that is used by the Operation Object. 
// It is not mandatory to have a Tag Object per tag used there.
type Tag struct {
	Name		string			`json:"name"`
	Description	string			`json:"description"`
	ExternalDocs	ExtDocObject		`json:"externalDocs,omitempty"`
}

var spec20		*SwaggerAPI20

func newSpec20(basePath string, numSvcTypes int, numEndPoints int) *SwaggerAPI20 {
	spec20 = new(SwaggerAPI20)

	spec20.SwaggerVersion = "2.0"
	spec20.Host = strings.TrimSuffix(strings.TrimPrefix(basePath, "http://"), "/")
	spec20.BasePath = "/"
	spec20.Schemes = make([]string, 0)
	spec20.Consumes = make([]string, numSvcTypes)
	spec20.Produces = make([]string, 0)
	spec20.Paths = make(map[string]PathItem, numEndPoints)
	spec20.Definitions = make(map[string]SchemaObject, 0)
	spec20.Parameters = make(map[string]ParameterObject, 0)
	spec20.SecurityDefs = make(map[string]SecurityScheme, 0)
	spec20.Tags = make([]Tag, 0)

	return spec20
}

func _spec20() *SwaggerAPI20 {
	return spec20
}

func swaggerDocumentor20(basePath string, svcTypes map[string]gorest.ServiceMetaData, endPoints map[string]gorest.EndPointStruct) interface{} {
	spec20 = newSpec20(basePath, len(svcTypes), len(endPoints))

	x := 0
	var svcInt 	reflect.Type 
	for _, st := range svcTypes {
		spec20.Produces = append(spec20.Produces, st.ProducesMime...)
		spec20.Consumes[x] = st.ConsumesMime
	
        	svcInt = reflect.TypeOf(st.Template)

	        if svcInt.Kind() == reflect.Ptr {
	                svcInt = svcInt.Elem()
       		}

		if field, found := svcInt.FieldByName("RestService"); found {
			temp := strings.Join(strings.Fields(string(field.Tag)), " ")
			tags := reflect.StructTag(temp)
			spec20.Info = populateInfoObject(tags)
		}
	}

	// skip authorizations for now

	x = 0
	for _, ep := range endPoints {
		var api		PathItem

		path := "/" + cleanPath(ep.Signiture)

		var op		OperationObject

		if field, found := svcInt.FieldByName(ep.Name); found {
			temp := strings.Join(strings.Fields(string(field.Tag)), " ")
			tags := reflect.StructTag(temp)
			op = populateOperationObject(tags, ep)
		}

		op.Consumes = spec20.Consumes
		op.Produces = spec20.Produces

		switch (ep.RequestMethod) {
		case "GET":
			api.Get = &op
		case "POST":
			api.Post = &op
		case "PUT":
			api.Put = &op
		case "DELETE":
			api.Delete = &op
		case "OPTIONS":
			api.Options = &op
		case "PATCH":
			api.Patch = &op
		case "HEAD":
			api.Head = &op
		}

		op.Parameters = make([]ParameterObject, len(ep.Params) + len(ep.QueryParams))
		pnum := 0
		for j := 0; j < len(ep.Params); j++ {
			var par		ParameterObject

			par.In = "path"
			par.Name = ep.Params[j].Name
			par.Type = ep.Params[j].TypeName
			par.Format = ep.Params[j].TypeName
			par.Description = ""
			par.Required = true

			op.Parameters[pnum] = par
			pnum++
		}

		for j := 0; j < len(ep.QueryParams); j++ {
			var par		ParameterObject

			par.In = "query"
			par.Name = ep.QueryParams[j].Name
			par.Type = ep.QueryParams[j].TypeName
			par.Format = ep.QueryParams[j].TypeName
			par.Description = ""
			par.Required = false

			op.Parameters[pnum] = par
			pnum++
		}

		if ep.PostdataType != "" {
			var par		ParameterObject

			par.In = "body"
			par.Name = ep.PostdataType
			par.Description = ""
			par.Required = true

			var schema	SchemaObject
			schema.Ref = "#/definitions/" + ep.PostdataType
			par.Schema = &schema

			op.Parameters = append(op.Parameters, par)
		}

		spec20.Paths[path] = api

		x++

		methType := svcInt.Method(ep.MethodNumberInParent).Type
		// skip the fuction class pointer
		for i := 1; i < methType.NumIn(); i++ {
			inType := methType.In(i)
			if inType.Kind() == reflect.Struct {
				if _, ok := spec20.Definitions[inType.Name()]; ok {
					continue  // definition already exists
				}

				schema := populateDefinitions(inType)

				spec20.Definitions[inType.Name()] = schema
			}

			// inType.Kind() == reflect.Slice (arrays)
		}

		for i := 0; i < methType.NumOut(); i++ {
			outType := methType.Out(i)
			if outType.Kind() == reflect.Struct {
				if _, ok := spec20.Definitions[outType.Name()]; ok {
					continue  // definition already exists
				}

				schema := populateDefinitions(outType)

				spec20.Definitions[outType.Name()] = schema
			}
			// inType.Kind() == reflect.Slice (arrays)
		}
	}	

	return *spec20
}

func populateInfoObject(tags reflect.StructTag) InfoObject {
	var info	InfoObject

	if tag := tags.Get("sw.title"); tag != "" {
		info.Title = tag
	}
	if tag := tags.Get("sw.description"); tag != "" {
		info.Description = tag
	}
	if tag := tags.Get("sw.termsOfService"); tag != "" {
		info.TermsOfService = tag
	}
	if tag := tags.Get("sw.apiVersion"); tag != "" {
		info.Version = tag
	}
	if tag := tags.Get("sw.contactName"); tag != "" {
		info.Contact.Name = tag
	}
	if tag := tags.Get("sw.contactUrl"); tag != "" {
		info.Contact.Url = tag
	}
	if tag := tags.Get("sw.contactEmail"); tag != "" {
		info.Contact.Email = tag
	}
	if tag := tags.Get("sw.licenseName"); tag != "" {
		info.License.Name = tag
	}
	if tag := tags.Get("sw.licenseUrl"); tag != "" {
		info.License.Url = tag
	}
	
	return info
}

func populateOperationObject(tags reflect.StructTag, ep gorest.EndPointStruct) OperationObject {
	var op	OperationObject

	op.Tags = make([]string, 0)

	if tag := tags.Get("sw.summary"); tag != "" {
		op.Summary = tag
	}
	if tag := tags.Get("sw.notes"); tag != "" {
		op.Description = tag
	}
	if tag := tags.Get("sw.description"); tag != "" {
		op.Description = tag
	}
	if tag := tags.Get("sw.nickname"); tag != "" {
		op.OperationId = tag
	}
	if tag := tags.Get("sw.operationId"); tag != "" {
		op.OperationId = tag
	}
	if op.OperationId == "" {
		op.OperationId = ep.Name
	}

	op.Tags = append(op.Tags, "Challenge")
	op.Responses = populateResponseObject(tags, ep)
	return op
}

func populateResponseObject(tags reflect.StructTag, ep gorest.EndPointStruct) map[string]ResponseObject {
	var responses	map[string]ResponseObject
	var tag		string

	responses = make(map[string]ResponseObject, 0)
	if tag = tags.Get("sw.response"); tag != "" {
		reg := regexp.MustCompile("{[^}]+}")
		parts := reg.FindAllString(tag, -1)
		for i := 0; i < len(parts); i++ {
			var resp	ResponseObject

			resp.Headers = make(map[string]HeaderObject, 0);
			cd_msg := strings.Split(parts[i], ":")
			code := strings.TrimPrefix(cd_msg[0], "{")
			if len(cd_msg) == 2 {
				resp.Description = strings.TrimSuffix(cd_msg[1], "}")
			} else {
				resp.Description = cd_msg[1]
				
				if cd_msg[2] == "output}" {
					var schema	SchemaObject

					if ep.OutputTypeIsArray {
						schema.Type = "array"
						var items	SchemaObject
						items.Type = ep.OutputType
						items.Format = ep.OutputType
						schema.Items = &items
					} else {
						schema.Ref = "#/definitions/" + ep.OutputType
					}
					resp.Schema = &schema
				}
			}

			responses[code] = resp
		}
	}
	return responses
}

func populateDefinitions(t reflect.Type) SchemaObject {
	var model	SchemaObject

	model.Description = ""			// not able to tag struct definition
	model.Required = make([]string, 0)
	model.Properties = make(map[string]SchemaObject)

	for k := 0; k < t.NumField(); k++ {
		sMem := t.Field(k)
		switch sMem.Type.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
		//		prop, required := populatePropertyArray(sMem)
		//		model.Properties[sMem.Name] = prop
		//		if required {
		//			model.Required = append(model.Required, sMem.Name)
		//		}
			default:
				prop, required := populateDefinition(sMem)
				model.Properties[sMem.Name] = prop
				if required {
					model.Required = append(model.Required, sMem.Name)
				}
		}
	}

	return model
}

func populateDefinition(sf reflect.StructField) (SchemaObject, bool) {
	var prop	SchemaObject

	stmp := strings.Join(strings.Fields(string(sf.Tag)), " ")
	tags := reflect.StructTag(stmp)
	if sf.Type.String() == "bool" {
		prop.Type = "boolean"
	} else {
		prop.Type = sf.Type.String()
	}

	if sf.Type.String() == "time.Time" {
		prop.Type = "string"
		prop.Format = "date-time"
	} else if sf.Type.String() == "int" {
		prop.Type = "integer"
		prop.Format = "int32"
	} else if sf.Type.String() == "float32" {
		prop.Type = "number"
		prop.Format = "float"
	} else if sf.Type.Kind() == reflect.Struct {
		parts := strings.Split(sf.Type.String(), ".")
		if len(parts) > 1 {
			prop.Type = parts[1]
		} else {
			prop.Type = parts[0]
		}

		if _, ok := spec20.Definitions[sf.Type.Name()]; !ok {
			schema := populateDefinitions(sf.Type)
			_spec20().Definitions[sf.Type.Name()] = schema
		}
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
/*
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
		if _, ok := spec20.Models[et.Name()]; !ok {
			model := populateModel(et)
			_spec20().Models[model.ID] = model
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
*/
