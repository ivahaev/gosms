package gosms

import (
	"encoding/json"
	"github.com/ivahaev/gosms/sdb"
	log "github.com/ivahaev/go-logger"
	"time"
)

var bucket = "sms"

func insertMessage(sms *SMS) error {
	log.Info("insertMessage ", sms)
	err := sdb.Save(bucket, sms.UUID, sms)
	if err != nil {
		log.Error("Error when inserting message: ", err)
	}
	return nil
}

func updateMessageStatus(sms SMS) error {
	log.Info("updateMessageStatus ", sms)
	encoded, err := sdb.Get(bucket, sms.UUID)
	if err != nil {
		log.Error("Error when getting message: ", err)
		return err
	}
	oldSms := SMS{}
	err = json.Unmarshal(encoded, &oldSms)
	if err != nil {
		log.Error("Error when unmarshaling message: ", err)
		return err
	}
	oldSms.Status = sms.Status
	oldSms.Retries = sms.Retries
	oldSms.Device = sms.Device
	oldSms.UpdatedAt = time.Now()
	err = sdb.Save(bucket, oldSms.UUID, oldSms)
	if err != nil {
		log.Error("Error when inserting message: ", err)
	}
	return err
}

func getPendingMessages(bufferSize int) (result []SMS, err error) {
	log.Info("getPendingMessages ")
	allMessages, err := sdb.GetAll(bucket)
	if err != nil {
		log.Error(err)
		return
	}
	result = []SMS{}
	for _, _m := range allMessages {
		sms := SMS{}
		err := json.Unmarshal(_m, &sms)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		if sms.Status != SMSProcessed && sms.Retries < SMSRetryLimit {
			result = append(result, sms)
		}
		if len(result) >= bufferSize {
			break
		}
	}

	return
}

func GetMessages(filter string) (result []SMS, err error) {
	log.Info("GetMessages")
	allMessages, err := sdb.GetAll(bucket)
	if err != nil {
		log.Error(err)
		return
	}
	result = []SMS{}
	for _, _m := range allMessages {
		sms := SMS{}
		err := json.Unmarshal(_m, &sms)
		if err != nil {
			log.Error("Error when unmarshaling message: ", err)
			return nil, err
		}
//		if sms.Status != SMSProcessed && sms.Retries < SMSRetryLimit {
			result = append(result, sms)
//		}
	}
	return result, nil
}

func GetLast7DaysMessageCount() (map[string]int, error) {
	log.Info("GetLast7DaysMessageCount")
	allMessages, err := GetMessages("")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	dayCount := make(map[string]int)
	fromDate := time.Now().Add(-time.Hour * 24 * 8)
	for _, sms := range allMessages {
		if sms.CreatedAt.After(fromDate) {
			createdAt := sms.CreatedAt.Format("2006-01-02")
			count, ok := dayCount[createdAt]
			if ok {
				count++
			} else {
				count = 1
			}
			dayCount[createdAt] = count
		}
	}
	return dayCount, nil
}

func GetStatusSummary() ([]int, error) {
	log.Info("GetStatusSummary")
	allMessages, err := GetMessages("")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	statusSummary := make([]int, 3)
	for _, sms := range allMessages {
		statusSummary[sms.Status]++
	}
	return statusSummary, nil
}
