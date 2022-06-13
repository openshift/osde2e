package aws

import (
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//Refactor: Some of these variables should be set through Vyper.(Availability Zone, Subnets Cidr Block, etc)
//Should I make this a struct to hold these variables?
var (
	ByoVpcName              string
	vpcId                   string
	internetGatewayId       string
	publicSubnetId          string
	privateSubnetId         string
	elasticIpId             string
	natGatewayId            string
	publicRouteTableId      string
	privateRouteTableId     string
	vpcCidrBlock            = "10.0.0.0/16"
	publicSubnetsCidrBlock  = "10.0.0.0/17"
	privateSubnetsCidrBlock = "10.0.128.0/17"
	availabilityZone        = "us-east-1a"
)

func init() {
	//To be refactord later into the overall init and to be trigged by the flags that enable this job.
	ByoVpcName = "BYO-VPC-" + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
}

//Prepare VPC for BYO environment using Squid Proxy
func ByoVpcSetup() ([]string, error) {
	CcsAwsSession.getIamClient()
	var err error

	vpcId, err = createByoVpc(ByoVpcName, vpcCidrBlock)
	if err != nil {
		return nil, err
	}

	internetGatewayId, err = byoVpcInternetGatewaySetup(vpcId)
	if err != nil {
		return nil, err
	}

	publicSubnetId, err = byoVpcPublicSubnetSetup(vpcId, availabilityZone, publicSubnetsCidrBlock)
	if err != nil {
		return nil, err
	}

	privateSubnetId, err = byoVpcPrivateSubnetSetup(vpcId, availabilityZone, privateSubnetsCidrBlock)
	if err != nil {
		return nil, err
	}

	elasticIpId, err = byoVpcElasticIpSetup()
	if err != nil {
		return nil, err
	}

	natGatewayId, err = byoVpcNatGatewaySetup(vpcId, publicSubnetId, elasticIpId)
	if err != nil {
		return nil, err
	}

	publicRouteTableId, err = byoVpcCreatePublicRouteTable(vpcId)
	if err != nil {
		return nil, err
	} else {
		err = byoVpcAssociatePublicTable(publicRouteTableId, publicSubnetId)
		if err != nil {
			return nil, err
		} else {
			err = byoVpcCreatePublicRoute(publicRouteTableId, natGatewayId)
			if err != nil {
				return nil, err
			}
		}
	}

	privateRouteTableId, err = byoVpcCreatePrivateRouteTable(vpcId)
	if err != nil {
		return nil, err
	} else {
		err = byoVpcAssociatePrivateTable(privateRouteTableId, privateSubnetId)
		if err != nil {
			return nil, err
		} else {
			err = byoVpcCreatePrivateRoute(privateRouteTableId, natGatewayId)
			if err != nil {
				return nil, err
			}
		}
	}

	vpcSubnetIds, err := byoVpcGetVpcSubnetIds(vpcId)
	if err != nil {
		return nil, err
	}

	return vpcSubnetIds, nil
}

// createVpc creates a VPC with the given CIDR block and tags it with the given tagSpecification
func createByoVpc(name string, cidrBlock string) (string, error) {
	var err error

	// Create a new VPC
	input := &ec2.CreateVpcInput{
		CidrBlock: aws.String(cidrBlock),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("vpc"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(name),
					},
				},
			},
		},
	}

	result, err := CcsAwsSession.ec2.CreateVpc(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidVpc.Cidr":
				log.Printf("The CIDR block you specified is invalid.")
				return "", err
			case "InvalidVpc.CidrConflict":
				log.Printf("The CIDR block you specified is already in use by another VPC.")
				return "", err
			default:
				log.Printf("Error creating VPC: %v", aerr.Error())
				return "", err
			}
		} else {
			log.Printf("Error creating VPC: %v", err)
			return "", err
		}
	}
	log.Printf("Created VPC %s", *result.Vpc.VpcId)

	_, err = CcsAwsSession.ec2.ModifyVpcAttribute(&ec2.ModifyVpcAttributeInput{
		VpcId: result.Vpc.VpcId,
		EnableDnsHostnames: &ec2.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	})
	if err != nil {
		log.Printf("Error enabling DNS hostnames: %v", err)
		return "", err
	}
	log.Printf("Enabled DNS hostnames for VPC %s", *result.Vpc.VpcId)

	return *result.Vpc.VpcId, err
}

func byoVpcInternetGatewaySetup(vpcId string) (string, error) {

	input := &ec2.CreateInternetGatewayInput{
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("internet-gateway"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(ByoVpcName + "-IGW"),
					},
				},
			},
		},
	}

	result, err := CcsAwsSession.ec2.CreateInternetGateway(input)
	if err != nil {
		log.Printf("Error creating Internet Gateway: %v", err)
		return "", err
	} else {
		log.Printf("Created Internet Gateway %s", *result.InternetGateway.InternetGatewayId)
	}

	_, err = CcsAwsSession.ec2.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: result.InternetGateway.InternetGatewayId,
		VpcId:             aws.String(vpcId),
	})
	if err != nil {
		log.Printf("Error attaching Internet Gateway: %v", err)
		return "", err
	} else {
		log.Printf("Attached Internet Gateway %s to VPC %s", *result.InternetGateway.InternetGatewayId, vpcId)
	}

	return *result.InternetGateway.InternetGatewayId, err
}

func byoVpcPublicSubnetSetup(vpcId string, availabilityZone string, cidrBlock string) (string, error) {

	publicSubnet := &ec2.CreateSubnetInput{
		VpcId:            aws.String(vpcId),
		CidrBlock:        aws.String(publicSubnetsCidrBlock),
		AvailabilityZone: aws.String(availabilityZone),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("subnet"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(ByoVpcName + "-PubSub"),
					},
				},
			},
		},
	}

	result, err := CcsAwsSession.ec2.CreateSubnet(publicSubnet)
	if err != nil {
		log.Printf("Error creating Public Subnet: %v", err)
		return "", err
	} else {
		log.Printf("Created Public Subnet %s", *result.Subnet.SubnetId)
	}

	return *result.Subnet.SubnetId, err
}

func byoVpcPrivateSubnetSetup(vpcId string, availabilityZone string, cidrBlock string) (string, error) {

	privateSubnet := &ec2.CreateSubnetInput{
		VpcId:            aws.String(vpcId),
		CidrBlock:        aws.String(privateSubnetsCidrBlock),
		AvailabilityZone: aws.String(availabilityZone),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("subnet"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(ByoVpcName + "-PrvSub"),
					},
				},
			},
		},
	}

	result, err := CcsAwsSession.ec2.CreateSubnet(privateSubnet)
	if err != nil {
		log.Printf("Error creating Private Subnet: %v", err)
		return "", err
	} else {
		log.Printf("Created Private Subnet %s", *result.Subnet.SubnetId)
	}

	return *result.Subnet.SubnetId, err
}

func byoVpcElasticIpSetup() (string, error) {

	input := &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	}

	result, err := CcsAwsSession.ec2.AllocateAddress(input)
	if err != nil {
		log.Printf("Error allocating EIP: %v", err)
		return "", err
	} else {
		log.Printf("Allocated EIP %s", *result.AllocationId)
	}

	return *result.AllocationId, err
}

func byoVpcNatGatewaySetup(vpcId string, publicSubnetId string, elasticIp string) (string, error) {

	input := &ec2.CreateNatGatewayInput{
		SubnetId:     aws.String(publicSubnetId),
		AllocationId: aws.String(elasticIp),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("natgateway"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(ByoVpcName + "-NAT"),
					},
				},
			},
		},
	}

	result, err := CcsAwsSession.ec2.CreateNatGateway(input)
	if err != nil {
		log.Printf("Error creating NAT Gateway: %v", err)
		return "", err
	} else {
		CcsAwsSession.ec2.WaitUntilNatGatewayAvailable(&ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []*string{result.NatGateway.NatGatewayId},
		})
		log.Printf("Created NAT Gateway %s", *result.NatGateway.NatGatewayId)
	}

	return *result.NatGateway.NatGatewayId, err
}

func byoVpcCreatePublicRouteTable(vpcId string) (string, error) {

	input := &ec2.CreateRouteTableInput{
		VpcId: aws.String(vpcId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("route-table"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(ByoVpcName + "-PubRoute"),
					},
				},
			},
		},
	}

	result, err := CcsAwsSession.ec2.CreateRouteTable(input)
	if err != nil {
		log.Printf("Error creating Public Route Table: %v", err)
		return "", err
	} else {
		log.Printf("Created Public Route Table %s", *result.RouteTable.RouteTableId)
	}

	return *result.RouteTable.RouteTableId, err
}

func byoVpcAssociatePublicTable(publicRouteTableId string, publicSubnetId string) error {

	input := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(publicRouteTableId),
		SubnetId:     aws.String(publicSubnetId),
	}

	result, err := CcsAwsSession.ec2.AssociateRouteTable(input)
	if err != nil {
		log.Printf("Error associating Public Route Table: %v", err)
		return err
	} else {
		log.Printf("Associated Public Route Table %s", *result.AssociationId)
	}

	return err
}

func byoVpcCreatePublicRoute(publicRouteTableId string, internetGatewayId string) error {

	input := &ec2.CreateRouteInput{
		RouteTableId:         aws.String(publicRouteTableId),
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            aws.String(internetGatewayId),
	}

	_, err := CcsAwsSession.ec2.CreateRoute(input)
	if err != nil {
		log.Printf("Error creating Public Route: %v", err)
		return err
	} else {
		log.Printf("Created Public Route")
	}

	return err
}

func byoVpcCreatePrivateRouteTable(vpcId string) (string, error) {

	input := &ec2.CreateRouteTableInput{
		VpcId: aws.String(vpcId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("route-table"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(ByoVpcName + "-PrvRoute"),
					},
				},
			},
		},
	}

	result, err := CcsAwsSession.ec2.CreateRouteTable(input)
	if err != nil {
		log.Printf("Error creating Private Route Table: %v", err)
		return "", err
	} else {
		log.Printf("Created Private Route Table %s", *result.RouteTable.RouteTableId)
	}

	return *result.RouteTable.RouteTableId, err
}

func byoVpcAssociatePrivateTable(privateRouteTableId string, privateSubnetId string) error {

	input := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(privateRouteTableId),
		SubnetId:     aws.String(privateSubnetId),
	}

	result, err := CcsAwsSession.ec2.AssociateRouteTable(input)
	if err != nil {
		log.Printf("Error associating Private Route Table: %v", err)
		return err
	} else {
		log.Printf("Associated Private Route Table %s", *result.AssociationId)
	}

	return err
}

func byoVpcCreatePrivateRoute(privateRouteTableId string, natGatewayId string) error {

	input := &ec2.CreateRouteInput{
		RouteTableId:         aws.String(privateRouteTableId),
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		NatGatewayId:         aws.String(natGatewayId),
	}

	_, err := CcsAwsSession.ec2.CreateRoute(input)
	if err != nil {
		log.Printf("Error creating Private Route: %v", err)
		return err
	} else {
		log.Printf("Created Private Route")
	}

	return err
}

func byoVpcGetVpcSubnetIds(vpcId string) ([]string, error) {

	input := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcId)},
			},
		},
	}

	result, err := CcsAwsSession.ec2.DescribeSubnets(input)
	if err != nil {
		log.Printf("Error getting Subnets: %v", err)
		return nil, err
	} else {
		log.Printf("Got Subnets %s", result)
	}

	subnetIds := make([]string, len(result.Subnets))
	for i, subnet := range result.Subnets {
		subnetIds[i] = *subnet.SubnetId
	}

	return subnetIds, err
}

//Deletes all the resources created by the BYO VPC
func ByoVpcCleanUp() []error {
	var errs []error

	//Delete Private Route Table
	inputRouteTable := &ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(privateRouteTableId),
	}
	_, err := CcsAwsSession.ec2.DeleteRouteTable(inputRouteTable)
	if err != nil {
		log.Printf("Error deleting Private Route Table: %v", err)
		errs = append(errs, err)
	} else {
		log.Printf("Deleted Private Route Table %s", privateRouteTableId)
	}

	//Delete Public Route Table
	inputRouteTable = &ec2.DeleteRouteTableInput{
		RouteTableId: aws.String(publicRouteTableId),
	}
	_, err = CcsAwsSession.ec2.DeleteRouteTable(inputRouteTable)
	if err != nil {
		log.Printf("Error deleting Public Route Table: %v", err)
		errs = append(errs, err)
	} else {
		log.Printf("Deleted Public Route Table %s", publicRouteTableId)
	}

	//Delete NAT Gateway
	inputNatGateway := &ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(natGatewayId),
	}
	_, err = CcsAwsSession.ec2.DeleteNatGateway(inputNatGateway)
	if err != nil {
		log.Printf("Error deleting NAT Gateway: %v", err)
		errs = append(errs, err)
	} else {
		log.Printf("Deleted NAT Gateway %s", natGatewayId)
	}

	//Delete Elastic IP
	inputElasticIp := &ec2.ReleaseAddressInput{
		AllocationId: aws.String(elasticIpId),
	}
	_, err = CcsAwsSession.ec2.ReleaseAddress(inputElasticIp)
	if err != nil {
		log.Printf("Error deleting Elastic IP: %v", err)
		errs = append(errs, err)
	} else {
		log.Printf("Deleted Elastic IP %s", elasticIpId)
	}

	//Delete Private Subnet
	inputSubnet := &ec2.DeleteSubnetInput{
		SubnetId: aws.String(privateSubnetId),
	}
	_, err = CcsAwsSession.ec2.DeleteSubnet(inputSubnet)
	if err != nil {
		log.Printf("Error deleting Private Subnet: %v", err)
		errs = append(errs, err)
	} else {
		log.Printf("Deleted Private Subnet %s", privateSubnetId)
	}

	//Delete Public Subnet
	inputSubnet = &ec2.DeleteSubnetInput{
		SubnetId: aws.String(publicSubnetId),
	}
	_, err = CcsAwsSession.ec2.DeleteSubnet(inputSubnet)
	if err != nil {
		log.Printf("Error deleting Public Subnet: %v", err)
		errs = append(errs, err)
	} else {
		log.Printf("Deleted Public Subnet %s", publicSubnetId)
	}

	//Delete VPC
	inputVpc := &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcId),
	}
	_, err = CcsAwsSession.ec2.DeleteVpc(inputVpc)
	if err != nil {
		log.Printf("Error deleting VPC: %v", err)
		errs = append(errs, err)
	} else {
		log.Printf("Deleted VPC %s", vpcId)
	}

	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("Error: %v", err)
		}

		return errs
	}

	return nil
}
