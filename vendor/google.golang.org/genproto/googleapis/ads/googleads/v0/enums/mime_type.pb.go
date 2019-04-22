// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v0/enums/mime_type.proto

package enums // import "google.golang.org/genproto/googleapis/ads/googleads/v0/enums"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// The mime type
type MimeTypeEnum_MimeType int32

const (
	// The mime type has not been specified.
	MimeTypeEnum_UNSPECIFIED MimeTypeEnum_MimeType = 0
	// The received value is not known in this version.
	//
	// This is a response-only value.
	MimeTypeEnum_UNKNOWN MimeTypeEnum_MimeType = 1
	// MIME type of image/jpeg.
	MimeTypeEnum_IMAGE_JPEG MimeTypeEnum_MimeType = 2
	// MIME type of image/gif.
	MimeTypeEnum_IMAGE_GIF MimeTypeEnum_MimeType = 3
	// MIME type of image/png.
	MimeTypeEnum_IMAGE_PNG MimeTypeEnum_MimeType = 4
	// MIME type of application/x-shockwave-flash.
	MimeTypeEnum_FLASH MimeTypeEnum_MimeType = 5
	// MIME type of text/html.
	MimeTypeEnum_TEXT_HTML MimeTypeEnum_MimeType = 6
	// MIME type of application/pdf.
	MimeTypeEnum_PDF MimeTypeEnum_MimeType = 7
	// MIME type of application/msword.
	MimeTypeEnum_MSWORD MimeTypeEnum_MimeType = 8
	// MIME type of application/vnd.ms-excel.
	MimeTypeEnum_MSEXCEL MimeTypeEnum_MimeType = 9
	// MIME type of application/rtf.
	MimeTypeEnum_RTF MimeTypeEnum_MimeType = 10
	// MIME type of audio/wav.
	MimeTypeEnum_AUDIO_WAV MimeTypeEnum_MimeType = 11
	// MIME type of audio/mp3.
	MimeTypeEnum_AUDIO_MP3 MimeTypeEnum_MimeType = 12
	// MIME type of application/x-html5-ad-zip.
	MimeTypeEnum_HTML5_AD_ZIP MimeTypeEnum_MimeType = 13
)

var MimeTypeEnum_MimeType_name = map[int32]string{
	0:  "UNSPECIFIED",
	1:  "UNKNOWN",
	2:  "IMAGE_JPEG",
	3:  "IMAGE_GIF",
	4:  "IMAGE_PNG",
	5:  "FLASH",
	6:  "TEXT_HTML",
	7:  "PDF",
	8:  "MSWORD",
	9:  "MSEXCEL",
	10: "RTF",
	11: "AUDIO_WAV",
	12: "AUDIO_MP3",
	13: "HTML5_AD_ZIP",
}
var MimeTypeEnum_MimeType_value = map[string]int32{
	"UNSPECIFIED":  0,
	"UNKNOWN":      1,
	"IMAGE_JPEG":   2,
	"IMAGE_GIF":    3,
	"IMAGE_PNG":    4,
	"FLASH":        5,
	"TEXT_HTML":    6,
	"PDF":          7,
	"MSWORD":       8,
	"MSEXCEL":      9,
	"RTF":          10,
	"AUDIO_WAV":    11,
	"AUDIO_MP3":    12,
	"HTML5_AD_ZIP": 13,
}

func (x MimeTypeEnum_MimeType) String() string {
	return proto.EnumName(MimeTypeEnum_MimeType_name, int32(x))
}
func (MimeTypeEnum_MimeType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_mime_type_9d68a7e3d7ca4947, []int{0, 0}
}

// Container for enum describing the mime types.
type MimeTypeEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MimeTypeEnum) Reset()         { *m = MimeTypeEnum{} }
func (m *MimeTypeEnum) String() string { return proto.CompactTextString(m) }
func (*MimeTypeEnum) ProtoMessage()    {}
func (*MimeTypeEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_mime_type_9d68a7e3d7ca4947, []int{0}
}
func (m *MimeTypeEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MimeTypeEnum.Unmarshal(m, b)
}
func (m *MimeTypeEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MimeTypeEnum.Marshal(b, m, deterministic)
}
func (dst *MimeTypeEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MimeTypeEnum.Merge(dst, src)
}
func (m *MimeTypeEnum) XXX_Size() int {
	return xxx_messageInfo_MimeTypeEnum.Size(m)
}
func (m *MimeTypeEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_MimeTypeEnum.DiscardUnknown(m)
}

var xxx_messageInfo_MimeTypeEnum proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MimeTypeEnum)(nil), "google.ads.googleads.v0.enums.MimeTypeEnum")
	proto.RegisterEnum("google.ads.googleads.v0.enums.MimeTypeEnum_MimeType", MimeTypeEnum_MimeType_name, MimeTypeEnum_MimeType_value)
}

func init() {
	proto.RegisterFile("google/ads/googleads/v0/enums/mime_type.proto", fileDescriptor_mime_type_9d68a7e3d7ca4947)
}

var fileDescriptor_mime_type_9d68a7e3d7ca4947 = []byte{
	// 375 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xbd, 0xee, 0xda, 0x30,
	0x14, 0xc5, 0x9b, 0xd0, 0x3f, 0x1f, 0x06, 0x5a, 0xcb, 0x3b, 0x03, 0xec, 0x75, 0x22, 0xa1, 0x2e,
	0xee, 0x64, 0x88, 0x13, 0xd2, 0x92, 0x60, 0x91, 0x10, 0x10, 0x8a, 0x14, 0xd1, 0x26, 0x8a, 0x90,
	0xc8, 0x87, 0x30, 0x20, 0xf1, 0x3a, 0x1d, 0xfb, 0x28, 0xed, 0x2b, 0x74, 0xea, 0xd8, 0xa7, 0xa8,
	0x9c, 0x34, 0x30, 0xb5, 0x8b, 0x75, 0xee, 0x3d, 0x3f, 0x5f, 0xdb, 0xc7, 0xe0, 0x5d, 0x5a, 0x14,
	0xe9, 0x29, 0xd1, 0x0e, 0xb1, 0xd0, 0x6a, 0x29, 0xd5, 0x4d, 0xd7, 0x92, 0xfc, 0x9a, 0x09, 0x2d,
	0x3b, 0x66, 0x49, 0x74, 0xb9, 0x97, 0x09, 0x2e, 0xcf, 0xc5, 0xa5, 0x40, 0xa3, 0x9a, 0xc1, 0x87,
	0x58, 0xe0, 0x07, 0x8e, 0x6f, 0x3a, 0xae, 0xf0, 0xc9, 0x4f, 0x05, 0x0c, 0x9c, 0x63, 0x96, 0xf8,
	0xf7, 0x32, 0x61, 0xf9, 0x35, 0x9b, 0xfc, 0x50, 0x40, 0xb7, 0x69, 0xa0, 0xb7, 0xa0, 0xbf, 0x71,
	0x3d, 0xce, 0xe6, 0xb6, 0x69, 0x33, 0x03, 0xbe, 0x42, 0x7d, 0xd0, 0xd9, 0xb8, 0x9f, 0xdc, 0xd5,
	0xd6, 0x85, 0x0a, 0x7a, 0x03, 0x80, 0xed, 0x50, 0x8b, 0x45, 0x1f, 0x39, 0xb3, 0xa0, 0x8a, 0x86,
	0xa0, 0x57, 0xd7, 0x96, 0x6d, 0xc2, 0xd6, 0xb3, 0xe4, 0xae, 0x05, 0x5f, 0xa3, 0x1e, 0x78, 0x31,
	0x97, 0xd4, 0x5b, 0xc0, 0x17, 0xe9, 0xf8, 0x6c, 0xe7, 0x47, 0x0b, 0xdf, 0x59, 0xc2, 0x36, 0xea,
	0x80, 0x16, 0x37, 0x4c, 0xd8, 0x41, 0x00, 0xb4, 0x1d, 0x6f, 0xbb, 0x5a, 0x1b, 0xb0, 0x2b, 0x4f,
	0x72, 0x3c, 0xb6, 0x9b, 0xb3, 0x25, 0xec, 0x49, 0x62, 0xed, 0x9b, 0x10, 0xc8, 0x9d, 0x74, 0x63,
	0xd8, 0xab, 0x68, 0x4b, 0x03, 0xd8, 0x7f, 0x96, 0x0e, 0x9f, 0xc2, 0x01, 0x82, 0x60, 0x20, 0x47,
	0xbe, 0x8f, 0xa8, 0x11, 0xed, 0x6d, 0x0e, 0x87, 0xb3, 0x5f, 0x0a, 0x18, 0x7f, 0x29, 0x32, 0xfc,
	0xdf, 0x10, 0x66, 0xc3, 0xe6, 0xc1, 0x5c, 0x46, 0xc6, 0x95, 0xfd, 0xec, 0x2f, 0x9f, 0x16, 0xa7,
	0x43, 0x9e, 0xe2, 0xe2, 0x9c, 0x6a, 0x69, 0x92, 0x57, 0x81, 0x36, 0x99, 0x97, 0x47, 0xf1, 0x8f,
	0x2f, 0xf8, 0x50, 0xad, 0x5f, 0xd5, 0x96, 0x45, 0xe9, 0x37, 0x75, 0x64, 0xd5, 0xa3, 0x68, 0x2c,
	0x70, 0x2d, 0xa5, 0x0a, 0x74, 0x2c, 0xd3, 0x16, 0xdf, 0x1b, 0x3f, 0xa4, 0xb1, 0x08, 0x1f, 0x7e,
	0x18, 0xe8, 0x61, 0xe5, 0xff, 0x56, 0xc7, 0x75, 0x93, 0x10, 0x1a, 0x0b, 0x42, 0x1e, 0x04, 0x21,
	0x81, 0x4e, 0x48, 0xc5, 0x7c, 0x6e, 0x57, 0x17, 0x9b, 0xfe, 0x09, 0x00, 0x00, 0xff, 0xff, 0xc7,
	0xb3, 0xac, 0x09, 0x1a, 0x02, 0x00, 0x00,
}
