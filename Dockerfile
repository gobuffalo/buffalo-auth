FROM gobuffalo/buffalo:v0.12.0

RUN rm -rf $GOPATH/src/github.com/gobuffalo/buffalo-auth
ADD . $GOPATH/src/github.com/gobuffalo/buffalo-auth

WORKDIR $GOPATH/src/github.com/gobuffalo/buffalo-auth
RUN go install -v

WORKDIR $GOPATH/src

RUN buffalo new  --db-type=sqlite3 --skip-webpack app 
WORKDIR ./app

RUN buffalo g auth 
RUN buffalo db migrate
RUN buffalo test -v ./...

WORKDIR $GOPATH/src
RUN buffalo new --db-type=sqlite3 --skip-webpack other
WORKDIR ./other
RUN buffalo g auth first_name last_name phone_number
RUN buffalo db migrate
RUN buffalo test -v ./...
RUN filetest -c $GOPATH/src/github.com/gobuffalo/buffalo-auth/filetests/aditional_fields.json
