package snmptrapper

import (
	"time"
	"net"
	"strings"

	types "github.com/chrusty/prometheus_webhook_snmptrapper/types"

	logrus "github.com/Sirupsen/logrus"
	snmpgo "github.com/k-sone/snmpgo"
)

func sendTrap(alert types.Alert, uptime uint32) {

	// Prepare an SNMP handler:
	snmp, err := snmpgo.NewSNMP(snmpgo.SNMPArguments{
		Version:   snmpgo.V2c,
		Address:   myConfig.SNMPTrapAddress,
		Retries:   myConfig.SNMPRetries,
		Community: myConfig.SNMPCommunity,
	})
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to create snmpgo.SNMP object")
		return
	} else {
		log.WithFields(logrus.Fields{"address": myConfig.SNMPTrapAddress, "retries": myConfig.SNMPRetries, "community": myConfig.SNMPCommunity}).Debug("Created snmpgo.SNMP object")
	}

	// Build VarBind list:
	var varBinds snmpgo.VarBinds

	// Insert uptime varbind
	varBinds = append(varBinds, snmpgo.NewVarBind(snmpgo.OidSysUpTime, snmpgo.NewTimeTicks(uptime)))

	// The "enterprise OID" for the trap (rising/firing or falling/recovery):
	if alert.Status == "firing" {
		varBinds = append(varBinds, snmpgo.NewVarBind(snmpgo.OidSnmpTrap, trapOIDs.FiringTrap))
	} else {
		varBinds = append(varBinds, snmpgo.NewVarBind(snmpgo.OidSnmpTrap, trapOIDs.RecoveryTrap))
	}

	// Insert the AlertManager variables:
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Instance, snmpgo.NewOctetString([]byte(alert.Labels["instance"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Service, snmpgo.NewOctetString([]byte(alert.Labels["service"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Location, snmpgo.NewOctetString([]byte(alert.Labels["location"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Severity, snmpgo.NewOctetString([]byte(alert.Labels["severity"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.Description, snmpgo.NewOctetString([]byte(alert.Annotations["description"]))))
	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.JobName, snmpgo.NewOctetString([]byte(alert.Labels["job"]))))

	// Insert the timestamp varbind
	if alert.Status == "firing" {
		varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.TimeStamp, snmpgo.NewOctetString([]byte(alert.StartsAt.Format(time.RFC3339)))))
	} else {
		varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.TimeStamp, snmpgo.NewOctetString([]byte(alert.EndsAt.Format(time.RFC3339)))))
	}

	varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.URL, snmpgo.NewOctetString([]byte(alert.GeneratorURL))))

	// Set trap source address
	ips, err := net.LookupIP(strings.Split(alert.Labels["instance"], ":")[0])
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to resolve name")
	} else {
		ip := ips[0]
		varBinds = append(varBinds, snmpgo.NewVarBind(trapOIDs.TrapAddress, snmpgo.NewIpaddress(ip[12], ip[13], ip[14], ip[15])))
	}

	// Create an SNMP "connection":
	if err = snmp.Open(); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to open SNMP connection")
		return
	}
	defer snmp.Close()

	// Send the trap:
	if err = snmp.V2Trap(varBinds); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to send SNMP trap")
		return
	} else {
		log.WithFields(logrus.Fields{"status": alert.Status}).Info("It's a trap!")
	}
}
