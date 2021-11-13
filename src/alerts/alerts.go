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
	Volume   int
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
	VolumeSettings   = map[string]int{
		// Subscription Volumes
		"Subscriber1.wav":      60,
		"Subscriber2.wav":      70,
		"Subscriber3.wav":      72,
		"Subscriber6.wav":      74,
		"Subscriber9.wav":      76,
		"Subscriber12.wav":     78,
		"Subscriber18.wav":     80,
		"Subscriber24.wav":     82,
		"Subscriber30.wav":     84,
		"Subscriber36.wav":     86,
		"Subscriber42.wav":     88,
		"Subscriber48.wav":     90,
		"Subscriber54.wav":     92,
		"Subscriber60.wav":     94,
		"SubscriberSuper.wav":  96,
		"SubscriberMega.wav":   20,
		"SubscriberGift.wav":   50,
		"SubscriberGift5.wav":  60,
		"SubscriberGift25.wav": 90,
		"SubscriberGift95.wav": 100,
		// Donation Volumes
		"DonationDefault.wav": 80,
		"Donation420.wav":     80,
		"Donation10.wav":      40,
		"Donation50.wav":      100,
		// Cheer Volumes
		"CheerDefault.wav": 64,
		"Cheer1.wav":       64,
		"Cheer100.wav":     64,
		"Cheer500.wav":     34,
		"Cheer1000.wav":    28,
		"Cheer10000.wav":   44,
		"Cheer100000.wav":  100,
	}
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
			Volume:   VolumeSettings[v],
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
			Volume:   VolumeSettings[v],
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
			Volume:   VolumeSettings[v],
		}
	}
}
