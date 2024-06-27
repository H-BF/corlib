//go:build linux
// +build linux

package nftables

import (
	"fmt"
	"net"

	di "github.com/H-BF/corlib/pkg/dict"
	rc "github.com/H-BF/corlib/pkg/net/resources"

	"github.com/c-robinson/iplib"
	nftlib "github.com/google/nftables"
	bu "github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// RuleBuilder -
type RuleBuilder struct {
	sets  di.HDict[uint32, NfSet]
	exprs []expr.Any
}

// ApplyRule -
func (rb RuleBuilder) ApplyRule(chn *nftlib.Chain, c *nftlib.Conn) {
	if len(rb.exprs) > 0 {
		rb.sets.Iterate(func(id uint32, s NfSet) bool {
			s.Table = chn.Table
			if e := c.AddSet(s.Set, s.Elements); e != nil {
				panic(e)
			}
			return true
		})
		_ = c.AddRule(&nftlib.Rule{
			Table: chn.Table,
			Chain: chn,
			Exprs: rb.exprs,
		})
	}
}

// DLogs -
func (rb RuleBuilder) DLogs(f expr.LogFlags) RuleBuilder { //nolint:unparam
	rb.exprs = append(rb.exprs,
		&expr.Log{
			Flags: f,
			Level: expr.LogLevelDebug,
			Key: (1<<unix.NFTA_LOG_FLAGS)*tern(f == 0, uint32(0), 1) |
				(1 << unix.NFTA_LOG_LEVEL),
		})
	return rb
}

// Accept -
func (rb RuleBuilder) Accept() RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Verdict{Kind: expr.VerdictAccept},
	)
	return rb
}

// Drop -
func (rb RuleBuilder) Drop() RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Verdict{Kind: expr.VerdictDrop},
	)
	return rb
}

// RuleAction2Verdict -
func (rb RuleBuilder) RuleAction2Verdict(a RuleAction) RuleBuilder {
	var f func() RuleBuilder
	switch a {
	case RA_ACCEPT:
		f = rb.Accept
	case RA_DROP:
		f = rb.Drop
	}
	return f()
}

// Jump -
func (rb RuleBuilder) Jump(chain string) RuleBuilder { //nolint:unused
	rb.exprs = append(rb.exprs,
		&expr.Verdict{Kind: expr.VerdictJump, Chain: chain},
	)
	return rb
}

// GoTo -
func (rb RuleBuilder) GoTo(chain string) RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Verdict{Kind: expr.VerdictGoto, Chain: chain},
	)
	return rb
}

// Counter -
func (rb RuleBuilder) Counter() RuleBuilder {
	rb.exprs = append(rb.exprs, &expr.Counter{})
	return rb
}

// InSet -
func (rb RuleBuilder) InSet(s *nftlib.Set) RuleBuilder {
	if s != nil {
		n := s.Name
		if s.Anonymous {
			n = fmt.Sprintf(s.Name, s.ID)
		}
		rb.exprs = append(rb.exprs,
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        n,
				SetID:          s.ID,
			})
	}
	return rb
}

// SAddr -
func (rb RuleBuilder) SAddr(ipVer int) RuleBuilder {
	switch ipVer {
	case iplib.IP4Version:
		return rb.SAddr4()
	case iplib.IP6Version:
		return rb.SAddr6()
	default:
		panic(fmt.Errorf("unsuppoeted proto ver '%v'", ipVer))
	}
}

// DAddr -
func (rb RuleBuilder) DAddr(ipVer int) RuleBuilder {
	switch ipVer {
	case iplib.IP4Version:
		return rb.DAddr4()
	case iplib.IP6Version:
		return rb.DAddr6()
	default:
		panic(fmt.Errorf("unsuppoeted proto ver '%v'", ipVer))
	}
}

// SAddr6 -
func (rb RuleBuilder) SAddr6() RuleBuilder {
	rb.exprs = append(rb.IP6().exprs,
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       uint32(8),  //nolint:mnd
			Len:          uint32(16), //nolint:mnd
		},
	)
	return rb
}

// DAddr6 -
func (rb RuleBuilder) DAddr6() RuleBuilder {
	rb.exprs = append(rb.IP6().exprs,
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       uint32(24), //nolint:mnd
			Len:          uint32(16), //nolint:mnd
		},
	)
	return rb
}

// SAddr4 -
func (rb RuleBuilder) SAddr4() RuleBuilder {
	rb.exprs = append(rb.IP4().exprs, //ip
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       uint32(12), //nolint:mnd
			Len:          uint32(4),  //nolint:mnd
		}, //saddr
	)
	return rb
}

// DAddr4 -
func (rb RuleBuilder) DAddr4() RuleBuilder {
	rb.exprs = append(rb.IP4().exprs, //ip
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       uint32(16), //nolint:mnd
			Len:          uint32(4),  //nolint:mnd
		}, //daddr
	)
	return rb
}

// SPort -
func (rb RuleBuilder) SPort() RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       0,
			Len:          2,
		},
	)
	return rb
}

// DPort -
func (rb RuleBuilder) DPort() RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       2,
			Len:          2,
		},
	)
	return rb
}

// MetaL4PROTO -
func (rb RuleBuilder) MetaL4PROTO() RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
	)
	return rb
}

// ProtoIP -
func (rb RuleBuilder) ProtoIP(tr rc.NetworkTransport) RuleBuilder {
	var t byte
	switch tr {
	case rc.TCP:
		t = unix.IPPROTO_TCP
	case rc.UDP:
		t = unix.IPPROTO_UDP
	default:
		panic("UB")
	}
	rb.exprs = append(rb.MetaL4PROTO().exprs,
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     []byte{t},
		},
	)
	return rb
}

// ProtoICMP -
func (rb RuleBuilder) ProtoICMP(d rc.ICMP) RuleBuilder {
	var proto byte
	switch d.IPv {
	case rc.IPv4:
		proto = unix.IPPROTO_ICMP
	case rc.IPv6:
		proto = unix.IPPROTO_ICMPV6
	default:
		panic(
			errors.Errorf("unsusable proto family(%v)", d.IPv),
		)
	}
	rb.exprs = append(rb.MetaL4PROTO().exprs,
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     []byte{proto},
		},
	)
	if n := d.Types.Len(); n > 0 {
		set := &nftlib.Set{
			ID:        NextSetID(),
			Name:      "__set%d",
			Anonymous: true,
			Constant:  true,
			KeyType: tern(d.IPv == rc.IPv4,
				nftlib.TypeICMPType, nftlib.TypeICMP6Type),
		}
		elements := make([]nftlib.SetElement, 0, n)
		d.Types.Iterate(func(v uint8) bool {
			elements = append(elements,
				nftlib.SetElement{Key: []byte{v}},
			)
			return true
		})
		rb.exprs = append(rb.exprs,
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       0,
				Len:          1,
			},
		)
		rb = rb.InSet(set)
		rb.sets.Put(set.ID, NfSet{Set: set, Elements: elements})
	}
	return rb
}

// IP4 -
func (rb RuleBuilder) IP4() RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Meta{Key: expr.MetaKeyNFPROTO, Register: 1},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     []byte{unix.NFPROTO_IPV4},
		}, //ip
	)
	return rb
}

// IP6 -
func (rb RuleBuilder) IP6() RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Meta{Key: expr.MetaKeyNFPROTO, Register: 1},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     []byte{unix.NFPROTO_IPV6},
		}, //ip6
	)
	return rb
}

// CTState -
func (rb RuleBuilder) CTState(ctStateBitMask uint32) RuleBuilder {
	rb.exprs = append(rb.exprs,
		&expr.Ct{Key: expr.CtKeySTATE, Register: 1},
		&expr.Bitwise{
			SourceRegister: 1,
			DestRegister:   1,
			Len:            4,
			Mask:           bu.NativeEndian.PutUint32(ctStateBitMask),
			Xor:            bu.NativeEndian.PutUint32(0),
		},
		&expr.Cmp{
			Op:       expr.CmpOpNeq,
			Data:     bu.NativeEndian.PutUint32(0),
			Register: 1,
		},
	)
	return rb
}

// IIF -
func (rb RuleBuilder) IIF() RuleBuilder { //nolint:unused
	rb.exprs = append(rb.exprs,
		&expr.Meta{Key: expr.MetaKeyIIFNAME, Register: 1},
	)
	return rb
}

// OIF -
func (rb RuleBuilder) OIF() RuleBuilder { //nolint:unused
	rb.exprs = append(rb.exprs,
		&expr.Meta{Key: expr.MetaKeyOIFNAME, Register: 1},
	)
	return rb
}

// NeqS -
func (rb RuleBuilder) NeqS(s string) RuleBuilder { //nolint:unused
	rb.exprs = append(rb.exprs,
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpNeq,
			Data:     bu.PutString(zs(s)),
		},
	)
	return rb
}

// EqU16 -
func (rb RuleBuilder) EqU16(val uint16) RuleBuilder {
	return rb.CmpU16(expr.CmpOpEq, val)
}

// LeU16 -
func (rb RuleBuilder) LeU16(val uint16) RuleBuilder {
	return rb.CmpU16(expr.CmpOpLte, val)
}

// LtU16 -
func (rb RuleBuilder) LtU16(val uint16) RuleBuilder { //nolint:unused
	return rb.CmpU16(expr.CmpOpLt, val)
}

// GeU16 -
func (rb RuleBuilder) GeU16(val uint16) RuleBuilder {
	return rb.CmpU16(expr.CmpOpGte, val)
}

// GtU16 -
func (rb RuleBuilder) GtU16(val uint16) RuleBuilder { //nolint:unused
	return rb.CmpU16(expr.CmpOpGt, val)
}

// CmpU16 -
func (rb RuleBuilder) CmpU16(op expr.CmpOp, val uint16) RuleBuilder {
	rb.exprs = append(rb.exprs, &expr.Cmp{
		Register: 1,
		Op:       op,
		Data:     bu.BigEndian.PutUint16(val),
	})
	return rb
}

// EqS -
func (rb RuleBuilder) EqS(s string) RuleBuilder { //nolint:unused
	rb.exprs = append(rb.exprs, &expr.Cmp{
		Register: 1,
		Op:       expr.CmpOpEq,
		Data:     bu.PutString(zs(s)),
	})
	return rb
}

// MetaNFTRACE -
func (rb RuleBuilder) MetaNFTRACE(on bool) RuleBuilder {
	if on {
		rb.exprs = append(rb.exprs,
			&expr.Immediate{
				Register: 1,
				Data:     []byte{1},
			},
			&expr.Meta{
				Key:            expr.MetaKeyNFTRACE,
				Register:       1,
				SourceRegister: true,
			}, //meta nftrace set 1|0
		)
	}
	return rb
}

// SrcOrDstSingleIpNet -
func (rb RuleBuilder) SrcOrDstSingleIpNet(n net.IPNet, isSource bool) RuleBuilder {
	var isIP4 bool
	switch len(n.IP) {
	case net.IPv4len:
		isIP4 = true
	case net.IPv6len:
	default:
		panic(
			errors.Errorf("wrong IPNet '%s'", n),
		)
	}
	set := NfSet{
		Elements: Nets2SetElements(sli(n),
			tern(isIP4, iplib.IP4Version, iplib.IP6Version)),
		Set: &nftlib.Set{
			ID:        NextSetID(),
			Name:      "__set%d",
			Constant:  true,
			KeyType:   tern(isIP4, nftlib.TypeIPAddr, nftlib.TypeIP6Addr),
			Interval:  true,
			Anonymous: true,
		},
	}
	_ = rb.sets.Insert(set.ID, set)

	return tern(isSource,
		tern(isIP4, rb.SAddr4, rb.SAddr6),
		tern(isIP4, rb.DAddr4, rb.DAddr6),
	)().InSet(set.Set)
}

// NDPI -
func (rb RuleBuilder) NDPI(dom rc.FQDN, protocols ...string) RuleBuilder { //nolint:unused
	n, e := expr.NewNdpi(expr.NdpiWithHost(dom.String()), expr.NdpiWithProtocols(protocols...))
	if e != nil {
		panic(e)
	}
	rb.exprs = append(rb.exprs, n)
	return rb
}

// PutSet -
func (rb *RuleBuilder) PutSet(setID uint32, set NfSet) {
	rb.sets.Put(setID, set)
}
