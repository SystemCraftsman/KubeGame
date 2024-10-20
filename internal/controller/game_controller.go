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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"systemcraftsman.com/kubegame/api/v1alpha1"
)

// GameReconciler reconciles a Game object
type GameReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=games,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=games/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubegame.systemcraftsman.com,resources=games/finalizers,verbs=update

func (r *GameReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Game resource
	var game v1alpha1.Game
	if err := r.Get(ctx, req.NamespacedName, &game); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Game resource not found, might be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Game resource")
		return ctrl.Result{}, err
	}

	// Check if the resource is marked for deletion
	if !game.DeletionTimestamp.IsZero() {
		logger.Info("Handling deletion")
		// handle deletion logic
		return ctrl.Result{}, nil
	}

	// Example of status update logic
	game.Status.Ready = true
	if err := r.Status().Update(ctx, &game); err != nil {
		logger.Error(err, "Failed to update Game status")
		return ctrl.Result{}, err
	}

	// Handle creation or updates for dependent resources, like Services or Deployments
	// For example, you might want to ensure a Deployment exists for the game
	// Check if the Deployment already exists
	var deployment appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: game.Name + "-postgres", Namespace: game.Namespace}, &deployment); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating a new Deployment for Game")
			deployment = *getPostgresDeployment(&game)
			if err := ctrl.SetControllerReference(&game, &deployment, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			err := r.Create(ctx, &deployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		logger.Error(err, "Failed to get Deployment for Game")
		return ctrl.Result{}, err
	}

	// Return success if nothing needs to be done
	return ctrl.Result{}, nil
}

func getPostgresDeployment(game *v1alpha1.Game) *appsv1.Deployment {
	labels := map[string]string{
		"app":  "postgres",
		"game": game.Name,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      game.Name + "-postgres",
			Namespace: game.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1), // Define how many replicas you need
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgres",
							Image: "postgres:13", // PostgreSQL version
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 5432, // PostgreSQL default port
									Name:          "postgres",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "POSTGRES_DB",
									Value: "postres",
								},
								{
									Name:  "POSTGRES_USER",
									Value: game.Spec.Database.Username,
								},
								{
									Name:  "POSTGRES_PASSWORD",
									Value: game.Spec.Database.Password,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "postgres-storage",
									MountPath: "/var/lib/postgresql/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "postgres-storage",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
	return deployment
}

func int32Ptr(i int32) *int32 {
	return &i
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Game{}).
		Complete(r)
}
