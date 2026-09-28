//go:build go1.23

package json_test

import (
	"crypto/x509"
	"testing"
)

func TestEncodeStdOID(t *testing.T) {
	oid, err := x509.ParseOID("1.2.840.113549.1.1.11")
	if err != nil {
		t.Fatal(err)
	}
	checkStdEncoding(t, oid, &oid, map[string]x509.OID{"sha256WithRSA": oid}, []x509.OID{oid})
}
