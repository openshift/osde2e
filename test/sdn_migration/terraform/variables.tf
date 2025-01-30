variable "vpc_cidr_block" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16" # Default to a /16 block within the 10.0.0.0 private network
}

# Prefix to add to name tags
variable "name_prefix" {
  description = "prefix to add to the Name tag associated with most of the resources created by these scripts"
  type        = string
  default     = "osde2e-sdnOvn-proxy"
}

variable "ca_bundle_file" {
  description = "the path for CA bundle file used to sign squid certificates"
  type        = string
  default     = "ca-bundle.crt"
}

variable "region" {
  description = ""
  type = string
  default = "us-east-1"
}

variable "aws_access_key_id" {
  description = ""
  type = string
}

variable "aws_secret_access_key" {
  description = ""
  type = string
}
