package notify

import (
	"fmt"
	"github.com/SongZihuan/https-watcher/src/config"
	"github.com/SongZihuan/https-watcher/src/smtpserver"
	"github.com/SongZihuan/https-watcher/src/utils"
	"github.com/SongZihuan/https-watcher/src/wxrobot"
	"strings"
	"sync"
	"time"
)

const (
	StatusError     = 1
	StatusOutOfDate = 2
)

type urlRecord struct {
	Name     string
	URL      string
	Status   int
	Deadline time.Duration
	ErrorMsg string
}

var startTime time.Time
var outOfDateRecords sync.Map
var errorRecords sync.Map

func InitNotify() error {
	if !config.IsReady() {
		panic("config is not ready")
	}

	startTime = time.Now().In(config.TimeZone())

	err := smtpserver.InitSmtp()
	if err != nil {
		return err
	}

	return nil
}

func NewOutOfDateRecord(name string, url string, deadline time.Duration) {
	if name == "" {
		name = url
	}

	outOfDateRecords.Store(name, &urlRecord{
		Name:     name,
		URL:      url,
		Status:   StatusOutOfDate,
		Deadline: deadline,
	})
}

func NewErrorRecord(name string, url string, err string) {
	if name == "" {
		name = url
	}

	errorRecords.Store(name, &urlRecord{
		Name:     name,
		URL:      url,
		Status:   StatusError,
		ErrorMsg: err,
	})
}

func SendOutOfDateNotify() {
	var res strings.Builder
	var expiredCount uint64 = 0
	var expiringSoonCount uint64 = 0

	res.WriteString(fmt.Sprintf("日期：%s %s\n", startTime.Format("2006-01-02 15:04:05"), startTime.Location().String()))

	outOfDateRecords.Range(func(key, value any) bool {
		record, ok := value.(*urlRecord)
		if !ok {
			return true
		}

		if record.Status != StatusOutOfDate {
			return true
		}

		if record.Deadline <= 0 {
			expiredCount += 1
			res.WriteString(fmt.Sprintf("- %s 已过期\n", record.Name))
		} else {
			expiringSoonCount += 1
			res.WriteString(fmt.Sprintf("- %s 剩余时间: %s\n", record.Name, utils.TimeDurationToStringCN(record.Deadline)))
		}

		return true
	})

	if expiredCount+expiringSoonCount <= 0 {
		// 无任何记录
		return
	}

	res.WriteString(fmt.Sprintf("共计：过期 %d 条，即将过期 %d 条。\n", expiredCount, expiringSoonCount))
	res.WriteString("完毕\n")
	msg := res.String()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		wxrobot.SendNotify(msg)
	}()

	go func() {
		defer wg.Done()
		smtpserver.SendNotify(msg)
	}()

	wg.Wait()
}

func SendErrorNotify() {
	var res strings.Builder
	var count uint64 = 0

	res.WriteString(fmt.Sprintf("日期：%s %s\n", startTime.Format("2006-01-02 15:04:05"), startTime.Location().String()))

	errorRecords.Range(func(key, value any) bool {
		record, ok := value.(*urlRecord)
		if !ok {
			return true
		}

		if record.Status != StatusError {
			return true
		}

		count += 1
		res.WriteString(fmt.Sprintf("- 检查 %s 出错: %s\n", record.Name, record.ErrorMsg))

		return true
	})

	if count <= 0 {
		// 无任何记录
		return
	}

	res.WriteString(fmt.Sprintf("共计：出粗 %d 条。\n", count))
	res.WriteString("完毕\n")
	msg := res.String()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		wxrobot.SendNotify(msg)
	}()

	go func() {
		defer wg.Done()
		smtpserver.SendNotify(msg)
	}()

	wg.Wait()
}

func SendNotify() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		SendOutOfDateNotify()
	}()

	go func() {
		defer wg.Done()
		SendErrorNotify()
	}()

	wg.Wait()
}
