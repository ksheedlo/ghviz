package simulate

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ksheedlo/ghviz/github"
	"github.com/stretchr/testify/assert"
)

func TestStarCounts(t *testing.T) {
	t.Parallel()

	starEvents := []github.StarEvent{
		github.StarEvent{StarredAt: time.Unix(1, 0)},
		github.StarEvent{StarredAt: time.Unix(2, 0)},
		github.StarEvent{StarredAt: time.Unix(3, 0)},
	}
	starCounts := StarCounts(starEvents)

	for i := 0; (i + 1) < len(starCounts); i++ {
		assert.True(t,
			starCounts[i].Timestamp.Before(starCounts[i+1].Timestamp),
			fmt.Sprintf("Expected starCounts[%d] to be before starCounts[%d] !", i, i+1),
		)
		assert.Equal(t, starCounts[i].Stars, i+1)
	}
	assert.Equal(t, starCounts[len(starCounts)-1].Stars, len(starCounts))
}

func TestMarshalStarCount(t *testing.T) {
	t.Parallel()

	jsonBytes, err := json.Marshal(&StarCount{
		Stars:     5,
		Timestamp: time.Unix(1458966366, 892000000).UTC(),
	})
	assert.NoError(t, err)
	var starCount map[string]interface{}
	assert.NoError(t, json.Unmarshal(jsonBytes, &starCount))
	assert.Equal(t, 5.0, starCount["stars"].(float64))
	assert.Equal(t, "2016-03-26T04:26:06.892Z", starCount["timestamp"].(string))
}
