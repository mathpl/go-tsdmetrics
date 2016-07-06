package tsdmetrics

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type TagsID string
type Tags map[string]string

func TagsFromString(tags string) (Tags, error) {
	splitTagVals := strings.Split(tags, ",")
	parsedTags := make(Tags, len(splitTagVals))
	if tags == "" {
		return parsedTags, nil
	}

	for _, tagVal := range splitTagVals {
		splitTagVal := strings.Split(tagVal, "=")
		if len(splitTagVal) == 0 {
			continue
		} else if len(splitTagVal) != 2 {
			return nil, fmt.Errorf("Comma delimited tag should follow the format tag=value: %s", tags)
		} else {
			parsedTags[splitTagVal[0]] = splitTagVal[1]
		}
	}
	return parsedTags, nil
}

func (tm Tags) TagsID() TagsID {
	keys := make([]string, len(tm))
	i := 0
	for k, _ := range tm {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	var tagid bytes.Buffer
	for _, k := range keys {
		tagid.Write([]byte(fmt.Sprintf("%s=%s;", k, tm[k])))
	}

	return TagsID(tagid.String())
}

func (tm Tags) String() string {
	var tagid bytes.Buffer
	for k, v := range tm {
		tagid.Write([]byte(fmt.Sprintf("%s=%s ", k, v)))
	}

	return tagid.String()
}

func (tm Tags) AddTags(tags Tags) Tags {
	newTags := make(map[string]string, len(tm)+len(tags))
	for t, v := range tm {
		newTags[t] = v
	}

	for t, v := range tags {
		if _, ok := newTags[t]; !ok {
			newTags[t] = v
		}
	}

	return newTags
}
