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

package sshkey

import (
	"context"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/pkg/errors"
	"github.com/yaskoo/provider-hetzner/internal/controller/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/hetznercloud/hcloud-go/hcloud"

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
	errNotSSHKey    = "managed resource is not a SSHKey custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

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

// Setup adds a controller that reconciles SSHKey managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.SSHKeyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.SSHKeyGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: hCloudService}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.SSHKey{}).
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
	cr, ok := mg.(*v1alpha1.SSHKey)
	if !ok {
		return nil, errors.New(errNotSSHKey)
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
	cr, ok := mg.(*v1alpha1.SSHKey)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSSHKey)
	}

	key, _, err := c.service.client.SSHKey.GetByName(ctx, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	exists := key != nil && key.ID > 0
	if exists {
		cr.Status.SetConditions(xpv1.Available())

		cr.Status.AtProvider.Id = key.ID
		cr.Status.AtProvider.Created = &metav1.Time{Time: key.Created}
		cr.Status.AtProvider.Fingerprint = key.Fingerprint
	}

	return managed.ExternalObservation{
		ResourceExists:    exists,
		ResourceUpToDate:  exists && util.LabelsUpToDate(cr.Spec.ForProvider.Labels, key.Labels),
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.SSHKey)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSSHKey)
	}

	_, _, err := c.service.client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      meta.GetExternalName(cr),
		PublicKey: cr.Spec.ForProvider.PublicKey,
		Labels:    cr.Spec.ForProvider.Labels,
	})

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.SSHKey)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSSHKey)
	}

	_, _, err := c.service.client.SSHKey.Update(ctx, &hcloud.SSHKey{ID: cr.Status.AtProvider.Id}, hcloud.SSHKeyUpdateOpts{
		Name:   meta.GetExternalName(cr),
		Labels: cr.Spec.ForProvider.Labels,
	})

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, err
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.SSHKey)
	if !ok {
		return errors.New(errNotSSHKey)
	}

	_, err := c.service.client.SSHKey.Delete(ctx, &hcloud.SSHKey{ID: cr.Status.AtProvider.Id})
	return err
}
