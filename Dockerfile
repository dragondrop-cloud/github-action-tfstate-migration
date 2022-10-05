# This dockerfile performs a multi-stage build.
# Stage 1) Builds the tfswitch executable.
# Stage 2) Builds the tfmigrate executable.
# Stage 3) Builds the job that wraps these previous two steps and executes the code
# in this repository.
###################################################################################################
# 1) Reference to tfswitch binary
###################################################################################################
FROM golang:1.19-alpine3.15 as tfswitch
RUN apk update && apk add --no-cache bash curl git make
RUN curl -L https://raw.githubusercontent.com/warrensbox/terraform-switcher/release/install.sh | bash

###################################################################################################
# 2) Building the tfmigrate binary
###################################################################################################
FROM golang:1.19-alpine3.15 as tfmigrate
RUN apk update && apk add --no-cache bash git make

# Building tfmigrate executable
RUN git clone https://github.com/minamijoyo/tfmigrate /tfmigrate
WORKDIR /tfmigrate

RUN go mod download
RUN make install

###################################################################################################
# 3) Building the go binary
###################################################################################################
FROM golang:1.19-alpine3.15 as tfstate-migration
RUN apk update && apk add --no-cache bash git make

# Building the src code
WORKDIR $GOPATH/github-action-ftstate-migration/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
     go build -ldflags='-w -s -extldflags "-static"' -a \
     -o /go/bin/github-action-tfstate-migration .

###################################################################################################
# 4) Final lightweight container
###################################################################################################
FROM golang:1.19-alpine3.15
RUN apk update && apk add --no-cache bash git make

# Copying compiled executables from upstream builds
COPY --from=tfswitch /usr/local/bin/tfswitch /usr/local/bin/
COPY --from=tfmigrate go/bin/tfmigrate /usr/local/bin/
COPY --from=tfstate-migration /go/bin/github-action-tfstate-migration /go/bin/github-action-tfstate-migration

ENTRYPOINT ["/go/bin/github-action-tfstate-migration"]
