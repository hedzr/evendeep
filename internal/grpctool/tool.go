package grpctool

import (
	"time"
	// "golang.org/x/sys/windows"
)

// // toTime converts an 8-byte Windows Filetime to time.Time.
// func toTime(t [8]byte) time.Time {
//	ft := &windows.Filetime{
//		LowDateTime:  binary.LittleEndian.Uint32(t[:4]),
//		HighDateTime: binary.LittleEndian.Uint32(t[4:]),
//	}
//	return time.Unix(0, ft.Nanoseconds())
// }

// Int64ToTime converts utcTimeNanoSeconds (int64) to time.Time.
// utcTimeNanoSeconds is a value in nanoseconds.
//
func Int64ToTime(utcTimeNanoSeconds int64) (tm time.Time) {
	// if utcTime == DefaultNilTimeNano {
	// 	return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	// }
	return time.Unix(0, utcTimeNanoSeconds)
}

// Int64SecondsToTime converts utcTimeSeconds (int64) to time.Time.
// utcTimeSeconds is a value in seconds.
//
// utcTimeSeconds is a standard unix timestamp.
//
// Unix returns the local Time corresponding to the given Unix
// time, sec seconds and nsec nanoseconds since January 1, 1970
// UTC. It is valid to pass nsec outside the range [0, 999999999].
// Not all sec values have a corresponding time value. One such
// value is 1<<63-1 (the largest int64 value).
func Int64SecondsToTime(utcTimeSeconds int64) (tm time.Time) {
	return time.Unix(utcTimeSeconds, 0)
}

// func TimestampToTime(ts *timestamp.Timestamp) (tm time.Time) {
//	var err error
//	tm, err = ptypes.Timestamp(ts)
//	if err != nil {
//		logrus.Warnf("CAN'T extract pb ptypes.timestamp to time: %v", err)
//	}
//	return
// }
//
// func TimeToTimestamp(tm time.Time) (ts *timestamp.Timestamp) {
//	var err error
//	ts, err = ptypes.TimestampProto(tm)
//	if err != nil {
//		logrus.Warnf("CAN'T convert time to pb ptypes.timestamp: %v", err)
//	}
//	return
// }
//
// func Int64ToTimestamp(utcTime int64) (ts *timestamp.Timestamp) {
//	tm := Int64ToTime(utcTime)
//	ts = &timestamp.Timestamp{Seconds: int64(tm.Unix()), Nanos: int32(tm.Nanosecond())}
//	return
// }

// DecodeZigZagInt decodes a protobuffer variable integer (zigzag format).
func DecodeZigZagInt(b []byte) (r int64, ate int) {
	var b1 byte
	var sh uint
	for i := 0; i < len(b); i++ {
		b1 = b[i]
		if b1&0x80 == 0 {
			r += int64(uint64(b1)) << sh
			break
		} else {
			r += int64(b1&0x7f) << sh //nolint:gomnd
			sh += 7
			ate++
		}
	}
	ate++
	return
}

// DecodeZigZagUint decodes a protobuffer variable integer (zigzag format) and return it as an uint64 umber.
func DecodeZigZagUint(b []byte) (r uint64, ate int) {
	var b1 byte
	var sh uint
	for i := 0; i < len(b); i++ {
		b1 = b[i]
		if b1&0x80 == 0 {
			r += uint64(b1) << sh
			break
		} else {
			r += uint64(b1&0x7f) << sh //nolint:gomnd
			sh += 7
			ate++
		}
	}
	ate++
	return
}
