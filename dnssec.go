/*
Checks:
	1. Existence of DNSSEC Rr
		- DNSKEY, DS, NSEC, NSEC3, NSEC3PARAM, RRSIG
	2. NSEC3 Existence --> Zone Walking
	3. Key Strength
*/

package main

import (
	"os"
	"regexp"
	"strconv"

	"github.com/miekg/dns"
)

const resultsPath string = "./dnssec.json"

type Out struct {
	DNSSEC          bool  `json:"dnssec"`
	NSEC3           bool  `json:"nsec3"`
	Used            bool  `json:"used"`
	KeyCount        int   `json:"keycount"`
	runningRollover bool  `json:"runningRollover"`
	Keys            []Key `json:"keys",omitempty`
}

type Key struct {
	Type      string `json:"type"`
	Hash      string `json:"hash"`
	HComment  string `json:"hComment"`
	HUntil    string `json:"hUntil"`
	Alg       string `json:"alg"`
	keyLength int32  `json:"keyLength"`
	AComment  string `json:"aComment"`
	AUntil    string `json:"aUntil"`
}

func main() {
	internal_id := os.Args[1]
	report_id := os.Args[2]
	hostname := os.Args[3]
	out := Out()
	checkKeys(os.Args[1], &out)
}
// TODO: Throw error handling
func dnssecQuery(fqdn string, rrType uint16) dns.Msg {
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	c := new(dns.Client)
	c.Net = "udp"
	k.dns.Msg)
	k.tion(dns.Fqdn(fqdn), rrType)
	k.tative = true
	k.onDesired = true
	r, _, _ := c.Exchange(k.inHostPort(config.Servers[0], config.Port))
	if r == nil || r.Rcode != dns.RcodeSuccess {
		r = nil
	}
	return *r
}

// Checks if the DNSKEY uses accepted (?) algorithms
// Signature Algorithm:
// 	1. RSA/SHA-256 (IETF Recom)
//  2. RSA/SHA-1 (IETF accepted alternative)
//  -  RSA/MD5 (IETF shouldnt be considered)
//  Key Length: BSI TR-02102-2
//  - RSA min 2048 bit bis 2022
//  - RSA min 3072 bit ab 2023
//  - DSA min 2000 bit bis 2022
//	- ECDSA min 250 bis 2022

func checkKeys(fqdn string, out *Out) {
	r := dnssecQuery(fqdn, dns.TypeDNSKEY)
	for _, i := range r.Answer {
		x := regexp.MustCompile("( +|\t+)").Split(i.String(), -1)
		if x[5] == "3" {
			out.KeyCount = len(x)
			k := Key()
			if x[4] == "256" {
				k.Type = "ZSK"
			} else if x[4] == "257" {
				k.Type = "KSK key strength"
			}
			s, _ := strconv.ParseInt(x[6], 10, 8)
			switch s {
			case 1: // RSA/MD5
				k.Hash = "MD5"
				k.HComment = "NON-COMPLIANT"
			case 3: // DSA/SHA-1
				// Check key length
				k.Hash = "SHA-1"
				k.HComment = "COMPLIANT"
			case 5: // RSA/SHA-1
				// SHA-256 would be better
				k.Hash = "SHA-256"
				k.HComment = "COMPLIANT"
			case 6: // RSA/SHA-1/NSEC3
				// Could be better
				k.Hash = "SHA-1"
				k.HComment = "COMPLIANT"
			case 7: // RSA/SHA-1/NSEC3
				// Could be better
				k.Hash = "SHA-1"
				k.HComment = "COMPLIANT"
			case 8: // RSA/SHA-256
				// Check key length
				// BSI Recommended -> perfectly fine
				k.Hash = "SHA-256"
			case 10: // RSA/SHA-512
				// check key length
				// perfectly fine
				k.Hash = "SHA-512"
			case 13: // ECDSA P-256 (128bit sec) with SHA-256
				// SHA-256 is perfectly fine
				k.Hash = "None"
			case 14: //ECDSA P-384 (192bit sec)
				// perfectly fine
				k.Hash = "None"
			case 15: // Ed25519 (128bit sec)
				k.Hash = "-"
			case 16: // ED448
				k.Hash = "-"
			default:
			}
		}
	}
}



/*
package main

import (
	"encoding/base64"
	//"strconv"
	"fmt"
)

func keyLength(keyIn string) (e, n, l int) {
	// Base64 encoding
	keyBinary := make([]byte, base64.StdEncoding.DecodedLen(len(keyIn)))
	base64.StdEncoding.Decode(keyBinary, []byte(keyIn))

	err := keyBinary
	if err == nil {
		fmt.Println("Error:", err)
		return
	}

	if keyBinary[0] == 0 {
		e := keyBinary[1:3]
		n := keyBinary[3:]
		l := len(n) * 8
		fmt.Printf("e: %s\nn: %s\nl: %d", e, n, l)
	} else {
		// requires import "strconv"
		// e := strconv.ParseInt(keyBinary[1], 2, 64)
		e := keyBinary[1]
		n := keyBinary[2:]
		l := len(n) * 8
		fmt.Printf("e: %s\nn: %s\nl: %d\n", e, n, l)
	}
	return e, n, l
}

func main() {
	keyInput := "AwEAAcPXtQjs85qD8rnBCxGLRcm1Ghc0jWAS8ExiEaKUBK24yp6DpvuqQFevVfFXT3SUcrMw9La9dUHk0ZLFMZTC+irx4+/iaR9UYG6WW7xpWD12l0NotT0Z7GELKk5mCCnWUe72hVolxrvmaMT3J0GcP0FvSqFicuDEjAzYEoGEiYD5"
	keyLength(keyInput)
}
*/
