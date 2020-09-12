package controllers

import (
	"context"

	. "github.com/onsi/gomega"
	apiv1 "github.com/operatify/operatify/api/v1alpha1"
	"github.com/operatify/operatify/reconciler"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func deleteObjectA(key types.NamespacedName) error {
	f, err := getObjectA(key)
	if err != nil {
		return err
	}
	return k8sClient.Delete(context.Background(), f)
}

func getObjectA(key types.NamespacedName) (*apiv1.ATest, error) {
	f := &apiv1.ATest{}
	err := k8sClient.Get(context.Background(), key, f)
	return f, err
}

func nameAndSpecA(aId string) (types.NamespacedName, *apiv1.ATest) {
	return nameAndSpecWithAnnotationsA(aId, nil)
}

func nameAndSpecWithAnnotationsA(aId string, annotations map[string]string) (types.NamespacedName, *apiv1.ATest) {
	key := types.NamespacedName{
		Name:      aId,
		Namespace: "default",
	}
	spec := &apiv1.ATest{
		ObjectMeta: v1.ObjectMeta{
			Name:        key.Name,
			Namespace:   key.Namespace,
			Annotations: annotations,
		},
		Spec: apiv1.ASpec{
			Spec: apiv1.Spec{
				Id: aId,
			},
		},
	}

	return key, spec
}

func waitUntilReconcileStateA(key types.NamespacedName, state reconciler.ReconcileState) {
	Eventually(func() reconciler.ReconcileState {
		f, _ := getObjectA(key)
		return reconciler.ReconcileState(f.Status.State)
	}, timeout, interval).Should(Equal(state))
}

func waitUntilObjectMissingA(key types.NamespacedName) {
	Eventually(func() error {
		_, err := getObjectA(key)
		return err
	}, timeout, interval).ShouldNot(Succeed())
}
