package vault

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAWSSecret(t *testing.T) {
	mountPath := acctest.RandomWithPrefix("aws")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if accessKey == "" {
		t.Skip("AWS_ACCESS_KEY_ID not set")
	}
	if secretKey == "" {
		t.Skip("AWS_SECRET_ACCESS_KEY not set")
	}
	resource.Test(t, resource.TestCase{
		Providers: testProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:             testAccDataSourceAWSSecret_configBasic(mountPath, accessKey, secretKey),
				Check:              testAccDataSourceAWSSecret_check(mountPath),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSourceAWSSecret_configBasic(mountPath, accessKey, secretKey string) string {
	return fmt.Sprintf(`
resource "vault_aws_secret_backend" "aws" {
    path = "%s"
    description = "Obtain AWS credentials."
    access_key = "%s"
    secret_key = "%s"
}

resource "vault_generic_secret" "policy" {
    path = "${vault_aws_secret_backend.aws.path}/roles/test"
    data_json = <<EOT
{
    "policy": "{\"Version\": \"2012-10-17\", \"Statement\": [{\"Effect\": \"Allow\", \"Action\": \"iam:*\", \"Resource\": \"*\"}]}"
}
EOT
}

data "vault_aws_secret" "test" {
    path = "${vault_aws_secret_backend.aws.path}/creds/test"
    depends_on = ["vault_generic_secret.policy"]
}`, mountPath, accessKey, secretKey)
}

func testAccDataSourceAWSSecret_check(mountPath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState := s.Modules[0].Resources["data.vault_aws_secret.test"]
		if resourceState == nil {
			return fmt.Errorf("resource not found in state %v", s.Modules[0].Resources)
		}

		iState := resourceState.Primary
		if iState == nil {
			return fmt.Errorf("resource has no primary instance")
		}

		return nil
	}
}
