package controllers

import (
	"context"
	"github.com/szoio/resource-operator-factory/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func deleteObject(key types.NamespacedName) error {
	f, err := getObject(key)
	if err != nil {
		return err
	}
	return k8sClient.Delete(context.Background(), f)
}

func getObject(key types.NamespacedName) (*v1alpha1.A, error) {
	f := &v1alpha1.A{}
	err := k8sClient.Get(context.Background(), key, f)
	return f, err
}

func nameAndSpec(aId string) (types.NamespacedName, *v1alpha1.A) {
	key := types.NamespacedName{
		Name:      aId,
		Namespace: "default",
	}
	spec := &v1alpha1.A{
		ObjectMeta: v1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: v1alpha1.ASpec{Id: aId},
	}

	return key, spec
}
