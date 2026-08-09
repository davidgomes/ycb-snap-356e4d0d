// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package discoverychain

import (
	"testing"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/stretchr/testify/require"
)

func TestHTTPServiceDefaultForDestination(t *testing.T) {
	t.Parallel()

	t.Run("nil destination", func(t *testing.T) {
		require.Nil(t, httpServiceDefaultForDestination(nil))
	})

	t.Run("empty service", func(t *testing.T) {
		require.Nil(t, httpServiceDefaultForDestination(&structs.ServiceRouteDestination{
			Namespace: "other",
			Partition: "part",
		}))
	})

	t.Run("same tenancy", func(t *testing.T) {
		got := httpServiceDefaultForDestination(&structs.ServiceRouteDestination{
			Service: "bar",
		})
		require.NotNil(t, got)
		require.Equal(t, structs.ServiceDefaults, got.Kind)
		require.Equal(t, "bar", got.Name)
		require.Equal(t, "http", got.Protocol)
		require.Equal(t, structs.NewEnterpriseMetaWithPartition("", ""), got.EnterpriseMeta)
	})

	t.Run("different namespace and partition", func(t *testing.T) {
		got := httpServiceDefaultForDestination(&structs.ServiceRouteDestination{
			Service:   "bar",
			Namespace: "other-ns",
			Partition: "other-ap",
		})
		require.NotNil(t, got)
		require.Equal(t, structs.ServiceDefaults, got.Kind)
		require.Equal(t, "bar", got.Name)
		require.Equal(t, "http", got.Protocol)
		require.Equal(t, structs.NewEnterpriseMetaWithPartition("other-ap", "other-ns"), got.EnterpriseMeta)
	})
}

func TestHTTPRouteToDiscoveryChain_ComposedDestinationGetsHTTPDefaults(t *testing.T) {
	t.Parallel()

	route := structs.HTTPRouteConfigEntry{
		Kind: structs.HTTPRoute,
		Name: "http-route",
		Rules: []structs.HTTPRouteRule{{
			Services: []structs.HTTPService{{
				Name: "foo",
			}},
		}},
	}

	t.Run("composed destination without service-defaults", func(t *testing.T) {
		serviceRouters := map[structs.ServiceName][]*structs.ServiceRoute{
			structs.NewServiceName("foo", nil): {{
				Match: &structs.ServiceRouteMatch{
					HTTP: &structs.ServiceRouteHTTPMatch{PathPrefix: "/v1"},
				},
				Destination: &structs.ServiceRouteDestination{
					Service: "bar",
				},
			}},
		}

		router, _, defaults := httpRouteToDiscoveryChain(route, serviceRouters)
		require.NotNil(t, router)
		requireHTTPServiceDefault(t, defaults, "foo")
		requireHTTPServiceDefault(t, defaults, "bar")

		var foundComposed bool
		for _, r := range router.Routes {
			if r.Destination != nil && r.Destination.Service == "bar" {
				foundComposed = true
			}
		}
		require.True(t, foundComposed, "expected composed route to target bar")
	})

	t.Run("composed destination in another namespace and partition", func(t *testing.T) {
		serviceRouters := map[structs.ServiceName][]*structs.ServiceRoute{
			structs.NewServiceName("foo", nil): {{
				Match: &structs.ServiceRouteMatch{
					HTTP: &structs.ServiceRouteHTTPMatch{PathPrefix: "/v1"},
				},
				Destination: &structs.ServiceRouteDestination{
					Service:   "bar",
					Namespace: "other-ns",
					Partition: "other-ap",
				},
			}},
		}

		_, _, defaults := httpRouteToDiscoveryChain(route, serviceRouters)
		got := requireHTTPServiceDefault(t, defaults, "bar")
		require.Equal(t, structs.NewEnterpriseMetaWithPartition("other-ap", "other-ns"), got.EnterpriseMeta)
	})
}

func requireHTTPServiceDefault(t *testing.T, defaults []*structs.ServiceConfigEntry, name string) *structs.ServiceConfigEntry {
	t.Helper()
	for _, d := range defaults {
		if d != nil && d.Name == name && d.Kind == structs.ServiceDefaults && d.Protocol == "http" {
			return d
		}
	}
	t.Fatalf("missing http service-defaults for %q in %#v", name, defaults)
	return nil
}
