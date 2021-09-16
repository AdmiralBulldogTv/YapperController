package parts

import "github.com/troydota/tts-textparser/src/datastructures"

type PartType int

const (
	PartTypeCurrency PartType = iota
	PartTypeRaw
	PartTypeOverride
)

type SpaceType int

const (
	SpaceTypeShortPause  = iota // space
	SpaceTypeMediumPause        // unused
	SpaceTypeLongPause          // fullstop
)

type VoicePartMeta struct {
	Space SpaceType
	Idx   int
}

type VoicePartType int

const (
	VoicePartTypeReader VoicePartType = iota
	VoicePartTypeByte
)

type VoicePart struct {
	VoicePartMeta
	Value string
	Voice Voice
	Type  PartType
}

type Voice struct {
	Name  string
	Type  VoicePartType
	Entry datastructures.AudioConfig
}

type VoicePartList []VoicePart

func (pm VoicePartList) Map() map[Voice]VoicePartList {
	mp := make(map[Voice]VoicePartList)
	for i, v := range pm {
		v.Idx = i
		mp[v.Voice] = append(mp[v.Voice], v)
	}

	return mp
}

func (vpl VoicePartList) Unique() map[VoicePart][]VoicePartMeta {
	mp := make(map[VoicePart][]VoicePartMeta)
	for _, v := range vpl {
		meta := v.VoicePartMeta
		v.VoicePartMeta = VoicePartMeta{}
		mp[v] = append(mp[v], meta)
	}

	return mp
}
