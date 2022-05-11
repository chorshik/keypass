package action

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

// clearClipboard ...
func clearClipboard(content []byte, timeout int) error {
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	cmd := exec.Command(os.Args[0], "unclip", "--timeout", strconv.Itoa(timeout))
	// https://groups.google.com/d/msg/golang-nuts/shST-SDqIp4/za4oxEiVtI0J
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Env = append(os.Environ(), "KEYPASS_UNCLIP_CHECKSUM="+hash)
	return cmd.Start()
}
