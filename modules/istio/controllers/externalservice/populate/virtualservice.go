package populate

import (
	"github.com/rancher/rio/modules/istio/controllers/service/populate"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/wrangler/pkg/objectset"
)

func VirtualServiceForExternalService(namespace string, es *riov1.ExternalService, serviceSet *serviceset.ServiceSet, clusterDomain *v1.ClusterDomain,
	svc *riov1.Service, os *objectset.ObjectSet) {

	dests := populate.DestsForService(serviceSet)
	serviceVS := populate.VirtualServiceFromSpec(true, namespace, svc.Name, svc.Namespace, clusterDomain, svc, dests...)

	// override host match with external service
	serviceVS.Name = es.Name
	serviceVS.Namespace = es.Namespace
	os.Add(serviceVS)
}
