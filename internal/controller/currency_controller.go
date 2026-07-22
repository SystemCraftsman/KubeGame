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

const currencyFinalizer = "kubegame.systemcraftsman.com/finalizer"

type CurrencyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=currencies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=currencies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=currencies/finalizers,verbs=update

func (r *CurrencyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var currency v1alpha1.Currency
	if err := r.Get(ctx, req.NamespacedName, &currency); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !currency.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&currency, currencyFinalizer) {
			var game v1alpha1.Game
			if err := r.Get(ctx, types.NamespacedName{Name: currency.Spec.Game, Namespace: currency.Namespace}, &game); err == nil && game.Status.Ready {
				db, err := getGameDB(ctx, r.Client, &game)
				if err == nil {
					db.Where("name = ? AND game = ?", currency.Name, currency.Spec.Game).Delete(&persistence.CurrencyDefinition{})
					logger.Info("Deleted currency record from database", "currency", currency.Name)
				}
			}

			controllerutil.RemoveFinalizer(&currency, currencyFinalizer)
			if err := r.Update(ctx, &currency); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	var game v1alpha1.Game
	if err := r.Get(ctx, types.NamespacedName{Name: currency.Spec.Game, Namespace: currency.Namespace}, &game); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Game resource not found", "game", currency.Spec.Game)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(&currency, currencyFinalizer) {
		controllerutil.AddFinalizer(&currency, currencyFinalizer)
		if err := r.Update(ctx, &currency); err != nil {
			return ctrl.Result{}, err
		}
	}

	if _, updated, err := ensureLabels(ctx, r.Client, &currency, map[string]string{
		labelGame: currency.Spec.Game,
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

	if err := persistence.RunMigrations(db, persistence.CurrencyModels()...); err != nil {
		logger.Error(err, "Failed to run currency migrations")
		return ctrl.Result{}, err
	}

	record := &persistence.CurrencyDefinition{
		Name:           currency.Name,
		Game:           currency.Spec.Game,
		Symbol:         currency.Spec.Symbol,
		Tradeable:      currency.Spec.Tradeable,
		MaxBalance:     currency.Spec.MaxBalance,
		InitialBalance: currency.Spec.InitialBalance,
	}
	if result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}, {Name: "game"}},
		DoUpdates: clause.AssignmentColumns([]string{"symbol", "tradeable", "max_balance", "initial_balance"}),
	}).Create(record); result.Error != nil {
		logger.Error(result.Error, "Failed to upsert currency definition", "currency", currency.Name)
		return ctrl.Result{}, result.Error
	}

	return ctrl.Result{}, nil
}

func (r *CurrencyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Currency{}).
		Complete(r)
}
