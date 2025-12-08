package dogtags

// Service defines the interface for the dogtags service
type Service interface {
	GetBool(key BoolDogTag) bool
	GetString(key StringDogTag) string
	GetInt(key IntDogTag) int64
	GetFloat(key FloatDogTag) float64
	GetAllDogTags() map[string]any
}

type service struct {
	provider Provider
}

func NewService(provider Provider) Service {
	return &service{provider: provider}
}

// NewDefaultService creates a service with the NoopProvider (all defaults)
func NewDefaultService() Service {
	return &service{provider: NewNoopProvider()}
}

func (s *service) GetBool(key BoolDogTag) bool {
	if val, err := s.provider.GetBool(string(key)); err == nil {
		return val
	}
	return AllBoolDogTags[key].Default
}

func (s *service) GetString(key StringDogTag) string {
	if val, err := s.provider.GetString(string(key)); err == nil {
		return val
	}
	return AllStringDogTags[key].Default
}

func (s *service) GetInt(key IntDogTag) int64 {
	if val, err := s.provider.GetInt(string(key)); err == nil {
		return val
	}
	return AllIntDogTags[key].Default
}

func (s *service) GetFloat(key FloatDogTag) float64 {
	if val, err := s.provider.GetFloat(string(key)); err == nil {
		return val
	}
	return AllFloatDogTags[key].Default
}

func (s *service) GetAllDogTags() map[string]any {
	result := make(map[string]any)

	for key := range AllBoolDogTags {
		result[string(key)] = s.GetBool(key)
	}
	for key := range AllStringDogTags {
		result[string(key)] = s.GetString(key)
	}
	for key := range AllIntDogTags {
		result[string(key)] = s.GetInt(key)
	}
	for key := range AllFloatDogTags {
		result[string(key)] = s.GetFloat(key)
	}

	return result
}
