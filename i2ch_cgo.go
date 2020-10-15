// +build linux,cgo

package i2c

// #include <linux/i2c-dev.h>
import "C"

// Get I2C_SLAVE constant value from
// Linux OS I2C declaration file.
const (
	i2cSlave = C.I2C_SLAVE
)
