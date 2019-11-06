package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/log"
	"github.com/szoio/resource-operator-factory/api/v1alpha1"
	"github.com/szoio/resource-operator-factory/reconciler"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Test Controller", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Create and Delete", func() {
		It("should create and delete an A", func() {
			log.Info("Creating an A...")

			key := types.NamespacedName{
				Name:      "test-cluster",
				Namespace: "default",
			}

			id := "a-" + RandomString(10)

			created := &v1alpha1.A{
				ObjectMeta: v1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.ASpec{Id: id},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() bool {
				f := &v1alpha1.A{}
				k8sClient.Get(context.Background(), key, f)
				return f.Status.State == string(reconciler.Succeeded)
			}, timeout, interval).Should(BeTrue())

			record := store.GetRecord(id)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultProvisioning,
				reconciler.VerifyResultReady,
			}))

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &v1alpha1.A{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &v1alpha1.A{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

			record = store.GetRecord(id)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultProvisioning,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))

			log.Info("Finished creating an A...")
		})
	})
})
