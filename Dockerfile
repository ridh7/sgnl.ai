# Copyright 2023 SGNL.ai, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG GOLANG_IMAGE=golang:1.21.3-bookworm
ARG BASE_IMAGE=gcr.io/distroless/static

FROM ${GOLANG_IMAGE} as build

RUN apt-get update && apt-get install -y \
    unzip=6.0*

ARG PROTOBUF_VERSION=23.3
RUN curl -fSsL "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip" > /tmp/protoc.zip \
    && (cd /usr/local && unzip /tmp/protoc.zip 'bin/protoc' 'include/*')  \
    && chmod +x /usr/local/bin/protoc \
    && rm -f /tmp/protoc.zip

ARG PROTOC_GEN_GO_VERSION=1.28.1
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v${PROTOC_GEN_GO_VERSION}

ARG PROTOC_GEN_GO_RPC_VERSION=1.3.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v${PROTOC_GEN_GO_RPC_VERSION}

ARG GOPS_VERSION=v0.3.27
RUN CGO_ENABLED=0 go install -ldflags "-s -w" github.com/google/gops@${GOPS_VERSION}

ARG AUTH_TOKENS_PATH=$AUTH_TOKENS_PATH 

WORKDIR /build
COPY . ./

RUN CGO_ENABLED=0 go build -ldflags "-s -w" ./cmd/adapter

FROM ${BASE_IMAGE} AS run
USER nonroot:nonroot

WORKDIR /

COPY --from=build /go/bin/gops /gops
COPY --from=build /build/adapter /adapter
COPY --from=build /build/$AUTH_TOKENS_PATH /$AUTH_TOKENS_PATH

ENTRYPOINT [ "/adapter" ]