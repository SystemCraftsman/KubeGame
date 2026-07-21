package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
	getDB := func(game string) (*gorm.DB, error) {
		return resolveGameDB(k8sClient, game)
	}

	h := NewHandler(getDB)
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		isInstanceList := strings.HasSuffix(path, "/avatars") || strings.HasSuffix(path, "/avatars/")
		hasInstanceName := !isInstanceList && strings.Contains(path, "/avatars/")

		switch {
		case r.Method == http.MethodPost && isInstanceList:
			h.CreateAvatarInstance(w, r)
		case r.Method == http.MethodGet && isInstanceList:
			h.ListAvatarInstances(w, r)
		case r.Method == http.MethodGet && hasInstanceName:
			h.GetAvatarInstance(w, r)
		case r.Method == http.MethodDelete && hasInstanceName:
			h.DeleteAvatarInstance(w, r)
		default:
			writeError(w, http.StatusNotFound, "not found")
		}
	})

	return &Server{
		handler: h,
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
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

func resolveGameDB(k8sClient client.Client, gameName string) (*gorm.DB, error) {
	var game v1alpha1.Game
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: gameName, Namespace: "default"}, &game); err != nil {
		return nil, fmt.Errorf("game %q not found: %v", gameName, err)
	}

	serviceName := gameName + common.PostgresSuffix
	return persistence.CreateDatabaseConnection(serviceName, game.Spec.Database.Username, game.Spec.Database.Password)
}
