package openstack_utlities


import (
	"errors"
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"log"
)

func GetAllImages(client *gophercloud.ServiceClient, opts images.ListOpts) []images.Image{
	allPages, err := images.ListDetail(client, opts).AllPages()
	if err != nil {
		log.Panic(err)
	}

	allImages, err := images.ExtractImages(allPages)
	if err != nil {
		log.Panic(err)
	}

	return allImages
}

func GetImageByID(client *gophercloud.ServiceClient, imageID string) images.Image {
	var i images.Image
	allImages := GetAllImages(client, images.ListOpts{})

	for _, image := range allImages {
		if image.ID == imageID {
			i = image
		}
	}

	return i
}

func GetImageByName(client *gophercloud.ServiceClient, imageName string) (images.Image, error) {
	listOpst := images.ListOpts{
		Name: imageName,
	}

	allImages := GetAllImages(client, listOpst)

	if len(allImages) > 1 {
		return images.Image{}, errors.New(fmt.Sprintf("More than One Image Found With Name %s\n", imageName))
	}
	return allImages[0], nil
}