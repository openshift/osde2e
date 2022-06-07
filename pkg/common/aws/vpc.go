package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// createVpc creates a VPC with the given CIDR block and tags it with the given tagSpecification
func CreateVpc() (string, error) {
	var err error

	// Create a new VPC
	input := &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("vpc"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("BYO-VPC"),
					},
				},
			},
		},
	}

	VerifyCCS()
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

	return *result.Vpc.VpcId, nil
}
