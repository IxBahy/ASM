package nmap

import (
	"encoding/xml"

	"github.com/IxBahy/ASM/internal/scanners"
	"github.com/IxBahy/ASM/pkg/client"
)

type NmapScanner struct {
	*scanners.BaseScanner
	installClient client.ToolInstaller
}
type NmapRun struct {
	XMLName xml.Name `xml:"nmaprun"`
	Hosts   []Host   `xml:"host"`
}

type Host struct {
	Ports []Port `xml:"ports>port"`
}

type Port struct {
	PortID       string   `xml:"portid,attr"`
	Protocol     string   `xml:"protocol,attr"`
	StateDetails State    `xml:"state"`
	Service      Service  `xml:"service"`
	Scripts      []Script `xml:"script"`
}

type Script struct {
	Name   string `xml:"id,attr"`
	Result string `xml:"output,attr"`
}

type State struct {
	Value     string `xml:"state,attr"`
	Reason    string `xml:"reason,attr"`
	ReasonTTL string `xml:"reason_ttl,attr"`
}

type Service struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr,omitempty"`
}
