terraform {
  required_version = ">=0.14.7"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.75.0"
    }
  }
}
# AWS provider configuration
provider "aws" {
  region  = var.region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}

# creat ca-cert
resource "null_resource" "run_script" {
  provisioner "local-exec" {
    command = "bash ./assets/create-ca-cert.sh"
  }
  # Ensure that the script is complete before moving on to any other resources
  triggers = {
    always_run = timestamp() # This makes sure this resource runs every time you run `terraform apply`
  }
}
data "local_sensitive_file" "ca_bundle_file" {
  depends_on = [null_resource.run_script]
  filename = var.ca_bundle_file
}

#
### RESOURCES
# Create a VPC
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr_block
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags                 = { Name = "${var.name_prefix}-vpc" }
}

# Create an Internet Gateway and attach it to the VPC
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id
  tags   = { Name = "${var.name_prefix}-igw" }
}

# Create public subnet A
resource "aws_subnet" "public_subnet_a" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.0.0/19" # CIDR block for public subnet
  availability_zone       = "${var.region}a"
  map_public_ip_on_launch = true
  tags                    = { Name = "${var.name_prefix}-publicSubA" }
}
# Create public subnet B
resource "aws_subnet" "public_subnet_b" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.32.0/19" # CIDR block for public subnet
  availability_zone       = "${var.region}b"
  map_public_ip_on_launch = true
  tags                    = { Name = "${var.name_prefix}-publicSubB" }
}

# Create public subnet C
resource "aws_subnet" "public_subnet_c" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.64.0/19" # CIDR block for public subnet
  availability_zone       = "${var.region}c"
  map_public_ip_on_launch = true
  tags                    = { Name = "${var.name_prefix}-publicSubC" }
}

# Create private subnets A, B, and C
resource "aws_subnet" "private_subnet_a" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.96.0/19"
  availability_zone = "${var.region}a"
  tags = {
    Name = "${var.name_prefix}-privateSubA"
  }
}

resource "aws_subnet" "private_subnet_b" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.128.0/19"
  availability_zone = "${var.region}b"
  tags = {
    Name = "${var.name_prefix}-privateSubB"
  }
}

resource "aws_subnet" "private_subnet_c" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.160.0/19"
  availability_zone = "${var.region}c"
  tags = {
    Name = "${var.name_prefix}-privateSubC"
  }
}
# Allocate Elastic IP for NAT Gateway A
resource "aws_eip" "nat_eip_a" {
  tags = {
    Name = "${var.name_prefix}-a"
  }
}

# Create NAT Gateway A
resource "aws_nat_gateway" "nat_gateway_a" {
  subnet_id     = aws_subnet.public_subnet_a.id
  allocation_id = aws_eip.nat_eip_a.id
  tags = {
    Name = "${var.name_prefix}-natA"
  }
}

# Allocate Elastic IP for NAT Gateway B
resource "aws_eip" "nat_eip_b" {
  tags = {
    Name = "${var.name_prefix}-b"
  }
}

# Create NAT Gateway B
resource "aws_nat_gateway" "nat_gateway_b" {
  subnet_id     = aws_subnet.public_subnet_b.id
  allocation_id = aws_eip.nat_eip_b.id
  tags = {
    Name = "${var.name_prefix}-natB"
  }
}

# Allocate Elastic IP for NAT Gateway C
resource "aws_eip" "nat_eip_c" {
  tags = {
    Name = "${var.name_prefix}-c"
  }
}

# Create NAT Gateway C
resource "aws_nat_gateway" "nat_gateway_c" {
  subnet_id     = aws_subnet.public_subnet_c.id
  allocation_id = aws_eip.nat_eip_c.id
  tags = {
    Name = "${var.name_prefix}-natC"
  }
}

# Create the Public Route Table and associate it with public subnets
resource "aws_route_table" "public_rt" {
  vpc_id = aws_vpc.main.id
  tags = {
    Name = "${var.name_prefix}-publicRT"
  }
}

resource "aws_route" "public_route" {
  route_table_id         = aws_route_table.public_rt.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.igw.id
}

resource "aws_route_table_association" "public_subnet_a_association" {
  subnet_id      = aws_subnet.public_subnet_a.id
  route_table_id = aws_route_table.public_rt.id
}

resource "aws_route_table_association" "public_subnet_b_association" {
  subnet_id      = aws_subnet.public_subnet_b.id
  route_table_id = aws_route_table.public_rt.id
}

resource "aws_route_table_association" "public_subnet_c_association" {
  subnet_id      = aws_subnet.public_subnet_c.id
  route_table_id = aws_route_table.public_rt.id
}

# Create Private Route Tables and associate them with private subnets
resource "aws_route_table" "private_rt_a" {
  vpc_id = aws_vpc.main.id
  tags = {
    Name = "${var.name_prefix}-privateRtA"
  }
}

resource "aws_route" "private_route_a" {
  route_table_id         = aws_route_table.private_rt_a.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_nat_gateway.nat_gateway_a.id
}

resource "aws_route_table_association" "private_subnet_a_association" {
  subnet_id      = aws_subnet.private_subnet_a.id
  route_table_id = aws_route_table.private_rt_a.id
}

resource "aws_route_table" "private_rt_b" {
  vpc_id = aws_vpc.main.id
  tags = {
    Name = "${var.name_prefix}-privateRtB"
  }
}

resource "aws_route" "private_route_b" {
  route_table_id         = aws_route_table.private_rt_b.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_nat_gateway.nat_gateway_b.id
}

resource "aws_route_table_association" "private_subnet_b_association" {
  subnet_id      = aws_subnet.private_subnet_b.id
  route_table_id = aws_route_table.private_rt_b.id
}

resource "aws_route_table" "private_rt_c" {
  vpc_id = aws_vpc.main.id
  tags = {
    Name = "${var.name_prefix}-privateRtC"
  }
}

resource "aws_route" "private_route_c" {
  route_table_id         = aws_route_table.private_rt_c.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_nat_gateway.nat_gateway_c.id
}

resource "aws_route_table_association" "private_subnet_c_association" {
  subnet_id      = aws_subnet.private_subnet_c.id
  route_table_id = aws_route_table.private_rt_c.id
}



# Create a security group for the proxy machine (rules below)
resource "aws_security_group" "proxy_machine_sg" {
  name_prefix = var.name_prefix
  description = "Allow all outbound traffic and inbound traffic from proxied subnet or developer SSH client"
  vpc_id      = aws_vpc.main.id
  tags = {
    osd   = "proxy"
    owner = "terraform"
  }
}

resource "aws_security_group_rule" "squid_http" {
  type              = "ingress"
  from_port         = 3128
  to_port           = 3128
  protocol          = "tcp"
  cidr_blocks       = ["10.0.0.0/16"]
  security_group_id = aws_security_group.proxy_machine_sg.id
}

resource "aws_security_group_rule" "public_in_https" {
  type              = "ingress"
  from_port         = 3129
  to_port           = 3129
  protocol          = "tcp"
  cidr_blocks       = ["10.0.0.0/16"]
  security_group_id = aws_security_group.proxy_machine_sg.id
}

resource "aws_security_group_rule" "public_out_allow_all" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "all"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.proxy_machine_sg.id
}
# End proxy_machine_sg rules



# Setup squid proxy instance
resource "aws_instance" "proxy_machine" {
  ami           = data.aws_ami.rhel9.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.private_subnet_a.id
#  key_name      = "sdnOvnMigrationKeyPair"
  depends_on = [data.local_sensitive_file.ca_bundle_file]

  user_data = templatefile("assets/proxy-setup.sh.tftpl", {
    CA_BUNDLE_FILE = data.local_sensitive_file.ca_bundle_file.content
  })

  vpc_security_group_ids = [aws_security_group.proxy_machine_sg.id]
  tags = {
    osd   = "proxy"
    owner = "terraform"
    Name = "${var.name_prefix}-proxy-machine"
  }

  lifecycle {
    ignore_changes = [user_data]
  }
}


output "http_proxy_var" {
  description = "value for http_proxy environmental variable"
  value       = "http://${aws_instance.proxy_machine.private_ip}:3128"
}
output "https_proxy_var" {
  description = "value for HTTPS_PROXY environmental variable"
  value       = "https://${aws_instance.proxy_machine.private_ip}:3129"
}
output "subnets" {
  description = "value for all subnets"
  value       = "${aws_subnet.public_subnet_a.id},${aws_subnet.public_subnet_b.id},${aws_subnet.public_subnet_c.id},${aws_subnet.private_subnet_a.id},${aws_subnet.private_subnet_b.id},${aws_subnet.private_subnet_c.id}"
}

## DATA
# Get the current AWS region
data "aws_region" "current" {}

# Automatic lookup of the latest official RHEL 9 AMI
data "aws_ami" "rhel9" {
  most_recent = true

  filter {
    name   = "platform-details"
    values = ["Red Hat Enterprise Linux"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "manifest-location"
    values = ["amazon/RHEL-9.*_HVM-*-x86_64-*-Hourly2-GP2"]
  }

  owners = ["309956199498"] # Amazon's "Official Red Hat" account
}
