package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kubegamev1alpha1 "systemcraftsman.com/kubegame/api/v1alpha1"
	"systemcraftsman.com/kubegame/internal/persistence"
)

var _ = Describe("WorldReconciler", func() {
	const (
		timeout  = 30 * time.Second
		interval = 250 * time.Millisecond
	)

	Context("when a World CR is created with a ready Game", func() {
		var (
			game  *kubegamev1alpha1.Game
			world *kubegamev1alpha1.World
		)

		BeforeEach(func() {
			game = &kubegamev1alpha1.Game{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "world-test-game",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.GameSpec{
					Database: kubegamev1alpha1.Database{
						Username: testDBUser,
						Password: testDBPassword,
					},
				},
			}
			Expect(k8sClient.Create(ctx, game)).To(Succeed())

			gameKey := types.NamespacedName{Name: game.Name, Namespace: game.Namespace}
			Eventually(func() bool {
				var g kubegamev1alpha1.Game
				if err := k8sClient.Get(ctx, gameKey, &g); err != nil {
					return false
				}
				return g.Status.Ready
			}, timeout, interval).Should(BeTrue())

			world = &kubegamev1alpha1.World{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ludus",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.WorldSpec{
					Game:        game.Name,
					Description: "The school planet in the OASIS",
				},
			}
			Expect(k8sClient.Create(ctx, world)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, world)).To(Succeed())
			Expect(k8sClient.Delete(ctx, game)).To(Succeed())
		})

		It("should add a finalizer to the World", func() {
			worldKey := types.NamespacedName{Name: world.Name, Namespace: world.Namespace}

			Eventually(func() []string {
				var w kubegamev1alpha1.World
				if err := k8sClient.Get(ctx, worldKey, &w); err != nil {
					return nil
				}
				return w.Finalizers
			}, timeout, interval).Should(ContainElement(worldFinalizer))
		})

		It("should persist the World record in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() bool {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return false
				}
				var record persistence.World
				result := db.Where("name = ?", world.Name).First(&record)
				return result.Error == nil && record.Name == world.Name
			}, timeout, interval).Should(BeTrue())
		})
	})
})
