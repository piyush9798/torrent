package iplist

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sort"
)

type IPList struct {
	ranges []Range
}

type Range struct {
	First, Last net.IP
	Description string
}

func (r *Range) String() string {
	return fmt.Sprintf("%s-%s (%s)", r.First, r.Last, r.Description)
}

// Create a new IP list. The given range must already sorted by the lower IP
// in the range. Behaviour is undefined for lists of overlapping ranges.
func New(initSorted []Range) *IPList {
	return &IPList{
		ranges: initSorted,
	}
}

func (me *IPList) NumRanges() int {
	if me == nil {
		return 0
	}
	return len(me.ranges)
}

// Return the range the given IP is in. Returns nil if no range is found.
func (me *IPList) Lookup(ip net.IP) (r *Range) {
	if me == nil {
		return nil
	}
	// TODO: Perhaps all addresses should be converted to IPv6, if the future
	// of IP is to always be backwards compatible. But this will cost 4x the
	// memory for IPv4 addresses?
	if v4 := ip.To4(); v4 != nil {
		r = me.lookup(v4)
		if r != nil {
			return
		}
	}
	if v6 := ip.To16(); v6 != nil {
		return me.lookup(v6)
	}
	return nil
}

// Return the range the given IP is in. Returns nil if no range is found.
func (me *IPList) lookup(ip net.IP) (r *Range) {
	// Find the index of the first range for which the following range exceeds
	// it.
	i := sort.Search(len(me.ranges), func(i int) bool {
		if i+1 >= len(me.ranges) {
			return true
		}
		return bytes.Compare(ip, me.ranges[i+1].First) < 0
	})
	if i == len(me.ranges) {
		return
	}
	r = &me.ranges[i]
	if bytes.Compare(ip, r.First) < 0 || bytes.Compare(ip, r.Last) > 0 {
		r = nil
	}
	return
}

// Parse a line of the PeerGuardian Text Lists (P2P) Format. Returns !ok but
// no error if a line doesn't contain a range but isn't erroneous, such as
// comment and blank lines.
func ParseBlocklistP2PLine(l []byte) (r Range, ok bool, err error) {
	l = bytes.TrimSpace(l)
	if len(l) == 0 || bytes.HasPrefix(l, []byte("#")) {
		return
	}
	colon := bytes.IndexByte(l, ':')
	if colon == -1 {
		err = errors.New("missing colon")
		return
	}
	hyphen := bytes.IndexByte(l[colon+1:], '-')
	if hyphen == -1 {
		err = errors.New("missing hyphen")
		return
	}
	hyphen += colon + 1
	r.Description = string(l[:colon])
	r.First = net.ParseIP(string(l[colon+1 : hyphen]))
	r.Last = net.ParseIP(string(l[hyphen+1:]))
	if r.First == nil || r.Last == nil {
		return
	}
	ok = true
	return
}
