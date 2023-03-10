package matcher

import (
	"app/model"
	"app/svc"
	"app/utils"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/log"
	_ "github.com/lib/pq"
)

const (
	interval  = 40 //扫一次数据库的时间
	batchSize = 10000
)

var Projects = []string{
	"anyswap", "across",
}

type Matcher struct {
	svc      *svc.ServiceContext
	projects map[string]model.Matcher
}

func NewMatcher(svc *svc.ServiceContext, startIds map[string]uint64) *Matcher {
	if _, ok := startIds["anyswap"]; !ok {
		return nil
	}
	if _, ok := startIds["across"]; !ok {
		return nil
	}
	var projects = make(map[string]model.Matcher)
	for _, project := range Projects {
		projects[project] = NewSimpleInMatcher(project, svc.ProjectsDao, startIds[project])
	}
	return &Matcher{
		svc:      svc,
		projects: projects,
	}
}

func (m *Matcher) Start() {
	for _, matcher := range m.projects {
		go m.StartMatch(matcher)
	}
}

func (m *Matcher) StartMatch(matcher model.Matcher) {
	m.svc.Wg.Add(1)
	defer m.svc.Wg.Done()
	timer := time.NewTimer(1 * time.Second)
	//var last = matcher.LastUnmatchId()
	var last = matcher.LastUnmatchId()
	log.Info("matcher start", "project", matcher.Project(), "Start ID", last)
	for {
		select {
		case <-m.svc.Ctx.Done():
			log.Info("match svc done", "project", matcher.Project(), "current Id", last)
			return
		case <-timer.C:
			latest, err := m.svc.Dao.LatestId(matcher.Project())
			if err != nil {
				log.Error("get latest id failed", "projet", matcher.Project(), "err", err)
				break
			}
			for last < latest {
				var shouldBreak bool
				select {
				case <-m.svc.Ctx.Done():
					shouldBreak = true
				default:
				}
				if shouldBreak {
					break
				}
				right := utils.Min(latest, last+batchSize)
				last -= 300
				total, matched, err := m.BeginMatch(last+1, right, matcher.Project(), matcher)
				if err != nil {
					log.Error("match job failed", "project", matcher.Project(), "from", last+1, "to", right, "err", err)
				} else {
					last = right
					log.Info("match done", "project", matcher.Project(), "current Id", last, "total", total, "matched crossIn", matched)
				}
			}
		}
		timer.Reset(interval * time.Second)
	}
}

func (m *Matcher) BeginMatch(from, to uint64, project string, matcher model.Matcher) (total int, matched int, err error) {
	var stmt string
	switch matcher.(type) {
	case *SimpleInMatcher:
		//stmt = fmt.Sprintf("select * from %s where project = '%s' and direction = '%s' and id >= $1 and id <= $2 and match_id is null and match_tag not in ('0', '1', '2', '3', '4')", m.svc.Dao.Table(), project, model.InDirection)
		stmt = fmt.Sprintf("select %s from %s where direction = '%s' and id >= $1 and id <= $2 and match_id is null order by id asc",
			model.ResultRows, project, model.InDirection)
	default:
		panic("invalid matcher")
	}
	var results model.Datas

	if project == "anyswap" {
		stmt_ := fmt.Sprintf("select %s from %s where id >= $1 and id <= $2 and match_id is null ", model.ResultRows, project)
		err = m.svc.Dao.DB().Select(&results, stmt_, from, to)
		if err != nil {
			return
		}

		cnt, errs := matcher.UpdateAnyswapMatchTag(results)
		if errs != nil {
			utils.LogError(errs, "./error.log")
		}
		log.Info("update anyswap tag done", "updated", cnt)
	}

	err = m.svc.Dao.DB().Select(&results, stmt, from, to)
	if err != nil {
		return
	}
	shouldUpdates, err := matcher.Match(results)
	if err != nil {
		return
	}
	err = m.svc.Dao.Update(project, shouldUpdates)
	if err != nil {
		return
	}
	matched = len(shouldUpdates) / 2
	total = len(results)
	return
}
