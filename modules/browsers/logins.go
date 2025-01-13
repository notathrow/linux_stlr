package browsers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func (c *Chromium) GetLogins(path string) (logins []Login, err error) {
	db, err := sql.Open("sqlite", filepath.Join(path, "Login Data"))
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT origin_url, username_value, password_value, date_created FROM logins")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			url, username string
			pwd, password []byte
			create        int64
		)
		if err := rows.Scan(&url, &username, &pwd, &create); err != nil {
			continue
		}
		if url == "" || username == "" || pwd == nil {
			continue
		}

		login := Login{
			Username: username,
			LoginURL: url,
		}
		if len(pwd) > 0 {
			if len(c.MasterKey) == 0 {
				password, err = DecryptWithDPAPI(pwd)
			} else {
				password, err = DecryptWithChromium(c.MasterKey, pwd)
			}
			if err != nil {
				continue
			}
		}
		login.Password = string(password)
		logins = append(logins, login)
	}

	return logins, nil
}

func (g *Gecko) GetLogins(path string) (logins []Login, err error) {
	s, err := os.ReadFile(filepath.Join(path, "logins.json"))
	if err != nil {
		return nil, err
	}

	var data struct {
		NextId int `json:"nextId"`
		Logins []struct {
			Hostname          string `json:"hostname"`
			EncryptedUsername string `json:"encryptedUsername"`
			EncryptedPassword string `json:"encryptedPassword"`
		}
	}
	err = json.Unmarshal(s, &data)
	if err != nil {
		return nil, err
	}

	for _, v := range data.Logins {
		decodedUser, err := base64.StdEncoding.DecodeString(v.EncryptedUsername)
		if err != nil {
			return nil, err
		}
		decodedPass, err := base64.StdEncoding.DecodeString(v.EncryptedPassword)
		if err != nil {
			return nil, err
		}

		userPBE, err := NewASN1PBE(decodedUser)
		if err != nil {
			return nil, err
		}
		pwdPBE, err := NewASN1PBE(decodedPass)
		if err != nil {
			return nil, err
		}
		user, err := userPBE.Decrypt(g.MasterKey)
		if err != nil {
			return nil, err
		}
		pwd, err := pwdPBE.Decrypt(g.MasterKey)
		if err != nil {
			return nil, err
		}
		logins = append(logins, Login{
			Username: string(user),
			Password: string(pwd),
			LoginURL: v.Hostname,
		})
	}
	return logins, nil
}
