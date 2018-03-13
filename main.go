package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ashb/jqrepl/jq"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "faq",
		Short: "format agnostic querier",
		Long:  "faq is like `sed`, but for object-like data using libjq.",
		Args:  cobra.MinimumNArgs(2),
		RunE:  runCmdFunc,
	}

	rootCmd.Flags().Bool("debug", false, "enable debug logging")
	rootCmd.Flags().StringP("format", "f", "json", "object format (e.g. json, yaml, bencode)")
	rootCmd.Execute()
}

func runCmdFunc(cmd *cobra.Command, args []string) error {
	prog, err := jq.New()
	if err != nil {
		return fmt.Errorf("failed to initialize libjq: %s", err)
	}
	defer prog.Close()

	path := os.ExpandEnv(args[1])
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file at %s: `%s`", path, err)
	}

	formatName, err := cmd.Flags().GetString("format")
	if err != nil {
		panic("failed to find format flag")
	}

	unmarshaledFile, err := unmarshal(formatName, fileBytes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal file at %s: `%s`", path, err)
	}

	fileJv, err := jq.JvFromInterface(unmarshaledFile)
	if err != nil {
		panic("failed to reflect a jv from unmarshalled file")
	}

	errs := prog.Compile(args[0], jq.JvArray())
	for _, err := range errs {
		if err != nil {
			return fmt.Errorf("failed to compile jq program for file at %s: %s", path, err)
		}
	}

	resultJvs, err := prog.Execute(fileJv)
	if err != nil {
		return fmt.Errorf("failed to execute jq program for file at %s: %s", path, err)
	}

	for _, resultJv := range resultJvs {
		fmt.Println(resultJv.Dump(jq.JvPrintPretty | jq.JvPrintSpace1 | jq.JvPrintColour))
	}

	return nil
}
