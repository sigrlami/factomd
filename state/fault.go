// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"

	"github.com/FactomProject/factomd/common/messages"
)

type FaultState struct {
	FaultCore          FaultCore
	AmINegotiator      bool
	MyVoteTallied      bool
	VoteMap            map[[32]byte]interfaces.IFullSignature
	NegotiationOngoing bool
}

type FaultCore struct {
	// The following 5 fields represent the "Core" of the message
	// This should match the Core of FullServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	VMIndex       byte
	DBHeight      uint32
	Height        uint32
}

func fault(pl *ProcessList, vm *VM, vmIndex, height, tag int) {
	now := time.Now().Unix()

	if vm.whenFaulted == 0 {
		// if we did not previously consider this VM faulted
		// we simply mark it as faulted (by assigning it a nonzero whenFaulted time)
		vm.whenFaulted = now
	} else {
		if now-vm.whenFaulted > 20 {
			if !vm.faultInitiatedAlready {
				// after 20 seconds, we take initiative and
				// issue a server fault vote of our own
				craftAndSubmitFault(pl, vm, vmIndex, height)
				vm.faultInitiatedAlready = true
				//if I am negotiator... {
				go handleNegotiations(pl)
				//}
			}
			if now-vm.whenFaulted > 40 {
				// if !vm(f+1).goodNegotiator {
				//fault(vm(f+1))
				//}
			}

		}
	}
}

func handleNegotiations(pl *ProcessList) {
	for {
		for faultID, faultState := range pl.FaultMap {
			if faultState.AmINegotiator {
				craftAndSubmitFullFault(pl, faultID)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func craftAndSubmitFullFault(pl *ProcessList, faultID [32]byte) {
	fmt.Println("JUSTIN CRAFTANDSUB", pl.State.FactomNodeName)
	faultState := pl.FaultMap[faultID]
	fc := faultState.FaultCore

	sf := messages.NewServerFault(pl.State.GetTimestamp(), fc.ServerID, fc.AuditServerID, int(fc.VMIndex), fc.DBHeight, fc.Height)

	var listOfSigs []interfaces.IFullSignature
	for _, sig := range faultState.VoteMap {
		listOfSigs = append(listOfSigs, sig)
	}

	fullFault := messages.NewFullServerFault(sf, listOfSigs)
	absf := fullFault.ToAdminBlockEntry()
	pl.State.LeaderPL.AdminBlock.AddServerFault(absf)
	if fullFault != nil {
		fullFault.Sign(pl.State.serverPrivKey)
		pl.State.NetworkOutMsgQueue() <- fullFault
		fullFault.FollowerExecute(pl.State)
		//pl.AmINegotiator = false
		//delete(pl.FaultMap, faultID)
	}
}

func craftFault(pl *ProcessList, vm *VM, vmIndex int, height int) *messages.ServerFault {
	// TODO: if I am the Leader being faulted, I should respond by sending out
	// a MissingMsgResponse to everyone for the msg I'm being faulted for
	auditServerList := pl.State.GetOnlineAuditServers(pl.DBHeight)
	if len(auditServerList) > 0 {
		replacementServer := auditServerList[0]
		leaderMin := getLeaderMin(pl)

		faultedFed := pl.ServerMap[leaderMin][vmIndex]

		pl.FedServers[faultedFed].SetOnline(false)
		id := pl.FedServers[faultedFed].GetChainID()
		//NOMINATE
		sf := messages.NewServerFault(pl.State.GetTimestamp(), id, replacementServer.GetChainID(), vmIndex, pl.DBHeight, uint32(height))
		if sf != nil {
			sf.Sign(pl.State.serverPrivKey)
			return sf
		}
	}
	return nil
}

func craftAndSubmitFault(pl *ProcessList, vm *VM, vmIndex int, height int) {
	fmt.Println("JUSTIN CRASF", pl.State.FactomNodeName)
	// TODO: if I am the Leader being faulted, I should respond by sending out
	// a MissingMsgResponse to everyone for the msg I'm being faulted for
	auditServerList := pl.State.GetOnlineAuditServers(pl.DBHeight)
	if len(auditServerList) > 0 {
		replacementServer := auditServerList[0]
		leaderMin := getLeaderMin(pl)

		faultedFed := pl.ServerMap[leaderMin][vmIndex]

		pl.FedServers[faultedFed].SetOnline(false)
		id := pl.FedServers[faultedFed].GetChainID()
		//NOMINATE
		sf := messages.NewServerFault(pl.State.GetTimestamp(), id, replacementServer.GetChainID(), vmIndex, pl.DBHeight, uint32(height))
		if sf != nil {
			sf.Sign(pl.State.serverPrivKey)
			pl.State.NetworkOutMsgQueue() <- sf
			pl.State.InMsgQueue() <- sf
		}
	} else {
		for _, aud := range pl.AuditServers {
			aud.SetOnline(true)
		}
	}
}

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
	sf, _ := m.(*messages.ServerFault)
	pl := s.ProcessLists.Get(sf.DBHeight)

	if pl == nil {
		return
	}

	if pl.VMs[sf.VMIndex].whenFaulted == 0 {
		return
	}

	fmt.Println("JUSTIN FOLLEXSF", pl.State.FactomNodeName, sf.GetCoreHash().String()[:10])

	s.regularFaultExecution(sf, pl)
}

func (s *State) regularFaultExecution(sf *messages.ServerFault, pl *ProcessList) {
	var issuerID [32]byte
	rawIssuerID := sf.GetSignature().GetKey()
	for i := 0; i < 32; i++ {
		if i < len(rawIssuerID) {
			issuerID[i] = rawIssuerID[i]
		}
	}

	coreHash := sf.GetCoreHash().Fixed()
	fmt.Println("JUSTIN COREH:", pl.State.FactomNodeName, sf.GetCoreHash().String()[:10], len(pl.FaultMap))
	faultState, haveFaultMapped := pl.FaultMap[sf.GetCoreHash().Fixed()]
	if haveFaultMapped {
		fmt.Println(faultState)
	} else {
		fcore := FaultCore{ServerID: sf.ServerID, AuditServerID: sf.AuditServerID, VMIndex: sf.VMIndex, DBHeight: sf.DBHeight, Height: sf.Height}
		pl.FaultMap[coreHash] = FaultState{FaultCore: fcore, AmINegotiator: false, MyVoteTallied: false, VoteMap: make(map[[32]byte]interfaces.IFullSignature), NegotiationOngoing: false}
		faultState = pl.FaultMap[coreHash]
	}

	if faultState.VoteMap == nil {
		faultState.VoteMap = make(map[[32]byte]interfaces.IFullSignature)
	}

	lbytes, err := sf.MarshalForSignature()

	sfSig := sf.Signature.GetSignature()

	isPledge := false
	auth, _ := s.GetAuthority(sf.AuditServerID)
	if auth == nil {
		isPledge = false
	} else {
		valid, err := auth.VerifySignature(lbytes, sfSig)
		if err == nil && valid {
			isPledge = true
		}
	}

	sfSigned, err := s.VerifyAuthoritySignature(lbytes, sfSig, sf.DBHeight)

	if err == nil && (sfSigned > 0 || (sfSigned == 0 && isPledge)) {
		faultState.VoteMap[issuerID] = sf.GetSignature()
	}

	if s.Leader || s.IdentityChainID.IsSameAs(sf.AuditServerID) {
		cnt := len(faultState.VoteMap)
		var fedServerCnt int
		if pl != nil {
			fedServerCnt = len(pl.FedServers)
		} else {
			fedServerCnt = len(s.GetFedServers(sf.DBHeight))
		}
		responsibleFaulterIdx := (int(sf.VMIndex) + 1) % fedServerCnt
		if s.Leader && s.LeaderVMIndex == responsibleFaulterIdx {
			faultState.AmINegotiator = true
			pl.AmINegotiator = true
		}

		if !faultState.MyVoteTallied {
			s.matchFault(sf)
		}
		fmt.Println("JUSTIN CNT:", cnt, s.FactomNodeName)
	}

	pl.FaultMap[sf.GetCoreHash().Fixed()] = faultState
}

func (s *State) matchFault(sf *messages.ServerFault) {
	fmt.Println("JUSTIN MATCH FAULT", s.FactomNodeName, sf.GetCoreHash().String()[:10], sf.Signature.Bytes())
	if sf != nil {
		sf.Sign(s.serverPrivKey)
		s.NetworkOutMsgQueue() <- sf
		s.InMsgQueue() <- sf
		fmt.Println("JUSTIN MATCHED FAULT", s.FactomNodeName, sf.GetCoreHash().String()[:10], sf.Signature.Bytes())
	}
}

func wipeOutFaultsFor(pl *ProcessList, faultedServerID interfaces.IHash) {
	for faultID, faultState := range pl.FaultMap {
		if faultState.FaultCore.ServerID.IsSameAs(faultedServerID) {
			delete(pl.FaultMap, faultID)
		}
	}
}

func (s *State) FollowerExecuteFullFault(m interfaces.IMsg) {
	fullFault, _ := m.(*messages.FullServerFault)
	relevantPL := s.ProcessLists.Get(fullFault.DBHeight)

	auditServerList := s.GetAuditServers(fullFault.DBHeight)
	var theAuditReplacement interfaces.IFctServer

	for _, auditServer := range auditServerList {
		if auditServer.GetChainID().IsSameAs(fullFault.AuditServerID) {
			theAuditReplacement = auditServer
		}
	}
	if theAuditReplacement == nil {
		return
	}

	hasSignatureQuorum := m.Validate(s)
	if hasSignatureQuorum > 0 {
		if s.pledgedByAudit(fullFault) {
			fmt.Println("JUSTIN PLEDGE SUCCESSSSSSSSSSSSSSSSSSSSSSS", s.FactomNodeName, fullFault.AuditServerID.String()[:10])
			for listIdx, fedServ := range relevantPL.FedServers {
				if fedServ.GetChainID().IsSameAs(fullFault.ServerID) {
					relevantPL.FedServers[listIdx] = theAuditReplacement
					relevantPL.FedServers[listIdx].SetOnline(true)
					relevantPL.AddAuditServer(fedServ.GetChainID())
					s.RemoveAuditServer(fullFault.DBHeight, theAuditReplacement.GetChainID())
					if foundVM, vmindex := relevantPL.GetVirtualServers(s.CurrentMinute, theAuditReplacement.GetChainID()); foundVM {
						relevantPL.VMs[vmindex].faultHeight = -1
						relevantPL.VMs[vmindex].faultingEOM = 0
						relevantPL.VMs[vmindex].whenFaulted = 0
					}
					wipeOutFaultsFor(relevantPL, fullFault.ServerID)
					break
				}
			}

			s.Leader, s.LeaderVMIndex = s.LeaderPL.GetVirtualServers(s.CurrentMinute, s.IdentityChainID)
			delete(relevantPL.FaultMap, fullFault.GetCoreHash().Fixed())
			delete(relevantPL.FaultMap, fullFault.GetCoreHash().Fixed())
			return
		} else {
			fmt.Println("JUSTIN PLEDGE FAIL", s.FactomNodeName, fullFault.AuditServerID.String()[:10])

		}
	} else if hasSignatureQuorum == 0 {
		fmt.Println("JUSTIN not enough sigs!", s.FactomNodeName, fullFault.GetCoreHash().String()[:10])
	}

	theFaultState := relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()]

	if !theFaultState.MyVoteTallied {
		lbytes, err := fullFault.MarshalForSF()
		auth, _ := s.GetAuthority(s.IdentityChainID)
		if auth == nil || err != nil {
			return
		}
		for _, sig := range fullFault.SignatureList.List {
			ffSig := sig.GetSignature()
			valid, err := auth.VerifySignature(lbytes, ffSig)
			if err == nil && valid {
				theFaultState.MyVoteTallied = true
				relevantPL.FaultMap[fullFault.GetCoreHash().Fixed()] = theFaultState
				fmt.Println("JUSTIN TALLIED", s.FactomNodeName, fullFault.GetCoreHash().String()[:10])
				return
			}
		}
	}

}

func (s *State) pledgedByAudit(fullFault *messages.FullServerFault) bool {
	for _, a := range s.Authorities {
		if a.AuthorityChainID.IsSameAs(fullFault.AuditServerID) {
			marshalledSF, err := fullFault.MarshalForSF()
			if err == nil {
				for _, sig := range fullFault.SignatureList.List {
					sigVer, err := a.VerifySignature(marshalledSF, sig.GetSignature())
					if err == nil && sigVer {
						return true
					}
				}
			}
			break
		}
	}
	return false
}