package openstack_utlities


import (
	"errors"
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"log"
)

func GetAllFlavors(client *gophercloud.ServiceClient) []flavors.Flavor  {
	var allFlavors []flavors.Flavor

	listOpts := flavors.ListOpts{
		// Nothing to set here
	}

	allPages, err := flavors.ListDetail(client, listOpts).AllPages()
	if err != nil {
		log.Panic(err)
	}

	allFlavors, err = flavors.ExtractFlavors(allPages)
	if err != nil {
		log.Panic(err)
	}

	return allFlavors
}

func GetFlavorByName(client *gophercloud.ServiceClient, name string) (flavors.Flavor, error) {
	var f flavors.Flavor

	for _, flavor := range GetAllFlavors(client) {
		if flavor.Name == name {
			if f.ID != "" {
				return flavors.Flavor{}, errors.New(fmt.Sprintf("More than one flavor found with name: %s\n", name))
			}
			f = flavor
		}
	}

	// If No Image Was Found
	if f.ID != "" {
		return f, nil
	} else {
		return flavors.Flavor{}, errors.New(fmt.Sprintf("No Image Was Found With Name: %s\n", name))
	}
}