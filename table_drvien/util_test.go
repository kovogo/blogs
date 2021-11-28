package table_drvien

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func parseTime(timeStr string) (t time.Time, err error) {
	t, err = time.Parse("2006-01-02 15:04:05", timeStr)
	return
}

//Test_GetZeroTime_HardCode 硬编码测试用例
func Test_GetZeroTime_HardCode(t *testing.T) {
	correctTime, err := parseTime("2021-11-28 00:00:00" )
	assert.Empty(t, err)
	t1, err := parseTime("2021-11-28 10:00:51")
	assert.Empty(t, err)
	assert.Equal(t, GetZeroTimeOfDay(t1), correctTime)
	t2, err := parseTime("2021-11-28 12:00:51")
	assert.Empty(t, err)
	assert.Equal(t, GetZeroTimeOfDay(t2), correctTime)
	t3, err := parseTime("2021-11-28 00:00:00")
	assert.Empty(t, err)
	assert.Equal(t, GetZeroTimeOfDay(t3), correctTime)
	t4, err := parseTime("2021-11-28 23:59:59")
	assert.Empty(t, err)
	assert.Equal(t, GetZeroTimeOfDay(t4), correctTime)
}

type TestCase struct {
	TimeStr []string
	ExpectTime string
}

//Test_GetZeroTime_TableDriven  使用表驱动进行测试
func Test_GetZeroTime_TableDriven(t *testing.T) {
	testCases := []TestCase{
		{
			TimeStr: []string{
				"2021-11-28 10:00:51",
				"2021-11-28 12:00:51",
				"2021-11-28 00:00:00",
				"2021-11-28 23:59:59",
			},
			ExpectTime: "2021-11-28 00:00:00",
		},
	}
	for _, testCase := range testCases{
		expectTime, err := parseTime(testCase.ExpectTime)
		assert.Empty(t, err)
		for _, timeStr := range testCase.TimeStr {
			testTime, err := parseTime(timeStr)
			assert.Empty(t, err)
			result := GetZeroTimeOfDay(testTime)
			assert.Equal(t, result, expectTime)
		}
	}
}

type TestData struct{
	ZeroTimeTestCase []TestCase
}

func Test_GetZeroTime_ReadFile(t *testing.T) {
	bytes, err := os.ReadFile("./testdata.json")
	assert.Empty(t, err)
	data := &TestData{}
	err = json.Unmarshal(bytes, data)
	assert.Empty(t, err)
	for _, testCase := range data.ZeroTimeTestCase {
		expectTime, err := parseTime(testCase.ExpectTime)
		assert.Empty(t, err)
		for _, timeStr := range testCase.TimeStr {
			testTime, err := parseTime(timeStr)
			assert.Empty(t, err)
			result := GetZeroTimeOfDay(testTime)
			assert.Equal(t, result, expectTime)
		}
	}
}