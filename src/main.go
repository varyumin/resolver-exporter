package main

import (
	"github.com/prometheus/common/log"
	"flag"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
	"net"
	"time"
	"io/ioutil"
	"github.com/facebookgo/flagenv"
	"fmt"
)
var (
	sites listSites
	addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	file = flag.String("sites-file", "site.yaml", "YAML file with list sites")
	timeOutScrape = flag.Int64("timeout-scrape", 30, "YAML file with list sites")
	siteResolve = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "resolve_site",
			Help: "Check to resolve site in DNS server. False == 0 and True == 1",
		},
		[]string{"site"},
	)
	)

type listSites struct {
	Sites []string `yaml:"site"`
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Healthz!")
}

func (c *listSites) getSites() *listSites{
	yamlFile, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatalf("Open file. Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func init() {
	prometheus.MustRegister(siteResolve)
	sites.getSites()
	flagenv.Parse()
	flag.Parse()
}
func main() {
	go func() {
		for {
			for _, site := range sites.Sites{
				siteResolve.With(prometheus.Labels{"site": site}).Set(boolToFloat(resolvHost(site)))
			}
			time.Sleep(time.Duration(*timeOutScrape) * time.Second)
		}
	}()
	log.Infof("Start web server: %v", *addr)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", healthzHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func resolvHost(hostname string) (status bool){
	ips, err := net.LookupIP(hostname)
	if err != nil {
		log.Errorf("Hostname %s could not get IPs::: %s", hostname, err)
		status = false
	} else {
		log.Infof("%s:", hostname)
		for _, ip := range ips {
			log.Infof("%s", ip.String())
		}
		status = true
	}
	return status
}

func boolToFloat(status bool) float64{
	if status{
		return 1
	} else {
		return 0
	}
}