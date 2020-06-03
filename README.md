# 检测网站IPv4/v6的HTTP、HTTPS、HTTP2服务状态

## 使用说明  
### 编译
```
#go mod init main
go build httpstatus.go
```
### 下载

直接下载编译[releases](https://github.com/abanger/httpstatus/releases)

### 使用


```
httpstatus -h example.com  #（域名） 
```


### 说明

url | IPv4 | httpv4 | httpsv4 | http2v4 | IPv6 | httpipv6 | httpsv6 | http2v6 
---|---|---|---|---|---|---|---|---
example.com | 127.0.0.1  | 1 | 1 | 1 | ::1: | 1 | 1 | 1 

结果：0为不能到达，<9为正常，9为出错。