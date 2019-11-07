package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/log"
	"github.com/szoio/resource-operator-factory/controllers/manager"
	"github.com/szoio/resource-operator-factory/reconciler"
)

var _ = FDescribe("Test Reconciler", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Create and Delete", func() {
		It("should create and delete with async operations", func() {
			log.Info("Creating an A asynchronously...")

			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() bool {
				f, _ := getObject(key)
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
				return deleteObject(key)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				_, err := getObject(key)
				return err
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

		It("should create and delete sync operations", func() {
			log.Info("Creating an A synchronously...")

			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// tell it to run the create synchronously
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventCreate,
				Operation: manager.CreateSync.AsOperation(),
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() bool {
				f, _ := getObject(key)
				return f.Status.State == string(reconciler.Succeeded)
			}, timeout, interval).Should(BeTrue())

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultReady,
			}))

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				return deleteObject(key)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				_, err := getObject(key)
				return err
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

	Context("Handle failure", func() {
		It("should fail if fails to create", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// tell it to fail to create
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventCreate,
				Operation: manager.CreateFail.AsOperation(),
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() bool {
				f, _ := getObject(key)
				return f.Status.State == string(reconciler.Failed)
			}, timeout, interval).Should(BeTrue())

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultError,
			}))

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				_, err := getObject(key)
				return err
			}, timeout, interval).ShouldNot(Succeed())
		})

		It("should fail if fails complete async creation", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// tell it to fail to create
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventCreate,
				Operation: manager.CreateCompleteFail.AsOperation(),
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() bool {
				f, _ := getObject(key)
				return f.Status.State == string(reconciler.Failed)
			}, timeout, interval).Should(BeTrue())

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultProvisioning,
				reconciler.VerifyResultError,
			}))

			// Remove the behaviours
			resourceManager.ClearBehaviours(aId)

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				_, err := getObject(key)
				return err
			}, timeout, interval).ShouldNot(Succeed())

			// TODO: should it this?
			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultProvisioning,
				reconciler.VerifyResultError,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))
		})

		It("should fail if fails to verify creation after completing async creation", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// tell it to fail verification
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventGet,
				Operation: manager.VerifyFail.AsOperation(),
				Count:     1,
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			Eventually(func() bool {
				f, _ := getObject(key)
				return f.Status.State == string(reconciler.Failed)
			}, timeout, interval).Should(BeTrue())

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultProvisioning,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultError,
				reconciler.VerifyResultError, // TODO: should it this?
			}))

			// Remove the behaviours
			resourceManager.ClearBehaviours(aId)

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).Should(Succeed())

			By("Expecting to delete finish")
			Eventually(func() error {
				_, err := getObject(key)
				return err
			}, timeout, interval).ShouldNot(Succeed())

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultProvisioning,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultError,
				reconciler.VerifyResultError, // TODO: should it this?
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))
		})
	})
})
