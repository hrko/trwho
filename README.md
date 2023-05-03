# trwho
A TUI application to show rwhod information saved as  `/var/spool/rwho/whod.*`.

## Build
```
$ git clone git@github.com:hrko/trwho.git
$ cd trwho
$ go build
```

## Usage
First, make sure that you can read `/var/spool/rwho/`.
```
$ sudo chmod 755 /var/spool/rwho
```

Then, just run `trwho`.
```
$ trwho
```
