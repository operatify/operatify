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
	"github.com/szoio/resource-operator-factory/api/v1alpha1"
	"github.com/szoio/resource-operator-factory/controllers/shared"
	"github.com/szoio/resource-operator-factory/reconciler"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

type ControllerFactory struct {
	ClientCreator func(logr.Logger, record.EventRecorder, *shared.Store) ResourceManager
	Scheme        *runtime.Scheme
	Store    	  *shared.Store
}

// +kubebuilder:rbac:groups=test.stephenzoio.com,resources=as,verbs=get;list;watch;Create;update;patch;delete
// +kubebuilder:rbac:groups=test.stephenzoio.com,resources=as/status,verbs=get;update;patch

const ResourceKind = "A"
const FinalizerName = "a.finalizers.com"

func (factory *ControllerFactory) SetupWithManager(mgr ctrl.Manager, parameters reconciler.ReconcileParameters, log *logr.Logger) error {
	if log == nil {
		l := ctrl.Log.WithName("controllers")
		log = &l
	}
	gc, err := factory.create(mgr.GetClient(),
		(*log).WithName(ResourceKind),
		mgr.GetEventRecorderFor(ResourceKind+"-controller"), parameters)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.A{}).
		Complete(gc)
}

func (factory *ControllerFactory) create(kubeClient client.Client, logger logr.Logger, recorder record.EventRecorder, parameters reconciler.ReconcileParameters) (*reconciler.GenericController, error) {
	resourceManagerClient := factory.ClientCreator(logger, recorder, factory.Store)

	return reconciler.CreateGenericController(parameters, ResourceKind, kubeClient, logger, recorder, factory.Scheme, &resourceManagerClient, &definitionManager{}, FinalizerName, shared.AnnotationBaseName, nil)
}

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
