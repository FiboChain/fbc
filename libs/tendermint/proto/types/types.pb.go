// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/types/types.proto

package types

import (
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/timestamp"
	merkle "github.com/FiboChain/fbc/libs/tendermint/proto/crypto/merkle"
	bits "github.com/FiboChain/fbc/libs/tendermint/proto/libs/bits"
	version "github.com/FiboChain/fbc/libs/tendermint/proto/version"
	math "math"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// BlockIdFlag indicates which BlcokID the signature is for
type BlockIDFlag int32

const (
	BLOCKD_ID_FLAG_UNKNOWN BlockIDFlag = 0
	BlockIDFlagAbsent      BlockIDFlag = 1
	BlockIDFlagCommit      BlockIDFlag = 2
	BlockIDFlagNil         BlockIDFlag = 3
)

var BlockIDFlag_name = map[int32]string{
	0: "BLOCKD_ID_FLAG_UNKNOWN",
	1: "BLOCK_ID_FLAG_ABSENT",
	2: "BLOCK_ID_FLAG_COMMIT",
	3: "BLOCK_ID_FLAG_NIL",
}

var BlockIDFlag_value = map[string]int32{
	"BLOCKD_ID_FLAG_UNKNOWN": 0,
	"BLOCK_ID_FLAG_ABSENT":   1,
	"BLOCK_ID_FLAG_COMMIT":   2,
	"BLOCK_ID_FLAG_NIL":      3,
}

func (BlockIDFlag) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{0}
}

// SignedMsgType is a type of signed message in the consensus.
type SignedMsgType int32

const (
	SIGNED_MSG_TYPE_UNKNOWN SignedMsgType = 0
	PrevoteType             SignedMsgType = 1
	PrecommitType           SignedMsgType = 2
	ProposalType            SignedMsgType = 3
)

var SignedMsgType_name = map[int32]string{
	0: "SIGNED_MSG_TYPE_UNKNOWN",
	1: "PREVOTE_TYPE",
	2: "PRECOMMIT_TYPE",
	3: "PROPOSAL_TYPE",
}

var SignedMsgType_value = map[string]int32{
	"SIGNED_MSG_TYPE_UNKNOWN": 0,
	"PREVOTE_TYPE":            1,
	"PRECOMMIT_TYPE":          2,
	"PROPOSAL_TYPE":           3,
}

func (SignedMsgType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{1}
}

// PartsetHeader
type PartSetHeader struct {
	Total                int64    `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	Hash                 []byte   `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PartSetHeader) Reset()         { *m = PartSetHeader{} }
func (m *PartSetHeader) String() string { return proto.CompactTextString(m) }
func (*PartSetHeader) ProtoMessage()    {}
func (*PartSetHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{0}
}
func (m *PartSetHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PartSetHeader.Unmarshal(m, b)
}
func (m *PartSetHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PartSetHeader.Marshal(b, m, deterministic)
}
func (m *PartSetHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PartSetHeader.Merge(m, src)
}
func (m *PartSetHeader) XXX_Size() int {
	return xxx_messageInfo_PartSetHeader.Size(m)
}
func (m *PartSetHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_PartSetHeader.DiscardUnknown(m)
}

var xxx_messageInfo_PartSetHeader proto.InternalMessageInfo

func (m *PartSetHeader) GetTotal() int64 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *PartSetHeader) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

type Part struct {
	Index                uint32             `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Bytes                []byte             `protobuf:"bytes,2,opt,name=bytes,proto3" json:"bytes,omitempty"`
	Proof                merkle.SimpleProof `protobuf:"bytes,3,opt,name=proof,proto3" json:"proof"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *Part) Reset()         { *m = Part{} }
func (m *Part) String() string { return proto.CompactTextString(m) }
func (*Part) ProtoMessage()    {}
func (*Part) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{1}
}
func (m *Part) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Part.Unmarshal(m, b)
}
func (m *Part) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Part.Marshal(b, m, deterministic)
}
func (m *Part) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Part.Merge(m, src)
}
func (m *Part) XXX_Size() int {
	return xxx_messageInfo_Part.Size(m)
}
func (m *Part) XXX_DiscardUnknown() {
	xxx_messageInfo_Part.DiscardUnknown(m)
}

var xxx_messageInfo_Part proto.InternalMessageInfo

func (m *Part) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *Part) GetBytes() []byte {
	if m != nil {
		return m.Bytes
	}
	return nil
}

func (m *Part) GetProof() merkle.SimpleProof {
	if m != nil {
		return m.Proof
	}
	return merkle.SimpleProof{}
}

// BlockID
type BlockID struct {
	Hash                 []byte        `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	PartsHeader          PartSetHeader `protobuf:"bytes,2,opt,name=parts_header,json=partsHeader,proto3" json:"parts_header"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *BlockID) Reset()         { *m = BlockID{} }
func (m *BlockID) String() string { return proto.CompactTextString(m) }
func (*BlockID) ProtoMessage()    {}
func (*BlockID) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{2}
}
func (m *BlockID) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockID.Unmarshal(m, b)
}
func (m *BlockID) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockID.Marshal(b, m, deterministic)
}
func (m *BlockID) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockID.Merge(m, src)
}
func (m *BlockID) XXX_Size() int {
	return xxx_messageInfo_BlockID.Size(m)
}
func (m *BlockID) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockID.DiscardUnknown(m)
}

var xxx_messageInfo_BlockID proto.InternalMessageInfo

func (m *BlockID) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *BlockID) GetPartsHeader() PartSetHeader {
	if m != nil {
		return m.PartsHeader
	}
	return PartSetHeader{}
}

// Header defines the structure of a Tendermint block header.
type Header struct {
	// basic block info
	Version version.Consensus `protobuf:"bytes,1,opt,name=version,proto3" json:"version"`
	ChainID string            `protobuf:"bytes,2,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	Height  int64             `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	Time    time.Time         `protobuf:"bytes,4,opt,name=time,proto3,stdtime" json:"time"`
	// prev block info
	LastBlockID BlockID `protobuf:"bytes,5,opt,name=last_block_id,json=lastBlockId,proto3" json:"last_block_id"`
	// hashes of block data
	LastCommitHash []byte `protobuf:"bytes,6,opt,name=last_commit_hash,json=lastCommitHash,proto3" json:"last_commit_hash,omitempty"`
	DataHash       []byte `protobuf:"bytes,7,opt,name=data_hash,json=dataHash,proto3" json:"data_hash,omitempty"`
	// hashes from the app output from the prev block
	ValidatorsHash     []byte `protobuf:"bytes,8,opt,name=validators_hash,json=validatorsHash,proto3" json:"validators_hash,omitempty"`
	NextValidatorsHash []byte `protobuf:"bytes,9,opt,name=next_validators_hash,json=nextValidatorsHash,proto3" json:"next_validators_hash,omitempty"`
	ConsensusHash      []byte `protobuf:"bytes,10,opt,name=consensus_hash,json=consensusHash,proto3" json:"consensus_hash,omitempty"`
	AppHash            []byte `protobuf:"bytes,11,opt,name=app_hash,json=appHash,proto3" json:"app_hash,omitempty"`
	LastResultsHash    []byte `protobuf:"bytes,12,opt,name=last_results_hash,json=lastResultsHash,proto3" json:"last_results_hash,omitempty"`
	// consensus info
	EvidenceHash         []byte   `protobuf:"bytes,13,opt,name=evidence_hash,json=evidenceHash,proto3" json:"evidence_hash,omitempty"`
	ProposerAddress      []byte   `protobuf:"bytes,14,opt,name=proposer_address,json=proposerAddress,proto3" json:"proposer_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Header) Reset()         { *m = Header{} }
func (m *Header) String() string { return proto.CompactTextString(m) }
func (*Header) ProtoMessage()    {}
func (*Header) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{3}
}
func (m *Header) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Header.Unmarshal(m, b)
}
func (m *Header) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Header.Marshal(b, m, deterministic)
}
func (m *Header) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Header.Merge(m, src)
}
func (m *Header) XXX_Size() int {
	return xxx_messageInfo_Header.Size(m)
}
func (m *Header) XXX_DiscardUnknown() {
	xxx_messageInfo_Header.DiscardUnknown(m)
}

var xxx_messageInfo_Header proto.InternalMessageInfo

func (m *Header) GetVersion() version.Consensus {
	if m != nil {
		return m.Version
	}
	return version.Consensus{}
}

func (m *Header) GetChainID() string {
	if m != nil {
		return m.ChainID
	}
	return ""
}

func (m *Header) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *Header) GetTime() time.Time {
	if m != nil {
		return m.Time
	}
	return time.Time{}
}

func (m *Header) GetLastBlockID() BlockID {
	if m != nil {
		return m.LastBlockID
	}
	return BlockID{}
}

func (m *Header) GetLastCommitHash() []byte {
	if m != nil {
		return m.LastCommitHash
	}
	return nil
}

func (m *Header) GetDataHash() []byte {
	if m != nil {
		return m.DataHash
	}
	return nil
}

func (m *Header) GetValidatorsHash() []byte {
	if m != nil {
		return m.ValidatorsHash
	}
	return nil
}

func (m *Header) GetNextValidatorsHash() []byte {
	if m != nil {
		return m.NextValidatorsHash
	}
	return nil
}

func (m *Header) GetConsensusHash() []byte {
	if m != nil {
		return m.ConsensusHash
	}
	return nil
}

func (m *Header) GetAppHash() []byte {
	if m != nil {
		return m.AppHash
	}
	return nil
}

func (m *Header) GetLastResultsHash() []byte {
	if m != nil {
		return m.LastResultsHash
	}
	return nil
}

func (m *Header) GetEvidenceHash() []byte {
	if m != nil {
		return m.EvidenceHash
	}
	return nil
}

func (m *Header) GetProposerAddress() []byte {
	if m != nil {
		return m.ProposerAddress
	}
	return nil
}

// Data contains the set of transactions included in the block
type Data struct {
	// Txs that will be applied by state @ block.Height+1.
	// NOTE: not all txs here are valid.  We're just agreeing on the order first.
	// This means that block.AppHash does not include these txs.
	Txs [][]byte `protobuf:"bytes,1,rep,name=txs,proto3" json:"txs,omitempty"`
	// Volatile
	Hash                 []byte   `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Data) Reset()         { *m = Data{} }
func (m *Data) String() string { return proto.CompactTextString(m) }
func (*Data) ProtoMessage()    {}
func (*Data) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{4}
}
func (m *Data) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Data.Unmarshal(m, b)
}
func (m *Data) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Data.Marshal(b, m, deterministic)
}
func (m *Data) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Data.Merge(m, src)
}
func (m *Data) XXX_Size() int {
	return xxx_messageInfo_Data.Size(m)
}
func (m *Data) XXX_DiscardUnknown() {
	xxx_messageInfo_Data.DiscardUnknown(m)
}

var xxx_messageInfo_Data proto.InternalMessageInfo

func (m *Data) GetTxs() [][]byte {
	if m != nil {
		return m.Txs
	}
	return nil
}

func (m *Data) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

// Vote represents a prevote, precommit, or commit vote from validators for
// consensus.
type Vote struct {
	Type                 SignedMsgType `protobuf:"varint,1,opt,name=type,proto3,enum=tendermint.proto.types.SignedMsgType" json:"type,omitempty"`
	Height               int64         `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	Round                int64         `protobuf:"varint,3,opt,name=round,proto3" json:"round,omitempty"`
	BlockID              BlockID       `protobuf:"bytes,4,opt,name=block_id,json=blockId,proto3" json:"block_id"`
	Timestamp            time.Time     `protobuf:"bytes,5,opt,name=timestamp,proto3,stdtime" json:"timestamp"`
	ValidatorAddress     []byte        `protobuf:"bytes,6,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	ValidatorIndex       int64         `protobuf:"varint,7,opt,name=validator_index,json=validatorIndex,proto3" json:"validator_index,omitempty"`
	Signature            []byte        `protobuf:"bytes,8,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Vote) Reset()         { *m = Vote{} }
func (m *Vote) String() string { return proto.CompactTextString(m) }
func (*Vote) ProtoMessage()    {}
func (*Vote) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{5}
}
func (m *Vote) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Vote.Unmarshal(m, b)
}
func (m *Vote) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Vote.Marshal(b, m, deterministic)
}
func (m *Vote) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Vote.Merge(m, src)
}
func (m *Vote) XXX_Size() int {
	return xxx_messageInfo_Vote.Size(m)
}
func (m *Vote) XXX_DiscardUnknown() {
	xxx_messageInfo_Vote.DiscardUnknown(m)
}

var xxx_messageInfo_Vote proto.InternalMessageInfo

func (m *Vote) GetType() SignedMsgType {
	if m != nil {
		return m.Type
	}
	return SIGNED_MSG_TYPE_UNKNOWN
}

func (m *Vote) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *Vote) GetRound() int64 {
	if m != nil {
		return m.Round
	}
	return 0
}

func (m *Vote) GetBlockID() BlockID {
	if m != nil {
		return m.BlockID
	}
	return BlockID{}
}

func (m *Vote) GetTimestamp() time.Time {
	if m != nil {
		return m.Timestamp
	}
	return time.Time{}
}

func (m *Vote) GetValidatorAddress() []byte {
	if m != nil {
		return m.ValidatorAddress
	}
	return nil
}

func (m *Vote) GetValidatorIndex() int64 {
	if m != nil {
		return m.ValidatorIndex
	}
	return 0
}

func (m *Vote) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

// Commit contains the evidence that a block was committed by a set of validators.
type Commit struct {
	Height               int64          `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Round                int32          `protobuf:"varint,2,opt,name=round,proto3" json:"round,omitempty"`
	BlockID              BlockID        `protobuf:"bytes,3,opt,name=block_id,json=blockId,proto3" json:"block_id"`
	Signatures           []CommitSig    `protobuf:"bytes,4,rep,name=signatures,proto3" json:"signatures"`
	Hash                 []byte         `protobuf:"bytes,5,opt,name=hash,proto3" json:"hash,omitempty"`
	BitArray             *bits.BitArray `protobuf:"bytes,6,opt,name=bit_array,json=bitArray,proto3" json:"bit_array,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Commit) Reset()         { *m = Commit{} }
func (m *Commit) String() string { return proto.CompactTextString(m) }
func (*Commit) ProtoMessage()    {}
func (*Commit) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{6}
}
func (m *Commit) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Commit.Unmarshal(m, b)
}
func (m *Commit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Commit.Marshal(b, m, deterministic)
}
func (m *Commit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Commit.Merge(m, src)
}
func (m *Commit) XXX_Size() int {
	return xxx_messageInfo_Commit.Size(m)
}
func (m *Commit) XXX_DiscardUnknown() {
	xxx_messageInfo_Commit.DiscardUnknown(m)
}

var xxx_messageInfo_Commit proto.InternalMessageInfo

func (m *Commit) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *Commit) GetRound() int32 {
	if m != nil {
		return m.Round
	}
	return 0
}

func (m *Commit) GetBlockID() BlockID {
	if m != nil {
		return m.BlockID
	}
	return BlockID{}
}

func (m *Commit) GetSignatures() []CommitSig {
	if m != nil {
		return m.Signatures
	}
	return nil
}

func (m *Commit) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Commit) GetBitArray() *bits.BitArray {
	if m != nil {
		return m.BitArray
	}
	return nil
}

// CommitSig is a part of the Vote included in a Commit.
type CommitSig struct {
	BlockIdFlag          BlockIDFlag `protobuf:"varint,1,opt,name=block_id_flag,json=blockIdFlag,proto3,enum=tendermint.proto.types.BlockIDFlag" json:"block_id_flag,omitempty"`
	ValidatorAddress     []byte      `protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	Timestamp            time.Time   `protobuf:"bytes,3,opt,name=timestamp,proto3,stdtime" json:"timestamp"`
	Signature            []byte      `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *CommitSig) Reset()         { *m = CommitSig{} }
func (m *CommitSig) String() string { return proto.CompactTextString(m) }
func (*CommitSig) ProtoMessage()    {}
func (*CommitSig) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{7}
}
func (m *CommitSig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommitSig.Unmarshal(m, b)
}
func (m *CommitSig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommitSig.Marshal(b, m, deterministic)
}
func (m *CommitSig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommitSig.Merge(m, src)
}
func (m *CommitSig) XXX_Size() int {
	return xxx_messageInfo_CommitSig.Size(m)
}
func (m *CommitSig) XXX_DiscardUnknown() {
	xxx_messageInfo_CommitSig.DiscardUnknown(m)
}

var xxx_messageInfo_CommitSig proto.InternalMessageInfo

func (m *CommitSig) GetBlockIdFlag() BlockIDFlag {
	if m != nil {
		return m.BlockIdFlag
	}
	return BLOCKD_ID_FLAG_UNKNOWN
}

func (m *CommitSig) GetValidatorAddress() []byte {
	if m != nil {
		return m.ValidatorAddress
	}
	return nil
}

func (m *CommitSig) GetTimestamp() time.Time {
	if m != nil {
		return m.Timestamp
	}
	return time.Time{}
}

func (m *CommitSig) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

type Proposal struct {
	Type                 SignedMsgType `protobuf:"varint,1,opt,name=type,proto3,enum=tendermint.proto.types.SignedMsgType" json:"type,omitempty"`
	Height               int64         `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	Round                int32         `protobuf:"varint,3,opt,name=round,proto3" json:"round,omitempty"`
	PolRound             int32         `protobuf:"varint,4,opt,name=pol_round,json=polRound,proto3" json:"pol_round,omitempty"`
	BlockID              BlockID       `protobuf:"bytes,5,opt,name=block_id,json=blockId,proto3" json:"block_id"`
	Timestamp            time.Time     `protobuf:"bytes,6,opt,name=timestamp,proto3,stdtime" json:"timestamp"`
	Signature            []byte        `protobuf:"bytes,7,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Proposal) Reset()         { *m = Proposal{} }
func (m *Proposal) String() string { return proto.CompactTextString(m) }
func (*Proposal) ProtoMessage()    {}
func (*Proposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{8}
}
func (m *Proposal) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Proposal.Unmarshal(m, b)
}
func (m *Proposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Proposal.Marshal(b, m, deterministic)
}
func (m *Proposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Proposal.Merge(m, src)
}
func (m *Proposal) XXX_Size() int {
	return xxx_messageInfo_Proposal.Size(m)
}
func (m *Proposal) XXX_DiscardUnknown() {
	xxx_messageInfo_Proposal.DiscardUnknown(m)
}

var xxx_messageInfo_Proposal proto.InternalMessageInfo

func (m *Proposal) GetType() SignedMsgType {
	if m != nil {
		return m.Type
	}
	return SIGNED_MSG_TYPE_UNKNOWN
}

func (m *Proposal) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *Proposal) GetRound() int32 {
	if m != nil {
		return m.Round
	}
	return 0
}

func (m *Proposal) GetPolRound() int32 {
	if m != nil {
		return m.PolRound
	}
	return 0
}

func (m *Proposal) GetBlockID() BlockID {
	if m != nil {
		return m.BlockID
	}
	return BlockID{}
}

func (m *Proposal) GetTimestamp() time.Time {
	if m != nil {
		return m.Timestamp
	}
	return time.Time{}
}

func (m *Proposal) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

type SignedHeader struct {
	Header               *Header  `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Commit               *Commit  `protobuf:"bytes,2,opt,name=commit,proto3" json:"commit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SignedHeader) Reset()         { *m = SignedHeader{} }
func (m *SignedHeader) String() string { return proto.CompactTextString(m) }
func (*SignedHeader) ProtoMessage()    {}
func (*SignedHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{9}
}
func (m *SignedHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignedHeader.Unmarshal(m, b)
}
func (m *SignedHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignedHeader.Marshal(b, m, deterministic)
}
func (m *SignedHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignedHeader.Merge(m, src)
}
func (m *SignedHeader) XXX_Size() int {
	return xxx_messageInfo_SignedHeader.Size(m)
}
func (m *SignedHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_SignedHeader.DiscardUnknown(m)
}

var xxx_messageInfo_SignedHeader proto.InternalMessageInfo

func (m *SignedHeader) GetHeader() *Header {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *SignedHeader) GetCommit() *Commit {
	if m != nil {
		return m.Commit
	}
	return nil
}

type BlockMeta struct {
	BlockID              BlockID  `protobuf:"bytes,1,opt,name=block_id,json=blockId,proto3" json:"block_id"`
	BlockSize            int64    `protobuf:"varint,2,opt,name=block_size,json=blockSize,proto3" json:"block_size,omitempty"`
	Header               Header   `protobuf:"bytes,3,opt,name=header,proto3" json:"header"`
	NumTxs               int64    `protobuf:"varint,4,opt,name=num_txs,json=numTxs,proto3" json:"num_txs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockMeta) Reset()         { *m = BlockMeta{} }
func (m *BlockMeta) String() string { return proto.CompactTextString(m) }
func (*BlockMeta) ProtoMessage()    {}
func (*BlockMeta) Descriptor() ([]byte, []int) {
	return fileDescriptor_ff06f8095857fb18, []int{10}
}
func (m *BlockMeta) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockMeta.Unmarshal(m, b)
}
func (m *BlockMeta) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockMeta.Marshal(b, m, deterministic)
}
func (m *BlockMeta) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockMeta.Merge(m, src)
}
func (m *BlockMeta) XXX_Size() int {
	return xxx_messageInfo_BlockMeta.Size(m)
}
func (m *BlockMeta) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockMeta.DiscardUnknown(m)
}

var xxx_messageInfo_BlockMeta proto.InternalMessageInfo

func (m *BlockMeta) GetBlockID() BlockID {
	if m != nil {
		return m.BlockID
	}
	return BlockID{}
}

func (m *BlockMeta) GetBlockSize() int64 {
	if m != nil {
		return m.BlockSize
	}
	return 0
}

func (m *BlockMeta) GetHeader() Header {
	if m != nil {
		return m.Header
	}
	return Header{}
}

func (m *BlockMeta) GetNumTxs() int64 {
	if m != nil {
		return m.NumTxs
	}
	return 0
}

func init() {
	proto.RegisterEnum("tendermint.proto.types.BlockIDFlag", BlockIDFlag_name, BlockIDFlag_value)
	proto.RegisterEnum("tendermint.proto.types.SignedMsgType", SignedMsgType_name, SignedMsgType_value)
	proto.RegisterType((*PartSetHeader)(nil), "tendermint.proto.types.PartSetHeader")
	proto.RegisterType((*Part)(nil), "tendermint.proto.types.Part")
	proto.RegisterType((*BlockID)(nil), "tendermint.proto.types.BlockID")
	proto.RegisterType((*Header)(nil), "tendermint.proto.types.Header")
	proto.RegisterType((*Data)(nil), "tendermint.proto.types.Data")
	proto.RegisterType((*Vote)(nil), "tendermint.proto.types.Vote")
	proto.RegisterType((*Commit)(nil), "tendermint.proto.types.Commit")
	proto.RegisterType((*CommitSig)(nil), "tendermint.proto.types.CommitSig")
	proto.RegisterType((*Proposal)(nil), "tendermint.proto.types.Proposal")
	proto.RegisterType((*SignedHeader)(nil), "tendermint.proto.types.SignedHeader")
	proto.RegisterType((*BlockMeta)(nil), "tendermint.proto.types.BlockMeta")
}

func init() { proto.RegisterFile("proto/types/types.proto", fileDescriptor_ff06f8095857fb18) }

var fileDescriptor_ff06f8095857fb18 = []byte{
	// 1274 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x57, 0xcd, 0x6e, 0xdb, 0xc6,
	0x13, 0x37, 0x25, 0xca, 0x92, 0x86, 0x92, 0x2d, 0xf3, 0xef, 0x7f, 0xa2, 0xca, 0xad, 0xa5, 0xc8,
	0x4d, 0xea, 0x7c, 0x80, 0x2a, 0x5c, 0xa0, 0x68, 0x80, 0x5e, 0x24, 0xdb, 0x71, 0x84, 0xd8, 0xb2,
	0x40, 0xa9, 0xe9, 0xc7, 0x85, 0x58, 0x89, 0x1b, 0x8a, 0x08, 0x45, 0x12, 0xdc, 0x95, 0x61, 0xa7,
	0x40, 0x81, 0xde, 0x0a, 0x9f, 0xfa, 0x02, 0x3e, 0xa5, 0x05, 0xfa, 0x16, 0xed, 0xb1, 0xa7, 0x3e,
	0x42, 0x0a, 0xa4, 0xaf, 0xd0, 0x07, 0x28, 0xf6, 0x83, 0x94, 0x14, 0xd9, 0x6d, 0xd0, 0xa4, 0x17,
	0x9b, 0x3b, 0xf3, 0x9b, 0xd9, 0x9d, 0xdf, 0xfc, 0x66, 0xd7, 0x86, 0xeb, 0x61, 0x14, 0xd0, 0xa0,
	0x41, 0xcf, 0x42, 0x4c, 0xc4, 0x4f, 0x83, 0x5b, 0xf4, 0x6b, 0x14, 0xfb, 0x36, 0x8e, 0xc6, 0xae,
	0x4f, 0x85, 0xc5, 0xe0, 0xde, 0xca, 0x2d, 0x3a, 0x72, 0x23, 0xdb, 0x0a, 0x51, 0x44, 0xcf, 0x1a,
	0x22, 0xd8, 0x09, 0x9c, 0x60, 0xfa, 0x25, 0xd0, 0x95, 0xaa, 0x13, 0x04, 0x8e, 0x87, 0x05, 0x64,
	0x30, 0x79, 0xd2, 0xa0, 0xee, 0x18, 0x13, 0x8a, 0xc6, 0xa1, 0x04, 0x6c, 0x88, 0x10, 0xcf, 0x1d,
	0x90, 0xc6, 0xc0, 0xa5, 0x73, 0xbb, 0x57, 0xaa, 0xc2, 0x39, 0x8c, 0xce, 0x42, 0x1a, 0x34, 0xc6,
	0x38, 0x7a, 0xea, 0xe1, 0x39, 0x80, 0x8c, 0x3e, 0xc1, 0x11, 0x71, 0x03, 0x3f, 0xfe, 0x2d, 0x9c,
	0xf5, 0xfb, 0x50, 0xec, 0xa2, 0x88, 0xf6, 0x30, 0x7d, 0x88, 0x91, 0x8d, 0x23, 0x7d, 0x1d, 0x32,
	0x34, 0xa0, 0xc8, 0x2b, 0x2b, 0x35, 0x65, 0x3b, 0x6d, 0x8a, 0x85, 0xae, 0x83, 0x3a, 0x42, 0x64,
	0x54, 0x4e, 0xd5, 0x94, 0xed, 0x82, 0xc9, 0xbf, 0xeb, 0x5f, 0x83, 0xca, 0x42, 0x59, 0x84, 0xeb,
	0xdb, 0xf8, 0x94, 0x47, 0x14, 0x4d, 0xb1, 0x60, 0xd6, 0xc1, 0x19, 0xc5, 0x44, 0x86, 0x88, 0x85,
	0x7e, 0x00, 0x99, 0x30, 0x0a, 0x82, 0x27, 0xe5, 0x74, 0x4d, 0xd9, 0xd6, 0x76, 0xee, 0x1a, 0x0b,
	0xd4, 0x89, 0x3a, 0x0c, 0x51, 0x87, 0xd1, 0x73, 0xc7, 0xa1, 0x87, 0xbb, 0x2c, 0xa4, 0xa5, 0xfe,
	0xfa, 0xa2, 0xba, 0x64, 0x8a, 0xf8, 0xfa, 0x18, 0xb2, 0x2d, 0x2f, 0x18, 0x3e, 0x6d, 0xef, 0x25,
	0x67, 0x53, 0xa6, 0x67, 0xd3, 0x3b, 0x50, 0x60, 0xb4, 0x13, 0x6b, 0xc4, 0xab, 0xe2, 0x87, 0xd0,
	0x76, 0x6e, 0x1a, 0x97, 0x77, 0xca, 0x98, 0xa3, 0x40, 0x6e, 0xa4, 0xf1, 0x04, 0xc2, 0x54, 0xff,
	0x36, 0x03, 0xcb, 0x92, 0xa0, 0x5d, 0xc8, 0x4a, 0x0a, 0xf9, 0x8e, 0xda, 0xce, 0xd6, 0x62, 0xd6,
	0x98, 0xe3, 0xdd, 0xc0, 0x27, 0xd8, 0x27, 0x13, 0x22, 0x73, 0xc6, 0x91, 0xfa, 0x2d, 0xc8, 0x0d,
	0x47, 0xc8, 0xf5, 0x2d, 0xd7, 0xe6, 0x67, 0xcb, 0xb7, 0xb4, 0x97, 0x2f, 0xaa, 0xd9, 0x5d, 0x66,
	0x6b, 0xef, 0x99, 0x59, 0xee, 0x6c, 0xdb, 0xfa, 0x35, 0x58, 0x1e, 0x61, 0xd7, 0x19, 0x51, 0x4e,
	0x58, 0xda, 0x94, 0x2b, 0xfd, 0x13, 0x50, 0x99, 0x48, 0xca, 0x2a, 0x3f, 0x41, 0xc5, 0x10, 0x0a,
	0x32, 0x62, 0x05, 0x19, 0xfd, 0x58, 0x41, 0xad, 0x1c, 0xdb, 0xf8, 0xfb, 0xdf, 0xab, 0x8a, 0xc9,
	0x23, 0xf4, 0x2f, 0xa0, 0xe8, 0x21, 0x42, 0xad, 0x01, 0x63, 0x8f, 0x6d, 0x9f, 0xe1, 0x29, 0xaa,
	0x57, 0x51, 0x23, 0x59, 0x6e, 0xfd, 0x8f, 0xe5, 0x79, 0xf9, 0xa2, 0xaa, 0x1d, 0x22, 0x42, 0xa5,
	0xd1, 0xd4, 0xbc, 0x64, 0x61, 0xeb, 0xdb, 0x50, 0xe2, 0x99, 0x87, 0xc1, 0x78, 0xec, 0x52, 0x8b,
	0xf7, 0x64, 0x99, 0xf7, 0x64, 0x85, 0xd9, 0x77, 0xb9, 0xf9, 0x21, 0xeb, 0xce, 0x06, 0xe4, 0x6d,
	0x44, 0x91, 0x80, 0x64, 0x39, 0x24, 0xc7, 0x0c, 0xdc, 0xf9, 0x01, 0xac, 0x9e, 0x20, 0xcf, 0xb5,
	0x11, 0x0d, 0x22, 0x22, 0x20, 0x39, 0x91, 0x65, 0x6a, 0xe6, 0xc0, 0x0f, 0x61, 0xdd, 0xc7, 0xa7,
	0xd4, 0x7a, 0x15, 0x9d, 0xe7, 0x68, 0x9d, 0xf9, 0x1e, 0xcf, 0x47, 0xdc, 0x84, 0x95, 0x61, 0xdc,
	0x11, 0x81, 0x05, 0x8e, 0x2d, 0x26, 0x56, 0x0e, 0x7b, 0x07, 0x72, 0x28, 0x0c, 0x05, 0x40, 0xe3,
	0x80, 0x2c, 0x0a, 0x43, 0xee, 0xba, 0x03, 0x6b, 0xbc, 0xc6, 0x08, 0x93, 0x89, 0x47, 0x65, 0x92,
	0x02, 0xc7, 0xac, 0x32, 0x87, 0x29, 0xec, 0x1c, 0xbb, 0x05, 0x45, 0x7c, 0xe2, 0xda, 0xd8, 0x1f,
	0x62, 0x81, 0x2b, 0x72, 0x5c, 0x21, 0x36, 0x72, 0xd0, 0x6d, 0x28, 0x85, 0x51, 0x10, 0x06, 0x04,
	0x47, 0x16, 0xb2, 0xed, 0x08, 0x13, 0x52, 0x5e, 0x11, 0xf9, 0x62, 0x7b, 0x53, 0x98, 0xeb, 0xf7,
	0x40, 0xdd, 0x43, 0x14, 0xe9, 0x25, 0x48, 0xd3, 0x53, 0x52, 0x56, 0x6a, 0xe9, 0xed, 0x82, 0xc9,
	0x3e, 0x2f, 0x9d, 0xce, 0x3f, 0x53, 0xa0, 0x3e, 0x0e, 0x28, 0xd6, 0xef, 0x83, 0xca, 0x3a, 0xc9,
	0xc5, 0xba, 0x72, 0xf5, 0x08, 0xf4, 0x5c, 0xc7, 0xc7, 0xf6, 0x11, 0x71, 0xfa, 0x67, 0x21, 0x36,
	0x79, 0xc8, 0x8c, 0xfa, 0x52, 0x73, 0xea, 0x5b, 0x87, 0x4c, 0x14, 0x4c, 0x7c, 0x5b, 0x8a, 0x52,
	0x2c, 0xf4, 0x47, 0x90, 0x4b, 0x44, 0xa5, 0xbe, 0x9e, 0xa8, 0x56, 0xa5, 0xa8, 0xe2, 0x59, 0x36,
	0xb3, 0x03, 0x29, 0xa6, 0x16, 0xe4, 0x93, 0x5b, 0x50, 0x4a, 0xf4, 0xf5, 0x54, 0x3e, 0x0d, 0xd3,
	0xef, 0xc2, 0x5a, 0xa2, 0x8d, 0x84, 0x5c, 0xa1, 0xc8, 0x52, 0xe2, 0x90, 0xec, 0xce, 0xc9, 0xce,
	0x12, 0xf7, 0x59, 0x96, 0x57, 0x37, 0x95, 0x5d, 0x9b, 0x5f, 0x6c, 0xef, 0x42, 0x9e, 0xb8, 0x8e,
	0x8f, 0xe8, 0x24, 0xc2, 0x52, 0x99, 0x53, 0x43, 0xfd, 0x79, 0x0a, 0x96, 0x85, 0xd2, 0x67, 0xd8,
	0x53, 0x2e, 0x67, 0x8f, 0x91, 0x9a, 0xb9, 0x8c, 0xbd, 0xf4, 0x9b, 0xb2, 0x77, 0x00, 0x90, 0x1c,
	0x89, 0x94, 0xd5, 0x5a, 0x7a, 0x5b, 0xdb, 0xb9, 0x71, 0x55, 0x3a, 0x71, 0xdc, 0x9e, 0xeb, 0xc8,
	0x4b, 0x6a, 0x26, 0x34, 0x51, 0x56, 0x66, 0xe6, 0x6e, 0x6d, 0x42, 0x7e, 0xe0, 0x52, 0x0b, 0x45,
	0x11, 0x3a, 0xe3, 0x74, 0x6a, 0x3b, 0xef, 0x2f, 0xe6, 0x66, 0x8f, 0x95, 0xc1, 0x1e, 0x2b, 0xa3,
	0xe5, 0xd2, 0x26, 0xc3, 0x9a, 0xb9, 0x81, 0xfc, 0xaa, 0xff, 0xa1, 0x40, 0x3e, 0xd9, 0x56, 0x3f,
	0x80, 0x62, 0x5c, 0xba, 0xf5, 0xc4, 0x43, 0x8e, 0x94, 0xea, 0xd6, 0x3f, 0xd4, 0xff, 0xc0, 0x43,
	0x8e, 0xa9, 0xc9, 0x92, 0xd9, 0xe2, 0xf2, 0x86, 0xa7, 0xae, 0x68, 0xf8, 0x9c, 0xc2, 0xd2, 0xff,
	0x4e, 0x61, 0x73, 0x5a, 0x50, 0x5f, 0xd5, 0xc2, 0xcf, 0x29, 0xc8, 0x75, 0xf9, 0x10, 0x23, 0xef,
	0x3f, 0x1f, 0xc3, 0x44, 0x48, 0x1b, 0x90, 0x0f, 0x03, 0xcf, 0x12, 0x1e, 0x95, 0x7b, 0x72, 0x61,
	0xe0, 0x99, 0x0b, 0x2a, 0xcb, 0xbc, 0xd5, 0x19, 0x5d, 0x7e, 0x0b, 0x0c, 0x66, 0x5f, 0x65, 0xf0,
	0x1b, 0x28, 0x08, 0x42, 0xe4, 0xdb, 0xfb, 0x31, 0x63, 0x82, 0x3f, 0xe8, 0xe2, 0xe9, 0xdd, 0xbc,
	0xea, 0xf0, 0x02, 0x6f, 0x4a, 0x34, 0x8b, 0x13, 0xaf, 0x92, 0xfc, 0x43, 0x60, 0xf3, 0xef, 0x67,
	0xc1, 0x94, 0xe8, 0xfa, 0x6f, 0x0a, 0xe4, 0x79, 0xd9, 0x47, 0x98, 0xa2, 0x39, 0xf2, 0x94, 0x37,
	0x25, 0xef, 0x3d, 0x00, 0x91, 0x8c, 0xb8, 0xcf, 0xb0, 0x6c, 0x6c, 0x9e, 0x5b, 0x7a, 0xee, 0x33,
	0xac, 0x7f, 0x9a, 0x54, 0x9a, 0x7e, 0x9d, 0x4a, 0xe5, 0xe8, 0xc6, 0xf5, 0x5e, 0x87, 0xac, 0x3f,
	0x19, 0x5b, 0xec, 0x99, 0x50, 0x85, 0x64, 0xfc, 0xc9, 0xb8, 0x7f, 0x4a, 0xee, 0xfc, 0xa2, 0x80,
	0x36, 0x33, 0x3e, 0x7a, 0x05, 0xae, 0xb5, 0x0e, 0x8f, 0x77, 0x1f, 0xed, 0x59, 0xed, 0x3d, 0xeb,
	0xc1, 0x61, 0xf3, 0xc0, 0xfa, 0xac, 0xf3, 0xa8, 0x73, 0xfc, 0x79, 0xa7, 0xb4, 0xa4, 0x37, 0x60,
	0x9d, 0xfb, 0x12, 0x57, 0xb3, 0xd5, 0xdb, 0xef, 0xf4, 0x4b, 0x4a, 0xe5, 0xff, 0xe7, 0x17, 0xb5,
	0xb5, 0x99, 0x34, 0xcd, 0x01, 0xc1, 0x3e, 0x5d, 0x0c, 0xd8, 0x3d, 0x3e, 0x3a, 0x6a, 0xf7, 0x4b,
	0xa9, 0x85, 0x00, 0x79, 0x43, 0xde, 0x86, 0xb5, 0xf9, 0x80, 0x4e, 0xfb, 0xb0, 0x94, 0xae, 0xe8,
	0xe7, 0x17, 0xb5, 0x95, 0x19, 0x74, 0xc7, 0xf5, 0x2a, 0xb9, 0xef, 0x9e, 0x6f, 0x2e, 0xfd, 0xf4,
	0xc3, 0xe6, 0xd2, 0x9d, 0x1f, 0x15, 0x28, 0xce, 0x4d, 0x89, 0xbe, 0x01, 0xd7, 0x7b, 0xed, 0x83,
	0xce, 0xfe, 0x9e, 0x75, 0xd4, 0x3b, 0xb0, 0xfa, 0x5f, 0x76, 0xf7, 0x67, 0xaa, 0xb8, 0x01, 0x85,
	0xae, 0xb9, 0xff, 0xf8, 0xb8, 0xbf, 0xcf, 0x3d, 0x25, 0xa5, 0xb2, 0x7a, 0x7e, 0x51, 0xd3, 0xba,
	0x11, 0x3e, 0x09, 0x28, 0xe6, 0xf1, 0x37, 0x61, 0xa5, 0x6b, 0xee, 0x8b, 0xc3, 0x0a, 0x50, 0xaa,
	0xb2, 0x76, 0x7e, 0x51, 0x2b, 0x76, 0x23, 0x2c, 0x84, 0xc0, 0x61, 0x5b, 0x50, 0xec, 0x9a, 0xc7,
	0xdd, 0xe3, 0x5e, 0xf3, 0x50, 0xa0, 0xd2, 0x95, 0xd2, 0xf9, 0x45, 0xad, 0x10, 0x8f, 0x38, 0x03,
	0x4d, 0xcf, 0xd9, 0x32, 0xbe, 0xba, 0xe7, 0xb8, 0x74, 0x34, 0x19, 0x18, 0xc3, 0x60, 0xdc, 0x98,
	0x76, 0x6f, 0xf6, 0x73, 0xe6, 0x3f, 0x8a, 0xc1, 0x32, 0x5f, 0x7c, 0xf4, 0x57, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xfb, 0xb3, 0xf9, 0x43, 0x67, 0x0c, 0x00, 0x00,
}