/*
Copyright 2026 The TabTabAI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"os"

	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/klog/v2"
)

type PullSecret struct {
	Name string `json:"name" yaml:"name"`
}
type ACKCluster struct {
	Enabled                  bool              `json:"enabled" yaml:"enabled"`
	PodAdditionalLabels      map[string]string `json:"podAdditionalLabels" yaml:"podAdditionalLabels"`
	PodAdditionalAnnotations map[string]string `json:"podAdditionalAnnotations" yaml:"podAdditionalAnnotations"`

	WarnupImage string `json:"warnupImage" yaml:"warnupImage"`
	WarnupCount int    `json:"warnupCount" yaml:"warnupCount"`
}

type Redis struct {
	Addr     string `json:"addr" yaml:"addr"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
}

// Gateway holds OpenClaw gateway settings used during instance initialisation.
// Init holds LiteLLM integration settings.
type Init struct {
	LiteLLM LiteLLM `json:"litellm" yaml:"litellm"`
}

type LiteLLM struct {
	BaseURL   string `json:"baseurl" yaml:"baseurl"`
	MasterKey string `json:"master_key" yaml:"master_key"`
	// DefaultTeam is the default team ID assigned to new LiteLLM users
	DefaultTeam string `json:"default_team" yaml:"default_team"`
	// DefaultMaxBudget is the default maximum budget (in USD) applied to each generated key/user.
	// 0 means unlimited.
	DefaultMaxBudget float64 `json:"default_max_budget" yaml:"default_max_budget"`
	// DefaultBudgetDuration is the reset period for the budget (e.g. "30d", "1h").
	// Empty means no reset.
	DefaultBudgetDuration string `json:"default_budget_duration" yaml:"default_budget_duration"`
	// Api Key should be configured when allocating openclaw to user.
	Models struct {
		Lite struct {
			ModelID string `json:"modelID" yaml:"modelID"`
		} `json:"lite" yaml:"lite"`
		Pro struct {
			ModelID string `json:"modelID" yaml:"modelID"`
		} `json:"pro" yaml:"pro"`
	}
}

// Server holds HTTP API server settings.
type Server struct {
	Port int `json:"port" yaml:"port"`
	// JWTSecret is used to sign JWT tokens. If empty, a random secret is generated on startup
	// (tokens will be invalidated on restart). Set this to a stable value in production.
	JWTSecret string `json:"jwt_secret" yaml:"jwt_secret"`
	// DataDir is the directory used for persistent data (SQLite DB, JWT secret file).
	// Created automatically if absent. Defaults to ".claw-swarm".
	DataDir string `json:"data_dir" yaml:"data_dir"`
}

type DB struct {
	// sqlite db file
	File string `json:"file" yaml:"file"`
}

// K8sConfig holds Kubernetes-specific adapter settings.
type K8sConfig struct {
	// Kubeconfig path; empty means in-cluster or default kubeconfig rules.
	Kubeconfig string `json:"kubeconfig" yaml:"kubeconfig"`
	// Namespace to manage claw instances in; falls back to POD_NAMESPACE env var then "default".
	Namespace string `json:"namespace" yaml:"namespace"`
}

// Adapter holds adapter backend settings.
type Adapter struct {
	// Type selects the adapter backend. Default "k8s".
	Type string `json:"type" yaml:"type"`
	// K8s holds Kubernetes-specific settings (used when type = "k8s").
	K8s K8sConfig `json:"k8s" yaml:"k8s"`
}

type Config struct {
	Server  Server  `json:"server" yaml:"server"`
	Init    Init    `json:"init" yaml:"init"`
	Adapter Adapter `json:"adapter" yaml:"adapter"`

	DB DB `json:"db" yaml:"db"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Server: Server{
			Port:    8088,
			DataDir: ".claw-swarm",
		},
		Adapter: Adapter{
			Type: "k8s",
		},
		Init: Init{
			LiteLLM: func() LiteLLM {
				l := LiteLLM{BaseURL: "http://localhost:4000"}
				l.Models.Lite.ModelID = "tabtab-lite"
				l.Models.Pro.ModelID = "tabtab-pro"
				return l
			}(),
		},
	}
}

// ReadConf reads and parses a YAML config file.
// If the file does not exist, defaults are returned without error.
// If the file exists but cannot be parsed, the program panics.
func ReadConf(path string) Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			klog.V(4).Infof("config file %s not found, using defaults", path)
			return cfg
		}
		panic(err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
