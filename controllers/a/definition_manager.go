/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package a

import (
	"context"

	"github.com/operatify/operatify/api/v1alpha1"
	"github.com/operatify/operatify/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type definitionManager struct{}

func (dm *definitionManager) GetDefinition(ctx context.Context, namespacedName types.NamespacedName) *reconciler.ResourceDefinition {
	return &reconciler.ResourceDefinition{
		InitialInstance: &v1alpha1.A{},
		StatusAccessor:  GetStatus,
		StatusUpdater:   updateStatus,
	}
}

func (dm *definitionManager) GetDependencies(ctx context.Context, thisInstance runtime.Object) (*reconciler.DependencyDefinitions, error) {
	return &reconciler.NoDependencies, nil
}
