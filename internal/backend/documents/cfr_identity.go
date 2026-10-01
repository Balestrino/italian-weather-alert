package documents

import (
	"bytes"
	"encoding/binary"
	"regexp"
	"strings"
	"time"
)

var (
	cfrPDFDates = regexp.MustCompile(`/(LastModified|M|CreationDate|ModDate) \(D:[0-9]{14}[+-][0-9]{2}'[0-9]{2}'\)`)
	cfrXMPDates = regexp.MustCompile(`<xmp:(CreateDate|ModifyDate|MetadataDate)>[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[+-][0-9]{2}:[0-9]{2}</xmp:(CreateDate|ModifyDate|MetadataDate)>`)
	cfrXMPIDs   = regexp.MustCompile(`<xmpMM:(DocumentID|InstanceID)>uuid:[0-9a-fA-F-]{36}</xmpMM:(DocumentID|InstanceID)>`)
	cfrPDFIDs   = regexp.MustCompile(`/ID \[ <[0-9a-fA-F]{32}> <[0-9a-fA-F]{32}> \]`)
)

// cfrContentIdentity removes only generation metadata observed in the CFR
// product's TCPDF print output and PNG map exports. It never changes retained
// bytes, and unrecognized formats fall back to byte-exact identity.
func cfrContentIdentity(raw []byte, mediaType string) []byte {
	switch strings.ToLower(strings.TrimSpace(strings.Split(mediaType, ";")[0])) {
	case "application/pdf":
		if !bytes.HasPrefix(raw, []byte("%PDF-")) || !bytes.Contains(raw, []byte("/Producer (TCPDF ")) {
			return raw
		}
		canonical := cfrPDFDates.ReplaceAll(raw, []byte("/$1 (D:generated)"))
		canonical = cfrXMPDates.ReplaceAll(canonical, []byte("<xmp:$1>generated</xmp:$2>"))
		canonical = cfrXMPIDs.ReplaceAll(canonical, []byte("<xmpMM:$1>uuid:generated</xmpMM:$2>"))
		return cfrPDFIDs.ReplaceAll(canonical, []byte("/ID [ <generated> <generated> ]"))
	case "image/png":
		return cfrPNGContentIdentity(raw)
	default:
		return raw
	}
}

func cfrPNGContentIdentity(raw []byte) []byte {
	if !bytes.HasPrefix(raw, []byte("\x89PNG\r\n\x1a\n")) {
		return raw
	}
	canonical := make([]byte, 0, len(raw))
	canonical = append(canonical, raw[:8]...)
	for offset := 8; offset < len(raw); {
		if len(raw)-offset < 12 {
			return raw
		}
		length := uint64(binary.BigEndian.Uint32(raw[offset : offset+4]))
		if length > uint64(len(raw)-offset-12) {
			return raw
		}
		end := offset + 12 + int(length)
		kind := raw[offset+4 : offset+8]
		data := raw[offset+8 : end-4]
		drop := (bytes.Equal(kind, []byte("tIME")) && len(data) == 7) ||
			(bytes.Equal(kind, []byte("tEXt")) && bytes.HasPrefix(data, []byte("date:modify\x00")) &&
				validCFRPNGDate(data[len("date:modify\x00"):]))
		if !drop {
			canonical = append(canonical, raw[offset:end]...)
		}
		offset = end
		if bytes.Equal(kind, []byte("IEND")) {
			if offset != len(raw) {
				return raw
			}
			return canonical
		}
	}
	return raw
}

func validCFRPNGDate(value []byte) bool {
	_, err := time.Parse(time.RFC3339, string(value))
	return err == nil
}
