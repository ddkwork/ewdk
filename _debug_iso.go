package main
import (
    fmt
    syscall
    unsafe
    golang.org/x/sys/windows
    ewdkvhd ewdk/vhd
    github.com/Microsoft/go-winio/pkg/guid
)
func main() {
    isoPath := d:\ewdk\EWDK_br_release_28000_251103-1709.iso
    vendorMS := guid.GUID{Data1: 0xec984aec, Data2: 0xa0f9, Data3: 0x47e9, Data4: [8]byte{0x90, 0x1f, 0x71, 0x41, 0x5a, 0x66, 0x34, 0x5b}}
    isoType := ewdkvhd.VirtualStorageType{DeviceID: 1, VendorID: vendorMS}
    type pV1 struct { v uint32; d uint32 }
    params := pV1{v: 1, d: 1}
    var h syscall.Handle
    err := ewdkvhd.OpenVirtualDiskRaw(&isoType, isoPath, 0x00020000, 0, (*ewdkvhd.OpenVirtualDiskParametersRaw)(unsafe.Pointer(&params)), &h)
    fmt.Printf(err=%v h=%d, err, h)
    if err != nil { fmt.Printf( errno=%d, err.(syscall.Errno)) }
}
