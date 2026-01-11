// Copyright 2026 The Casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/casbin/policywall/pkg/audit"
	casbinpkg "github.com/casbin/policywall/pkg/casbin"
)

var (
	kubeconfig string
	namespace  string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "policywall",
		Short: "PolicyWall CLI for managing admission policies",
		Long:  `A command-line tool for managing Kubernetes admission policies with Casbin`,
	}

	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")

	rootCmd.AddCommand(newAuditCmd())
	rootCmd.AddCommand(newTemplatesCmd())
	rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newAuditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit existing resources against policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to build config: %w", err)
			}

			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				return fmt.Errorf("failed to create clientset: %w", err)
			}

			enforcer := casbinpkg.NewPolicyEnforcer()
			auditor := audit.NewAuditor(clientset, enforcer)

			if namespace != "" {
				return auditor.AuditNamespace(context.Background(), namespace)
			}
			return auditor.AuditAll(context.Background())
		},
	}

	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to audit (default: all)")

	return cmd
}

func newTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "List available policy templates",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Available policy templates:")
			for name, tmpl := range casbinpkg.Templates {
				fmt.Printf("\n  %s\n    %s\n", name, tmpl.Description)
			}
		},
	}

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("PolicyWall v0.1.0")
		},
	}
}
