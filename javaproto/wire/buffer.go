package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"unicode/utf8"

	"github.com/go-mclib/nbt"
)

// maxPrealloc caps slice preallocation from wire counts; larger collections
// grow as their elements decode, so a forged count cannot force a huge allocation.
const maxPrealloc = 1024

var (
	errNotReadMode  = errors.New("buffer not in read mode")
	errNotWriteMode = errors.New("buffer not in write mode")
)

// ByteArray is a raw byte sequence used throughout the protocol.
type ByteArray = []byte

// PacketBuffer provides methods for reading and writing Minecraft protocol data types.
// It wraps io.Reader and io.Writer interfaces for streaming network communication.
type PacketBuffer struct {
	reader io.Reader
	writer io.Writer

	// For writer mode, we also keep a bytes.Buffer to retrieve written bytes
	buf *bytes.Buffer
}

// NewReader creates a PacketBuffer for reading from data.
func NewReader(data []byte) *PacketBuffer {
	return &PacketBuffer{
		reader: bytes.NewReader(data),
	}
}

// NewReaderFrom creates a PacketBuffer for reading from an io.Reader.
func NewReaderFrom(r io.Reader) *PacketBuffer {
	return &PacketBuffer{
		reader: r,
	}
}

// NewWriter creates a PacketBuffer for writing data.
func NewWriter() *PacketBuffer {
	buf := &bytes.Buffer{}
	return &PacketBuffer{
		writer: buf,
		buf:    buf,
	}
}

// NewWriterTo creates a PacketBuffer that writes directly to an io.Writer.
func NewWriterTo(w io.Writer) *PacketBuffer {
	return &PacketBuffer{
		writer: w,
	}
}

// Bytes returns the written bytes. Only valid for buffers created with NewWriter.
func (pb *PacketBuffer) Bytes() []byte {
	if pb.buf != nil {
		return pb.buf.Bytes()
	}
	return nil
}

// Len returns the number of written bytes. Only valid for buffers created with NewWriter.
func (pb *PacketBuffer) Len() int {
	if pb.buf != nil {
		return pb.buf.Len()
	}
	return 0
}

// Reset resets the buffer for reuse. Only valid for buffers created with NewWriter.
func (pb *PacketBuffer) Reset() {
	if pb.buf != nil {
		pb.buf.Reset()
	}
}

// Read reads exactly len(p) bytes from the buffer.
func (pb *PacketBuffer) Read(p []byte) (int, error) {
	if pb.reader == nil {
		return 0, errNotReadMode
	}
	return io.ReadFull(pb.reader, p)
}

// Write writes p to the buffer.
func (pb *PacketBuffer) Write(p []byte) (int, error) {
	if pb.writer == nil {
		return 0, errNotWriteMode
	}
	return pb.writer.Write(p)
}

// ReadByte reads a single byte.
func (pb *PacketBuffer) ReadByte() (byte, error) {
	var b [1]byte
	_, err := pb.Read(b[:])
	return b[0], err
}

// WriteByte writes a single byte.
func (pb *PacketBuffer) WriteByte(b byte) error {
	_, err := pb.Write([]byte{b})
	return err
}

// Reader returns the underlying io.Reader.
func (pb *PacketBuffer) Reader() io.Reader {
	return pb.reader
}

// Writer returns the underlying io.Writer.
func (pb *PacketBuffer) Writer() io.Writer {
	return pb.writer
}

// ReadVarInt reads a variable-length 32-bit integer.
func (pb *PacketBuffer) ReadVarInt() (VarInt, error) {
	return DecodeVarInt(pb)
}

// WriteVarInt writes a variable-length 32-bit integer.
func (pb *PacketBuffer) WriteVarInt(v VarInt) error {
	var b [binary.MaxVarintLen32]byte
	_, err := pb.Write(v.Append(b[:0]))
	return err
}

// ReadVarLong reads a variable-length 64-bit integer.
func (pb *PacketBuffer) ReadVarLong() (VarLong, error) {
	var value int64
	var position uint
	for {
		b, err := pb.ReadByte()
		if err != nil {
			return 0, err
		}
		value |= int64(b&0x7F) << position
		if b&0x80 == 0 {
			return VarLong(value), nil
		}
		position += 7
		if position >= 70 {
			return 0, errVarLongTooBig
		}
	}
}

// WriteVarLong writes a variable-length 64-bit integer.
func (pb *PacketBuffer) WriteVarLong(v VarLong) error {
	var b [binary.MaxVarintLen64]byte
	_, err := pb.Write(v.Append(b[:0]))
	return err
}

// ReadBool reads a boolean (1 byte: 0x00 = false, any other value = true).
func (pb *PacketBuffer) ReadBool() (Boolean, error) {
	b, err := pb.ReadByte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

// WriteBool writes a boolean.
func (pb *PacketBuffer) WriteBool(v Boolean) error {
	var b byte
	if v {
		b = 0x01
	}
	return pb.WriteByte(b)
}

// ReadInt8 reads a signed 8-bit integer.
func (pb *PacketBuffer) ReadInt8() (Int8, error) {
	b, err := pb.ReadByte()
	return Int8(b), err
}

// WriteInt8 writes a signed 8-bit integer.
func (pb *PacketBuffer) WriteInt8(v Int8) error {
	return pb.WriteByte(byte(v))
}

// ReadUint8 reads an unsigned 8-bit integer.
func (pb *PacketBuffer) ReadUint8() (Uint8, error) {
	b, err := pb.ReadByte()
	return Uint8(b), err
}

// WriteUint8 writes an unsigned 8-bit integer.
func (pb *PacketBuffer) WriteUint8(v Uint8) error {
	return pb.WriteByte(byte(v))
}

// ReadInt16 reads a big-endian signed 16-bit integer.
func (pb *PacketBuffer) ReadInt16() (Int16, error) {
	v, err := pb.ReadUint16()
	return Int16(v), err
}

// WriteInt16 writes a big-endian signed 16-bit integer.
func (pb *PacketBuffer) WriteInt16(v Int16) error {
	return pb.WriteUint16(Uint16(v))
}

// ReadUint16 reads a big-endian unsigned 16-bit integer.
func (pb *PacketBuffer) ReadUint16() (Uint16, error) {
	var b [2]byte
	if _, err := pb.Read(b[:]); err != nil {
		return 0, err
	}
	return Uint16(binary.BigEndian.Uint16(b[:])), nil
}

// WriteUint16 writes a big-endian unsigned 16-bit integer.
func (pb *PacketBuffer) WriteUint16(v Uint16) error {
	var b [2]byte
	_, err := pb.Write(binary.BigEndian.AppendUint16(b[:0], uint16(v)))
	return err
}

// ReadInt32 reads a big-endian signed 32-bit integer.
func (pb *PacketBuffer) ReadInt32() (Int32, error) {
	v, err := pb.readUint32()
	return Int32(v), err
}

// WriteInt32 writes a big-endian signed 32-bit integer.
func (pb *PacketBuffer) WriteInt32(v Int32) error {
	return pb.writeUint32(uint32(v))
}

// ReadInt64 reads a big-endian signed 64-bit integer.
func (pb *PacketBuffer) ReadInt64() (Int64, error) {
	v, err := pb.readUint64()
	return Int64(v), err
}

// WriteInt64 writes a big-endian signed 64-bit integer.
func (pb *PacketBuffer) WriteInt64(v Int64) error {
	return pb.writeUint64(uint64(v))
}

// ReadFloat32 reads a big-endian 32-bit IEEE 754 float.
func (pb *PacketBuffer) ReadFloat32() (Float32, error) {
	v, err := pb.readUint32()
	return Float32(math.Float32frombits(v)), err
}

// WriteFloat32 writes a big-endian 32-bit IEEE 754 float.
func (pb *PacketBuffer) WriteFloat32(v Float32) error {
	return pb.writeUint32(math.Float32bits(float32(v)))
}

// ReadFloat64 reads a big-endian 64-bit IEEE 754 double.
func (pb *PacketBuffer) ReadFloat64() (Float64, error) {
	v, err := pb.readUint64()
	return Float64(math.Float64frombits(v)), err
}

// WriteFloat64 writes a big-endian 64-bit IEEE 754 double.
func (pb *PacketBuffer) WriteFloat64(v Float64) error {
	return pb.writeUint64(math.Float64bits(float64(v)))
}

func (pb *PacketBuffer) readUint32() (uint32, error) {
	var b [4]byte
	if _, err := pb.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b[:]), nil
}

func (pb *PacketBuffer) writeUint32(v uint32) error {
	var b [4]byte
	_, err := pb.Write(binary.BigEndian.AppendUint32(b[:0], v))
	return err
}

func (pb *PacketBuffer) readUint64() (uint64, error) {
	var b [8]byte
	if _, err := pb.Read(b[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(b[:]), nil
}

func (pb *PacketBuffer) writeUint64(v uint64) error {
	var b [8]byte
	_, err := pb.Write(binary.BigEndian.AppendUint64(b[:0], v))
	return err
}

// ReadString reads a UTF-8 string with VarInt length prefix.
// maxLen is the maximum allowed string length in characters (usually 32767, 0 means no limit).
func (pb *PacketBuffer) ReadString(maxLen int) (String, error) {
	length, err := pb.ReadVarInt()
	if err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}
	if length < 0 {
		return "", fmt.Errorf("negative string length: %d", length)
	}

	// a UTF-8 character takes at most 4 bytes
	if maxBytes := maxLen * 4; maxLen > 0 && int(length) > maxBytes {
		return "", fmt.Errorf("string byte length %d exceeds maximum %d", length, maxBytes)
	}

	data, err := pb.ReadFixedByteArray(int(length))
	if err != nil {
		return "", fmt.Errorf("failed to read string data: %w", err)
	}
	if maxLen > 0 {
		if n := utf8.RuneCount(data); n > maxLen {
			return "", fmt.Errorf("string length %d exceeds maximum %d characters", n, maxLen)
		}
	}
	return String(data), nil
}

// WriteString writes a UTF-8 string with VarInt length prefix.
func (pb *PacketBuffer) WriteString(v String) error {
	if err := pb.WriteVarInt(VarInt(len(v))); err != nil {
		return fmt.Errorf("failed to write string length: %w", err)
	}
	// the length write succeeded, so pb.writer is set
	if _, err := io.WriteString(pb.writer, string(v)); err != nil {
		return fmt.Errorf("failed to write string data: %w", err)
	}
	return nil
}

// ReadIdentifier reads a Minecraft identifier (namespaced ID).
func (pb *PacketBuffer) ReadIdentifier() (Identifier, error) {
	s, err := pb.ReadString(32767)
	return Identifier(s), err
}

// WriteIdentifier writes a Minecraft identifier.
func (pb *PacketBuffer) WriteIdentifier(v Identifier) error {
	return pb.WriteString(String(v))
}

// ReadByteArray reads a byte array with VarInt length prefix.
func (pb *PacketBuffer) ReadByteArray(maxLen int) (ByteArray, error) {
	length, err := pb.ReadVarInt()
	if err != nil {
		return nil, fmt.Errorf("failed to read byte array length: %w", err)
	}

	if maxLen > 0 && int(length) > maxLen {
		return nil, fmt.Errorf("byte array length %d exceeds maximum %d", length, maxLen)
	}

	data, err := pb.ReadFixedByteArray(int(length))
	if err != nil {
		return nil, fmt.Errorf("failed to read byte array data: %w", err)
	}
	return data, nil
}

// WriteByteArray writes a byte array with VarInt length prefix.
func (pb *PacketBuffer) WriteByteArray(v ByteArray) error {
	if err := pb.WriteVarInt(VarInt(len(v))); err != nil {
		return fmt.Errorf("failed to write byte array length: %w", err)
	}
	if _, err := pb.Write(v); err != nil {
		return fmt.Errorf("failed to write byte array data: %w", err)
	}
	return nil
}

// ReadFixedByteArray reads exactly n bytes.
func (pb *PacketBuffer) ReadFixedByteArray(n int) (ByteArray, error) {
	if pb.reader == nil {
		return nil, errNotReadMode
	}
	return readN(pb.reader, n)
}

// ReadRemaining reads all bytes left in the buffer.
func (pb *PacketBuffer) ReadRemaining() ([]byte, error) {
	if pb.reader == nil {
		return nil, errNotReadMode
	}
	return io.ReadAll(pb.reader)
}

// ReadNBT reads a network-format (nameless root) NBT tag.
func (pb *PacketBuffer) ReadNBT() (nbt.Tag, error) {
	if pb.reader == nil {
		return nil, errNotReadMode
	}
	tag, _, err := nbt.NewReaderFrom(pb.reader).ReadTag(nbt.FormatNetwork)
	return tag, err
}

// WriteNBT writes tag in network format (nameless root).
func (pb *PacketBuffer) WriteNBT(tag nbt.Tag) error {
	if pb.writer == nil {
		return errNotWriteMode
	}
	return nbt.NewWriterTo(pb.writer).WriteTag(tag, "", nbt.FormatNetwork)
}

// remaining returns the unread byte count when the reader can report it.
func remaining(r io.Reader) (int, bool) {
	if l, ok := r.(interface{ Len() int }); ok {
		return l.Len(), true
	}
	return 0, false
}

// readN reads exactly n bytes. The length comes from the wire, so it is only
// trusted for preallocation once the input is known to hold that many bytes.
func readN(r io.Reader, n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("negative length %d", n)
	}
	rem, known := remaining(r)
	if known && n > rem {
		return nil, io.ErrUnexpectedEOF
	}
	if known || n <= maxPrealloc {
		data := make([]byte, n)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
		return data, nil
	}
	var b bytes.Buffer
	if _, err := io.CopyN(&b, r, int64(n)); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return b.Bytes(), nil
}

// readCount reads a VarInt element count, rejecting negative counts and counts
// whose elements (at least minSize bytes each) cannot fit in the remaining input.
func (pb *PacketBuffer) readCount(minSize int) (int, error) {
	n, err := pb.ReadVarInt()
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, fmt.Errorf("negative count %d", n)
	}
	if rem, ok := remaining(pb.reader); ok && int64(n)*int64(minSize) > int64(rem) {
		return 0, fmt.Errorf("count %d exceeds remaining %d bytes", n, rem)
	}
	return int(n), nil
}

// makeCapped returns an empty slice for n wire-counted elements, preallocating
// at most maxPrealloc of them.
func makeCapped[T any](n int) []T {
	return make([]T, 0, min(n, maxPrealloc))
}

// WriteFixedByteArray writes bytes without length prefix.
func (pb *PacketBuffer) WriteFixedByteArray(v ByteArray) error {
	_, err := pb.Write(v)
	return err
}

// ReadPosition reads a block position packed into a 64-bit integer.
func (pb *PacketBuffer) ReadPosition() (Position, error) {
	v, err := pb.readUint64()
	if err != nil {
		return Position{}, err
	}
	return unpackPosition(int64(v)), nil
}

// WritePosition writes a block position packed into a 64-bit integer.
func (pb *PacketBuffer) WritePosition(v Position) error {
	return pb.writeUint64(uint64(v.Pack()))
}

// ReadGlobalPos reads a global position (dimension + block position).
func (pb *PacketBuffer) ReadGlobalPos() (GlobalPos, error) {
	dim, err := pb.ReadIdentifier()
	if err != nil {
		return GlobalPos{}, err
	}
	pos, err := pb.ReadPosition()
	if err != nil {
		return GlobalPos{}, err
	}
	return GlobalPos{Dimension: dim, Pos: pos}, nil
}

// WriteGlobalPos writes a global position (dimension + block position).
func (pb *PacketBuffer) WriteGlobalPos(v GlobalPos) error {
	if err := pb.WriteIdentifier(v.Dimension); err != nil {
		return err
	}
	return pb.WritePosition(v.Pos)
}

// ReadUUID reads a 128-bit UUID (two 64-bit integers).
func (pb *PacketBuffer) ReadUUID() (UUID, error) {
	var u UUID
	if _, err := pb.Read(u[:]); err != nil {
		return UUID{}, err
	}
	return u, nil
}

// WriteUUID writes a 128-bit UUID.
func (pb *PacketBuffer) WriteUUID(v UUID) error {
	_, err := pb.Write(v[:])
	return err
}

// ReadAngle reads a rotation angle (1 byte = 1/256 of a full turn).
func (pb *PacketBuffer) ReadAngle() (Angle, error) {
	b, err := pb.ReadByte()
	return Angle(b), err
}

// WriteAngle writes a rotation angle.
func (pb *PacketBuffer) WriteAngle(v Angle) error {
	return pb.WriteByte(byte(v))
}

// CopyVarInt copies a VarInt from src to this buffer.
func (pb *PacketBuffer) CopyVarInt(src *PacketBuffer) error {
	v, err := src.ReadVarInt()
	if err != nil {
		return err
	}
	return pb.WriteVarInt(v)
}

// CopyVarLong copies a VarLong from src to this buffer.
func (pb *PacketBuffer) CopyVarLong(src *PacketBuffer) error {
	v, err := src.ReadVarLong()
	if err != nil {
		return err
	}
	return pb.WriteVarLong(v)
}

// CopyBool copies a Boolean from src to this buffer.
func (pb *PacketBuffer) CopyBool(src *PacketBuffer) error {
	v, err := src.ReadBool()
	if err != nil {
		return err
	}
	return pb.WriteBool(v)
}

// CopyInt8 copies an Int8 from src to this buffer.
func (pb *PacketBuffer) CopyInt8(src *PacketBuffer) error {
	v, err := src.ReadInt8()
	if err != nil {
		return err
	}
	return pb.WriteInt8(v)
}

// CopyInt32 copies an Int32 from src to this buffer.
func (pb *PacketBuffer) CopyInt32(src *PacketBuffer) error {
	v, err := src.ReadInt32()
	if err != nil {
		return err
	}
	return pb.WriteInt32(v)
}

// CopyFloat32 copies a Float32 from src to this buffer.
func (pb *PacketBuffer) CopyFloat32(src *PacketBuffer) error {
	v, err := src.ReadFloat32()
	if err != nil {
		return err
	}
	return pb.WriteFloat32(v)
}

// CopyString copies a String from src to this buffer.
func (pb *PacketBuffer) CopyString(src *PacketBuffer, maxLen int) error {
	v, err := src.ReadString(maxLen)
	if err != nil {
		return err
	}
	return pb.WriteString(v)
}

// CopyUUID copies a UUID from src to this buffer.
func (pb *PacketBuffer) CopyUUID(src *PacketBuffer) error {
	v, err := src.ReadUUID()
	if err != nil {
		return err
	}
	return pb.WriteUUID(v)
}

// CopyPosition copies a Position from src to this buffer.
func (pb *PacketBuffer) CopyPosition(src *PacketBuffer) error {
	v, err := src.ReadPosition()
	if err != nil {
		return err
	}
	return pb.WritePosition(v)
}
