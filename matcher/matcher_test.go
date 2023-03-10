package matcher

import (
	"app/config"
	"app/dao"
	"app/model"
	"app/provider/chainbase"
	"app/svc"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"os"
	"testing"
)

var cfg config.Config
var srvCtx *svc.ServiceContext
var d = dao.NewDao("postgres://xiaohui_hu:xiaohui_hu_blocksec888@192.168.3.155:8888/cross_chain?sslmode=disable")

func init() {
	config.LoadCfg(&cfg, "../config.yaml")
	srvCtx = svc.NewServiceContext(context.Background(), &cfg)
	log.Root().SetHandler(log.LvlFilterHandler(
		log.LvlTrace, log.StreamHandler(os.Stderr, log.TerminalFormat(false)),
	))
	chainbase.SetupLimit(10)
}

func TestSimpleInMatcher_UpdateAnyswapMatchTag(t *testing.T) {

	//m.BeginMatch(id-1, id+1, "across", a)
	var data model.Datas
	start_id := uint64(438486)
	id := uint64(438980)
	//stmt := fmt.Sprintf("select %s from across where id = %d", model.ResultRows, id)
	stmt := fmt.Sprintf("select %s from anyswap where direction = 'out' and id > $1 and id < $2 and to_chain in(1,10,56,137,250,42161,43114) and match_id is null", model.ResultRows)
	_ = d.DB().Select(&data, stmt, start_id, id)
	println(len(data))
	var m = &Matcher{}
	m.svc = srvCtx
	a := NewSimpleInMatcher("anyswap", srvCtx.ProjectsDao, id)
	//m.BeginMatch(id-1, id+1, "anyswap", a)
	n, _ := a.UpdateAnyswapMatchTag(data)
	println(n)
}

func TestNewSimpleInMatcher(t *testing.T) {
	ss := "0x8dceda860bf5d56dce08dbe172dafe6faf79b9ea197b5357fb3939add6b49cb8"
	tt := common.BytesToHash([]byte(ss))
	fmt.Println(tt)
}

func TestSimpleInMatcher_Match(t *testing.T) {
	id := 7552917
	stmt := "select * from anyswap where id = 7552917"
	var data model.Datas
	if err := d.DB().Select(&data, stmt); err != nil {
		fmt.Println(err)
	}
	a := NewSimpleInMatcher("anyswap", srvCtx.ProjectsDao, uint64(id))
	a.Match(data)
}

func TestSup(t *testing.T) {
	if _, ok := SupportedChainIds["anyswap"]["3"]; ok {
		println("any")
	}
	if _, ok := SupportedChainIds["across"]["1"]; ok {
		println("ac")
	}
	if _, ok := SupportedChainIds["anyswap"]["1"]; ok {
		println("ay")
	}
}
