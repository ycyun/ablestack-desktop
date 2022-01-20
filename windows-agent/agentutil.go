package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type AGENTCONFIG struct {
	WorksServer string `json:"WorksServer"`
	WorksPort   int    `json:"WorksPort"`
	Type        string `json:"Type"`
	UUID        string `json:"UUID"`
	Silent      bool   `json:"Silent"`
	Interval    int    `json:"Interval"`
	Status      string `json:"Status"`
	HostName    string `json:"HostName"`
	Domain    	string `json:"Domain"`
	ADusername	string `json:"ADusername"`
	ADpassword	string `json:"ADpassword"`
}
var Agentconfig = AGENTCONFIG{Silent: false, Interval: 10, ADusername: "Administrator", ADpassword: "Ablecloud1!"}

func AgentInit() (err error) {
	data, err := os.Open("conf.json")
	if err != nil {
		log.Fatalf("ADinit: %s", err)
		return err
	}

	byteValue, err := ioutil.ReadAll(data)
	if err != nil {
		log.Fatalf("ADinit: %s", err)
		return err
	}

	err = json.Unmarshal(byteValue, &Agentconfig)
	if err != nil {
		log.Fatalf("ADinit: %s", err)
		return err
	}
	log.Infof("Agentconfig: %v", Agentconfig)
	return nil

}


func AgentSave() (err error) {
	data, err := os.OpenFile("conf.json", os.O_WRONLY, 0777)
	if err != nil {
		log.Fatalf("agentsave: %s", err)
		return err
	}
	byteValue, err:=json.MarshalIndent(Agentconfig, "", "  ")
	if err != nil {
		log.Fatalf("ADinit: %s", err)
		return err
	}
	log.Errorf("Save %s", byteValue)
	_, err = data.Write(byteValue)
	if err != nil {
		log.Fatalf("ADinit: %s", err)
		return err
	}
	return nil

}
