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

package b

import (
	"fmt"

	api "github.com/szoio/operatify/api/v1alpha1"
	"github.com/szoio/operatify/reconciler"

	"k8s.io/apimachinery/pkg/runtime"
)

func GetStatus(instance runtime.Object) (*reconciler.Status, error) {
	x, err := convertInstance(instance)
	if err != nil {
		return nil, err
	}
	status := x.Status

	return &reconciler.Status{
		State:   reconciler.ReconcileState(status.State),
		Message: status.Message,
	}, nil
}

func updateStatus(instance runtime.Object, status *reconciler.Status) error {
	x, err := convertInstance(instance)
	if err != nil {
		return err
	}
	x.Status.State = string(status.State)
	x.Status.Message = status.Message
	return nil
}

func convertInstance(obj runtime.Object) (*api.B, error) {
	local, ok := obj.(*api.B)
	if !ok {
		return nil, fmt.Errorf("failed type assertion on kind: A")
	}
	return local, nil
}

func getSpec(object runtime.Object) (*api.Spec, error) {
	instance, err := convertInstance(object)
	if err != nil {
		return nil, err
	}
	return &instance.Spec.Spec, nil
}
