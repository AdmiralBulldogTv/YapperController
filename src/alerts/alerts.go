package alerts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gobuffalo/packr/v2"
)

type Alert struct {
	Name     string
	Data     []byte
	CheckSum string
}

func (a Alert) ToName() string {
	file := a.Name
	idx := strings.LastIndexByte(file, '.')
	suffix := file[idx:]
	file = file[:idx]

	return fmt.Sprintf("%s.%s%s", file, a.CheckSum, suffix)
}

var (
	CheerAlerts      = map[string]Alert{}
	DonationAlerts   = map[string]Alert{}
	SubscriberAlerts = map[string]Alert{}
)

func init() {
	cheerBox := packr.New("cheer-alerts", "./DonationAlerts/Cheer")
	donationBox := packr.New("donation-alerts", "./DonationAlerts/Donation")
	subscriberBox := packr.New("subscriber-alerts", "./DonationAlerts/Subscriber")

	h := sha256.New()
	for _, v := range cheerBox.List() {
		data, _ := cheerBox.Find(v)
		h.Reset()
		_, _ = h.Write(data)

		CheerAlerts[v] = Alert{
			Name:     v,
			Data:     data,
			CheckSum: hex.EncodeToString(h.Sum(nil))[:8],
		}
	}

	for _, v := range donationBox.List() {
		data, _ := donationBox.Find(v)
		h.Reset()
		_, _ = h.Write(data)

		DonationAlerts[v] = Alert{
			Name:     v,
			Data:     data,
			CheckSum: hex.EncodeToString(h.Sum(nil))[:8],
		}
	}

	for _, v := range subscriberBox.List() {
		data, _ := subscriberBox.Find(v)
		h.Reset()
		_, _ = h.Write(data)

		SubscriberAlerts[v] = Alert{
			Name:     v,
			Data:     data,
			CheckSum: hex.EncodeToString(h.Sum(nil))[:8],
		}
	}
}
