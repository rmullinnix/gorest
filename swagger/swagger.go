package swagger

import (
	"github.com/rmullinnix/gorest"
	"strings"
)

type dataType struct {
	swtype		string
	swformat	string
	primitive	bool
}

var primitives		map[string]dataType

// creates a new Swagger Documentor
//   versions supported - 1.2 and 2.0
func NewSwaggerDocumentor(version string) *gorest.Documentor {
	var doc		gorest.Documentor

	if version == "1.2" {
        	doc = gorest.Documentor{swaggerDocumentor12}
	} else if version == "2.0" {
		doc = gorest.Documentor{swaggerDocumentor20}
	}

	primitives = make(map[string]dataType)

	primitives["int"] = dataType{"integer", "int32", true}
	primitives["int32"] = dataType{"integer", "int32", true}
	primitives["int64"] = dataType{"long", "int64", true}
	primitives["uint32"] = dataType{"integer", "int32", true}
	primitives["uint64"] = dataType{"long", "int64", true}
	primitives["float32"] = dataType{"number", "float", true} 
	primitives["float64"] = dataType{"number", "float", true}
	primitives["string"] = dataType{"string", "", true}
	primitives["bool"] = dataType{"boolean", "", true}
	primitives["date"] = dataType{"string", "date", true}
	primitives["time.Time"] = dataType{"string", "dateTime", true}
	primitives["Time"] = dataType{"string", "dateTime", true}
	primitives["byte"] = dataType{"string", "byte", true}
	primitives["interface {}"] = dataType{"object", "object", true}

        return &doc
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

func isPrimitive(varType string) bool {
	_, found := primitives[varType]
	return found
}

func primitiveFormat(varType string) (string, string) {
	item, found := primitives[varType]
	if found {
		return item.swtype, item.swformat
	} else {
		return "object", ""
	}
}
