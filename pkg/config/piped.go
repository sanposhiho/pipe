// Copyright 2020 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pipe-cd/pipe/pkg/model"
)

var DefaultKubernetesCloudProvider = PipedCloudProvider{
	Name:             "kubernetes-default",
	Type:             model.CloudProviderKubernetes,
	KubernetesConfig: &CloudProviderKubernetesConfig{},
}

// PipedSpec contains configurable data used to while running Piped.
type PipedSpec struct {
	// The identifier of the PipeCD project where this piped belongs to.
	ProjectID string
	// The unique identifier generated for this piped.
	PipedID string
	// The path to the file containing the generated Key string for this piped.
	PipedKeyFile string
	// The address used to connect to the control-plane's API.
	APIAddress string `json:"apiAddress"`
	// The address to the control-plane's Web.
	WebAddress string `json:"webAddress"`
	// How often to check whether an application should be synced.
	// Default is 1m.
	SyncInterval Duration `json:"syncInterval"`
	// Git configuration needed for git commands.
	Git PipedGit `json:"git"`
	// List of git repositories this piped will handle.
	Repositories []PipedRepository `json:"repositories"`
	// List of helm chart repositories that should be added while starting up.
	ChartRepositories []HelmChartRepository `json:"chartRepositories"`
	// List of cloud providers can be used by this piped.
	CloudProviders []PipedCloudProvider `json:"cloudProviders"`
	// List of analysis providers can be used by this piped.
	AnalysisProviders []PipedAnalysisProvider `json:"analysisProviders"`
	// List of image providers can be used by this piped.
	ImageProviders []PipedImageProvider `json:"imageProviders"`
	// Sending notification to Slack, Webhook…
	Notifications Notifications `json:"notifications"`
	// How the sealed secret should be managed.
	SealedSecretManagement *SealedSecretManagement `json:"sealedSecretManagement"`
}

// Validate validates configured data of all fields.
func (s *PipedSpec) Validate() error {
	if s.ProjectID == "" {
		return fmt.Errorf("projectID must be set")
	}
	if s.PipedID == "" {
		return fmt.Errorf("pipedID must be set")
	}
	if s.PipedKeyFile == "" {
		return fmt.Errorf("pipedKeyFile must be set")
	}
	if s.APIAddress == "" {
		return fmt.Errorf("apiAddress must be set")
	}
	if s.WebAddress == "" {
		return fmt.Errorf("webAddress must be set")
	}
	if s.SyncInterval < 0 {
		s.SyncInterval = Duration(time.Minute)
	}
	if s.SealedSecretManagement != nil {
		if err := s.SealedSecretManagement.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// EnableDefaultKubernetesCloudProvider adds the default kubernetes cloud provider if it was not specified.
func (s *PipedSpec) EnableDefaultKubernetesCloudProvider() {
	for _, cp := range s.CloudProviders {
		if cp.Name == DefaultKubernetesCloudProvider.Name {
			return
		}
	}
	s.CloudProviders = append(s.CloudProviders, DefaultKubernetesCloudProvider)
}

// HasCloudProvider checks whether the given provider is configured or not.
func (s *PipedSpec) HasCloudProvider(name string, t model.CloudProviderType) bool {
	for _, cp := range s.CloudProviders {
		if cp.Name != name {
			continue
		}
		if cp.Type != t {
			continue
		}
		return true
	}
	return false
}

// FindCloudProvider finds and returns a Cloud Provider by name and type.
func (s *PipedSpec) FindCloudProvider(name string, t model.CloudProviderType) (PipedCloudProvider, bool) {
	for _, p := range s.CloudProviders {
		if p.Name != name {
			continue
		}
		if p.Type != t {
			continue
		}
		return p, true
	}
	return PipedCloudProvider{}, false
}

// GetRepositoryMap returns a map of repositories where key is repo id.
func (s *PipedSpec) GetRepositoryMap() map[string]PipedRepository {
	m := make(map[string]PipedRepository, len(s.Repositories))
	for _, repo := range s.Repositories {
		m[repo.RepoID] = repo
	}
	return m
}

// GetRepository finds a repository with the given ID from the configured list.
func (s *PipedSpec) GetRepository(id string) (PipedRepository, bool) {
	for _, repo := range s.Repositories {
		if repo.RepoID == id {
			return repo, true
		}
	}
	return PipedRepository{}, false
}

// GetAnalysisProvider finds and returns an Analysis Provider config whose name is the given string.
func (s *PipedSpec) GetAnalysisProvider(name string) (PipedAnalysisProvider, bool) {
	for _, p := range s.AnalysisProviders {
		if p.Name == name {
			return p, true
		}
	}
	return PipedAnalysisProvider{}, false
}

type PipedGit struct {
	// The username that will be configured for `git` user.
	Username string `json:"username"`
	// The email that will be configured for `git` user.
	Email string `json:"email"`
	// Where to write ssh config file.
	// Default is "/home/pipecd/.ssh/config".
	SSHConfigFilePath string `json:"sshConfigFilePath"`
	// The host name.
	// e.g. github.com, gitlab.com
	// Default is "github.com".
	Host string `json:"host"`
	// The hostname or IP address of the remote git server.
	// e.g. github.com, gitlab.com
	// Default is the same value with Host.
	HostName string `json:"hostName"`
	// The path to the private ssh key file.
	// This will be used to clone the source code of the specified git repositories.
	SSHKeyFile string `json:"sshKeyFile"`
}

func (g PipedGit) ShouldConfigureSSHConfig() bool {
	return g.SSHKeyFile != ""
}

type PipedRepository struct {
	// Unique identifier for this repository.
	// This must be unique in the piped scope.
	RepoID string `json:"repoId"`
	// Remote address of the repository used to clone the source code.
	// e.g. git@github.com:org/repo.git
	Remote string `json:"remote"`
	// The branch will be handled.
	Branch string `json:"branch"`
}

type HelmChartRepository struct {
	// The name of the Helm chart repository.
	Name string `json:"name"`
	// The address to the Helm chart repository.
	Address string `json:"address"`
	// Username used for the repository backed by HTTP basic authentication.
	Username string `json:"username"`
	// Password used for the repository backed by HTTP basic authentication.
	Password string `json:"password"`
}

type PipedCloudProvider struct {
	Name string
	Type model.CloudProviderType

	KubernetesConfig *CloudProviderKubernetesConfig
	TerraformConfig  *CloudProviderTerraformConfig
	CloudRunConfig   *CloudProviderCloudRunConfig
	LambdaConfig     *CloudProviderLambdaConfig
}

type genericPipedCloudProvider struct {
	Name   string                  `json:"name"`
	Type   model.CloudProviderType `json:"type"`
	Config json.RawMessage         `json:"config"`
}

func (p *PipedCloudProvider) UnmarshalJSON(data []byte) error {
	var err error
	gp := genericPipedCloudProvider{}
	if err = json.Unmarshal(data, &gp); err != nil {
		return err
	}
	p.Name = gp.Name
	p.Type = gp.Type

	switch p.Type {
	case model.CloudProviderKubernetes:
		p.KubernetesConfig = &CloudProviderKubernetesConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.KubernetesConfig)
		}
	case model.CloudProviderTerraform:
		p.TerraformConfig = &CloudProviderTerraformConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.TerraformConfig)
		}
	case model.CloudProviderCloudRun:
		p.CloudRunConfig = &CloudProviderCloudRunConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.CloudRunConfig)
		}
	case model.CloudProviderLambda:
		p.LambdaConfig = &CloudProviderLambdaConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.LambdaConfig)
		}
	default:
		err = fmt.Errorf("unsupported cloud provider type: %s", p.Name)
	}
	return err
}

type CloudProviderKubernetesConfig struct {
	// The master URL of the kubernetes cluster.
	// Empty means in-cluster.
	MasterURL string `json:"masterURL"`
	// The path to the kubeconfig file.
	// Empty means in-cluster.
	KubeConfigPath string `json:"kubeConfigPath"`
	// Configuration for application resource informer.
	AppStateInformer KubernetesAppStateInformer `json:"appStateInformer"`
}

type KubernetesAppStateInformer struct {
	// Only watches the specified namespace.
	// Empty means watching all namespaces.
	Namespace string `json:"namespace"`
	// List of resources that should be added to the watching targets.
	IncludeResources []KubernetesResourceMatcher `json:"includeResources"`
	// List of resources that should be ignored from the watching targets.
	ExcludeResources []KubernetesResourceMatcher `json:"excludeResources"`
}

type KubernetesResourceMatcher struct {
	// The APIVersion of the kubernetes resource.
	APIVersion string `json:"apiVersion"`
	// The kind name of the kubernetes resource.
	// Empty means all kinds are matching.
	Kind string `json:"kind"`
}

type CloudProviderTerraformConfig struct {
	// List of variables that will be set directly on terraform commands with "-var" flag.
	// The variable must be formatted by "key=value" as below:
	// "image_id=ami-abc123"
	// 'image_id_list=["ami-abc123","ami-def456"]'
	// 'image_id_map={"us-east-1":"ami-abc123","us-east-2":"ami-def456"}'
	Vars []string `json:"vars"`
}

type CloudProviderCloudRunConfig struct {
	// The GCP project hosting the CloudRun service.
	Project string `json:"project"`
	// The region of running CloudRun service.
	Region string `json:"region"`
	// The path to the service account file for accessing CloudRun service.
	CredentialsFile string `json:"credentialsFile"`
}

type CloudProviderLambdaConfig struct {
	Region string `json:"region"`
}

type PipedAnalysisProvider struct {
	Name string                     `json:"name"`
	Type model.AnalysisProviderType `json:"type"`

	PrometheusConfig  *AnalysisProviderPrometheusConfig  `json:"prometheus"`
	DatadogConfig     *AnalysisProviderDatadogConfig     `json:"datadog"`
	StackdriverConfig *AnalysisProviderStackdriverConfig `json:"stackdriver"`
}

type genericPipedAnalysisProvider struct {
	Name   string                     `json:"name"`
	Type   model.AnalysisProviderType `json:"type"`
	Config json.RawMessage            `json:"config"`
}

func (p *PipedAnalysisProvider) UnmarshalJSON(data []byte) error {
	var err error
	gp := genericPipedAnalysisProvider{}
	if err = json.Unmarshal(data, &gp); err != nil {
		return err
	}
	p.Name = gp.Name
	p.Type = gp.Type

	switch p.Type {
	case model.AnalysisProviderPrometheus:
		p.PrometheusConfig = &AnalysisProviderPrometheusConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.PrometheusConfig)
		}
	case model.AnalysisProviderDatadog:
		p.DatadogConfig = &AnalysisProviderDatadogConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.DatadogConfig)
		}
	case model.AnalysisProviderStackdriver:
		p.StackdriverConfig = &AnalysisProviderStackdriverConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.StackdriverConfig)
		}
	default:
		err = fmt.Errorf("unsupported analysis provider type: %s", p.Name)
	}
	return err
}

type AnalysisProviderPrometheusConfig struct {
	Address string `json:"address"`
	// The path to the username file.
	UsernameFile string `json:"usernameFile"`
	// The path to the password file.
	PasswordFile string `json:"passwordFile"`
}

type AnalysisProviderDatadogConfig struct {
	Address string `json:"address"`
	// The path to the api key file.
	APIKeyFile string `json:"apiKeyFile"`
	// The path to the application key file.
	ApplicationKeyFile string `json:"applicationKeyFile"`
}

type AnalysisProviderStackdriverConfig struct {
	// The path to the service account file.
	ServiceAccountFile string `json:"serviceAccountFile"`
}

type PipedImageProvider struct {
	Name string                  `json:"name"`
	Type model.ImageProviderType `json:"type"`
	// Default is five minute.
	PullInterval Duration `json:"pullInterval"`

	DockerhubConfig *ImageProviderDockerhubConfig
	GCRConfig       *ImageProviderGCRConfig
	ECRConfig       *ImageProviderECRConfig
}

type genericPipedImageProvider struct {
	Name         string                  `json:"name"`
	Type         model.ImageProviderType `json:"type"`
	PullInterval Duration                `json:"pullInterval"`

	Config json.RawMessage `json:"config"`
}

func (p *PipedImageProvider) UnmarshalJSON(data []byte) error {
	var err error
	gp := genericPipedImageProvider{}
	if err = json.Unmarshal(data, &gp); err != nil {
		return err
	}
	p.Name = gp.Name
	p.Type = gp.Type
	p.PullInterval = gp.PullInterval
	if p.PullInterval == 0 {
		p.PullInterval = Duration(time.Minute * 5)
	}

	switch p.Type {
	case model.ImageProviderTypeDockerhub:
		p.DockerhubConfig = &ImageProviderDockerhubConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.DockerhubConfig)
		}
	case model.ImageProviderTypeGCR:
		p.GCRConfig = &ImageProviderGCRConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.GCRConfig)
		}
	case model.ImageProviderTypeECR:
		p.ECRConfig = &ImageProviderECRConfig{}
		if len(gp.Config) > 0 {
			err = json.Unmarshal(gp.Config, p.ECRConfig)
		}
	default:
		err = fmt.Errorf("unsupported image provider type: %s", p.Name)
	}
	return err
}

type ImageProviderGCRConfig struct {
	Domain string `json:"domain"`
}

type ImageProviderDockerhubConfig struct {
	Username     string `json:"username"`
	PasswordFile string `json:"passwordFile"`
}

type ImageProviderECRConfig struct {
}

type Notifications struct {
	// List of notification routes.
	Routes []NotificationRoute `json:"routes"`
	// List of notification receivers.
	Receivers []NotificationReceiver `json:"receivers"`
}

type NotificationRoute struct {
	Name         string   `json:"name"`
	Receiver     string   `json:"receiver"`
	Events       []string `json:"events"`
	IgnoreEvents []string `json:"ignoreEvents"`
	Groups       []string `json:"groups"`
	IgnoreGroups []string `json:"ignoreGroups"`
	Apps         []string `json:"apps"`
	IgnoreApps   []string `json:"ignoreApps"`
	Envs         []string `json:"envs"`
	IgnoreEnvs   []string `json:"ignoreEnvs"`
}

type NotificationReceiver struct {
	Name    string                       `json:"name"`
	Slack   *NotificationReceiverSlack   `json:"slack"`
	Webhook *NotificationReceiverWebhook `json:"webhook"`
}

type NotificationReceiverSlack struct {
	HookURL string `json:"hookURL"`
}

type NotificationReceiverWebhook struct {
	URL string `json:"url"`
}

type SealedSecretManagement struct {
	// Which management service should be used.
	// Available values: SEALING_KEY, GCP_KMS, AWS_KMS
	Type model.SealedSecretManagementType `json:"type"`

	SealingKeyConfig *SealedSecretManagementSealingKey
	GCPKMSConfig     *SealedSecretManagementGCPKMS
}

func (m *SealedSecretManagement) Validate() error {
	switch m.Type {
	case model.SealedSecretManagementSealingKey:
		return m.SealingKeyConfig.Validate()
	case model.SealedSecretManagementGCPKMS:
		return m.GCPKMSConfig.Validate()
	default:
		return fmt.Errorf("unsupported sealed secret management type: %s", m.Type)
	}
}

type SealedSecretManagementSealingKey struct {
	// Configurable fields for SEALING_KEY.
	// The path to the private RSA key file.
	PrivateKeyFile string `json:"privateKeyFile"`
	// The path to the public RSA key file.
	PublicKeyFile string `json:"publicKeyFile"`
}

func (m *SealedSecretManagementSealingKey) Validate() error {
	if m.PrivateKeyFile == "" {
		return fmt.Errorf("privateKeyFile must be set")
	}
	if m.PublicKeyFile == "" {
		return fmt.Errorf("publicKeyFile must be set")
	}
	return nil
}

type SealedSecretManagementGCPKMS struct {
	// Configurable fields when using Google Cloud KMS.
	// The key name used for decrypting the sealed secret.
	KeyName string `json:"keyName"`
	// The path to the service account used to decrypt secret.
	DecryptServiceAccountFile string `json:"decryptServiceAccountFile"`
	// The path to the service account used to encrypt secret.
	EncryptServiceAccountFile string `json:"encryptServiceAccountFile"`
}

func (m *SealedSecretManagementGCPKMS) Validate() error {
	if m.KeyName == "" {
		return fmt.Errorf("keyName must be set")
	}
	if m.DecryptServiceAccountFile == "" {
		return fmt.Errorf("decryptServiceAccountFile must be set")
	}
	if m.EncryptServiceAccountFile == "" {
		return fmt.Errorf("encryptServiceAccountFile must be set")
	}
	return nil
}

type genericSealedSecretManagement struct {
	Type   model.SealedSecretManagementType `json:"type"`
	Config json.RawMessage                  `json:"config"`
}

func (p *SealedSecretManagement) UnmarshalJSON(data []byte) error {
	var err error
	g := genericSealedSecretManagement{}
	if err = json.Unmarshal(data, &g); err != nil {
		return err
	}
	p.Type = g.Type

	switch p.Type {
	case model.SealedSecretManagementSealingKey:
		p.SealingKeyConfig = &SealedSecretManagementSealingKey{}
		if len(g.Config) > 0 {
			err = json.Unmarshal(g.Config, p.SealingKeyConfig)
		}
	case model.SealedSecretManagementGCPKMS:
		p.GCPKMSConfig = &SealedSecretManagementGCPKMS{}
		if len(g.Config) > 0 {
			err = json.Unmarshal(g.Config, p.GCPKMSConfig)
		}
	default:
		err = fmt.Errorf("unsupported sealed secret management type: %s", p.Type)
	}
	return err
}
