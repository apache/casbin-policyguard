// Copyright 2026 The Casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casbin

import (
	"testing"
)

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantOk   bool
	}{
		{
			name:     "pod-security template exists",
			template: "pod-security",
			wantOk:   true,
		},
		{
			name:     "image-validation template exists",
			template: "image-validation",
			wantOk:   true,
		},
		{
			name:     "resource-quota template exists",
			template: "resource-quota",
			wantOk:   true,
		},
		{
			name:     "namespace-isolation template exists",
			template: "namespace-isolation",
			wantOk:   true,
		},
		{
			name:     "non-existent template",
			template: "non-existent",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, ok := GetTemplate(tt.template)
			if ok != tt.wantOk {
				t.Errorf("GetTemplate() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && tmpl.Name != tt.template {
				t.Errorf("GetTemplate() name = %v, want %v", tmpl.Name, tt.template)
			}
		})
	}
}

func TestTemplateModel(t *testing.T) {
	for name, tmpl := range Templates {
		t.Run(name, func(t *testing.T) {
			if tmpl.Model == "" {
				t.Errorf("template %s has empty model", name)
			}
			if tmpl.Description == "" {
				t.Errorf("template %s has empty description", name)
			}
		})
	}
}
