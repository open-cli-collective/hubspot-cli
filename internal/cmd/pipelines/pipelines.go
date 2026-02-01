package pipelines

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/hubspot-cli/api"
	"github.com/open-cli-collective/hubspot-cli/internal/cmd/root"
)

// Register registers the pipelines command and subcommands
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use:   "pipelines",
		Short: "Manage HubSpot pipelines",
		Long:  "Commands for listing and viewing pipelines and their stages.",
	}

	cmd.AddCommand(newListCmd(opts))
	cmd.AddCommand(newGetCmd(opts))
	cmd.AddCommand(newStagesCmd(opts))

	parent.AddCommand(cmd)
}

func newListCmd(opts *root.Options) *cobra.Command {
	var objectType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List pipelines",
		Long:  "List all pipelines for an object type.",
		Example: `  # List deal pipelines
  hspt pipelines list --object-type deals

  # List ticket pipelines
  hspt pipelines list --object-type tickets`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if objectType == "" {
				return fmt.Errorf("--object-type is required (deals or tickets)")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			result, err := client.ListPipelines(api.ObjectType(objectType))
			if err != nil {
				return err
			}

			if len(result.Results) == 0 {
				v.Info("No pipelines found")
				return nil
			}

			headers := []string{"ID", "LABEL", "DISPLAY ORDER", "STAGES"}
			rows := make([][]string, 0, len(result.Results))
			for _, pipeline := range result.Results {
				rows = append(rows, []string{
					pipeline.ID,
					pipeline.Label,
					fmt.Sprintf("%d", pipeline.DisplayOrder),
					fmt.Sprintf("%d", len(pipeline.Stages)),
				})
			}

			return v.Render(headers, rows, result)
		},
	}

	cmd.Flags().StringVar(&objectType, "object-type", "", "Object type (deals or tickets)")

	return cmd
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var objectType, pipelineID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a pipeline",
		Long:  "Get details of a specific pipeline.",
		Example: `  # Get a deal pipeline
  hspt pipelines get --object-type deals --id default

  # Get a ticket pipeline
  hspt pipelines get --object-type tickets --id 12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if objectType == "" || pipelineID == "" {
				return fmt.Errorf("--object-type and --id are required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			pipeline, err := client.GetPipeline(api.ObjectType(objectType), pipelineID)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Pipeline %s not found for %s", pipelineID, objectType)
					return nil
				}
				return err
			}

			headers := []string{"PROPERTY", "VALUE"}
			rows := [][]string{
				{"ID", pipeline.ID},
				{"Label", pipeline.Label},
				{"Display Order", fmt.Sprintf("%d", pipeline.DisplayOrder)},
				{"Archived", formatBool(pipeline.Archived)},
				{"Created", pipeline.CreatedAt},
				{"Updated", pipeline.UpdatedAt},
			}

			if len(pipeline.Stages) > 0 {
				rows = append(rows, []string{"Stages", fmt.Sprintf("%d stages", len(pipeline.Stages))})
			}

			return v.Render(headers, rows, pipeline)
		},
	}

	cmd.Flags().StringVar(&objectType, "object-type", "", "Object type (deals or tickets)")
	cmd.Flags().StringVar(&pipelineID, "id", "", "Pipeline ID")

	return cmd
}

func newStagesCmd(opts *root.Options) *cobra.Command {
	var objectType, pipelineID string

	cmd := &cobra.Command{
		Use:   "stages",
		Short: "List pipeline stages",
		Long:  "List all stages for a specific pipeline.",
		Example: `  # List stages for default deal pipeline
  hspt pipelines stages --object-type deals --id default

  # List stages for a ticket pipeline
  hspt pipelines stages --object-type tickets --id 12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			v := opts.View()

			if objectType == "" || pipelineID == "" {
				return fmt.Errorf("--object-type and --id are required")
			}

			client, err := opts.APIClient()
			if err != nil {
				return err
			}

			stages, err := client.GetPipelineStages(api.ObjectType(objectType), pipelineID)
			if err != nil {
				if api.IsNotFound(err) {
					v.Error("Pipeline %s not found for %s", pipelineID, objectType)
					return nil
				}
				return err
			}

			if len(stages) == 0 {
				v.Info("No stages found")
				return nil
			}

			headers := []string{"ID", "LABEL", "DISPLAY ORDER", "ARCHIVED"}
			rows := make([][]string, 0, len(stages))
			for _, stage := range stages {
				rows = append(rows, []string{
					stage.ID,
					stage.Label,
					fmt.Sprintf("%d", stage.DisplayOrder),
					formatBool(stage.Archived),
				})
			}

			return v.Render(headers, rows, stages)
		},
	}

	cmd.Flags().StringVar(&objectType, "object-type", "", "Object type (deals or tickets)")
	cmd.Flags().StringVar(&pipelineID, "id", "", "Pipeline ID")

	return cmd
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
