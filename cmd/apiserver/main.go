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

// cmd/apiserver is the standalone HTTP API Server binary.
// The backend adapter is selected by cfg.Adapter.Type ("k8s" or "docker").
// It exposes the full /auth, /users, /audit, /claw REST API and can run
// as multiple replicas behind a load balancer.
package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	adapterfactory "gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/adapter/factory"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/apiserver"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/apiserver/service"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/audit"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/auth"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/config"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/db"
	"gitlab.botnow.cn/agentic/claw-swarm-operator/pkg/utils"
)

var setupLog = ctrl.Log.WithName("setup")

func main() {
	var (
		configFile string
		apiPort    int
		dataDir    string
	)

	// Parse --config early to seed flag defaults.
	flag.StringVar(&configFile, "config", "config.yaml", "Path to the YAML configuration file.")
	for i, arg := range os.Args {
		if (arg == "--config" || arg == "-config") && i+1 < len(os.Args) {
			configFile = os.Args[i+1]
			break
		}
		if strings.HasPrefix(arg, "--config=") {
			configFile = strings.SplitN(arg, "=", 2)[1]
			break
		}
	}
	cfg := config.ReadConf(configFile)

	flag.IntVar(&apiPort, "port", cfg.Server.Port, "HTTP API server listen port")
	flag.StringVar(&dataDir, "data-dir", cfg.Server.DataDir,
		"Directory for persistent data (SQLite DB and JWT secret). Created automatically if absent.")

	opts := zap.Options{Development: true}
	klog.InitFlags(nil)
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// ── Data directory & DB ───────────────────────────────────────────────
	if err := utils.EnsureDataDir(dataDir); err != nil {
		setupLog.Error(err, "unable to create data directory", "dir", dataDir)
		os.Exit(1)
	}
	if cfg.DB.File == "" {
		cfg.DB.File = filepath.Join(dataDir, "claw.db")
	}
	db.InitDB(cfg)

	// ── User store ────────────────────────────────────────────────────────
	userStore := auth.NewUserStore(db.DB())
	if err := userStore.InitSchema(); err != nil {
		setupLog.Error(err, "unable to initialize user schema")
		os.Exit(1)
	}
	if err := userStore.MigrateSchema(); err != nil {
		setupLog.Error(err, "unable to migrate user schema")
		os.Exit(1)
	}
	if err := userStore.EnsureAdminUser(); err != nil {
		setupLog.Error(err, "unable to ensure admin user")
		os.Exit(1)
	}

	// ── JWT secret ────────────────────────────────────────────────────────
	jwtSecret := cfg.Server.JWTSecret
	if jwtSecret == "" {
		var err error
		jwtSecret, err = utils.LoadOrCreateJWTSecret(dataDir)
		if err != nil {
			setupLog.Error(err, "unable to load or create JWT secret")
			os.Exit(1)
		}
		setupLog.Info("JWT secret loaded", "path", filepath.Join(dataDir, ".jwt_secret"))
	}

	// ── Audit store ───────────────────────────────────────────────────────
	auditStore := audit.NewAuditStore(db.DB())
	if err := auditStore.InitSchema(); err != nil {
		setupLog.Error(err, "unable to initialize audit schema")
		os.Exit(1)
	}

	// ── Claw Adapter ──────────────────────────────────────────────────────
	// Type is determined by cfg.Adapter.Type (default: "k8s").
	clawAdapter, err := adapterfactory.New(cfg)
	if err != nil {
		setupLog.Error(err, "unable to create claw adapter", "runtime", cfg.Adapter.Type)
		os.Exit(1)
	}
	setupLog.Info("claw adapter initialized", "runtime", cfg.Adapter.Type)

	// ── Service + HTTP Server ─────────────────────────────────────────────
	clawSvc := service.New(clawAdapter, auditStore)
	srv := apiserver.NewServer(userStore, auditStore, jwtSecret, clawSvc)

	setupLog.Info("starting apiserver", "port", apiPort)
	ctx := ctrl.SetupSignalHandler()
	srv.Run(ctx, apiPort) // blocks until signal
}
