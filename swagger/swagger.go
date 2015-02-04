package swagger

import (
	"github.com/rmullinnix/gorest"
	"strings"
)

// creates a new Swagger Documentor
//   versions supported - 1.2 and 2.0
func NewSwaggerDocumentor(version string) *gorest.Documentor {
	var doc		gorest.Documentor

	if version == "1.2" {
        	doc = gorest.Documentor{swaggerDocumentor12}
	} else if version == "2.0" {
		doc = gorest.Documentor{swaggerDocumentor20}
	}

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
