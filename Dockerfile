FROM  golang:1.19.5-bullseye AS build-env

ARG TARGETPLATFORM
ARG BUILDPLATFORM 
RUN echo "GOARCH=$(echo $TARGETPLATFORM | cut -d / -f 2)"

WORKDIR /go/src/github.com/evmos/evmos

RUN apt-get update -y
RUN apt-get install git -y

COPY . .

RUN make build 

FROM golang:1.19.5-bullseye

RUN apt-get update -y
RUN apt-get install ca-certificates jq -y

WORKDIR /root

COPY --from=build-env /go/src/github.com/evmos/evmos/build/hustd /usr/bin/hustd

ENTRYPOINT [ "hustd" ]
