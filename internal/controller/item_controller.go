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

const itemFinalizer = "kubegame.systemcraftsman.com/finalizer"

type ItemReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=items,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=items/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=items/finalizers,verbs=update

func (r *ItemReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var item v1alpha1.Item
	if err := r.Get(ctx, req.NamespacedName, &item); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var game v1alpha1.Game
	if err := r.Get(ctx, types.NamespacedName{Name: item.Spec.Game, Namespace: item.Namespace}, &game); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Game resource not found", "game", item.Spec.Game)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !item.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&item, itemFinalizer) {
			if game.Status.Ready {
				db, err := getGameDB(ctx, r.Client, &game)
				if err == nil {
					db.Where("item_name = ? AND game = ?", item.Name, item.Spec.Game).Delete(&persistence.ItemEffectRecord{})
					db.Where("name = ? AND game = ?", item.Name, item.Spec.Game).Delete(&persistence.ItemDefinition{})
					logger.Info("Deleted item record from database", "item", item.Name)
				}
			}

			controllerutil.RemoveFinalizer(&item, itemFinalizer)
			if err := r.Update(ctx, &item); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&item, itemFinalizer) {
		controllerutil.AddFinalizer(&item, itemFinalizer)
		if err := r.Update(ctx, &item); err != nil {
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

	if err := persistence.RunMigrations(db, persistence.ItemModels()...); err != nil {
		logger.Error(err, "Failed to run item migrations")
		return ctrl.Result{}, err
	}

	record := &persistence.ItemDefinition{
		Name:      item.Name,
		Game:      item.Spec.Game,
		Category:  item.Spec.Category,
		Rarity:    item.Spec.Rarity,
		Stackable: item.Spec.Stackable,
		MaxStack:  item.Spec.MaxStack,
		Duration:  item.Spec.Duration,
	}
	if result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}, {Name: "game"}},
		DoUpdates: clause.AssignmentColumns([]string{"category", "rarity", "stackable", "max_stack", "duration"}),
	}).Create(record); result.Error != nil {
		logger.Error(result.Error, "Failed to upsert item definition", "item", item.Name)
		return ctrl.Result{}, result.Error
	}

	for _, effect := range item.Spec.Effects {
		effectRecord := &persistence.ItemEffectRecord{
			ItemName:  item.Name,
			Game:      item.Spec.Game,
			Attribute: effect.Attribute,
			Modifier:  effect.Modifier,
		}
		if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(effectRecord); result.Error != nil {
			logger.Error(result.Error, "Failed to insert item effect record")
			return ctrl.Result{}, result.Error
		}
	}

	return ctrl.Result{}, nil
}

func (r *ItemReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Item{}).
		Complete(r)
}
