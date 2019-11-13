package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/szoio/resource-operator-factory/controllers/manager"
	"github.com/szoio/resource-operator-factory/reconciler"
)

var _ = Describe("Test Update and Recreate", func() {

	Context("when updating and recreating", func() {

		It("should update synchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecA(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).To(Succeed())
			waitUntilReconcileStateA(key, reconciler.Succeeded)

			// tell it update is required (ony for the next verify)
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventGet,
				Operation: manager.VerifyNeedsUpdate.AsOperation(),
				From:      resourceManager.CountEvents(aId, manager.EventGet),
				Count:     1,
			})

			toUpdate, _ := getObjectA(key)
			toUpdate.Spec.IntData = 1
			toUpdate.Spec.StringData = "Updated"
			Expect(k8sClient.Update(context.Background(), toUpdate)).To(Succeed())

			Eventually(func() []reconciler.VerifyResult {
				record := resourceManager.GetRecord(aId)
				return record.States
			}, timeout, interval).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultUpdateRequired,
				reconciler.VerifyResultReady,
			}))

			Expect(resourceManager.CountEvents(aId, manager.EventCreate)).To(Equal(1))
			Expect(resourceManager.CountEvents(aId, manager.EventUpdate)).To(Equal(1))

			updated, _ := getObjectA(key)
			Expect(updated.Spec.StringData).To(Equal("Updated"))
		})

		It("should update asynchronously", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecA(aId)

			// Create
			Expect(k8sClient.Create(context.Background(), created)).To(Succeed())
			waitUntilReconcileStateA(key, reconciler.Succeeded)

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

			toUpdate, _ := getObjectA(key)
			toUpdate.Spec.IntData = 1
			toUpdate.Spec.StringData = "Updated"
			Expect(k8sClient.Update(context.Background(), toUpdate)).To(Succeed())
			waitUntilReconcileStateA(key, reconciler.Succeeded)

			Eventually(func() []reconciler.VerifyResult {
				record := resourceManager.GetRecord(aId)
				return record.States
			}, timeout, interval).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
				reconciler.VerifyResultUpdateRequired,
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultReady,
			}))

			Expect(resourceManager.CountEvents(aId, manager.EventCreate)).To(Equal(1))
			Expect(resourceManager.CountEvents(aId, manager.EventUpdate)).To(Equal(1))

			updated, _ := getObjectA(key)
			Expect(updated.Spec.StringData).To(Equal("Updated"))
		})
	})
})
