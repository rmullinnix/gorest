gorest/swagger
======

* Implementation of swagger 1.2 and swagger 2.0 spec generation based on gorest EndPointStruct and tags in StructField
* Provides basics for generating swagger specification from service to be viewed in swagger-ui
* Not yet implemented - security sections of specification, all attributes of schema, external docs, xml attribute
* Requires use of github.com/rmullinnix/gorest library
* Uses Go tags instead of annotations

### Connecting to gorest service
```
gorest.RegisterDocumentor("swagger", swagger.NewSwaggerDocumentor("1.2")) // 1.2 spec generation
   or
gorest.RegisterDocumentor("swagger", swagger.NewSwaggerDocumentor("2.0")) // 2.0 spec generation
```
"swagger" represents the endpoint which will generate the spec.  The gorest libary will intercept this before searching other endpoints

### gorest Service Specification with swagger tags
```
type SecurityService struct {
        gorest.RestService      `realm:"Security" root:"/contra/security" consumes:"application/json"
                                produces:"application/vnd.siren+json,application/json,application/hal+json"
                                swagger:"/swagger" sw.apiVersion:"1.0"
                                sw.contactUrl:"http://www.abc.com" sw.contactEmail:"rmullinnix@yahoo.com"`
        getRole                 gorest.EndPoint `method:"GET"   path:"/role/{RoleName:string}" output:"RoleAccess"
                                sw.summary:"Retrieve the access rights for the role"
                                sw.notes:"A role has access rights for a set of resources"
                                sw.nickname:"Get Role Access"
                                sw.tags:"Security,Role"
                                sw.response:"{200:OK:output},{404:Role not found},{500:Internal Server Error}"`
        getRoleList             gorest.EndPoint `method:"GET"   path:"/role" output:"[]RoleAccess"
                                sw.summary:"Retrieve the list of roles and associated access rights"
                                sw.notes:"Each role has access rights for a set of resources"
                                sw.nickname:"Get Role List"
                                sw.tags:"Security,Role"
                                sw.response:"{200:OK:output},{500:Internal Server Error}"`
        createRole              gorest.EndPoint `method:"POST"  path:"/role" postdata:"RoleAccess"
                                sw.summary:"Create a new role"
                                sw.notes:"A role must be unique and access must be for valid resources"
                                sw.nickname:"Create Role"
                                sw.tags:"Security,Role"
                                sw.response:"{201:Role created},{409:Role already exists},{500:Internal Server Error}"`
}
```

Tags on structures returned from Endpoints
```
type RoleAccess struct {
        Role            string                          `sw.description:"Name of the role" sw.required:"true"`
        ResourceACL     []ResourceAccess                `sw.description:"List of access rights to resource"`
}

type ResourceAccess struct{
        Resource        string                          `sw.description:"Name of the resource" sw.required:"true"`
        Href            string                          `sw.description:"Path to the resource" sw.required:"true"`
        Access          string                          `sw.description:"Access rights to resource (create, read, update or delete)" sw.required:"true"`
}
```
