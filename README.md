# i2c
[![Build Status](https://travis-ci.org/theckman/i2c.svg?branch=master)](https://travis-ci.org/theckman/i2c)
[![Go Report Card](https://goreportcard.com/badge/github.com/theckman/i2c)](https://goreportcard.com/report/github.com/theckman/i2c)
[![GoDoc](https://godoc.org/github.com/theckman/i2c?status.svg)](https://godoc.org/github.com/theckman/i2c)
[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

This project is a set of Go bindings for the I²C bus, focused on sensors compatible with Raspberry Pi (and RPi clones). This project was forked from [github.com/d2r2/go-i2c](https://github.com/d2r2/go-i2c), which itself was forked from [github.com/davecheney/i2c](https://github.com/davecheney/i2c).

## License

i2c is licensed under MIT License.

## Compatibility

Pre-fork this project was tested on Raspberry Pi 1 (Model B), Raspberry Pi 3
(Model B+), Banana Pi (model M1), Orange Pi Zero, Orange Pi One.

## Usage

This section is a work in progress. Ideally the
[GoDoc](http://godoc.org/github.com/theckman/i2c) page provides enough context
to get started.

## Troubleshooting

- *How to enable I2C bus on RPi device:*
If you employ RaspberryPI, use raspi-config utility to activate i2c-bus on the OS level.
Go to "Interfacing Options" menu, to active I2C bus.
Probably you will need to reboot to load i2c kernel module.
Finally you should have device like /dev/i2c-1 present in the system.

- *How to find I2C bus allocation and device address:*
Use i2cdetect utility in format "i2cdetect -y X", where X may vary from 0 to 5 or more,
to discover address occupied by peripheral device. To install utility you should run
`apt install i2c-tools` on debian-kind system. `i2cdetect -y 1` sample output:
	```
	     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
	00:          -- -- -- -- -- -- -- -- -- -- -- -- --
	10: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	30: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	40: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	70: -- -- -- -- -- -- 76 --    
	```
