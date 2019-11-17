package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/szoio/operatify/controllers/manager"
	"github.com/szoio/operatify/reconciler"
)

var _ = Describe("Test Failure Scenarios", func() {

	Context("when there are failures", func() {

		It("should fail if fails to create", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecA(aId)

			// tell it to fail to create
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventCreate,
				Operation: manager.CreateFail.AsOperation(),
			})

			// Create
			By("Expecting to create successfully")
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileStateA(key, reconciler.Failed)

			By("Expecting correct state transitions")
			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultError,
			}))

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObjectA(key)).Should(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissingA(key)
		})

		It("should fail if fails complete async creation", func() {
			aId := "a-" + RandomString(10)
			key, created := nameAndSpecA(aId)

			// tell it to fail to create
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventCreate,
				Operation: manager.CreateCompleteFail.AsOperation(),
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileStateA(key, reconciler.Failed)

			record := resourceManager.GetRecord(aId)
			Expect(record.States).Should(Equal([]reconciler.VerifyResult{
				reconciler.VerifyResultInProgress,
				reconciler.VerifyResultError,
			}))

			// Remove the behaviours
			resourceManager.ClearBehaviours(aId)

			// Delete
			By("Expecting to delete successfully")
			Expect(deleteObjectA(key)).Should(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissingA(key)

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
			key, created := nameAndSpecA(aId)

			// tell it to fail verification
			resourceManager.AddBehaviour(aId, manager.Behaviour{
				Event:     manager.EventGet,
				Operation: manager.VerifyFail.AsOperation(),
				From:      1,
			})

			// Create
			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())
			waitUntilReconcileStateA(key, reconciler.Failed)

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
			Expect(deleteObjectA(key)).Should(Succeed())

			By("Expecting to delete finish")
			waitUntilObjectMissingA(key)

			record := resourceManager.GetRecord(aId)
			expectedStates := append(recordPreStates,
				reconciler.VerifyResultDeleting,
				reconciler.VerifyResultMissing)
			Expect(record.States).Should(Equal(expectedStates))
		})

		It("should fail if fails to update", func() {
			// TODO:
		})

		It("should fail if fails to delete and recreate", func() {
			// TODO:
		})
	})
})
