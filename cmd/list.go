package cmd

import (
	"encoding/json"

	. "github.com/duanqiaobb/BlogExporter/pkg"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sites supported",
	Long:  `List all sites supported`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		avaliable_sites := list_sites()
		format, _ := cmd.Flags().GetString("format")
		if format == "text" {
			cmd.Println(textify_sites(avaliable_sites))
		} else if format == "json" {
			cmd.Println(jsonify_sites(avaliable_sites))
		}
	},
}

func textify_sites(sites []string) string {
	text_sites := "Available sites: \n"
	for _, site := range sites {
		text_sites += site + "\n"
	}
	return text_sites
}

func jsonify_sites(sites []string) string {
	jsonify_sites := ""
	json_sites := map[string][]string{}
	json_sites["avaliable_sites"] = sites
	jsonify_sites_bytes, _ := json.Marshal(json_sites)
	jsonify_sites = string(jsonify_sites_bytes)
	return jsonify_sites

}

func list_sites() []string {
	sites := GetRegisteredSpiderNames()
	return sites
}

func init() {
	listCmd.Flags().StringP("format", "f", "text", "format of output [json, text]")
}
