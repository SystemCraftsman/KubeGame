package api

import (
	"context"
	"fmt"
	"net/http"

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

	mux.HandleFunc("POST /api/v1/namespaces/{namespace}/games/{game}/avatars", h.CreateAvatarInstance)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars", h.ListAvatarInstances)
	mux.HandleFunc("GET /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}", h.GetAvatarInstance)
	mux.HandleFunc("DELETE /api/v1/namespaces/{namespace}/games/{game}/avatars/{name}", h.DeleteAvatarInstance)

	mux.HandleFunc("POST /api/v1/games/{game}/avatars", h.CreateAvatarInstance)
	mux.HandleFunc("GET /api/v1/games/{game}/avatars", h.ListAvatarInstances)
	mux.HandleFunc("GET /api/v1/games/{game}/avatars/{name}", h.GetAvatarInstance)
	mux.HandleFunc("DELETE /api/v1/games/{game}/avatars/{name}", h.DeleteAvatarInstance)

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

	serviceName := gameName + common.PostgresSuffix

	username := game.Spec.Database.Username
	password := game.Spec.Database.Password

	return persistence.GetOrCreateConnection(serviceName, username, password)
}
