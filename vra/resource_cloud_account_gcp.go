package vra

import (
	"fmt"

	"github.com/vmware/vra-sdk-go/pkg/client/cloud_account"
	"github.com/vmware/vra-sdk-go/pkg/models"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCloudAccountGCP() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudAccountGCPCreate,
		Read:   resourceCloudAccountGCPRead,
		Update: resourceCloudAccountGCPUpdate,
		Delete: resourceCloudAccountGCPDelete,

		Schema: map[string]*schema.Schema{

			"client_email": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"private_key": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"private_key_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"project_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"regions": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"region_ids": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceCloudAccountGCPCreate(d *schema.ResourceData, m interface{}) error {
	var regions []string

	apiClient := m.(*Client).apiClient

	if v, ok := d.GetOk("regions"); ok {
		if !compareUnique(v.([]interface{})) {
			return fmt.Errorf("specified regions are not unique")
		}
		regions = expandStringList(v.([]interface{}))
	}

	createResp, err := apiClient.CloudAccount.CreateGcpCloudAccount(cloud_account.NewCreateGcpCloudAccountParams().WithBody(&models.CloudAccountGcpSpecification{
		Description:        d.Get("description").(string),
		Name:               withString(d.Get("name").(string)),
		ClientEmail:        withString(d.Get("client_email").(string)),
		PrivateKey:         withString(d.Get("private_key").(string)),
		PrivateKeyID:       withString(d.Get("private_key_id").(string)),
		ProjectID:          withString(d.Get("project_id").(string)),
		CreateDefaultZones: false,
		RegionIds:          regions,
		Tags:               expandTags(d.Get("tags").(*schema.Set).List()),
	}))

	if err != nil {
		return err
	}

	d.SetId(*createResp.Payload.ID)

	return resourceCloudAccountGCPRead(d, m)
}

func resourceCloudAccountGCPRead(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*Client).apiClient

	id := d.Id()
	ret, err := apiClient.CloudAccount.GetGcpCloudAccount(cloud_account.NewGetGcpCloudAccountParams().WithID(id))
	if err != nil {
		switch err.(type) {
		case *cloud_account.GetGcpCloudAccountNotFound:
			d.SetId("")
			return nil
		}
		return err
	}
	gcpAccount := *ret.Payload

	d.Set("description", gcpAccount.Description)
	d.Set("name", gcpAccount.Name)

	d.Set("client_email", gcpAccount.ClientEmail)
	d.Set("private_key_id", gcpAccount.PrivateKeyID)
	d.Set("project_id", gcpAccount.ProjectID)

	regions := gcpAccount.EnabledRegionIds
	d.Set("regions", regions)

	// The returned EnabledRegionIds and Hrefs containing the region ids can be in a different order than the request order.
	// Call a routine to normalize the order to correspond with the users region order.
	regionsIds, err := flattenAndNormalizeCLoudAccountGcpRegionIds(regions, &gcpAccount)
	if err != nil {
		return err
	}
	d.Set("region_ids", regionsIds)

	if err := d.Set("tags", flattenTags(gcpAccount.Tags)); err != nil {
		return fmt.Errorf("error setting cloud account tags - error: %#v", err)
	}

	return nil
}

func resourceCloudAccountGCPUpdate(d *schema.ResourceData, m interface{}) error {
	var regions []string

	apiClient := m.(*Client).apiClient

	id := d.Id()

	if v, ok := d.GetOk("regions"); ok {
		if !compareUnique(v.([]interface{})) {
			return fmt.Errorf("specified regions are not unique")
		}
		regions = expandStringList(v.([]interface{}))
	}
	tags := expandTags(d.Get("tags").(*schema.Set).List())

	_, err := apiClient.CloudAccount.UpdateGcpCloudAccount(cloud_account.NewUpdateGcpCloudAccountParams().WithID(id).WithBody(&models.UpdateCloudAccountGcpSpecification{
		Description:        d.Get("description").(string),
		CreateDefaultZones: false,
		RegionIds:          regions,
		Tags:               tags,
	}))
	if err != nil {
		return err
	}

	return resourceCloudAccountGCPRead(d, m)
}

func resourceCloudAccountGCPDelete(d *schema.ResourceData, m interface{}) error {
	apiClient := m.(*Client).apiClient

	id := d.Id()
	_, err := apiClient.CloudAccount.DeleteGcpCloudAccount(cloud_account.NewDeleteGcpCloudAccountParams().WithID(id))
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
