package controllers

import (
	"context"

	. "github.com/onsi/gomega"
	apiv1 "github.com/operatify/operatify/api/v1alpha1"
	"github.com/operatify/operatify/reconciler"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func deleteObjectB(key types.NamespacedName) error {
	f, err := getObjectB(key)
	if err != nil {
		return err
	}
	return k8sClient.Delete(context.Background(), f)
}

func getObjectB(key types.NamespacedName) (*apiv1.BTest, error) {
	f := &apiv1.BTest{}
	err := k8sClient.Get(context.Background(), key, f)
	return f, err
}

func nameAndSpecB(aId string, owner string, dependencies []string) (types.NamespacedName, *apiv1.BTest) {
	return nameAndSpecWithAnnotationsB(aId, owner, dependencies, nil)
}

func nameAndSpecWithAnnotationsB(bId string, owner string, dependencies []string, annotations map[string]string) (types.NamespacedName, *apiv1.BTest) {
	key := types.NamespacedName{
		Name:      bId,
		Namespace: "default",
	}
	spec := &apiv1.BTest{
		ObjectMeta: v1.ObjectMeta{
			Name:        key.Name,
			Namespace:   key.Namespace,
			Annotations: annotations,
		},
		Spec: apiv1.BSpec{
			Spec: apiv1.Spec{
				Id: bId,
			},
			Owner:        owner,
			Dependencies: dependencies,
		},
	}

	return key, spec
}

func waitUntilReconcileStateB(key types.NamespacedName, state reconciler.ReconcileState) {
	Eventually(func() reconciler.ReconcileState {
		f, _ := getObjectB(key)
		return reconciler.ReconcileState(f.Status.State)
	}, timeout, interval).Should(Equal(state))
}

func waitUntilObjectMissingB(key types.NamespacedName) {
	Eventually(func() error {
		_, err := getObjectB(key)
		return err
	}, timeout, interval).ShouldNot(Succeed())
}
