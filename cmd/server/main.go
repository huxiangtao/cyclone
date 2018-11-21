/*
Copyright 2018 caicloud authors. All rights reserved.

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

package main

import (
	"flag"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/profiling"

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1/descriptor"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1/handler"
	"github.com/caicloud/cyclone/pkg/server/config"
)

// APIServerOptions contains all options(config) for api server
type APIServerOptions struct {
	KubeHost   string
	KubeConfig string

	CyclonePort int
	CycloneAddr string

	Loglevel string
}

// NewAPIServerOptions returns a new APIServerOptions
func NewAPIServerOptions() *APIServerOptions {
	return &APIServerOptions{
		CyclonePort: 7099,
	}
}

// AddFlags adds flags to APIServerOptions.
func (opts *APIServerOptions) AddFlags() {

	flag.StringVar(&opts.KubeHost, config.EnvKubeHost, config.KubeHost, "Kube host address")
	flag.StringVar(&opts.KubeConfig, config.EnvKubeConfig, config.KubeConfig, "Kube config file path")

	flag.IntVar(&opts.CyclonePort, config.EnvCycloneServerPort, config.CycloneServerPort, "The port for the cyclone server to serve on.")
	flag.StringVar(&opts.CycloneAddr, config.EnvCycloneServerHost, config.CycloneServerHost, "The IP address for the cyclone server to serve on.")

	flag.StringVar(&opts.Loglevel, config.EnvLogLevel, config.LogLevel, "Log level.")

	flag.Parse()
}

func initialize(opts *APIServerOptions) {
	// Init k8s client
	client, err := common.GetClient(opts.KubeHost, opts.KubeConfig)
	if err != nil {
		log.Fatalf("Create k8s client error: %v", err)
	}

	handler.InitHandlers(client)

	log.Info("Init k8s client success")

	return
}

func main() {
	opts := NewAPIServerOptions()
	opts.AddFlags()

	initialize(opts)

	config := nirvana.NewDefaultConfig()
	nirvana.IP(opts.CycloneAddr)(config)
	nirvana.Port(uint16(opts.CyclonePort))(config)
	config.Configure(
		metrics.Path("/metrics"),
		profiling.Path("/debug/pprof/"),
		profiling.Contention(true),
	)

	config.Configure(nirvana.Descriptor(descriptor.Descriptor()))

	log.Infof("Cyclone service listening on %s:%d", opts.CycloneAddr, opts.CyclonePort)
	if err := nirvana.NewServer(config).Serve(); err != nil {
		log.Fatal(err)
	}

	log.Info("Cyclone server stopped")
}