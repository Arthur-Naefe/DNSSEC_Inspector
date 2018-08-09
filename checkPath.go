package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

const (
	ZSK uint16 = 256
	KSK uint16 = 257
)

type noDSerror struct {
	key dns.DNSKEY
	msg string
}

func (e *noDSerror) Error() string {
	return fmt.Sprintf("%v\n -- Error: %s", e.key.String(), e.msg)
}

/* Checks the complete path up to the trust anchor
 */
func checkPath(fqdn string, res *Result) {
	anchor := false
	f := strings.Split(fqdn, ".")
	l := len(f) - 1
	var zoneList []string
	for i := range f {
		zoneList = append([]string{strings.Join(f[l-i:], ".")}, zoneList...)
	}
	zoneList = append(zoneList, ".")
	for _, fqdn = range zoneList {
		z := &Zone{}
		z.FQDN = fqdn
		checkNSEC3Existence(fqdn, z)
		// 1. Check section (includes checking the validation of ZSK)
		checkRRValidation(fqdn, z)
		// 2. Check KSK validity by checking with DS RR in higher z one
		m := dnssecQuery(fqdn, dns.TypeDNSKEY, "")
		keys := getDNSKEYs(m, ZSK)
		keyRes1 := make([]Key, len(keys))
		for i, k := range keys {
			checkKey(k, &keyRes1[i])
		}
		keys = getDNSKEYs(m, KSK)
		keyRes2 := make([]Key, len(keys))
		for i, k := range keys {
			checkKSKverifiability(fqdn, k, &keyRes2[i])
			checkKey(k, &keyRes2[i])
			if keyRes2[i].TrustAnchor {
				anchor = true
			}
		}
		z.Keys = append(keyRes1, keyRes2...)
		z.KeyCount = len(z.Keys)
		res.Zones = append(res.Zones, *z)
		if anchor == true {
			if fqdn != "." {
				res.TrustIsland = true
				res.TrustIslandAnchorZone = fqdn
			}
			break
		}
	}
	return
}

// Checks the validity of a KSK DNSKEY RR by checking the DS RR in the authoritative zone above
func checkKSKverifiability(fqdn string, key dns.DNSKEY, res *Key) (bool, error) {
	ds, err := getDSforKey(fqdn, key)
	res.Verifiable = false
	res.TrustAnchor = false
	if err == nil {
		newDS := key.ToDS(ds.DigestType)
		if ds.Digest == (*newDS).Digest {
			res.Verifiable = true
			return true, nil
		} else {
			return false, errors.New("DS does not match")
		}
	}
	res.TrustAnchor = true
	return false, err
}

// Gets the DS RR for a given key
func getDSforKey(fqdn string, key dns.DNSKEY) (dns.DS, error) {
	m := dnssecQuery(fqdn, dns.TypeDS, "")
	for _, r := range m.Answer {
		if r.Header().Rrtype == dns.TypeDS {
			if r.(*dns.DS).KeyTag == key.KeyTag() {
				return *(r.(*dns.DS)), nil
			}
		}
	}
	return dns.DS{}, errors.New("No DS RR for given key")
}

/* Removes all RRSIG artefacts from list and returns a list of true DNSKEY RRs
 */
func getDNSKEYs(m dns.Msg, t uint16) (ret []dns.DNSKEY) {
	for _, r := range m.Answer {
		if r.Header().Rrtype == dns.TypeDNSKEY && r.(*dns.DNSKEY).Flags == t {
			ret = append(ret, *r.(*dns.DNSKEY))
		}
	}
	return
}

/* Filters all RRsigs covering a given type t and returns them.
 */
func getRRsigs(m dns.Msg, t uint16) (ret []*dns.RRSIG) {
	x := m.Answer
	if t == dns.TypeNS {
		x = m.Ns
	}
	for _, r := range x {
		if r.Header().Rrtype == dns.TypeRRSIG && r.(*dns.RRSIG).TypeCovered == t {
			ret = append(ret, r.(*dns.RRSIG))
		}
	}
	return
}
