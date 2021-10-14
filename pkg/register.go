package spiders

import (
	"github.com/duanqiaobb/BlogExporter/pkg/inter"
	log "github.com/sirupsen/logrus"
)

var spiders map[string]inter.Spider = make(map[string]inter.Spider)

func RegisterSpider(name string, spider inter.Spider) {
	spiders[name] = spider
	log.Infof("Spider %s registered.", name)
}

func GetRegisteredSpiderNames() []string {
	spider_names := []string{}
	for name, _ := range spiders {
		spider_names = append(spider_names, name)
	}
	return spider_names
}

func GetResigerSpiderByName(name string) inter.Spider {
	log.Infof("Loading registered spider %s .", name)
	spider := spiders[name]
	log.Infof("Loaded registered spider %s .", name)
	return spider
}
