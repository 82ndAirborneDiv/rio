package publicdomain

import (
	"context"
	"fmt"

	certmanagerapi "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/rancher/rio/modules/system/features/letsencrypt/pkg/issuers"
	projectv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	adminv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	v12 "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/generic"
	name2 "github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/objectset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContexts *types.Context) error {
	p := &publicDomainHandler{
		namespace: rContexts.Namespace,
		apply: rContexts.Apply.WithSetID("letsencrypt-publicdomain").
			WithStrictCaching().
			WithCacheTypes(rContexts.CertManager.Certmanager().V1alpha1().Certificate()),
		featureClientCache: rContexts.Global.Admin().V1().Feature().Cache(),
		publicDomains:      rContexts.Global.Admin().V1().PublicDomain(),
		publicDomainCache:  rContexts.Global.Admin().V1().PublicDomain().Cache(),
	}

	rContexts.Global.Admin().V1().PublicDomain().AddGenericHandler(ctx, "letsencrypt-handler", generic.UpdateOnChange(rContexts.Global.Admin().V1().PublicDomain().Updater(), p.onChange))
	rContexts.Global.Admin().V1().PublicDomain().OnRemove(ctx, "letsencrypt-handler", p.onRemove)
	rContexts.Global.Admin().V1().Feature().OnChange(ctx, "letsencrypt-handler", p.featureChanged)

	return nil
}

type publicDomainHandler struct {
	namespace          string
	apply              apply.Apply
	publicDomains      adminv1controller.PublicDomainController
	publicDomainCache  adminv1controller.PublicDomainCache
	featureClientCache v12.FeatureCache
}

func (p *publicDomainHandler) featureChanged(key string, feature *projectv1.Feature) (*projectv1.Feature, error) {
	if feature == nil {
		return feature, nil
	}

	if feature.Namespace != p.namespace || feature.Name != "letsencrypt" {
		return feature, nil
	}

	pds, err := p.publicDomainCache.List(p.namespace, labels.Everything())
	if err != nil {
		return feature, err
	}

	for _, pd := range pds {
		p.publicDomains.Enqueue(pd.Namespace, pd.Name)
	}

	return feature, nil
}

func (p *publicDomainHandler) onChange(key string, obj runtime.Object) (runtime.Object, error) {
	if obj == nil {
		return nil, nil
	}
	domain := obj.(*projectv1.PublicDomain)

	if domain.Spec.DisableLetsencrypt {
		return domain, nil
	}

	domain.Status.Endpoint = fmt.Sprintf("https://%s", domain.Spec.DomainName)

	feature, err := p.featureClientCache.Get(p.namespace, "letsencrypt")
	if err != nil {
		return domain, err
	}

	publicdomainType := feature.Spec.Answers[constants.PublicDomainType]
	issuerName := issuers.IssuerTypeToName[publicdomainType]

	os := objectset.NewObjectSet()

	if issuerName != "" {
		os.Add(certificateHTTP(p.namespace, domain, issuerName))
	}
	domain.Spec.SecretRef.Name = fmt.Sprintf("%s-%s", domain.Namespace, domain.Name)
	domain.Spec.SecretRef.Namespace = domain.Namespace
	domain.Status.IssuerName = issuerName

	return domain, p.apply.WithOwner(domain).Apply(os)
}

func (p *publicDomainHandler) onRemove(key string, domain *projectv1.PublicDomain) (*projectv1.PublicDomain, error) {
	if domain == nil {
		return domain, nil
	}

	if domain.Namespace != p.namespace {
		return domain, nil
	}

	return domain, p.apply.WithOwner(domain).Apply(nil)
}

func certificateHTTP(namespace string, domain *projectv1.PublicDomain, issuerName string) runtime.Object {
	name := fmt.Sprintf("%s-%s", domain.Namespace, domain.Name)
	cert := constructors.NewCertificate(namespace, name,
		certmanagerapi.Certificate{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "rio.cattle.io/v1",
						Kind:       "PublicDomain",
						Name:       domain.Name,
						UID:        domain.UID,
					},
				},
			},
			Spec: certmanagerapi.CertificateSpec{
				SecretName: name,
				IssuerRef: certmanagerapi.ObjectReference{
					Kind: "ClusterIssuer",
					Name: issuerName,
				},
				DNSNames: []string{
					domain.Spec.DomainName,
				},
				ACME: &certmanagerapi.ACMECertificateConfig{
					Config: []certmanagerapi.DomainSolverConfig{
						{
							Domains: []string{
								domain.Spec.DomainName,
							},
							SolverConfig: certmanagerapi.SolverConfig{
								HTTP01: &certmanagerapi.HTTP01SolverConfig{
									Ingress: name2.SafeConcatName(domain.Name, name2.Hex(domain.Spec.DomainName, 5)),
								},
							},
						},
					},
				},
			},
		})
	return cert
}
