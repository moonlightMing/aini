package aini

import (
	"strings"
	"testing"
)

var input1 string = `
myhost1

[dbs]
dbhost1
dbhost2

[apps]
my-app-server1
my-app-server2 ansible_ssh_port=3000 public_ip=192.168.31.33

`

func createHosts(input string) Hosts {
	testInput := strings.NewReader(input)
	v, _ := NewParser(testInput)
	return *v
}

func createHostsFromFile(f string) Hosts {
	v, _ := NewFile(f)
	return *v
}

func TestGroupExists(t *testing.T) {
	v := createHosts(input1)
	matched := false
	if _, ok := v.Groups.Get("dbs"); ok {
		matched = true
	}
	if !matched {
		t.Error("Expected to find the group \"dbs\"")
	}
}

func TestHostExistsInGroups(t *testing.T) {
	v := createHosts(input1)
	exportedHosts := map[string][]Host{
		"dbs": []Host{Host{Name: "dbhost1"},
			Host{Name: "dbhost2"}},
		"ungrouped": []Host{Host{Name: "myhost1"}},
		"apps":      []Host{Host{Name: "my-app-server2", Port: 3000, PublicIP: "192.168.31.33"}},
	}

	for group, ehosts := range exportedHosts {
		for _, ehost := range ehosts {
			if hosts, ok := v.Groups.Get(group); ok {
				matched := false
				for _, host := range hosts {
					if host.Name == ehost.Name {
						matched = true
						if host.Port != ehost.Port {
							t.Errorf("Host port '%v' does not match expected port of '%v'.\n", host.Port, ehost.Port)
						}
						if host.PublicIP != ehost.PublicIP {
							t.Errorf("Host public_ip '%v' does not match expected public_ip of '%v'\n", host.PublicIP, ehost.PublicIP)
						}
					}
				}
				if !matched {
					t.Errorf("Server '%+v' was not found in '%+v'.\n", ehost.Name, group)
				}
			} else {
				t.Errorf("'%v' group doesn't exist.\n", group)
			}
		}

	}
}

func TestHostMatching(t *testing.T) {
	v := createHosts(input1)
	hosts := v.Match("db*")
	if len(hosts) != 2 {
		t.Errorf("Number of hosts don't match 2, but were %v instead", len(hosts))
	}
	if !matchHosts(hosts, "dbhost1") {
		t.Errorf("dbhost1 should be in the list, but found %+v", hosts)
	}
	if !matchHosts(hosts, "dbhost2") {
		t.Errorf("dbhost2 should be in the list, but found %+v", hosts)
	}
}

func TestFromFileGroupExists(t *testing.T) {
	v := createHostsFromFile("sample_hosts")
	matched := false
	if _, ok := v.Groups.Get("dbs"); ok {
		matched = true
	}
	if !matched {
		t.Error("Expected to find the group \"dbs\"")
	}

}

func matchHosts(hosts []Host, hostToMatch string) bool {
	match := false
	for _, host := range hosts {
		if host.Name == hostToMatch {
			match = true
		}
	}
	return match
}

func TestReadSSHParameters(t *testing.T) {
	expectedHosts := []Host{
		Host{Name: "sql-host1", Port: 3306, User: "ubuntu", Pass: "ubuntu"},
		Host{Name: "sql-host2", Port: 3306, User: "ubuntu", PrivateKey: "/tmp/some/key"},
	}
	i := createHostsFromFile("sample_hosts")
	sqlGroup,_ := i.Groups.Get("sql")
	for i, host := range sqlGroup {
		if expectedHosts[i].User != host.User {
			t.Errorf("mismatched users: %v / %v", expectedHosts[i], host)
		}
		if expectedHosts[i].Pass != host.Pass {
			t.Errorf("mismatched pass: %v / %v", expectedHosts[i], host)
		}
		if expectedHosts[i].PrivateKey != host.PrivateKey {
			t.Errorf("mismatched private key: %v / %v", expectedHosts[i], host)
		}
	}
}
