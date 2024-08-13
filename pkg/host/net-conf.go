package host

import (
	"github.com/H-BF/corlib/pkg/dict"
	"github.com/H-BF/corlib/pkg/nl"
)

type (
	//UpdStrategy update strategy
	UpdStrategy uint

	//LinkID link ID dev
	LinkID int

	//IpDev ip device
	IpDev struct { //nolint:revive
		Name string
		ID   LinkID
	}

	//IpDevs ip devices
	IpDevs struct {
		dict.HDict[LinkID, IpDev] //nolint:revive
	}

	//NetConf network conf
	NetConf struct {
		Devs     IpDevs
		Adresses LinkAddresses
	}
)

const (
	//Update use insert/update
	Update UpdStrategy = iota

	//Delete use delete
	Delete
)

// Upd update devs
func (devs *IpDevs) Upd(d IpDev, how UpdStrategy) bool {
	switch how {
	case Update:
		if v, ok := devs.Get(d.ID); ok && !(v == d) {
			devs.Put(d.ID, d)
		}
	case Delete:
		n := devs.Len()
		devs.Del(d.ID)
		return n-devs.Len() > 0
	default:
		panic("UB")
	}
	return true
}

// Clone makes a copy
func (devs IpDevs) Clone() IpDevs {
	var ret IpDevs
	devs.Iterate(ret.Insert)
	return ret
}

// Eq -
func (devs IpDevs) Eq(other IpDevs) bool {
	return devs.HDict.Eq(&other.HDict, func(vL, vR IpDev) bool {
		return vL == vR
	})
}

// Eq -
func (conf NetConf) Eq(other NetConf) bool {
	return conf.Devs.Eq(other.Devs) &&
		conf.Adresses.Eq(other.Adresses)
}

// Clone -
func (conf NetConf) Clone() NetConf {
	return NetConf{
		Adresses: conf.Adresses.Clone(),
		Devs:     conf.Devs.Clone(),
	}
}

// LocalIPs get effective local unique IP lists
func (conf NetConf) LocalIPs() (ip4set IPvSet[IP4], ip6set IPvSet[IP6]) {
	conf.Adresses.IPSets.Iterate(func(_ LinkID, v IPSet) bool {
		v.IPs.Iterate(func(ip IPAddr) bool {
			if ip.Is4() {
				ip4set.Put(ip.As4())
			} else if ip.Is6() {
				ip6set.Put(ip.As16())
			}
			return true
		})
		return true
	})
	return ip4set, ip6set
}

// UpdFromWatcher updates conf with messages came from netlink-watcher
func (conf *NetConf) UpdFromWatcher(msgs ...nl.WatcherMsg) bool {
	var cnt int
	var ok bool
	for i := range msgs {
		switch v := msgs[i].(type) {
		case nl.AddrUpdateMsg:
			ok = conf.Adresses.Upd(
				LinkID(v.LinkIndex),
				v.Address,
				tern(v.Deleted, Delete, Update))
		case nl.LinkUpdateMsg:
			attrs := v.Link.Attrs()
			ok = conf.Devs.Upd(
				IpDev{
					ID:   LinkID(attrs.Index),
					Name: attrs.Name,
				},
				tern(v.Deleted, Delete, Update),
			)
		}
		cnt += tern(ok, 1, 0)
	}
	return cnt > 0
}

func tern[tval any](cond bool, v1, v2 tval) tval {
	if cond {
		return v1
	}
	return v2
}
