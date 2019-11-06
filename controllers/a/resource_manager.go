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
	"fmt"
	"github.com/go-logr/logr"
	"github.com/szoio/resource-operator-factory/controllers/shared"
	"github.com/szoio/resource-operator-factory/reconciler"
	"k8s.io/client-go/tools/record"
)

type ResourceManager struct {
	Logger   logr.Logger
	Recorder record.EventRecorder
	Store    *shared.Store
}

func CreateResourceManager(logger logr.Logger, recorder record.EventRecorder) ResourceManager {
	return ResourceManager{
		Logger:   logger,
		Recorder: recorder,
		Store:    shared.CreateStore(),
	}
}

func (r *ResourceManager) Create(ctx context.Context, s reconciler.ResourceSpec) (reconciler.ApplyResponse, error) {
	instance, err := convertInstance(s.Instance)
	if err != nil {
		return reconciler.ApplyError, err
	}

	id := instance.Spec.Data
	r.Store.Create(id)

	return reconciler.ApplyAwaitingVerification, err
}

func (_ *ResourceManager) Update(ctx context.Context, r reconciler.ResourceSpec) (reconciler.ApplyResponse, error) {
	return reconciler.ApplyError, fmt.Errorf("updating not currently supported")
}

func (r *ResourceManager) Verify(ctx context.Context, s reconciler.ResourceSpec) (reconciler.VerifyResponse, error) {
	instance, err := convertInstance(s.Instance)
	if err != nil {
		return reconciler.VerifyError, err
	}

	id := instance.Spec.Data
	result := r.Store.Get(id)
	return reconciler.VerifyResponse{
		Result: result,
		Status: nil,
	}, nil
}

func (r *ResourceManager) Delete(ctx context.Context, s reconciler.ResourceSpec) (reconciler.DeleteResult, error) {
	_, err := convertInstance(s.Instance)
	if err != nil {
		return reconciler.DeleteError, err
	}
	return reconciler.DeleteSucceeded, nil
}
