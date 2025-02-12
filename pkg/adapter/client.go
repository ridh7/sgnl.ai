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

package adapter

import (
	"context"

	framework "github.com/sgnl-ai/adapter-framework"
)

// Client is a client that allows querying the datasource which
// contains JSON objects.
type Client interface {
	// GetPage returns a page of JSON objects from the datasource for the
	// requested entity.
	// Returns a (possibly empty) list of JSON objects, each object being
	// unmarshaled into a map by Golang's JSON unmarshaler.
	GetPage(ctx context.Context, request *Request) (*Response, *framework.Error)
}

// SCAFFOLDING #5 - pkg/adapter/client.go: Add/Remove/Update any fields to model the request for the SoR API.

// Request is a request to the datasource.
type Request struct {
	// BaseURL is the Base URL of the datasource to query.
	BaseURL string

	// Username is the username to use to authenticate with the datasource.
	Username string

	// Password is the password to use to authenticate with the datasource.
	Password string

	// Token is the Authorization token to use to authentication with the datasource.
	Token string

	// PageSize is the maximum number of objects to return from the entity.
	PageSize int64

	// EntityExternalID is the external ID of the entity.
	// The external ID should match the API's resource name.
	EntityExternalID string

	// Cursor identifies the first object of the page to return, as returned by
	// the last request for the entity.
	// Optional. If not set, return the first page for this entity.
	Cursor string
}

// SCAFFOLDING #6 - pkg/adapter/client.go: Add/Remove/Update any fields to model the response from the SoR API.
// Response is a response returned by the datasource.
type Response struct {
	// StatusCode is an HTTP status code.
	StatusCode int

	// RetryAfterHeader is the Retry-After response HTTP header, if set.
	RetryAfterHeader string

	// Objects is the list of
	// May be empty.
	Objects []map[string]any

	// NextCursor is the cursor that identifies the first object of the next
	// page.
	// May be empty.
	NextCursor string
}
