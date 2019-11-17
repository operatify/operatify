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

package shared

import (
	"context"

	"github.com/szoio/operatify/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/go-logr/logr"
	"github.com/szoio/operatify/controllers/manager"
	"github.com/szoio/operatify/reconciler"
	"k8s.io/client-go/tools/record"
)

type SpecGetter func(object runtime.Object) (*v1alpha1.Spec, error)

type ResourceManager struct {
	Logger     logr.Logger
	Recorder   record.EventRecorder
	Manager    *manager.Manager
	SpecGetter SpecGetter
}

func CreateResourceManager(logger logr.Logger, recorder record.EventRecorder, manager *manager.Manager, specGetter SpecGetter) ResourceManager {
	return ResourceManager{
		Logger:     logger,
		Recorder:   recorder,
		Manager:    manager,
		SpecGetter: specGetter,
	}
}

func (r *ResourceManager) Create(ctx context.Context, s reconciler.ResourceSpec) (reconciler.ApplyResponse, error) {
	spec, err := r.SpecGetter(s.Instance)
	if err != nil {
		return reconciler.ApplyError, err
	}

	result, err := r.Manager.Create(spec.Id)
	return reconciler.ApplyResponse{
		Result: result,
		Status: &spec,
	}, err
}

func (r *ResourceManager) Update(ctx context.Context, s reconciler.ResourceSpec) (reconciler.ApplyResponse, error) {
	spec, err := r.SpecGetter(s.Instance)
	if err != nil {
		return reconciler.ApplyError, err
	}

	result, err := r.Manager.Update(spec.Id)
	return reconciler.ApplyResponse{
		Result: result,
		Status: &spec,
	}, err
}

func (r *ResourceManager) Verify(ctx context.Context, s reconciler.ResourceSpec) (reconciler.VerifyResponse, error) {
	spec, err := r.SpecGetter(s.Instance)
	if err != nil {
		return reconciler.VerifyError, err
	}

	result, err := r.Manager.Get(spec.Id)
	return reconciler.VerifyResponse{
		Result: result,
		Status: &spec,
	}, err
}

func (r *ResourceManager) Delete(ctx context.Context, s reconciler.ResourceSpec) (reconciler.DeleteResult, error) {
	spec, err := r.SpecGetter(s.Instance)
	if err != nil {
		return reconciler.DeleteError, err
	}

	return r.Manager.Delete(spec.Id)
}
