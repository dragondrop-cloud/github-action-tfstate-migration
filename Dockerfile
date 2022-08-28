# This dockerfile performs a multi-stage build.
# Stage 1) Creates a reference to the desired terraform version
# Stage 2) Builds the tfmigrate executable.
# Stage 3) Builds the job that wraps these previous two steps and executes the code
# in this repository.
###################################################################################################
# 1) Reference to terraform binary
###################################################################################################
ARG TERRAFORM_VERSION=latest
FROM hashicorp/terraform:${TERRAFORM_VERSION} as tf

###################################################################################################
# 2) Building the tfmigrate binary
###################################################################################################
FROM golang:1.19-alpine3.15 as tfm
RUN apk update && apk add --no-cache bash git make

# Building tfmigrate executable
COPY --from=tf /bin/terraform /usr/local/bin/
RUN git clone https://github.com/minamijoyo/tfmigrate /tfmigrate
WORKDIR /tfmigrate

RUN go mod download
RUN make install
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
#    go build -ldflags='-w -s -extldflags "-static"' -a \
#    -o /go/bin/tfmigrate -v .

###################################################################################################
# 3) Building the github action logic
###################################################################################################
FROM golang:1.19-alpine3.15
RUN apk update && apk add --no-cache bash git make

# Copying compiled executables from tf-requirements
COPY --from=tf /bin/terraform /usr/local/bin/
COPY --from=tfm go/bin/tfmigrate /usr/local/bin/

# Building the src code
WORKDIR $GOPATH/github-action-ftstate-migration/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install

ENTRYPOINT ["/go/bin/github-action-tfstate-migration"]
