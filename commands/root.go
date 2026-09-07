package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/redhat-consulting-services/kustomize-validator/k8s"
	"github.com/redhat-consulting-services/kustomize-validator/validate"
	"github.com/spf13/cobra"
)

var (
	// checkArbitrary is a slice of strings to check for arbitrary validation in the rendered kustomize output
	// It is set via command line flag
	checkArbitrary *[]string = &[]string{}
)

var RootCmd = &cobra.Command{
	Use:  "kustomize-validator",
	Long: "A tool to validate Kustomization files",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("No arguments provided")
			return nil
		}

		fmt.Println("Validating Kustomization files", args[0])
		isVerbose := cmd.Flag("verbose").Value.String() == "true"
		isErrorOnly := cmd.Flag("error-only").Value.String() == "true"
		isTable := cmd.Flag("table").Value.String() == "true"

		var tableRows [][]string
		cwd, _ := os.Getwd()

		if isTable {
			tableRows = append(tableRows, []string{"Relative path", "ApiVersion", "Kind", "Name", "Namespace", "Validation Error"})
		}

		msgChan := validate.KustomizeBuild(args[0])
		ctx, cf := context.WithTimeout(context.Background(), 2*time.Second)
		defer cf()

		totalCounter := 0
		successCounter := 0
		failureCounter := 0

	BREAK:
		for {
			select {
			case <-ctx.Done():
				break BREAK
			case msg, ok := <-msgChan:
				if !ok {
					break BREAK
				}
				totalCounter++
				isError := false
				if msg.Err != nil {
					isError = true
				}

				// all rendered resources from kustomize output
				resources := k8s.ParseKustomizeOutput(msg.Stdout, msg.Path, cwd)

				// if no error, validate content
				rsrcs := validate.ValidateContent(resources, *checkArbitrary)
				if msg.Err == nil {
					msg.Err = rsrcs.Error()
				}
				if msg.Err != nil {
					isError = true
				}

				if isTable {
					for _, resource := range resources {
						tableRows = append(tableRows, []string{
							resource.SourcePath,
							resource.ApiVersion,
							resource.Kind,
							resource.Name,
							resource.Namespace,
							rsrcs.Find(resource.ApiVersion, resource.Kind, resource.Namespace, resource.Name).Error(),
						})
					}
				} else {
					fmt.Print(msg.Msg(isErrorOnly, isVerbose))
				}

				if isError {
					failureCounter++
				} else {
					successCounter++
				}
			}
		}

		if isTable && len(tableRows) > 1 {
			table := tablewriter.NewTable(os.Stdout)

			// Convert header to []any
			header := make([]any, len(tableRows[0]))
			for i, v := range tableRows[0] {
				header[i] = v
			}
			table.Header(header...)

			// Convert data rows to [][]any
			data := make([][]any, len(tableRows)-1)
			for i := 1; i < len(tableRows); i++ {
				row := make([]any, len(tableRows[i]))
				for j, v := range tableRows[i] {
					row[j] = v
				}
				data[i-1] = row
			}
			table.Bulk(data)

			table.Render()
		}
		fmt.Println("Total: ", validate.ColorF(validate.ColorBlue, "%d", totalCounter))
		fmt.Println("Success: ", validate.ColorF(validate.ColorGreen, "%d", successCounter))
		fmt.Println("Error: ", validate.ColorF(validate.ColorRed, "%d", failureCounter))
		fmt.Println("Failed in %: ", validate.ColorF(validate.ColorRed, "%.2f%%", float64(failureCounter)/float64(totalCounter)*100))

		if failureCounter > 0 {
			return fmt.Errorf("validation failed")
		}
		return nil
	},
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().BoolP("error-only", "e", false, "whether we should only log errors")
	RootCmd.PersistentFlags().BoolP("table", "t", false, "output resources in table format")
	checkArbitrary = RootCmd.PersistentFlags().StringSliceP("check", "c", []string{"PATCH_ME", "patch_me"}, "check for arbitrary validation in rendered kustomize output.\nUse glob:pattern for glob matching, e.g., glob:PAT*_ME to match PAT123_ME\nor use the regex match pattern regex:app-.* to match app-123.\nIf no prefix is provided, literal substring matching is used (default).")
}
