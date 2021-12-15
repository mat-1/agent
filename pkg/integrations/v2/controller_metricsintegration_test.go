package integrations

import (
	"testing"

	"github.com/go-kit/log"
	"github.com/grafana/agent/pkg/util"
	"github.com/prometheus/common/model"
	prom_config "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/http"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/stretchr/testify/require"
)

//
// Tests for controller's utilization of the MetricsIntegration interface.
//

func Test_controller_MetricsIntegration_Targets(t *testing.T) {
	integrationWithTarget := func(targetName string) Integration {
		return mockMetricsIntegration{
			Integration: NoOpIntegration,
			TargetsFunc: func(prefix string) []*targetgroup.Group {
				return []*targetgroup.Group{{
					Targets: []model.LabelSet{{model.AddressLabel: model.LabelValue(targetName)}},
				}}
			},
			ScrapeConfigsFunc: func(c discovery.Configs) []*prom_config.ScrapeConfig { return nil },
		}
	}

	integrations := []Config{
		mockConfigNameTuple(t, "a", "instanceA").WithNewIntegrationFunc(func(l log.Logger, g Globals) (Integration, error) {
			return integrationWithTarget("a"), nil
		}),
		mockConfigNameTuple(t, "b", "instanceB").WithNewIntegrationFunc(func(l log.Logger, g Globals) (Integration, error) {
			return integrationWithTarget("b"), nil
		}),
	}

	t.Run("All", func(t *testing.T) {
		ctrl, err := newController(
			util.TestLogger(t),
			controllerConfig(integrations),
			Globals{},
		)
		require.NoError(t, err)
		_ = newSyncController(t, ctrl)

		result := ctrl.Targets("/", TargetOptions{})
		expect := []*targetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
			{Targets: []model.LabelSet{{model.AddressLabel: "b"}}},
		}
		require.Equal(t, expect, result)
	})

	t.Run("All by Integration", func(t *testing.T) {
		ctrl, err := newController(
			util.TestLogger(t),
			controllerConfig(integrations),
			Globals{},
		)
		require.NoError(t, err)
		_ = newSyncController(t, ctrl)

		result := ctrl.Targets("/", TargetOptions{
			Integrations: []string{"a", "b"},
		})
		expect := []*targetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
			{Targets: []model.LabelSet{{model.AddressLabel: "b"}}},
		}
		require.Equal(t, expect, result)
	})

	t.Run("Specific Integration", func(t *testing.T) {
		ctrl, err := newController(
			util.TestLogger(t),
			controllerConfig(integrations),
			Globals{},
		)
		require.NoError(t, err)
		_ = newSyncController(t, ctrl)

		result := ctrl.Targets("/", TargetOptions{
			Integrations: []string{"a"},
		})
		expect := []*targetGroup{
			{Targets: []model.LabelSet{{model.AddressLabel: "a"}}},
		}
		require.Equal(t, expect, result)
	})
}

func Test_controller_MetricsIntegration_ScrapeConfig(t *testing.T) {
	integrationWithTarget := func(targetName string) Integration {
		return mockMetricsIntegration{
			Integration: NoOpIntegration,
			ScrapeConfigsFunc: func(c discovery.Configs) []*prom_config.ScrapeConfig {
				return []*prom_config.ScrapeConfig{{JobName: targetName}}
			},
		}
	}

	integrations := []Config{
		mockConfigNameTuple(t, "a", "instanceA").WithNewIntegrationFunc(func(l log.Logger, g Globals) (Integration, error) {
			return integrationWithTarget("a"), nil
		}),
		mockConfigNameTuple(t, "b", "instanceB").WithNewIntegrationFunc(func(l log.Logger, g Globals) (Integration, error) {
			return integrationWithTarget("b"), nil
		}),
	}

	ctrl, err := newController(
		util.TestLogger(t),
		controllerConfig(integrations),
		Globals{},
	)
	require.NoError(t, err)
	_ = newSyncController(t, ctrl)

	result := ctrl.ScrapeConfigs("/", &http.DefaultSDConfig)
	expect := []*prom_config.ScrapeConfig{
		{JobName: "a"},
		{JobName: "b"},
	}
	require.Equal(t, expect, result)
}

//
// Tests for controller's utilization of the MetricsIntegration interface.
//

type mockMetricsIntegration struct {
	Integration
	TargetsFunc       func(prefix string) []*targetgroup.Group
	ScrapeConfigsFunc func(discovery.Configs) []*prom_config.ScrapeConfig
}

func (m mockMetricsIntegration) Targets(prefix string) []*targetgroup.Group {
	return m.TargetsFunc(prefix)
}

func (m mockMetricsIntegration) ScrapeConfigs(cfgs discovery.Configs) []*prom_config.ScrapeConfig {
	return m.ScrapeConfigsFunc(cfgs)
}
