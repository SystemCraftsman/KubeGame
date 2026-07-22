/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"gorm.io/gorm/clause"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "systemcraftsman.com/kubegame/api/v1alpha1"
	"systemcraftsman.com/kubegame/internal/persistence"
)

const worldFinalizer = "kubegame.systemcraftsman.com/finalizer"

type WorldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=worlds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=worlds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=worlds/finalizers,verbs=update

func (r *WorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var world v1alpha1.World
	if err := r.Get(ctx, req.NamespacedName, &world); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("World resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get World resource")
		return ctrl.Result{}, err
	}

	if !world.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&world, worldFinalizer) {
			var game v1alpha1.Game
			if err := r.Get(ctx, types.NamespacedName{Name: world.Spec.Game, Namespace: world.Namespace}, &game); err == nil && game.Status.Ready {
				db, err := getGameDB(ctx, r.Client, &game)
				if err == nil {
					db.Where("name = ?", world.Name).Delete(&persistence.World{})
					logger.Info("Deleted world record from database", "world", world.Name)
				}
			}

			controllerutil.RemoveFinalizer(&world, worldFinalizer)
			if err := r.Update(ctx, &world); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	var game v1alpha1.Game
	if err := r.Get(ctx, types.NamespacedName{Name: world.Spec.Game, Namespace: world.Namespace}, &game); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Game resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Game resource")
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(&world, worldFinalizer) {
		controllerutil.AddFinalizer(&world, worldFinalizer)
		if err := r.Update(ctx, &world); err != nil {
			return ctrl.Result{}, err
		}
	}

	if _, updated, err := ensureLabels(ctx, r.Client, &world, map[string]string{
		labelGame: world.Spec.Game,
	}); err != nil {
		return ctrl.Result{}, err
	} else if updated {
		return ctrl.Result{}, nil
	}

	if !game.Status.Ready {
		logger.Info("Game not ready yet, requeuing", "game", game.Name)
		return ctrl.Result{Requeue: true}, nil
	}

	db, err := getGameDB(ctx, r.Client, &game)
	if err != nil {
		logger.Error(err, "Failed to get database connection")
		return ctrl.Result{}, err
	}

	if err := persistence.RunMigrations(db, persistence.WorldModels()...); err != nil {
		logger.Error(err, "Failed to run world migrations")
		return ctrl.Result{}, err
	}

	worldRecord := &persistence.World{
		Name:        world.Name,
		Game:        world.Spec.Game,
		Description: world.Spec.Description,
	}

	if result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(worldRecord); result.Error != nil {
		logger.Error(result.Error, "Failed to insert the record into the World table")
		return ctrl.Result{}, result.Error
	}

	return ctrl.Result{}, nil
}

func (r *WorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.World{}).
		Complete(r)
}
