package browsers

import (
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite"
)

func (c *Chromium) GetCookies(path string) (cookies []Cookie, err error) {
	var dbpath string
	if runtime.GOOS == "windows" {
		dbpath = filepath.Join(path, "Network", "Cookies")
	} else if runtime.GOOS == "linux" {
		dbpath = filepath.Join(path, "Cookies")
	}

	db, err := GetDBConnection(dbpath)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT name, encrypted_value, host_key, path, expires_utc FROM cookies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name, host, path       string
			encrypted_value, value []byte
			expires_utc            int64
		)
		if err = rows.Scan(&name, &encrypted_value, &host, &path, &expires_utc); err != nil {
			continue
		}

		if name == "" || host == "" || path == "" || encrypted_value == nil {
			continue
		}

		cookie := Cookie{
			Name:       name,
			Host:       host,
			Path:       path,
			ExpireDate: expires_utc,
		}
		value, err = DecryptWithChromium(c.MasterKey, encrypted_value)
		if err != nil {
			continue
		}
		cookie.Value = string(value)
		cookies = append(cookies, cookie)
	}

	return cookies, nil
}

func (g *Gecko) GetCookies(path string) (cookies []Cookie, err error) {
	db, err := GetDBConnection(filepath.Join(path, "cookies.sqlite"))
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT name, value, host, path, expiry FROM moz_cookies")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			name, host, path string
			value            []byte
			expiry           int64
		)
		if err = rows.Scan(&name, &value, &host, &path, &expiry); err != nil {
			continue
		}

		if name == "" || host == "" || path == "" || value == nil {
			continue
		}

		cookie := Cookie{
			Name:       name,
			Host:       host,
			Path:       path,
			ExpireDate: expiry,
			Value:      string(value),
		}
		cookies = append(cookies, cookie)
	}

	return cookies, nil
}
