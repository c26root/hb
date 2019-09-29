package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

func center(s string, width int) string {
	n := width - len(s)
	if n <= 0 {
		return s
	}
	half := n / 2
	if n%2 != 0 && width%2 != 0 {
		half = half + 1
	}
	return strings.Repeat(" ", half) + s + strings.Repeat(" ", (n-half))
}

func rjust(s string, width int) string {
	n := width - len(s)
	if n <= 0 {
		return s
	}
	return strings.Repeat(" ", n) + s
}

func ljust(s string, width int) string {
	n := width - len(s)
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}

func IsValidIPV4(ip string) bool {
	b := net.ParseIP(ip)
	if b.To4() == nil {
		return false
	}
	return true
}

func ParsePort(portString string) ([]int, error) {

	var portList []int

	pair := strings.Split(portString, ",")
	for _, item := range pair {
		if strings.Contains(item, "-") {
			portRange := strings.Split(item, "-")
			if len(portRange) != 2 {
				return portList, fmt.Errorf("%s is invalid port range", portString)
			}
			start, _ := strconv.Atoi(portRange[0])
			end, _ := strconv.Atoi(portRange[1])
			for i := start; i <= end; i++ {
				portList = append(portList, i)
			}
		} else {
			if item != "" {
				item, _ := strconv.Atoi(item)
				portList = append(portList, item)
			}
		}
	}

	sort.Ints(portList)
	return portList, nil
}

func isValidIPPart(s interface{}) bool {
	var i int
	switch s.(type) {
	case int:
		i = s.(int)
	case string:
		var err error
		i, err = strconv.Atoi(s.(string))
		if err != nil {
			return false
		}
	}
	return i >= 0 && i <= 256
}

func isDomain(s string) bool {
	ns, err := net.LookupHost(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s", err.Error())
		return false
	}
	for _, n := range ns {
		if n == s {
			return false
		}
	}
	return true
}

func ParseIP(ipString string) ([]string, error) {
	ipList := []string{}

	pair := strings.Split(ipString, ",")
	for _, item := range pair {
		// single ip 192.168.1.1
		if net.ParseIP(item) != nil {
			ipList = append(ipList, item)
			// cidr 192.168.1.1/24
		} else if ip, network, err := net.ParseCIDR(item); err == nil {
			s := []string{}
			for ip := ip.Mask(network.Mask); network.Contains(ip); increaseIP(ip) {
				s = append(s, ip.String())
			}
			for _, ip := range s[1 : len(s)-1] {
				ipList = append(ipList, ip)
			}
		} else if strings.Contains(item, "-") {
			if isDomain(item) {
				ipList = append(ipList, ipString)
				continue
			}
			// 192.168.1.1-254
			// 192.168.1-254.1
			// 192.168.1-2.5-6
			ipSplit := strings.Split(item, ".")
			tmpSlice := make([][]int, 4)
			for idx, val := range strings.Split(item, ".") {
				valSplit := strings.Split(val, "-")
				if len(valSplit) == 1 {
					if !isValidIPPart(val) {
						return []string{}, fmt.Errorf("Invalid ip range %s", item)
					}
				}
				if len(valSplit) != 2 {
					continue
				}

				if !isValidIPPart(valSplit[0]) || !isValidIPPart(valSplit[1]) {
					return []string{}, fmt.Errorf("Invalid ip range %s", item)
				}

				start, _ := strconv.Atoi(valSplit[0])
				end, _ := strconv.Atoi(valSplit[1])
				if start < end {
					var tmpInts []int
					for i := start; i <= end; i++ {
						tmpInts = append(tmpInts, i)
					}
					tmpSlice[idx] = tmpInts
				}
			}
			for idx, vals := range tmpSlice {
				if len(vals) < 2 {
					continue
				}
				tmpIPSplit := ipSplit
				for _, val := range vals {
					tmpIPSplit[idx] = strconv.Itoa(val)
					ipString := strings.Join(tmpIPSplit, ".")
					ipList = append(ipList, ipString)
				}
			}
		} else {
			if ipString != "" && isDomain(item) {
				ipList = append(ipList, item)
			}
		}
	}
	// return ipList, fmt.Errorf("%s is not an IP Address or CIDR Network", item)
	return ipList, nil
}

// LinesToIPList processes a list of IP addresses or networks in CIDR format.
// Returning a list of all possible IP addresses.
func LinesToIPList(lines []string) ([]string, error) {
	ipList := []string{}
	for _, line := range lines {
		_ipList, err := ParseIP(line)
		if err != nil {
			return _ipList, fmt.Errorf("%s is not an IP Address", line)
		}
		for _, line := range _ipList {
			ipList = append(ipList, line)
		}
	}
	return ipList, nil
}

// increases an IP by a single address.
func increaseIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func isStartingIPLower(start, end net.IP) bool {
	if len(start) != len(end) {
		return false
	}
	for i := range start {
		if start[i] > end[i] {
			return false
		}
	}
	return true
}

func ParseLines(l []string) []string {
	var lines []string
	for _, line := range l {
		ips, _ := ParseIP(line)
		if len(ips) != 0 {
			for _, ip := range ips {
				lines = append(lines, ip)
			}
		} else {
			lines = append(lines, line)
		}
	}
	return lines
}

// ReadFileLines returns all the lines in a file.
func ReadFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
