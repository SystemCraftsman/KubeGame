package api

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "systemcraftsman.com/kubegame/api/v1alpha1"
	"systemcraftsman.com/kubegame/internal/common"
	"systemcraftsman.com/kubegame/internal/persistence"
)

type Server struct {
	handler    *Handler
	httpServer *http.Server
}

func NewServer(k8sClient client.Client, addr string) *Server {
	getDB := func(game, namespace string) (*gorm.DB, error) {
		return resolveGameDB(k8sClient, game, namespace)
	}

	h := NewHandler(getDB)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/worlds", h.ListWorlds)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/worlds/{name}", h.GetWorld)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars", h.CreateAvatarInstance)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars", h.ListAvatarInstances)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}", h.GetAvatarInstance)
	mux.HandleFunc("DELETE /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}", h.DeleteAvatarInstance)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/worlds/{world}/areas", h.ListAreas)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/worlds/{world}/areas/{name}", h.GetArea)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/items", h.ListItems)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/items/{name}", h.GetItem)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/inventory", h.GrantItem)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/equip", h.EquipItem)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/unequip", h.UnequipItem)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/powerups/activate", h.ActivatePowerup)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/powerups", h.ListActivePowerups)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/currencies", h.ListCurrencies)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/currencies/{name}", h.GetCurrency)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/wallet", h.GetWallet)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/wallet/credit", h.CreditWallet)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/wallet/debit", h.DebitWallet)
	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}/wallet/transfer", h.TransferWallet)

	mux.HandleFunc("GET /api/v1/games/{game}/worlds", h.ListWorlds)
	mux.HandleFunc("GET /api/v1/games/{game}/worlds/{name}", h.GetWorld)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars", h.CreateAvatarInstance)
	mux.HandleFunc("GET /api/v1/games/{game}/avatars", h.ListAvatarInstances)
	mux.HandleFunc("GET /api/v1/games/{game}/avatars/{name}", h.GetAvatarInstance)
	mux.HandleFunc("DELETE /api/v1/games/{game}/avatars/{name}", h.DeleteAvatarInstance)
	mux.HandleFunc("GET /api/v1/games/{game}/worlds/{world}/areas", h.ListAreas)
	mux.HandleFunc("GET /api/v1/games/{game}/worlds/{world}/areas/{name}", h.GetArea)
	mux.HandleFunc("GET /api/v1/games/{game}/items", h.ListItems)
	mux.HandleFunc("GET /api/v1/games/{game}/items/{name}", h.GetItem)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars/{name}/inventory", h.GrantItem)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars/{name}/equip", h.EquipItem)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars/{name}/unequip", h.UnequipItem)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars/{name}/powerups/activate", h.ActivatePowerup)
	mux.HandleFunc("GET /api/v1/games/{game}/avatars/{name}/powerups", h.ListActivePowerups)
	mux.HandleFunc("GET /api/v1/games/{game}/currencies", h.ListCurrencies)
	mux.HandleFunc("GET /api/v1/games/{game}/currencies/{name}", h.GetCurrency)
	mux.HandleFunc("GET /api/v1/games/{game}/avatars/{name}/wallet", h.GetWallet)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars/{name}/wallet/credit", h.CreditWallet)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars/{name}/wallet/debit", h.DebitWallet)
	mux.HandleFunc("POST /api/v1/games/{game}/avatars/{name}/wallet/transfer", h.TransferWallet)

	handler := withCORS(withRecovery(withLogging(mux)))

	return &Server{
		handler: h,
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting API server", "addr", s.httpServer.Addr)

	go func() {
		<-ctx.Done()
		s.httpServer.Shutdown(context.Background())
	}()

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("API server error: %v", err)
	}
	return nil
}

func resolveGameDB(k8sClient client.Client, gameName, namespace string) (*gorm.DB, error) {
	if namespace == "" {
		namespace = "default"
	}

	var game v1alpha1.Game
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: gameName, Namespace: namespace}, &game); err != nil {
		return nil, fmt.Errorf("game %q not found in namespace %q: %v", gameName, namespace, err)
	}

	var secret corev1.Secret
	if err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      game.Spec.Database.SecretRef,
		Namespace: namespace,
	}, &secret); err != nil {
		return nil, fmt.Errorf("secret %q not found: %v", game.Spec.Database.SecretRef, err)
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	serviceName := gameName + common.PostgresSuffix

	return persistence.GetOrCreateConnection(serviceName, username, password)
}
