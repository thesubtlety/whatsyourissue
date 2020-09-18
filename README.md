Utility to get the /etc/issue.net or /etc/motd from an SSH server after making a login attempt

### Usage

`./whatsyourissue -t 10.0.0.1/24 -n 100 -v -t 3`

```
$cat hosts | ./whatsyourissue
127.0.0.1               ########################
                        This is a test of the emergency broadcast system
                        ########################

127.0.0.2               *****************************************************

                                    This is a private system!!
                        All connections attempts will be logged and monitored.

                        *****************************************************
```

```
$ ./whatsyourissue -h                                                                                                                                                                          130 ↵
  -h    Print help
  -n int
        number of threads (default 100)
  -t string
        target ip or cidr range
  -timeout int
        ssh timeout in seconds (default 10)
  -v    print hosts without an issue/motd
```

