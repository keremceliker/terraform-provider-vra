package vra

import (
	"fmt"

	"github.com/vmware/vra-sdk-go/pkg/client/cloud_account"
	"github.com/vmware/vra-sdk-go/pkg/models"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCloudAccountNSXV() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudAccountNSXVCreate,
		Read:   resourceCloudAccountNSXVRead,
		Update: resourceCloudAccountNSXVUpdate,
		Delete: resourceCloudAccountNSXVDelete,

		Schema: map[string]*schema.Schema{
			"accept_self_signed_cert": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"dc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"hostname": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"tags": tagsSchema(),
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceCloudAccountNSXVCreate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*Client).apiClient

	tags := expandTags(d.Get("tags").(*schema.Set).List())

	createResp, err := apiClient.CloudAccount.CreateNsxVCloudAccount(cloud_account.NewCreateNsxVCloudAccountParams().WithBody(&models.CloudAccountNsxVSpecification{
		AcceptSelfSignedCertificate: d.Get("accept_self_signed_cert").(bool),
		Dcid:                        withString(d.Get("dc_id").(string)),
		Description:                 d.Get("description").(string),
		HostName:                    withString(d.Get("hostname").(string)),
		Name:                        withString(d.Get("name").(string)),
		Password:                    withString(d.Get("password").(string)),
		Tags:                        tags,
		Username:                    withString(d.Get("username").(string)),
	}))

	if err != nil {
		return err
	}

	if err := d.Set("tags", flattenTags(tags)); err != nil {
		return fmt.Errorf("error setting cloud account tags - error: %#v", err)
	}
	d.SetId(*createResp.Payload.ID)

	return resourceCloudAccountNSXVRead(d, m)
}

func resourceCloudAccountNSXVRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*Client).apiClient

	id := d.Id()
	ret, err := apiClient.CloudAccount.GetNsxVCloudAccount(cloud_account.NewGetNsxVCloudAccountParams().WithID(id))
	if err != nil {
		switch err.(type) {
		case *cloud_account.GetNsxVCloudAccountNotFound:
			d.SetId("")
			return nil
		}
		return err
	}
	nsxvAccount := *ret.Payload

	d.Set("dc_id", nsxvAccount.Dcid)
	d.Set("description", nsxvAccount.Description)
	d.Set("name", nsxvAccount.Name)
	d.Set("username", nsxvAccount.Username)

	if err := d.Set("tags", flattenTags(nsxvAccount.Tags)); err != nil {
		return fmt.Errorf("error setting cloud account tags - error: %#v", err)
	}

	return nil
}

func resourceCloudAccountNSXVUpdate(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*Client).apiClient

	id := d.Id()

	_, err := apiClient.CloudAccount.UpdateNsxVCloudAccount(cloud_account.NewUpdateNsxVCloudAccountParams().WithID(id).WithBody(&models.UpdateCloudAccountSpecificationBase{
		Description: d.Get("description").(string),
		Tags:        expandTags(d.Get("tags").(*schema.Set).List()),
	}))
	if err != nil {
		return err
	}

	return resourceCloudAccountNSXVRead(d, m)
}

func resourceCloudAccountNSXVDelete(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*Client).apiClient

	id := d.Id()
	_, err := apiClient.CloudAccount.DeleteCloudAccountNsxV(cloud_account.NewDeleteCloudAccountNsxVParams().WithID(id))
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
