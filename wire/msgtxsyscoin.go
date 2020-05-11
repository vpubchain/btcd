// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"io"
	"errors"
)

type AssetOutType struct {
	N uint32
	ValueSat int64
}
type AssetAllocationType struct {
	VoutAssets map[uint32][]AssetOutType
}

type AssetType struct {
	Allocation AssetAllocationType
	Contract []byte
	PrevContract  []byte
	Symbol string
	PubData []byte
	PrevPubData []byte
	Balance int64
	TotalSupply int64
	MaxSupply int64
	Precision uint8
	UpdateFlags uint8
	PrevUpdateFlags uint8
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

func PutUint(w io.Writer, n uint64) error
{
    tmp := make([]uint8, (len(n)*8+6)/7)
    var len int=0
    for  {
        tmp[len] = (n & 0x7F) | (len ? 0x80 : 0x00)
        if (n <= 0x7F)
            break
        n = (n >> 7) - 1
        len++
	}
	for n = len; n >= 0; n-- {
		err = binarySerializer.PutUint8(w, tmp[len])
		if err != nil {
			return err
		}
	}
}

func ReadUint(r io.Reader) (uint64, error)
{
    var n uint64 = 0
    for {
		chData, err := binarySerializer.Uint8(r)
		if err != nil {
			return 0, err
		}
        n = (n << 7) | (chData & 0x7F)
        if (chData & 0x80) {
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


func (a *AssetType) Deserialize(r io.Reader) error {
	var err error
	err = a.Allocation.Deserialize(r)
	if err != nil {
		return err
	}
	a.Precision, err = binarySerializer.Uint8(r)
	a.Contract, err = ReadVarBytes(r, 0, 20, "Contract")
	if err != nil {
		return err
	}
	a.PubData, err = ReadVarBytes(r, 0, 512, "PubData")
	if err != nil {
		return err
	}
	symbol, err := ReadVarBytes(r, 0, 8, "Symbol")
	if err != nil {
		return err
	}
	a.Symbol = string(symbol)
	a.UpdateFlags, err = binarySerializer.Uint8(r)
	if err != nil {
		return err
	}
	a.PrevContract, err = ReadVarBytes(r, 0, 20, "PrevContract")
	if err != nil {
		return err
	}
	a.PrevPubData, err = ReadVarBytes(r, 0, 512, "PrevPubData")
	if err != nil {
		return err
	}
	a.PrevUpdateFlags, err = binarySerializer.Uint8(r)
	if err != nil {
		return err
	}
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

	return nil
}

func (a *AssetType) Serialize(w io.Writer) error {
	var err error
	err = a.Allocation.Serialize(w)
	if err != nil {
		return err
	}
	err = binarySerializer.PutUint8(w, a.Precision)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.Contract)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.PubData)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, ([]byte)(a.Symbol))
	if err != nil {
		return err
	}
	err = binarySerializer.PutUint8(w, a.UpdateFlags)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.PrevContract)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.PrevPubData)
	if err != nil {
		return err
	}
	err = binarySerializer.PutUint8(w, a.PrevUpdateFlags)
	if err != nil {
		return err
	}
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
	return nil
}


func (a *AssetAllocationType) Deserialize(r io.Reader) error {
	numAssets, err := ReadVarInt(r, 0)
	if err != nil {
		return err
	}
	a.VoutAssets = make(map[uint32][]AssetOutType, numAssets)
	for i := 0; i < int(numAssets); i++ {
		var assetGuid uint32
		err = readElement(r, &assetGuid)
		if err != nil {
			return err
		}
		numOutputs, err := ReadVarInt(r, 0)
		if err != nil {
			return err
		}
		assetOutArray, ok := a.VoutAssets[assetGuid]
		if !ok {
			assetOutArray = make([]AssetOutType, numOutputs)
			a.VoutAssets[assetGuid] = assetOutArray
		}
		for j := 0; j < int(numOutputs); j++ {
			err = assetOutArray[j].Deserialize(r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *AssetAllocationType) Serialize(w io.Writer) error {
	err := WriteVarInt(w, 0, uint64(len(a.VoutAssets)))
	if err != nil {
		return err
	}
	for k, v := range a.VoutAssets {
		err = writeElement(w, k)
		if err != nil {
			return err
		}
		err = WriteVarInt(w, 0, uint64(len(v)))
		if err != nil {
			return err
		}
		for _,voutAsset := range v {
			err = voutAsset.Serialize(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *AssetOutType) Serialize(w io.Writer) error {
	var err error
	err = PutUint(buf, uint64(a.N))
	if err != nil {
		return err
	}
	err = PutUint(buf, CompressAmount(uint64(a.ValueSat)))
	if err != nil {
		return err
	}
	return nil
}

func (a *AssetOutType) Deserialize(r io.Reader) error {
	var err error
	n, err := ReadUint(r)
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



func (a *MintSyscoinType) Deserialize(r io.Reader) error {
	var err error
	err = a.Allocation.Deserialize(r)
	if err != nil {
		return err
	}
	bridgeTransferId, err := ReadVarInt(r, 0)
	a.BridgeTransferId = uint32(bridgeTransferId)
	if err != nil {
		return err
	}
	blockNumber, err := ReadVarInt(r, 0)
	a.BlockNumber = uint32(blockNumber)
	if err != nil {
		return err
	}
	a.TxValue, err = ReadVarBytes(r, 0, 4096, "TxValue")
	if err != nil {
		return err
	}
	a.TxParentNodes, err = ReadVarBytes(r, 0, 4096, "TxParentNodes")
	if err != nil {
		return err
	}
	a.TxRoot, err = ReadVarBytes(r, 0, 4096, "TxRoot")
	if err != nil {
		return err
	}
	a.TxPath, err = ReadVarBytes(r, 0, 4096, "TxPath")
	if err != nil {
		return err
	}
	a.ReceiptValue, err = ReadVarBytes(r, 0, 4096, "ReceiptValue")
	if err != nil {
		return err
	}
	a.ReceiptParentNodes, err = ReadVarBytes(r, 0, 4096, "ReceiptParentNodes")
	if err != nil {
		return err
	}
	a.ReceiptRoot, err = ReadVarBytes(r, 0, 4096, "ReceiptRoot")
	if err != nil {
		return err
	}
	a.ReceiptPath, err = ReadVarBytes(r, 0, 4096, "ReceiptPath")
	if err != nil {
		return err
	}
	return nil
}

func (a *MintSyscoinType) Serialize(w io.Writer) error {
	var err error
	err = a.Allocation.Serialize(w)
	if err != nil {
		return err
	}
	err = WriteVarInt(w, 0, uint64(a.BridgeTransferId))
	if err != nil {
		return err
	}
	err = WriteVarInt(w, 0, uint64(a.BlockNumber))
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
	var err error
	err = a.Allocation.Deserialize(r)
	if err != nil {
		return err
	}
	a.EthAddress, err = ReadVarBytes(r, 0, 20, "ethAddress")
	if err != nil {
		return err
	}
	return nil
}

func (a *SyscoinBurnToEthereumType) Serialize(w io.Writer) error {
	var err error
	err = a.Allocation.Serialize(w)
	if err != nil {
		return err
	}
	err = WriteVarBytes(w, 0, a.EthAddress)
	if err != nil {
		return err
	}
	return nil
}