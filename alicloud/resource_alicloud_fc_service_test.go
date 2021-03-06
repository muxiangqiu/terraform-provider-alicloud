package alicloud

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/fc-go-sdk"
	"github.com/denverdino/aliyungo/ram"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("alicloud_fc_service", &resource.Sweeper{
		Name: "alicloud_fc_service",
		F:    testSweepFCServices,
	})
}

func testSweepFCServices(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting Alicloud client: %s", err)
	}
	conn := client.(*AliyunClient)

	prefixes := []string{
		"tf-testAcc",
		"tf_testAcc",
		"tf_test_",
		"tf-test-",
		"testAcc",
		"test-acc-alicloud",
	}

	fcconn, err := conn.Fcconn()
	if err != nil {
		return fmt.Errorf("error getting fc conn: %s", err)
	}

	if services, err := fcconn.ListServices(fc.NewListServicesInput()); err != nil {
		return fmt.Errorf("Error retrieving FC services: %s", err)
	} else {
		for _, v := range services.Services {
			name := *v.ServiceName
			id := *v.ServiceID
			skip := true
			for _, prefix := range prefixes {
				if strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
					skip = false
					break
				}
			}
			if skip {
				log.Printf("[INFO] Skipping FC services: %s (%s)", name, id)
				continue
			}
			log.Printf("[INFO] Deleting FC services: %s (%s)", name, id)
			if _, err := conn.fcconn.DeleteService(&fc.DeleteServiceInput{
				ServiceName: StringPointer(name),
			}); err != nil {
				log.Printf("[ERROR] Failed to delete FC services (%s (%s)): %s", name, id, err)
			}
		}
	}
	return nil
}

func TestAccAlicloudFCService_basic(t *testing.T) {
	if !isRegionSupports(FunctionCompute) {
		logTestSkippedBecauseOfUnsupportedRegionalFeatures(t.Name(), FunctionCompute)
		return
	}

	var service fc.GetServiceOutput
	var project sls.LogProject
	var store sls.LogStore

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAlicloudFCServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAlicloudFCServiceBasic("tf-testaccalicloudfcservicebasic", testFCRoleTemplate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlicloudLogProjectExists("alicloud_log_project.foo", &project),
					testAccCheckAlicloudLogStoreExists("alicloud_log_store.foo", &store),
					testAccCheckAlicloudFCServiceExists("alicloud_fc_service.foo", &service),
					resource.TestCheckResourceAttr("alicloud_fc_service.foo", "name", "tf-testaccalicloudfcservicebasic"),
					resource.TestCheckResourceAttr("alicloud_fc_service.foo", "description", "tf unit test"),
				),
			},
		},
	})
}

func TestAccAlicloudFCService_update(t *testing.T) {
	if !isRegionSupports(FunctionCompute) {
		logTestSkippedBecauseOfUnsupportedRegionalFeatures(t.Name(), FunctionCompute)
		return
	}

	var service fc.GetServiceOutput
	var vpcInstance vpc.DescribeVpcAttributeResponse
	var group ecs.DescribeSecurityGroupAttributeResponse
	var vsw vpc.DescribeVSwitchAttributesResponse
	var role ram.Role

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAlicloudFCServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAlicloudFCServiceUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAlicloudFCServiceExists("alicloud_fc_service.foo", &service),
					resource.TestCheckResourceAttr("alicloud_fc_service.foo", "name", "tf-testAlicloudFCServiceUpdate"),
					resource.TestCheckResourceAttr("alicloud_fc_service.foo", "description", "tf unit test"),
					resource.TestCheckResourceAttr("alicloud_fc_service.foo", "internet_access", "false"),
				),
			},
			{
				Config: testAlicloudFCServiceVpc(testFCRoleTemplate, testFCVpcPolicyTemplate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("alicloud_vpc.vpc", &vpcInstance),
					testAccCheckVswitchExists("alicloud_vswitch.vsw", &vsw),
					testAccCheckSecurityGroupExists("alicloud_security_group.group", &group),
					testAccCheckRamRoleExists("alicloud_ram_role.role", &role),
					testAccCheckAlicloudFCServiceExists("alicloud_fc_service.foo", &service),
					resource.TestCheckResourceAttr("alicloud_fc_service.foo", "name", "tf-testAlicloudFCServiceUpdate"),
					resource.TestCheckResourceAttr("alicloud_fc_service.foo", "description", "tf unit test"),
					resource.TestCheckResourceAttr("alicloud_fc_service.foo", "internet_access", "false"),
				),
			},
		},
	})
}

func testAccCheckAlicloudFCServiceExists(name string, service *fc.GetServiceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Log store ID is set")
		}

		client := testAccProvider.Meta().(*AliyunClient)

		ser, err := client.DescribeFcService(rs.Primary.ID)
		if err != nil {
			return err
		}

		service = ser

		return nil
	}
}

func testAccCheckAlicloudFCServiceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*AliyunClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "alicloud_fc_service" {
			continue
		}

		ser, err := client.DescribeFcService(rs.Primary.ID)
		if err != nil {
			if NotFoundError(err) {
				continue
			}
			return fmt.Errorf("Check fc service got an error: %#v.", err)
		}

		if ser == nil {
			return nil
		}

		return fmt.Errorf("FC service %s still exists.", rs.Primary.ID)
	}

	return nil
}

func testAlicloudFCServiceBasic(name, role string) string {
	return fmt.Sprintf(`
variable "name" {
    default = "%s"
}

resource "alicloud_log_project" "foo" {
    name = "${var.name}"
}

resource "alicloud_log_store" "foo" {
    project = "${alicloud_log_project.foo.name}"
    name = "${var.name}"
}

resource "alicloud_ram_role" "role" {
  name = "${var.name}"
  document = <<DEFINITION
  %s
  DEFINITION
  description = "this is a test"
  force = true
}

resource "alicloud_ram_role_policy_attachment" "attac" {
  role_name = "${alicloud_ram_role.role.name}"
  policy_name = "AliyunLogFullAccess"
  policy_type = "System"
}

resource "alicloud_fc_service" "foo" {
    name = "${var.name}"
    description = "tf unit test"
    log_config {
	project = "${alicloud_log_project.foo.name}"
	logstore = "${alicloud_log_store.foo.name}"
    }
    role = "${alicloud_ram_role.role.arn}"
    depends_on = ["alicloud_ram_role_policy_attachment.attac"]
}
`, name, role)
}

const testAlicloudFCServiceUpdate = `
variable "name" {
    default = "tf-testAlicloudFCServiceUpdate"
}
resource "alicloud_fc_service" "foo" {
    name = "${var.name}"
    description = "tf unit test"
    internet_access = false
}
`

func testAlicloudFCServiceVpc(role, policy string) string {
	return fmt.Sprintf(`
variable "name" {
    default = "tf-testAlicloudFCServiceUpdate"
}
resource "alicloud_vpc" "vpc" {
  name = "${var.name}"
  cidr_block = "172.16.0.0/16"
}

data "alicloud_zones" "zones" {
    available_resource_creation = "VSwitch"
}

resource "alicloud_vswitch" "vsw" {
  name = "${var.name}"
  cidr_block = "172.16.0.0/24"
  availability_zone = "${data.alicloud_zones.zones.zones.0.id}"
  vpc_id = "${alicloud_vpc.vpc.id}"
}
resource "alicloud_security_group" "group" {
  name = "${var.name}"
  vpc_id = "${alicloud_vpc.vpc.id}"
}

resource "alicloud_ram_role" "role" {
  name = "${var.name}"
  document = <<DEFINITION
  %s
  DEFINITION
  description = "this is a test"
  force = true
}

resource "alicloud_ram_role_policy_attachment" "attac" {
  role_name = "${alicloud_ram_role.role.name}"
  policy_name = "AliyunLogFullAccess"
  policy_type = "System"
}

resource "alicloud_ram_policy" "vpc" {
  name = "${var.name}"
  document = <<DEFINITION
  %s
  DEFINITION
}
resource "alicloud_ram_role_policy_attachment" "vpc" {
  role_name = "${alicloud_ram_role.role.name}"
  policy_name = "${alicloud_ram_policy.vpc.name}"
  policy_type = "Custom"
}
resource "alicloud_fc_service" "foo" {
  name = "${var.name}"
  description = "tf unit test"
  vpc_config {
    vswitch_ids = [
      "${alicloud_vswitch.vsw.id}"]
    security_group_id = "${alicloud_security_group.group.id}"
  }
  internet_access = false
  role = "${alicloud_ram_role.role.arn}"
  depends_on = ["alicloud_ram_role_policy_attachment.attac", "alicloud_ram_role_policy_attachment.vpc"]
}
`, role, policy)
}

var testFCRoleTemplate = `
{
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "fc.aliyuncs.com"
        ]
      }
    }
  ],
  "Version": "1"
}
`

var testFCVpcPolicyTemplate = `
{
  "Version": "1",
  "Statement": [
    {
      "Action": "vpc:*",
      "Resource": "*",
      "Effect": "Allow"
    },
    {
      "Action": [
        "ecs:*NetworkInterface*"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
`
