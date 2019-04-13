package cas

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	tango "github.com/vmware/terraform-provider-cas/sdk"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/cas-sdk-go/pkg/client/compute"
)

func TestAccTangoMachine_Basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckAWS(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTangoMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckTangoMachineConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTangoMachineExists("cas_machine.my_machine"),
					resource.TestMatchResourceAttr(
						"cas_machine.my_machine", "name", regexp.MustCompile("^terraform_cas_machine-"+strconv.Itoa(rInt))),
					resource.TestCheckResourceAttr(
						"cas_machine.my_machine", "image", "ubuntu"),
					resource.TestCheckResourceAttr(
						"cas_machine.my_machine", "flavor", "small"),
					resource.TestCheckResourceAttr(
						"cas_machine.my_machine", "tags.#", "1"),
					resource.TestCheckResourceAttr(
						"cas_machine.my_machine", "tags.0.key", "stoyan"),
					resource.TestCheckResourceAttr(
						"cas_machine.my_machine", "tags.0.value", "genchev"),
				),
			},
		},
	})
}

func testAccCheckTangoMachineExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no machine ID is set")
		}

		return nil
	}
}

func testAccCheckTangoMachineDestroy(s *terraform.State) error {
	client := testAccProviderCAS.Meta().(*tango.Client)
	apiClient := client.GetAPIClient()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cas_machine" {
			continue
		}

		_, err := apiClient.Compute.GetMachine(compute.NewGetMachineParams().WithID(rs.Primary.ID))
		_, ok := err.(*compute.GetMachineNotFound)
		if err != nil && !ok {
			return fmt.Errorf("error waiting for machine (%s) to be destroyed: %s", rs.Primary.ID, err)
		}
	}
	return nil
}

func testAccCheckTangoMachineConfig(rInt int) string {
	return fmt.Sprintf(`
data "cas_project" "test_project" {
	name = "test-project"
}

resource "cas_machine" "my_machine" {
	name = "terraform_cas_machine-%d"
	project_id = "${data.cas_project.test_project.id}"
  image = "ubuntu"
  flavor = "small"

  tags {
	key = "stoyan"
    value = "genchev"
  }
}`, rInt)
}
