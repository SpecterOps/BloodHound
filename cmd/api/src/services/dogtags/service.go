package dogtags

// Service defines the interface for the dogtags service
type Service interface {
	GetFlagAsBool(key BoolDogTag) bool
	GetFlagAsString(key StringDogTag) string
	GetFlagAsInt(key IntDogTag) int64
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

func (s *service) GetFlagAsBool(key BoolDogTag) bool {
	if val, err := s.provider.GetFlagAsBool(string(key)); err == nil {
		return val
	}
	return AllBoolDogTags[key].Default
}

func (s *service) GetFlagAsString(key StringDogTag) string {
	if val, err := s.provider.GetFlagAsString(string(key)); err == nil {
		return val
	}
	return AllStringDogTags[key].Default
}

func (s *service) GetFlagAsInt(key IntDogTag) int64 {
	if val, err := s.provider.GetFlagAsInt(string(key)); err == nil {
		return val
	}
	return AllIntDogTags[key].Default
}

func (s *service) GetAllDogTags() map[string]any {
	result := make(map[string]any)

	for key := range AllBoolDogTags {
		result[string(key)] = s.GetFlagAsBool(key)
	}
	for key := range AllStringDogTags {
		result[string(key)] = s.GetFlagAsString(key)
	}
	for key := range AllIntDogTags {
		result[string(key)] = s.GetFlagAsInt(key)
	}

	return result
}
