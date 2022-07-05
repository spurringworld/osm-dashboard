// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
	"context"
	"log"

	osmconfigclientset "github.com/openservicemesh/osm/pkg/gen/client/config/clientset/versioned"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubernetes/dashboard/src/app/backend/errors"
)

// MeshNameValidityMetadata is a specification for Mesh name validation request.
type MeshNameValidityMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// MeshNameValidity describes validity of the Mesh name.
type MeshNameValidity struct {
	// True when the Mesh name is valid.
	Valid bool `json:"valid"`
}

// ValidateMeshName validates Mesh name. When error is returned, name validity could not be
// determined.
func ValidateMeshName(metadata *MeshNameValidityMetadata, osmConfigClient osmconfigclientset.Interface) (*MeshNameValidity, error) {
	log.Printf("Validating %s mesh config name in %s namespace", metadata.Name, metadata.Namespace)

	isValid := false
	_, err := osmConfigClient.ConfigV1alpha2().MeshConfigs(metadata.Namespace).Get(context.TODO(), metadata.Name+"-mesh-config", metaV1.GetOptions{})

	println(errors.IsNotFoundError(err))
	if err != nil {
		if errors.IsNotFoundError(err) || errors.IsForbiddenError(err) {
			isValid = true
		} else {
			return nil, err
		}
	}

	log.Printf("Validation result for %s mesh config name in %s namespace is %t", metadata.Name,
		metadata.Namespace, isValid)

	return &MeshNameValidity{Valid: isValid}, nil
}
