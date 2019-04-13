package cas

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vmware/cas-sdk-go/pkg/client"
	"github.com/vmware/cas-sdk-go/pkg/client/compute"
	"github.com/vmware/cas-sdk-go/pkg/client/request"
	"github.com/vmware/cas-sdk-go/pkg/models"
	tango "github.com/vmware/terraform-provider-cas/sdk"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceMachine() *schema.Resource {
	return &schema.Resource{
		Create: resourceMachineCreate,
		Read:   resourceMachineRead,
		Update: resourceMachineUpdate,
		Delete: resourceMachineDelete,

		Schema: map[string]*schema.Schema{
			"image": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"flavor": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"project_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			//"image_ref": &schema.Schema{
			//	Type:     schema.TypeString,
			//	Required: true,
			//},
			"power_state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"machine_count": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"constraints": constraintsSchema(),
			"tags":        tagsSchema(),
			"custom_properties": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Optional: true,
			},
			"nics": nicsSchema(false),
			"disks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"block_device_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"boot_config": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"external_zone_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_region_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return !strings.HasPrefix(new, old)
				},
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMachineCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*tango.Client)
	apiClient := client.GetAPIClient()

	createResp, err := apiClient.Compute.CreateMachine(compute.NewCreateMachineParams().WithBody(&models.MachineSpecification{
		Name:             withString(d.Get("name").(string)),
		Image:            withString(d.Get("image").(string)),
		Flavor:           withString(d.Get("flavor").(string)),
		ProjectID:        withString(d.Get("project_id").(string)),
		MachineCount:     int32(d.Get("machine_count").(int)),
		Constraints:      expandSDKConstraints(d.Get("constraints").([]interface{})),
		Tags:             expandSDKTags(d.Get("tags").([]interface{})),
		CustomProperties: expandCustomProperties(d.Get("custom_properties").(map[string]interface{})),
		Nics:             expandSDKNics(d.Get("nics").([]interface{})),
		Disks:            expandSDKDisks(d.Get("disks").([]interface{})),
	}))

	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"INPROGRESS", "FAILED"},
		Target:     []string{"FINISHED"},
		Refresh:    machineStateRefreshFunc(*apiClient, *createResp.Payload.ID),
		Timeout:    5 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	// Wait for resource to be created
	newMachines, err := stateConf.WaitForState()
	log.Printf("Waitforstate returned: %T %+v %+v\n", newMachines, newMachines, err)

	if err != nil {
		return err
	}

	machines := newMachines.([]string)
	if len(machines) != 1 {
		return fmt.Errorf("total number of machines created was not 1 (found %d)", len(machines))
	}

	d.SetId(machines[0])

	return resourceMachineRead(d, m)
}

func machineStateRefreshFunc(apiClient client.MulticloudIaaS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := apiClient.Request.GetRequestTracker(request.NewGetRequestTrackerParams().WithID(id))
		if err != nil {
			return "", "FAILED", err
		}

		status := resp.Payload.Status
		switch *status {
		case "FAILED":
			return [...]string{""}, *status, fmt.Errorf(resp.Payload.Message)
		case "INPROGRESS":
			return [...]string{id}, *status, nil
		case "FINISHED":
			newMachines := make([]string, len(resp.Payload.Resources))
			for i, r := range resp.Payload.Resources {
				newMachines[i] = strings.TrimPrefix(r, "/iaas/api/machines/")
			}
			return newMachines, "FINISHED", nil
		default:
			return [...]string{id}, resp.Payload.Message, fmt.Errorf("machineStateRefreshFunc: unknown status %s", *status)
		}
	}
}

func expandDisks(configDisks []interface{}) []tango.Disk {
	disks := make([]tango.Disk, 0, len(configDisks))

	for _, configDisk := range configDisks {
		diskMap := configDisk.(map[string]interface{})

		disk := tango.Disk{
			BlockDeviceID: diskMap["block_device_id"].(string),
		}

		if v, ok := diskMap["name"].(string); ok && v != "" {
			disk.Name = v
		}

		if v, ok := diskMap["description"].(string); ok && v != "" {
			disk.Description = v
		}

		disks = append(disks, disk)
	}

	return disks
}

func expandSDKDisks(configDisks []interface{}) []*models.DiskAttachmentSpecification {
	disks := make([]*models.DiskAttachmentSpecification, 0, len(configDisks))

	for _, configDisk := range configDisks {
		diskMap := configDisk.(map[string]interface{})

		disk := models.DiskAttachmentSpecification{
			BlockDeviceID: withString(diskMap["block_device_id"].(string)),
		}

		if v, ok := diskMap["name"].(string); ok && v != "" {
			disk.Name = &v
		}

		if v, ok := diskMap["description"].(string); ok && v != "" {
			disk.Description = v
		}

		disks = append(disks, &disk)
	}

	return disks
}

func resourceMachineRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*tango.Client)
	apiClient := client.GetAPIClient()

	id := d.Id()
	ret, err := apiClient.Compute.GetMachine(compute.NewGetMachineParams().WithID(id))
	if err != nil {
		switch err.(type) {
		case *compute.GetMachineNotFound:
			d.SetId("")
			return nil
		}
		return err
	}

	machine := *ret.Payload

	d.Set("power_state", machine.PowerState)
	d.Set("address", machine.Address)
	d.Set("external_zone_id", machine.ExternalZoneID)
	d.Set("external_region_id", machine.ExternalRegionID)
	d.Set("external_id", machine.ExternalID)
	d.Set("name", machine.Name)
	d.Set("description", machine.Description)
	d.Set("created_at", machine.CreatedAt)
	d.Set("updated_at", machine.UpdatedAt)
	d.Set("owner", machine.Owner)
	d.Set("organization_id", machine.OrganizationID)
	d.Set("custom_properties", machine.CustomProperties)

	return nil
}

func resourceMachineUpdate(d *schema.ResourceData, m interface{}) error {

	return fmt.Errorf("resourceMachineUpdate not implemented")
}

func resourceMachineDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*tango.Client)
	apiClient := client.GetAPIClient()

	id := d.Id()
	_, err := apiClient.Compute.DeleteMachine(compute.NewDeleteMachineParams().WithID(id))
	if err != nil {
		return err
	}

	d.SetId("")

	return nil
}
