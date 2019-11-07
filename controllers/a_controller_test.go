package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/log"
	api "github.com/szoio/resource-operator-factory/api/v1alpha1"
	"github.com/szoio/resource-operator-factory/controllers/manager"
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
		It("should create and delete an A with async operations", func() {
			log.Info("Creating an A asynchronously...")

			aId := "a-" + RandomString(10)
			key := types.NamespacedName{
				Name:      "test-cluster",
				Namespace: "default",
			}

			created := &api.A{
				ObjectMeta: v1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: api.ASpec{Id: aId},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() bool {
				f := &api.A{}
				k8sClient.Get(context.Background(), key, f)
				return f.Status.State == string(reconciler.Succeeded)
			}, timeout, interval).Should(BeTrue())

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultProvisioning,
				reconciler.VerifyResultReady,
			}))

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &api.A{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &api.A{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultProvisioning,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))

			log.Info("Finished creating an A...")
		})

		It("should create and delete an A with sync operations", func() {
			log.Info("Creating an A synchronously...")

			key := types.NamespacedName{
				Name:      "test-cluster",
				Namespace: "default",
			}

			aId := "a-" + RandomString(10)

			// tell it to run the create synchronously
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventCreate,
				Operation: manager.CreateSync.AsOperation(),
			})

			created := &api.A{
				ObjectMeta: v1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: api.ASpec{Id: aId},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() bool {
				f := &api.A{}
				k8sClient.Get(context.Background(), key, f)
				return f.Status.State == string(reconciler.Succeeded)
			}, timeout, interval).Should(BeTrue())

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultReady,
			}))

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				f := &api.A{}
				k8sClient.Get(context.Background(), key, f)
				return k8sClient.Delete(context.Background(), f)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				f := &api.A{}
				return k8sClient.Get(context.Background(), key, f)
			}, timeout, interval).ShouldNot(Succeed())

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultReady,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))

			log.Info("Finished creating an A...")
		})
	})
})
