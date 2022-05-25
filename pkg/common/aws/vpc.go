package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type awsVpc struct {
	cidrBblock string
	tags       []*ec2.Tag
}

// VpcId is the id of the AWS VPC that will get passed to OCM for the BYO-VPC work flow
var VpcId string

// createVpc creates a VPC with the given CIDR block and tags it with the given tagSpecification
func createVpc(cidrBlock string, tags string) (string, error) {
	var err error

	// Create a new VPC
	input := &ec2.CreateVpcInput{
		CidrBlock: aws.String(cidrBlock),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("vpc"),
				Tags: []*ec2.Tag{
					{
						Key: aws.String("Name"),
						//Need to add the cluster name here?
						Value: aws.String("BYO-VPC"),
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

	return "hello", nil
}
