package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

var sendAlert func(user string, alert Alert)

func main() {
	sendAlert = func(user string, alert Alert) {
		go func(user string, alert Alert) {
			alert.User = user
			fmt.Println(alert)
		}(user, alert)
	}
	deviceConfig := DeviceConfig{
		Name: "test",
		Type: "foraO2",
	}

	var ret DeviceConfig
	//wcRes := `{"name":"test", "deviceId":"XXXXXXXXXXXXX", "patBed":"1-1", "patId":"3524761", "type":"whizConnect","config":{"ssid":"test", "password":"123", "registeredDevice":{"whizConnect":"XXXXXXXXXXXX","whizPad":"FFFFFFFFFFFF","foraD40":"AAAAAAAAAAAA","foraIR42":"BBBBBBBBBBBB","foraO2":"CCCCCCCCCCCC","foraP30":"DDDDDDDDDDDDD","locate":"EEEEEEEEEEEE","miTemp":"FFFFFFFFFFFF"}}}`
	foraRes := `{"name":"test", "deviceId":"XXXXXXXXXXXXX", "patBed":"1-1", "patId":"3524761", "type":"foraO2","config":{"pulse":28, "spO2":30}}`
	errWC := json.Unmarshal([]byte(foraRes), &ret)
	if errWC != nil {
		fmt.Errorf("unmarshal whizconnect data error")
	}
	//errFora := json.Unmarshal([]byte(foraRes), &ret)
	//if errFora != nil {
	//	fmt.Errorf("unmarshal fora data error")
	//}

	fmt.Println(deviceConfig)
	foraO2Config := ForaO2Config{}
	deviceConfig.Config = foraO2Config

	foraO2Event := ForaO2Event{SpO2: 10, Pulse: 10}
	needAlert, err1 := deviceConfig.Config.checkAlert(foraO2Event)
	needAlert, err2 := deviceConfig.Config.checkAlert(BedEvent{
		BedSensor: "111111111111111111111111111111",
	})

	fmt.Println("fora alert error:", err1)
	fmt.Println("bed alert error:", err2)
	fmt.Println(needAlert)
	fmt.Println(ret)
}

type DeviceConfig struct {
	Name     string          `json:"name"`
	DeviceID string          `json:"deviceId"`
	PatBed   string          `json:"patBed"`
	PatID    string          `json:"patId"`
	Type     string          `json:"type"`
	Config   ConfigInterface `json:"config"`
}

func (m *DeviceConfig) UnmarshalJSON(b []byte) error {
	tempData := &struct {
		Name     string `json:"name"`
		DeviceID string `json:"deviceId"`
		PatBed   string `json:"patBed"`
		PatID    string `json:"patId"`
		Type     string `json:"type"`
	}{}
	type rec struct {
		Name     string          `json:"name"`
		DeviceID string          `json:"deviceId"`
		PatBed   string          `json:"patBed"`
		PatID    string          `json:"patId"`
		Type     string          `json:"type"`
		Config   ConfigInterface `json:"config"`
	}
	//get type
	err := json.Unmarshal(b, tempData)
	if err != nil {
		fmt.Println("A", err)
	}
	//different type of struct
	var reallyData rec
	switch tempData.Type {
	case ForaO2Config{}.Type():
		reallyData = rec{
			Name:     "",
			DeviceID: "",
			PatBed:   "",
			PatID:    "",
			Type:     "",
			Config:   ForaO2Config{},
		}
		err = json.Unmarshal(b, &reallyData)
	case WhizConnectConfig{}.Type():
		reallyData = rec{
			Name:     "",
			DeviceID: "",
			PatBed:   "",
			PatID:    "",
			Type:     "",
			Config:   WhizConnectConfig{},
		}
		err = json.Unmarshal(b, &reallyData)
	default:
		return fmt.Errorf("not found")
	}
	fmt.Println(err.Error())
	m.Name = reallyData.Name
	m.DeviceID = reallyData.DeviceID
	m.PatBed = reallyData.PatBed
	m.PatID = reallyData.PatID
	m.Type = reallyData.Type
	return nil
}

type ConfigInterface interface {
	Type() string
	Raw() interface{}
	checkAlert(interface{}) (bool, error) //event in, result out
}

type ForaO2Config struct {
	PulseT int `json:"pulse"`
	SpO2T  int `json:"spO2"`
	//Alert struct {
	//	Pulse  DeviceAlert `json:"pulse"`
	//	SpO2   DeviceAlert `json:"spO2"`
	//	Enable bool        `json:"enable"`
	//} `json:"alert" bson:"alert"`
}

type DeviceAlert struct {
	UpperBound int  `json:"upperBound" bson:"upper_bound"`
	LowerBound int  `json:"lowerBound" bson:"lower_bound"`
	Threshold  int  `json:"threshold" bson:"threshold"`
	Duration   int  `json:"duration" bson:"duration"`
	StartHour  int  `json:"startHour" bson:"start_hour"`
	EndHour    int  `json:"endHour" bson:"end_hour"`
	Enable     bool `json:"enable" bson:"enable"`
}

func (f ForaO2Config) Type() string {
	return "foraO2"
}

func (f ForaO2Config) Raw() interface{} {
	return f
}

func (f WhizConnectConfig) Type() string {
	return "whizConnect"
}

func (f WhizConnectConfig) Raw() interface{} {
	return f
}

func (f ForaO2Config) checkAlert(i interface{}) (bool, error) {
	fmt.Println("i", i)
	foraO2Event, ok := i.(ForaO2Event)
	fmt.Println("foraO2 event", foraO2Event)
	if !ok {
		return false, errors.New("type error")
	}
	if foraO2Event.SpO2 > f.SpO2T {
		sendAlert("test", Alert{Name: "SpO2 over Alert"})
		return true, nil
	}
	if foraO2Event.Pulse > f.PulseT {
		sendAlert("test", Alert{Name: "Pulse over Alert"})
		return true, nil
	}
	return false, nil
}

func (f WhizConnectConfig) checkAlert(i interface{}) (bool, error) {
	return false, nil
}

type ForaO2Event struct {
	SpO2  int `json:"SpO2" bson:"SpO2"`
	Pulse int `json:"pulse" bson:"pulse"`
}

type BedEvent struct {
	BedSensor string `json:"bedSensor"`
}

type Alert struct {
	User     string `json:"-" bson:"user"`
	DeviceID string `json:"deviceId" bson:"device_id"`
	Epoch    int64  `json:"epoch" bson:"epoch"`
	Type     string `json:"type" bson:"type"`
	Event    int    `json:"event" bson:"event"`
	Name     string `json:"name" bson:"name"`
}

type WhizConnectConfig struct {
	SSId             string `json:"ssid" bson:"ssid"`
	Password         string `json:"password" bson:"password"`
	RegisteredDevice struct {
		WhizConnect string `json:"whizConnect" bson:"whizConnect"`
		WhizPad     string `json:"whizPad" bson:"whizPad"`
		ForaD40     string `json:"foraD40" bson:"foraD40"`
		ForaIR42    string `json:"foraIR42" bson:"foraIR42"`
		ForaO2      string `json:"foraO2" bson:"foraO2"`
		ForaP30     string `json:"foraP30" bson:"foraP30"`
		Locate      string `json:"locate" bson:"locate"`
		MiTemp      string `json:"miTemp" bson:"mi_temp"`
		Diaper      string `json:"diaper" bson:"diaper"`
	} `json:"registeredDevice" bson:"registered_device"`
}

type BedConfig struct {
	Name string `json:"name"`
	//StageAlert struct {
	//	Day   StageAlert `json:"day"`
	//	Noon  StageAlert `json:"noon"`
	//	Night StageAlert `json:"night"`
	//} `json:"stageAlert"`
	Gender        int `json:"gender"`
	ActivityAlert struct {
		UniColor struct {
			AllDay bool `json:"allDay"`
		} `json:"uniColor"`
		Over struct {
			Checked   bool `json:"checked"`
			Threshold int  `json:"threshold"`
		} `json:"over"`
		Range struct {
			AllDay bool `json:"allDay"`
			From   int  `json:"from"`
			To     int  `json:"to"`
		} `json:"range"`
		Lie struct {
			Checked   bool `json:"checked"`
			Threshold int  `json:"threshold"`
		} `json:"lie"`
		Lower struct {
			Checked   bool `json:"checked"`
			Threshold int  `json:"threshold"`
			Duration  int  `json:"duration"`
		} `json:"lower"`
		RollOver struct {
			Checked   bool `json:"checked"`
			Threshold int  `json:"threshold"`
			Duration  int  `json:"duration"`
		} `json:"rollOver"`
	} `json:"activityAlert"`
	DeviceID       string `json:"deviceId"`
	PatBed         string `json:"patBed"`
	PatID          string `json:"patId"`
	Created        int64  `json:"created"`
	Index          int    `json:"index"`
	MiTempDeviceID string `json:"miTempDeviceId"`
}
