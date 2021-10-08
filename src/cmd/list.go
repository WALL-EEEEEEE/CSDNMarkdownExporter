package cmd

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sites supported",
	Long:  `List all sites supported`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		avaliable_sites, err := list_sites()
		if err != nil {
			cmd.PrintErrf("Error: %s", err)
			return
		}
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

func list_sites() ([]string, error) {
	var sites []string
	_, currFilePath, _, _ := runtime.Caller(0)
	spiderDir := path.Join(path.Dir(path.Dir(currFilePath)), "spiders")
	if _, err := os.Stat(spiderDir); err != nil {
		return sites, err
	}
	filepath.Walk(spiderDir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			site_name := strings.Split(info.Name(), filepath.Ext(info.Name()))[0]
			sites = append(sites, strings.ToUpper(site_name))
		}
		return nil
	})
	return sites, nil
}

func init() {
	listCmd.Flags().StringP("format", "f", "text", "format of output [json, text]")
}
