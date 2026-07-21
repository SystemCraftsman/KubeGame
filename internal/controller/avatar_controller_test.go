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

var _ = Describe("AvatarReconciler", func() {
	const (
		timeout  = 30 * time.Second
		interval = 250 * time.Millisecond
	)

	Context("when an Avatar CR is created with a ready Game", func() {
		var (
			game   *kubegamev1alpha1.Game
			avatar *kubegamev1alpha1.Avatar
		)

		BeforeEach(func() {
			game = &kubegamev1alpha1.Game{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "avatar-test-game",
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

			avatar = &kubegamev1alpha1.Avatar{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-avatar",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.AvatarSpec{
					Game:        game.Name,
					Type:        "Adventurer",
					Description: "Test adventurer avatar",
					AttributeTypes: []kubegamev1alpha1.AttributeType{
						{Name: "strength", ValueType: "int"},
						{Name: "intelligence", ValueType: "int"},
					},
					InventoryTypes: []kubegamev1alpha1.InventoryType{
						{Name: "Weapon", Category: "Equipment"},
						{Name: "Potion", Category: "Consumable"},
					},
					AchievementTypes: []kubegamev1alpha1.AchievementType{
						{Name: "First Blood", Description: "Win your first battle"},
					},
				},
			}
			Expect(k8sClient.Create(ctx, avatar)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, avatar)).To(Succeed())
			Expect(k8sClient.Delete(ctx, game)).To(Succeed())
		})

		It("should add a finalizer to the Avatar", func() {
			avatarKey := types.NamespacedName{Name: avatar.Name, Namespace: avatar.Namespace}

			Eventually(func() []string {
				var a kubegamev1alpha1.Avatar
				if err := k8sClient.Get(ctx, avatarKey, &a); err != nil {
					return nil
				}
				return a.Finalizers
			}, timeout, interval).Should(ContainElement(avatarFinalizer))
		})

		It("should persist the AvatarType record in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() bool {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return false
				}
				var record persistence.AvatarType
				result := db.Where("name = ?", avatar.Name).First(&record)
				return result.Error == nil && record.Type == "Adventurer"
			}, timeout, interval).Should(BeTrue())
		})

		It("should persist attribute types in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() int {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return 0
				}
				var records []persistence.AttributeType
				db.Where("avatar_name = ?", avatar.Name).Find(&records)
				return len(records)
			}, timeout, interval).Should(Equal(2))
		})

		It("should persist inventory types in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() int {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return 0
				}
				var records []persistence.InventoryType
				db.Where("avatar_name = ?", avatar.Name).Find(&records)
				return len(records)
			}, timeout, interval).Should(Equal(2))
		})

		It("should persist achievement types in PostgreSQL", func() {
			serviceName := game.Name + "-postgres"

			Eventually(func() int {
				db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
				if err != nil {
					return 0
				}
				var records []persistence.AchievementType
				db.Where("avatar_name = ?", avatar.Name).Find(&records)
				return len(records)
			}, timeout, interval).Should(Equal(1))
		})
	})
})
