package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kubegamev1alpha1 "systemcraftsman.com/kubegame/api/v1alpha1"
	"systemcraftsman.com/kubegame/internal/common"
)

func createDBSecret(name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte(testDBUser),
			"password": []byte(testDBPassword),
		},
	}
}

var _ = Describe("GameReconciler", func() {
	const (
		timeout  = 30 * time.Second
		interval = 250 * time.Millisecond
	)

	Context("when a Game CR is created", func() {
		var (
			game   *kubegamev1alpha1.Game
			secret *corev1.Secret
		)

		BeforeEach(func() {
			secret = createDBSecret("test-game-db-creds")
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())

			game = &kubegamev1alpha1.Game{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-game",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.GameSpec{
					Database: kubegamev1alpha1.Database{
						SecretRef: "test-game-db-creds",
					},
				},
			}
			Expect(k8sClient.Create(ctx, game)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, game)).To(Succeed())
			Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
		})

		It("should create a PostgreSQL Deployment", func() {
			deploymentKey := types.NamespacedName{
				Name:      game.Name + common.PostgresSuffix,
				Namespace: game.Namespace,
			}

			Eventually(func() error {
				return k8sClient.Get(ctx, deploymentKey, &appsv1.Deployment{})
			}, timeout, interval).Should(Succeed())
		})

		It("should create a PostgreSQL Service", func() {
			serviceKey := types.NamespacedName{
				Name:      game.Name + common.PostgresSuffix,
				Namespace: game.Namespace,
			}

			Eventually(func() error {
				return k8sClient.Get(ctx, serviceKey, &corev1.Service{})
			}, timeout, interval).Should(Succeed())
		})

		It("should set the Game status to Ready", func() {
			gameKey := types.NamespacedName{Name: game.Name, Namespace: game.Namespace}

			Eventually(func() bool {
				var g kubegamev1alpha1.Game
				if err := k8sClient.Get(ctx, gameKey, &g); err != nil {
					return false
				}
				return g.Status.Ready
			}, timeout, interval).Should(BeTrue())
		})

		It("should add a finalizer", func() {
			gameKey := types.NamespacedName{Name: game.Name, Namespace: game.Namespace}

			Eventually(func() []string {
				var g kubegamev1alpha1.Game
				if err := k8sClient.Get(ctx, gameKey, &g); err != nil {
					return nil
				}
				return g.Finalizers
			}, timeout, interval).Should(ContainElement(gameFinalizer))
		})
	})

	Context("when a Game CR references a missing Secret", func() {
		var game *kubegamev1alpha1.Game

		BeforeEach(func() {
			game = &kubegamev1alpha1.Game{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-game-bad-secret",
					Namespace: "default",
				},
				Spec: kubegamev1alpha1.GameSpec{
					Database: kubegamev1alpha1.Database{
						SecretRef: "nonexistent-secret",
					},
				},
			}
			Expect(k8sClient.Create(ctx, game)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, game)).To(Succeed())
		})

		It("should not set the Game status to Ready", func() {
			gameKey := types.NamespacedName{Name: game.Name, Namespace: game.Namespace}

			Consistently(func() bool {
				var g kubegamev1alpha1.Game
				if err := k8sClient.Get(ctx, gameKey, &g); err != nil {
					return false
				}
				return g.Status.Ready
			}, 3*time.Second, interval).Should(BeFalse())
		})
	})
})
