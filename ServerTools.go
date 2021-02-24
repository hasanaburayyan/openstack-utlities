package openstack_utlities

import (
	"errors"
	"fmt"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"html/template"
	"log"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"os"
	"regexp"
	"strings"
)

func GetServerClient() *gophercloud.ServiceClient {
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
client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
Region: "RegionOne",
})

if err != nil {
log.Panic(err)
}
return client
}

func GetBlockStorageClient() *gophercloud.ServiceClient {
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
	client, err := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
		Region: "RegionOne",
	})

	if err != nil {
		log.Panic(err)
	}
	return client
}


func ListServersInCurrentTenant(client *gophercloud.ServiceClient, t string) {
	allServers := GetAllServers(client)


	//temp = "{{.Name}}\t\t||\t\t{{index .Image `id`}}\t\t||\t\t{{index .Flavor `id`}}\n"
	var temp string
	var tmpl *template.Template
	var err error

	if t != "" {
		temp = t
		tmpl, err = template.New("Server").Parse(temp)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Server Name \t\t||\t\t\tImage\t\t\t||\t\tFlavor\t\t\t\t||\t\tNetworks")
		for i := 0; i < 200; i ++ {
			fmt.Print("=")
		}
		fmt.Println()
	}


	for _, server := range allServers {

		if temp == "" {
			fmt.Printf("%s\t\t||\t%s\t||\t%s\t||\t", server.Name, server.Image["id"], server.Flavor["id"])
			var networks string
			for k, v := range server.Addresses {
				re := regexp.MustCompile("([0-9]{1,3}[.]?){4}")
				match := re.Find([]byte(fmt.Sprintf("%s", v)))
				networks += fmt.Sprintf("%s : %s ", k, match)
			}
			fmt.Printf("{ %s }\n", networks)
		} else {
			err = tmpl.Execute(os.Stdout, server)
			if err != nil {
				log.Fatal(err)
			}
		}


	}
}

func GetAllServers(client *gophercloud.ServiceClient) []servers.Server {
	// Options for listing servers
	listOpts := servers.ListOpts{
		AllTenants: false,
	}

	// Get all pages of servers
	allPages, err := servers.List(client, listOpts).AllPages()

	if err != nil {
		log.Panic(err)
	}

	// Extract all servers from pages
	allServers, err := servers.ExtractServers(allPages)

	if err != nil {
		log.Panic(err)
	}

	return allServers
}

func FindServersByName(client *gophercloud.ServiceClient, name string) []servers.Server {
	matchingServers := []servers.Server{}
	allServers := GetAllServers(client)

	for _, server := range allServers {
		if strings.Contains(server.Name, name) {
			matchingServers = append(matchingServers, server)
		}
	}

	return matchingServers
}

// FindServerByExactName will search the current tenant for servers matching that name, if more than one server is found
// an error will be returned to indicate that no unique server could be retrieved. (searching by ID would be required)
func FindServerByExactName(client *gophercloud.ServiceClient, name string) (servers.Server, error) {
	allServers := GetAllServers(client)
	var s servers.Server
	for _, server := range allServers {
		if server.Name == name {
			if s.ID != "" {
				return s, errors.New(fmt.Sprintf("Multiple Servers Found With Name %s!\n", name))
			} else {
				s = server
			}
		}
	}
	if s.ID == "" {
		return s, errors.New(fmt.Sprintf("No Servers Found With Name %s!\n", name))
	}
	return s, nil
}

func ListAllKeypairs(client *gophercloud.ServiceClient) {
	allPages, err := keypairs.List(client).AllPages()
	if err != nil {
		log.Panic(err)
	}

	allKeyPairs, err := keypairs.ExtractKeyPairs(allPages)
	if err != nil {
		log.Panic(err)
	}

	for _, kp := range allKeyPairs {
		fmt.Println(kp)
	}
}

func DeleteServer(client *gophercloud.ServiceClient, serverId string) {
	err := servers.Delete(client, serverId).ExtractErr()
	if err != nil {
		panic(err)
	}
	fmt.Println("Waiting For server to delete will timeout in 600 seconds")
	servers.WaitForStatus(client, serverId, "", 600)
	fmt.Printf("server %s deleted!", serverId)
}

func prepareNetworkCreateOpts(networks []networks.Network) []servers.Network {
	var s []servers.Network
	for _, n := range networks {
		s = append(s, servers.Network{
			UUID: n.ID,
		})
	}
	return s
}

func AttachNetworkToOpts(opts *servers.CreateOpts, n []networks.Network) {
	networkList := prepareNetworkCreateOpts(n)
	opsnet := GetTenantOpsNet(GetNetworkClient())

	networkList = append(networkList, servers.Network{UUID: opsnet.ID})
	opts.Networks = networkList
}

func CreateServerWithOptions(client *gophercloud.ServiceClient, opts servers.CreateOpts) *servers.Server {
	createOpts := keypairs.CreateOptsExt{
		CreateOptsBuilder: opts,
		KeyName:           "opskey",
	}

	server, err := servers.Create(client, createOpts).Extract()
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("Waiting For server to transition to ACTIVE, will timeout in 600 seconds")
	servers.WaitForStatus(client, server.ID, "ACTIVE", 600)
	fmt.Printf("server %s created!\n", server.ID)

	return server
}

func CreateServer(client *gophercloud.ServiceClient, serverName, imageName, flavorName string, n []networks.Network) *servers.Server {
	// prepare network list
	networkList := prepareNetworkCreateOpts(n)
	// add tenant opsnet by default
	opsnet := GetTenantOpsNet(GetNetworkClient())
	networkList = append(networkList, servers.Network{UUID: opsnet.ID})

	serverCreateOpts := servers.CreateOpts{
		Name:      serverName,
		ImageRef:  imageName,
		FlavorRef: flavorName,
		Networks: networkList,
		Metadata: map[string]string{
			"hsa29-test": "true",
		},
	}

	createOpts := keypairs.CreateOptsExt{
		CreateOptsBuilder: serverCreateOpts,
		KeyName:           "opskey",
	}

	server, err := servers.Create(client, createOpts).Extract()
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("Waiting For server to transition to ACTIVE, will timeout in 600 seconds")
	servers.WaitForStatus(client, server.ID, "ACTIVE", 600)
	fmt.Printf("server %s created!\n", server.ID)

	return server
}

func CreateVolume(client *gophercloud.ServiceClient, volumeName string) *volumes.Volume {
	opts := volumes.CreateOpts{Size: 10, Name: volumeName, VolumeType: "d4559dc6-3abc-49a1-aed6-a2c4f0b4ceac"}
	vol, err := volumes.Create(client, opts).Extract()

	if err != nil {
		log.Panic(err)
	}
	volumes.WaitForStatus(client, vol.ID, "available", 600)
	fmt.Printf("volume %s created!\n", volumeName)
	return vol
}

func AttachVolume(client *gophercloud.ServiceClient, volume *volumes.Volume, server *servers.Server) {
	fmt.Printf("Attempting to attach volume %s to %s!\n", volume.ID, server.ID)
	createOpts := volumeattach.CreateOpts{
		Device:   "/dev/vdc",
		VolumeID: volume.ID,
	}

	_, err := volumeattach.Create(client, server.ID, createOpts).Extract()
	if err != nil {
		panic(err)
	}
	volumes.WaitForStatus(client, volume.ID, "in-use", 600)
}
