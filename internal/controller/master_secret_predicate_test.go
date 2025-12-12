package controller

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func Test_masterSecret_Create(t *testing.T) {
	type predicateResult = predicate.TypedPredicate[client.Object]

	const (
		masterSecretNamespace = "master-secret-namespace"
		masterSecretName      = "master-secret"
	)

	var (
		newTypedCreateEvent = func(name, namespace string) event.TypedCreateEvent[client.Object] {
			return event.TypedCreateEvent[client.Object]{
				Object: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
					},
				},
			}

		}

		newTestMasterSecretPredicate = func() predicateResult {
			return &masterSecret{
				log: slog.Default(),
				NamespacedName: types.NamespacedName{
					Name:      masterSecretName,
					Namespace: masterSecretNamespace,
				},
			}
		}
	)

	tcs := []struct {
		name     string
		expected bool
		p        predicate.TypedPredicate[client.Object]
		e        event.TypedCreateEvent[client.Object]
	}{
		{
			name:     "create master-secret in master-secret-namespace",
			expected: true,
			p:        newTestMasterSecretPredicate(),
			e:        newTypedCreateEvent(masterSecretName, masterSecretNamespace),
		},
		{
			name:     "create test secret in test namespace",
			expected: false,
			p:        newTestMasterSecretPredicate(),
			e:        newTypedCreateEvent("test", "test-namespace"),
		},
		{
			name:     "create test secret in master-secret-namespace",
			expected: false,
			p:        newTestMasterSecretPredicate(),
			e:        newTypedCreateEvent("test", masterSecretNamespace),
		},
		{
			name:     "create master-secret in test namespace",
			expected: false,
			p:        newTestMasterSecretPredicate(),
			e:        newTypedCreateEvent(masterSecretName, "test-namespace"),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.p.Create(tc.e)
			assert.Equal(t, tc.expected, actual)
		})
	}

}
