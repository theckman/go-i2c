// Package i2c provides low level interactions with the Linux I²C bus.
//
// Before usage you should load the i2c-dev kernel module
//
//      sudo modprobe i2c-dev
//
// Each I²C bus can address 127 independent I²C devices, and most
// Linux systems contain several buses.
package i2c

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"syscall"
)

// NOOPDebugf is a no-op formatted debug function. Exported so that you can use
// it to toggle debug logging back off using Device.SetDebugf().
func NOOPDebugf(string, ...interface{}) {}

// Device is a connection to a device on the I²C bus. It contains a file handle
// to a specific device address on a numbered I²C bus.
type Device struct {
	addr   uint8
	bus    int
	rc     *os.File
	debugf func(string, ...interface{})
}

// New opens a new file handle on the provided I²C bus, making an ioctl call
// to request read/write access to the device at the specified address.
//
// Most interactions start with either reads or writes at a specific register
// address. See ReadReg and WriteReg.
func New(bus int, addr uint8) (*Device, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", bus), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	if err := ioctl(f.Fd(), i2cSlave, uintptr(addr)); err != nil {
		return nil, err
	}

	i := &Device{
		rc:     f,
		bus:    bus,
		addr:   addr,
		debugf: NOOPDebugf,
	}

	return i, nil
}

// SetDebugf sets a formatted debug function, which can be used to hook in to
// your logging system.
func (d *Device) SetDebugf(debugf func(format string, args ...interface{})) {
	d.debugf = debugf
}

// Bus return bus number to create this device.
func (d *Device) Bus() int {
	return d.bus
}

// Addr returns the device's address on the I²C bus.
func (d *Device) Addr() uint8 {
	return d.addr
}

// Write satisfies io.Writer, sending data to the I2C device.
func (d *Device) Write(p []byte) (int, error) {
	n := len(p)
	if n > 512 {
		return 0, fmt.Errorf("maximum message length 512, was %d", n)
	}

	if n == 0 {
		return 0, errors.New("minimum message length 1")
	}

	d.debugf("Write %d bytes: [%+v]", n, hex.EncodeToString(p))

	return d.rc.Write(p)
}

// WriteByte writes a single byte to the I2C device.
func (d *Device) WriteByte(b byte) (int, error) {
	buf := [1]byte{b}

	return d.Write(buf[:])
}

// WriteReg writes a series of bytes to a specific register address.
func (d *Device) WriteReg(p []byte, reg byte) (int, error) {
	n := len(p)
	if n > 511 {
		return 0, fmt.Errorf("maximum message length 511, was %d", n)
	}

	if n == 0 {
		return 0, errors.New("minimum message length 1")
	}

	buf := make([]byte, n+1)
	buf[0] = reg

	copy(buf[1:], p)

	d.debugf("Write %d bytes: [%+v]", len(buf), hex.EncodeToString(buf))

	return d.Write(buf)
}

// Read satisfies io.Reader, reading data from the I2C device.
func (d *Device) Read(p []byte) (int, error) {
	n, err := d.rc.Read(p)
	if err != nil {
		return n, err
	}

	d.debugf("Read %d bytes: [%+v]", n, hex.EncodeToString(p[:n]))
	return n, nil
}

// ReadReg reads I2C device data at the specified register address into the
// buffer provided. This expects you to right-size the buffer so that it only
// reads the appropriate amount of data.
func (d *Device) ReadReg(p []byte, reg byte) (int, error) {
	d.debugf("Reading %d bytes from register 0x%0X", len(p), reg)

	_, err := d.WriteByte(reg)
	if err != nil {
		return 0, err
	}

	n, err := d.Read(p)
	if err != nil {
		return n, err
	}

	return n, nil
}

// Close I²C file handle.
func (d *Device) Close() error {
	err := d.rc.Close()
	d.bus = 0
	d.addr = 0

	return err
}

// ReadRegU8 reads byte from I2C device register specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) ReadRegU8(reg byte) (byte, error) {
	_, err := d.WriteByte(reg)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 1)

	_, err = d.Read(buf)
	if err != nil {
		return 0, err
	}

	d.debugf("Read U8 %d from reg 0x%0X", buf[0], reg)
	return buf[0], nil
}

// WriteRegU8 writes byte to I2C device register specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) WriteRegU8(reg byte, value byte) error {
	buf := [2]byte{reg, value}

	_, err := d.Write(buf[:])
	if err != nil {
		return err
	}

	d.debugf("Write U8 %d to reg 0x%0X", value, reg)
	return nil
}

// ReadRegU16BE reads unsigned big endian word (16 bits)
// from I2C device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) ReadRegU16BE(reg byte) (uint16, error) {
	_, err := d.WriteByte(reg)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 2)

	_, err = d.Read(buf)
	if err != nil {
		return 0, err
	}

	w := uint16(buf[0])<<8 + uint16(buf[1])

	d.debugf("Read U16 %d from reg 0x%0X", w, reg)
	return w, nil
}

// ReadRegU16LE reads unsigned little endian word (16 bits)
// from I2C device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) ReadRegU16LE(reg byte) (uint16, error) {
	w, err := d.ReadRegU16BE(reg)
	if err != nil {
		return 0, err
	}

	// exchange bytes
	w = (w&0xFF)<<8 + w>>8

	return w, nil
}

// ReadRegS16BE reads signed big endian word (16 bits)
// from I2C device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) ReadRegS16BE(reg byte) (int16, error) {
	_, err := d.WriteByte(reg)
	if err != nil {
		return 0, err
	}

	buf := make([]byte, 2)

	_, err = d.Read(buf)
	if err != nil {
		return 0, err
	}

	w := int16(buf[0])<<8 + int16(buf[1])

	d.debugf("Read S16 %d from reg 0x%0X", w, reg)
	return w, nil
}

// ReadRegS16LE reads signed little endian word (16 bits)
// from I2C device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) ReadRegS16LE(reg byte) (int16, error) {
	w, err := d.ReadRegS16BE(reg)
	if err != nil {
		return 0, err
	}

	// exchange bytes
	w = (w&0xFF)<<8 + w>>8

	return w, nil
}

// WriteRegU16BE writes unsigned big endian word (16 bits)
// value to I2C device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) WriteRegU16BE(reg byte, value uint16) error {
	buf := [3]byte{reg, byte((value & 0xFF00) >> 8), byte(value & 0xFF)}

	_, err := d.Write(buf[:])
	if err != nil {
		return err
	}

	d.debugf("Write U16 %d to reg 0x%0X", value, reg)
	return nil
}

// WriteRegU16LE writes unsigned little endian word (16 bits)
// value to I2C device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) WriteRegU16LE(reg byte, value uint16) error {
	w := (value*0xFF00)>>8 + value<<8

	return d.WriteRegU16BE(reg, w)
}

// WriteRegS16BE writes signed big endian word (16 bits)
// value to I2C device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) WriteRegS16BE(reg byte, value int16) error {
	buf := [3]byte{reg, byte((uint16(value) & 0xFF00) >> 8), byte(value & 0xFF)}

	_, err := d.Write(buf[:])
	if err != nil {
		return err
	}

	d.debugf("Write S16 %d to reg 0x%0X", value, reg)
	return nil
}

// WriteRegS16LE writes signed little endian word (16 bits)
// value to I2C device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (d *Device) WriteRegS16LE(reg byte, value int16) error {
	w := int16((uint16(value)*0xFF00)>>8) + value<<8

	return d.WriteRegS16BE(reg, w)
}

func ioctl(fd, cmd, arg uintptr) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if err != 0 {
		return err
	}

	return nil
}
