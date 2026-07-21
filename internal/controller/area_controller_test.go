package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kubegamev1alpha1 "systemcraftsman.com/kubegame/api/v1alpha1"
	"systemcraftsman.com/kubegame/internal/persistence"
)

var _ = Describe("AreaReconciler", func() {
	const (
		timeout  = 30 * time.Second
		interval = 250 * time.Millisecond
	)

	Context("when an Area CR is created with a ready Game", func() {
		var (
			game   *kubegamev1alpha1.Game
			world  *kubegamev1alpha1.World
			area   *kubegamev1alpha1.Area
			secret *corev1.Secret
		)

		BeforeEach(func() {
			secret = createDBSecret("area-test-db-creds")
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			game = &kubegamev1alpha1.Game{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "area-test-game",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.GameSpec{
					Database: kubegamev1alpha1.Database{
						SecretRef: "area-test-db-creds",
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
					Name:      "area-test-world",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.WorldSpec{
					Game:        game.Name,
					Description: "Test world for area tests",
				},
			}
			Expect(k8sClient.Create(ctx, world)).To(Succeed())

			area = &kubegamev1alpha1.Area{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-area",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.AreaSpec{
					Game:        game.Name,
					World:       world.Name,
					Description: "A test area",
					ConnectedAreas: []string{
						"other-area-1",
						"other-area-2",
					},
					Properties: []kubegamev1alpha1.AreaProperty{
						{Name: "pvpEnabled", Value: "true"},
						{Name: "level", Value: "5"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, area)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, area)).To(Succeed())
			Expect(k8sClient.Delete(ctx, world)).To(Succeed())
			Expect(k8sClient.Delete(ctx, game)).To(Succeed())
			Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
		})

		It("should add a finalizer to the Area", func() {
			areaKey := types.NamespacedName{Name: area.Name, Namespace: area.Namespace}

			Eventually(func() []string {
				var a kubegamev1alpha1.Area
				if err := k8sClient.Get(ctx, areaKey, &a); err != nil {
					return nil
				}
				return a.Finalizers
			}, timeout, interval).Should(ContainElement(areaFinalizer))
		})

		It("should persist the Area record in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() bool {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return false
				}
				var record persistence.Area
				result := db.Where("name = ?", area.Name).First(&record)
				return result.Error == nil && record.World == world.Name
			}, timeout, interval).Should(BeTrue())
		})

		It("should persist area connections in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() int {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return 0
				}
				var records []persistence.AreaConnection
				db.Where("area_name = ?", area.Name).Find(&records)
				return len(records)
			}, timeout, interval).Should(Equal(2))
		})

		It("should persist area properties in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() int {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return 0
				}
				var records []persistence.AreaPropertyRecord
				db.Where("area_name = ?", area.Name).Find(&records)
				return len(records)
			}, timeout, interval).Should(Equal(2))
		})
	})
})
