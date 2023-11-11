/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package app

import (
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc"
	"k8s.io/apimachinery/pkg/util/net"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/spf13/cobra"
	zapOpt "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/fluid-cloudnative/fluid"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	efcctl "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/efc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	// Use compiler to check if the struct implements all the interface
	_ base.Implement = (*efc.EFCEngine)(nil)

	eventDriven             bool
	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	development             bool
	maxConcurrentReconciles int
	pprofAddr               string
	portRange               string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start efcruntime-controller in Kubernetes",
	Run: func(cmd *cobra.Command, args []string) {
		handle()
	},
}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = datav1alpha1.AddToScheme(scheme)

	startCmd.Flags().StringVarP(&metricsAddr, "metrics-addr", "", ":8080", "The address the metric endpoint binds to.")
	startCmd.Flags().BoolVarP(&enableLeaderElection, "enable-leader-election", "", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	startCmd.Flags().StringVarP(&leaderElectionNamespace, "leader-election-namespace", "", "fluid-system", "The namespace in which the leader election resource will be created.")
	startCmd.Flags().StringVarP(&pprofAddr, "pprof-addr", "", "", "The address for pprof to use while exporting profiling results")
	startCmd.Flags().BoolVarP(&development, "development", "", true, "Enable development mode for fluid controller.")
	startCmd.Flags().BoolVar(&eventDriven, "event-driven", true, "The reconciler's loop strategy. if it's false, it indicates period driven.")
	startCmd.Flags().StringVar(&portRange, "runtime-node-port-range", "16000-17999", "Set available port range for EFC")
}

func handle() {
	fluid.LogVersion()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = development
	}, func(o *zap.Options) {
		o.ZapOpts = append(o.ZapOpts, zapOpt.AddCaller())
	}, func(o *zap.Options) {
		if !development {
			encCfg := zapOpt.NewProductionEncoderConfig()
			encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
			encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
			o.Encoder = zapcore.NewConsoleEncoder(encCfg)
		}
	}))

	utils.NewPprofServer(setupLog, pprofAddr, development)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      metricsAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: leaderElectionNamespace,
		LeaderElectionID:        "efc.data.fluid.io",
		Port:                    9443,
		NewClient:               controllers.NewFluidControllerClient,
	})
	if err != nil {
		setupLog.Error(err, "unable to start efcruntime manager")
		os.Exit(1)
	}

	controllerOptions := controller.Options{
		MaxConcurrentReconciles: maxConcurrentReconciles,
	}

	if err = (efcctl.NewRuntimeReconciler(mgr.GetClient(),
		ctrl.Log.WithName("efcctl").WithName("EFCRuntime"),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("EFCRuntime"),
	)).SetupWithManager(mgr, controllerOptions, eventDriven); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "EFCRuntime")
		os.Exit(1)
	}

	pr, err := net.ParsePortRange(portRange)
	if err != nil {
		setupLog.Error(err, "can't parse port range. Port range must be like <min>-<max>")
		os.Exit(1)
	}
	setupLog.Info("port range parsed", "port range", pr.String())

	// portallocator.SetupRuntimePortAllocator(mgr.GetClient(), pr, efc.GetReservedPorts)
	portallocator.SetupRuntimePortAllocatorWithType(mgr.GetClient(),
		pr,
		portallocator.Random,
		efc.GetReservedPorts)

	if err := mgr.Add(efc.NewSessMgrInitializer(mgr.GetClient())); err != nil {
		setupLog.Error(err, "fail to add SessMgrInitializer to controller manager")
		os.Exit(1)
	}

	setupLog.Info("starting efcruntime-controller")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem efcruntime-controller")
		os.Exit(1)
	}
}
