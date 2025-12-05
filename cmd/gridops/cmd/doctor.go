package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/chanzuckerberg/fogg/config/markers"
	"github.com/spf13/cobra"
	"github.com/terraconstructs/grid/pkg/sdk"
)

var (
	verbose         bool
	veryVerbose     bool
	withGridPreview bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check for marker inconsistencies and preview sync actions",
	Long: `Scans for .grid-state.yaml files and checks for duplicates or conflicts.
With -v, shows detailed marker information.
With -vv, shows dependency resolution details.
With --with-grid-preview, connects to Grid API to preview what sync would do (dry-run).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		markers, err := ScanMarkers(cwd, excludeDirs)
		if err != nil {
			return err
		}

		issues := ValidateMarkers(markers)
		fmt.Printf("Found %d markers.\n", len(markers))

		if verbose || veryVerbose {
			fmt.Println("\n📋 Marker Details:")
			for _, marker := range markers {
				relPath, _ := filepath.Rel(cwd, marker.Path)
				fmt.Printf("\n  📄 %s\n", relPath)
				fmt.Printf("     GUID:      %s\n", marker.Marker.GUID)
				fmt.Printf("     LogicalID: %s\n", marker.Marker.LogicalID)

				if len(marker.Marker.Labels) > 0 {
					fmt.Printf("     Labels:\n")
					for k, v := range marker.Marker.Labels {
						fmt.Printf("       %s: %s\n", k, v)
					}
				}

				if len(marker.Marker.Dependencies) > 0 {
					fmt.Printf("     Dependencies:\n")
					for _, dep := range marker.Marker.Dependencies {
						if dep.Output != "" {
							fmt.Printf("       - GUID: %s, Output: %s", dep.GUID, dep.Output)
						} else {
							fmt.Printf("       - GUID: %s, Output: <unspecified>", dep.GUID)
						}
						if dep.Input != "" {
							fmt.Printf(", Input: %s", dep.Input)
						}
						fmt.Println()
					}
				}
			}
		}

		hasErrors := false
		if len(issues) > 0 {
			fmt.Println("\n❌ Validation Issues:")
			hasErrors = true
			for _, issue := range issues {
				fmt.Printf("   %s\n", issue)
			}
		}

		fmt.Println("\n🔍 Checking dependencies...")
		for _, marker := range markers {
			deps := marker.Marker.Dependencies
			if !needsOutputInference(deps) {
				if verbose || veryVerbose {
					fmt.Printf("   ✅ [%s] All dependencies have outputs specified\n", marker.Marker.LogicalID)
				}
				continue
			}

			// Notify that inference will happen
			unspecifiedCount := 0
			for _, dep := range deps {
				if strings.TrimSpace(dep.Output) == "" {
					unspecifiedCount++
				}
			}

			if verbose || veryVerbose {
				fmt.Printf("   🔎 [%s] %d dependencies with unspecified outputs - inferring from Terraform files...\n",
					marker.Marker.LogicalID, unspecifiedCount)
			}

			resolved, err := resolveMarkerDependencies(marker)
			if err != nil {
				fmt.Printf("   ❌ [%s] Failed to resolve dependencies: %v\n", marker.Path, err)
				hasErrors = true
				continue
			}

			// Find what was inferred
			inferredCount := 0
			for _, r := range resolved {
				found := false
				for _, d := range deps {
					if d.GUID == r.GUID && d.Output == r.Output && d.Input == r.Input {
						found = true
						break
					}
				}
				if !found {
					inferredCount++
					if verbose || veryVerbose {
						fmt.Printf("      ↳ Inferred: %s -> %s\n", r.GUID, r.Output)
					}
				}
			}

			if inferredCount > 0 {
				if !verbose && !veryVerbose {
					fmt.Printf("   ✅ [%s] Inferred %d outputs from Terraform files\n", marker.Marker.LogicalID, inferredCount)
				} else {
					fmt.Printf("   ✅ [%s] Successfully inferred %d outputs (total dependencies: %d)\n",
						marker.Marker.LogicalID, inferredCount, len(resolved))
				}
			} else {
				// This could happen if the dependency has unspecified outputs but no outputs were found in .tf files
				fmt.Printf("   ⚠️  [%s] No outputs could be inferred - check your Terraform files\n", marker.Marker.LogicalID)
			}
		}

		// Grid preview mode: show what Grid API calls would be made
		if withGridPreview {
			if err := previewGridAPIActions(cmd.Context(), markers); err != nil {
				fmt.Printf("\n⚠️  Could not preview Grid API actions: %v\n", err)
				fmt.Println("   (This is expected if Grid is not accessible or credentials are not configured)")
			}
		}

		if hasErrors {
			return fmt.Errorf("doctor found issues")
		}

		fmt.Println("\n✅ No issues found.")
		return nil
	},
}

func previewGridAPIActions(ctx context.Context, markers []LoadedMarker) error {
	fmt.Println("\n🔮 Grid API Actions Preview (dry-run):")

	// Server URL is required, but credentials are optional (depends on Grid server config)
	if opts.serverURL == "" {
		return fmt.Errorf("Grid server URL not configured (use --server or GRID_API_URL)")
	}

	session := sessionConfig{
		ServerURL:    opts.serverURL,
		ClientID:     opts.clientID,
		ClientSecret: opts.clientSecret,
	}

	client, err := newGridClient(ctx, session)
	if err != nil {
		return err
	}

	for _, marker := range markers {
		fmt.Printf("\n   📍 %s (GUID: %s)\n", marker.Marker.LogicalID, marker.Marker.GUID)

		if err := previewMarkerSync(ctx, client, marker); err != nil {
			fmt.Printf("      ⚠️  Error: %v\n", err)
		}
	}

	return nil
}

func previewMarkerSync(ctx context.Context, client *sdk.Client, marker LoadedMarker) error {
	desiredLabels := markerLabelsToLabelMap(marker.Marker.Labels)
	desiredDeps, err := resolveMarkerDependencies(marker)
	if err != nil {
		return err
	}

	info, err := client.GetStateInfo(ctx, sdk.StateReference{GUID: marker.Marker.GUID})
	if err != nil {
		var cerr *connect.Error
		if errors.As(err, &cerr) && cerr.Code() == connect.CodeNotFound {
			fmt.Printf("      🆕 Would CREATE state:\n")
			fmt.Printf("         - GUID: %s\n", marker.Marker.GUID)
			fmt.Printf("         - LogicalID: %s\n", marker.Marker.LogicalID)
			if len(desiredLabels) > 0 {
				fmt.Printf("         - Labels: %v\n", formatLabelMap(desiredLabels))
			}
			if len(desiredDeps) > 0 {
				fmt.Printf("         - Dependencies: %d\n", len(desiredDeps))
				for _, dep := range desiredDeps {
					output := dep.Output
					if output == "" {
						output = "default"
					}
					fmt.Printf("           • from=%s output=%s", dep.GUID, output)
					if dep.Input != "" {
						fmt.Printf(" input=%s", dep.Input)
					}
					fmt.Println()
				}
			}
			return nil
		}
		return fmt.Errorf("failed to fetch state: %w", err)
	}

	// State exists, check what would be updated
	fmt.Printf("      ✓ State exists (LogicalID: %s)\n", info.State.LogicID)

	if info.State.LogicID != marker.Marker.LogicalID {
		fmt.Printf("      ⚠️  LogicalID mismatch: grid=%s, marker=%s (rename not supported)\n",
			info.State.LogicID, marker.Marker.LogicalID)
	}

	adds, removals := diffLabels(info.Labels, desiredLabels)
	if len(adds) > 0 || len(removals) > 0 {
		fmt.Printf("      🏷️  Would UPDATE labels:\n")
		if len(adds) > 0 {
			fmt.Printf("         Add: %v\n", formatLabelMap(adds))
		}
		if len(removals) > 0 {
			fmt.Printf("         Remove: %v\n", removals)
		}
	} else {
		fmt.Printf("      ✓ Labels are up to date\n")
	}

	// Preview dependency changes
	if err := previewDependencyChanges(info, marker, desiredDeps); err != nil {
		return err
	}

	return nil
}

func previewDependencyChanges(info *sdk.StateInfo, marker LoadedMarker, desired []markers.Dependency) error {
	const defaultOutput = "default"

	desiredMap := make(map[string]markers.Dependency)
	for _, dep := range desired {
		guid := strings.TrimSpace(dep.GUID)
		if guid == "" {
			continue
		}
		output := dep.Output
		if output == "" {
			output = defaultOutput
		}
		key := depKey(guid, output)
		desiredMap[key] = dep
	}

	existing := make(map[string]sdk.DependencyEdge)
	for _, edge := range info.Dependencies {
		// Only consider incoming edges for this state
		if edge.To.GUID != "" && edge.To.GUID != info.State.GUID {
			continue
		}
		if edge.To.GUID == "" && info.State.LogicID != "" && edge.To.LogicID != info.State.LogicID {
			continue
		}

		fromGUID := edge.From.GUID
		if fromGUID == "" {
			fromGUID = edge.From.LogicID
		}
		if fromGUID == "" {
			continue
		}
		output := edge.FromOutput
		if output == "" {
			output = defaultOutput
		}
		key := depKey(fromGUID, output)
		existing[key] = edge
	}

	// Find what would be removed
	var toRemove []sdk.DependencyEdge
	for key, edge := range existing {
		if _, ok := desiredMap[key]; !ok {
			toRemove = append(toRemove, edge)
		}
	}

	// Find what would be added
	var toAdd []markers.Dependency
	for key, dep := range desiredMap {
		if _, ok := existing[key]; !ok {
			toAdd = append(toAdd, dep)
		}
	}

	if len(toRemove) > 0 {
		fmt.Printf("      ➖ Would REMOVE dependencies:\n")
		for _, edge := range toRemove {
			fmt.Printf("         • from=%s output=%s", edge.From.GUID, edge.FromOutput)
			if edge.ToInputName != "" {
				fmt.Printf(" input=%s", edge.ToInputName)
			}
			fmt.Println()
		}
	}

	if len(toAdd) > 0 {
		fmt.Printf("      ➕ Would ADD dependencies:\n")
		for _, dep := range toAdd {
			output := dep.Output
			if output == "" {
				output = defaultOutput
			}
			fmt.Printf("         • from=%s output=%s", dep.GUID, output)
			if dep.Input != "" {
				fmt.Printf(" input=%s", dep.Input)
			}
			fmt.Println()
		}
	}

	if len(toRemove) == 0 && len(toAdd) == 0 {
		fmt.Printf("      ✓ Dependencies are up to date\n")
	}

	return nil
}

func formatLabelMap(labels sdk.LabelMap) string {
	if len(labels) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func init() {
	rootCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed marker information")
	doctorCmd.Flags().BoolVarP(&veryVerbose, "vv", "", false, "Show dependency resolution details")
	doctorCmd.Flags().BoolVar(&withGridPreview, "with-grid-preview", false, "Connect to Grid API to preview sync actions (dry-run)")
}
