package openstack_utlities


import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"log"
	"strings"
)

func GetNetworkClient() *gophercloud.ServiceClient {
	opts, err := openstack.AuthOptionsFromEnv()

	if err != nil {
		log.Panic(err)
	}

	// Create a provider to authenticate all services we use
	provider, err := openstack.AuthenticatedClient(opts)

	if err != nil {
		log.Panic(err)
	}

	// Use the provider to authenticate a new client
	client, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{
		Region: "RegionOne",
	})

	if err != nil {
		log.Panic(err)
	}
	return client
}

func GetNetworkByName(client *gophercloud.ServiceClient, name string) []networks.Network {
	listOpts := networks.ListOpts{
		Name: name,
	}

	allPages, err := networks.List(client, listOpts).AllPages()
	if err != nil {
		log.Panic(err)
	}

	allNetworks, err := networks.ExtractNetworks(allPages)
	if err != nil {
		log.Panic(err)
	}

	return allNetworks
}

func GetTenantOpsNet(client *gophercloud.ServiceClient) networks.Network {
	var n networks.Network
	networks := GetNetworkByName(client, "")
	for _, network := range networks {
		if strings.Contains(network.Name, "ops-net") {
			n = network
		}
	}
	return n
}