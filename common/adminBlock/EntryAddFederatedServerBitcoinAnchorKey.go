package adminBlock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type AddFederatedServerBitcoinAnchorKey struct {
	IdentityChainID interfaces.IHash
	KeyPriority     byte
	KeyType         byte //0=P2PKH 1=P2SH
	ECDSAPublicKey  primitives.ByteSlice20
}

var _ interfaces.IABEntry = (*AddFederatedServerBitcoinAnchorKey)(nil)
var _ interfaces.BinaryMarshallable = (*AddFederatedServerBitcoinAnchorKey)(nil)

func (c *AddFederatedServerBitcoinAnchorKey) UpdateState(state interfaces.IState) {

}

// Create a new DB Signature Entry
func NewAddFederatedServerBitcoinAnchorKey(identityChainID interfaces.IHash, keyPriority byte, keyType byte, ecdsaPublicKey primitives.ByteSlice20) (e *AddFederatedServerBitcoinAnchorKey) {
	e = new(AddFederatedServerBitcoinAnchorKey)
	e.IdentityChainID = identityChainID
	e.KeyPriority = keyPriority
	e.KeyType = keyType
	e.ECDSAPublicKey = ecdsaPublicKey
	return
}

func (e *AddFederatedServerBitcoinAnchorKey) Type() byte {
	return constants.TYPE_ADD_BTC_ANCHOR_KEY
}

func (e *AddFederatedServerBitcoinAnchorKey) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.Type()})

	data, err := e.IdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	buf.Write([]byte{e.KeyPriority})
	buf.Write([]byte{e.KeyType})

	data, err = e.ECDSAPublicKey.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Bytes(), nil
}

func (e *AddFederatedServerBitcoinAnchorKey) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	newData = data
	if newData[0] != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}
	newData = newData[1:]

	e.IdentityChainID = new(primitives.Hash)
	newData, err = e.IdentityChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	e.KeyPriority, newData = newData[0], newData[1:]
	e.KeyType, newData = newData[0], newData[1:]

	newData, err = e.ECDSAPublicKey.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	return
}

func (e *AddFederatedServerBitcoinAnchorKey) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddFederatedServerBitcoinAnchorKey) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddFederatedServerBitcoinAnchorKey) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AddFederatedServerBitcoinAnchorKey) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *AddFederatedServerBitcoinAnchorKey) String() string {
	str := fmt.Sprintf("AddFederatedServerBitcoinAnchorKey with Identity Chain ID = %x", e.IdentityChainID.Bytes()[:5])
	return str
}

func (e *AddFederatedServerBitcoinAnchorKey) IsInterpretable() bool {
	return false
}

func (e *AddFederatedServerBitcoinAnchorKey) Interpret() string {
	return ""
}

func (e *AddFederatedServerBitcoinAnchorKey) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
