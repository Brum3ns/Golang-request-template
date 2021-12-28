//usr/bin/env go run $0 "$@"; exit
package main

import (
	"os"
	"fmt"
	"log"
	"net"
	"time"
	"flag"
	"bufio"
	"strings"
	"net/http" // <====[Maybe try fasthttp?]
	"crypto/tls"
)



type Options struct {
	url string
	filenameUrls string
	lst_urls []string

	method string

	timeout int
	threads int

	verbose bool
}

type Target struct {
	URL string
}

type Response struct {
	R_url string
	R_status int
	R_ContentLength string

	Error string
}

type Client struct{
	client *http.Client
}


/*----------------------[README]-------------------------*\
| Make sure to understand the code before continue on it. |
| Check every *struct* and *function* and how it is       |
| processed before you start to add your own code		  |
|														  |
|			~ Happy Hacking by Brumens					  |
|			~ Made with coffine <3 						  |
\*-------------------------------------------------------*/

func main() {
	banner()
	opt := userArguments()
	client := client(opt)
	
	configure(opt)

	//Disclaimer text:
	fmt.Printf("\033[1;1mStay ethical. You are responsible for your actions.\ncreator of the tool is not responsible for any misuse or damage.\033[0m\n")
	fmt.Println(strings.Repeat("-", 64))


	//Set a *timer* that calculate the total process time of all code executions and requests:
	timer := time.Now()

	//Job and Result pools: (jobs=[Add processes to a queue], result=[Response data storage] )
	jobs := make(chan Target)
	results := make(chan Response)

	//Starting up threads to do the jobs: (opt.threads == threads to use)
	for thread := 1; thread <= opt.threads; thread++ {
		go request(opt, client, thread, jobs, results)
	}

	
	//Take urls and add each to the "Job pool": -{TEMP}
	//fmt.Println(opt.urlLst)
	for _, url := range opt.lst_urls {
		jobs <- Target{URL: url}
	}
	close(jobs)


	//Check the stored result in the "result pool"
	for item := 1; item <= len(opt.lst_urls); item++ {
		result := <-results
		log.Printf("\033[1;36m:\033[0m %v \n", result)
	}


	log.Printf("\033[1;36m|\033[0m Process finished. Time [%v]\n", time.Since(timer) )
	os.Exit(0)
}

func userArguments() *Options {
	opt := &Options{}

	flag.StringVar(&opt.url, "u", "", "Url to test")
	flag.StringVar(&opt.filenameUrls, "l", "", "File containing urls to test")
	flag.IntVar(&opt.timeout, "T", 7000, "Timeout (ms) before giving up on response")
	flag.IntVar(&opt.threads, "t", 15, "Threads(\"Workers\") to use")
	flag.StringVar(&opt.method, "m", "GET", "HTTP method to use in the request")
	flag.BoolVar(&opt.verbose, "v", false, "Verbose output")

	flag.Parse()

	return opt
}


func configure(opt *Options) {

	if opt.filenameUrls != "" {
			
		file, _ := os.Open(opt.filenameUrls)
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {

			if scanner.Text() != "" && scanner.Text() != "\n" {

				opt.lst_urls = append(opt.lst_urls, scanner.Text())
			}
		}

	}else if opt.url != "" {
	
		opt.lst_urls = append(opt.lst_urls, opt.url)
	}else {
		fmt.Printf("\033[1;36m|\033[0m Usage: scar \033[1;36m-h\033[0m\n")
		os.Exit(0)
	}
}

func request(opt *Options, client *http.Client, Thread_id int, jobs <-chan Target, results chan<- Response) {

	//Extract each urls inside the "job pool":
	for url := range jobs {

		//Request each urls and store it's response data into the "results pool":
		req, err := http.NewRequest(opt.method, url.URL, nil)
		if err != nil {
			if opt.verbose == true {log.Printf(":\033[1;31mx\033[0m: Request process failed \033[1;31m->\033[0m [%v]", err.Error())}

			results <- Response{Error: "url"}
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			if opt.verbose == true {log.Printf(":\033[1;31mx\033[0m: Failed to send Request \033[1;31m->\033[0m [%v]", err.Error())}
			
			results <- Response{Error: "url"}
			return
		}

		results <- Response{
			R_url: "url",
			R_status: resp.StatusCode,
			R_ContentLength: "BodySize",
		}
	}
}

func client(opt *Options) *http.Client {
	timeout := time.Duration(opt.timeout) * time.Millisecond

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
		Timeout:	timeout,
		Transport: &http.Transport{
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 500,
			MaxConnsPerHost:     500,
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				Renegotiation: tls.RenegotiateOnceAsClient,
			},
		},
	}
	return client
}


func banner() {
	fmt.Printf(`%v
   __  __  __  __ 
  /_  /   /_/ /_/
 ._/ /_, / / / )%v

"Make a scar in the web"

Author: Brumens
Version: 1.0

`, "\033[1;36m", "\033[0m") //<===[Colors for the banned]
}