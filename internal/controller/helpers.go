package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "systemcraftsman.com/kubegame/api/v1alpha1"
	"systemcraftsman.com/kubegame/internal/common"
	"systemcraftsman.com/kubegame/internal/persistence"

	"gorm.io/gorm"
)

const (
	labelGame  = "kubegame.systemcraftsman.com/game"
	labelWorld = "kubegame.systemcraftsman.com/world"
)

func resolveCredentials(ctx context.Context, c client.Client, game *v1alpha1.Game) (string, string, error) {
	var secret corev1.Secret
	if err := c.Get(ctx, types.NamespacedName{
		Name:      game.Spec.Database.SecretRef,
		Namespace: game.Namespace,
	}, &secret); err != nil {
		return "", "", fmt.Errorf("secret %q not found: %v", game.Spec.Database.SecretRef, err)
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	if username == "" || password == "" {
		return "", "", fmt.Errorf("secret %q must contain 'username' and 'password' keys", game.Spec.Database.SecretRef)
	}
	return username, password, nil
}

func getGameDB(ctx context.Context, c client.Client, game *v1alpha1.Game) (*gorm.DB, error) {
	var postgresService corev1.Service
	if err := c.Get(ctx, types.NamespacedName{
		Name:      game.Name + common.PostgresSuffix,
		Namespace: game.Namespace,
	}, &postgresService); err != nil {
		return nil, fmt.Errorf("postgres service not found for game %q: %v", game.Name, err)
	}

	username, password, err := resolveCredentials(ctx, c, game)
	if err != nil {
		return nil, err
	}

	return persistence.GetOrCreateConnection(postgresService.Name, username, password)
}

func ensureLabels(ctx context.Context, c client.Client, obj client.Object, desired map[string]string) (ctrl.Result, bool, error) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	needsUpdate := false
	for k, v := range desired {
		if labels[k] != v {
			labels[k] = v
			needsUpdate = true
		}
	}
	if needsUpdate {
		obj.SetLabels(labels)
		if err := c.Update(ctx, obj); err != nil {
			return ctrl.Result{}, false, err
		}
		return ctrl.Result{}, true, nil
	}
	return ctrl.Result{}, false, nil
}

