FROM gobuffalo/buffalo:v0.12.0

ADD . $GOPATH/src/github.com/gobuffalo/buffalo-auth
WORKDIR $GOPATH/src/github.com/gobuffalo/buffalo-auth
RUN go install -v

WORKDIR $GOPATH/src

RUN buffalo new  --db-type=sqlite3 --skip-webpack app 
WORKDIR ./app

RUN buffalo g auth
RUN buffalo db migrate
RUN buffalo test -v ./...