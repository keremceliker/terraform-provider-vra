provider "vra" {
  url           = var.url
  refresh_token = var.refresh_token
}

data "vra_cloud_account_aws" "this" {
  name = var.cloud_account
}

data "vra_region" "this" {
  cloud_account_id = data.vra_cloud_account_aws.this.id
  region           = var.region
}

data "vra_fabric_network" "subnet" {
  filter = "name eq '${var.subnet_name}' and cloudAccountId eq '${data.vra_cloud_account_aws.this.id}' and externalRegionId eq '${var.region}'"
}

data "vra_network_domain" "vpc" {
  filter = "name eq '${var.network_domain_name}' and cloudAccountId eq '${data.vra_cloud_account_aws.this.id}' and externalRegionId eq '${var.region}'"
}

data "vra_fabric_network" "external_subnet" {
  filter = "name eq '${var.external_subnet_name}' and cloudAccountId eq '${data.vra_cloud_account_aws.this.id}' and externalRegionId eq '${var.region}'"
}

resource "vra_network_profile" "external_subnet_isolation" {
  name        = "isolation-with-external-subnet"
  description = "On-demand networks are created for outbound and private networks."
  region_id   = data.vra_region.this.id

  fabric_network_ids = [
    data.vra_fabric_network.subnet.id
  ]

  isolation_type                       = "SUBNET"
  isolation_network_domain_id          = data.vra_network_domain.vpc.id
  isolated_network_cidr_prefix         = var.cidr_prefix
  isolation_external_fabric_network_id = data.vra_fabric_network.external_subnet.id

  tags {
    key   = "foo"
    value = "bar"
  }
}
