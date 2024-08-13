package host

import (
	"net"
	"net/netip"

	"github.com/H-BF/corlib/pkg/dict"
)

type (
	// IP4 -
	IP4 [4]byte

	// IP6 -
	IP6 [16]byte

	// IPvList -
	IPvList[ipV IP4 | IP6] []ipV

	// IPvSet -
	IPvSet[ipV IP4 | IP6] dict.HSet[ipV]

	// IPAddr -
	IPAddr struct {
		netip.Addr
	}

	// IPSet -
	IPSet struct {
		IPs dict.RBSet[IPAddr]
	}

	//LinkAddresses -
	LinkAddresses struct {
		IPSets dict.HDict[LinkID, IPSet]
	}
)

// Cmp -
func (addr IPAddr) Cmp(other IPAddr) int {
	return addr.Compare(other.Addr)
}

// Eq -
func (addr IPAddr) Eq(other IPAddr) bool {
	return addr.Compare(other.Addr) == 0
}

func (ipset IPSet) Eq(other IPSet) bool {
	return ipset.IPs.Eq(&other.IPs)
}

// Clone -
func (ipset IPSet) Clone(other IPSet) (ret IPSet) {
	ipset.IPs.Iterate(ret.IPs.Insert)
	return ret
}

// Upd -
func (ipset *IPSet) Upd(ips ...IPAddr) int {
	var cnt int
	for i := range ips {
		if ips[i].IsValid() {
			cnt += tern(ipset.IPs.Insert(ips[i]), 1, 0)
		}
	}
	return cnt
}

// Del -
func (ipset *IPSet) Del(ips ...IPAddr) int {
	cnt := ipset.IPs.Len()
	ipset.IPs.Del(ips...)
	return cnt - ipset.IPs.Len()
}

// Clone -
func (la LinkAddresses) Clone() (ret LinkAddresses) {
	la.IPSets.Iterate(ret.IPSets.Insert)
	return ret
}

// Eq -
func (la LinkAddresses) Eq(other LinkAddresses) bool {
	return la.IPSets.Eq(&other.IPSets, func(vL, vR IPSet) bool {
		return vL.Eq(vR)
	})
}

// Upd -
func (la *LinkAddresses) Upd(lnk LinkID, addr net.IPNet, how UpdStrategy) bool {
	var ipaddr IPAddr
	var ok bool
	var set IPSet
	ipaddr.Addr, ok = netip.AddrFromSlice(addr.IP)
	if !(ok && ipaddr.IsValid()) {
		return false
	}
	set, ok = la.IPSets.Get(lnk)
	switch how {
	case Update:
		if set.Upd(ipaddr) == 0 {
			return false
		}
	case Delete:
		if !ok || set.Del(ipaddr) == 0 {
			return false
		}
	default:
		panic("unreacheable code")
	}
	la.IPSets.Put(lnk, set)
	return true
}

func (lst IPvList[ipV]) NetIPs() []net.IP {
	ret := make([]net.IP, 0, len(lst))
	for i := range lst {
		switch v := any(lst[i]).(type) {
		case IP4:
			ret = append(ret, net.IP(v[:]))
		case IP6:
			ret = append(ret, net.IP(v[:]))
		}
	}
	return ret
}

// IPvSet2List -
func IPvSet2List[ipV IP4 | IP6](set IPvSet[ipV]) (ret IPvList[ipV]) {
	return set.Values()
}
