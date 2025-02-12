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
	"errors"
)

// Config is the optional configuration passed in each GetPage calls to the
// adapter.
type Config struct {
	// SCAFFOLDING #3 - pkg/adapter/config.go - pass Adapter config fields.
	// Every field MUST have a `json` tag.

	// Example config field.
	APIVersion string `json:"apiVersion,omitempty"`
}

// ValidateConfig validates that a Config received in a GetPage call is valid.
func (c *Config) Validate(_ context.Context) error {
	// SCAFFOLDING #4 - pkg/adapter/config.go: Validate fields passed in Adapter config.
	// Update the checks below to validate the fields in Config.
	switch {
	case c == nil:
		return errors.New("request contains no config")
	case c.APIVersion == "":
		return errors.New("apiVersion is not set")
	default:
		return nil
	}
}
