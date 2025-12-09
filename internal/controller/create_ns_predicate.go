package controller

import (
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ predicate.TypedPredicate[client.Object] = &createNsPredicate{}

type createNsPredicate struct {
	log                      *slog.Logger
	masterSecretNamspaceName string
}

func (p createNsPredicate) isMasterSecretNamespace(namespaceName string) bool {
	return p.masterSecretNamspaceName == namespaceName
}

// Create - handles the case of namespace creation (omits events comming from
// the master secret namespace)
func (p createNsPredicate) Create(e event.TypedCreateEvent[client.Object]) bool {
	_, accept := e.Object.(*corev1.Namespace)

	accept = accept && !p.isMasterSecretNamespace(e.Object.GetName())

	args := []any{
		"accept", accept,
		"name", e.Object.GetName(),
	}

	p.log.Debug("incomming create ns event", args...)
	return accept
}

// Delete - omit event
func (p createNsPredicate) Delete(event.TypedDeleteEvent[client.Object]) bool {
	return false
}

// Update - omit event
func (p createNsPredicate) Update(event.TypedUpdateEvent[client.Object]) bool {
	return false
}

// Generic - omit event
func (p createNsPredicate) Generic(event.TypedGenericEvent[client.Object]) bool {
	return false
}
