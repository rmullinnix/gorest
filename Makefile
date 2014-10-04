DEP=github.com/rmullinnix/logger
TARG=github.com/rmullinnix/gorest

GOFILES=\
    doc.go\
	api.go\
	decorator.go\
	siren.go\
	gorest.go\
	mime.go\
	parse.go\
	reflect.go\
	marshaller.go\
	client.go\
	util.go\
	sec.go\

install:
	go install $(DEP)
	go install $(TARG)

build:
	go build $(TARG)

