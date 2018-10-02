FROM gobuffalo/buffalo:latest

RUN rm -rf $GOPATH/src/github.com/gobuffalo/buffalo-auth
ADD . $GOPATH/src/github.com/gobuffalo/buffalo-auth

WORKDIR $GOPATH/src/github.com/gobuffalo/buffalo-auth
RUN go install -v
RUN go test -race -tags sqlite -v ./...

WORKDIR $GOPATH/src

RUN buffalo new  --db-type=sqlite3 --skip-webpack app
WORKDIR $GOPATH/src/app

RUN buffalo g auth
RUN buffalo db migrate
RUN buffalo test -v ./...

WORKDIR $GOPATH/src
RUN buffalo new --db-type=sqlite3 --skip-webpack -f aditional_fields
WORKDIR $GOPATH/src/aditional_fields

RUN buffalo g auth first_name last_name phone_number email:string
RUN buffalo db migrate
RUN buffalo test -v ./...
RUN filetest -c $GOPATH/src/github.com/gobuffalo/buffalo-auth/filetests/aditional_fields.json
