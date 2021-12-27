//usr/bin/env go run $0 "$@"; exit
package main

import (
	//"os"
	"fmt"
	"log"
	"net"
	"time"
	"flag"
	"net/http" // <====[Maybe try fasthttp?]
	"crypto/tls"
)



type Options struct {
	url string
	urlLst string

	method string

	timeout int
	threads int
}

type Target struct {
	URL string
}

type Response struct {
	R_url string
	R_status int
	R_ContentLength string
}

type Client struct{
	client *http.Client
}


/*----------------------[README]-------------------------*\
| Make sure to understand the code before continue on it. |
| Setup every *struct* and *function* and sketch how to   |
| process it before you start do the actuall code.		  |
|														  |
|			~ Happy Hacking by Brumens					  |
|			~ Made with coffine <3 						  |
\*-------------------------------------------------------*/


func main() {
	banner()
	opt := userArguments()
	client := client(opt)


	//urls lists (7) -{TEMP}
	lst_urls := []string{
		"https://www.google.com",
		"https://support.google.com",
		"https://www.example.com",
		"https://www.youtube.com",
		"https://www.nordvpn.com",
		"https://www.bing.com",
		"https://www.yandex.com",
	}

	//Set a *timer* that calculate the total process time of all code executions and requests:
	timer := time.Now()

	//Job and Result pools: (jobs=[Add processes to a queue], result=[Response data storage] )
	jobs := make(chan Target)
	results := make(chan Response)


	/*DEBUGG*/fmt.Printf("[DBG]\033[1;31m[HTTP Client] : [%v] \033[0m\n", client)//*/


	//Starting up threads to do the jobs: (opt.threads == threads to use)
	for thread := 1; thread <= opt.threads; thread++ {
		/*DEBUGG*/fmt.Printf("[DBG]\033[1;31m[Thread] : [%v] \033[0m\n", thread)//*/
		//go request(opt, client, thread, jobs, results)
	}

	
	//Take urls and add each to the "Job pool": -{TEMP}
	//fmt.Println(opt.urlLst)
	for _, url := range lst_urls {
		/*DEBUGG*/fmt.Printf("[DBG]\033[1;31m[Url] : [%v] \033[0m\n", url)//*/
		//jobs <- Target{URL: url}
	}
	close(jobs)


	//Check the stored result in the "result pool"
	for a := 1; a <= 7; a++ {
		result := <-results
		log.Println(result)
	}


	log.Printf(":: Process finished. Time[%v]\n", time.Since(timer) )
}

func userArguments() *Options {
	opt := &Options{}

	flag.StringVar(&opt.url, "u", "", "Url to test")
	flag.StringVar(&opt.urlLst, "l", "", "File containing urls to test")
	flag.IntVar(&opt.timeout, "T", 7000, "Timeout (ms) before giving up on response")
	flag.IntVar(&opt.threads, "t", 15, "Threads(\"Workers\") to use")
	flag.StringVar(&opt.method, "m", "GET", "HTTP method to use in the request")

	flag.Parse()

	return opt
}


func configure(opt *Options) {

	if opt.urlLst != "" {
		fmt.Println("List with url(FILE) >>", opt.urlLst)
	}else if opt.url != "" {
		fmt.Println("Single url >>", opt.url)
	}

	//file_urlLst, _ := os.Open(opt.urlList)
	//for _, url := range
}

func request(opt *Options, client *http.Client, Thread_id int, jobs <-chan Target, results chan<- Response) {

	//Extract each urls inside the "job pool":
	for url := range jobs {

		/*DEBUGG*/fmt.Printf("[DBG]\033[1;31m[Thread_id] : [%v]\033[0m\n", Thread_id)//*/


		//Request each urls and store it's response data into the "results pool":
		req, err := http.NewRequest(opt.method, url.URL, nil)
		if err != nil {
			log.Printf(":E: Request process failed -> ", err.Error())
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf(":E: Failed to send Request ->", err.Error())
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
	/*DEBUGG*/fmt.Printf("[DBG]\033[1;31m[Request Timeout] : [%v]\033[0m\n", timeout)//*/

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

Author: Brumens
Version: 1.0

Stay ethical. You are responsible for your actions.
creator of the tool is not responsible for any misuse or damage.

`, "\033[1;36m", "\033[0m") //<===[Colors for the banned]
}