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

var _ = Describe("ItemCatalogReconciler", func() {
	const (
		timeout  = 30 * time.Second
		interval = 250 * time.Millisecond
	)

	Context("when an ItemCatalog CR is created with a ready Game", func() {
		var (
			game    *kubegamev1alpha1.Game
			catalog *kubegamev1alpha1.ItemCatalog
			secret  *corev1.Secret
		)

		BeforeEach(func() {
			secret = createDBSecret("item-test-db-creds")
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			game = &kubegamev1alpha1.Game{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "item-test-game",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.GameSpec{
					Database: kubegamev1alpha1.Database{
						SecretRef: "item-test-db-creds",
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

			catalog = &kubegamev1alpha1.ItemCatalog{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-items",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.ItemCatalogSpec{
					Game: game.Name,
					Items: []kubegamev1alpha1.ItemSpec{
						{
							Name:      "Iron Sword",
							Category:  "Equipment",
							Rarity:    "Common",
							Stackable: false,
							Effects: []kubegamev1alpha1.ItemEffect{
								{Attribute: "strength", Modifier: "+5"},
							},
						},
						{
							Name:      "Health Potion",
							Category:  "Powerup",
							Rarity:    "Common",
							Stackable: true,
							MaxStack:  10,
							Duration:  300,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, catalog)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, catalog)).To(Succeed())
			Expect(k8sClient.Delete(ctx, game)).To(Succeed())
			Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
		})

		It("should add a finalizer to the ItemCatalog", func() {
			key := types.NamespacedName{Name: catalog.Name, Namespace: catalog.Namespace}

			Eventually(func() []string {
				var c kubegamev1alpha1.ItemCatalog
				if err := k8sClient.Get(ctx, key, &c); err != nil {
					return nil
				}
				return c.Finalizers
			}, timeout, interval).Should(ContainElement(itemCatalogFinalizer))
		})

		It("should persist item definitions in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() int {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return 0
				}
				var records []persistence.ItemDefinition
				db.Where("game = ?", game.Name).Find(&records)
				return len(records)
			}, timeout, interval).Should(Equal(2))
		})

		It("should persist item effects in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() int {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return 0
				}
				var records []persistence.ItemEffectRecord
				db.Where("game = ?", game.Name).Find(&records)
				return len(records)
			}, timeout, interval).Should(Equal(1))
		})

		It("should persist stackable item properties", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() bool {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return false
				}
				var record persistence.ItemDefinition
				result := db.Where("name = ? AND game = ?", "Health Potion", game.Name).First(&record)
				return result.Error == nil && record.Stackable && record.MaxStack == 10 && record.Duration == 300
			}, timeout, interval).Should(BeTrue())
		})
	})
})
