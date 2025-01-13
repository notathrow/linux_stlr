package browsers

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
	keyring "github.com/m00nm00nn/go-dbus-keyring"
)

func (c *Chromium) GetMasterKey(unused string) error {
	// what is d-bus @https://dbus.freedesktop.org/
	// don't need chromium key file for Linux
	//defer os.Remove(types.ChromiumKey.TempFilename())

	conn, _ := dbus.SessionBus()

	svc, err := keyring.GetSecretService(conn)
	if err != nil {
		return err
	}
	session, err := svc.OpenSession()
	if err != nil {
		return err
	}
	defer func() {
		if err := session.Close(); err != nil {
			slog.Error("close dbus session error", "err", err.Error())
		}
	}()
	collections, err := svc.GetAllCollections()
	if err != nil {
		return err
	}
	var secret []byte
	for _, col := range collections {
		items, err := col.GetAllItems()
		if err != nil {
			return err
		}
		for _, i := range items {
			label, err := i.GetLabel()
			if err != nil {
				slog.Warn("get label from dbus", "err", err.Error())
				continue
			}
			if strings.Contains(label, " Safe Storage") {
				se, err := i.GetSecret(session.Path())
				if err != nil {
					return fmt.Errorf("get storage from dbus: %w", err)
				}
				secret = se.Value
			}
		}
	}

	if len(secret) == 0 {
		// set default secret @https://source.chromium.org/chromium/chromium/src/+/main:components/os_crypt/os_crypt_linux.cc;l=100
		secret = []byte("peanuts")
	}
	salt := []byte("saltysalt")
	// @https://source.chromium.org/chromium/chromium/src/+/master:components/os_crypt/os_crypt_linux.cc

	key := PBKDF2Key(secret, salt, 1, 16, sha1.New)
	c.MasterKey = key
	return nil
}

func (g *Gecko) GetMasterKey(path string) error {
	var globalSalt, metaBytes, nssA11, nssA102 []byte
	keyring.GetSecretService(nil)
	keyDB, err := GetDBConnection(filepath.Join(path, "key4.db"))

	if err != nil {
		return err
	}

	if err = keyDB.QueryRow(`SELECT item1, item2 FROM metaData WHERE id = 'password'`).Scan(&globalSalt, &metaBytes); err != nil {
		return err
	}

	if err = keyDB.QueryRow(`SELECT a11, a102 from nssPrivate`).Scan(&nssA11, &nssA102); err != nil {
		return err
	}

	metaPBE, err := NewASN1PBE(metaBytes)
	if err != nil {
		return err
	}

	k, err := metaPBE.Decrypt(globalSalt)
	if err != nil {
		return err
	}

	if !bytes.Contains(k, []byte("password-check")) {
		return errors.New("password check error")
	}

	if !bytes.Equal(nssA102, []byte{248, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}) {
		return errors.New("nssA102 error")
	}

	nssPBE, err := NewASN1PBE(nssA11)
	if err != nil {
		return err
	}

	finallyKey, err := nssPBE.Decrypt(globalSalt)
	if err != nil {
		return err
	}

	g.MasterKey = finallyKey[:24]
	return nil
}
