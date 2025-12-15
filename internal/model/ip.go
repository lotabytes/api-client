package model

import "net/netip"

type IPAddress = netip.Addr

func ParseAddr(ipAddr string) (IPAddress, error) {
	return netip.ParseAddr(ipAddr)
}

func MustParseAddr(ipAddr string) IPAddress {
	return netip.MustParseAddr(ipAddr)
}
