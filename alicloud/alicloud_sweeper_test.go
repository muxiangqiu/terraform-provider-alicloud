package alicloud

import (
	"fmt"
	"os"
	"testing"

	"github.com/denverdino/aliyungo/common"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// sharedClientForRegion returns a common AWSClient setup needed for the sweeper
// functions for a given region
func sharedClientForRegion(region string) (interface{}, error) {
	var accessKey, secretKey string
	if accessKey = os.Getenv("ALICLOUD_ACCESS_KEY"); accessKey == "" {
		return nil, fmt.Errorf("empty ALICLOUD_ACCESS_KEY")
	}

	if secretKey = os.Getenv("ALICLOUD_SECRET_KEY"); secretKey == "" {
		return nil, fmt.Errorf("empty ALICLOUD_SECRET_KEY")
	}

	conf := Config{
		Region:    common.Region(region),
		RegionId:  region,
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
	if accountId := os.Getenv("ALICLOUD_ACCOUNT_ID"); accountId != "" {
		conf.AccountId = accountId
	}

	// configures a default client for the region, using the above env vars
	client, err := conf.Client()
	if err != nil {
		return nil, err
	}

	return client, nil
}
