package tsdmetrics

type TaggedMetric interface {
	GetTags() Tags
	GetMetric() interface{}
	GetTagsID() TagsID
	TagString() string

	AddTags(Tags) TaggedMetric
}

type DefaultTaggedMetric struct {
	Tags   Tags
	Metric interface{}
}

func (m *DefaultTaggedMetric) GetTags() Tags {
	return m.Tags
}

func (m *DefaultTaggedMetric) GetMetric() interface{} {
	return m.Metric
}

func (m *DefaultTaggedMetric) GetTagsID() TagsID {
	return m.GetTagsID()
}

func (m *DefaultTaggedMetric) AddTags(tags Tags) TaggedMetric {
	var newStm DefaultTaggedMetric

	newStm.Metric = m.Metric
	newStm.Tags = m.Tags.AddTags(tags)

	return &newStm
}

func (m *DefaultTaggedMetric) TagString() string {
	return m.Tags.String()
}
