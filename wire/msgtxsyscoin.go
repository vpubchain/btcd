// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"
	"encoding/binary"
	"encoding/base64"
)
const (
	MAX_GUID_LENGTH = 20
	MAX_VALUE_LENGTH = 512
	MAX_SYMBOL_SIZE = 12 // up to 9 characters base64 decoded
	MAX_SIG_SIZE = 65
	MAX_RLP_SIZE = 4096
)
const ( 	
	ASSET_UPDATE_DATA = 1 // can you update public data field?
  	ASSET_UPDATE_CONTRACT = 2 // can you update smart contract?
 	ASSET_UPDATE_SUPPLY = 4 // can you update supply?
 	ASSET_UPDATE_NOTARY_KEY = 8 // can you update notary?
 	ASSET_UPDATE_NOTARY_DETAILS = 16 // can you update notary details?
 	ASSET_UPDATE_AUXFEE_KEY = 32 // can you update aux fees?
	ASSET_UPDATE_AUXFEE_DETAILS = 64 // can you update aux fees details?
 	ASSET_UPDATE_CAPABILITYFLAGS = 128 // can you update capability flags?
 	ASSET_UPDATE_ALL = 255
)
type AssetOutValueType struct {
	N uint32
	ValueSat int64
}
type AssetOutType struct {
	AssetGuid uint32
	Values []AssetOutValueType
	NotarySig []byte
}
type AssetAllocationType struct {
	VoutAssets []AssetOutType
}
type NotaryDetailsType struct {
	EndPoint []byte
	InstantTransfers uint8
	HDRequired uint8
}
type AuxFeesType struct {
	Bound int64
	Percent uint16
}
type AuxFeeDetailsType struct {
	AuxFees []AuxFeesType
}
type AssetType struct {
	Allocation AssetAllocationType
	Contract []byte
	PrevContract []byte
	Symbol string
	PubData []byte
	PrevPubData []byte
	NotaryKeyID []byte
	PrevNotaryKeyID []byte
	NotaryDetails NotaryDetailsType
	PrevNotaryDetails NotaryDetailsType
	AuxFeeKeyID []byte
	PrevAuxFeeKeyID []byte
	AuxFeeDetails AuxFeeDetailsType
	PrevAuxFeeDetails AuxFeeDetailsType
	Balance int64
	TotalSupply int64
	MaxSupply int64
	Precision uint8
	UpdateCapabilityFlags uint8
	PrevUpdateCapabilityFlags uint8
	UpdateFlags uint8
}

type MintSyscoinType struct {
	Allocation AssetAllocationType
    TxValue []byte
    TxParentNodes []byte
    TxRoot []byte
    TxPath []byte
    ReceiptValue []byte
    ReceiptParentNodes []byte
    ReceiptRoot []byte
    ReceiptPath []byte
    BlockNumber uint32
    BridgeTransferId uint32
}

type SyscoinBurnToEthereumType struct {
	Allocation AssetAllocationType
	EthAddress []byte
}

func PutUint(w io.Writer, n uint64) error {
    tmp := make([]uint8, 10)
    var len uint8=0
    for  {
		var mask uint64
		if len > 0 {
			mask = 0x80
		}
		tmpI := (n & 0x7F) | mask
		tmp[len] = uint8(tmpI)
        if n <= 0x7F {
			break
		}
        n = (n >> 7) - 1
        len++
	}
	for len > 0 {
		len--
		err := binarySerializer.PutUint8(w, tmp[len])
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadUint(r io.Reader) (uint64, error) {
    var n uint64 = 0
    for {
		chData, err := binarySerializer.Uint8(r)
		if err != nil {
			return 0, err
		}
        n = (n << 7) | (uint64(chData) & 0x7F)
        if (chData & 0x80) > 0 {
            n++
        } else {
            return n, nil
        }
	}
	return n, nil
}
// Amount compression:
// * If the amount is 0, output 0
// * first, divide the amount (in base units) by the largest power of 10 possible; call the exponent e (e is max 9)
// * if e<9, the last digit of the resulting number cannot be 0; store it as d, and drop it (divide by 10)
//   * call the result n
//   * output 1 + 10*(9*n + d - 1) + e
// * if e==9, we only know the resulting number is not zero, so output 1 + 10*(n - 1) + 9
// (this is decodable, as d is in [1-9] and e is in [0-9])

func CompressAmount(n uint64) uint64 {
    if n == 0 {
		return 0
	}
    var e int = 0;
    for ((n % 10) == 0) && e < 9 {
        n /= 10
        e++
    }
    if e < 9 {
        var d int = int(n % 10)
        n /= 10
        return 1 + (n*9 + uint64(d) - 1)*10 + uint64(e)
    } else {
        return 1 + (n - 1)*10 + 9
    }
}

func DecompressAmount(x uint64) uint64 {
    // x = 0  OR  x = 1+10*(9*n + d - 1) + e  OR  x = 1+10*(n - 1) + 9
    if x == 0 {
		return 0
	}
    x--
    // x = 10*(9*n + d - 1) + e
    var e int = int(x % 10)
    x /= 10
    var n uint64 = 0
    if e < 9 {
        // x = 9*n + d - 1
        var d int = int(x % 9) + 1
        x /= 9
        // x = n
        n = x*10 + uint64(d)
    } else {
        n = x+1
    }
    for e > 0 {
        n *= 10
        e--
    }
    return n
}

func (a *NotaryDetailsType) Deserialize(r io.Reader) error {
	var err error
	a.EndPoint, err = ReadVarBytes(r, 0, MAX_VALUE_LENGTH, "EndPoint")
	if err != nil {
		return err
	}
	a.InstantTransfers, err = binarySerializer.Uint8(r)
	if err != nil {
		return err
	}
	a.HDRequired, err = binarySerializer.Uint8(r)
	if err != nil {
		return err
	}
	return nil
}
func (a *AuxFeesType) Deserialize(r io.Reader) error {
	valueSat, err := ReadUint(r)
	if err != nil {
		return err
	}
	a.Bound = int64(DecompressAmount(valueSat))
	a.Percent, err = binarySerializer.Uint16(r, binary.LittleEndian)
	if err != nil {
		return err
	}
	return nil
}
func (a *AuxFeeDetailsType) Deserialize(r io.Reader) error {
	numAuxFees, err := ReadVarInt(r, 0)
	if err != nil {
		return err
	}
	a.AuxFees = make([]AuxFeesType, numAuxFees)
	for i := 0; i < int(numAuxFees); i++ {
		err = a.AuxFees[i].Deserialize(r)
		if err != nil {
			return err
		}
	}
	return nil
}
func (a *AssetType) Deserialize(r io.Reader) error {
	err := a.Allocation.Deserialize(r)
	if err != nil {
		return err
	}
	a.Precision, err = binarySerializer.Uint8(r)
	symbol, err = ReadVarBytes(r, 0, MAX_SYMBOL_SIZE, "Symbol")
	if err != nil {
		return err
	}
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(symbol)))
	n, err := base64.StdEncoding.Decode(base64Text, []byte(symbol))
	if err != nil {
		return err
	}
	a.Symbol = string(base64Text[:n])
	a.UpdateFlags, err = binarySerializer.Uint8(r)
	if err != nil {
		return err
	}
	if (a.UpdateFlags & ASSET_UPDATE_CONTRACT) != 0 {
		a.Contract, err = ReadVarBytes(r, 0, MAX_GUID_LENGTH, "Contract")
		if err != nil {
			return err
		}
		a.PrevContract, err = ReadVarBytes(r, 0, MAX_GUID_LENGTH, "PrevContract")
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_DATA) != 0 {
		a.PubData, err = ReadVarBytes(r, 0, MAX_VALUE_LENGTH, "PubData")
		if err != nil {
			return err
		}
		a.PrevPubData, err = ReadVarBytes(r, 0, MAX_VALUE_LENGTH, "PrevPubData")
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_SUPPLY) != 0 {
		valueSat, err := ReadUint(r)
		if err != nil {
			return err
		}
		a.Balance = int64(DecompressAmount(valueSat))

		valueSat, err = ReadUint(r)
		if err != nil {
			return err
		}
		a.TotalSupply = int64(DecompressAmount(valueSat))

		valueSat, err = ReadUint(r)
		if err != nil {
			return err
		}
		a.MaxSupply = int64(DecompressAmount(valueSat))
	}
	if (a.UpdateFlags & ASSET_UPDATE_NOTARY_KEY) != 0 {
		a.NotaryKeyID, err = ReadVarBytes(r, 0, MAX_GUID_LENGTH, "NotaryKeyID")
		if err != nil {
			return err
		}
		a.PrevNotaryKeyID, err = ReadVarBytes(r, 0, MAX_GUID_LENGTH, "PrevNotaryKeyID")
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_NOTARY_DETAILS) != 0 {
		err = a.NotaryDetails.Deserialize(r)
		if err != nil {
			return err
		}
		err = a.PrevNotaryDetails.Deserialize(r)
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_AUXFEE_KEY) != 0 {
		a.AuxFeeKeyID, err = ReadVarBytes(r, 0, MAX_GUID_LENGTH, "AuxFeeKeyID")
		if err != nil {
			return err
		}
		a.PrevAuxFeeKeyID, err = ReadVarBytes(r, 0, MAX_GUID_LENGTH, "PrevAuxFeeKeyID")
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_AUXFEE_DETAILS) != 0 {
		err = a.AuxFeeDetails.Deserialize(r)
		if err != nil {
			return err
		}
		err = a.PrevAuxFeeDetails.Deserialize(r)
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_CAPABILITYFLAGS) != 0 {
		a.UpdateCapabilityFlags, err = binarySerializer.Uint8(r)
		if err != nil {
			return err
		}
		a.PrevUpdateCapabilityFlags, err = binarySerializer.Uint8(r)
		if err != nil {
			return err
		}
	}

	return nil
}


func (a *NotaryDetailsType) Serialize(w io.Writer) error {
	err := WriteVarBytes(w, 0, a.EndPoint)
	if err != nil {
		return err
	}
	err = binarySerializer.PutUint8(w, a.InstantTransfers)
	if err != nil {
		return err
	}
	err = binarySerializer.PutUint8(w, a.HDRequired)
	if err != nil {
		return err
	}
	return nil
}
func (a *AuxFeesType) Serialize(w io.Writer) error {
	err := PutUint(w, CompressAmount(uint64(a.Bound)))
	if err != nil {
		return err
	}
	err = binarySerializer.PutUint16(w, binary.LittleEndian, a.Percent)
	if err != nil {
		return err
	}
	return nil
}
func (a *AuxFeeDetailsType) Serialize(w io.Writer) error {
	lenAuxFees := len(a.AuxFees)
	err := WriteVarInt(w, 0, uint64(lenAuxFees))
	if err != nil {
		return err
	}
	for i := 0; i < lenAuxFees; i++ {
		err = a.AuxFees[i].Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AssetType) Serialize(w io.Writer) error {
	err := a.Allocation.Serialize(w)
	if err != nil {
		return err
	}
	err = binarySerializer.PutUint8(w, a.Precision)
	if err != nil {
		return err
	}
	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(a.Symbol)))
	base64.StdEncoding.Encode(base64Text, []byte(a.Symbol))
	err = WriteVarBytes(w, 0, base64Text)
	if err != nil {
		return err
	}
	err = binarySerializer.PutUint8(w, a.UpdateFlags)
	if err != nil {
		return err
	}
	if (a.UpdateFlags & ASSET_UPDATE_CONTRACT) != 0 {
		err = WriteVarBytes(w, 0, a.Contract)
		if err != nil {
			return err
		}
		err = WriteVarBytes(w, 0, a.PrevContract)
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_DATA) != 0 {
		err = WriteVarBytes(w, 0, a.PubData)
		if err != nil {
			return err
		}
		err = WriteVarBytes(w, 0, a.PrevPubData)
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_SUPPLY) != 0 {
		err = PutUint(w, CompressAmount(uint64(a.Balance)))
		if err != nil {
			return err
		}
		err = PutUint(w, CompressAmount(uint64(a.TotalSupply)))
		if err != nil {
			return err
		}
		err = PutUint(w, CompressAmount(uint64(a.MaxSupply)))
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_NOTARY_KEY) != 0 {
		err = WriteVarBytes(w, 0, a.NotaryKeyID)
		if err != nil {
			return err
		}
		err = WriteVarBytes(w, 0, a.PrevNotaryKeyID)
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_NOTARY_DETAILS) != 0 {
		err = a.NotaryDetails.Serialize(w)
		if err != nil {
			return err
		}
		err = a.PrevNotaryDetails.Serialize(w)
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_AUXFEE_KEY) != 0 {
		err = WriteVarBytes(w, 0, a.AuxFeeKeyID)
		if err != nil {
			return err
		}
		err = WriteVarBytes(w, 0, a.PrevAuxFeeKeyID)
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_AUXFEE_DETAILS) != 0 {
		err = a.AuxFeeDetails.Serialize(w)
		if err != nil {
			return err
		}
		err = a.PrevAuxFeeDetails.Serialize(w)
		if err != nil {
			return err
		}
	}
	if (a.UpdateFlags & ASSET_UPDATE_CAPABILITYFLAGS) != 0 {
		err = binarySerializer.PutUint8(w, a.UpdateCapabilityFlags)
		if err != nil {
			return err
		}
		err = binarySerializer.PutUint8(w, a.PrevUpdateCapabilityFlags)
		if err != nil {
			return err
		}
	}

	return nil
}


func (a *AssetAllocationType) Deserialize(r io.Reader) error {
	numAssets, err := ReadVarInt(r, 0)
	if err != nil {
		return err
	}
	a.VoutAssets = make([]AssetOutType, numAssets)
	for i := 0; i < int(numAssets); i++ {
		err = a.VoutAssets[i].Deserialize(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AssetAllocationType) Serialize(w io.Writer) error {
	lenAssets := len(a.VoutAssets)
	err := WriteVarInt(w, 0, uint64(lenAssets))
	if err != nil {
		return err
	}
	for i := 0; i < lenAssets; i++ {
		err = a.VoutAssets[i].Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AssetOutValueType) Serialize(w io.Writer) error {
	err := WriteVarInt(w, 0, uint64(a.N))
	if err != nil {
		return err
	}
	err = PutUint(w, CompressAmount(uint64(a.ValueSat)))
	if err != nil {
		return err
	}
	return nil
}

func (a *AssetOutValueType) Deserialize(r io.Reader) error {
	n, err := ReadVarInt(r, 0)
	if err != nil {
		return err
	}
	a.N = uint32(n)
	valueSat, err := ReadUint(r)
	if err != nil {
		return err
	}
	a.ValueSat = int64(DecompressAmount(valueSat))
	return nil
}

func (a *AssetOutType) Serialize(w io.Writer) error {
	err := writeElement(w, a.AssetGuid)
	if err != nil {
		return err
	}
	lenValues := len(a.Values)
	err = WriteVarInt(w, 0, uint64(lenValues))
	if err != nil {
		return err
	}
	for i := 0; i < lenValues; i++ {
		err = a.Values[i].Serialize(w)
		if err != nil {
			return err
		}
	}
	err = WriteVarBytes(w, 0, a.NotarySig)
	if err != nil {
		return err
	}
	return nil
}

func (a *AssetOutType) Deserialize(r io.Reader) error {
	err := readElement(r, &a.AssetGuid)
	if err != nil {
		return err
	}
	numOutputs, err := ReadVarInt(r, 0)
	if err != nil {
		return err
	}
	a.Values = make([]AssetOutValueType, numOutputs)
	for i := 0; i < int(numOutputs); i++ {
		err = a.Values[i].Deserialize(r)
		if err != nil {
			return err
		}
	}
	a.NotarySig, err = ReadVarBytes(r, 0, MAX_SIG_SIZE, "NotarySig")
	if err != nil {
		return err
	}
	return nil
}


func (a *MintSyscoinType) Deserialize(r io.Reader) error {
	err := a.Allocation.Deserialize(r)
	if err != nil {
		return err
	}
	err = readElement(r, &a.BridgeTransferId)
	if err != nil {
		return err
	}
	err = readElement(r, &a.BlockNumber)
	if err != nil {
		return err
	}
	a.TxValue, err = ReadVarBytes(r, 0, MAX_RLP_SIZE, "TxValue")
	if err != nil {
		return err
	}
	a.TxParentNodes, err = ReadVarBytes(r, 0, MAX_RLP_SIZE, "TxParentNodes")
	if err != nil {
		return err
	}
	a.TxRoot, err = ReadVarBytes(r, 0, MAX_RLP_SIZE, "TxRoot")
	if err != nil {
		return err
	}
	a.TxPath, err = ReadVarBytes(r, 0, MAX_RLP_SIZE, "TxPath")
	if err != nil {
		return err
	}
	a.ReceiptValue, err = ReadVarBytes(r, 0, MAX_RLP_SIZE, "ReceiptValue")
	if err != nil {
		return err
	}
	a.ReceiptParentNodes, err = ReadVarBytes(r, 0, MAX_RLP_SIZE, "ReceiptParentNodes")
	if err != nil {
		return err
	}
	a.ReceiptRoot, err = ReadVarBytes(r, 0, MAX_RLP_SIZE, "ReceiptRoot")
	if err != nil {
		return err
	}
	a.ReceiptPath, err = ReadVarBytes(r, 0, MAX_RLP_SIZE, "ReceiptPath")
	if err != nil {
		return err
	}
	return nil
}

func (a *MintSyscoinType) Serialize(w io.Writer) error {
	err := a.Allocation.Serialize(w)
	if err != nil {
		return err
	}
	err = writeElement(w, a.BridgeTransferId)
	if err != nil {
		return err
	}
	err = writeElement(w, a.BlockNumber)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.TxValue)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.TxParentNodes)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.TxRoot)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.TxPath)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.ReceiptValue)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.ReceiptParentNodes)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.ReceiptRoot)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.ReceiptPath)
	if err != nil {
		return err
	}
	return nil
}

func (a *SyscoinBurnToEthereumType) Deserialize(r io.Reader) error {
	err := a.Allocation.Deserialize(r)
	if err != nil {
		return err
	}
	a.EthAddress, err = ReadVarBytes(r, 0, MAX_GUID_LENGTH, "ethAddress")
	if err != nil {
		return err
	}
	return nil
}

func (a *SyscoinBurnToEthereumType) Serialize(w io.Writer) error {
	err := a.Allocation.Serialize(w)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.EthAddress)
	if err != nil {
		return err
	}
	return nil
}