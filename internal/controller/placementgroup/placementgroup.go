/*
Copyright 2022 The Crossplane Authors.

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

package placementgroup

import (
	"context"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/hetznercloud/hcloud-go/hcloud"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/yaskoo/provider-hetzner/apis/cloud/v1alpha1"
	apisv1alpha1 "github.com/yaskoo/provider-hetzner/apis/v1alpha1"
	"github.com/yaskoo/provider-hetzner/internal/controller/features"
)

const (
	errNotPlacementGroup = "managed resource is not a PlacementGroup custom resource"
	errTrackPCUsage      = "cannot track ProviderConfig usage"
	errGetPC             = "cannot get ProviderConfig"
	errGetCreds          = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// A HCloudService is the interface to the Hetzner cloud API.
type HCloudService struct {
	client *hcloud.Client
}

var (
	hCloudService = func(creds []byte) (*HCloudService, error) {
		return &HCloudService{
			client: hcloud.NewClient(hcloud.WithToken(string(creds))),
		}, nil
	}
)

// Setup adds a controller that reconciles PlacementGroup managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.PlacementGroupGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PlacementGroupGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: hCloudService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.PlacementGroup{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(creds []byte) (*HCloudService, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.PlacementGroup)
	if !ok {
		return nil, errors.New(errNotPlacementGroup)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(data)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *HCloudService
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.PlacementGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPlacementGroup)
	}

	pg, _, err := c.service.client.PlacementGroup.GetByName(ctx, meta.GetExternalName(cr))
	exists := pg != nil && pg.ID > 0

	var upToDate bool
	if exists {
		cr.Status.SetConditions(xpv1.Available())
		cr.Status.AtProvider.Id = pg.ID
		cr.Status.AtProvider.Created = &metav1.Time{Time: pg.Created}

		switch {
		case cr.Spec.ForProvider.Labels == nil && len(pg.Labels) > 0:
			upToDate = false
		case cr.Spec.ForProvider.Labels == nil && len(pg.Labels) == 0:
			upToDate = true
		case reflect.DeepEqual(*cr.Spec.ForProvider.Labels, pg.Labels):
			upToDate = true
		case !reflect.DeepEqual(*cr.Spec.ForProvider.Labels, pg.Labels):
			upToDate = false
		}
	}

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: exists,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: exists && upToDate,

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.PlacementGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPlacementGroup)
	}

	var labels map[string]string
	if cr.Spec.ForProvider.Labels != nil {
		labels = *cr.Spec.ForProvider.Labels
	}

	_, _, err := c.service.client.PlacementGroup.Create(ctx, hcloud.PlacementGroupCreateOpts{
		Name:   meta.GetExternalName(cr),
		Type:   hcloud.PlacementGroupType(cr.Spec.ForProvider.Type),
		Labels: labels,
	})

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.PlacementGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPlacementGroup)
	}

	var labels map[string]string
	if cr.Spec.ForProvider.Labels == nil {
		labels = make(map[string]string)
	} else {
		labels = *cr.Spec.ForProvider.Labels
	}

	pg := &hcloud.PlacementGroup{
		ID: cr.Status.AtProvider.Id,
	}

	_, _, err := c.service.client.PlacementGroup.Update(ctx, pg, hcloud.PlacementGroupUpdateOpts{
		Labels: labels,
	})
	return managed.ExternalUpdate{}, err
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.PlacementGroup)
	if !ok {
		return errors.New(errNotPlacementGroup)
	}

	_, err := c.service.client.PlacementGroup.Delete(ctx, &hcloud.PlacementGroup{ID: cr.Status.AtProvider.Id})

	return err
}
