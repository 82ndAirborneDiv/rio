package telemetry

import (
	"context"

	"github.com/rancher/rio/pkg/constants"

	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "mixer",
		FeatureSpec: v1.FeatureSpec{
			Description: "Istio Mixer telemetry",
			Answers: map[string]string{
				"GRAFANA_USERNAME": "admin",
				"GRAFANA_PASSWORD": "admin",
			},
			Requires: []string{
				"prometheus",
			},
			Enabled: !constants.DisableMixer,
		},
		SystemStacks: []*systemstack.SystemStack{
			systemstack.NewStack(apply, rContext.Namespace, "istio-telemetry", true),
		},
		FixedAnswers: map[string]string{
			"NAMESPACE": rContext.Namespace,
			"TAG":       "1.1.3",
		},
	}

	return feature.Register()
}
