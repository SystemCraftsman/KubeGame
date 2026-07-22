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

var _ = Describe("CurrencyController", func() {
	const (
		timeout  = time.Second * 30
		interval = time.Millisecond * 250
	)

	var (
		game     *kubegamev1alpha1.Game
		currency *kubegamev1alpha1.Currency
		secret   *corev1.Secret
	)

	BeforeEach(func() {
		secret = createDBSecret("currency-test-db-creds")
		Expect(k8sClient.Create(ctx, secret)).To(Succeed())

		game = &kubegamev1alpha1.Game{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "currency-test-game",
				Namespace: "default",
			},
			Spec: kubegamev1alpha1.GameSpec{
				Database: kubegamev1alpha1.Database{
					SecretRef: "currency-test-db-creds",
				},
			},
		}
		Expect(k8sClient.Create(ctx, game)).To(Succeed())

		gameKey := types.NamespacedName{Name: game.Name, Namespace: game.Namespace}
		Eventually(func() bool {
			var g kubegamev1alpha1.Game
			_ = k8sClient.Get(ctx, gameKey, &g)
			return g.Status.Ready
		}, timeout, interval).Should(BeTrue())
	})

	AfterEach(func() {
		if currency != nil {
			_ = k8sClient.Delete(ctx, currency)
			currency = nil
		}
		Expect(k8sClient.Delete(ctx, game)).To(Succeed())
		Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
	})

	It("should add finalizer and game label", func() {
		currency = &kubegamev1alpha1.Currency{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-gold",
				Namespace: "default",
			},
			Spec: kubegamev1alpha1.CurrencySpec{
				Game:           "currency-test-game",
				Symbol:         "G",
				Tradeable:      true,
				MaxBalance:     1000000,
				InitialBalance: 100,
			},
		}
		Expect(k8sClient.Create(ctx, currency)).To(Succeed())

		currencyKey := types.NamespacedName{Name: currency.Name, Namespace: currency.Namespace}
		Eventually(func() []string {
			var c kubegamev1alpha1.Currency
			_ = k8sClient.Get(ctx, currencyKey, &c)
			return c.Finalizers
		}, timeout, interval).Should(ContainElement("kubegame.systemcraftsman.com/finalizer"))

		Eventually(func() string {
			var c kubegamev1alpha1.Currency
			_ = k8sClient.Get(ctx, currencyKey, &c)
			return c.Labels[labelGame]
		}, timeout, interval).Should(Equal("currency-test-game"))
	})

	It("should persist currency definition to PostgreSQL", func() {
		currency = &kubegamev1alpha1.Currency{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-credit",
				Namespace: "default",
			},
			Spec: kubegamev1alpha1.CurrencySpec{
				Game:           "currency-test-game",
				Symbol:         "CR",
				Tradeable:      false,
				MaxBalance:     500000,
				InitialBalance: 0,
			},
		}
		Expect(k8sClient.Create(ctx, currency)).To(Succeed())

		serviceName := game.Name + "-postgres"

		Eventually(func() int {
			db, err := persistence.GetOrCreateConnection(serviceName, testDBUser, testDBPassword)
			if err != nil {
				return 0
			}
			var records []persistence.CurrencyDefinition
			db.Where("name = ? AND game = ?", currency.Name, game.Name).Find(&records)
			return len(records)
		}, timeout, interval).Should(Equal(1))
	})
})
