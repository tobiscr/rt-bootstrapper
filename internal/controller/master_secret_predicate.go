package controller

import (
	"bytes"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.TypedPredicate[client.Object] = &masterSecret{}

type masterSecret struct {
	log *slog.Logger
	types.NamespacedName
}

// Create - handles the case of master secret creation
func (p masterSecret) Create(e event.TypedCreateEvent[client.Object]) bool {

	secretName := e.Object.GetName()
	secretNamespace := e.Object.GetNamespace()

	args := []any{
		"secret-name", secretName,
		"secret-namespace", secretNamespace,
	}

	accept := secretName == p.Name && secretNamespace == p.Namespace
	p.log.With(args...).Debug("incomming create secret event", "accept", accept)

	return accept
}

// Delete - omit event
func (p masterSecret) Delete(e event.TypedDeleteEvent[client.Object]) bool {
	return false
}

// Update - handles the case of an update when a secret has the same name
// as the master secret and '.dockerconfigjson' entry changed
func (p masterSecret) Update(e event.TypedUpdateEvent[client.Object]) bool {

	secretNew := e.ObjectNew.(*corev1.Secret)
	secretOld := e.ObjectOld.(*corev1.Secret)

	valOld := secretOld.Data[corev1.DockerConfigJsonKey]
	valNew := secretNew.Data[corev1.DockerConfigJsonKey]

	args := []any{
		"secret-name", secretNew.Name,
		"secret-namespace", secretNew.Namespace,
	}

	idMatch := p.Name == secretNew.Name
	accept := idMatch && !bytes.Equal(valNew, valOld)

	p.log.With(args...).Debug("incomming update secret event", "accept", accept)
	return accept
}

// Generic - omit event
func (p masterSecret) Generic(e event.TypedGenericEvent[client.Object]) bool {
	return false
}
