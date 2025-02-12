// Copyright 2023 SGNL.ai, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	api_adapter_v1 "github.com/sgnl-ai/adapter-framework/api/adapter/v1"
	"github.com/sgnl-ai/adapter-framework/server"
	"github.com/sgnl-ai/adapter-template/pkg/adapter"
	"google.golang.org/grpc"
)

var (
	// Port is the port at which the gRPC server will listen.
	Port = flag.Int("port", 8080, "The server port")

	// Timeout is the timeout for the HTTP client used to make requests to the datasource (seconds).
	Timeout = flag.Int("timeout", 30, "The timeout for the HTTP client used to make requests to the datasource (seconds)")
)

func main() {
	logger := log.New(os.Stdout, "adapter", log.Lmicroseconds|log.LUTC|log.Lshortfile)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *Port))
	if err != nil {
		logger.Fatalf("Failed to open server port: %v", err)
	}

	// SCAFFOLDING #1 - cmd/adapter/main.go: Pass options to configure TLS, connection parameters.
	s := grpc.NewServer()

	stop := make(chan struct{})

	adapterServer := server.New(stop)

	// SCAFFOLDING #2 - cmd/adapter/main.go: Update Adapter type.
	// The Adapter type below must be unique across all registered Adapters and match the Adapter
	// type configured on the Adapter object via the SGNL Config API.
	//
	// If you need to run multiple adapters on the same gRPC server, they can be registered here.
	err = server.RegisterAdapter(adapterServer, "Test-1.0.0", adapter.NewAdapter(adapter.NewClient(*Timeout)))
	if err != nil {
		logger.Fatalf("Failed to register adapter: %v", err)
	}

	api_adapter_v1.RegisterAdapterServer(s, adapterServer)

	logger.Printf("Started adapter gRPC server on port %d", *Port)

	if err := s.Serve(listener); err != nil {
		logger.Fatalf("Failed to listen on server port: %v", err)
	}
}
