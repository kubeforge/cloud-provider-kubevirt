package kubevirt

import (
	"bytes"
	"fmt"
	"io"

	"github.com/golang/glog"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/cloudprovider"
	"k8s.io/kubernetes/pkg/controller"
	"kubevirt.io/kubevirt/pkg/kubecli"
)

const (
	// ProviderName is the name of the kubevirt provider
	ProviderName = "kubevirt"
)

type cloud struct {
	namespace string
	kubevirt  kubecli.KubevirtClient
}

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName, kubevirtCloudProviderFactory)
}

func kubevirtCloudProviderFactory(config io.Reader) (cloudprovider.Interface, error) {
	if config == nil {
		return nil, fmt.Errorf("No %s cloud provider config file given", ProviderName)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(config)
	clientConfig, err := clientcmd.NewClientConfigFromBytes(buf.Bytes())
	if err != nil {
		return nil, err
	}
	kubevirtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		glog.Errorf("Failed to create KubeVirt client: %v", err)
		return nil, err
	}
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		glog.Errorf("Could not find namespace in client config: %v", err)
		return nil, err
	}
	return &cloud{
		namespace: namespace,
		kubevirt:  kubevirtClient,
	}, nil
}

// Initialize provides the cloud with a kubernetes client builder and may spawn goroutines
// to perform housekeeping activities within the cloud provider.
func (c *cloud) Initialize(clientBuilder controller.ControllerClientBuilder) {}

// LoadBalancer returns a balancer interface. Also returns true if the interface is supported, false otherwise.
func (c *cloud) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	return &loadbalancer{
		namespace: c.namespace,
		kubevirt:  c.kubevirt,
	}, true
}

// Instances returns an instances interface. Also returns true if the interface is supported, false otherwise.
func (c *cloud) Instances() (cloudprovider.Instances, bool) {
	return &instances{
		namespace: c.namespace,
		kubevirt:  c.kubevirt,
	}, true
}

// Zones returns a zones interface. Also returns true if the interface is supported, false otherwise.
func (c *cloud) Zones() (cloudprovider.Zones, bool) {
	return &zones{
		namespace: c.namespace,
		kubevirt:  c.kubevirt,
	}, true
}

// Clusters returns a clusters interface.  Also returns true if the interface is supported, false otherwise.
func (c *cloud) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// Routes returns a routes interface along with whether the interface is supported.
func (c *cloud) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

// ProviderName returns the cloud provider ID.
func (c *cloud) ProviderName() string {
	return ProviderName
}

// HasClusterID returns true if a ClusterID is required and set
func (c *cloud) HasClusterID() bool {
	return true
}
