package controllers

import (
	"context"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/szoio/resource-operator-factory/controllers/manager"
	"github.com/szoio/resource-operator-factory/reconciler"
)

var _ = Describe("Test Reconciler", func() {

	Context("Create and Delete", func() {
		It("should create asynchronously and delete asynchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
			}))

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).To(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))
		})

		It("should create synchronously and delete asynchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// tell it to run the create synchronously
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventCreate,
				Operation: manager.CreateSync.AsOperation(),
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultReady,
			}))

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).To(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultReady,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing,
			}))
		})

		It("should create asynchronously and delete synchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
			}))

			// tell it to run delete synchronously
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventDelete,
				Operation: manager.DeleteSync.AsOperation(),
			})

			// Delete
			By("Expecting to delete successfully")
			Eventually(func() error {
				return deleteObject(key)
			}, timeout, interval).Should(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)

			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultMissing,
			}))
		})
	})

	Context("Update/Recreate", func() {
		It("should update synchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).To(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			// tell it update is required (ony for the next verify)
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventGet,
				Operation: manager.VerifyNeedsUpdate.AsOperation(),
				From:      resourceManager.CountEvents(aId, manager.EventGet),
				Count:     1,
			})

			toUpdate, _ := getObject(key)
			toUpdate.Spec.IntData = 1
			toUpdate.Spec.StringData = "Updated"
			Expect(k8sClient.Update(context.Background(), toUpdate)).To(Succeed())

			Eventually(func() bool {
				record := resourceManager.GetRecord(aId)
				return reflect.DeepEqual(record.States, []reconciler.VerifyResult{
					reconciler.VerifyResultInProgress,
					reconciler.VerifyResultReady,
					reconciler.VerifyResultUpdateRequired,
					reconciler.VerifyResultReady,
				})
			}, timeout, interval).Should(BeTrue())

			Expect(resourceManager.CountEvents(aId, manager.EventCreate)).To(Equal(1))
			Expect(resourceManager.CountEvents(aId, manager.EventUpdate)).To(Equal(1))

			updated, _ := getObject(key)
			Expect(updated.Spec.StringData).To(Equal("Updated"))
		})

		It("should update asynchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpec(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).To(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			// tell it update is required (ony for the next verify)
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventGet,
				Operation: manager.VerifyNeedsUpdate.AsOperation(),
				OneTime:   true,
			})

			// tell to update asynchronously
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventUpdate,
				Operation: manager.UpdateAsync.AsOperation(),
			})

			toUpdate, _ := getObject(key)
			toUpdate.Spec.IntData = 1
			toUpdate.Spec.StringData = "Updated"
			Expect(k8sClient.Update(context.Background(), toUpdate)).To(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			Eventually(func() bool {
				record := resourceManager.GetRecord(aId)
				return reflect.DeepEqual(record.States, []reconciler.VerifyResult{
					reconciler.VerifyResultInProgress,
					reconciler.VerifyResultReady,
					reconciler.VerifyResultUpdateRequired,
					reconciler.VerifyResultInProgress,
					reconciler.VerifyResultReady,
				})
			}, timeout, interval).Should(BeTrue())

			Expect(resourceManager.CountEvents(aId, manager.EventCreate)).To(Equal(1))
			Expect(resourceManager.CountEvents(aId, manager.EventUpdate)).To(Equal(1))

			updated, _ := getObject(key)
			Expect(updated.Spec.StringData).To(Equal("Updated"))
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
			By("Expecting to create successfully")
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Failed)

			By("Expecting correct state transitions")
			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultError,
			}))

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).Should(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)
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
			waitUntilReconcileState(key, reconciler.Failed)

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultError,
			}))

			// Remove the behaviours
			resourceManager.ClearBehaviours(aId)

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).Should(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)

			// TODO: should it this?
			record = resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
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
				From:      1,
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Failed)

			// wait until Ready is received
			Eventually(func() bool {
				record := resourceManager.GetRecord(aId)
				for _, r := range record.States {
					if r == reconciler.VerifyResultReady {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
			var recordPreStates = resourceManager.GetRecord(aId).States

			// Remove the behaviours
			resourceManager.ClearBehaviours(aId)

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObject(key)).Should(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissing(key)

			record := resourceManager.GetRecord(aId)
			expectedStates := append(recordPreStates,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing)
			Expect(record.States).Should(Equal(expectedStates))
		})
	})

	Context("Handle permissions", func() {
		It("should create if create permission is present", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecWithAnnotations(aId, map[string]string{accessPermissionAnnotation: "C"})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.Events).To(ContainElement(manager.EventCreate))
		})

		It("should stop creating if create permission is missing", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecWithAnnotations(aId, map[string]string{accessPermissionAnnotation: "none"})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Failed)

			record := resourceManager.GetRecord(aId)
			Expect(record.Events).To(Not(ContainElement(manager.EventCreate)))

			object, _ := getObject(key)
			Expect(object.Status.Message).To(HavePrefix("permission to create Azure resource is not set"))
		})

		It("should update if update if permission is present", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecWithAnnotations(aId, map[string]string{accessPermissionAnnotation: "CU"})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).To(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			// tell it update is required (ony for the next verify)
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventGet,
				Operation: manager.VerifyNeedsUpdate.AsOperation(),
				From:      resourceManager.CountEvents(aId, manager.EventGet),
				Count:     1,
			})

			toUpdate, _ := getObject(key)
			toUpdate.Spec.StringData = "Updated"
			Expect(k8sClient.Update(context.Background(), toUpdate)).To(Succeed())

			Eventually(func() []reconciler.VerifyResult {
				return resourceManager.GetRecord(aId).States
			}, timeout, interval).Should(ContainElement(reconciler.VerifyResultUpdateRequired))
			waitUntilReconcileState(key, reconciler.Succeeded)

			updated, _ := getObject(key)
			Expect(updated.Spec.StringData).To(Equal("Updated"))
		})

		It("should fail to update if update if permission is absent", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecWithAnnotations(aId, map[string]string{accessPermissionAnnotation: "C"})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).To(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			// tell it update is required (ony for the next verify)
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventGet,
				Operation: manager.VerifyNeedsUpdate.AsOperation(),
				From:      resourceManager.CountEvents(aId, manager.EventGet),
				Count:     1,
			})

			toUpdate, _ := getObject(key)
			toUpdate.Spec.StringData = "Updated"
			Expect(k8sClient.Update(context.Background(), toUpdate)).To(Succeed())

			Eventually(func() []reconciler.VerifyResult {
				return resourceManager.GetRecord(aId).States
			}, timeout, interval).Should(ContainElement(reconciler.VerifyResultUpdateRequired))

			waitUntilReconcileState(key, reconciler.Failed)

			updated, _ := getObject(key)
			Expect(updated.Status.Message).To(HavePrefix("permission to update Azure resource is not set"))
		})

	})
})
