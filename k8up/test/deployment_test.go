package test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
)

var (
	tplDeployment = []string{"templates/deployment.yaml"}
)

func Test_Deployment_ShouldRender_EnvironmentVariables(t *testing.T) {
	wantRepo := "repository"
	wantTag := "tag"
	wantVar := "BACKUP_IMAGE"
	options := &helm.Options{
		ValuesFiles: []string{"testdata/deployment_1.yaml"},
		SetValues: map[string]string{
			"k8up.backupImage.repository": wantRepo,
			"k8up.backupImage.tag": wantTag,
		},
	}

	got := renderDeployment(t, options, false)

	envs := got.Spec.Template.Spec.Containers[0].Env
	assert.Equalf(t, wantVar, envs[0].Name, "Deployment does not use required Env %s", wantVar)
	assert.Equalf(t, wantRepo+":"+wantTag, envs[0].Value, "Deployment does not use required Env Value from %s", wantVar)
	assert.Equal(t, "VARIABLE", envs[1].Name, "Deployment does not use configured Env Name")
	assert.Equal(t, "VALUE", envs[1].Value, "Deployment does not use configured Env Value")
}

func Test_Deployment_ShouldRender_Affinity(t *testing.T) {
	options := &helm.Options{
		ValuesFiles: []string{"testdata/deployment_1.yaml"},
	}

	got := renderDeployment(t, options, false)

	host := got.Spec.Template.Spec.
		Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values[0]
	assert.Equal(t, "host", host, "Deployment does not render configured host affinity")
}

func Test_Deployment_ShouldRender_DefaultServiceAccount(t *testing.T) {
	want := releaseName + "-k8up"
	options := &helm.Options{}

	got := renderDeployment(t, options, false)
	serviceName := got.Spec.Template.Spec.ServiceAccountName
	assert.Equal(t, want, serviceName, "Deployment does not render configured serviceName")
}

func Test_Deployment_ShouldRender_CustomServiceAccount(t *testing.T) {
	want := "test"
	options := &helm.Options{
		SetValues: map[string]string{
			"serviceAccount.name": want,
		},
	}

	got := renderDeployment(t, options, false)

	serviceName := got.Spec.Template.Spec.ServiceAccountName
	assert.Equal(t, want, serviceName, "Deployment does not render configured serviceName")
}

func Test_Deployment_ShouldRender_DefaultResources(t *testing.T) {
	want := "1Gi"
	options := &helm.Options{
		SetValues: map[string]string{
			"resources.limits.memory": want,
		},
	}

	got := renderDeployment(t, options, false)
	resources := got.Spec.Template.Spec.Containers[0].Resources
	assert.Equal(t, want, resources.Limits.Memory().String(), "Deployment does not render configured memory limit")
}

func Test_Deployment_ShouldRender_Labels(t *testing.T) {
	options := &helm.Options{}

	got := renderDeployment(t, options, false)

	selector := got.Spec.Selector.MatchLabels
	matchLabels := got.Spec.Template.Labels
	assert.Equal(t, selector, matchLabels, "Deployment does not render matching labels")
}

func renderDeployment(t *testing.T, options *helm.Options, wantErr bool) *appv1.Deployment {
	output, err := helm.RenderTemplateE(t, options, helmChartPath, releaseName, tplDeployment)
	if wantErr {
		require.Error(t, err)
		return nil
	}
	require.NoError(t, err)
	deployment := appv1.Deployment{}
	helm.UnmarshalK8SYaml(t, output, &deployment)
	return &deployment
}