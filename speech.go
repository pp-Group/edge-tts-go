package edge_tts_go

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/pp-group/edge-tts-go/biz/service/tts/edge"
	file_helper "github.com/pp-group/file-helper"
	storage "github.com/pp-group/file-helper/storage"
)

type ISpeech interface {
	GenTTS() error
	URL(filename string) (string, error)
	FileName() string
}

var _ ISpeech = new(LocalSpeech)

type LocalSpeech struct {
	*Speech
}

func NewLocalSpeech(c *edge.Communicate, folder string) (*LocalSpeech, error) {

	fileStorage, err := file_helper.FileStorageFactory(folder)()
	if err != nil {
		return nil, err
	}

	s, err := NewSpeech(c, fileStorage, folder)
	if err != nil {
		return nil, err
	}

	return &LocalSpeech{
		Speech: s,
	}, nil

}

func (speech *LocalSpeech) GenTTS() error {
	return gentts(speech.Speech, func() (storage.IWriteBroker, error) {
		return speech.Writer(speech.fileName, nil)
	})
}

func (speech *LocalSpeech) URL(filename string) (string, error) {
	return url(func() (storage.IReadBroker, error) {
		return speech.Reader(filename, nil)
	})
}

var _ ISpeech = new(OssSpeech)

type OssSpeech struct {
	*Speech
	bucket string
}

func NewOssSpeech(c *edge.Communicate, endpoint, ak, sk, folder, bucket string) (*OssSpeech, error) {

	ossStorage, err := file_helper.OssStorageFactory(endpoint, ak, sk, folder)()
	if err != nil {
		return nil, err
	}

	s, err := NewSpeech(c, ossStorage, folder)
	if err != nil {
		return nil, err
	}

	return &OssSpeech{
		Speech: s,
		bucket: bucket,
	}, nil
}

func (speech *OssSpeech) GenTTS() error {

	return gentts(speech.Speech, func() (storage.IWriteBroker, error) {
		return speech.Writer(speech.fileName, func() interface{} {
			return speech.bucket
		})
	})
}

func (speech *OssSpeech) URL(filename string) (string, error) {
	return url(func() (storage.IReadBroker, error) {
		return speech.Reader(filename, func() interface{} {
			return speech.bucket
		})
	})
}
func gentts(speech *Speech, brokerFunc func() (storage.IWriteBroker, error)) error {

	fileName := generateHashName(speech.Text, speech.VoiceLangRegion) + ".mp3"

	speech.fileName = fileName

	broker, err := brokerFunc()
	if err != nil {
		return err
	}

	err = speech.gen(broker)
	if err != nil {
		return err
	}
	return nil
}

func url(brokerFunc func() (storage.IReadBroker, error)) (string, error) {

	broker, err := brokerFunc()
	if err != nil {
		return "", err
	}
	return broker.URL()
}

type Speech struct {
	*edge.Communicate
	storage.IStorage
	folder   string
	fileName string
}

func (s *Speech) FileName() string {
	return s.fileName
}

func NewSpeech(c *edge.Communicate, storage storage.IStorage, folder string) (*Speech, error) {
	s := &Speech{
		Communicate: c,
		IStorage:    storage,
		folder:      folder,
	}
	return s, nil

}

func (s *Speech) gen(broker storage.IWriteBroker) error {
	op, err := s.Stream()
	if err != nil {
		return err
	}
	defer s.CloseOutput()
	solveCount := 0
	audioData := make([][][]byte, s.AudioDataIndex)
	for i := range op {
		if _, ok := i["end"]; ok {
			solveCount++
			if solveCount == s.AudioDataIndex {
				break
			}
		}
		t, ok := i["type"]
		if ok && t == "audio" {
			data := i["data"].(edge.AudioData)
			audioData[data.Index] = append(audioData[data.Index], data.Data)
		}
		e, ok := i["error"]
		if ok {
			fmt.Printf("has error err: %v\n", e)
		}
	}
	// write data, sort by index
	for _, v := range audioData {
		for _, data := range v {
			broker.Write(data)
		}
	}
	broker.Close()
	return nil
}

func generateHashName(name, voice string) string {
	hash := sha256.Sum256([]byte(name))
	return fmt.Sprintf("%s_%s", voice, hex.EncodeToString(hash[:]))
}
