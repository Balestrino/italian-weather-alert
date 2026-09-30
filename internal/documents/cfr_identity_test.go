package documents

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"
)

func TestCFRPDFIdentityIgnoresOnlyGenerationMetadata(t *testing.T) {
	first := []byte("%PDF-1.7\n/Producer (TCPDF 6.6.5) /CreationDate (D:20260925105921+01'00')\n" +
		"<xmp:CreateDate>2026-09-25T10:59:21+01:00</xmp:CreateDate>\n" +
		"<xmpMM:DocumentID>uuid:234e7d91-5fa6-e475-8a97-a5d04df621a0</xmpMM:DocumentID>\n" +
		"/ID [ <234e7d915fa6e4758a97a5d04df621a0> <234e7d915fa6e4758a97a5d04df621a0> ]\n" +
		"stream\nlevel: green\nendstream")
	original := bytes.Clone(first)
	second := bytes.ReplaceAll(first, []byte("20260925105921"), []byte("20260925115925"))
	second = bytes.ReplaceAll(second, []byte("2026-09-25T10:59:21"), []byte("2026-09-25T11:59:25"))
	second = bytes.ReplaceAll(second, []byte("234e7d91-5fa6-e475-8a97-a5d04df621a0"), []byte("5dab9fe9-b1ee-ac3c-4d6c-316593ee9e1b"))
	second = bytes.ReplaceAll(second, []byte("234e7d915fa6e4758a97a5d04df621a0"), []byte("5dab9fe9b1eeac3c4d6c316593ee9e1b"))
	if !bytes.Equal(cfrContentIdentity(first, "application/pdf"), cfrContentIdentity(second, "application/pdf")) {
		t.Fatal("generation-only changes must have the same identity")
	}
	changed := bytes.ReplaceAll(second, []byte("level: green"), []byte("level: orange"))
	if bytes.Equal(cfrContentIdentity(first, "application/pdf"), cfrContentIdentity(changed, "application/pdf")) {
		t.Fatal("changed publication content must have a new identity")
	}
	if !bytes.Equal(first, original) {
		t.Fatal("identity calculation must not modify original bytes")
	}
}

func TestCFRPNGIdentityIgnoresOnlyGenerationChunks(t *testing.T) {
	first := append([]byte("\x89PNG\r\n\x1a\n"), pngChunk("IHDR", []byte("dimensions"))...)
	first = append(first, pngChunk("tIME", []byte{7, 234, 9, 25, 10, 59, 21})...)
	first = append(first, pngChunk("tEXt", []byte("date:modify\x002026-09-25T10:59:21+02:00"))...)
	first = append(first, pngChunk("IDAT", []byte("pixels"))...)
	first = append(first, pngChunk("IEND", nil)...)
	original := bytes.Clone(first)
	second := append([]byte("\x89PNG\r\n\x1a\n"), pngChunk("IHDR", []byte("dimensions"))...)
	second = append(second, pngChunk("tIME", []byte{7, 234, 9, 25, 11, 59, 25})...)
	second = append(second, pngChunk("tEXt", []byte("date:modify\x002026-09-25T11:59:25+02:00"))...)
	second = append(second, pngChunk("IDAT", []byte("pixels"))...)
	second = append(second, pngChunk("IEND", nil)...)
	if !bytes.Equal(cfrContentIdentity(first, "image/png"), cfrContentIdentity(second, "image/png")) {
		t.Fatal("generation-only changes must have the same identity")
	}
	changed := bytes.ReplaceAll(second, []byte("pixels"), []byte("P1xels"))
	if bytes.Equal(cfrContentIdentity(first, "image/png"), cfrContentIdentity(changed, "image/png")) {
		t.Fatal("changed map pixels must have a new identity")
	}
	if !bytes.Equal(first, original) {
		t.Fatal("identity calculation must not modify original bytes")
	}
}

func TestCFRVersionIdentityKeepsOriginalObjectHashes(t *testing.T) {
	makeAcquisition := func(date string) Acquisition {
		pdf := []byte("%PDF-1.7\n/Producer (TCPDF 6.6.5) /CreationDate (D:" + date + "+01'00')\nstream\nwarning\nendstream")
		return Acquisition{
			ID: "cfr-" + date, SourceID: "cfr-criticality", Configuration: 1, URL: "https://cfr.toscana.it/product", Metadata: []byte(`{}`),
			Resources: []Resource{
				{URL: "https://cfr.toscana.it/product", Role: "original", Required: true, SourceID: "cfr-criticality", Configuration: 1, MediaType: "text/html", Bytes: []byte("<h1>warning</h1>")},
				{URL: "https://cfr.toscana.it/print.pdf", Role: "resource", Required: true, SourceID: "cfr-criticality", Configuration: 1, MediaType: "application/pdf", Bytes: pdf},
			},
		}
	}
	first, firstRefs, firstHash, _, err := prepare(makeAcquisition("20260925105921"))
	if err != nil {
		t.Fatal(err)
	}
	second, secondRefs, secondHash, _, err := prepare(makeAcquisition("20260925115925"))
	if err != nil {
		t.Fatal(err)
	}
	if firstHash != secondHash {
		t.Fatal("generation metadata created a new version")
	}
	if firstRefs[0].Hash == secondRefs[0].Hash || firstRefs[0].Hash != digest(first.Resources[0].Bytes) || secondRefs[0].Hash != digest(second.Resources[0].Bytes) {
		t.Fatal("original print bytes must retain distinct object hashes")
	}
	changed := makeAcquisition("20260925115925")
	changed.Resources[1].Bytes = bytes.ReplaceAll(changed.Resources[1].Bytes, []byte("warning"), []byte("critical"))
	_, _, changedHash, _, err := prepare(changed)
	if err != nil || changedHash == firstHash {
		t.Fatal("changed print content must create a new version", err)
	}
}

func pngChunk(kind string, data []byte) []byte {
	chunk := make([]byte, 12+len(data))
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(data)))
	copy(chunk[4:8], kind)
	copy(chunk[8:], data)
	binary.BigEndian.PutUint32(chunk[8+len(data):], crc32.ChecksumIEEE(chunk[4:8+len(data)]))
	return chunk
}
