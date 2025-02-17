package vra

import (
	"errors"

	"github.com/hashicorp/terraform/helper/schema"
)

// Provider represents the VRA provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VRA_URL", nil),
				Description: "The base url for API operations.",
			},
			"refresh_token": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"access_token"},
				DefaultFunc:   schema.EnvDefaultFunc("VRA_REFRESH_TOKEN", nil),
				Description:   "The refresh token for API operations.",
			},
			"access_token": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"refresh_token"},
				DefaultFunc:   schema.EnvDefaultFunc("VRA_ACCESS_TOKEN", nil),
				Description:   "The access token for API operations.",
			},
			"insecure": {
				Type:        schema.TypeBool,
				DefaultFunc: schema.EnvDefaultFunc("VRA7_INSECURE", nil),
				Optional:    true,
				Description: "Specify whether to validate TLS certificates.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"vra_cloud_account_aws":   dataSourceCloudAccountAWS(),
			"vra_cloud_account_azure": dataSourceCloudAccountAzure(),
			"vra_cloud_account_gcp":   dataSourceCloudAccountGCP(),
			"vra_cloud_account_nsxt":  dataSourceCloudAccountNSXT(),
			"vra_cloud_account_nsxv":  dataSourceCloudAccountNSXV(),
			"vra_cloud_account_vmc":   dataSourceCloudAccountVMC(),
			"vra_data_collector":      dataSourceDataCollector(),
			"vra_fabric_network":      dataSourceFabricNetwork(),
			"vra_image":               dataSourceImage(),
			"vra_network":             dataSourceNetwork(),
			"vra_network_domain":      dataSourceNetworkDomain(),
			"vra_project":             dataSourceProject(),
			"vra_region":              dataSourceRegion(),
			"vra_region_enumeration":  dataSourceRegionEnumeration(),
			"vra_security_group":      dataSourceSecurityGroup(),
			"vra_zone":                dataSourceZone(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"vra_block_device":          resourceBlockDevice(),
			"vra_cloud_account_aws":     resourceCloudAccountAWS(),
			"vra_cloud_account_azure":   resourceCloudAccountAzure(),
			"vra_cloud_account_gcp":     resourceCloudAccountGCP(),
			"vra_cloud_account_nsxt":    resourceCloudAccountNSXT(),
			"vra_cloud_account_nsxv":    resourceCloudAccountNSXV(),
			"vra_cloud_account_vmc":     resourceCloudAccountVMC(),
			"vra_cloud_account_vsphere": resourceCloudAccountVsphere(),
			"vra_flavor_profile":        resourceFlavorProfile(),
			"vra_image_profile":         resourceImageProfile(),
			"vra_load_balancer":         resourceLoadBalancer(),
			"vra_machine":               resourceMachine(),
			"vra_network":               resourceNetwork(),
			"vra_network_profile":       resourceNetworkProfile(),
			"vra_project":               resourceProject(),
			"vra_storage_profile":       resourceStorageProfile(),
			"vra_storage_profile_aws":   resourceStorageProfileAws(),
			"vra_storage_profile_azure": resourceStorageProfileAzure(),
			"vra_zone":                  resourceZone(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	url := d.Get("url").(string)
	refreshToken := ""
	accessToken := ""

	if v, ok := d.GetOk("refresh_token"); ok {
		refreshToken = v.(string)
	}

	if v, ok := d.GetOk("access_token"); ok {
		accessToken = v.(string)
	}

	insecure := d.Get("insecure").(bool)

	if accessToken == "" && refreshToken == "" {
		return nil, errors.New("refresh_token or access_token required")
	}

	if accessToken != "" {
		return NewClientFromAccessToken(url, accessToken, insecure)
	}

	return NewClientFromRefreshToken(url, refreshToken, insecure)
}
