package mutex

import (
	"encoding/hex"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
)

func mutexpath(mutexseed []byte) string {
	mutexValue := hex.EncodeToString(mutexseed)
	if len(mutexValue) < 32 {
		for len(mutexValue) < 32 {
			gen := rand.New(rand.NewSource(1337))
			mutexValue += strconv.Itoa(gen.Intn(9))
		}
	}
	mutexValue = mutexValue[len(mutexValue)-32:]
	tmp := []byte(mutexValue)
	tmp[0] = 'S'
	mutexValue = string(tmp)
	return os.TempDir() + "/systemd-private-" + mutexValue
}

func Createmutx() error {
	out, err := exec.Command("who", "-b").Output()
	if err != nil {
		return err
	}
	err = os.Mkdir(mutexpath(out), 0700)
	return err
}

// true is mutex is valid and exit, false is otherwise
func Checkmutx() bool {
	out, err := exec.Command("who", "-b").Output()
	if err != nil {
		return false
	}
	_, err = os.ReadDir(mutexpath(out))
	return !os.IsNotExist(err)
}

func Removemutex() {
	out, _ := exec.Command("who", "-b").Output()
	os.Remove(mutexpath(out))
}

func Cleanexit(code int) {
	Removemutex()
	os.Exit(code)
}
