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
	"regexp"
	"strings"
	"net/http" // <====[Fasthttp can be replaced. If so you do need to modify the request handling and client inside the function "requests()" and "client()"]
	"crypto/tls"
)


//Options variable must be defined here before it's added as a argument inside the function "userArguments()": ("-h/--help")
type Options struct {
	url string
	filenameUrls string
	lst_urls []string

	method string

	timeout int
	threads int

	verbose bool
}

/*Target : Jobs will use this struct to collect and add new jobs for it's channel(chan) at "jobs": [CUSTOMIZE] - (If prefered)
  :[ Add variables if more than the URL should be use as a target. (NOTE: method, timeout etc. Is used inside the struct "Options")
  !(Do not add)! ] */
type Target struct {
	URL string
}

/*Response : data will be collected and added to the variables temporary before it's stored inside the channel(chan) at "results": [CUSTOMIZE] - (Recommended)
  :[ Add more variable if you like. This variables is the collected data/information from the response as a result.
  This variables will be stored inside "results" when the process take place => ( "results := make(chan Response)" ) ] */
type Response struct {
	R_url string
	R_status int
	R_ContentLength string

	Error string
}

//Define the HTTP client that is configured in the function "client()"
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

	//Import the returned values from "Options" which is then configured before taking action in the process
	opt := userArguments()
	client := client(opt)
	configure(opt)


	//Disclaimer text:
	fmt.Printf("\033[1;1mStay ethical. You are responsible for your actions.\ncreator of the tool is not responsible for any misuse or damage.\033[0m\n")
	fmt.Println(strings.Repeat("-", 64))


	//Set a *timer* that calculate the total process time of all code executions and requests:
	timer := time.Now()


	//Job and Result channels. Aka: "workpools": (jobs=[Add processes to a queue (Struct "Target")], result=[Response data storage(Struct "Response")] )
	jobs := make(chan Target)
	results := make(chan Response)

	//Starting up threads to do the "jobs" channel: (opt.threads == Amount of threads)
	for thread := 1; thread <= opt.threads; thread++ {
		go request(opt, client, thread, jobs, results)
	}

	
	//Take urls from the file and add each url to the "Jobs" channel:
	for _, url := range opt.lst_urls {
		jobs <- Target{URL: url}
	}
	//Important to close the "jobs" channel when no more jobs should be added. Otherwise it can cause a freeze of the process:
	close(jobs)


	/*Check what job process that have been done and are stored in "result pool".
	  When an request failes (Ex: Timeout issues) it will still be added to the "result" channel and because of that "len(opt.lst_urls)"
	  is used to calculate the amount of urls that the file contained and therefore can track the process when it's properly finished. */ 
	for item := 1; item <= len(opt.lst_urls); item++ {
		result := <-results
		log.Printf(": %v \n", result)
	}

	//Process is finished and kill the tool with "os.Exit(0)" with status code "0" to close all threads and memory files properly.
	log.Printf("| Process finished. Time [%v]\n", time.Since(timer) )
	os.Exit(0)
}


/*User parse arguments options: [CUSTOMIZE] - (Recommended)
  :[ Define inside "Response" struct first! ] */
func userArguments() *Options {
	opt := &Options{}

	flag.StringVar(&opt.url, "u", "", "Url to test")
	flag.StringVar(&opt.filenameUrls, "l", "", "File containing urls to test")
	flag.IntVar(&opt.timeout, "T", 7000, "Timeout (ms) before giving up on response")
	flag.IntVar(&opt.threads, "t", 15, "Threads(\"Workers\") to use")
	flag.StringVar(&opt.method, "m", "GET", "HTTP method to use in the request")
	flag.BoolVar(&opt.verbose, "v", false, "Verbose output")
	/*Add more options here and at the stuct "Options" above:
	.. ..... .. . .. . 
	... . ..  . . ..
	*/ 

	flag.Parse()

	return opt
}

//Configure user options before taking action in the process:
func configure(opt *Options) {

	//Regrex to check if url start with "http[s]://" to avoid request errors:
	urlRegrex := regexp.MustCompile(`http*`)


	//Check if either (-u) or (-l) have been set as an user argument:
	if opt.filenameUrls != "" {
			
		file, _ := os.Open(opt.filenameUrls)
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {

			//Check if the urls start with a protocol. If not then warn the user: (http[s])
			if urlRegrex.MatchString(string(scanner.Text())) == false {
				log.Printf(":\033[1;31mx\033[0m: One or more urls in the file do not have a http protocol as a prefix. This will cause request errors in the process\n")
				os.Exit(0)
			
			}else {
			
				if scanner.Text() != "" && scanner.Text() != "\n" {

				opt.lst_urls = append(opt.lst_urls, scanner.Text())
				}
			}
		}


	//If url regrex match and the URL has a HTTP protocol then add it to the "lst_urls" list:
	}else if opt.url != "" && urlRegrex.MatchString(string(opt.url)) == true {
		opt.lst_urls = append(opt.lst_urls, opt.url)


	/*If (-u) or (-l) haven't been set. Display an [USAGE] text for the user on how to use the tool and then EXIT: [CUSTOMIZE] - (Recommended)
	  :["Tool name" + "Options/Help" ] */
	}else {
		fmt.Println(":\033[1;31mx\033[0m: (-u) or (-l) was invalid. Make sure that the URL do have a HTTP protocol as a prefix.")
		fmt.Printf("Usage: TOOL_NAME -h \n")
		os.Exit(0)
	}
}


//Reques module that send and add the response data to the "results" channel and use "Response" as struct for dynamic temp variables:
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

/*Configure the HTTP client fully before making any request: [CUSTOMIZE] - (If prefered)
  :[ The Client options can be modifyed for preformance adjustment or add "opt.OPTION_NAME".
  So that it can be modifyed by the user parse arguments (NOTE: Define inside "response" struct and add as a "flag" inside the function "userArguments()") ] */
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

//Designed banner: [CUSTOMIZE] - (Recommended)
func banner() {
	fmt.Printf(`

	 [[ %vBANNER%v ]]
	
~ Custom ASCII banners: "https://patorjk.com/software/taag"

Author: Brumens
Version: 1.0

`, "\033[1;36m", "\033[0m") //<===[Custom colors for the banned]
}
