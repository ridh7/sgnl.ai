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
	"fmt"
	"strings"
	"time"

	framework "github.com/sgnl-ai/adapter-framework"
	api_adapter_v1 "github.com/sgnl-ai/adapter-framework/api/adapter/v1"
	"github.com/sgnl-ai/adapter-framework/web"
)

// Adapter implements the framework.Adapter interface to query pages of objects
// from datasources.
type Adapter struct {
	// SCAFFOLDING #20 - pkg/adapter/adapter.go: Add or remove fields to configure the adapter.

	// Client provides access to the datasource.
	Client Client
}

// NewAdapter instantiates a new Adapter.
//
// SCAFFOLDING #21 - pkg/adapter/adapter.go: Add or remove parameters to match field updates above.
func NewAdapter(client Client) framework.Adapter[Config] {
	return &Adapter{
		Client: client,
	}
}

// GetPage is called by SGNL's ingestion service to query a page of objects
// from a datasource.
func (a *Adapter) GetPage(ctx context.Context, request *framework.Request[Config]) framework.Response {
	if err := a.ValidateGetPageRequest(ctx, request); err != nil {
		return framework.NewGetPageResponseError(err)
	}

	return a.RequestPageFromDatasource(ctx, request)
}

// RequestPageFromDatasource requests a page of objects from a datasource.
func (a *Adapter) RequestPageFromDatasource(
	ctx context.Context, request *framework.Request[Config],
) framework.Response {

	// SCAFFOLDING #22 - pkg/adapter/adapter.go: Modify implementation to query your SoR.
	// If necessary, update this entire method to query your SoR. All of the code in this function
	// can be updated to match your SoR requirements.

	if !strings.HasPrefix(request.Address, "https://") {
		request.Address = "https://" + request.Address
	}
	req := &Request{
		BaseURL: request.Address,

		// Basic Auth
		Username: request.Auth.Basic.Username,
		Password: request.Auth.Basic.Password,

		// API Key or OAuth2 Token
		Token:            request.Auth.HTTPAuthorization,
		
		PageSize:         request.PageSize,
		EntityExternalID: request.Entity.ExternalId,
		Cursor:           request.Cursor,
	}

	resp, err := a.Client.GetPage(ctx, req)
	if err != nil {
		return framework.NewGetPageResponseError(err)
	}

	// An adapter error message is generated if the response status code is not
	// successful (i.e. if not statusCode >= 200 && statusCode < 300).
	if adapterErr := web.HTTPError(resp.StatusCode, resp.RetryAfterHeader); adapterErr != nil {
		return framework.NewGetPageResponseError(adapterErr)
	}

	// The raw JSON objects from the response must be parsed and converted into framework.Objects.
	// Nested attributes are flattened and delimited by the delimiter specified.
	// DateTime values are parsed using the specified DateTimeFormatWithTimeZone.
	parsedObjects, parserErr := web.ConvertJSONObjectList(
		&request.Entity,
		resp.Objects,

		// SCAFFOLDING #23 - pkg/adapter/adapter.go: Disable JSONPathAttributeNames.
		// Disable JSONPathAttributeNames if your datasource does not support
		// JSONPath attribute names. This should be enabled for most datasources.
		web.WithJSONPathAttributeNames(),

		// SCAFFOLDING #24 - pkg/adapter/adapter.go: List datetime formats supported by your SoR.
		// Provide a list of datetime formats supported by your datasource if
		// they are known. This will optimize the parsing of datetime values.
		// If this is not known, you can omit this option which will try
		// a list of common datetime formats.
		web.WithDateTimeFormats(
			[]web.DateTimeFormatWithTimeZone{
				{Format: time.RFC3339, HasTimeZone: true},
				{Format: time.RFC3339Nano, HasTimeZone: true},
				{Format: "2006-01-02T15:04:05.000Z0700", HasTimeZone: true},
				{Format: "2006-01-02", HasTimeZone: false},
			}...,
		),

		// SCAFFOLDING #25 - pkg/adapter/adapter.go: Uncomment to set the default timezone in case the SoR datetime attribute does not have timezone specified.
		// This can be provided to be used as a default value when parsing
		// datetime values lacking timezone info. This defaults to UTC.
		// web.WithLocalTimeZoneOffset(-7),
	)
	if parserErr != nil {
		return framework.NewGetPageResponseError(
			&framework.Error{
				Message: fmt.Sprintf("Failed to convert datasource response objects: %v.", parserErr),
				Code:    api_adapter_v1.ErrorCode_ERROR_CODE_INTERNAL,
			},
		)
	}

	page := &framework.Page{
		Objects: parsedObjects,
	}

	page.NextCursor = resp.NextCursor

	return framework.NewGetPageResponseSuccess(page)
}
