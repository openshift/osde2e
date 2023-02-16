package terraform

import (
	"context"
	"fmt"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type runner struct {
	installer  *releases.LatestVersion
	runner     *tfexec.Terraform
	workingDir string
}

// New creates a terraform runner for the provided working directory where
// terraform files are. It returns a TerraformRunner struct which has the
// runner to perform any commands terraform-exec package provides.
func New(workingDir string) (*runner, error) {
	installer := releases.LatestVersion{
		Product:    product.Terraform,
		InstallDir: "/tmp",
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error installing terraform: %w", err)
	}

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return nil, fmt.Errorf("error configuring terraform: %w", err)
	}

	return &runner{
		installer:  &installer,
		runner:     tf,
		workingDir: workingDir,
	}, err
}

// Uninstalls the terraform instance installed at runtime.
func (r *runner) Uninstall(ctx context.Context) error {
	err := r.installer.Remove(ctx)
	if err != nil {
		return fmt.Errorf("error uninstalling terraform: %w", err)
	}

	return nil
}

// Init performs a terraform init using the provided TerraformRunner receiver.
func (r *runner) Init(ctx context.Context) error {
	err := r.runner.Init(ctx)
	if err != nil {
		return fmt.Errorf("error running terraform init: %w", err)
	}

	return nil
}

// Plan performs a terraform plan using the provided TerraformRunner receiver.
func (r *runner) Plan(ctx context.Context, args ...tfexec.PlanOption) error {
	args = append(args, tfexec.Out(fmt.Sprintf("%s/plan.out", r.workingDir)))

	_, err := r.runner.Plan(
		ctx,
		args...,
	)
	if err != nil {
		return fmt.Errorf("error running terraform plan: %s", err)
	}

	return nil
}

// Apply performs a terraform apply using the provided TerraformRunner receiver.
func (r *runner) Apply(ctx context.Context) error {
	err := r.runner.Apply(
		ctx,
		tfexec.DirOrPlan(fmt.Sprintf("%s/plan.out", r.workingDir)),
	)
	if err != nil {
		return fmt.Errorf("error running terraform apply: %w", err)
	}

	return nil
}

// Destroy performs a terraform destroy using the provided TerraformRunner receiver.
func (r *runner) Destroy(ctx context.Context) error {
	err := r.runner.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("error running terraform destroy: %w", err)
	}

	return nil
}

// Output performs a terraform output using the provided TerraformRunner receiver.
func (r *runner) Output(ctx context.Context) (map[string]tfexec.OutputMeta, error) {
	output, err := r.runner.Output(ctx)
	if err != nil {
		return output, fmt.Errorf("error running terraform output: %w", err)
	}

	return output, nil
}
