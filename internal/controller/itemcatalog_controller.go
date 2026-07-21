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

const itemCatalogFinalizer = "kubegame.systemcraftsman.com/finalizer"

type ItemCatalogReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=itemcatalogs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=itemcatalogs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=itemcatalogs/finalizers,verbs=update

func (r *ItemCatalogReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var catalog v1alpha1.ItemCatalog
	if err := r.Get(ctx, req.NamespacedName, &catalog); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var game v1alpha1.Game
	if err := r.Get(ctx, types.NamespacedName{Name: catalog.Spec.Game, Namespace: catalog.Namespace}, &game); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Game resource not found", "game", catalog.Spec.Game)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !catalog.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&catalog, itemCatalogFinalizer) {
			if game.Status.Ready {
				db, err := getGameDB(ctx, r.Client, &game)
				if err == nil {
					for _, item := range catalog.Spec.Items {
						db.Where("item_name = ? AND game = ?", item.Name, catalog.Spec.Game).Delete(&persistence.ItemEffectRecord{})
					}
					for _, item := range catalog.Spec.Items {
						db.Where("name = ? AND game = ?", item.Name, catalog.Spec.Game).Delete(&persistence.ItemDefinition{})
					}
					logger.Info("Deleted item catalog records from database", "catalog", catalog.Name)
				}
			}

			controllerutil.RemoveFinalizer(&catalog, itemCatalogFinalizer)
			if err := r.Update(ctx, &catalog); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&catalog, itemCatalogFinalizer) {
		controllerutil.AddFinalizer(&catalog, itemCatalogFinalizer)
		if err := r.Update(ctx, &catalog); err != nil {
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

	if err := persistence.RunMigrations(db, persistence.ItemCatalogModels()...); err != nil {
		logger.Error(err, "Failed to run item catalog migrations")
		return ctrl.Result{}, err
	}

	for _, item := range catalog.Spec.Items {
		record := &persistence.ItemDefinition{
			Name:      item.Name,
			Game:      catalog.Spec.Game,
			Category:  item.Category,
			Rarity:    item.Rarity,
			Stackable: item.Stackable,
			MaxStack:  item.MaxStack,
			Duration:  item.Duration,
		}
		if result := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}, {Name: "game"}},
			DoUpdates: clause.AssignmentColumns([]string{"category", "rarity", "stackable", "max_stack", "duration"}),
		}).Create(record); result.Error != nil {
			logger.Error(result.Error, "Failed to upsert item definition", "item", item.Name)
			return ctrl.Result{}, result.Error
		}

		for _, effect := range item.Effects {
			effectRecord := &persistence.ItemEffectRecord{
				ItemName:  item.Name,
				Game:      catalog.Spec.Game,
				Attribute: effect.Attribute,
				Modifier:  effect.Modifier,
			}
			if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(effectRecord); result.Error != nil {
				logger.Error(result.Error, "Failed to insert item effect record")
				return ctrl.Result{}, result.Error
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *ItemCatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ItemCatalog{}).
		Complete(r)
}
