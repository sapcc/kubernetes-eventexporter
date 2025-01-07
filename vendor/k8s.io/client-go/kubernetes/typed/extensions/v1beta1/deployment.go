/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	context "context"
	fmt "fmt"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	applyconfigurationsextensionsv1beta1 "k8s.io/client-go/applyconfigurations/extensions/v1beta1"
	gentype "k8s.io/client-go/gentype"
	scheme "k8s.io/client-go/kubernetes/scheme"
	apply "k8s.io/client-go/util/apply"
)

// DeploymentsGetter has a method to return a DeploymentInterface.
// A group's client should implement this interface.
type DeploymentsGetter interface {
	Deployments(namespace string) DeploymentInterface
}

// DeploymentInterface has methods to work with Deployment resources.
type DeploymentInterface interface {
	Create(ctx context.Context, deployment *extensionsv1beta1.Deployment, opts v1.CreateOptions) (*extensionsv1beta1.Deployment, error)
	Update(ctx context.Context, deployment *extensionsv1beta1.Deployment, opts v1.UpdateOptions) (*extensionsv1beta1.Deployment, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, deployment *extensionsv1beta1.Deployment, opts v1.UpdateOptions) (*extensionsv1beta1.Deployment, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*extensionsv1beta1.Deployment, error)
	List(ctx context.Context, opts v1.ListOptions) (*extensionsv1beta1.DeploymentList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *extensionsv1beta1.Deployment, err error)
	Apply(ctx context.Context, deployment *applyconfigurationsextensionsv1beta1.DeploymentApplyConfiguration, opts v1.ApplyOptions) (result *extensionsv1beta1.Deployment, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, deployment *applyconfigurationsextensionsv1beta1.DeploymentApplyConfiguration, opts v1.ApplyOptions) (result *extensionsv1beta1.Deployment, err error)
	GetScale(ctx context.Context, deploymentName string, options v1.GetOptions) (*extensionsv1beta1.Scale, error)
	UpdateScale(ctx context.Context, deploymentName string, scale *extensionsv1beta1.Scale, opts v1.UpdateOptions) (*extensionsv1beta1.Scale, error)
	ApplyScale(ctx context.Context, deploymentName string, scale *applyconfigurationsextensionsv1beta1.ScaleApplyConfiguration, opts v1.ApplyOptions) (*extensionsv1beta1.Scale, error)

	DeploymentExpansion
}

// deployments implements DeploymentInterface
type deployments struct {
	*gentype.ClientWithListAndApply[*extensionsv1beta1.Deployment, *extensionsv1beta1.DeploymentList, *applyconfigurationsextensionsv1beta1.DeploymentApplyConfiguration]
}

// newDeployments returns a Deployments
func newDeployments(c *ExtensionsV1beta1Client, namespace string) *deployments {
	return &deployments{
		gentype.NewClientWithListAndApply[*extensionsv1beta1.Deployment, *extensionsv1beta1.DeploymentList, *applyconfigurationsextensionsv1beta1.DeploymentApplyConfiguration](
			"deployments",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *extensionsv1beta1.Deployment { return &extensionsv1beta1.Deployment{} },
			func() *extensionsv1beta1.DeploymentList { return &extensionsv1beta1.DeploymentList{} },
			gentype.PrefersProtobuf[*extensionsv1beta1.Deployment](),
		),
	}
}

// GetScale takes name of the deployment, and returns the corresponding extensionsv1beta1.Scale object, and an error if there is any.
func (c *deployments) GetScale(ctx context.Context, deploymentName string, options v1.GetOptions) (result *extensionsv1beta1.Scale, err error) {
	result = &extensionsv1beta1.Scale{}
	err = c.GetClient().Get().
		UseProtobufAsDefault().
		Namespace(c.GetNamespace()).
		Resource("deployments").
		Name(deploymentName).
		SubResource("scale").
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// UpdateScale takes the top resource name and the representation of a scale and updates it. Returns the server's representation of the scale, and an error, if there is any.
func (c *deployments) UpdateScale(ctx context.Context, deploymentName string, scale *extensionsv1beta1.Scale, opts v1.UpdateOptions) (result *extensionsv1beta1.Scale, err error) {
	result = &extensionsv1beta1.Scale{}
	err = c.GetClient().Put().
		UseProtobufAsDefault().
		Namespace(c.GetNamespace()).
		Resource("deployments").
		Name(deploymentName).
		SubResource("scale").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(scale).
		Do(ctx).
		Into(result)
	return
}

// ApplyScale takes top resource name and the apply declarative configuration for scale,
// applies it and returns the applied scale, and an error, if there is any.
func (c *deployments) ApplyScale(ctx context.Context, deploymentName string, scale *applyconfigurationsextensionsv1beta1.ScaleApplyConfiguration, opts v1.ApplyOptions) (result *extensionsv1beta1.Scale, err error) {
	if scale == nil {
		return nil, fmt.Errorf("scale provided to ApplyScale must not be nil")
	}
	patchOpts := opts.ToPatchOptions()
	request, err := apply.NewRequest(c.GetClient(), scale)
	if err != nil {
		return nil, err
	}

	result = &extensionsv1beta1.Scale{}
	err = request.
		UseProtobufAsDefault().
		Namespace(c.GetNamespace()).
		Resource("deployments").
		Name(deploymentName).
		SubResource("scale").
		VersionedParams(&patchOpts, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}
