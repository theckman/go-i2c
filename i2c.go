// Package i2c provides low level control over the Linux i2c bus.
//
// Before usage you should load the i2c-dev kernel module
//
//      sudo modprobe i2c-dev
//
// Each i2c bus can address 127 independent i2c devices, and most
// Linux systems contain several buses.
package i2c

import (
	"encoding/hex"
	"fmt"
	"os"
	"syscall"
)

// NOOPDebugf is a no-op formatted debug function. Exported so that you can use
// it to toggle debug logging back off using I2C.SetDebugf().
func NOOPDebugf(string, ...interface{}) {}

// I2C represents a connection to I2C-device.
type I2C struct {
	addr   uint8
	bus    int
	rc     *os.File
	debugf func(string, ...interface{})
}

// NewI2C opens a connection for I2C-device.
// SMBus (System Management Bus) protocol over I2C
// supported as well: you should preliminary specify
// register address to read from, either write register
// together with the data in case of write operations.
func NewI2C(addr uint8, bus int) (*I2C, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", bus), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	if err := ioctl(f.Fd(), i2cSlave, uintptr(addr)); err != nil {
		return nil, err
	}

	i := &I2C{
		rc:     f,
		bus:    bus,
		addr:   addr,
		debugf: NOOPDebugf,
	}

	return i, nil
}

// SetDebugf sets a formatted debug function, which can be used to hook in to
// your logging system.
func (i *I2C) SetDebugf(debugf func(format string, args ...interface{})) {
	i.debugf = debugf
}

// GetBus return bus line, where I2C-device is allocated.
func (i *I2C) GetBus() int {
	return i.bus
}

// GetAddr return device occupied address in the bus.
func (i *I2C) GetAddr() uint8 {
	return i.addr
}

// Write satisfies io.Writer, sending data to the I2C-device.
func (i *I2C) Write(buf []byte) (int, error) {
	i.debugf("Write %d hex bytes: [%+v]", len(buf), hex.EncodeToString(buf))

	return i.rc.Write(buf)
}

// WriteByte writes a single byte to the I2C device.
func (i *I2C) WriteByte(b byte) (int, error) {
	buf := [1]byte{b}

	return i.Write(buf[:])
}

// Read satisfies io.Reader, reading data from the I2C-device.
func (i *I2C) Read(buf []byte) (int, error) {
	n, err := i.rc.Read(buf)
	if err != nil {
		return n, err
	}

	i.debugf("Read %d hex bytes: [%+v]", len(buf), hex.EncodeToString(buf))
	return n, nil
}

// Close I2C-connection.
func (i *I2C) Close() error {
	return i.rc.Close()
}

// ReadRegBytes read count of n byte's sequence from I2C-device
// starting from reg address.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) ReadRegBytes(reg byte, n int) ([]byte, int, error) {
	i.debugf("Read %d bytes starting from reg 0x%0X...", n, reg)

	_, err := i.WriteByte(reg)
	if err != nil {
		return nil, 0, err
	}

	buf := make([]byte, n)

	c, err := i.Read(buf)
	if err != nil {
		return nil, 0, err
	}

	return buf, c, nil
}

// ReadRegU8 reads byte from I2C-device register specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) ReadRegU8(reg byte) (byte, error) {
	_, err := i.WriteByte(reg)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 1)

	_, err = i.Read(buf)
	if err != nil {
		return 0, err
	}

	i.debugf("Read U8 %d from reg 0x%0X", buf[0], reg)
	return buf[0], nil
}

// WriteRegU8 writes byte to I2C-device register specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) WriteRegU8(reg byte, value byte) error {
	buf := []byte{reg, value}

	_, err := i.Write(buf)
	if err != nil {
		return err
	}

	i.debugf("Write U8 %d to reg 0x%0X", value, reg)
	return nil
}

// ReadRegU16BE reads unsigned big endian word (16 bits)
// from I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) ReadRegU16BE(reg byte) (uint16, error) {
	_, err := i.WriteByte(reg)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 2)

	_, err = i.Read(buf)
	if err != nil {
		return 0, err
	}

	w := uint16(buf[0])<<8 + uint16(buf[1])

	i.debugf("Read U16 %d from reg 0x%0X", w, reg)
	return w, nil
}

// ReadRegU16LE reads unsigned little endian word (16 bits)
// from I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) ReadRegU16LE(reg byte) (uint16, error) {
	w, err := i.ReadRegU16BE(reg)
	if err != nil {
		return 0, err
	}

	// exchange bytes
	w = (w&0xFF)<<8 + w>>8

	return w, nil
}

// ReadRegS16BE reads signed big endian word (16 bits)
// from I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) ReadRegS16BE(reg byte) (int16, error) {
	_, err := i.WriteByte(reg)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 2)

	_, err = i.Read(buf)
	if err != nil {
		return 0, err
	}

	w := int16(buf[0])<<8 + int16(buf[1])

	i.debugf("Read S16 %d from reg 0x%0X", w, reg)
	return w, nil
}

// ReadRegS16LE reads signed little endian word (16 bits)
// from I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) ReadRegS16LE(reg byte) (int16, error) {
	w, err := i.ReadRegS16BE(reg)
	if err != nil {
		return 0, err
	}

	// exchange bytes
	w = (w&0xFF)<<8 + w>>8

	return w, nil
}

// WriteRegU16BE writes unsigned big endian word (16 bits)
// value to I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) WriteRegU16BE(reg byte, value uint16) error {
	buf := []byte{reg, byte((value & 0xFF00) >> 8), byte(value & 0xFF)}

	_, err := i.Write(buf)
	if err != nil {
		return err
	}

	i.debugf("Write U16 %d to reg 0x%0X", value, reg)
	return nil
}

// WriteRegU16LE writes unsigned little endian word (16 bits)
// value to I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) WriteRegU16LE(reg byte, value uint16) error {
	w := (value*0xFF00)>>8 + value<<8

	return i.WriteRegU16BE(reg, w)
}

// WriteRegS16BE writes signed big endian word (16 bits)
// value to I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) WriteRegS16BE(reg byte, value int16) error {
	buf := []byte{reg, byte((uint16(value) & 0xFF00) >> 8), byte(value & 0xFF)}

	_, err := i.Write(buf)
	if err != nil {
		return err
	}

	i.debugf("Write S16 %d to reg 0x%0X", value, reg)
	return nil
}

// WriteRegS16LE writes signed little endian word (16 bits)
// value to I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (i *I2C) WriteRegS16LE(reg byte, value int16) error {
	w := int16((uint16(value)*0xFF00)>>8) + value<<8

	return i.WriteRegS16BE(reg, w)
}

func ioctl(fd, cmd, arg uintptr) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if err != 0 {
		return err
	}

	return nil
}
