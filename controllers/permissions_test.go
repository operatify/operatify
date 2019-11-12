package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/szoio/resource-operator-factory/controllers/manager"
	"github.com/szoio/resource-operator-factory/reconciler"
)

var _ = Describe("Test permissions", func() {
	Context("when permissions are set", func() {
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

		It("should delete if delete permission present", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecWithAnnotations(aId, map[string]string{accessPermissionAnnotation: "CD"})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.States[len(record.States)-1]).To(Equal(reconciler.VerifyResultReady))

			By("Expecting to delete successfully")
			Expect(deleteObject(key)).To(Succeed())

			By("Expecting delete to finish")
			waitUntilObjectMissing(key)

			By("Underlying resource should be deleted")
			record = resourceManager.GetRecord(aId)
			Expect(record.States[len(record.States)-1]).To(Equal(reconciler.VerifyResultMissing))
		})

		It("should leave the underlying resource alone when deleting if delete permission is not present", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecWithAnnotations(aId, map[string]string{accessPermissionAnnotation: "C"})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileState(key, reconciler.Succeeded)

			record := resourceManager.GetRecord(aId)
			Expect(record.States[len(record.States)-1]).To(Equal(reconciler.VerifyResultReady))

			By("Expecting to delete successfully")
			Expect(deleteObject(key)).To(Succeed())

			By("Expecting delete to finish")
			waitUntilObjectMissing(key)

			By("Underlying resource should be untouched")
			record = resourceManager.GetRecord(aId)
			Expect(record.States[len(record.States)-1]).To(Equal(reconciler.VerifyResultReady))
		})

	})
})
