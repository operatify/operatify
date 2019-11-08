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
	"github.com/go-logr/logr"
	"github.com/szoio/resource-operator-factory/controllers/manager"
	"github.com/szoio/resource-operator-factory/reconciler"
	"k8s.io/client-go/tools/record"
)

type ResourceManager struct {
	Logger   logr.Logger
	Recorder record.EventRecorder
	Manager  *manager.Manager
}

func CreateResourceManager(logger logr.Logger, recorder record.EventRecorder, manager *manager.Manager) ResourceManager {
	return ResourceManager{
		Logger:   logger,
		Recorder: recorder,
		Manager:  manager,
	}
}

func (r *ResourceManager) Create(ctx context.Context, s reconciler.ResourceSpec) (reconciler.ApplyResponse, error) {
	instance, err := convertInstance(s.Instance)
	if err != nil {
		return reconciler.ApplyError, err
	}

	id := instance.Spec.Id
	result, err := r.Manager.Create(id)
	return reconciler.ApplyResponse{
		Result: result,
		Status: nil,
	}, err
}

func (r *ResourceManager) Update(ctx context.Context, s reconciler.ResourceSpec) (reconciler.ApplyResponse, error) {
	instance, err := convertInstance(s.Instance)
	if err != nil {
		return reconciler.ApplyError, err
	}

	id := instance.Spec.Id
	result, err := r.Manager.Update(id)
	return reconciler.ApplyResponse{
		Result: result,
		Status: &instance.Spec,
	}, err
}

func (r *ResourceManager) Verify(ctx context.Context, s reconciler.ResourceSpec) (reconciler.VerifyResponse, error) {
	instance, err := convertInstance(s.Instance)
	if err != nil {
		return reconciler.VerifyError, err
	}

	id := instance.Spec.Id
	result, err := r.Manager.Get(id)
	return reconciler.VerifyResponse{
		Result: result,
		Status: &instance.Spec,
	}, err
}

func (r *ResourceManager) Delete(ctx context.Context, s reconciler.ResourceSpec) (reconciler.DeleteResult, error) {
	instance, err := convertInstance(s.Instance)
	if err != nil {
		return reconciler.DeleteError, err
	}

	id := instance.Spec.Id
	return r.Manager.Delete(id)
}
