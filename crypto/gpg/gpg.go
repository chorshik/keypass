package gpg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	fileMode = 0600
	dirPerm  = 0700
)

var (
	reUIDComment = regexp.MustCompile(`([^(<]+)\s+(\([^)]+\))\s+<([^>]+)>`)
	reUID        = regexp.MustCompile(`([^(<]+)\s+<([^>]+)>`)
	// GPGArgs gpg argument
	GPGArgs = []string{"--quiet", "--yes", "--compress-algo=none", "--no-encrypt-to", "--no-auto-check-trustdb"}
	// GPGBin location of the gpg binary
	GPGBin = "gpg"
	// Debug ...
	Debug = false
)

// KeyList ...
type KeyList []Key

// Key ...
type Key struct {
	KeyType        string
	KeyLength      int
	Validity       string
	CreationDate   time.Time
	ExpirationDate time.Time
	Ownertrust     string
	Fingerprint    string
	Identities     map[string]Identity
	SubKeys        map[string]struct{}
}

// Identity ...
type Identity struct {
	Name           string
	Comment        string
	Email          string
	CreationDate   time.Time
	ExpirationDate time.Time
}

// ImportPublicKey ...
func ImportPublicKey(filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	args := append(GPGArgs, "--import")
	cmd := exec.Command(GPGBin, args...)
	if Debug {
		fmt.Printf("gpg.ImportPublicKey: %s %+v\n", cmd.Path, cmd.Args)
	}
	cmd.Stdin = bytes.NewReader(buf)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExportPublicKey ...
func ExportPublicKey(id, filename string) error {
	args := append(GPGArgs, "--armor", "--export", id)
	cmd := exec.Command(GPGBin, args...)
	if Debug {
		fmt.Printf("gpg.ExportPublicKey: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, out, fileMode)
}

// ListPublicKeys ...
func ListPublicKeys(search ...string) (KeyList, error) {
	return listKeys("public", search...)
}

// ListPrivateKeys ...
func ListPrivateKeys(search ...string) (KeyList, error) {
	return listKeys("secret", search...)
}

func listKeys(typ string, search ...string) (KeyList, error) {
	args := []string{"--with-colons", "--with-fingerprint", "--fixed-list-mode", "--list-" + typ + "-keys"}
	args = append(args, search...)
	cmd := exec.Command(GPGBin, args...)
	if Debug {
		fmt.Printf("gpg.listKeys: %s %+v\n", cmd.Path, cmd.Args)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(out, []byte("secret key not available")) {
			return KeyList{}, nil
		}
		return KeyList{}, err
	}

	return ParseColons(bytes.NewBuffer(out)), nil
}

// ParseColons ...
func ParseColons(reader io.Reader) KeyList {
	// http://git.gnupg.org/cgi-bin/gitweb.cgi?p=gnupg.git;a=blob_plain;f=doc/DETAILS
	// Fields:
	// 0 - Type of record
	//     Types:
	//     pub - Public Key
	//     crt - X.509 cert
	//     crs - X.509 cert and private key
	//     sub - Subkey (Secondary Key)
	//     sec - Secret / Private Key
	//     ssb - Secret Subkey
	//     uid - User ID
	//     uat - User attribute
	//     sig - Signature
	//     rev - Revocation Signature
	//     fpr - Fingerprint (field 9)
	//     pkd - Public Key Data
	//     grp - Keygrip
	//     rvk - Revocation KEy
	//     tfs - TOFU stats
	//     tru - Trust database info
	//     spk - Signature subpacket
	//     cfg - Configuration data
	// 1 - Validity
	// 2 - Key length
	// 3 - Public Key Algo
	// 4 - KeyID
	// 5 - Creation Date (UTC)
	// 6 - Expiration Date
	// 7 - Cert S/N
	// 8 - Ownertrust
	// 9 - User-ID
	// 10 - Sign. Class
	// 11 - Key Caps.
	// 12 - Issuer cert fp
	// 13 - Flag
	// 14 - S/N of a token
	// 15 - Hash algo (2 - SHA-1, 8 - SHA-256)
	// 16 - Curve Name
	
	kl := make(KeyList, 0, 100)

	scanner := bufio.NewScanner(reader)

	var cur Key

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Split(line, ":")

		switch fields[0] {
		case "pub":
			fallthrough
		case "sec":
			if cur.Fingerprint != "" && cur.KeyLength > 0 {
				kl = append(kl, cur)
			}
			validity := fields[1]
			if validity == "" && fields[0] == "sec" {
				validity = "u"
			}
			cur = Key{
				KeyType:        fields[0],
				Validity:       validity,
				KeyLength:      parseInt(fields[2]),
				CreationDate:   parseTS(fields[5]),
				ExpirationDate: parseTS(fields[6]),
				Ownertrust:     fields[8],
				Identities:     make(map[string]Identity, 1),
				SubKeys:        make(map[string]struct{}, 1),
			}
		case "sub":
			fallthrough
		case "ssb":
			cur.SubKeys[fields[4]] = struct{}{}
		case "fpr":
			if cur.Fingerprint == "" {
				cur.Fingerprint = fields[9]
			}
		case "uid":
			sn := fields[7]
			id := fields[9]
			ni := Identity{}
			if reUIDComment.MatchString(id) {
				if m := reUIDComment.FindStringSubmatch(id); len(m) > 3 {
					ni.Name = m[1]
					ni.Comment = strings.Trim(m[2], "()")
					ni.Email = m[3]
				}
			} else if reUID.MatchString(id) {
				if m := reUID.FindStringSubmatch(id); len(m) > 2 {
					ni.Name = m[1]
					ni.Email = m[2]
				}
			}
			cur.Identities[sn] = ni
		}
	}

	if cur.Fingerprint != "" && cur.KeyLength > 0 {
		kl = append(kl, cur)
	}

	return kl
}

// IsUseable ...
func (k Key) IsUseable() bool {
	if !k.ExpirationDate.IsZero() && k.ExpirationDate.Before(time.Now()) {
		return false
	}
	switch k.Validity {
	case "m":
		return true
	case "f":
		return true
	case "u":
		return true
	}
	return false
}

// UseableKeys ...
func (kl KeyList) UseableKeys() KeyList {
	nkl := make(KeyList, 0, len(kl))
	for _, k := range kl {
		if !k.IsUseable() {
			continue
		}
		nkl = append(nkl, k)
	}
	return nkl
}

// ID ...
func (i Identity) ID() string {
	out := i.Name
	if i.Comment != "" {
		out += " (" + i.Comment + ")"
	}
	out += " <" + i.Email + ">"
	return out
}

// OneLine ...
func (k Key) OneLine() string {
	id := Identity{}
	for _, i := range k.Identities {
		id = i
		break
	}
	return fmt.Sprintf("0x%s - %s", k.Fingerprint[24:], id.ID())
}

// Encrypt ...
func Encrypt(path string, content []byte, recipients []string, alwaysTrust bool) error {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return err
	}

	args := append(GPGArgs, "--encrypt", "--output", path)
	if alwaysTrust {
		// changing the trustmodel is possibly dangerous. A user should always
		// explicitly opt-in to do this
		args = append(args, "--trust-model=always")
	}
	for _, r := range recipients {
		args = append(args, "--recipient", r)
	}

	cmd := exec.Command(GPGBin, args...)
	if Debug {
		fmt.Printf("gpg.Encrypt: %s %+v\n", cmd.Path, cmd.Args)
	}
	cmd.Stdin = bytes.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// Decrypt ...
func Decrypt(path string) ([]byte, error) {
	args := append(GPGArgs, "--decrypt", path)
	cmd := exec.Command(GPGBin, args...)
	if Debug {
		fmt.Printf("gpg.Decrypt: %s %+v\n", cmd.Path, cmd.Args)
	}
	return cmd.Output()
}

func parseTS(str string) time.Time {
	t := time.Time{}

	if sec, err := strconv.ParseInt(str, 10, 64); err == nil {
		t = time.Unix(sec, 0)
	}

	return t
}

func parseInt(str string) int {
	i := 0

	if iv, err := strconv.ParseInt(str, 10, 32); err == nil {
		i = int(iv)
	}

	return i
}
