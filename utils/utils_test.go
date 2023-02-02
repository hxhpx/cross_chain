package utils

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	v := "0x"
	fmt.Println(ParseStrToUint64(v))
	v = "0x1"
	fmt.Println(ParseStrToUint64(v))
	v = "0x23"
	fmt.Println(ParseStrToUint64(v))
	v = "1564"
	fmt.Println(ParseStrToUint64(v))
	v = "1ebf"
	fmt.Println(ParseStrToUint64(v))
	v = ""
	fmt.Println(ParseStrToUint64(v))
	v = "pppp"
	fmt.Println(ParseStrToUint64(v))
}

func TestParseTime(t *testing.T) {
	ts := "2020-08-29 03:24:24.111"
	tt, err := ParseDateTime(ts)
	fmt.Println(tt, err, tt.Unix())
}

func TestNetErr(t *testing.T) {
	fmt.Println(IsNetError(nil))
}

func TestDel(t *testing.T) {
	fmt.Println(DeleteSliceElementByValue([]int{1, 2, 3, 4, 5}, 1))
	fmt.Println(DeleteSliceElementByValue([]int{1, 2, 3, 4, 5}, 3))
	fmt.Println(DeleteSliceElementByValue([]int{1, 2, 3, 4, 5}, 5))
	fmt.Println(DeleteSliceElementByValue([]int{1, 2, 3, 4, 5}, 100))
}

func TestContain(t *testing.T) {
	fmt.Println(strings.Contains("apiKey", "invalid apiKey"))
}

func TestHttpGet(t *testing.T) {
	//var dest map[string]interface{}
	url := "https://docs.synapseprotocol.com/reference/contract-addresses"
	dest, err := HttpGet(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(dest))
	decoder.UseNumber()
	s2 := make(map[string]interface{})
	err = decoder.Decode(&s2)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s2)
	/*str := "evm0xc2cb89bbb5bba6e21db1dfe13493dfd7dcbabd68"
	if _, ok := s2[str]; ok {
		fmt.Println(s2[str])
	}*/
}

func TestGetCsv(t *testing.T) {
	resp, e := HttpGet("https://github.com/wormhole-foundation/wormhole-token-list/blob/main/content/by_source.csv")
	if e != nil {
		fmt.Println(e)
		panic(e)
	}
	if resp == nil {
		panic("resp nil")
	}

	r := csv.NewReader(bytes.NewReader(resp))

	records, e := r.ReadAll()
	if e != nil {
		panic(e)
	}

	fmt.Println(records)
}