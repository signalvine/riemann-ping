package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/amir/raidman"
	"github.com/codegangsta/cli"
)

var riemannSend func(url, method string, duration float64)

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "nil"
	}
	app := cli.NewApp()
	app.Name = "riemann-ping"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			Value:  "localhost",
			Usage:  "the host that riemann is running on",
			EnvVar: "RIEMANN_HOST",
		},
		cli.IntFlag{
			Name:   "port, p",
			Value:  5555,
			Usage:  "The port that riemann is running on",
			EnvVar: "RIEMANN_PORT",
		},
		cli.StringFlag{
			Name:   "event-host, e",
			Usage:  "The hostname to use for the event",
			Value: hostname,
		},
		cli.DurationFlag{
			Name:  "interval, i",
			Value: 10 * time.Second,
			Usage: "The interval between the pings in seconds",
		},
		cli.StringSliceFlag{
			Name:  "attribute, a",
			Value: &cli.StringSlice{},
			Usage: "A list of attibutes to add to the event. Should be in the form attribute=value",
		},
		cli.StringSliceFlag{
			Name:  "tags, t",
			Value: &cli.StringSlice{},
			Usage: "A list of tags to add to the event",
		},
		cli.DurationFlag{
			Name:  "ttl, l",
			Usage: "The ttl of the event in seconds, should be at least as long as the interval",
			Value: 30 * time.Second,
		},
		cli.BoolFlag{
			Name:  "tcp, c",
			Usage: "Use tcp instead of udp.",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "get",
			Usage: "Performs a get request against all of the provided urls",
			Action: func(c *cli.Context) {
				interval := processGlobalFlags(c)
				for _, arg := range c.Args() {
					go checkScheduler(arg, "get", interval)
				}

				for {
					time.Sleep(100 * time.Second)
				}
			},
		},
	}

	app.Run(os.Args)
}

func processGlobalFlags(c *cli.Context) time.Duration {
	var network string
	if c.GlobalBool("tcp") {
		network = "tcp"
	} else {
		network = "udp"
	}
	client, err := raidman.Dial(network, fmt.Sprintf("%s:%d", c.GlobalString("host"), c.GlobalInt("port")))
	if c.GlobalString("event-host") == "nil" {
		log.Panic("Failed to automatically get the hostname. Please specify it with --host")
	}
	if err != nil {
		log.Panicf("Failed to connect to the riemann host because %s", err)
	}
	attribute, err := processAttributes(c.GlobalStringSlice("attribute"))
	if err != nil {
		log.Panic(err)
	}
	eventTemplate := raidman.Event{
		Ttl:        float32(c.GlobalDuration("ttl").Seconds()),
		Tags:       c.GlobalStringSlice("tags"),
		Host:       c.GlobalString("event-host"),
		Attributes: attribute,
	}
	riemannSend = func(url, method string, duration float64) {
		event := eventTemplate
		event.Service = fmt.Sprintf("%s %s", url, method)
		event.Metric = duration
		client.Send(&event)
	}

	return c.GlobalDuration("interval")
}

func processAttributes(attributes []string) (map[string]string, error) {
	parsedAttributes := *new(map[string]string)
	for _, combined := range attributes {
		parts := strings.Split(combined, "=")
		if len(parts) != 2 {
			return parsedAttributes, fmt.Errorf("Failed to parse %s as attitube, the format is incorrect", combined)
		}
	}

	return parsedAttributes, nil
}

func checkScheduler(url, method string, interval time.Duration) {
	for {
		select {
		case <-time.After(interval):
			if method == "get" {
				go getRequest(url)
			}
		}
	}
}

func getRequest(url string) {
	startTime := time.Now()
	response, err := http.Get(url)
	duration := time.Since(startTime)
	if err != nil {
		log.Printf("Failed to get '%s' because %s", url, err)
		return
	}
	if response.StatusCode != 200 {
		log.Printf("'%s' is unhealthy", url)
		return
	}
	riemannSend(url, "get", duration.Seconds()*1000)
}
