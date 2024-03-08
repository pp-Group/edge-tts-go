package edge

type Option struct {
	OptID optionID
	Param string
}

type optionID int

const (
	optionIDVoice optionID = iota
	optionIDRate
	optionIDVolume
	optionIDProxy
	optionIDPitch
)

// WithVoice get voice config here: https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/language-support?tabs=tts
func WithVoice(voice string) Option {
	return Option{
		OptID: optionIDVoice,
		Param: voice,
	}
}

func GetVoiceByOption(opts []Option) string {
	for _, opt := range opts {
		if opt.OptID == optionIDVoice {
			return opt.Param
		}
	}
	return ""
}

func WithRate(voice string) Option {
	return Option{
		OptID: optionIDRate,
		Param: voice,
	}
}

func GetRateByOption(opts []Option) string {
	for _, opt := range opts {
		if opt.OptID == optionIDRate {
			return opt.Param
		}
	}
	return ""
}

func WithVolume(voice string) Option {
	return Option{
		OptID: optionIDVolume,
		Param: voice,
	}
}

func GetVolumeByOption(opts []Option) string {
	for _, opt := range opts {
		if opt.OptID == optionIDVolume {
			return opt.Param
		}
	}
	return ""
}

func WithProxy(voice string) Option {
	return Option{
		OptID: optionIDProxy,
		Param: voice,
	}
}

func GetProxyByOption(opts []Option) string {
	for _, opt := range opts {
		if opt.OptID == optionIDProxy {
			return opt.Param
		}
	}
	return ""
}

func WithPitch(pitch string) Option {
	return Option{
		OptID: optionIDPitch,
		Param: pitch,
	}
}

func GetPitchByOption(opts []Option) string {
	for _, opt := range opts {
		if opt.OptID == optionIDPitch {
			return opt.Param
		}
	}
	return ""
}
