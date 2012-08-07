package main

import (
	"dns"
	"strings"
)

// fks config
type Config struct {
	Zones map[string]*dns.Zone
}

func NewConfig() *Config {
	c := new(Config)
	c.Zones = make(map[string]*dns.Zone)
	return c
}

func formerr(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	w.Write(m.SetRcode(req, dns.RcodeFormatError))
}

func config(w dns.ResponseWriter, req *dns.Msg, c *Config) {
	// Set question to fks. IN TXT otherwise error
	// tsig signed, key = user
	// config stuff in Auth section (just as dynamic updates (*hint* *hint*)
	// SUBSYSTEM. IN TXT "OPERATION<SPACE>OPTIONS..."
	// ZONE. IN TXT "READ origin /z/bloep" - absolute path in fs

	if !req.IsUpdate() {
		logPrintf("non config command")
		formerr(w, req)
		return
	}

	// TODO: check tsig
	logPrintf("config commmand")
	for _, rr := range req.Ns {
		t, ok := rr.(*dns.RR_TXT)

		if !ok {
			// Not the TXT record -> error
			return
		}
		switch strings.ToUpper(t.Header().Name) {
		case "ZONE.":
			if e := configZONE(t, c); e != nil {
				formerr(w, req)
			}
		default:
			return
			// error back
		}
	}
}

// Deal with the zone options
func configZONE(t *dns.RR_TXT, c *Config) error {
	sx := strings.Split(t.Txt[0], " ")
	if len(sx) == 0 {
		return nil
	}
	switch strings.ToUpper(sx[0]) {
	case "READ":
		if len(sx) != 3 {
			return nil
		}
		logPrintf("config: READ %s %s\n", dns.Fqdn(sx[1]), sx[2])
		if e := c.ReadZoneFile(dns.Fqdn(sx[1]), sx[2]); e != nil {
			logPrintf("failed to read %s: %s\n", sx[2], e.Error())
			return e
		}
		logPrintf("config: added: READ %s %s\n", dns.Fqdn(sx[1]), sx[2])
	case "READXFR":
		if len(sx) != 3 {
			return nil
		}
		logPrintf("config: READXFR %s %s\n", dns.Fqdn(sx[1]), sx[2])
		if e := c.ReadZoneXfr(dns.Fqdn(sx[1]), sx[2]); e != nil {
			logPrintf("failed to axfr %s: %s\n", sx[2], e.Error())
			return e
		}
		logPrintf("config: added: READXFR %s %s\n", dns.Fqdn(sx[1]), sx[2])
	case "DROP":
		if len(sx) != 2 {
			return nil
		}
		logPrintf("config: DROP %s\n", dns.Fqdn(sx[1]))
		if e := c.DropZone(dns.Fqdn(sx[1])); e != nil {
			logPrintf("Failed to drop %s: %s\n", dns.Fqdn(sx[1]), e.Error())
			return e
		}
		logPrintf("config: dropped: DROP %s\n", dns.Fqdn(sx[1]))
	}
	return nil
}
