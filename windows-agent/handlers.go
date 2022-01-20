package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func adjoinHandler(c *gin.Context) {
	setLog()
	domain := c.Param("domain")
	domain = Agentconfig.Domain
	shell, err := setupShell()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
		return
	}
	cmd := fmt.Sprintf("$username = \"$%v\\administrator\"; ", domain) +
		"$password = \"Ablecloud1!\" | ConvertTo-SecureString -asPlainText -Force; " +
		"$credential = New-Object System.Management.Automation.PSCredential($username,$password); " +
		fmt.Sprintf("Add-Computer -DomainName %v -Credential $credential", domain)
	log.Debugln(cmd)
	output, err := shell.Exec(cmd)
	log.Debugf("output: %v", output)
	log.Debugf("error: %v", err)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
		return
	}
	cmd = "shutdown /r /t 10"
	log.Debugln(cmd)
	output, err = shell.Exec(cmd)
	log.Debugf("output: %v", output)
	log.Debugf("error: %v", err)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
		return
	}
	Agentconfig.Status = "Joining"
	_ = AgentSave()
	c.JSON(http.StatusOK, output)
	return
}

func checkStatus() {

	shell, err := setupShell()
	if err != nil {
		log.Errorf("%v", err)
	}

	//rename
	hostname, err := shell.Exec("hostname")
	if err != nil {
		log.Errorf("%v", err)
	}
	hostname = strings.TrimSpace(hostname)
	log.Println(hostname)
	log.Println(Agentconfig.HostName)
	if !strings.EqualFold(hostname, Agentconfig.HostName) && Agentconfig.HostName != "" {
		cmd := fmt.Sprintf("Rename-Computer -NewName %v", Agentconfig.HostName)
		log.Errorf("cmd: %v", cmd)
		out, err := shell.Exec(cmd)
		log.Errorf("out: %v", out)
		log.Errorf("err: %v", err)
		Agentconfig.Status = Agentconfig.Status + " renamehost"
		AgentSave()
		cmd = fmt.Sprintf("shutdown /r /t 10")
		log.Errorf("cmd: %v", cmd)
		out, err = shell.Exec(cmd)
		log.Errorf("out: %v", out)
		log.Errorf("err: %v", err)
	}

	//ad join
	if strings.Contains(Agentconfig.Status, "renamehost") {
		cmd := "get-computerinfo -property csdomainrole | Format-list"
		output, err := shell.Exec(cmd)
		log.Debugf("output: %v", output)
		log.Debugf("error: %v", err)
		outputs := strings.Split(strings.TrimSpace(output), ":")
		log.Infof("DomainRole: %v", outputs)
		if strings.TrimSpace(outputs[1]) == "StandaloneWorkstation" {
			cmd := fmt.Sprintf("$username = \"administrator@%v\"; ", Agentconfig.Domain) +
				"$password = \"Ablecloud1!\" | ConvertTo-SecureString -asPlainText -Force; " +
				"$credential = New-Object System.Management.Automation.PSCredential($username,$password); " +
				fmt.Sprintf("Add-Computer -DomainName %v -Credential $credential", Agentconfig.Domain)
			log.Debugln(cmd)
			output, err := shell.Exec(cmd)
			log.Debugf("output: %v", output)
			log.Debugf("error: %v", err)
			cmd = "shutdown /r /t 10"
			log.Debugln(cmd)
			Agentconfig.Status = Agentconfig.Status + " adjoin"
			AgentSave()
			output, err = shell.Exec(cmd)
		}
	}

	//service 권한 elevate
	cmd := "c:\\Works-DC\\nssm.exe get Works-DC ObjectName"
	output_, err := shell.Exec(cmd)
	output := strings.TrimSpace(output_)
	if strings.EqualFold(output, "localsystem") {
		aduser := fmt.Sprintf("%v\\%v", Agentconfig.Domain, Agentconfig.ADusername)
		service := fmt.Sprintf("c:\\Works-DC\\nssm.exe set Works-DC objectName %v %v", aduser, Agentconfig.ADpassword)
		stdout1, err1 := shell.Exec(service)
		log.Infof("stdout: %v, \nstderr: %v\n", stdout1, err1)
		//service = fmt.Sprintf("c:\\Works-DC\\nssm.exe restart Works-DC > c:\\Works-DC\\nssm.txt")
		service = "shutdown /r /f /t 10"
		Agentconfig.Status = Agentconfig.Status + " serviceupdated"
		AgentSave()
		stdout2, err2 := shell.Exec(service)
		log.Infof("stdout: %v, \nstderr: %v\n", stdout2, err2)
	}
}

func bootstrapHandler(c *gin.Context) {
	setLog()

	//computername := c.Param("computername")
	domain := c.PostForm("domain")
	shell, err := setupShell()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
		return
	}
	service1 := fmt.Sprintf("c:\\agent\\nssm.exe set \"Ablecloud Works Agent\" objectName %v\\administrator Ablecloud1!", domain)
	output1, err1 := shell.Exec(service1)
	if err1 != nil {

		log.Errorf("err1: %v", service1)
		log.Errorf("err1: %v", err1)
		c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err1.Error(), Target: "getComputer"})
		return
	}
	service2 := fmt.Sprintf("gpupdate /force")
	output2, err2 := shell.Exec(service2)
	if err2 != nil {
		log.Errorf("err1: %v", service2)
		log.Errorf("err2: %v", err2)
		c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err2.Error(), Target: "getComputer"})
		return
	}
	c.JSON(http.StatusOK, map[string]string{"output1": output1, "output2": output2})
	return
}

func adStatusHandler(c *gin.Context) {
	setLog()
	shell, err := setupShell()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
		return
	}
	output, err := shell.Exec("get-computerinfo -property csdomain | format-list")
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
		return
	}
	domain := strings.TrimSpace(strings.Split(strings.TrimSpace(output), ":")[1])

	if strings.EqualFold(domain, "WORKGROUP") && Agentconfig.Status != "Joining" {
		Agentconfig.Status = "Pending"
		_ = AgentSave()
		c.JSON(http.StatusOK,
			map[string]string{
				"status": Agentconfig.Status,
				"next":   "PUT <vdi>/ad/:domain/",
			})
		return
	} else {
		output, err := shell.Exec("get-computerinfo -property csname | format-list")
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
			return
		}
		comname := strings.TrimSpace(strings.Split(strings.TrimSpace(output), ":")[1])
		output, err = shell.Exec(fmt.Sprintf("gpresult /scope computer /r | select-string -pattern cn=%v", comname))
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
			return
		}
		if strings.Contains(strings.TrimSpace(output), "OU=") {
			Agentconfig.Status = "Joined"
			_ = AgentSave()
		}
		output, err = shell.Exec("gpresult /scope computer /r | select-string -pattern remote")
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, errorModel{Msg: err.Error(), Target: "getComputer"})
			return
		}
		if strings.Contains(strings.TrimSpace(output), "remotefx") {
			Agentconfig.Status = "Ready"
			_ = AgentSave()
		}
		if Agentconfig.Status == "Joining" {
			c.JSON(http.StatusOK,
				map[string]string{
					"status": Agentconfig.Status,
					"next":   "PUT <dc>/computer/:computername/:groupname",
				})
			return
		} else if Agentconfig.Status == "Joined" {
			c.JSON(http.StatusOK,
				map[string]string{
					"status": Agentconfig.Status,
					"next":   "GET <vdi>/cmd/?timeout=60&cmd=gpupdate /force",
				})
			return
		} else if Agentconfig.Status == "Ready" {
			c.JSON(http.StatusOK,
				map[string]string{
					"status": Agentconfig.Status,
					"next":   "Ready to use",
				})
			return
		}
	}
	c.JSON(http.StatusOK, output)
	return
}
