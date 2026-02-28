package id

import (
	"testing"
	"time"
)

func TestTimeLen(t *testing.T) {
	time1, _ := time.Parse("2006-01-02 15:04:05", "2020-01-01 11:11:11")
	time2, _ := time.Parse("2006-01-02 15:04:05", "2030-01-01 22:22:22")
	time1Nano := time1.UnixNano()
	time2Nano := time2.UnixNano()
	timeNano := time2Nano - time1Nano
	timeMilli := timeNano / int64(time.Millisecond)
	t.Log(timeNano)
	t.Log(timeMilli)
}

func TestCreateId(t *testing.T) {
	idCreated := make(map[int64]bool, 400000)

	begin := time.Now().UnixNano()
	for i := 0; i < 400000; i++ {
		id, _ := General()
		if _, ok := idCreated[id]; ok {
			t.Error("gid: general id duplicate")
		} else {
			idCreated[id] = true
		}
	}
	end := time.Now().UnixNano()
	used := (end - begin) / int64(time.Millisecond)
	t.Log("gid: create 40000 id use time = ", used)
}

func BenchmarkCreateId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		General()
	}
}
