//go:build linux
// +build linux

package nftables

import (
	"math/rand"
	"net"
	"strings"
	"sync/atomic"

	rc "github.com/H-BF/corlib/pkg/net/resources"
	"github.com/c-robinson/iplib"
	nftlib "github.com/google/nftables"
)

func tern[T any](cond bool, a1, a2 T) T {
	if cond {
		return a1
	}
	return a2
}

func sli[T any](d ...T) []T {
	return d
}

// NextSetID -
func NextSetID() uint32 {
	return atomic.AddUint32(&setID, 1)
}

var setID = rand.Uint32() //nolint:gosec

func zs(s string) string { //nolint:unused
	const z = "\x00"
	if n := len(s); n > 0 {
		n1 := strings.LastIndex(s, z)
		if n1 >= 0 && (n-n1) == 1 {
			return s
		}
	}
	return s + z
}

// Nets2SetElements -
func Nets2SetElements(nets []net.IPNet, ipV int) []nftlib.SetElement {
	const (
		b32  = 32
		b128 = 128
	)
	var elements []nftlib.SetElement
	for i := range nets {
		nw := nets[i]
		ones, _ := nw.Mask.Size()
		netIf := iplib.NewNet(nw.IP, ones)
		ipLast := iplib.NextIP(netIf.LastAddress())
		switch ipV {
		case iplib.IP4Version:
			ipLast = tern(ones < b32, iplib.NextIP(ipLast), ipLast)
		case iplib.IP6Version:
			ipLast = tern(ones < b128, iplib.NextIP(ipLast), ipLast)
		default:
			return nil
		}
		////TODO: need expert opinion
		//elements = append(elements, nftLib.SetElement{
		//	Key:    nw.IP,
		//	KeyEnd: ipLast,
		//})
		elements = append(elements,
			nftlib.SetElement{
				Key: nw.IP,
			},
			nftlib.SetElement{
				IntervalEnd: true,
				Key:         ipLast,
			})
	}
	return elements
}

// TransormPortRanges -
func TransormPortRanges(pr rc.PortRanges) (ret [][2]rc.PortNumber) {
	ret = make([][2]rc.PortNumber, 0, pr.Len())
	pr.Iterate(func(r rc.PortRange) bool {
		a, b := r.Bounds()
		var x [2]rc.PortNumber
		x[0], _ = a.AsIncluded().GetValue()
		x[1], _ = b.AsIncluded().GetValue()
		ret = append(ret, x)
		return true
	})
	return ret
}
