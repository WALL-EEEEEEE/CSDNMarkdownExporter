package cmd

import (
	"fmt"
	"strings"

	"github.com/duanqiaobb/BlogExporter/inter"
	"github.com/duanqiaobb/BlogExporter/spiders"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "run a blog exporter",
		Long:  `run a blog exporter`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			log.SetFormatter(&log.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			})
			site, _ := cmd.Flags().GetString("site")
			site = strings.ToUpper(site)
			if !validate_site(site) {
				cmd.PrintErrf("Invalid site name!\n%s", textify_sites(sites))
			}
			switch site {
			case "CSDN":
				user, _ := cmd.Flags().GetString("user")
				user = strings.TrimSpace(user)
				outputDir, _ := cmd.Flags().GetString("output")
				outputDir = strings.TrimSpace(outputDir)
				cookie, _ := cmd.Flags().GetString("cookie")
				cookie = strings.ReplaceAll(strings.TrimSpace(cookie), "\"", "")
				spider := spiders.GetResigerSpiderByName("CSDN").New(user, cookie, outputDir).(inter.Spider)
				if len(user) < 1 {
					cmd.Help()
					cmd.PrintErrln("\nError: user of CSDN must be specified !")
					return
				}
				if len(cookie) < 1 {
					cmd.Help()
					cmd.PrintErrln("\nError: cookie of CSDN must be specified !")

				}
				spider.Crawl()
			}

		},
	}
	sites = list_sites()
)

func validate_site(site string) bool {
	for _, v := range sites {
		if v == site {
			return true
		}
	}
	return false
}

func init() {
	usage := fmt.Sprintf("site to be exported [%s]", strings.Join(sites, ", "))
	runCmd.Flags().StringP("site", "s", "CSDN", usage)
	runCmd.Flags().StringP("user", "u", "", "user in blog site")
	runCmd.Flags().StringP("cookie", "c", "", "user cookie in blog site")
}
