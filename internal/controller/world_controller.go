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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"systemcraftsman.com/kubegame/internal/common"
	"systemcraftsman.com/kubegame/internal/persistence"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "systemcraftsman.com/kubegame/api/v1alpha1"
)

type WorldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=worlds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=worlds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=worlds/finalizers,verbs=update

func (r *WorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the World resource
	var world v1alpha1.World
	if err := r.Get(ctx, req.NamespacedName, &world); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("World resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get World resource")
		return ctrl.Result{}, err
	}

	// Fetch the Game resource
	var game v1alpha1.Game
	if err := r.Get(ctx, types.NamespacedName{Name: world.Spec.Game, Namespace: world.Namespace}, &game); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Game resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Game resource")
		return ctrl.Result{}, err
	}

	if game.Status.Ready == true {
		var postgresService corev1.Service
		if err := r.Get(ctx, types.NamespacedName{Name: game.Name + common.PostgresSuffix, Namespace: game.Namespace}, &postgresService); err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Postgres service not found, might be deleted")
				return ctrl.Result{}, nil
			}
			logger.Error(err, "Failed to get the postgres service")
			return ctrl.Result{}, err
		}

		db, err := persistence.CreateDatabaseConnection(postgresService.Name, game.Spec.Database.Username,
			game.Spec.Database.Password)
		if err != nil {
			logger.Error(err, "Failed to create database connection")
			return ctrl.Result{}, err
		}

		if !db.Migrator().HasTable(persistence.World{}) {
			if err := db.Migrator().CreateTable(persistence.World{}); err != nil {
				logger.Error(err, "Failed to create the World table")
				return ctrl.Result{}, err
			}
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
			logger.Error(err, "Failed to insert the record into the World table")
			return ctrl.Result{}, result.Error
		}

	}

	return ctrl.Result{}, nil
}

func (r *WorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.World{}).
		Complete(r)
}
