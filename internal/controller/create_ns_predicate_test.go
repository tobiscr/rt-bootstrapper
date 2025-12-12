package controller

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func Test_createNsPredicate(t *testing.T) {
	type predicateResult = predicate.TypedPredicate[client.Object]

	const masterSecretNamespace = "master-secret-namespace"

	var (
		masterNsCreateEvent = event.TypedCreateEvent[client.Object]{
			Object: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: masterSecretNamespace,
				},
			},
		}
		testNsCreateEvent = event.TypedCreateEvent[client.Object]{
			Object: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
				},
			},
		}

		newTestCreateNsPredicate = func() predicateResult {
			return &createNsPredicate{
				log:                      slog.Default(),
				masterSecretNamspaceName: masterSecretNamespace,
			}
		}
	)

	tcs := []struct {
		name      string
		predicate predicate.TypedPredicate[client.Object]
		e         event.TypedCreateEvent[client.Object]
		expected  bool
	}{
		{
			name:      "is master namespace",
			predicate: newTestCreateNsPredicate(),
			e:         masterNsCreateEvent,
			expected:  false,
		},
		{
			name:      "is not master namespace",
			predicate: newTestCreateNsPredicate(),
			e:         testNsCreateEvent,
			expected:  true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.predicate.Create(tc.e)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
