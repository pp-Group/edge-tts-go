package edge_tts_go

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	file_helper "github.com/pp-group/file-helper"
	storage "github.com/pp-group/file-helper/storage"

	"github.com/pp-group/edge-tts-go/biz/service/tts/edge"
)

type ISpeech interface {
	GenTTS() (string, func() error)
	URL(filename string) (string, error)
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

func (speech *LocalSpeech) GenTTS() (string, func() error) {
	fileName := speech.generateHashName() + ".mp3"
	return fileName, func() error {
		return gentts(speech.Speech, func() (storage.IWriteBroker, error) {
			return speech.Writer(speech.FileName, nil)
		})
	}
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

func (speech *OssSpeech) GenTTS() (string, func() error) {
	fileName := speech.generateHashName() + ".mp3"
	return fileName, func() error {
		return gentts(speech.Speech, func() (storage.IWriteBroker, error) {
			return speech.Writer(speech.FileName, func() interface{} {
				return speech.bucket
			})
		})
	}
}

func (speech *OssSpeech) URL(filename string) (string, error) {
	return url(func() (storage.IReadBroker, error) {
		return speech.Reader(filename, func() interface{} {
			return speech.bucket
		})
	})
}

func gentts(speech *Speech, brokerFunc func() (storage.IWriteBroker, error)) error {
	fileName := speech.generateHashName() + ".mp3"

	speech.FileName = fileName

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
	Folder   string
	FileName string
}

func NewSpeech(c *edge.Communicate, storage storage.IStorage, folder string) (*Speech, error) {
	s := &Speech{
		Communicate: c,
		IStorage:    storage,
		Folder:      folder,
	}
	return s, nil
}

func (s *Speech) generateHashName() string {
	hash := sha256.Sum256([]byte(s.Text + s.Rate + s.Volume + s.Pitch))
	return fmt.Sprintf("%s_%s", s.VoiceLangRegion, hex.EncodeToString(hash[:]))
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

type OssSpeechFactory struct {
	endpoint string
	ak       string
	sk       string
	bucket   string
	folder   string
}

func NewOssSpeechFactory(endpoint, ak, sk, bucket, folder string) *OssSpeechFactory {
	return &OssSpeechFactory{
		endpoint: endpoint,
		ak:       ak,
		sk:       sk,
		bucket:   bucket,
		folder:   folder,
	}
}

func (factory *OssSpeechFactory) OssSpeech(c *edge.Communicate, folder string) (*OssSpeech, error) {
	if folder != "" {
		return NewOssSpeech(c, factory.endpoint, factory.ak, factory.sk, folder, factory.bucket)
	}
	return NewOssSpeech(c, factory.endpoint, factory.ak, factory.sk, factory.folder, factory.bucket)
}
