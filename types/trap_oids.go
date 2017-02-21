package types

import (
	snmpgo "github.com/k-sone/snmpgo"
)

type TrapOIDs struct {
	TrapAddress  *snmpgo.Oid
	FiringTrap   *snmpgo.Oid
	RecoveryTrap *snmpgo.Oid
	Instance     *snmpgo.Oid
	Service      *snmpgo.Oid
	Location     *snmpgo.Oid
	Severity     *snmpgo.Oid
	Description  *snmpgo.Oid
	JobName      *snmpgo.Oid
	TimeStamp    *snmpgo.Oid
	URL          *snmpgo.Oid
}
