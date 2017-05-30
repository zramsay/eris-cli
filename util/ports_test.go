package util

import (
	"reflect"
	"testing"
)

var PortAndProtocolTests = []struct {
	in, out string
}{
	{"", "/tcp"},
	{"53/udp", "53/udp"},
	{"53", "53/tcp"},
	{"53/tcp", "53/tcp"},
	{"53/", "53/"},
	{"53/test", "53/test"},
}

func TestPortAndProtocol(t *testing.T) {
	for _, test := range PortAndProtocolTests {
		if actual := PortAndProtocol(test.in); actual != test.out {
			t.Fatalf("expected %q, got %v", test.out, actual)
		}
	}
}

var PortComponentsTests = []struct {
	in, ip, published, exposed string
}{
	{"", "", "", ""},
	{"8080", "", "8080", "8080/tcp"},
	{"8080:8080", "", "8080", "8080/tcp"},
	{"8080:8080/tcp", "", "8080", "8080/tcp"},
	{"53:53/udp", "", "53", "53/udp"},

	{"", "", "", ""},
	{"127.0.0.1:8080:8080", "127.0.0.1", "8080", "8080/tcp"},
	{"127.0.0.1:8080:8080/tcp", "127.0.0.1", "8080", "8080/tcp"},
	{"127.0.0.1:53:53/udp", "127.0.0.1", "53", "53/udp"},
	{"127.0.0.1:40:40/test", "127.0.0.1", "40", "40/test"},
}

func TestPortComponents(t *testing.T) {
	for _, test := range PortComponentsTests {
		ip, published, exposed := PortComponents(test.in)

		if ip != test.ip {
			t.Fatalf("expected IP %q, got %q, input %q", test.ip, ip, test.in)
		}

		if published != test.published {
			t.Fatalf("expected published port %q, got %q, input %q", test.published, published, test.in)
		}

		if exposed != test.exposed {
			t.Fatalf("expected exposed port %q, got %q, input %q", test.exposed, exposed, test.in)
		}

	}
}

var MapPortsTests = []struct {
	name     string
	ports    []string
	mappings []string
	out      map[string]string
}{
	{
		"#1",
		[]string{},
		[]string{},
		map[string]string{},
	},
	{
		"#2",
		[]string{"8080:8081"},
		[]string{"8081"},
		map[string]string{
			"8081/tcp": "8081",
		},
	},
	{
		"#3",
		[]string{"8080:8081"},
		[]string{"8088"},
		map[string]string{
			"8081/tcp": "8088",
		},
	},
	{
		"#4",
		[]string{"8080:8081"},
		[]string{"8088:8080"},
		map[string]string{
			"8081/tcp": "8080",
			"8080/tcp": "8088",
		},
	},
	{
		"#5",
		[]string{"46656", "46657", "1337"},
		[]string{},
		map[string]string{
			"1337/tcp":  "1337",
			"46656/tcp": "46656",
			"46657/tcp": "46657",
		},
	},
	{
		"#6",
		[]string{"9000:9000", "9001:9001", "9002:9002"},
		[]string{},
		map[string]string{
			"9000/tcp": "9000",
			"9001/tcp": "9001",
			"9002/tcp": "9002",
		},
	},
	{
		"#7. mix",
		[]string{"9002:9000", "9001:9001", "9000:9002"},
		[]string{},
		map[string]string{
			"9000/tcp": "9002",
			"9001/tcp": "9001",
			"9002/tcp": "9000",
		},
	},
	{
		"#8. mix",
		[]string{"9002:9000", "9001:9001", "9000:9002"},
		[]string{"6001:9002", "6000:9001", "5999:9000", "5998:8999"},
		map[string]string{
			"9000/tcp": "5999",
			"9001/tcp": "6000",
			"9002/tcp": "6001",
		},
	},

	{
		"#9",
		[]string{"9000:9001", "9001:9002", "9002:9003"},
		[]string{},
		map[string]string{
			"9001/tcp": "9000",
			"9002/tcp": "9001",
			"9003/tcp": "9002",
		},
	},
	{
		"#10",
		[]string{"9000", "9001", "9002"},
		[]string{"9900-"},
		map[string]string{
			"9000/tcp": "9900",
			"9001/tcp": "9901",
			"9002/tcp": "9902",
		},
	},
	{
		"#11",
		[]string{"9002", "9001", "9000"},
		[]string{"9900-"},
		map[string]string{
			"9002/tcp": "9900",
			"9001/tcp": "9901",
			"9000/tcp": "9902",
		},
	},
	{
		"#12",
		[]string{"9000:9000", "9001:9001", "9002:9002"},
		[]string{"5000", "5010", "5020"},
		map[string]string{
			"9000/tcp": "5000",
			"9001/tcp": "5010",
			"9002/tcp": "5020",
		},
	},
	{
		"#13. handling ips",
		[]string{"127.0.0.1:9000:9000", "9001:9001", "127.0.0.1:9002:9002"},
		[]string{"5000", "5010", "5020"},
		map[string]string{
			"9000/tcp": "5000",
			"9001/tcp": "5010",
			"9002/tcp": "5020",
		},
	},
	{
		"#14",
		[]string{"127.0.0.1:9000:9000/tcp", "9001:9001/tcp", "127.0.0.1:9002:9002/tcp"},
		[]string{"5000", "5010", "5020"},
		map[string]string{
			"9000/tcp": "5000",
			"9001/tcp": "5010",
			"9002/tcp": "5020",
		},
	},
	{
		"#15",
		[]string{"9000:9000", "9001:9001", "9002:9002"},
		[]string{"5000-", "6000-"},
		map[string]string{
			"9000/tcp": "5000",
			"9001/tcp": "6000",
			"9002/tcp": "6001",
		},
	},
	{
		"#16. broken range",
		[]string{"9001:9000"},
		[]string{"xxxx-", "6000-"},
		map[string]string{
			"9000/tcp": "9001",
		},
	},
	{
		"#17. broken range",
		[]string{"9001:9000"},
		[]string{"-", "6000-"},
		map[string]string{
			"9000/tcp": "9001",
		},
	},
}

func TestMapPorts(t *testing.T) {
	for _, test := range MapPortsTests {
		if actual := MapPorts(test.ports, test.mappings); !reflect.DeepEqual(actual, test.out) {
			t.Fatalf("%s. expected %v, got %v, input ports:%v, mappings:%v", test.name, test.out, actual, test.ports, test.mappings)
		}
	}
}
