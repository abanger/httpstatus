// httpstatus.go 抓取网站的http服务状态
// Created on 15:47 2020/4/29
// Modify  on 14:31 2020/5/28
// @author: abanger 
// ver 0.1
// Copyright 2020  All Rights Reserved.
// 

package main

import (
        "flag"
        "fmt"        
		"io/ioutil"
		"net"
        "net/http"        
		"strings"
		"os"
		"sync"
		"log"
		"regexp"
		"crypto/tls"
        "golang.org/x/net/http2" 
)



type httpStatus struct{
	Domain string
    IPv4 string
    IPv6 string
	IPv4_http int
	Ipv6_http int
	IPv4_https int
	Ipv6_https int
	IPv4_http2 int
	Ipv6_http2 int    
}

var (
	httpStatusOut = make(chan httpStatus,9)
	outputWG = sync.WaitGroup{} 
	fetchWG = sync.WaitGroup{}
    checkIpv4 = true
    checkIpv6 = true
)


func main() {
	hostPtr := flag.String("h", "", "A valid internet site domain e.g example.com")
   
    flag.Parse()
    
	if strings.Compare(*hostPtr, "") == 0 {
		fmt.Println("Usage: httpstatus -h example.com")
		os.Exit(0)
	}
    url := fmt.Sprintf("%s", *hostPtr)
    
	outputWG.Add(1)
	go printResult()
    
    //fetch site
    fetchWG.Add(1)
    fetchSite(url)

	fetchWG.Wait()
	close(httpStatusOut)
	outputWG.Wait()

}


func fetchSite(url string){
    n := httpStatus{Domain: url}
	//configure response
    find,err := regexp.Compile("Connection*")
    
	if err !=nil{
		log.Println(err)
	}

    if checkIpv4 {
        n.IPv4_http = getHttp("tcp4",url,find)
        n.IPv4_https = getHttps("tcp4",url,find)
        n.IPv4_http2,n.IPv4 = getHttp2("tcp4",url)
        //try other 
        if n.IPv4_http==0{
            n.IPv4_http = getHttpTry("tcp4",url,find)
        }
	}
    if checkIpv6 {
        n.Ipv6_http = getHttp("tcp6",url,find)
        n.Ipv6_https = getHttps("tcp6",url,find)
        n.Ipv6_http2,n.IPv6 = getHttp2("tcp6",url)
	}
	httpStatusOut <- n  //发送值n到httpStatusOut中
	fetchWG.Done()
}

func printResult(){
	for v := range httpStatusOut {
        fmt.Println(v.Domain,v.IPv4,printStatus(v.IPv4_http),printStatus(v.IPv4_https),
            v.IPv4_http2,v.IPv6,printStatus(v.Ipv6_http),printStatus(v.Ipv6_https),
            v.Ipv6_http2) 
    }
	outputWG.Done()
}

func printStatus(s int) string{
	switch s {
	case 1:
		return "1";
	case 0:
		return "0";        
	case 200:
		return "2";
	case 300:
		return "3";
	case 400:
		return "4";
	case 500:
		return "5";
	}    
	return "9";
}

func getHttp(tcp, url string,find *regexp.Regexp) int {
	conn, err := net.Dial(tcp, url+":http")
	//handleError(err)
    if err != nil {
        return 0 //error return IP NULL
    }
	_, err = conn.Write([]byte("GET / HTTP/1.0\nHOST: "+url+"\r\n\r\n"))
    //_, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n" ))
    //请求头包括三个部分  请求方式  请求脚本的绝对路径   协议的版本
    if err != nil {
        _, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n" ))
        if err != nil {
            return 0 //error return IP NULL
        }
    }    
	result, _ := ioutil.ReadAll(conn)
    //log.Println(string(result))
	return outputConversion(result,find,false,false)
}

func getHttpTry(tcp, url string,find *regexp.Regexp) int {
	conn, err := net.Dial(tcp, url+":http")
	//handleError(err)
    if err != nil {
        return 0 //error return IP NULL
    }
	//_, err = conn.Write([]byte("GET / HTTP/1.0\nHOST: "+url+"\r\n\r\n"))
    _, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n" ))
    if err != nil {
        return 0 //error return IP NULL
    }    
	result, _ := ioutil.ReadAll(conn)
    //log.Println(string(result))
	return outputConversion(result,find,false,false)
}



func getHttps(tcp, url string,find *regexp.Regexp) int {
	conf := tls.Config{}
	conn, err := tls.Dial(tcp, url+":https",&conf)
    if err != nil {
        return 0 //error return IP NULL
    }
	_, err = conn.Write([]byte("GET / HTTP/1.0\nHOST: "+url+"\r\n\r\n"))
    //_, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n" ))
	//handleError(err)
    if err != nil {
        return 0 //error return IP NULL
    }    
	result, _ := ioutil.ReadAll(conn)
    //log.Println(string(result))
	ssl := false;
	out := outputConversion(result,find,true,ssl)
	return out;
}


//http2 h2
func getHttp2(tcp, url string) (int,string) {
    tcpAddr, err := net.ResolveTCPAddr(tcp, url+":443")
    if err != nil {
        return 0 ,""  //error return IP NULL
    }    
    // HTTP2 Transport
    client := http.Client{
        Transport: &http2.Transport{},
    }

    resp, err := client.Get("https://"+url)
    if err != nil {
        return 0 ,tcpAddr.IP.String()
    }
    if resp.Proto!="HTTP/2.0" {
        //fmt.Println("Not HTTP/2")
        return 0 ,tcpAddr.IP.String()
    }else{
        return 1 ,tcpAddr.IP.String()
    }
}


func outputConversion(result []byte,find *regexp.Regexp,ssl bool, sslwork bool) int{
	result_s := string(result)
    //fmt.Println(string(result_s))
    if strings.Contains(result_s,"HTTP/1.1 4") || strings.Contains(result_s,"HTTP/1.0 4"){
        return 400
    }
    if strings.Contains(result_s,"HTTP/1.1 5") || strings.Contains(result_s,"HTTP/1.0 5"){
        return 500
    }
    if strings.Contains(result_s,"HTTP/1.1 3") || strings.Contains(result_s,"HTTP/1.0 3"){
        return 300
    }    
	if strings.Contains(result_s,"HTTP/1.1 2") || strings.Contains(result_s,"HTTP/1.0 2"){
        //return 1
		if find.Match(result){
			return 1;
		}else{
            return 200
		}
	}
	return 0;
}


func handleError(err error){
	if err != nil {
			log.Println(err)
			os.Exit(1)
	}
}