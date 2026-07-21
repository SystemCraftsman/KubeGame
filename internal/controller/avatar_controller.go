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

const avatarFinalizer = "kubegame.systemcraftsman.com/finalizer"

type AvatarReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=avatars,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=avatars/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=avatars/finalizers,verbs=update

func (r *AvatarReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var avatar v1alpha1.Avatar
	if err := r.Get(ctx, req.NamespacedName, &avatar); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Avatar resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Avatar resource")
		return ctrl.Result{}, err
	}

	var game v1alpha1.Game
	if err := r.Get(ctx, types.NamespacedName{Name: avatar.Spec.Game, Namespace: avatar.Namespace}, &game); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Game resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Game resource")
		return ctrl.Result{}, err
	}

	if !avatar.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&avatar, avatarFinalizer) {
			if game.Status.Ready {
				db, err := getGameDB(ctx, r.Client, &game)
				if err == nil {
					db.Where("avatar_name = ?", avatar.Name).Delete(&persistence.CustomizationOption{})
					db.Where("avatar_name = ?", avatar.Name).Delete(&persistence.CustomizationTypeRecord{})
					db.Where("avatar_name = ?", avatar.Name).Delete(&persistence.AchievementType{})
					db.Where("avatar_name = ?", avatar.Name).Delete(&persistence.InventoryType{})
					db.Where("avatar_name = ?", avatar.Name).Delete(&persistence.AttributeType{})
					db.Where("name = ?", avatar.Name).Delete(&persistence.AvatarType{})
					logger.Info("Deleted avatar type records from database", "avatar", avatar.Name)
				}
			}

			controllerutil.RemoveFinalizer(&avatar, avatarFinalizer)
			if err := r.Update(ctx, &avatar); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&avatar, avatarFinalizer) {
		controllerutil.AddFinalizer(&avatar, avatarFinalizer)
		if err := r.Update(ctx, &avatar); err != nil {
			return ctrl.Result{}, err
		}
	}

	if _, updated, err := ensureLabels(ctx, r.Client, &avatar, map[string]string{
		labelGame: avatar.Spec.Game,
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

	if err := persistence.RunMigrations(db, persistence.AvatarModels()...); err != nil {
		logger.Error(err, "Failed to run avatar migrations")
		return ctrl.Result{}, err
	}

	avatarRecord := &persistence.AvatarType{
		Name:        avatar.Name,
		Game:        avatar.Spec.Game,
		Type:        avatar.Spec.Type,
		Description: avatar.Spec.Description,
	}

	if result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(avatarRecord); result.Error != nil {
		logger.Error(result.Error, "Failed to insert the record into the AvatarType table")
		return ctrl.Result{}, result.Error
	}

	for _, at := range avatar.Spec.AttributeTypes {
		record := &persistence.AttributeType{
			AvatarName: avatar.Name,
			Name:       at.Name,
			ValueType:  at.ValueType,
		}
		if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(record); result.Error != nil {
			logger.Error(result.Error, "Failed to insert attribute type record")
			return ctrl.Result{}, result.Error
		}
	}

	for _, it := range avatar.Spec.InventoryTypes {
		record := &persistence.InventoryType{
			AvatarName: avatar.Name,
			Name:       it.Name,
			Category:   it.Category,
		}
		if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(record); result.Error != nil {
			logger.Error(result.Error, "Failed to insert inventory type record")
			return ctrl.Result{}, result.Error
		}
	}

	for _, ach := range avatar.Spec.AchievementTypes {
		record := &persistence.AchievementType{
			AvatarName:  avatar.Name,
			Name:        ach.Name,
			Description: ach.Description,
		}
		if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(record); result.Error != nil {
			logger.Error(result.Error, "Failed to insert achievement type record")
			return ctrl.Result{}, result.Error
		}
	}

	for _, ct := range avatar.Spec.CustomizationTypes {
		record := &persistence.CustomizationTypeRecord{
			AvatarName: avatar.Name,
			Name:       ct.Name,
		}
		if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(record); result.Error != nil {
			logger.Error(result.Error, "Failed to insert customization type record")
			return ctrl.Result{}, result.Error
		}

		for _, opt := range ct.Options {
			optRecord := &persistence.CustomizationOption{
				CustomizationName: ct.Name,
				AvatarName:        avatar.Name,
				Value:             opt,
			}
			if result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(optRecord); result.Error != nil {
				logger.Error(result.Error, "Failed to insert customization option record")
				return ctrl.Result{}, result.Error
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *AvatarReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Avatar{}).
		Complete(r)
}
