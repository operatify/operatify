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
	api "github.com/szoio/operatify/api/v1alpha1"
	"github.com/szoio/operatify/controllers/manager"
	"github.com/szoio/operatify/controllers/shared"
	"github.com/szoio/operatify/reconciler"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

type ControllerFactory struct {
	ResourceManagerCreator func(logr.Logger, record.EventRecorder, *manager.Manager) shared.ResourceManager
	Scheme                 *runtime.Scheme
	Manager                *manager.Manager
}

// +kubebuilder:rbac:groups=test.stephenzoio.com,resources=bs,verbs=get;list;watch;Create;update;patch;delete
// +kubebuilder:rbac:groups=test.stephenzoio.com,resources=bs/status,verbs=get;update;patch

const ResourceKind = "B"
const FinalizerName = "b.finalizers.com"

func (factory *ControllerFactory) SetupWithManager(mgr ctrl.Manager, parameters reconciler.ReconcileParameters, log *logr.Logger) error {
	if log == nil {
		l := ctrl.Log.WithName("controllers")
		log = &l
	}
	gc, err := factory.createGenericController(mgr.GetClient(),
		(*log).WithName(ResourceKind),
		mgr.GetEventRecorderFor(ResourceKind+"-controller"), parameters)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&api.B{}).
		Complete(gc)
}

func (factory *ControllerFactory) createGenericController(kubeClient client.Client, logger logr.Logger, recorder record.EventRecorder, parameters reconciler.ReconcileParameters) (*reconciler.GenericController, error) {
	resourceManagerClient := factory.ResourceManagerCreator(logger, recorder, factory.Manager)

	return reconciler.CreateGenericController(parameters, ResourceKind, kubeClient, logger, recorder, factory.Scheme, &resourceManagerClient, &definitionManager{}, FinalizerName, shared.AnnotationBaseName, nil)
}

func CreateResourceManager(logger logr.Logger, recorder record.EventRecorder, manager *manager.Manager) shared.ResourceManager {
	return shared.ResourceManager{
		Logger:     logger,
		Recorder:   recorder,
		Manager:    manager,
		SpecGetter: getSpec,
	}
}
