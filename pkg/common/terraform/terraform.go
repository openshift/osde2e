package terraform

import (
	"context"
	"fmt"

	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type runner struct {
	Context    context.Context
	Installer  *releases.LatestVersion
	Runner     *tfexec.Terraform
	WorkingDir string
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
		Context:    context.Background(),
		Installer:  &installer,
		Runner:     tf,
		WorkingDir: workingDir,
	}, err
}

// Uninstalls the terraform instance installed at runtime.
func (r *runner) Uninstall() error {
	err := r.Installer.Remove(context.Background())
	if err != nil {
		return fmt.Errorf("error uninstalling terraform: %w", err)
	}

	return nil
}

// Init performs a terraform init using the provided TerraformRunner receiver.
func (r *runner) Init() error {
	err := r.Runner.Init(r.Context)
	if err != nil {
		return fmt.Errorf("error running terraform init: %w", err)
	}

	return nil
}

// Plan performs a terraform plan using the provided TerraformRunner receiver.
func (r *runner) Plan(args ...tfexec.PlanOption) error {
	args = append(args, tfexec.Out(fmt.Sprintf("%s/plan.out", r.WorkingDir)))

	_, err := r.Runner.Plan(
		r.Context,
		args...,
	)
	if err != nil {
		return fmt.Errorf("error running terraform plan: %s", err)
	}

	return nil
}

// Apply performs a terraform apply using the provided TerraformRunner receiver.
func (r *runner) Apply() error {
	err := r.Runner.Apply(
		r.Context,
		tfexec.DirOrPlan(fmt.Sprintf("%s/plan.out", r.WorkingDir)),
	)
	if err != nil {
		return fmt.Errorf("error running terraform apply: %w", err)
	}

	return nil
}

// Destroy performs a terraform destroy using the provided TerraformRunner receiver.
func (r *runner) Destroy() error {
	err := r.Runner.Destroy(r.Context)
	if err != nil {
		return fmt.Errorf("error running terraform destroy: %w", err)
	}

	return nil
}

// Output performs a terraform output using the provided TerraformRunner receiver.
func (r *runner) Output() (map[string]tfexec.OutputMeta, error) {
	var output map[string]tfexec.OutputMeta

	output, err := r.Runner.Output(r.Context)
	if err != nil {
		return output, fmt.Errorf("error running terraform output: %w", err)
	}

	return output, nil
}
