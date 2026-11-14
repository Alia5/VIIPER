package sdl

/*
#cgo CFLAGS: -I${SRCDIR}/../../../../_testing/e2e/deps/SDL/include

#include <stdlib.h>

#include <SDL3/SDL_audio.h>

static inline const char *audio_device_container_id(SDL_AudioDeviceID id)
{
	return SDL_GetStringProperty(SDL_GetAudioDeviceProperties(id), SDL_PROP_AUDIO_DEVICE_CONTAINER_ID_STRING, "");
}
*/
import "C"

import "unsafe"

type AudioDeviceID uint32

type AudioFormat uint32

const AudioFormatS16LE AudioFormat = C.SDL_AUDIO_S16LE

type AudioSpec struct {
	Format   AudioFormat
	Channels int
	Freq     int
}

type AudioStream struct {
	cStream *C.SDL_AudioStream
}

func GetAudioPlaybackDevices() []AudioDeviceID {
	var count C.int
	cIDs := C.SDL_GetAudioPlaybackDevices(&count)
	return audioDeviceIDs(cIDs, count)
}

func GetAudioRecordingDevices() []AudioDeviceID {
	var count C.int
	cIDs := C.SDL_GetAudioRecordingDevices(&count)
	return audioDeviceIDs(cIDs, count)
}

func audioDeviceIDs(cIDs *C.SDL_AudioDeviceID, count C.int) []AudioDeviceID {
	if cIDs == nil {
		return nil
	}
	defer C.SDL_free(unsafe.Pointer(cIDs))

	ids := make([]AudioDeviceID, 0, int(count))
	for _, id := range unsafe.Slice(cIDs, int(count)) {
		ids = append(ids, AudioDeviceID(id))
	}
	return ids
}

func GetAudioDeviceName(id AudioDeviceID) string {
	cName := C.SDL_GetAudioDeviceName(C.SDL_AudioDeviceID(id))
	if cName == nil {
		return ""
	}
	return C.GoString(cName)
}

func GetAudioDeviceContainerID(id AudioDeviceID) string {
	return C.GoString(C.audio_device_container_id(C.SDL_AudioDeviceID(id)))
}

func OpenAudioDeviceStream(id AudioDeviceID, spec *AudioSpec) (*AudioStream, error) {
	cSpec := C.SDL_AudioSpec{
		format:   C.SDL_AudioFormat(spec.Format),
		channels: C.int(spec.Channels),
		freq:     C.int(spec.Freq),
	}
	cStream := C.SDL_OpenAudioDeviceStream(C.SDL_AudioDeviceID(id), &cSpec, nil, nil)
	if cStream == nil {
		return nil, GetError()
	}
	return &AudioStream{
		cStream: cStream,
	}, nil
}

func (s *AudioStream) ResumeDevice() error {
	res := C.SDL_ResumeAudioStreamDevice(s.cStream)
	if !res {
		return GetError()
	}
	return nil
}

func (s *AudioStream) PutData(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	res := C.SDL_PutAudioStreamData(s.cStream, unsafe.Pointer(&data[0]), C.int(len(data)))
	if !res {
		return GetError()
	}
	return nil
}

func (s *AudioStream) GetData(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	return int(C.SDL_GetAudioStreamData(s.cStream, unsafe.Pointer(&buf[0]), C.int(len(buf))))
}

func (s *AudioStream) Destroy() {
	C.SDL_DestroyAudioStream(s.cStream)
}
