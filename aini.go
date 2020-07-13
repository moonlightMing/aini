package aini

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"io"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"github.com/flynn/go-shlex"
)

type groups struct {
	*orderedmap.OrderedMap
}

func (g groups) Get(key string) ([]Host,bool) {
	if hosts, ok := g.OrderedMap.Get(key); ok {
		return hosts.([]Host), ok
	} else {
		return nil, ok
	}
}

func (g groups) Set(key string, hosts []Host)  {
	g.OrderedMap.Set(key, hosts)
}

type Hosts struct {
	input  *bufio.Reader
	Groups *groups
}

type Host struct {
	Name       string
	Port       int
	User       string
	Pass       string
	PrivateKey string
	PublicIP   string
}

func NewFile(f string) (*Hosts, error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return &Hosts{}, err
	}

	h, err := NewParser(bytes.NewReader(bs))
	if err != nil {
		return &Hosts{}, err
	}

	return h, nil
}

func NewParser(r io.Reader) (*Hosts, error) {
	input := bufio.NewReader(r)
	hosts := &Hosts{input: input}
	hosts.parse()
	return hosts, nil
}

func (h *Hosts) parse() error {
	scanner := bufio.NewScanner(h.input)
	activeGroupName := "ungrouped"
	h.Groups = &groups{
		orderedmap.New(),
	}

	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		// fmt.Println(activeGroupName, ":", line)

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			replacer := strings.NewReplacer("[", "", "]", "")
			activeGroupName = replacer.Replace(line)

			if _, ok := h.Groups.Get(activeGroupName); !ok {
				h.Groups.Set(activeGroupName, make([]Host, 0))
			}
		} else if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || line == "" {
			// do nothing
		} else if activeGroupName != "" {
			parts, err := shlex.Split(line)
			if err != nil {
				fmt.Println("couldn't tokenize: ", line)
			}
			host := getHost(parts)

			group, _ := h.Groups.Get(activeGroupName)
			h.Groups.Set(activeGroupName, append(group, host))
		}
	}
	return nil
}

func (h *Hosts) Match(m string) []Host {
	matchedHosts := make([]Host, 0, 5)

	for _, hostsKey := range h.Groups.Keys() {
		hosts, _ := h.Groups.Get(hostsKey)
		for _, host := range hosts {
			if m, err := path.Match(m, host.Name); err == nil && m {
				matchedHosts = append(matchedHosts, host)
			}
		}
	}

	return matchedHosts
}

func getHost(parts []string) Host {
	hostname := parts[0]
	if (strings.Contains(hostname, "[") &&
		strings.Contains(hostname, "]") &&
		strings.Contains(hostname, ":") &&
		(strings.LastIndex(hostname, "]") < strings.LastIndex(hostname, ":"))) ||
		(!strings.Contains(hostname, "]") && strings.Contains(hostname, ":")) {

		splithost := strings.Split(hostname, ":")
		hostname = splithost[0]
	}
	params := parts[1:]
	host := Host{Name: hostname}
	parseParameters(params, &host)
	return host
}

func parseParameters(params []string, host *Host) {
	for _, p := range params {
		switch {
		case strings.Contains(p, "ansible_user"):
			host.User = strings.Split(p, "=")[1]
		case strings.Contains(p, "ansible_ssh_port"):
			host.Port, _ = strconv.Atoi(strings.Split(p, "=")[1])
		case strings.Contains(p, "ansible_ssh_pass"):
			host.Pass = strings.Split(p, "=")[1]
		case strings.Contains(p, "ansible_ssh_private_key_file"):
			host.PrivateKey = strings.Split(p, "=")[1]
		case strings.Contains(p, "public_ip"):
			host.PublicIP = strings.Split(p, "=")[1]
		}
	}
}
