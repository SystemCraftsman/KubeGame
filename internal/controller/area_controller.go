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

const areaFinalizer = "kubegame.systemcraftsman.com/finalizer"

type AreaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=areas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=areas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=areas/finalizers,verbs=update

func (r *AreaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var area v1alpha1.Area
	if err := r.Get(ctx, req.NamespacedName, &area); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Area resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Area resource")
		return ctrl.Result{}, err
	}

	var game v1alpha1.Game
	if err := r.Get(ctx, types.NamespacedName{Name: area.Spec.Game, Namespace: area.Namespace}, &game); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Game resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Game resource")
		return ctrl.Result{}, err
	}

	if !area.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&area, areaFinalizer) {
			if game.Status.Ready {
				db, err := getGameDB(ctx, r.Client, &game)
				if err == nil {
					db.Where("area_name = ?", area.Name).Delete(&persistence.AreaConnection{})
					db.Where("area_name = ?", area.Name).Delete(&persistence.AreaPropertyRecord{})
					db.Where("name = ?", area.Name).Delete(&persistence.Area{})
					logger.Info("Deleted area records from database", "area", area.Name)
				}
			}

			controllerutil.RemoveFinalizer(&area, areaFinalizer)
			if err := r.Update(ctx, &area); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&area, areaFinalizer) {
		controllerutil.AddFinalizer(&area, areaFinalizer)
		if err := r.Update(ctx, &area); err != nil {
			return ctrl.Result{}, err
		}
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

	if err := persistence.RunMigrations(db, persistence.AreaModels()...); err != nil {
		logger.Error(err, "Failed to run area migrations")
		return ctrl.Result{}, err
	}

	areaRecord := &persistence.Area{
		Name:        area.Name,
		Game:        area.Spec.Game,
		World:       area.Spec.World,
		Description: area.Spec.Description,
	}

	if result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(areaRecord); result.Error != nil {
		logger.Error(result.Error, "Failed to insert the record into the Area table")
		return ctrl.Result{}, result.Error
	}

	for _, conn := range area.Spec.ConnectedAreas {
		record := &persistence.AreaConnection{
			AreaName:   area.Name,
			ConnectsTo: conn,
		}
		if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(record); result.Error != nil {
			logger.Error(result.Error, "Failed to insert area connection record")
			return ctrl.Result{}, result.Error
		}
	}

	for _, prop := range area.Spec.Properties {
		record := &persistence.AreaPropertyRecord{
			AreaName: area.Name,
			Name:     prop.Name,
			Value:    prop.Value,
		}
		if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(record); result.Error != nil {
			logger.Error(result.Error, "Failed to insert area property record")
			return ctrl.Result{}, result.Error
		}
	}

	return ctrl.Result{}, nil
}

func (r *AreaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Area{}).
		Complete(r)
}
