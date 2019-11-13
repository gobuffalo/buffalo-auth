FROM gobuffalo/buffalo:latest
ENV GOPROXY=https://proxy.golang.org

RUN rm -rf $GOPATH/src/github.com/gobuffalo/buffalo-auth
ADD . $GOPATH/src/github.com/gobuffalo/buffalo-auth
RUN go get -u -v github.com/markbates/filetest

WORKDIR $GOPATH/src/github.com/gobuffalo/buffalo-auth
RUN go get -t -v ./...
RUN go install -v
RUN go test -race -tags sqlite -v ./...

WORKDIR $GOPATH/src
RUN buffalo new  --skip-webpack --db-type=sqlite3 app

WORKDIR $GOPATH/src/app
RUN buffalo plugins install github.com/gobuffalo/buffalo-auth

ENV ENV GO111MODULE=on
RUN buffalo g auth
RUN buffalo db migrate
RUN buffalo test -v ./...

WORKDIR $GOPATH/src

RUN buffalo new --db-type=sqlite3 --skip-webpack -f aditional_fields

WORKDIR $GOPATH/src/aditional_fields
RUN buffalo plugins install github.com/gobuffalo/buffalo-auth
ENV ENV GO111MODULE=on
RUN buffalo g auth first_name last_name phone_number email:string
RUN buffalo db migrate
RUN buffalo test -v ./...

RUN filetest -c $GOPATH/src/github.com/gobuffalo/buffalo-auth/filetests/aditional_fields.json
