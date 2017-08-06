// Package  atlasscientific provides interfaces to the EZO stamps.
// This driver controls the pH sensor.

package phezo

import (
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/kidoman/embd"
)

const (
	statusDelay = 300 * time.Millisecond
	calDelay    = 900 * time.Microsecond
	readDelay   = 900 * time.Microsecond
	calMidCmd   = "Cal,mid,7.00"
	calLowCmd   = "Cal,low,4.00"
	calHighCmd  = "Cal,high,10.00"
	calClrCmd   = "Cal,clear"
	calQCmd     = "Cal,?"
	readCmd     = 'R'
	tempCmd     = 'T'
	infoCmd     = 'i'
	ledCmd      = 'L'
	sleepCmd    = "Sleep"
	slopeCmd    = "Slope"
	statCmd     = "Status"
	findCmd     = "Find"
	queryCmd    = ",?"
)

const (
	ioctlCalLow byte = iota
	ioctlCalMid
	ioctlCalHigh
	ioctlCalQ
	ioctlInfo
	ioctlRead
	ioctlTemp
	ioctlLed
	ioctlSleep
	ioctlSlope
	ioctlFind
	ioctlStatus
)

type PHEZO struct {
	Bus  embd.I2CBus
	addr byte
	quit chan bool
}

// Returns a handle to a pHEZO Sensor
func New(Bus embd.I2CBus, addr byte) *PHEZO {
	return &PHEZO{Bus: Bus, addr: addr}
}

func (d *PHEZO) ioctl(cmd byte, value []byte) error {
	go func() {
		var retval error
		switch {
		case cmd == ioctlCalLow:
			retval = d.sendCmd(calLowCmd, nil)
		case cmd == ioctlCalMid:
			retval = d.sendCmd(calMidCmd, nil)
		case cmd == ioctlCalHigh:
			retval = d.sendCmd(calHighCmd, nil)
		}
		return retval
	}() error

}

// Send command which has no return value (no data)
// error will be nil if response code is true (1)
func (d *PHEZO) sendCmd(cmd string, value []byte) error {
	if value != nil {
		cmd = append(cmd, value)
	}
	d.Bus.WriteBytes(d.addr, (byte(cmd)[:] + value))
	time.Sleep(calDelay)
	ret, err := d.Bus.ReadByte(d.addr)
	if (err != nil) || (ret < 1) {
		glog.Infof("phezo-sendCmd: error reading byte %v, %v", err, ret)
		return err
	}
	return nil
}

// Sends single byte command and waits for return string
func (d *PHEZO) readCmd(cmd byte, cmdLen byte) ([]byte, error) {
	d.Bus.WriteByte(d.addr, cmd)
	if cmd == ioctlRead {
		time.Sleep(readDelay)
	} else {
		time.Sleep(statusDelay)
	}
	var data [cmdLen]byte
	err := d.Bus.ReadFromReg(addr, cmd, data[:])
	if (err != nil) || (ret < 1) {
		glog.Infof("phezo-readCmd: error reading byte %v", err)
		return nil, err
	}
	// drop the decimal 1 and return value
	return data[1:], err
}

func (d *PHEZO) Read() float32 {
	go func() {
		v, e := d.readCmd(readCmd, nil)
		if e != nil {
			return nil
		}
		return strconv.ParseFloat(string(v), 16)
	}()
}

// Close.
func (d *PHEZO) Close() {
	if d.quit != nil {
		d.quit <- true
	}
	return
}
