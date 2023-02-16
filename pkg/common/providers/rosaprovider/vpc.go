package rosaprovider

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/openshift/osde2e/assets"
	viper "github.com/openshift/osde2e/pkg/common/concurrentviper"
	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/terraform"
)

type HyperShiftVPC struct {
	PrivateSubnet     string
	PublicSubnet      string
	NodePrivateSubnet string
}

// copyFile copies the srcFile provided to the destFile
func copyFile(srcFile string, destFile string) error {
	srcReader, err := assets.FS.Open(srcFile)
	if err != nil {
		return fmt.Errorf("error opening %s file: %w", srcFile, err)
	}
	defer srcReader.Close()

	destReader, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("error creating runtime %s file: %w", destFile, err)
	}
	defer destReader.Close()

	_, err = io.Copy(destReader, srcReader)
	if err != nil {
		return fmt.Errorf("error copying source file to destination file: %w", err)
	}

	return nil
}

// createHyperShiftVPC creates the vpc to provision HyperShift clusters
func createHyperShiftVPC() (*HyperShiftVPC, error) {
	ctx := context.Background()

	var vpc HyperShiftVPC
	workingDir := viper.GetString(config.ReportDir)

	log.Println("Creating ROSA HyperShift aws vpc")

	err := copyFile("terraform/setup-vpc.tf", fmt.Sprintf("%s/setup-vpc.tf", workingDir))
	if err != nil {
		return nil, err
	}

	tf, err := terraform.New(workingDir)
	if err != nil {
		return nil, err
	}

	err = tf.Init(ctx)
	if err != nil {
		return nil, err
	}

	err = tf.Plan(
		ctx,
		tfexec.Var(fmt.Sprintf("aws_region=%s", viper.GetString(config.AWSRegion))),
		tfexec.Var(fmt.Sprintf("cluster_name=%s", viper.GetString(config.Cluster.Name))),
	)
	if err != nil {
		return &vpc, err
	}

	err = tf.Apply(ctx)
	if err != nil {
		return nil, err
	}

	output, err := tf.Output(ctx)
	if err != nil {
		return nil, err
	}

	vpc.PrivateSubnet = strings.ReplaceAll(string(output["cluster-private-subnet"].Value), "\"", "")
	vpc.PublicSubnet = strings.ReplaceAll(string(output["cluster-public-subnet"].Value), "\"", "")
	vpc.NodePrivateSubnet = strings.ReplaceAll(string(output["node-private-subnet"].Value), "\"", "")

	log.Println("ROSA HyperShift aws vpc created!")

	return &vpc, nil
}

// deleteHyperShiftVPC deletes the vpc created to provision HyperShift clusters
func deleteHyperShiftVPC(workingDir string) error {
	ctx := context.Background()

	log.Println("Deleting ROSA HyperShift aws vpc")

	tf, err := terraform.New(workingDir)
	if err != nil {
		return err
	}

	err = tf.Destroy(ctx)
	if err != nil {
		return err
	}

	log.Println("ROSA HyperShift aws vpc deleted!")

	err = tf.Uninstall(ctx)
	if err != nil {
		return err
	}

	return nil
}
