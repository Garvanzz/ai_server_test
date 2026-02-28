package id

// Package snowflake provides a very simple Twitter snowflake generator and parser.
// +-------------------------------------------------------+
// | 40 Bit Timestamp | 10 Bit WorkID | 14 Bit Sequence ID |
// +-------------------------------------------------------+

import (
	"errors"
	"sync"
	"time"
	"xfx/pkg/log"
)

const (
	numMachineBits  = 10
	numSequenceBits = 14
	MaxMachineId    = -1 ^ (-1 << numMachineBits)  //		...0000001111111111 1023
	MaxSequence     = -1 ^ (-1 << numSequenceBits) //		...0011111111111111 16383
)

type SnowFlake struct {
	time20200101Ms int64
	lastTimestamp  int64
	sequence       uint32
	machineId      uint32
	lock           sync.Mutex
}

var idCreater *SnowFlake

// Init machineId: 0~1000为保留段
// 其中100~200用于给机器人生成唯一id
func Init(machineId uint32) {
	var err error
	if idCreater, err = New(machineId); err != nil {
		log.Error("id: init error(%v)", err)
		panic(err)
	}
}

func General() (int64, error) {
	return idCreater.Generate()
}

func (sf *SnowFlake) pack() int64 {
	uuid := (sf.lastTimestamp << (numMachineBits + numSequenceBits)) | (int64(sf.machineId) << numSequenceBits) | (int64(sf.sequence))
	return uuid
}

func New(machineId uint32) (*SnowFlake, error) {
	// if machineId < 0 || machineId > MaxMachineId {
	if machineId > MaxMachineId {
		return nil, errors.New("id: invalid worker Id")
	}
	time20200101, _ := time.Parse("2006-01-02 15:04:05", "2020-01-01 11:11:11")
	time20200101Ms := time20200101.UnixNano() / int64(time.Millisecond)
	return &SnowFlake{
		machineId:      machineId,
		time20200101Ms: time20200101Ms,
	}, nil
}

func (sf *SnowFlake) Generate() (int64, error) {
	sf.lock.Lock()
	defer sf.lock.Unlock()

	ts := sf.timestamp()
	if ts == sf.lastTimestamp {
		sf.sequence = (sf.sequence + 1) & MaxSequence
		if sf.sequence == 0 {
			ts = sf.waitNextMilli(ts)
		}
	} else {
		sf.sequence = 0
	}

	// 服务器对时，可能会出现这个种情况
	if ts < sf.lastTimestamp {
		return 0, errors.New("id: invalid system clock")
	}

	sf.lastTimestamp = ts
	return sf.pack(), nil
}

func (sf *SnowFlake) waitNextMilli(ts int64) int64 {
	for ts == sf.lastTimestamp {
		time.Sleep(time.Millisecond)
		ts = sf.timestamp()
	}
	return ts
}

func (sf *SnowFlake) timestamp() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) - sf.time20200101Ms
}
