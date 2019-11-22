package controllers

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/operatify/operatify/api/v1alpha1"
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

func getObjectA(key types.NamespacedName) (*v1alpha1.A, error) {
	f := &v1alpha1.A{}
	err := k8sClient.Get(context.Background(), key, f)
	return f, err
}

func nameAndSpecA(aId string) (types.NamespacedName, *v1alpha1.A) {
	return nameAndSpecWithAnnotationsA(aId, nil)
}

func nameAndSpecWithAnnotationsA(aId string, annotations map[string]string) (types.NamespacedName, *v1alpha1.A) {
	key := types.NamespacedName{
		Name:      aId,
		Namespace: "default",
	}
	spec := &v1alpha1.A{
		ObjectMeta: v1.ObjectMeta{
			Name:        key.Name,
			Namespace:   key.Namespace,
			Annotations: annotations,
		},
		Spec: v1alpha1.ASpec{
			Spec: v1alpha1.Spec{
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
