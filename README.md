# kernctl

Version 0.0.1

kernctl is a library for interacting with OSX kernel extensions via the
[KEXT Control API](https://developer.apple.com/library/mac/documentation/darwin/conceptual/NKEConceptual/control/control.html)

## Installing
`go get github.com/aschepis/kernctl`

## Example
```go
package main

import(
    "fmt"
    "github.com/aschepis/kernctl"
)

type KextMessage struct {
    Version uint32
    Command uint32
}

func (msg *KextMessage) Bytes() []byte {
    buf := new(bytes.Buffer)
    binary.Write(buf, binary.LittleEndian, msg)
    return buf.Bytes()
}

func main() {
    msg := &KextMessage{
        Version: 1,
        Command: 17,
    }

  conn := kernctl.NewConnByName("my.kext.name")
    conn.Connect()
    conn.SendCommand(msg)
    conn.Close()
}
```

## Sending Messages

Sending messages via the `SendCommand` function requires only that your argument implements
the `kernctl.Message` interface

## TODO

* bi-directional communication functionality so that a kext can send information
  back to a connected client.
* ability to use the KEXT Notification API to receive notifications from a KEXT
* some tests? This stuff is unfortunately difficult to test but it may be possible
  to refactor things enough that there are some testable components that can
  verify that the interaction with the kext is called properly.

## Contributing

By all means, please do! Pull requests and issues are welcome. No offical workflow yet.

## Versioning

kernctl uses semantic versioning. This is a bit pointless at the moment since the library
is yet to reach 1.0
