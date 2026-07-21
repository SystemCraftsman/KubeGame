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

var _ = Describe("ItemReconciler", func() {
	const (
		timeout  = 30 * time.Second
		interval = 250 * time.Millisecond
	)

	Context("when an Item CR is created with a ready Game", func() {
		var (
			game   *kubegamev1alpha1.Game
			item   *kubegamev1alpha1.Item
			secret *corev1.Secret
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

			item = &kubegamev1alpha1.Item{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "iron-sword",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.ItemSpec{
					Game:      game.Name,
					Category:  "Equipment",
					Rarity:    "Common",
					Stackable: false,
					Effects: []kubegamev1alpha1.ItemEffect{
						{Attribute: "strength", Modifier: "+5"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, item)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, item)).To(Succeed())
			Expect(k8sClient.Delete(ctx, game)).To(Succeed())
			Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
		})

		It("should add a finalizer to the Item", func() {
			key := types.NamespacedName{Name: item.Name, Namespace: item.Namespace}

			Eventually(func() []string {
				var i kubegamev1alpha1.Item
				if err := k8sClient.Get(ctx, key, &i); err != nil {
					return nil
				}
				return i.Finalizers
			}, timeout, interval).Should(ContainElement(itemFinalizer))
		})

		It("should persist the item definition in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() bool {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return false
				}
				var record persistence.ItemDefinition
				result := db.Where("name = ? AND game = ?", item.Name, game.Name).First(&record)
				return result.Error == nil && record.Category == "Equipment"
			}, timeout, interval).Should(BeTrue())
		})

		It("should persist item effects in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() int {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return 0
				}
				var records []persistence.ItemEffectRecord
				db.Where("item_name = ? AND game = ?", item.Name, game.Name).Find(&records)
				return len(records)
			}, timeout, interval).Should(Equal(1))
		})
	})
})
